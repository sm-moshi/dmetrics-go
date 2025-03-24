#ifndef POWER_H
#define POWER_H

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>
#include <stdbool.h>
#include <stdint.h>

// Power source information
typedef struct
{
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
typedef struct
{
    double cpu_power;   // watts
    double gpu_power;   // watts
    double total_power; // watts
} system_power_t;

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

#endif // POWER_H
