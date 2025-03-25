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
#define SMC_ERROR_INIT_KEYS 1
#define SMC_ERROR_NO_SERVICE 2
#define SMC_ERROR_OPEN_FAILED 3

// Error severity levels
#define SMC_SEVERITY_INFO 0    // Informational messages
#define SMC_SEVERITY_WARNING 1 // Warning conditions
#define SMC_SEVERITY_ERROR 2   // Error conditions

// Error information structure for detailed error reporting
// All error messages are logged to syslog with appropriate priority levels:
// - Severity 0 (Info) -> LOG_INFO
// - Severity 1 (Warning) -> LOG_WARNING
// - Severity 2 (Error) -> LOG_ERR
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
void get_smc_error_info(smc_connection_t *conn, smc_error_info_t *error);
bool is_smc_limited_mode(smc_connection_t *conn);
void cleanup_smc_connection(smc_connection_t *conn);

// SMC data types for testing
#define SMC_TYPE_FP1F 0x66703166 // 'fp1f' in hex
#define SMC_TYPE_FP4C 0x66703463 // 'fp4c' in hex
#define SMC_TYPE_FP5B 0x6670356B // 'fp5b' in hex
#define SMC_TYPE_FP6A 0x66703661 // 'fp6a' in hex
#define SMC_TYPE_FP79 0x66703739 // 'fp79' in hex
#define SMC_TYPE_FP88 0x66703838 // 'fp88' in hex
#define SMC_TYPE_FPA6 0x66706136 // 'fpa6' in hex
#define SMC_TYPE_FPC4 0x66706334 // 'fpc4' in hex
#define SMC_TYPE_FPE2 0x66706532 // 'fpe2' in hex

// SMC command struct for testing
typedef struct {
  uint32_t key;
  uint32_t versioning;
  uint8_t cmd;
  uint32_t result;
  uint32_t unknown;
  uint8_t data[32];
  uint32_t keyInfo;
} smc_cmd_t;

// SMC float decoding function
float decode_smc_float(const smc_cmd_t *cmd);

#endif // POWER_H
