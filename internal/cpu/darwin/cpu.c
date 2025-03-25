#include "cpu.h"
#include <errno.h>
#include <mach/mach.h>
#include <mach/mach_host.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/sysctl.h>
#include <unistd.h>

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

  // Try getting current frequency
  if (sysctlbyname("hw.cpufrequency", &freq, &len, NULL, 0) == 0) {
    return freq / 1000000; // Convert to MHz
  }

  // Fallback to max frequency
  if (sysctlbyname("hw.cpufrequency_max", &freq, &len, NULL, 0) == 0) {
    return freq / 1000000;
  }

  // Return 0 to indicate error
  return 0;
}

uint64_t get_perf_core_freq(void) {
  uint64_t freq = 0;
  size_t len = sizeof(freq);

  // Try getting performance core frequency (P-cores)
  if (sysctlbyname("hw.perflevel0.freq_hz", &freq, &len, NULL, 0) == 0) {
    return freq / 1000000; // Convert to MHz
  }
  return 0;
}

uint64_t get_effi_core_freq(void) {
  uint64_t freq = 0;
  size_t len = sizeof(freq);

  // Try getting efficiency core frequency (E-cores)
  if (sysctlbyname("hw.perflevel1.freq_hz", &freq, &len, NULL, 0) == 0) {
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

  pthread_mutex_unlock(&cpu_mutex);
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
  processor_info_array_t info_array;
  mach_msg_type_number_t info_count;
  natural_t cpu_count;

  kern_return_t error =
      host_processor_info(mach_host_self(), PROCESSOR_CPU_LOAD_INFO, &cpu_count,
                          &info_array, &info_count);

  if (error != KERN_SUCCESS) {
    return CPU_ERROR_HOST_PROCESSOR_INFO;
  }

  processor_cpu_load_info_t cpu_load_info =
      (processor_cpu_load_info_t)info_array;
  *num_cores = cpu_count;

  for (natural_t i = 0; i < cpu_count; i++) {
    stats[i].core_id = i;
    stats[i].user = cpu_load_info[i].cpu_ticks[CPU_STATE_USER];
    stats[i].system = cpu_load_info[i].cpu_ticks[CPU_STATE_SYSTEM];
    stats[i].idle = cpu_load_info[i].cpu_ticks[CPU_STATE_IDLE];
    stats[i].nice = cpu_load_info[i].cpu_ticks[CPU_STATE_NICE];
  }

  vm_deallocate(mach_task_self(), (vm_address_t)info_array, info_count);
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