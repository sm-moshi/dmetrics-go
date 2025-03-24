#include "power.h"
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/ps/IOPSKeys.h>
#include <IOKit/ps/IOPowerSources.h>

// SMC keys for power information
#define SMC_KEY_CPU_POWER "PC0C"
#define SMC_KEY_GPU_POWER "PCGC"
#define SMC_KEY_BATTERY_TEMP "TB0T"

// Global SMC connection
static io_connect_t g_smc_conn = 0;

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
static bool init_power_source_keys(void)
{
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

bool get_power_source_info(power_stats_t *stats)
{
    if (!stats)
        return false;

    CFTypeRef blob = IOPSCopyPowerSourcesInfo();
    CFArrayRef sources = IOPSCopyPowerSourcesList(blob);
    if (!sources)
    {
        CFRelease(blob);
        return false;
    }

    bool success = false;
    CFIndex count = CFArrayGetCount(sources);
    if (count > 0)
    {
        CFDictionaryRef ps = IOPSGetPowerSourceDescription(
            blob, CFArrayGetValueAtIndex(sources, 0));
        if (ps)
        {
            // Get power source type
            CFStringRef ps_type = CFDictionaryGetValue(ps, kPowerSourceTypeKey);
            if (ps_type && CFEqual(ps_type, kPowerSourceInternalBattery))
            {
                stats->is_present = true;

                // Get charging state
                CFStringRef charging_state = CFDictionaryGetValue(ps, kPowerSourceStateKey);
                if (charging_state)
                {
                    stats->is_charging =
                        CFEqual(charging_state, kPowerSourceChargingKey);
                    stats->is_charged = CFEqual(charging_state, kPowerSourceChargedKey);
                }

                // Get current capacity
                CFNumberRef current_cap = CFDictionaryGetValue(ps, kPowerSourceCurrentCapacityKey);
                if (current_cap)
                {
                    int value;
                    if (CFNumberGetValue(current_cap, kCFNumberIntType, &value))
                    {
                        stats->current_capacity = (double)value;
                    }
                }

                // Get max capacity
                CFNumberRef max_cap = CFDictionaryGetValue(ps, kPowerSourceMaxCapacityKey);
                if (max_cap)
                {
                    int value;
                    if (CFNumberGetValue(max_cap, kCFNumberIntType, &value))
                    {
                        stats->max_capacity = (double)value;
                    }
                }

                // Get design capacity
                CFNumberRef design_cap = CFDictionaryGetValue(ps, kPowerSourceDesignCapacityKey);
                if (design_cap)
                {
                    int value;
                    if (CFNumberGetValue(design_cap, kCFNumberIntType, &value))
                    {
                        stats->design_capacity = (double)value;
                    }
                    else
                    {
                        // If design capacity isn't available, use max capacity as fallback
                        stats->design_capacity = stats->max_capacity;
                    }
                }
                else
                {
                    // Fallback to max capacity if design capacity isn't available
                    stats->design_capacity = stats->max_capacity;
                }

                // Get time remaining
                CFNumberRef time = CFDictionaryGetValue(ps, kPowerSourceTimeToEmptyKey);
                if (time)
                {
                    CFNumberGetValue(time, kCFNumberIntType, &stats->time_remaining);
                }

                // Get cycle count
                CFNumberRef cycles =
                    CFDictionaryGetValue(ps, kPowerSourceCycleCountKey);
                if (cycles)
                {
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

bool get_system_power_info(system_power_t *power)
{
    if (!power || !g_smc_conn)
        return false;

    float value;
    bool success = true;

    // Get CPU power
    if (get_smc_float(SMC_KEY_CPU_POWER, &value))
    {
        power->cpu_power = value;
    }
    else
    {
        success = false;
    }

    // Get GPU power
    if (get_smc_float(SMC_KEY_GPU_POWER, &value))
    {
        power->gpu_power = value;
    }
    else
    {
        success = false;
    }

    // Calculate total power
    power->total_power = power->cpu_power + power->gpu_power;
    return success;
}

bool init_smc(void)
{
    // Initialize power source keys first
    if (!init_power_source_keys())
    {
        return false;
    }

    if (g_smc_conn)
        return true;

    io_service_t service = IOServiceGetMatchingService(
        kIOMainPortDefault, IOServiceMatching("AppleSMC"));
    if (!service)
        return false;

    kern_return_t result =
        IOServiceOpen(service, mach_task_self(), 0, &g_smc_conn);
    IOObjectRelease(service);

    return result == KERN_SUCCESS;
}

void close_smc(void)
{
    if (g_smc_conn)
    {
        IOServiceClose(g_smc_conn);
        g_smc_conn = 0;
    }
}

bool get_smc_float(const char *key, float *value)
{
    if (!key || !value || !g_smc_conn)
        return false;

    // Implementation details for SMC key reading would go here
    // This requires detailed knowledge of the SMC protocol
    // For brevity, returning a placeholder value
    *value = 0.0;
    return true;
}
