/**
 * @file power.h
 * @brief Power source information retrieval for macOS using IOKit
 *
 * This header defines the interface for retrieving basic power source
 * information from macOS using the IOPowerSources API. The implementation
 * prioritises:
 * - Stability: Using well-documented, high-level IOKit APIs
 * - Simplicity: Focusing on essential power metrics
 * - Safety: Proper resource management and error handling
 *
 * The v0.1 implementation specifically avoids direct SMC access in favour of
 * the more stable IOPowerSources API for better compatibility and
 * maintainability.
 */

#ifndef POWER_H
#define POWER_H

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>
#include <os/lock.h>
#include <stdbool.h>
#include <stdint.h>

/**
 * @brief Basic power source statistics
 *
 * Contains essential power information retrieved from IOPowerSources.
 * This structure is designed to be minimal for v0.1, focusing on
 * the most commonly needed power metrics.
 */
typedef struct {
  bool is_present;    /**< Whether a battery is present */
  bool is_charging;   /**< Whether the battery is currently charging */
  bool is_charged;    /**< Whether the battery is fully charged */
  bool is_ac_present; /**< Whether AC power is connected */
  double percentage;  /**< Battery charge percentage (0-100) */
  double
      time_remaining; /**< Time remaining in minutes (negative when charging) */
  int cycle_count;    /**< Battery cycle count */
  double current_capacity; /**< Current capacity in mAh */
  double max_capacity;     /**< Maximum capacity in mAh */
  double design_capacity;  /**< Design capacity in mAh */
} power_stats_t;

/**
 * @brief System power consumption information.
 * Provides detailed power usage metrics for major system components.
 */
typedef struct {
  double cpu_power;   ///< CPU power consumption in Watts
  double gpu_power;   ///< GPU power consumption in Watts
  double total_power; ///< Total system power consumption in Watts
} system_power_t;

// Error codes for SMC operations
#define SMC_SUCCESS 0            ///< Operation completed successfully
#define SMC_ERROR_INIT_KEYS 1    ///< Failed to initialise power source keys
#define SMC_ERROR_NO_SERVICE 2   ///< SMC service not found
#define SMC_ERROR_OPEN_FAILED 3  ///< Failed to open SMC connection
#define SMC_ERROR_INVALID_ARGS 4 ///< Invalid arguments provided
#define SMC_ERROR_INIT_FAILED 5  ///< General initialisation failure

// Error severity levels for logging and diagnostics
#define SMC_SEVERITY_INFO 0    ///< Informational messages, no error
#define SMC_SEVERITY_WARNING 1 ///< Warning conditions, operation may proceed
#define SMC_SEVERITY_ERROR 2   ///< Error conditions, operation failed

/**
 * @brief Error information structure for detailed error reporting.
 * Provides comprehensive error context for debugging and logging.
 * Thread-safe: This structure is protected by the connection lock.
 */
typedef struct {
  int code;            ///< Error code from SMC_* constants
  const char *message; ///< Error message in British English
  int severity;        ///< Severity level from SMC_SEVERITY_* constants
} smc_error_info_t;

/**
 * @brief SMC connection structure with thread-safety support.
 * Maintains the state of a connection to the SMC and provides
 * synchronisation primitives for thread-safe access.
 */
typedef struct {
  io_connect_t connection; ///< IOKit connection handle
  smc_error_info_t error;  ///< Last error information
  os_unfair_lock lock;     ///< Thread-safe operations lock
  bool limited_mode;       ///< Operating with limited permissions
} smc_connection_t;

/**
 * @brief Initialisation options for SMC connection.
 * Controls the behaviour of the SMC connection and can be used
 * to optimise performance or handle permissions.
 */
typedef struct {
  bool allow_limited_mode; ///< Allow operation with limited permissions
  bool skip_power_keys;    ///< Skip power source key initialisation
  int timeout_ms;          ///< Connection timeout in milliseconds
} smc_init_options_t;

/**
 * @brief Retrieve current power source information
 *
 * Uses IOPowerSources API to gather basic power statistics. This function:
 * - Checks for battery presence
 * - Determines charging state
 * - Calculates current charge percentage
 *
 * @param[out] stats Pointer to power_stats_t structure to be populated
 * @return true if successful, false if an error occurred
 */
bool get_power_source_info(power_stats_t *stats);

/**
 * @brief Get system power consumption.
 * Retrieves current power consumption metrics for CPU, GPU, and total system.
 *
 * @param power Pointer to system_power_t structure to fill
 * @return true if successful, false on error
 */
bool get_system_power_info(system_power_t *power);

/**
 * @brief Initialise SMC connection with default options.
 * Thread-safe: Uses internal locking to ensure thread safety.
 *
 * @return true if successful, false on failure
 */
bool init_smc(void);

/**
 * @brief Close the SMC connection and clean up resources.
 * Thread-safe: Uses internal locking to ensure thread safety.
 *
 * @return true if successful, false on error
 */
bool close_smc(void);

/**
 * @brief Get SMC key value as float.
 * Thread-safe: Uses internal locking to ensure thread safety.
 *
 * @param key SMC key to read (e.g., "PC0C" for CPU power)
 * @param value Pointer to store the decoded float value
 * @return true if successful, false on failure
 */
bool get_smc_float(const char *key, float *value);

/**
 * @brief Initialise SMC connection with custom options.
 * Thread-safe: The caller must hold the connection lock.
 *
 * @param conn Pointer to the connection structure to initialise
 * @param options Configuration options for the connection
 * @return SMC_SUCCESS on success, or an error code on failure
 */
int init_smc_with_options(smc_connection_t *conn,
                          const smc_init_options_t *options);

/**
 * @brief Get error information from SMC connection.
 * Thread-safe: Uses internal locking to ensure thread safety.
 *
 * @param conn Pointer to the SMC connection
 * @param error Pointer to error structure to fill
 */
void get_smc_error_info(smc_connection_t *conn, smc_error_info_t *error);

/**
 * @brief Check if SMC is operating in limited mode.
 * Thread-safe: Uses internal locking to ensure thread safety.
 *
 * @param conn Pointer to the SMC connection
 * @return true if in limited mode, false otherwise
 */
bool is_smc_limited_mode(smc_connection_t *conn);

/**
 * @brief Clean up SMC connection resources.
 * Thread-safe: Uses internal locking to ensure thread safety.
 *
 * @param conn Pointer to the SMC connection to clean up
 */
void cleanup_smc_connection(smc_connection_t *conn);

// SMC data types for testing
#define SMC_TYPE_FP1F 0x66703166 ///< 'fp1f': 15.1 fixed point
#define SMC_TYPE_FP4C 0x66703463 ///< 'fp4c': 12.4 fixed point
#define SMC_TYPE_FP5B 0x6670356B ///< 'fp5b': 11.5 fixed point
#define SMC_TYPE_FP6A 0x66703661 ///< 'fp6a': 10.6 fixed point
#define SMC_TYPE_FP79 0x66703739 ///< 'fp79': 9.7 fixed point
#define SMC_TYPE_FP88 0x66703838 ///< 'fp88': 8.8 fixed point
#define SMC_TYPE_FPA6 0x66706136 ///< 'fpa6': 10.6 fixed point
#define SMC_TYPE_FPC4 0x66706334 ///< 'fpc4': 12.4 fixed point
#define SMC_TYPE_FPE2 0x66706532 ///< 'fpe2': 14.2 fixed point

/**
 * @brief SMC command structure for low-level operations.
 * Used for direct communication with the SMC.
 */
typedef struct {
  uint32_t key;        ///< SMC key to access
  uint32_t versioning; ///< Version information
  uint8_t cmd;         ///< Command to execute
  uint32_t result;     ///< Operation result
  uint32_t unknown;    ///< Reserved for SMC use
  uint8_t data[32];    ///< Data buffer
  uint32_t keyInfo;    ///< Key information
} smc_cmd_t;

/**
 * @brief Decode SMC float value from raw data.
 * Handles various SMC fixed-point formats.
 *
 * @param cmd Pointer to the SMC command structure containing the data
 * @return The decoded float value, or 0.0f on error
 */
float decode_smc_float(const smc_cmd_t *cmd);

/**
 * @brief Clean up power source key resources.
 * Thread-safe: Uses internal locking to ensure thread safety.
 * Releases all CoreFoundation objects used for power source keys.
 */
void cleanup_power_source_keys(void);

#endif /* POWER_H */
