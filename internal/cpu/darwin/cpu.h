#ifndef CPU_H
#define CPU_H

#include <mach/mach_host.h>
#include <mach/processor_info.h>
#include <stdint.h>
#include <sys/sysctl.h>

// Error codes
#define CPU_SUCCESS 0
#define CPU_ERROR_MEMORY -1
#define CPU_ERROR_SYSCTL -2
#define CPU_ERROR_HOST_PROCESSOR_INFO -3
#define CPU_ERROR_MUTEX -4 // Added for mutex operation failures

// CPU statistics structure
typedef struct {
  double user;
  double system;
  double idle;
  double nice;
} cpu_stats_t;

// Platform information structure
typedef struct {
  int is_apple_silicon;
  char brand_string[128];
  uint64_t frequency; // Base frequency in MHz
  uint64_t perf_freq; // Performance core frequency in MHz
  uint64_t effi_freq; // Efficiency core frequency in MHz
  int perf_cores;     // Number of performance cores
  int effi_cores;     // Number of efficiency cores
} cpu_platform_t;

// Core statistics structure
typedef struct {
  double user;
  double system;
  double idle;
  double nice;
  int core_id;
} cpu_core_stats_t;

// Function declarations
int get_cpu_count(void);
uint64_t get_cpu_freq(void);
uint64_t get_perf_core_freq(void);
uint64_t get_effi_core_freq(void);
int get_perf_core_count(void);
int get_effi_core_count(void);
int get_cpu_stats(cpu_stats_t *stats);
int get_cpu_platform(cpu_platform_t *platform);
int get_load_avg(double loadavg[3]);
void init_cpu_stats(void);
void cleanup_cpu_stats(void);
int get_cpu_core_stats(cpu_core_stats_t *stats, int *num_cores);
int get_per_core_cpu_stats(processor_cpu_load_info_t *cpu_load_info,
                           natural_t *cpu_count);

#endif // CPU_H