#ifndef POWER_H
#define POWER_H

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>
#include <os/lock.h>
#include <stdbool.h>
#include <stdint.h>

// Power source information
typedef struct {
  bool is_present;
  bool is_charging;
  bool is_charged;
  double percentage;
  double temperature;
  double voltage;
  double amperage;
  double power;
  double design_capacity;
  double max_capacity;
  double current_capacity;
  int cycle_count;
  int design_cycle_count;
  int time_remaining; // minutes
  int time_to_full;   // minutes
} power_stats_t;

// System power information
typedef struct {
  double cpu_power;   // watts
  double gpu_power;   // watts
  double total_power; // watts
} system_power_t;

// Error codes for SMC operations
#define SMC_SUCCESS 0
#define SMC_ERROR_INIT_KEYS -1
#define SMC_ERROR_NO_SERVICE -2
#define SMC_ERROR_OPEN_FAILED -3
#define SMC_ERROR_PERMISSION -4

// Error information structure for detailed error reporting
typedef struct {
  int code;            // Error code
  const char *message; // Error message in British English
  int severity;        // 0=info, 1=warning, 2=error
} smc_error_info_t;

// SMC connection structure with RAII support
typedef struct {
  io_connect_t connection;
  smc_error_info_t error;
  os_unfair_lock lock; // Thread-safe operations lock
  bool limited_mode;   // Operating with limited permissions
} smc_connection_t;

// Initialisation options for SMC connection
typedef struct {
  bool allow_limited_mode; // Allow operation with limited permissions
  bool skip_power_keys;    // Skip power source key initialisation
  int timeout_ms;          // Connection timeout in milliseconds
} smc_init_options_t;

// Get power source information
bool get_power_source_info(power_stats_t *stats);

// Get system power consumption
bool get_system_power_info(system_power_t *power);

// Initialize SMC connection
bool init_smc(void);

// Close SMC connection
void close_smc(void);

// Get SMC key value as float
bool get_smc_float(const char *key, float *value);

// SMC management functions
int init_smc_with_options(smc_connection_t *conn,
                          const smc_init_options_t *options);
void get_smc_error_info(const smc_connection_t *conn, smc_error_info_t *error);
bool is_smc_limited_mode(const smc_connection_t *conn);

#endif // POWER_H
