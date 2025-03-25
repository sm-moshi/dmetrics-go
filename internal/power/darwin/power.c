#include "power.h"
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>
#include <os/lock.h>
#include <stdbool.h>
#include <string.h>
#include <syslog.h>

#ifndef true
#define true 1
#endif

#ifndef false
#define false 0
#endif

// SMC keys for power information
#define SMC_KEY_CPU_POWER "PC0C"
#define SMC_KEY_GPU_POWER "PCGC"
#define SMC_KEY_BATTERY_TEMP "TB0T"

// SMC protocol definitions
#define SMC_CMD_READ_KEY 0x5
#define SMC_CMD_READ_INDEX 0x8
#define SMC_CMD_READ_KEYINFO 0x9

// SMC data types (as uint32_t constants)
#define SMC_TYPE_FP1F 0x66703166 // 'fp1f' in hex
#define SMC_TYPE_FP4C 0x66703463 // 'fp4c' in hex
#define SMC_TYPE_FP5B 0x6670356B // 'fp5b' in hex
#define SMC_TYPE_FP6A 0x66703661 // 'fp6a' in hex
#define SMC_TYPE_FP79 0x66703739 // 'fp79' in hex
#define SMC_TYPE_FP88 0x66703838 // 'fp88' in hex
#define SMC_TYPE_FPA6 0x66706136 // 'fpa6' in hex
#define SMC_TYPE_FPC4 0x66706334 // 'fpc4' in hex
#define SMC_TYPE_FPE2 0x66706532 // 'fpe2' in hex

// SMC key info struct
typedef struct {
  uint32_t dataSize;
  uint32_t dataType;
  uint8_t dataAttributes;
} smc_key_info_t;

// Global SMC connection state
static smc_connection_t g_smc_conn = {.connection = 0,
                                      .error = {.code = SMC_SUCCESS,
                                                .message = NULL,
                                                .severity = SMC_SEVERITY_INFO},
                                      .lock = OS_UNFAIR_LOCK_INIT,
                                      .limited_mode = false};

// Global power source keys
static CFStringRef kPowerSourceStateKey;
static CFStringRef kPowerSourceTypeKey;
static CFStringRef kPowerSourceInternalBattery;
static CFStringRef kPowerSourceChargingKey;
static CFStringRef kPowerSourceChargedKey;
static CFStringRef kPowerSourceCurrentCapacityKey;
static CFStringRef kPowerSourceMaxCapacityKey;
static CFStringRef kPowerSourceTimeToEmptyKey;
static CFStringRef kPowerSourceCycleCountKey;
static CFStringRef kPowerSourceDesignCapacityKey;

// Initialize power source keys
bool init_power_source_keys(void) {
  kPowerSourceStateKey = CFSTR("Power Source State");
  kPowerSourceTypeKey = CFSTR("Type");
  kPowerSourceInternalBattery = CFSTR("InternalBattery");
  kPowerSourceChargingKey = CFSTR("Charging");
  kPowerSourceChargedKey = CFSTR("Charged");
  kPowerSourceCurrentCapacityKey = CFSTR("Current Capacity");
  kPowerSourceMaxCapacityKey = CFSTR("Max Capacity");
  kPowerSourceTimeToEmptyKey = CFSTR("Time to Empty");
  kPowerSourceCycleCountKey = CFSTR("Cycle Count");
  kPowerSourceDesignCapacityKey = CFSTR("Design Capacity");
  return true;
}

bool get_power_source_info(power_stats_t *stats) {
  if (!stats)
    return false;

  // Initialize with defaults
  stats->is_present = false;
  stats->is_charging = false;
  stats->percentage = 0.0;

  // Get power source information
  CFTypeRef powerInfo = IOPSCopyPowerSourcesInfo();
  if (!powerInfo)
    return false;

  CFArrayRef powerSources = IOPSCopyPowerSourcesList(powerInfo);
  if (!powerSources) {
    CFRelease(powerInfo);
    return false;
  }

  // Get first power source (usually the internal battery)
  CFIndex count = CFArrayGetCount(powerSources);
  if (count > 0) {
    CFDictionaryRef powerSource = IOPSGetPowerSourceDescription(
        powerInfo, CFArrayGetValueAtIndex(powerSources, 0));

    if (powerSource) {
      // Check if it's an internal battery
      CFStringRef type = CFDictionaryGetValue(powerSource, CFSTR(kIOPSTypeKey));
      if (type && CFEqual(type, CFSTR(kIOPSInternalBatteryType))) {
        stats->is_present = true;

        // Get power source state and charging state
        CFStringRef powerState =
            CFDictionaryGetValue(powerSource, CFSTR(kIOPSPowerSourceStateKey));
        CFBooleanRef isCharging =
            CFDictionaryGetValue(powerSource, CFSTR(kIOPSIsChargingKey));
        CFBooleanRef isFinishCharging =
            CFDictionaryGetValue(powerSource, CFSTR(kIOPSIsChargedKey));

        if (powerState) {
          // Check if we're on AC power and either charging or fully charged
          bool onAC = CFEqual(powerState, CFSTR(kIOPSACPowerValue));
          bool charging = (isCharging == kCFBooleanTrue);
          bool fullyCharged = (isFinishCharging == kCFBooleanTrue);

          // Update charging state - we're charging if on AC and either actively
          // charging or fully charged
          stats->is_charging = onAC && (charging || fullyCharged);
        }

        // Get current capacity percentage
        CFNumberRef currentCapacity =
            CFDictionaryGetValue(powerSource, CFSTR(kIOPSCurrentCapacityKey));
        if (currentCapacity) {
          int value;
          if (CFNumberGetValue(currentCapacity, kCFNumberIntType, &value)) {
            stats->percentage = (double)value;
          }
        }
      }
    }
  }

  CFRelease(powerSources);
  CFRelease(powerInfo);
  return true;
}

bool get_system_power_info(system_power_t *power) {
  if (!power || !g_smc_conn.connection)
    return false;

  float value;
  bool success = true;

  // Get CPU power
  if (get_smc_float(SMC_KEY_CPU_POWER, &value)) {
    power->cpu_power = value;
  } else {
    success = false;
  }

  // Get GPU power
  if (get_smc_float(SMC_KEY_GPU_POWER, &value)) {
    power->gpu_power = value;
  } else {
    success = false;
  }

  // Calculate total power
  power->total_power = power->cpu_power + power->gpu_power;
  return success;
}

// Error logging function
static void log_smc_error(smc_error_info_t *error, const char *context) {
  if (!error)
    return;

  int priority;
  switch (error->severity) {
  case 0:
    priority = LOG_INFO;
    break;
  case 1:
    priority = LOG_WARNING;
    break;
  case 2:
    priority = LOG_ERR;
    break;
  default:
    priority = LOG_DEBUG;
  }

  syslog(priority, "SMC Error [%s]: %s (code: %d)", context, error->message,
         error->code);
}

/**
 * Initialises the SMC connection with the specified options.
 * This function handles the low-level setup of the SMC connection
 * and applies the provided configuration options.
 *
 * Thread-safety: The caller must hold the connection lock.
 *
 * @param conn Pointer to the connection structure to initialise
 * @param options Configuration options for the connection
 * @return SMC_SUCCESS on success, or an error code on failure
 */
int init_smc_with_options(smc_connection_t *conn,
                          const smc_init_options_t *options) {
  if (!conn || !options) {
    return SMC_ERROR_INVALID_ARGS;
  }

  // Find and open SMC service
  io_service_t service = IOServiceGetMatchingService(
      kIOMainPortDefault, IOServiceMatching("AppleSMC"));

  if (!service) {
    conn->error.code = SMC_ERROR_NO_SERVICE;
    conn->error.message = "SMC service not found";
    conn->error.severity = 2;
    log_smc_error(&conn->error, "Service Discovery");
    return SMC_ERROR_NO_SERVICE;
  }

  // Open connection to the SMC
  kern_return_t result =
      IOServiceOpen(service, mach_task_self(), 0, &conn->connection);

  IOObjectRelease(service);

  if (result != KERN_SUCCESS) {
    conn->error.code = SMC_ERROR_OPEN_FAILED;
    conn->error.message = "Failed to open SMC connection";
    conn->error.severity = 2;
    log_smc_error(&conn->error, "Connection Open");
    return SMC_ERROR_OPEN_FAILED;
  }

  // Apply options
  conn->limited_mode = options->allow_limited_mode;

  // Initialize power source keys if not in limited mode
  if (!options->skip_power_keys && !conn->limited_mode) {
    if (!init_power_source_keys()) {
      cleanup_smc_connection(conn);
      conn->error.code = SMC_ERROR_INIT_FAILED;
      conn->error.message = "Failed to initialize power source keys";
      conn->error.severity = 2;
      log_smc_error(&conn->error, "Power Keys Init");
      return SMC_ERROR_INIT_FAILED;
    }
  }

  conn->error.code = SMC_SUCCESS;
  conn->error.message = "SMC initialised successfully";
  conn->error.severity = 0;
  log_smc_error(&conn->error, "Initialisation");
  return SMC_SUCCESS;
}

void get_smc_error_info(smc_connection_t *conn, smc_error_info_t *error) {
  if (!conn || !error)
    return;

  os_unfair_lock_lock(&conn->lock);
  *error = conn->error;
  os_unfair_lock_unlock(&conn->lock);
}

bool is_smc_limited_mode(smc_connection_t *conn) {
  if (!conn)
    return false;

  os_unfair_lock_lock(&conn->lock);
  bool limited = conn->limited_mode;
  os_unfair_lock_unlock(&conn->lock);
  return limited;
}

/**
 * Reads an SMC key value with thread-safe locking.
 * This is an internal helper function used by get_smc_float.
 *
 * @param conn The SMC connection handle
 * @param key_str The 4-character SMC key to read
 * @param cmd Pointer to store the command results
 * @return true if successful, false on error
 */
static bool read_smc_key(io_connect_t conn, const char *key_str,
                         smc_cmd_t *cmd) {
  if (!key_str || !cmd) {
    syslog(LOG_ERR, "SMC Error: Invalid parameters for key reading");
    return false;
  }

  // Convert key string to uint32_t
  uint32_t key =
      (key_str[0] << 24) | (key_str[1] << 16) | (key_str[2] << 8) | key_str[3];

  // Get key info first
  smc_key_info_t key_info;
  memset(&key_info, 0, sizeof(key_info));

  cmd->key = key;
  cmd->cmd = SMC_CMD_READ_KEYINFO;
  cmd->keyInfo = 0;

  size_t size = sizeof(smc_cmd_t);
  kern_return_t result =
      IOConnectCallStructMethod(conn, 2, cmd, sizeof(smc_cmd_t), cmd, &size);

  if (result != KERN_SUCCESS) {
    syslog(LOG_ERR, "SMC Error: Failed to read key info for %s", key_str);
    return false;
  }

  // Now read the actual key value
  cmd->cmd = SMC_CMD_READ_KEY;
  cmd->keyInfo = 0;
  memset(cmd->data, 0, sizeof(cmd->data));

  result =
      IOConnectCallStructMethod(conn, 2, cmd, sizeof(smc_cmd_t), cmd, &size);

  if (result != KERN_SUCCESS) {
    syslog(LOG_ERR, "SMC Error: Failed to read key value for %s", key_str);
    return false;
  }

  syslog(LOG_DEBUG, "SMC: Successfully read key %s", key_str);
  return true;
}

bool init_smc(void) {
  os_unfair_lock_lock(&g_smc_conn.lock);
  smc_init_options_t options = {.allow_limited_mode = false,
                                .skip_power_keys = false,
                                .timeout_ms = 1000};
  int result = init_smc_with_options(&g_smc_conn, &options);
  if (result != SMC_SUCCESS) {
    g_smc_conn.connection = 0;
    g_smc_conn.limited_mode = false;
  }
  os_unfair_lock_unlock(&g_smc_conn.lock);
  return result == SMC_SUCCESS;
}

bool close_smc(void) {
  os_unfair_lock_lock(&g_smc_conn.lock);

  bool success = true;
  if (g_smc_conn.connection != 0) {
    kern_return_t result = IOServiceClose(g_smc_conn.connection);
    if (result != KERN_SUCCESS) {
      g_smc_conn.error.code = SMC_ERROR_INIT_FAILED;
      g_smc_conn.error.message = "Failed to close SMC connection";
      g_smc_conn.error.severity = SMC_SEVERITY_ERROR;
      log_smc_error(&g_smc_conn.error, "Connection Close");
      success = false;
    } else {
      // Cleanup power source keys if they were initialised
      if (!g_smc_conn.limited_mode) {
        cleanup_power_source_keys();
      }

      // Reset connection state
      g_smc_conn.connection = 0;
      g_smc_conn.error.code = SMC_SUCCESS;
      g_smc_conn.error.message = "SMC connection closed successfully";
      g_smc_conn.error.severity = SMC_SEVERITY_INFO;
      log_smc_error(&g_smc_conn.error, "Connection Close");
    }
  }

  os_unfair_lock_unlock(&g_smc_conn.lock);
  return success;
}

/**
 * Decodes an SMC float value from the raw command data.
 * Each SMC float type uses a different fixed-point format:
 *
 * - FP1F: 1-bit exponent, 15-bit fraction
 *   Format: [1-bit exp][15-bit frac]
 *   Range: 0 to ~1.999
 *
 * - FP4C: 4-bit exponent, 12-bit fraction
 *   Format: [4-bit exp][12-bit frac]
 *   Range: 0 to ~15.999
 *
 * - FP5B: 5-bit exponent, 11-bit fraction
 *   Format: [5-bit exp][11-bit frac]
 *   Range: 0 to ~31.999
 *
 * - FP6A: 6-bit exponent, 10-bit fraction
 *   Format: [6-bit exp][10-bit frac]
 *   Range: 0 to ~63.999
 *
 * - FP79: 7-bit exponent, 9-bit fraction
 *   Format: [7-bit exp][9-bit frac]
 *   Range: 0 to ~127.999
 *
 * - FP88: 8-bit integer, 8-bit fraction
 *   Format: [8-bit int][8-bit frac]
 *   Range: 0 to 255.99609375
 *
 * - FPA6: 10-bit integer, 6-bit fraction
 *   Format: [10-bit int][6-bit frac]
 *   Range: 0 to 1023.984375
 *
 * - FPC4: 12-bit integer, 4-bit fraction
 *   Format: [12-bit int][4-bit frac]
 *   Range: 0 to 4095.9375
 *
 * - FPE2: 14-bit integer, 2-bit fraction
 *   Format: [14-bit int][2-bit frac]
 *   Range: 0 to 16383.75
 *
 * @param cmd Pointer to the SMC command structure containing the data
 * @return The decoded float value, or 0.0f on error
 */
float decode_smc_float(const smc_cmd_t *cmd) {
  if (!cmd) {
    syslog(LOG_ERR, "SMC Error: NULL command pointer in decode_smc_float");
    return 0.0f;
  }

  uint32_t data_type = cmd->keyInfo & 0xFFFFFFFF;
  const uint8_t *data = cmd->data;

  switch (data_type) {
  case SMC_TYPE_FP1F: // 1-bit exponent, 15-bit fraction
    return (float)data[0] + ((float)(data[1] & 0x7F) / (1 << 7));

  case SMC_TYPE_FP4C: { // 4-bit exponent, 12-bit fraction
    uint16_t value = (data[0] << 8) | data[1];
    return (float)value / 4096.0f; // 2^12 = 4096
  }

  case SMC_TYPE_FP5B: { // 5-bit exponent, 11-bit fraction
    uint16_t value = (data[0] << 8) | data[1];
    return (float)value / 2048.0f; // 2^11 = 2048
  }

  case SMC_TYPE_FP6A: { // 6-bit exponent, 10-bit fraction
    uint16_t value = (data[0] << 8) | data[1];
    return (float)value / 1024.0f; // 2^10 = 1024
  }

  case SMC_TYPE_FP79: { // 7-bit exponent, 9-bit fraction
    uint16_t value = (data[0] << 8) | data[1];
    return (float)value / 512.0f; // 2^9 = 512
  }

  case SMC_TYPE_FP88: // 8-bit integer, 8-bit fraction
    return (float)data[0] + ((float)data[1] / (1 << 8));

  case SMC_TYPE_FPA6: { // 10-bit integer, 6-bit fraction
    uint16_t value = (data[0] << 8) | data[1];
    uint16_t integer = value >> 6;
    uint16_t fraction = value & 0x3F;
    return (float)integer + ((float)fraction / 64.0f); // 2^6 = 64
  }

  case SMC_TYPE_FPC4: { // 12-bit integer, 4-bit fraction
    uint8_t integer = data[0];
    uint8_t fraction = data[1] & 0xF; // Get the lower 4 bits
    return (float)integer + ((float)fraction / 16.0f);
  }

  case SMC_TYPE_FPE2: { // 14-bit integer, 2-bit fraction
    uint8_t integer = data[0];
    uint8_t fraction = data[1] & 0x3; // Get the bottom 2 bits
    return (float)integer + ((float)fraction / 4.0f);
  }

  default:
    syslog(LOG_WARNING, "SMC Warning: Unsupported float type: 0x%08x",
           data_type);
    return 0.0f;
  }
}

bool get_smc_float(const char *key, float *value) {
  if (!key || !value || !g_smc_conn.connection)
    return false;

  os_unfair_lock_lock(&g_smc_conn.lock);

  smc_cmd_t cmd;
  memset(&cmd, 0, sizeof(cmd));

  bool success = read_smc_key(g_smc_conn.connection, key, &cmd);
  if (success) {
    *value = decode_smc_float(&cmd);
  }

  os_unfair_lock_unlock(&g_smc_conn.lock);
  return success;
}

// Cleanup power source keys
void cleanup_power_source_keys(void) {
  if (kPowerSourceStateKey) {
    CFRelease(kPowerSourceStateKey);
    kPowerSourceStateKey = NULL;
  }
  if (kPowerSourceTypeKey) {
    CFRelease(kPowerSourceTypeKey);
    kPowerSourceTypeKey = NULL;
  }
  if (kPowerSourceInternalBattery) {
    CFRelease(kPowerSourceInternalBattery);
    kPowerSourceInternalBattery = NULL;
  }
  if (kPowerSourceChargingKey) {
    CFRelease(kPowerSourceChargingKey);
    kPowerSourceChargingKey = NULL;
  }
  if (kPowerSourceChargedKey) {
    CFRelease(kPowerSourceChargedKey);
    kPowerSourceChargedKey = NULL;
  }
  if (kPowerSourceCurrentCapacityKey) {
    CFRelease(kPowerSourceCurrentCapacityKey);
    kPowerSourceCurrentCapacityKey = NULL;
  }
  if (kPowerSourceMaxCapacityKey) {
    CFRelease(kPowerSourceMaxCapacityKey);
    kPowerSourceMaxCapacityKey = NULL;
  }
  if (kPowerSourceTimeToEmptyKey) {
    CFRelease(kPowerSourceTimeToEmptyKey);
    kPowerSourceTimeToEmptyKey = NULL;
  }
  if (kPowerSourceCycleCountKey) {
    CFRelease(kPowerSourceCycleCountKey);
    kPowerSourceCycleCountKey = NULL;
  }
  if (kPowerSourceDesignCapacityKey) {
    CFRelease(kPowerSourceDesignCapacityKey);
    kPowerSourceDesignCapacityKey = NULL;
  }
}

void cleanup_smc_connection(smc_connection_t *conn) {
  if (!conn)
    return;

  os_unfair_lock_lock(&conn->lock);

  if (conn->connection) {
    // Close the SMC connection
    IOServiceClose(conn->connection);
    conn->connection = 0;
  }

  // Cleanup power source keys if they were initialised
  if (!conn->limited_mode) {
    cleanup_power_source_keys();
  }

  // Reset error state
  conn->error.code = SMC_SUCCESS;
  conn->error.message = "Connection closed";
  conn->error.severity = 0;

  os_unfair_lock_unlock(&conn->lock);
}
