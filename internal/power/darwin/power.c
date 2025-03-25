#include "power.h"
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>
#include <os/lock.h>
#include <stdbool.h>
#include <string.h>

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

// SMC command struct
typedef struct {
  uint32_t key;
  uint32_t versioning;
  uint8_t cmd;
  uint32_t result;
  uint32_t unknown;
  uint8_t data[32];
  uint32_t keyInfo;
} smc_cmd_t;

// Global SMC connection with thread safety
static struct {
  io_connect_t conn;
  os_unfair_lock lock;
  bool initialised;
  smc_error_info_t last_error;
} g_smc = {
    .conn = 0,
    .lock = OS_UNFAIR_LOCK_INIT,
    .initialised = false,
    .last_error = {.code = SMC_SUCCESS, .message = "No error", .severity = 0}};

// Forward declarations for internal functions
static bool init_smc_internal(io_connect_t *conn, smc_error_info_t *error);
static void cleanup_smc_internal(io_connect_t conn);

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

// Add an initialization function
static bool init_power_source_keys(void) {
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

  CFTypeRef blob = IOPSCopyPowerSourcesInfo();
  CFArrayRef sources = IOPSCopyPowerSourcesList(blob);
  if (!sources) {
    CFRelease(blob);
    return false;
  }

  bool success = false;
  CFIndex count = CFArrayGetCount(sources);
  if (count > 0) {
    CFDictionaryRef ps =
        IOPSGetPowerSourceDescription(blob, CFArrayGetValueAtIndex(sources, 0));
    if (ps) {
      // Get power source type
      CFStringRef ps_type = CFDictionaryGetValue(ps, kPowerSourceTypeKey);
      if (ps_type && CFEqual(ps_type, kPowerSourceInternalBattery)) {
        stats->is_present = true;

        // Get charging state
        CFStringRef charging_state =
            CFDictionaryGetValue(ps, kPowerSourceStateKey);
        if (charging_state) {
          stats->is_charging = CFEqual(charging_state, kPowerSourceChargingKey);
          stats->is_charged = CFEqual(charging_state, kPowerSourceChargedKey);
        }

        // Get current capacity
        CFNumberRef current_cap =
            CFDictionaryGetValue(ps, kPowerSourceCurrentCapacityKey);
        if (current_cap) {
          int value;
          if (CFNumberGetValue(current_cap, kCFNumberIntType, &value)) {
            stats->current_capacity = (double)value;
          }
        }

        // Get max capacity
        CFNumberRef max_cap =
            CFDictionaryGetValue(ps, kPowerSourceMaxCapacityKey);
        if (max_cap) {
          int value;
          if (CFNumberGetValue(max_cap, kCFNumberIntType, &value)) {
            stats->max_capacity = (double)value;
          }
        }

        // Get design capacity
        CFNumberRef design_cap =
            CFDictionaryGetValue(ps, kPowerSourceDesignCapacityKey);
        if (design_cap) {
          int value;
          if (CFNumberGetValue(design_cap, kCFNumberIntType, &value)) {
            stats->design_capacity = (double)value;
          } else {
            // If design capacity isn't available, use max capacity as fallback
            stats->design_capacity = stats->max_capacity;
          }
        } else {
          // Fallback to max capacity if design capacity isn't available
          stats->design_capacity = stats->max_capacity;
        }

        // Get time remaining
        CFNumberRef time = CFDictionaryGetValue(ps, kPowerSourceTimeToEmptyKey);
        if (time) {
          CFNumberGetValue(time, kCFNumberIntType, &stats->time_remaining);
        }

        // Get cycle count
        CFNumberRef cycles =
            CFDictionaryGetValue(ps, kPowerSourceCycleCountKey);
        if (cycles) {
          CFNumberGetValue(cycles, kCFNumberIntType, &stats->cycle_count);
        }

        success = true;
      }

      // After getting the power source dictionary
      CFShow(ps); // This will print all available keys and values
    }
  }

  CFRelease(sources);
  CFRelease(blob);
  return success;
}

bool get_system_power_info(system_power_t *power) {
  if (!power || !g_smc.conn)
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

int init_smc_with_options(smc_connection_t *conn,
                          const smc_init_options_t *options) {
  if (!conn) {
    return SMC_ERROR_INIT_KEYS;
  }

  // Initialize the connection structure
  conn->connection = 0;
  conn->error.code = SMC_SUCCESS;
  conn->error.message = "Initialising";
  conn->error.severity = 0;
  conn->lock = OS_UNFAIR_LOCK_INIT;
  conn->limited_mode = options ? options->allow_limited_mode : false;

  // Find and open SMC service
  io_service_t service = IOServiceGetMatchingService(
      kIOMainPortDefault, IOServiceMatching("AppleSMC"));

  if (!service) {
    conn->error.code = SMC_ERROR_NO_SERVICE;
    conn->error.message = "SMC service not found";
    conn->error.severity = 2;
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
    return SMC_ERROR_OPEN_FAILED;
  }

  // Initialize SMC keys if needed
  if (!options || !options->skip_power_keys) {
    // Key initialization would go here
    // For now, just mark as successful
    conn->error.code = SMC_SUCCESS;
    conn->error.message = "SMC initialised successfully";
    conn->error.severity = 0;
  }

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

static bool read_smc_key(io_connect_t conn, const char *key_str,
                         smc_cmd_t *cmd) {
  if (!key_str || !cmd)
    return false;

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

  if (result != KERN_SUCCESS)
    return false;

  // Now read the actual key value
  cmd->cmd = SMC_CMD_READ_KEY;
  cmd->keyInfo = 0;
  memset(cmd->data, 0, sizeof(cmd->data));

  result =
      IOConnectCallStructMethod(conn, 2, cmd, sizeof(smc_cmd_t), cmd, &size);

  return result == KERN_SUCCESS;
}

static float decode_smc_float(const smc_cmd_t *cmd) {
  if (!cmd)
    return 0.0f;

  uint32_t data_type = cmd->keyInfo & 0xFFFFFFFF;
  const uint8_t *data = cmd->data;

  switch (data_type) {
  case SMC_TYPE_FP1F: // 1-bit exponent, 15-bit fraction
    return (float)data[0] + ((float)data[1] / (1 << 15));

  case SMC_TYPE_FP4C: // 4-bit exponent, 12-bit fraction
    return (float)((data[0] << 8 | data[1]) >> 4) +
           ((float)(data[1] & 0xF) / (1 << 12));

  case SMC_TYPE_FP88: // 8-bit integer, 8-bit fraction
    return (float)data[0] + ((float)data[1] / (1 << 8));

  default:
    return 0.0f; // Unsupported type
  }
}

bool get_smc_float(const char *key, float *value) {
  if (!key || !value || !g_smc.conn)
    return false;

  os_unfair_lock_lock(&g_smc.lock);

  smc_cmd_t cmd;
  memset(&cmd, 0, sizeof(cmd));

  bool success = read_smc_key(g_smc.conn, key, &cmd);
  if (success) {
    *value = decode_smc_float(&cmd);
  }

  os_unfair_lock_unlock(&g_smc.lock);
  return success;
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

  // Reset error state
  conn->error.code = SMC_SUCCESS;
  conn->error.message = "Connection closed";
  conn->error.severity = 0;
  conn->limited_mode = false;

  os_unfair_lock_unlock(&conn->lock);
}
