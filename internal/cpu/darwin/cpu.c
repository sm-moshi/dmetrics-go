#include "cpu.h"
#include <errno.h>
#include <mach/mach.h>
#include <mach/mach_host.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/sysctl.h>
#include <sys/types.h>
#include <unistd.h>

// Define CPU frequency sysctl constants if not available
#ifndef HW_CPU_FREQ
#define HW_CPU_FREQ 15
#endif

// Debug logging macro
#ifdef DEBUG
#define DEBUG_LOG(fmt, ...) fprintf(stderr, "DEBUG: " fmt "\n", ##__VA_ARGS__)
#else
#define DEBUG_LOG(fmt, ...)
#endif

// Error logging macro
#define ERROR_LOG(fmt, ...) fprintf(stderr, "ERROR: " fmt "\n", ##__VA_ARGS__)

// Static variables for caching
static processor_info_array_t prev_info_array = NULL;
static mach_msg_type_number_t prev_info_count;
static natural_t num_cpus;
static pthread_mutex_t cpu_mutex = PTHREAD_MUTEX_INITIALIZER;

int get_cpu_count(void) {
  int ncpu = 0;
  size_t len = sizeof(ncpu);
  if (sysctlbyname("hw.physicalcpu", &ncpu, &len, NULL, 0) == -1) {
    fprintf(stderr, "Error getting CPU count: %s\n", strerror(errno));
    return -1;
  }
  return ncpu;
}

uint64_t get_cpu_freq(void) {
  uint64_t freq = 0;
  size_t len = sizeof(freq);
  int error = 0;
  int mib[2];

  // Try direct sysctl for CPU frequency
  mib[0] = CTL_HW;
  mib[1] = HW_CPU_FREQ;
  error = sysctl(mib, 2, &freq, &len, NULL, 0);
  if (error == 0 && freq > 0 && freq < UINT64_MAX) {
    return freq / 1000000; // Convert to MHz
  }

  // Try getting current frequency
  if (sysctlbyname("hw.cpufrequency", &freq, &len, NULL, 0) == 0 && freq > 0) {
    return freq / 1000000;
  }

  // Try getting max frequency
  if (sysctlbyname("hw.cpufrequency_max", &freq, &len, NULL, 0) == 0 &&
      freq > 0) {
    return freq / 1000000;
  }

  // Try getting nominal frequency
  if (sysctlbyname("hw.cpufrequency_nominal", &freq, &len, NULL, 0) == 0 &&
      freq > 0) {
    return freq / 1000000;
  }

  // Try getting performance core frequency (Apple Silicon)
  if (sysctlbyname("hw.perflevel0.freq_hz", &freq, &len, NULL, 0) == 0 &&
      freq > 0) {
    return freq / 1000000;
  }

  // Try getting efficiency core frequency (Apple Silicon)
  if (sysctlbyname("hw.perflevel1.freq_hz", &freq, &len, NULL, 0) == 0 &&
      freq > 0) {
    return freq / 1000000;
  }

  // Try getting CPU speed (legacy method)
  if (sysctlbyname("hw.cpuspeed", &freq, &len, NULL, 0) == 0 && freq > 0) {
    return freq;
  }

  // Try getting CPU clock rate
  if (sysctlbyname("hw.clockrate", &freq, &len, NULL, 0) == 0 && freq > 0) {
    return freq;
  }

  // Try getting CPU frequency using sysctl machdep.tsc.frequency
  if (sysctlbyname("machdep.tsc.frequency", &freq, &len, NULL, 0) == 0 &&
      freq > 0) {
    return freq / 1000000;
  }

  // Try getting CPU frequency using sysctl hw.tbfrequency
  if (sysctlbyname("hw.tbfrequency", &freq, &len, NULL, 0) == 0 && freq > 0) {
    return freq / 1000000;
  }

  // Try getting CPU frequency using sysctl hw.busfrequency
  if (sysctlbyname("hw.busfrequency", &freq, &len, NULL, 0) == 0 && freq > 0) {
    return freq / 1000000;
  }

  // Only log error if all methods fail
  fprintf(stderr, "Warning: Failed to detect CPU frequency using any method\n");
  return 0;
}

uint64_t get_perf_core_freq(void) {
  uint64_t freq = 0;
  size_t len = sizeof(freq);

  // Try getting performance core frequency (P-cores)
  if (sysctlbyname("hw.perflevel0.freq_hz", &freq, &len, NULL, 0) == 0 &&
      freq > 0) {
    return freq / 1000000; // Convert to MHz
  }
  return 0;
}

uint64_t get_effi_core_freq(void) {
  uint64_t freq = 0;
  size_t len = sizeof(freq);

  // Try getting efficiency core frequency (E-cores)
  if (sysctlbyname("hw.perflevel1.freq_hz", &freq, &len, NULL, 0) == 0 &&
      freq > 0) {
    return freq / 1000000;
  }
  return 0;
}

int get_perf_core_count(void) {
  int count = 0;
  size_t len = sizeof(count);

  if (sysctlbyname("hw.perflevel0.logicalcpu", &count, &len, NULL, 0) == 0) {
    return count;
  }
  return 0;
}

int get_effi_core_count(void) {
  int count = 0;
  size_t len = sizeof(count);

  if (sysctlbyname("hw.perflevel1.logicalcpu", &count, &len, NULL, 0) == 0) {
    return count;
  }
  return 0;
}

int get_cpu_stats(cpu_stats_t *stats) {
  if (!stats)
    return CPU_ERROR_MEMORY;

  int ret = pthread_mutex_lock(&cpu_mutex);
  if (ret != 0) {
    ERROR_LOG("Failed to acquire mutex: %s", strerror(ret));
    return CPU_ERROR_MUTEX;
  }

  DEBUG_LOG("Collecting CPU stats");
  processor_info_array_t info_array;
  mach_msg_type_number_t info_count;

  kern_return_t error =
      host_processor_info(mach_host_self(), PROCESSOR_CPU_LOAD_INFO, &num_cpus,
                          &info_array, &info_count);

  if (error != KERN_SUCCESS) {
    ERROR_LOG("Failed to get processor info: %d", error);
    pthread_mutex_unlock(&cpu_mutex);
    return CPU_ERROR_HOST_PROCESSOR_INFO;
  }

  processor_cpu_load_info_t cpu_load_info =
      (processor_cpu_load_info_t)info_array;

  // If we don't have previous data, store it and wait
  if (!prev_info_array) {
    prev_info_array = info_array;
    prev_info_count = info_count;
    pthread_mutex_unlock(&cpu_mutex);
    usleep(500000); // 500ms for better sampling
    return get_cpu_stats(stats);
  }

  processor_cpu_load_info_t prev_cpu_load =
      (processor_cpu_load_info_t)prev_info_array;

  // Calculate per-core deltas
  for (natural_t i = 0; i < num_cpus; i++) {
    unsigned long long user = cpu_load_info[i].cpu_ticks[CPU_STATE_USER] -
                              prev_cpu_load[i].cpu_ticks[CPU_STATE_USER];
    unsigned long long system = cpu_load_info[i].cpu_ticks[CPU_STATE_SYSTEM] -
                                prev_cpu_load[i].cpu_ticks[CPU_STATE_SYSTEM];
    unsigned long long idle = cpu_load_info[i].cpu_ticks[CPU_STATE_IDLE] -
                              prev_cpu_load[i].cpu_ticks[CPU_STATE_IDLE];
    unsigned long long nice = cpu_load_info[i].cpu_ticks[CPU_STATE_NICE] -
                              prev_cpu_load[i].cpu_ticks[CPU_STATE_NICE];

    unsigned long long total = user + system + idle + nice;
    if (total > 0) {
      stats[i].user = (double)user / total * 100.0;
      stats[i].system = (double)system / total * 100.0;
      stats[i].idle = (double)idle / total * 100.0;
      stats[i].nice = (double)nice / total * 100.0;
    }
  }

  // Clean up and store current data for next time
  vm_deallocate(mach_task_self(), (vm_address_t)prev_info_array,
                prev_info_count);
  prev_info_array = info_array;
  prev_info_count = info_count;

  ret = pthread_mutex_unlock(&cpu_mutex);
  if (ret != 0) {
    ERROR_LOG("Failed to release mutex: %s", strerror(ret));
    return CPU_ERROR_MUTEX;
  }

  DEBUG_LOG("CPU stats collection completed successfully");
  return CPU_SUCCESS;
}

int get_load_avg(double loadavg[3]) {
  if (!loadavg)
    return CPU_ERROR_MEMORY;

  double sysloadavg[3];
  if (getloadavg(sysloadavg, 3) < 0) {
    return CPU_ERROR_SYSCTL;
  }

  loadavg[0] = sysloadavg[0];
  loadavg[1] = sysloadavg[1];
  loadavg[2] = sysloadavg[2];

  return CPU_SUCCESS;
}

int get_cpu_platform(cpu_platform_t *platform) {
  if (!platform)
    return CPU_ERROR_MEMORY;

  // Get CPU brand string
  size_t size = sizeof(platform->brand_string);
  if (sysctlbyname("machdep.cpu.brand_string", platform->brand_string, &size,
                   NULL, 0) < 0) {
    return CPU_ERROR_SYSCTL;
  }

  // Detect Apple Silicon
  platform->is_apple_silicon =
      (strstr(platform->brand_string, "Apple") != NULL);

  // Get actual CPU frequency
  platform->frequency = get_cpu_freq();
  // Don't treat 0 frequency as error since we have a default

  return CPU_SUCCESS;
}

void cleanup_cpu_stats(void) {
  pthread_mutex_lock(&cpu_mutex);
  if (prev_info_array) {
    vm_deallocate(mach_task_self(), (vm_address_t)prev_info_array,
                  prev_info_count);
    prev_info_array = NULL;
  }
  pthread_mutex_unlock(&cpu_mutex);
}

int get_cpu_core_stats(cpu_core_stats_t *stats, int *num_cores) {
  if (!stats || !num_cores)
    return CPU_ERROR_MEMORY;

  pthread_mutex_lock(&cpu_mutex);

  processor_info_array_t info_array;
  mach_msg_type_number_t info_count;

  kern_return_t error =
      host_processor_info(mach_host_self(), PROCESSOR_CPU_LOAD_INFO, &num_cpus,
                          &info_array, &info_count);

  if (error != KERN_SUCCESS) {
    pthread_mutex_unlock(&cpu_mutex);
    return CPU_ERROR_HOST_PROCESSOR_INFO;
  }

  processor_cpu_load_info_t cpu_load_info =
      (processor_cpu_load_info_t)info_array;

  // If we don't have previous data, store it and wait
  if (!prev_info_array) {
    prev_info_array = info_array;
    prev_info_count = info_count;
    pthread_mutex_unlock(&cpu_mutex);
    usleep(500000); // 500ms for better sampling
    return get_cpu_core_stats(stats, num_cores);
  }

  processor_cpu_load_info_t prev_cpu_load =
      (processor_cpu_load_info_t)prev_info_array;

  // Calculate per-core deltas and percentages
  for (natural_t i = 0; i < num_cpus; i++) {
    unsigned long long user = cpu_load_info[i].cpu_ticks[CPU_STATE_USER] -
                              prev_cpu_load[i].cpu_ticks[CPU_STATE_USER];
    unsigned long long system = cpu_load_info[i].cpu_ticks[CPU_STATE_SYSTEM] -
                                prev_cpu_load[i].cpu_ticks[CPU_STATE_SYSTEM];
    unsigned long long idle = cpu_load_info[i].cpu_ticks[CPU_STATE_IDLE] -
                              prev_cpu_load[i].cpu_ticks[CPU_STATE_IDLE];
    unsigned long long nice = cpu_load_info[i].cpu_ticks[CPU_STATE_NICE] -
                              prev_cpu_load[i].cpu_ticks[CPU_STATE_NICE];

    unsigned long long total = user + system + idle + nice;
    if (total > 0) {
      // Store individual percentages for each state
      stats[i].user = (double)user / total * 100.0;
      stats[i].system = (double)system / total * 100.0;
      stats[i].idle = (double)idle / total * 100.0;
      stats[i].nice = (double)nice / total * 100.0;
    } else {
      // If no ticks recorded, assume idle
      stats[i].user = 0;
      stats[i].system = 0;
      stats[i].idle = 100.0;
      stats[i].nice = 0;
    }
  }

  *num_cores = num_cpus;

  // Clean up old info array
  vm_deallocate(mach_task_self(), (vm_address_t)prev_info_array,
                prev_info_count * sizeof(natural_t));

  // Store current info array for next time
  prev_info_array = info_array;
  prev_info_count = info_count;

  pthread_mutex_unlock(&cpu_mutex);
  return CPU_SUCCESS;
}

int get_per_core_cpu_stats(processor_cpu_load_info_t *cpu_load_info,
                           natural_t *cpu_count) {
  processor_info_array_t info_array;
  mach_msg_type_number_t info_count;

  kern_return_t error =
      host_processor_info(mach_host_self(), PROCESSOR_CPU_LOAD_INFO, cpu_count,
                          &info_array, &info_count);

  if (error != KERN_SUCCESS) {
    return CPU_ERROR_HOST_PROCESSOR_INFO;
  }

  *cpu_load_info = (processor_cpu_load_info_t)info_array;
  return CPU_SUCCESS;
}

void init_cpu_stats(void) {
  // Initialize mutex if needed
  pthread_mutex_init(&cpu_mutex, NULL);

  // Ensure prev_info_array is NULL
  prev_info_array = NULL;
  prev_info_count = 0;
  num_cpus = 0;
}