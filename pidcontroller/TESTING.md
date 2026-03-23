# Testing Guide

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-97.0%25-brightgreen)](TESTING.md)
[![Test Specs](https://img.shields.io/badge/Tests%20Specs-18-green)]()
[![Test Asserts](https://img.shields.io/badge/Tests%20Asserts-~20-green)]()

Comprehensive testing documentation for the `pidcontroller` package, covering test execution, race detection, benchmarks, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Test Architecture](#test-architecture)
- [Framework & Tools](#framework--tools)
- [Quick Start](#quick-start)
- [Performance & Profiling](#performance--profiling)
- [Test Coverage](#test-coverage)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Reporting Bugs & Vulnerabilities](#reporting-bugs--vulnerabilities)
- [Resources](#resources)

---

## Overview

The `pidcontroller` package employs a BDD (Behavior-Driven Development) approach using the Ginkgo framework to ensure the correctness and reliability of the PID control logic. The test suite covers basic functionality, edge cases, context handling, and performance characteristics.

**Test Suite Summary**
- Total Specs: 18 across the package
- Total Assertions: ~20
- Overall Coverage: 97.0% (statements)
- Race Detection: ✅ Zero data races
- Execution Time: < 1s (approx.)

### Test Plan

This test suite provides a robust validation of the PID controller:

1.  **Functional Testing**: Verifies that the PID controller correctly generates a sequence of values transitioning from a start to an end point.
2.  **Context Validation**: Ensures that `context.Context` cancellation and timeouts are respected and handled gracefully.
3.  **Edge Case Testing**: Validates behavior for boundary conditions such as `min > max`, `min == max`, negative ranges, and extremely large numbers.
4.  **Helper Testing**: Checks the correctness of utility functions for type conversion and overflow handling.
5.  **Benchmarks**: Measures the CPU time and memory allocation of the PID generation loop under various conditions (small steps, large ranges, timeouts).

### Test Completeness

**Quality Indicators:**
- **Code Coverage**: 97.0% of statements (Target: >80%)
- **Race Conditions**: 0 detected across all scenarios
- **Flakiness**: 0 flaky tests detected

**Test Distribution:**
- ✅ **18 specifications** covering functionality and edge cases.
- ✅ **3 example tests** demonstrating usage patterns in `examples_test.go`.
- ✅ **6 performance benchmarks** measuring key metrics.
- ✅ **2 test files** (plus suite setup) organized by functional area.

---

## Test Architecture

### Test Matrix

The tests are organized to cover critical paths and potential failure modes.

| Category | Target | Specs | Priority | Dependencies |
|----------|--------|-------|----------|--------------|
| **Basic Tests** | Validate `New` constructor and basic `Range` functionality. | 4 | Critical | None |
| **Context Tests** | Verify `RangeCtx` handles cancellation and timeouts correctly. | 2 | Critical | `context` package |
| **Edge Cases** | Check boundary conditions (`min>max`, `min=max`, negative, large values). | 4 | High | None |
| **Helpers** | Unit tests for `Int64ToFloat64` and `Float64ToInt64` conversion logic. | 8 | Medium | None |
| **Examples** | Documentable examples for `Range` and `RangeCtx`. | 3 | High | None |

### Detailed Test Inventory

**Test ID Pattern by File:**
- **TC-PID-xxx**: PID Logic Tests (`pid_test.go`)
- **TC-HLP-xxx**: Helper Tests (`helper_test.go`)

| Test ID | File | Use Case | Priority | Expected Outcome |
|---|---|---|---|---|
| **TC-PID-001** | `pid_test.go` | Constructor `New` | Critical | Returns a valid non-nil interface. |
| **TC-PID-002** | `pid_test.go` | `Range` end value | Critical | Sequence ends exactly at `max`. |
| **TC-PID-003** | `pid_test.go` | `Range` progression | High | Values progress from `min` towards `max`. |
| **TC-PID-004** | `pid_test.go` | `Range` small steps | Medium | Handles small increments correctly. |
| **TC-PID-005** | `pid_test.go` | `RangeCtx` cancellation | Critical | Returns partial result + max on cancel. |
| **TC-PID-006** | `pid_test.go` | `RangeCtx` timeout | Critical | Returns partial result + max on timeout. |
| **TC-PID-007** | `pid_test.go` | Edge `min > max` | High | Returns `[max]` immediately. |
| **TC-PID-008** | `pid_test.go` | Edge `min == max` | High | Returns `[max]` immediately. |
| **TC-PID-009** | `pid_test.go` | Negative range | Medium | Correctly handles negative start/end values. |
| **TC-PID-010** | `pid_test.go` | Large range | Medium | Completes without hanging on large float values. |
| **TC-HLP-001** | `helper_test.go` | `Int64ToFloat64` | Low | Correct conversion. |
| **TC-HLP-002** | `helper_test.go` | `Float64ToInt64` | Low | Correct conversion and clamping. |

---

## Framework & Tools

### Testing Frameworks

#### Ginkgo v2 - BDD Testing Framework

- **Hierarchical organization**: Used `Describe`, `Context`, `It` to structure tests logically around methods and behaviors.
- **Lifecycle hooks**: `BeforeEach` is used to initialize a fresh PID controller for every test spec.

#### Gomega - Matcher Library

- **Expressive assertions**: `Expect(res).To(HaveLen(1))` and `Expect(res).To(BeNumerically(">", min))` ensure clear failure messages.

### Testing Concepts & Standards

#### ISTQB Alignment

1.  **Unit Testing**: Individual functions (`Int64ToFloat64`) are tested in isolation.
2.  **Component Testing**: The PID controller logic is tested as a cohesive unit via its interface.
3.  **Boundary Value Analysis**: Tests explicitly cover boundaries like `min=max` and integer overflows.

---

## Quick Start

### Installation

```bash
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

### Running Tests

```bash
# Run all tests
go test -v ./...
```

```bash
# With race detection
CGO_ENABLED=1 go test -v -race ./...
```

```bash
# Benchmarks
go test -v -bench=. -benchmem ./...
```

---

### Performance & Profiling

Benchmarks are implemented in `pid_benchmark_test.go`.

#### Performance Benchmarks

Command to launch:
```bash
go test -bench=. -benchmem ./...
```

**Results (Intel Core i7-4700HQ):**
| Benchmark | Time/Op | Allocs/Op |
|---|---|---|
| `BenchmarkPIDRange` | ~1228 ns/op | 5 |
| `BenchmarkPIDRangeCtx` | ~373 ns/op | 1 |
| `BenchmarkPIDRangeSmallSteps`| ~46458 ns/op | 11 |
| `BenchmarkPIDRangeLargeRange`| ~1759 ns/op | 5 |

#### CPU Profiling

Command to launch:
```bash
go test -bench=. -cpuprofile=cpu.out ./...
go tool pprof cpu.out
```

The CPU profile confirms that the most time-consuming functions are `RangeCtx` and `calc`, which is expected. The amortized context check (`check%100`) effectively minimizes the overhead of `ctx.Err()` in tight loops.

#### Memory Profiling

Command to launch:
```bash
go test -bench=. -memprofile=mem.out ./...
go tool pprof mem.out
```

The memory profile shows that `RangeCtx` is the primary allocator, mainly due to the creation and expansion of the result slice. The initial capacity of 100 helps reduce reallocations, and most benchmarks show a low number of allocations per operation.

---

## Test Coverage

**Current Status**: 97.0% statement coverage.

### Coverage By File

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Output**:
```
github.com/nabbar/golib/pidcontroller/helper.go:31:    Int64ToFloat64  100.0%
github.com/nabbar/golib/pidcontroller/helper.go:37:    Float64ToInt64  100.0%
github.com/nabbar/golib/pidcontroller/interface.go:81: New             100.0%
github.com/nabbar/golib/pidcontroller/model.go:63:     calc            100.0%
github.com/nabbar/golib/pidcontroller/model.go:96:     RangeCtx        94.1%
github.com/nabbar/golib/pidcontroller/model.go:154:    Range           100.0%
```

*Note: The slightly lower coverage in `RangeCtx` is due to amortized context checks which are statistically hard to hit in deterministic unit tests.*

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should return partial results on timeout", func() { ... })
```

**2. Follow AAA Pattern**
```go
It("should ...", func() {
    // Arrange
    min, max := 0.0, 10.0
    
    // Act
    res := pid.Range(min, max)
    
    // Assert
    Expect(res).To(HaveLen(1))
})
```

---

## Best Practices

### ✅ Good Practices
- **Test Independence**: Each test spec uses a fresh PID controller instance.
- **Context Usage**: Tests involving `RangeCtx` explicitly manage context cancellation/timeout.
- **Edge Case Coverage**: Boundaries are explicitly tested to prevent off-by-one errors or infinite loops.

### ❌ Avoid
- **Hardcoded Sleeps**: Used `context.WithTimeout` instead of sleeping in tests.
- **Global State**: No package-level variables are modified during tests.

---

## Reporting Bugs & Vulnerabilities

Please refer to the issue template in the repository for reporting bugs. Ensure to include:
- `go version`
- `go env`
- A minimal reproduction case
- Stack trace if applicable

---

## Resources

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for test generation, debugging, and documentation under human supervision. All tests are validated and reviewed by humans.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2020-2026 Nicolas JUHEL
