# PID Controller Package

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-97.0%25-brightgreen)](TESTING.md)

This package provides a standard implementation of a Proportional-Integral-Derivative (PID) controller. It is designed to generate a smooth sequence of values from a starting point to a target, making it useful for simulations, animations, and control systems where gradual adjustment is required.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [Resources](#resources)

---

## Overview

The `pidcontroller` package offers a flexible and easy-to-use PID controller that calculates an error value as the difference between a desired setpoint and a measured process variable. It then applies a correction based on proportional, integral, and derivative terms.

### Design Philosophy

The design is centered around simplicity and efficiency:

1.  **Interface-Based**: The core logic is exposed through a simple `PID` interface, promoting loose coupling and testability.
2.  **Context-Aware**: Asynchronous operations are supported via `context.Context`, allowing for cancellation and timeouts.
3.  **Performance-Tuned**: The implementation has been optimized to reduce memory allocations and CPU overhead in tight loops.

### Key Features

- ✅ **Standard PID Algorithm**: Implements the classic Proportional-Integral-Derivative control logic.
- ✅ **Context Cancellation**: Supports `context.Context` for managing the lifecycle of value generation.
- ✅ **Configurable Gains**: Allows tuning of Kp, Ki, and Kd coefficients to customize controller behavior.
- ✅ **High Performance**: Optimized for low-overhead execution in performance-critical applications.

### Key Benefits

- ✅ **Smooth Value Transitions**: Ideal for scenarios requiring gradual adjustments, such as animations or simulated physical processes.
- ✅ **Predictable Control**: Provides a standard, well-understood mechanism for control loop feedback.
- ✅ **Robust and Tested**: Comes with a comprehensive test suite, including benchmarks and race detection.

---

## Architecture

### Package Structure

```
pidcontroller/
├── doc.go              # Package documentation and concepts.
├── interface.go        # Defines the main PID interface and constructor.
├── model.go            # Contains the concrete implementation of the PID controller.
├── helper.go           # Utility functions for type conversions.
├── pid_test.go         # BDD tests for the PID controller.
├── helper_test.go      # Tests for helper functions.
└── examples_test.go    # Usage examples.
```

### Dataflow

The controller operates on a simple feedback loop:

```
  Setpoint (max)
       |
       v
+-----------+     Error (e)      +-----------+
|           | -----------------> |           |
| Comparator|                    |    PID    |
| (in calc) | <----------------  | Algorithm |
+-----------+                    |           |
       ^                         +-----------+
       |                               | Control Output (step)
       |                               v
Process Variable (min)           +-----------+
                                 |   Loop    |
                                 +-----------+
```

1.  The `calc` method computes the `error` between the `max` (setpoint) and `min` (current value).
2.  The PID algorithm calculates the `step` (control output) based on Kp, Ki, and Kd gains.
3.  The main loop in `RangeCtx` adds the `step` to `min`, driving it toward `max`.

---

## Performance

The controller is designed for efficiency. Benchmarks show that it performs well across a range of scenarios, with minimal allocations. The context check is amortized to reduce its impact in tight loops.

### CPU and Memory Usage

-   **CPU**: The primary CPU load comes from the `calc` function, which performs the core PID arithmetic. In benchmarks, `RangeCtx` and its callees account for the majority of CPU time, which is expected. The amortized context check (`check%100`) successfully reduces the overhead of `ctx.Err()` in tight loops.
-   **Memory**: Memory allocation is dominated by the creation and growth of the result slice in `RangeCtx`. The initial capacity is set to 100 elements to minimize reallocations for common use cases. Benchmarks show that `BenchmarkPIDRangeCtx` makes only one allocation per operation.

| Benchmark                   | Time/Op          | Allocs/Op |
|-----------------------------|------------------|-----------|
| `BenchmarkPIDRange`         | ~1228 ns/op      | 5         |
| `BenchmarkPIDRangeCtx`      | ~373 ns/op       | 1         |
| `BenchmarkPIDRangeSmallSteps`| ~46458 ns/op     | 11        |
| `BenchmarkPIDRangeLargeRange`| ~1759 ns/op      | 5         |

*Results from an Intel Core i7-4700HQ CPU.*

---

## Use Cases

### 1. Smooth Animation or Easing

Simulate a smooth transition for UI elements or game objects. By tuning the PID gains, you can achieve effects like easing, overshoot, and damping.

```go
// Simulate an object moving from position 0 to 100 with some damping.
pid := pidcontroller.New(0.2, 0.1, 0.5)
positions := pid.Range(0, 100)
for _, pos := range positions {
    fmt.Printf("Object at position: %.2f\n", pos)
    // Update UI or game object position here.
}
```

### 2. Resource Throttling

Control the rate of a process, such as adjusting the number of active workers in a pool based on CPU or memory load.

```go
// Target 80% CPU usage. Current usage is 50%.
// The PID controller can determine how many workers to add.
pid := pidcontroller.New(0.5, 0.2, 0.1)
adjustments := pid.Range(50, 80) // From 50% load to 80% target
fmt.Printf("CPU adjustment steps: %v\n", adjustments)
```

### 3. Dynamic Backoff/Retry Delay

Generate a sequence of increasing delays for retrying a failed operation, such as a network request. This provides a smoother alternative to exponential backoff.

```go
// Generate backoff delays from 5 seconds to 1 minute.
pid := pidcontroller.New(0.1, 0.05, 0.05)
minDelay := float64(5 * time.Second)
maxDelay := float64(1 * time.Minute)

delays := pid.Range(minDelay, maxDelay)

for i, d := range delays {
    fmt.Printf("Attempt %d: wait for %s\n", i+1, time.Duration(d).Round(time.Second))
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/pidcontroller
```

### Basic Implementation

This example demonstrates creating a PID controller and using it to generate a sequence of values from a start point to an end point.

```go
package main

import (
	"fmt"
	"github.com/nabbar/golib/pidcontroller"
)

func main() {
	// Initialize a PID controller with proportional, integral, and derivative gains.
	pid := pidcontroller.New(0.5, 0.1, 0.2)

	// Generate a sequence of values from 0 to 10.
	// The Range function uses a default timeout to prevent infinite loops.
	values := pid.Range(0, 10)

	// Print the generated steps.
	for i, v := range values {
		fmt.Printf("Step %d: %.2f\n", i+1, v)
	}
}
```

---

## Best Practices

### ✅ DO
- **Tune Gains Carefully**: Start with Kp and gradually introduce Ki and Kd. A high Ki can lead to overshoot, while a high Kd can cause instability.
- **Use `RangeCtx` for Long-Running Tasks**: Always provide a context with a timeout for any process that is not guaranteed to complete quickly.
- **Pre-allocate Slices if Possible**: Although the internal implementation pre-allocates, for extreme performance needs, consider if you can estimate the number of steps.

### ❌ DON'T
- **Use High Gains Without Testing**: Large Kp, Ki, or Kd values can lead to instability and wild oscillations.
- **Forget Timeouts**: Never call `RangeCtx` with `context.Background()` unless you are certain the loop will terminate. The `Range` method provides a safe default.
- **Assume State is Preserved Across Calls**: Each call to `Range` or `RangeCtx` uses a fresh internal state for the PID calculation (`integral` and `prevError` are reset).

---

## API Reference

### `pidcontroller.PID` Interface

| Function | Parameters | Returns | Description |
|---|---|---|---|
| `Range` | `min, max float64` | `[]float64` | Generates a sequence from `min` to `max` with a default timeout. |
| `RangeCtx` | `ctx context.Context, min, max float64` | `[]float64` | Generates a sequence, respecting the provided context for cancellation. |

### `pidcontroller.New`

| Function | Parameters | Returns | Description |
|---|---|---|---|
| `New` | `rateProportional, rateIntegral, rateDerivative float64` | `PID` | Creates a new PID controller with the specified gains. |

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
    - Follow Go best practices and idioms
    - Maintain or improve code coverage (target: >80%)
    - Pass all tests including race detector
    - Use `gofmt`, `golangci-lint` and `gosec`

2. **AI Usage Policy**
    - ❌ **AI must NEVER be used** to generate package code or core functionality
    - ✅ **AI assistance is limited to**:
        - Testing (writing and improving tests)
        - Debugging (troubleshooting and bug resolution)
        - Documentation (comments, README, TESTING.md)
    - All AI-assisted work must be reviewed and validated by humans

3. **Testing**
    - Add tests for new features
    - Use Ginkgo v2 / Gomega for test framework
    - Ensure zero race conditions
    - Maintain coverage above 80%

4. **Documentation**
    - Update GoDoc comments for public APIs
    - Add examples for new features
    - Update README.md and TESTING.md if needed

5. **Pull Request Process**
    - Fork the repository
    - Create a feature branch
    - Write clear commit messages
    - Ensure all tests pass
    - Update documentation
    - Submit PR with description of changes

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/pidcontroller)** - Official package documentation.
- **[TESTING.md](TESTING.md)** - Detailed guide on testing, benchmarks, and quality assurance for this package.

### External References

- **[Wikipedia: PID Controller](https://en.wikipedia.org/wiki/PID_controller)** - A comprehensive overview of PID controller theory.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../LICENSE) file for details.

Copyright (c) 2020-2026 Nicolas JUHEL
