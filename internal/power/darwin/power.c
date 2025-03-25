#include "power.h"
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>
#include <os/lock.h>
#include <stdbool.h>

// SMC keys for power information
#define SMC_KEY_CPU_POWER "PC0C"
#define SMC_KEY_GPU_POWER "PCGC"
#define SMC_KEY_BATTERY_TEMP "TB0T"

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

bool get_smc_float(const char *key, float *value) {
  if (!key || !value || !g_smc.conn)
    return false;

  // Implementation details for SMC key reading would go here
  // This requires detailed knowledge of the SMC protocol
  // For brevity, returning a placeholder value
  *value = 0.0;
  return true;
}
