# Testing Guide

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-91.0%25-brightgreen)](TESTING.md)
[![Test Specs](https://img.shields.io/badge/Tests%20Specs-106-green)]()
[![Test Asserts](https://img.shields.io/badge/Tests%20Asserts-~300-green)]()

Comprehensive testing documentation for the `info` package, covering test execution, race detection, benchmarks, and quality assurance.

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

The `info` package employs a Behavior-Driven Development (BDD) testing strategy using the Ginkgo framework. This ensures that tests serve as both verification and live documentation of the package's behavior.

**Test Suite Summary**
- Total Specs: 106 across 7 test files
- Overall Coverage: 91.0% of statements
- Race Detection: ✅ Zero data races
- Execution Time: ~0.12s (standard), ~1.2s (with race detection)

### Test Plan

This test suite provides complete coverage of the `info` package's lifecycle:

1.  **Functional Testing**: Verification of core API (`New`, `Name`, `Info`).
2.  **Manual Manipulation**: Validation of `SetData`, `AddData`, and override mechanisms.
3.  **Concurrency Testing**: Stress testing with parallel reads/writes to ensure thread safety.
4.  **Edge Cases**: Testing nil receivers, empty inputs, and error conditions.
5.  **Integration Testing**: Simulating real-world usage patterns (JSON marshaling, service discovery).
6.  **Internal Testing**: Verification of unexported methods and internal state consistency.

### Test Completeness

**Quality Indicators:**
- **Code Coverage**: 91.0% statements (Target: >85%)
- **Race Conditions**: 0 detected
- **Flakiness**: 0 flaky tests detected

**Test Distribution:**
- ✅ **106 specifications** covering all features
- ✅ **~300 assertions** validating precise behaviors
- ✅ **5+ example tests** demonstrating usage
- ✅ **Zero flaky tests** - deterministic execution

---

## Test Architecture

The tests are organized by functional area to maintain clarity and separation of concerns.

### Test Matrix

| Category | Target | Specs | Priority | Dependencies |
|----------|--------|-------|----------|--------------|
| **Core API** | `info.go`, `model.go` - Basic lifecycle | 16 | Critical | None |
| **Manual Data** | `manual_test.go` - `Set/Add/Del` operations | 19 | High | Core API |
| **Concurrency** | `info_test.go` - Parallel access safety | 5 | Critical | Core API |
| **Edge Cases** | `edge_cases_test.go` - Error handling, limits | 15 | Medium | Core API |
| **Integration** | `integration_test.go` - JSON/Text marshaling | 13 | Medium | Core API |
| **Internal** | `internal_test.go` - Nil receiver checks | ~10 | Low | None |

### Detailed Test Inventory

**Test ID Pattern:**
- **TC-CORE-xxx**: Core functionality (`info_test.go`)
- **TC-MAN-xxx**: Manual manipulation (`manual_test.go`)
- **TC-CONC-xxx**: Concurrency (`info_test.go`)
- **TC-EDGE-xxx**: Edge cases (`edge_cases_test.go`)
- **TC-INT-xxx**: Integration (`integration_test.go`)

| Test ID | File | Use Case | Priority | Expected Outcome |
|---------|------|----------|----------|------------------|
| **TC-CORE-001** | `info_test.go` | Constructor Validation | Critical | `New()` returns valid instance or error on empty name |
| **TC-CORE-002** | `info_test.go` | Name Retrieval | High | `Name()` returns registered or default name |
| **TC-CORE-003** | `info_test.go` | Empty Default Name | Critical | Error returned when default name is empty |
| **TC-CORE-004** | `info_test.go` | Name Default Fallback | High | Default name returned when no function registered |
| **TC-CORE-005** | `info_test.go` | Dynamic Name | High | Function result returned |
| **TC-CORE-006** | `info_test.go` | Name Function Error | Medium | Default name returned on function error |
| **TC-CORE-007** | `info_test.go` | Name Caching | Critical | Function called on every access (no caching) |
| **TC-CORE-008** | `info_test.go` | Name Re-registration | High | New function used immediately |
| **TC-CORE-009** | `info_test.go` | Info Default | High | Empty map returned when no function registered |
| **TC-CORE-010** | `info_test.go` | Dynamic Info | High | Function result returned |
| **TC-CORE-011** | `info_test.go` | Info Function Error | Medium | Empty map returned on function error |
| **TC-CORE-012** | `info_test.go` | Info Caching | Critical | Function called on every access (no caching) |
| **TC-CORE-013** | `info_test.go` | Info Re-registration | High | New function used immediately |
| **TC-CORE-014** | `info_test.go` | Empty Map Return | Medium | Empty map returned (not nil) |
| **TC-CORE-015** | `info_test.go` | Data Types | Medium | All Go types preserved in map |
| **TC-CORE-016** | `info_test.go` | Interface Compliance | Low | Implements `montps.Info` |
| **TC-CONC-001** | `info_test.go` | Concurrent Name Reg | Critical | No panic/race |
| **TC-CONC-002** | `info_test.go` | Concurrent Info Reg | Critical | No panic/race |
| **TC-CONC-003** | `info_test.go` | Concurrent Name Read | Critical | No panic/race |
| **TC-CONC-004** | `info_test.go` | Concurrent Info Read | Critical | No panic/race |
| **TC-CONC-005** | `info_test.go` | Mixed Concurrency | Critical | No panic/race |
| **TC-MAN-001** | `manual_test.go` | SetName Override | Medium | `SetName()` overrides default |
| **TC-MAN-002** | `manual_test.go` | SetName Empty | Medium | Revert to default name |
| **TC-MAN-003** | `manual_test.go` | SetData Replace | High | Replaces all data |
| **TC-MAN-004** | `manual_test.go` | SetData Nil | High | Clears all data |
| **TC-MAN-005** | `manual_test.go` | AddData Update | Medium | Adds/updates keys |
| **TC-MAN-006** | `manual_test.go` | AddData Nil | Medium | Deletes key |
| **TC-MAN-007** | `manual_test.go` | AddData Empty Key | Low | Ignored |
| **TC-MAN-008** | `manual_test.go` | DelData | Medium | Deletes specific key |
| **TC-MAN-009** | `manual_test.go` | DelData Empty Key | Low | Ignored |
| **TC-MAN-010** | `manual_test.go` | Unregister Name | High | Unregisters function via nil |
| **TC-MAN-011** | `manual_test.go` | Unregister Info | High | Unregisters function via nil |
| **TC-MAN-012..019**| `manual_test.go` | Nil Receivers | Medium | No panic on nil instance calls |
| **TC-EDGE-001** | `edge_cases_test.go` | Multiple Re-reg | Low | Correct state after multiple registrations |
| **TC-EDGE-002** | `edge_cases_test.go` | Info Re-reg | Low | Correct state after multiple registrations |
| **TC-EDGE-003** | `edge_cases_test.go` | Name Error Recovery | Medium | Recovers after function error |
| **TC-EDGE-004** | `edge_cases_test.go` | Info Error Recovery | Medium | Recovers after function error |
| **TC-EDGE-005** | `edge_cases_test.go` | Nil Map Return | Low | Handled gracefully |
| **TC-EDGE-006** | `edge_cases_test.go` | Sync Map Internal | Low | Internal key handling |
| **TC-EDGE-007** | `edge_cases_test.go` | Interleaved Ops | Medium | Correctness during interleaved calls |
| **TC-EDGE-008** | `edge_cases_test.go` | Empty String Name | Low | Handled as error/fallback |
| **TC-EDGE-009** | `edge_cases_test.go` | Whitespace Name | Low | Preserved |
| **TC-EDGE-010** | `edge_cases_test.go` | Complex Types | Low | Interface{} handling |
| **TC-EDGE-011** | `edge_cases_test.go` | Empty Info Text | Low | Text marshaling with empty info |
| **TC-EDGE-012** | `edge_cases_test.go` | Unicode Name | Low | UTF-8 support |
| **TC-EDGE-013** | `edge_cases_test.go` | Unicode Info | Low | UTF-8 support |
| **TC-EDGE-014** | `edge_cases_test.go` | Long Name | Low | Buffer handling |
| **TC-EDGE-015** | `edge_cases_test.go` | Long Info | Low | Buffer handling |
| **TC-INT-001** | `integration_test.go` | Text Marshaler | High | Interface implementation |
| **TC-INT-002** | `integration_test.go` | JSON Marshaler | High | Interface implementation |
| **TC-INT-003** | `integration_test.go` | JSON Structure | High | Correct JSON output |
| **TC-INT-004** | `integration_test.go` | Complex JSON | Medium | Nested structure support |
| **TC-INT-005** | `integration_test.go` | Service Discovery | Medium | Use case simulation |
| **TC-INT-006** | `integration_test.go` | Runtime Metrics | Medium | Use case simulation |
| **TC-INT-007** | `integration_test.go` | Env Info | Medium | Use case simulation |
| **TC-INT-008** | `integration_test.go` | Health Check | Medium | Use case simulation |
| **TC-INT-009** | `integration_test.go` | Config Info | Medium | Use case simulation |
| **TC-INT-010** | `integration_test.go` | Transient Errors | Medium | Retry logic |
| **TC-INT-011** | `integration_test.go` | Partial Data | Low | Robustness |
| **TC-INT-012** | `integration_test.go` | Caching Perf | High | Performance verification |
| **TC-INT-013** | `integration_test.go` | Invalidation | High | Cache clearing verification |

---

## Framework & Tools

### Testing Frameworks

#### Ginkgo v2 - BDD Testing Framework

Used for its expressive DSL and structural capabilities:
- **Hierarchical Contexts**: `Describe("Feature")` -> `Context("Scenario")` -> `It("Behavior")`
- **Parallel Execution**: Supports concurrent test runs for faster feedback.

#### Gomega - Matcher Library

Used for robust assertions:
- `Expect(val).To(Equal(expected))`
- `Expect(err).NotTo(HaveOccurred())`
- `Expect(map).To(HaveKey("version"))`

### Testing Standards

This suite aligns with **ISTQB** principles:
1.  **Unit Testing**: Isolated tests for individual methods (`manual_test.go`).
2.  **Integration Testing**: Verification of component interactions (`integration_test.go`).
3.  **Boundary Value Analysis**: Testing empty strings, nil maps, etc. (`edge_cases_test.go`).

---

## Quick Start

### Installation

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

### Running Tests

```bash
# Run all tests
go test -v ./...
```

```bash
# With coverage report
go test -v -cover ./...
```

```bash
# With race detection (Highly Recommended)
CGO_ENABLED=1 go test -v -race ./...
```

---

## Performance & Profiling

Benchmarks are executed to ensure the lock-free atomic implementation maintains high throughput.

### Hardware Environment
Tests and benchmarks are executed on:
- **Product**: ASUS G750JH Notebook
- **CPU**: Intel(R) Core(TM) i7-4700HQ CPU @ 2.40GHz (4 cores, 8 threads)
- **Memory**: 32GiB DDR3 1333 MHz
- **OS**: Linux 6.8.0

### Performance CPU

Performance benchmarks focus on the overhead of the atomic operations.

Command to launch:
```bash
# Benchmarks
go test -bench=. ./...
```

**Results (Core i7-4700HQ):**
| Benchmark | Ops/sec | ns/op | Alloc/op |
|-----------|---------|-------|----------|
| `BenchmarkConcurrentNameReads` | ~100M | ~19 ns | 0 B |
| `BenchmarkNameWithFunction` | ~29M | ~39 ns | 0 B |
| `BenchmarkConcurrentInfoReads` | ~2.4M | ~495 ns | 736 B (6 allocs) |
| `BenchmarkInfoWithFunction` | ~1.4M | ~1087 ns | 736 B (6 allocs) |
| `BenchmarkDelData` | ~42M | ~27 ns | 0 B |

*Note: Results may vary based on hardware.*

### Performance Memory

Memory profiling ensures zero unexpected allocations during read operations.

Command to launch:
```bash
# Benchmarks
go test -memprofile=mem.out -bench=. ./...
go tool pprof mem.out
```

**Results:**
- **Zero allocations** for `Name` reads when using registered functions or manual values.
- **Controlled allocations** for `Info` reads (map construction).

---

## Test Coverage

**Target**: ≥85% statement coverage
**Current**: **91.0%**

### Coverage By Package

```bash
# View coverage by package
go test -cover ./...
```

**Output**:
```
github.com/nabbar/golib/monitor/info  coverage: 91.0% of statements
```

### Coverage By File

```bash
# Generate detailed report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Detailed Report**:

| File | Function | Coverage |
|------|----------|----------|
| `encode.go` | All | 100.0% |
| `info.go` | All | 100.0% |
| `interface.go` | `New` | 100.0% |
| `model.go` | `SetName` | 100.0% |
| `model.go` | `SetData` | 88.2% |
| `model.go` | `RegisterName` | 100.0% |
| `model.go` | `getName` | 87.5% |
| `model.go` | `getInfo` | 90.9% |

---

## Writing Tests

### Guidelines

1.  **Descriptive Specs**: Use full sentences in `It` blocks.
    ```go
    It("should return an error when the name is empty", func() { ... })
    ```
2.  **Independence**: Every test must run in isolation. Use `BeforeEach` to reset state.
3.  **Black-Box Preference**: Test public APIs (`info_test.go`) over internal methods unless necessary (`internal_test.go`).

### Test Template

```go
var _ = Describe("Feature Name", func() {
    var i info.Info

    BeforeEach(func() {
        var err error
        i, err = info.New("test-instance")
        Expect(err).NotTo(HaveOccurred())
    })

    Context("When performing action X", func() {
        // TC-FEAT-001
        It("should result in Y", func() {
            // Test implementation
        })
    })
})
```

---

## Best Practices

### ✅ DO
- **Do** use `race` detection in CI/CD pipelines.
- **Do** test both success and error paths.
- **Do** use `Eventually` for any asynchronous operations.

### ❌ DON'T
- **Don't** use `time.Sleep` in tests; it causes flakiness.
- **Don't** depend on global state or execution order.

---

## Reporting Bugs & Vulnerabilities

### Bug Report Template

When reporting issues, please provide:
1.  **Description**: What happened vs. what was expected.
2.  **Reproduction**: Minimal code snippet to trigger the bug.
3.  **Environment**: Go version, OS, Architecture.
4.  **Logs**: Output of `go test -v -race`.

### Security Vulnerabilities

Please report security issues privately via email or GitHub Security Advisories. **DO NOT** open public issues for security flaws.

---

## Resources

- **Ginkgo**: [onsi.github.io/ginkgo](https://onsi.github.io/ginkgo/)
- **Gomega**: [onsi.github.io/gomega](https://onsi.github.io/gomega/)
- **Go Testing**: [pkg.go.dev/testing](https://pkg.go.dev/testing)
- **Race Detector**: [go.dev/doc/articles/race_detector](https://go.dev/doc/articles/race_detector)

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for test generation, debugging, and documentation under human supervision. All tests are validated and reviewed by humans.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2020-2026 Nicolas JUHEL

