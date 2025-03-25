# TODO List

## Version 0.1 (Current Release)

### Completed Tasks ✅

- [x] Implement thread-safe concurrent access
  - Added mutex protection for shared resources
  - Implemented atomic operations where appropriate
  - Verified thread safety with race detector
  - Added documentation for concurrency guarantees

- [x] Add proper context handling for cancellation
  - Implemented context.Context support in all APIs
  - Added timeout and cancellation handling
  - Ensured proper resource cleanup on cancellation
  - Added examples demonstrating context usage

- [x] Fix per-core CPU utilisation calculation
  - Corrected calculation methodology
  - Added validation against system tools
  - Implemented proper tick delta tracking
  - Added safeguards against overflow

- [x] Optimise memory management
  - Implemented proper C memory cleanup
  - Added finalizers for CGo resources
  - Reduced allocations in hot paths
  - Added memory usage documentation

- [x] Add comprehensive documentation
  - Added package overview and examples
  - Documented all exported symbols
  - Added performance characteristics
  - Included British English conventions

- [x] Implement proper error handling
  - Added domain-specific error types
  - Improved error messages
  - Added error wrapping
  - Documented error conditions

- [x] Add performance characterisation
  - Documented initial collection time
  - Measured subsequent call latency
  - Analysed memory usage patterns
  - Added benchmarks

- [x] Add test coverage for core functionality
  - Unit tests for all exported APIs
  - Integration tests for system calls
  - Benchmark suite
  - Race condition tests

### In Progress 🚧

- [ ] Fix CPU frequency detection
  - Current Status: Returns 0 for all cores
  - Investigation Areas:
    - sysctl calls on Intel processors
    - IOKit API usage for Apple Silicon
    - Frequency scaling support
  - Acceptance Criteria:
    - Accurate base frequency reporting
    - Dynamic frequency updates
    - Support for both architectures

- [ ] Optimise integration test complexity
  - Current Issues:
    - Long-running test suites
    - Flaky timing-dependent tests
    - Complex setup/teardown
  - Planned Improvements:
    - Refactor to use test helpers
    - Add proper mocking
    - Implement parallel testing
    - Reduce timing dependencies

### Power Metrics (Required for v0.1) 🔋

- [ ] Power source detection
  - Requirements:
    - AC adapter detection and status
    - Battery presence detection
    - Power source transition events
    - Multiple power source support
  - Implementation Notes:
    - Use IOPowerSources API
    - Implement power source change callbacks
    - Add power source type identification
    - Handle system sleep/wake events

- [ ] Battery metrics implementation
  - Requirements:
    - Current charge level
    - Time remaining estimation
    - Charge cycle count
    - Battery health status
  - Implementation Notes:
    - Use IOKit for battery info
    - Implement charge level monitoring
    - Add battery health diagnostics
    - Handle multiple battery configurations

- [ ] Power state monitoring
  - Requirements:
    - System power state
    - Sleep/wake detection
    - Power mode changes
    - Thermal state monitoring
  - Implementation Notes:
    - Implement IOKit power callbacks
    - Add power state change notifications
    - Monitor thermal conditions
    - Track power mode transitions

## Version 0.2 (Next Release)

### Memory Metrics 📊

- [ ] Memory utilisation tracking
  - Requirements:
    - Total system memory monitoring
    - Process-specific memory usage
    - Virtual memory statistics
    - Memory pressure indicators
  - Implementation Notes:
    - Use host_statistics64() for system metrics
    - Implement process_info() for per-process stats
    - Add memory pressure callbacks

- [ ] Swap usage monitoring
  - Features:
    - Swap space utilisation
    - Page-in/out rates
    - Swap file locations
    - Compression statistics

- [ ] Page fault statistics
  - Metrics to Track:
    - Major page faults
    - Minor page faults
    - Copy-on-write faults
    - Zero-fill page faults

- [ ] Memory pressure indicators
  - Implementation:
    - Memory pressure level detection
    - Pressure history tracking
    - Warning threshold configuration
    - Callback registration

### GPU Metrics 🎮

- [ ] GPU utilisation tracking
  - Requirements:
    - Support for integrated GPUs
    - Dedicated GPU monitoring
    - Multi-GPU systems
    - Metal API integration

- [ ] VRAM usage monitoring
  - Features:
    - Total VRAM capacity
    - Current utilisation
    - Per-process VRAM usage
    - Texture memory tracking

- [ ] Temperature monitoring
  - Implementation:
    - SMC interface integration
    - Temperature sensor mapping
    - Thermal zone monitoring
    - Warning threshold support

- [ ] Power consumption tracking
  - Metrics:
    - Current power draw
    - Average consumption
    - Peak usage tracking
    - Power state transitions

### Temperature Sensors 🌡️

- [ ] CPU temperature monitoring
  - Features:
    - Per-core temperature
    - Package temperature
    - Thermal throttling detection
    - Historical tracking

- [ ] GPU temperature monitoring
  - Implementation:
    - IOKit integration
    - Thermal sensor access
    - Multiple GPU support
    - Thermal zone mapping

- [ ] System temperature sensors
  - Coverage:
    - Ambient temperature
    - Logic board sensors
    - Power supply temperature
    - Custom sensor support

- [ ] Fan speed correlation
  - Features:
    - Temperature/fan speed correlation
    - Thermal zone mapping
    - Cooling system analysis
    - Thermal profile creation

### Fan Speed Metrics 💨

- [ ] System fan speed monitoring
  - Requirements:
    - Current RPM readings
    - Maximum/minimum speeds
    - Fan location mapping
    - Speed percentage calculation

- [ ] CPU fan speed tracking
  - Features:
    - Dynamic speed monitoring
    - Thermal correlation
    - Speed curve analysis
    - Historical tracking

- [ ] GPU fan speed tracking
  - Implementation:
    - IOKit integration
    - Multiple GPU support
    - Fan curve mapping
    - Speed normalisation

- [ ] Custom fan curve support
  - Features:
    - Fan curve definition
    - Temperature thresholds
    - Hysteresis support
    - Profile management

### Power Source Info 🔋

- [ ] Battery level monitoring
  - Requirements:
    - Current charge level
    - Time remaining estimate
    - Charge cycle count
    - Battery health metrics

- [ ] Power source detection
  - Features:
    - AC/Battery detection
    - Power adapter information
    - Multiple power source support
    - Source transition events

- [ ] Charging status tracking
  - Implementation:
    - Charge rate monitoring
    - Temperature influence
    - Power delivery analysis
    - Status change notifications

- [ ] Power consumption metrics
  - Metrics:
    - Current power draw
    - Energy impact scoring
    - Power nap statistics
    - Peak usage tracking

### Future Considerations 🔮

- [ ] Network Interface Metrics
- [ ] Disk I/O Statistics
- [ ] Process Resource Usage
- [ ] System Load Metrics
- [ ] Hardware Sensor Integration
- [ ] Energy Impact Analysis
- [ ] Thermal Management
- [ ] Resource Pressure Monitoring
