# Testing Guide

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-84.7%25-brightgreen)](TESTING.md)
[![Test Specs](https://img.shields.io/badge/Tests%20Specs-552-green)]()
[![Test Asserts](https://img.shields.io/badge/Tests%20Asserts-1407-green)]()

Comprehensive testing documentation for the monitor package, covering test execution, race detection, benchmarks, and quality assurance.

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

The monitor package employs a Behavior-Driven Development (BDD) approach using Ginkgo v2 and Gomega. This ensures that the health monitoring logic—including complex state transitions, hysteresis, and high-performance atomic metrics—is validated against expressive, human-readable specifications.

**Test Suite Summary**
- Total Specs: 552 across 4 packages
- Total Assertions: 1407
- Overall Coverage: 84.7% (aggregate statements)
- Race Detection: ✅ Zero data races detected
- Execution Time: ~14.5s (standard mode), ~31s (race mode)

---

## Test Plan (Monitor Package)

The test plan for the core `monitor` package is exhaustive. Any test not aligned with this plan is considered out-of-scope and non-essential.

1. **Functional Testing**: Validation of the 3-state machine (OK, Warn, KO) and hysteresis logic (Fall/Rise counts).
2. **Lifecycle Management**: Ensuring `Start()`, `Stop()`, and `Restart()` maintain a consistent internal state without leaking goroutines.
3. **High-Performance Metrics**: Accuracy validation of lock-free counters for Latency, Uptime, Downtime, and Transition times.
4. **Configuration Integrity**: Validation of `SetConfig` normalization engine and thread-safe metadata (`Info`) management.
5. **Middleware Execution**: Verification of the LIFO stack execution and result capturing via `mdlStatus`.
6. **Concurrent Stability**: High-pressure testing of the atomic read-path while background monitoring is active.
7. **Export Validation**: Correctness of JSON/Text marshalling and Prometheus metrics dispatching.

---

## Test Completeness (Monitor Package)

**Quality Indicators:**
- **Code Coverage**: **81.0%** of statements in the core `monitor` package.
- **Race Conditions**: **0** detected across 105 concurrent core scenarios.
- **Flakiness**: **Zero** flaky tests detected; all async state transitions are validated using deterministic polling via `Eventually`.

**Test Distribution:**
- ✅ **105 specifications** strictly covering the core implementation (`mon` struct).
- ✅ **412 assertions** validating behavior and state consistency.
- ✅ **12 example tests** demonstrating real-world API usage.
- ✅ **14 performance benchmarks** measuring throughput and atomic latency.
- ✅ **Zero out-of-scope tests**: All existing tests align with the functional or lifecycle categories.

---

## Test Architecture

The suite is architected to test sub-components in isolation (`status`, `info`) before validating the orchestrated `monitor` logic.

### Test Matrix

| Category           | Target                                   | Specs | Priority | Dependencies |
|--------------------|------------------------------------------|-------|----------|--------------|
| **Core Monitor**   | State Transitions, Lifecycle, Middleware | 105   | Critical | status, info |
| **Info Sub-pkg**   | Metadata evaluation & caching            | 106   | High     | None         |
| **Pool Sub-pkg**   | Batch operations & Aggregation           | 160   | Medium   | monitor      |
| **Status Sub-pkg** | Enumeration & multi-format encoding      | 181   | Major    | None         |

---

## Detailed Test Inventory (Monitor Core)

The following IDs uniquely identify all tests within the core `monitor` package.

**ID Pattern:** `TC-MON-<Category>-<Number>`
- **BS**: Basic & Functional
- **LC**: Lifecycle
- **TR**: Transitions
- **MT**: Metrics & Dispatch
- **EC**: Edge Cases & Context

| Test ID | File | Use Case | Priority | Expected Outcome |
|---------|------|----------|----------|------------------|
| **TC-MON-BS-001** | monitor_test.go | New() with valid info | Critical | Monitor instance created successfully |
| **TC-MON-BS-002** | monitor_test.go | New() with nil info | Critical | Returns error: "info cannot be nil" |
| **TC-MON-BS-003** | monitor_test.go | Set/Get HealthCheck | High | Function pointer correctly stored and retrieved |
| **TC-MON-BS-004** | monitor_test.go | Set/Get Config | High | Thresholds and intervals applied to internal state |
| **TC-MON-BS-005** | security_test.go | Nil handling (context in creation) | Major | Uses context.Background() if nil |
| **TC-MON-BS-006** | security_test.go | Nil handling (context in SetConfig) | Major | Uses internal context if nil |
| **TC-MON-BS-007** | security_test.go | HealthCheck change while running | High | Dynamic update of diagnostic logic without restart |
| **TC-MON-BS-008** | metrics_test.go | InfoMap metadata retrieval | Medium | Returns correct version and environment data |
| **TC-MON-BS-009** | metrics_test.go | InfoName retrieval | Medium | Returns identifier from info sub-package |
| **TC-MON-BS-010** | metrics_test.go | InfoUpd dynamic update | High | Correctly replaces info implementation |
| **TC-MON-LC-001** | monitor_test.go | Basic Start/Stop cycle | Critical | Monitor starts, sleeps, and stops correctly |
| **TC-MON-LC-002** | lifecycle_test.go | Successful Start & IsRunning | Critical | Background ticker becomes active |
| **TC-MON-LC-003** | lifecycle_test.go | Periodic execution | High | HealthCheck is invoked multiple times at interval |
| **TC-MON-LC-004** | lifecycle_test.go | Handle missing HealthCheck | Major | No crash; records "missing healthcheck" error |
| **TC-MON-LC-005** | lifecycle_test.go | Implicit Stop before Start | Major | Prevents duplicate ticker leaks |
| **TC-MON-LC-006** | lifecycle_test.go | Idempotent Stop | Medium | Multiple Stop calls are safe |
| **TC-MON-LC-007** | lifecycle_test.go | Execution halt after Stop | High | No further HealthChecks after Stop() returns |
| **TC-MON-LC-008** | lifecycle_test.go | Restart logic | High | Full Stop/Start cycle resumes monitoring |
| **TC-MON-LC-009** | lifecycle_test.go | Restart from stopped state | Medium | Monitor starts correctly |
| **TC-MON-LC-010** | lifecycle_test.go | HealthCheck context cancel | Major | Diagnostic respects middleware timeout |
| **TC-MON-LC-011** | lifecycle_test.go | Clone state and running status | High | Cloned monitor inherits configuration and state |
| **TC-MON-LC-012** | security_test.go | Resource cleanup after multiple cycles | Major | No resource exhaustion under rapid cycling |
| **TC-MON-LC-013** | security_test.go | Goroutine leak prevention | Critical | Background runner terminates immediately on Stop |
| **TC-MON-TR-001** | transitions_test.go | Initial KO state | Major | Starts in KO before first check |
| **TC-MON-TR-002** | transitions_test.go | Initial Error Message | Medium | Contains "no healcheck still run" initially |
| **TC-MON-TR-003** | transitions_test.go | KO -> Warn transition | Critical | Switches after RiseCountKO successes |
| **TC-MON-TR-004** | transitions_test.go | Warn -> OK transition | Critical | Switches after RiseCountWarn successes |
| **TC-MON-TR-005** | transitions_test.go | IsRise flag accuracy | High | True during recovery phase |
| **TC-MON-TR-006** | transitions_test.go | OK -> Warn transition | Critical | Switches after FallCountWarn failures |
| **TC-MON-TR-007** | transitions_test.go | Warn -> KO transition | Critical | Switches after FallCountKO failures |
| **TC-MON-TR-008** | transitions_test.go | IsFall flag accuracy | High | True during degradation phase |
| **TC-MON-TR-009** | transitions_test.go | Threshold enforcement | Major | No transition if successes < threshold |
| **TC-MON-TR-010** | transitions_test.go | Rise counter reset on failure | Major | Counter resets to 0 during recovery phase |
| **TC-MON-TR-011** | transitions_test.go | Message clearing on success | Medium | Success check clears previous error message |
| **TC-MON-TR-012** | transitions_test.go | Message update on new error | Medium | New failure replaces old error message |
| **TC-MON-TR-013** | transitions_test.go | Transition indicator reset | Medium | Rise/Fall flags reset after transition completes |
| **TC-MON-MT-001** | metrics_test.go | Latency tracking | High | Measured duration matches sleep time |
| **TC-MON-MT-002** | metrics_test.go | Dynamic latency update | High | Latency updates on every cycle |
| **TC-MON-MT-003** | metrics_test.go | Uptime accumulation | High | Increases while status is OK |
| **TC-MON-MT-004** | metrics_test.go | Downtime accumulation | High | Increases while status is Warn or KO |
| **TC-MON-MT-005** | metrics_test.go | Prometheus registration | Medium | Invokes collector with registered names |
| **TC-MON-MT-006** | metrics_test.go | CollectStatus snapshot | Medium | Returns atomic snapshot of status and flags |
| **TC-MON-MT-007** | metrics_test.go | Timing snapshots (Collect*) | Medium | All timing methods return consistent durations |
| **TC-MON-MT-008** | metrics_test.go | Initial latency zero | Low | Correct initialization of atomic counters |
| **TC-MON-MT-009** | metrics_test.go | Uptime pause on non-OK | High | Counter halts immediately on degradation |
| **TC-MON-MT-010** | metrics_test.go | Downtime in Warn status | High | Warn status correctly contributes to downtime |
| **TC-MON-MT-011** | metrics_test.go | Rise Time tracking | Medium | Accumulates during transition to OK |
| **TC-MON-MT-012** | metrics_test.go | Fall Time tracking | Medium | Accumulates during transition to KO |
| **TC-MON-MT-013** | metrics_test.go | Metric name de-duplication | Low | RegisterMetricsAddName handles duplicates |
| **TC-MON-MT-014** | metrics_test.go | Metrics dispatch guard | Low | No dispatch if names or function missing |
| **TC-MON-MT-015** | metrics_test.go | Metrics registration safety | Low | No panic when registering without collector |
| **TC-MON-MT-016** | encoding_test.go | MarshalText formatting | Major | Follows pattern STATUS: Name (Info) | Durations |
| **TC-MON-MT-017** | encoding_test.go | MarshalJSON structure | Major | Generates valid JSON with all metric keys |
| **TC-MON-EC-001** | security_test.go | Start Timeout | Major | poolIsRunning respects MaxPoolStart |
| **TC-MON-EC-002** | security_test.go | Long-running HealthCheck timeout | Major | MiddleWare enforces checkTimeout context |
| **TC-MON-EC-005** | config_test.go | Interval Normalization | Major | Micro-intervals are reset to safe minimums |
| **TC-MON-EC-006** | security_test.go | Cancelled context handling | Medium | Monitor handles already cancelled start context |
| **TC-MON-EC-007** | security_test.go | Panic recovery | Critical | Middleware captures panics in HealthCheck |
| **TC-MON-EC-008** | security_test.go | Zero duration handling | Major | Normalizes all intervals to minimum safety values |
| **TC-MON-EC-009** | security_test.go | Zero threshold counts | Major | Thresholds normalized to 1 |
| **TC-MON-EC-010** | security_test.go | Very high thresholds (255) | Medium | Supports max uint8 thresholds |
| **TC-MON-EC-011** | security_test.go | Extreme check frequency | Medium | Normalization prevents system overwhelming |
| **TC-MON-EC-012** | security_test.go | Long error message safety | Medium | Handles 10KB+ error strings |
| **TC-MON-EC-013** | security_test.go | Special char error safety | Medium | Handles quotes, backslashes, newlines in errors |

---

## Framework & Tools

### Testing Frameworks

#### Ginkgo v2 - BDD Testing Framework
- ✅ **Hierarchical organization**: `Describe`/`Context`/`It` for clear scenario mapping.
- ✅ **Async testing**: Heavy use of `Eventually` to poll background ticker state changes.
- ✅ **Lifecycle hooks**: `BeforeEach`/`AfterEach` ensuring fresh monitor instances for every spec.

#### Gomega - Matcher Library
- ✅ **Expressive matchers**: `BeNumerically` for duration comparison, `HaveOccurred` for error checking.
- ✅ **Async assertions**: `Eventually` polls for health check status updates.

### Testing Concepts & Standards

### ISTQB Alignment
- **Test Levels**: Component Testing (Core Monitor) and Integration Testing (Sub-packages).
- **Test Types**: Functional (State logic) and Non-functional (Benchmarks).
- **Techniques**: State Transition testing for hysteresis and Equivalence partitioning for configuration.

---

## Quick Start

```bash
# Install Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests in the suite
go test -v ./...

# Run with race detection (mandatory for concurrency verification)
CGO_ENABLED=1 go test -v -race ./...

# Run with coverage report
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Performance & Profiling

The monitor package is designed for high-concurrency environments where status polling frequency can reach thousands of requests per second. Performance validation is achieved through intensive benchmarking and PPROF profiling.

### 1. Performance Benchmarks

Captured on Linux amd64 (Intel Core i7-4700HQ @ 2.40GHz).

#### Read Path (Optimized Atomic Path)
The "hot path" used by metrics exporters and status probes is fully lock-free and allocation-free.

| Metric             | Operations  | Latency        | Memory | Allocations |
|--------------------|-------------|----------------|--------|-------------|
| `Status()` read    | 389,472,452 | **3.14 ns/op** | 0 B/op | 0 allocs/op |
| `Latency()` read   | 551,137,215 | **2.23 ns/op** | 0 B/op | 0 allocs/op |
| `Uptime()` read    | 525,178,024 | **2.24 ns/op** | 0 B/op | 0 allocs/op |
| `GetConfig()` read | 5,713,320   | 298.9 ns/op    | 0 B/op | 0 allocs/op |

#### Concurrency & Scaling
Tests validate near-linear scaling under multi-core pressure.

| Metric               | Operations    | Latency        | Efficiency                |
|----------------------|---------------|----------------|---------------------------|
| `Concurrent Status`  | 1,000,000,000 | **0.85 ns/op** | Near-linear scaling       |
| `Concurrent Metrics` | 665,232,284   | 1.98 ns/op     | Lock-free contention-less |

#### Administrative & Export Path
Infrequent operations (startup, serialization) use standard allocation patterns.

| Metric             | Operations | Latency      | Memory      | Allocations   |
|--------------------|------------|--------------|-------------|---------------|
| `SetConfig()`      | 24,315     | 49,407 ns/op | 24,823 B/op | 313 allocs/op |
| `MarshalText()`    | 111,370    | 9,126 ns/op  | 4,452 B/op  | 43 allocs/op  |
| `MarshalJSON()`    | 96,663     | 12,035 ns/op | 4,103 B/op  | 31 allocs/op  |
| `Clone()`          | 300,208    | 3,423 ns/op  | 1,542 B/op  | 54 allocs/op  |
| `Monitor Creation` | 311,746    | 3,239 ns/op  | 1,496 B/op  | 62 allocs/op  |

### 2. CPU Performance Analysis

- **Lock-Free Execution**: The CPU profile confirms that by replacing mutexes with `sync/atomic` primitives in the metrics container (`lastRun`), the "hot path" for status reads no longer triggers scheduler wait states.
- **Read Path Optimization**: Over 95% of CPU time during high-load polling is spent in raw memory loads rather than synchronization overhead.
- **Middleware Efficiency**: The primary CPU cost during an active health check cycle is the middleware stack traversal (`mdlStatus`), which is an acceptable trade-off for the structural safety and hysteresis logic it provides.

### 3. Memory & GC Pressure Analysis

- **Zero Garbage Path**: The most frequent operations (`Status`, `Latency`, etc.) produce **exactly 0 B/op**. This is critical for preventing Garbage Collector pauses in 24/7 monitoring services.
- **Allocation Isolation**: Heap allocations are isolated to the **Configuration Path** (`SetConfig`) and **Marshalling Path**. Since these occur at low frequency compared to status reads, the memory pressure remains minimal.
- **Buffer Management**: Optimized marshalling has resulted in a 60% reduction in allocation overhead compared to initial versions, ensuring high efficiency during large-scale reporting.

### 4. Profiling Commands

```bash
# Capture Statistics
go test -v -bench=. -benchmem ./monitor > res_bench.log

# Capture CPU Profile
go test -v -bench=BenchmarkMonitorStatusRead -cpuprofile=cpu.out ./monitor
go tool pprof -png cpu.out > res_cpu.png

# Capture Memory Profile
go test -v -bench=BenchmarkMonitorCreation -memprofile=mem.out ./monitor
go tool pprof -png mem.out > res_mem.png
```

---

## Test Coverage

**Aggregate Target**: ≥85% (Current: **84.7%**)

### Coverage By Package
```
github.com/nabbar/golib/monitor         coverage: 81.0% of statements
github.com/nabbar/golib/monitor/info    coverage: 91.0% of statements
github.com/nabbar/golib/monitor/pool    coverage: 91.2% of statements
github.com/nabbar/golib/monitor/status  coverage: 98.4% of statements
```

---

## Writing Tests

### Guidelines
1. **AAA Pattern**: Structure specs as Arrange (setup), Act (trigger check), Assert (Expect).
2. **Clean Lifecycle**: Always `Stop()` monitors in `AfterEach`.
3. **Avoid Sleeps**: Use Gomega's `Eventually()` for async state transitions.

### Test Template
```go
var _ = Describe("monitor/lifecycle", func() {
    var (
        ctx context.Context
        mon types.Monitor
    )

    BeforeEach(func() {
        ctx = context.Background()
        mon = createTestMonitor()
    })

    AfterEach(func() {
        if mon != nil {
            mon.Stop(ctx)
        }
    })

    It("should start background ticker", func() {
        // Act
        err := mon.Start(ctx)
        
        // Assert
        Expect(err).ToNot(HaveOccurred())
        Eventually(mon.IsRunning).Should(BeTrue())
    })
})
```

---

## Best Practices

### Test Independence
- ✅ Each test should be independent.
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup.
- ✅ Create fresh instances per test.
- ❌ Avoid global mutable state.

### Assertions
- ✅ Use appropriate matchers (`Equal`, `BeNumerically`, `HaveOccurred`).
- ✅ Use `Eventually` for async operations.

---

## Reporting Bugs & Vulnerabilities

### Bug Report Template

When reporting a bug in the test suite or the socket package, please use this template:

```markdown
**Title**: [BUG] Brief description of the bug

**Description**:
[A clear and concise description of what the bug is.]

**Steps to Reproduce:**
1. [First step]
2. [Second step]
3. [...]

**Expected Behavior**:
[A clear and concise description of what you expected to happen]

**Actual Behavior**:
[What actually happened]

**Code Example**:
[Minimal reproducible example]

**Test Case** (if applicable):
[Paste full test output with -v flag]

**Environment**:
- Go version: `go version`
- OS: Linux/macOS/Windows
- Architecture: amd64/arm64
- Package version: vX.Y.Z or commit hash

**Additional Context**:
[Any other relevant information]

**Logs/Error Messages**:
[Paste error messages or stack traces here]

**Possible Fix:**
[If you have suggestions]
```

### Security Vulnerability Template

**⚠️ IMPORTANT**: For security vulnerabilities, please **DO NOT** create a public issue.

Instead, report privately via:
1. GitHub Security Advisories (preferred)
2. Email to the maintainer (see footer)

**Vulnerability Report Template:**

```markdown
**Vulnerability Type:**
[e.g., Overflow, Race Condition, Memory Leak, Denial of Service]

**Severity:**
[Critical / High / Medium / Low]

**Affected Component:**
[e.g., interface.go, context.go, specific function]

**Affected Versions**:
[e.g., v1.0.0 - v1.2.3]

**Vulnerability Description:**
[Detailed description of the security issue]

**Attack Scenario**:
1. Attacker does X
2. System responds with Y
3. Attacker exploits Z

**Proof of Concept:**
[Minimal code to reproduce the vulnerability]
[DO NOT include actual exploit code]

**Impact**:
- Confidentiality: [High / Medium / Low]
- Integrity: [High / Medium / Low]
- Availability: [High / Medium / Low]

**Proposed Fix** (if known):
[Suggested approach to fix the vulnerability]

**CVE Request**:
[Yes / No / Unknown]

**Coordinated Disclosure**:
[Willing to work with maintainers on disclosure timeline]
```

### Issue Labels

When creating GitHub issues, use these labels:

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Improvements to docs
- `performance`: Performance issues
- `test`: Test-related issues
- `security`: Security vulnerability (private)
- `help wanted`: Community help appreciated
- `good first issue`: Good for newcomers

### Reporting Guidelines

**Before Reporting:**
1. ✅ Search existing issues to avoid duplicates
2. ✅ Verify the bug with the latest version
3. ✅ Run tests with `-race` detector
4. ✅ Check if it's a test issue or package issue
5. ✅ Collect all relevant logs and outputs

**What to Include:**
- Complete test output (use `-v` flag)
- Go version (`go version`)
- OS and architecture (`go env GOOS GOARCH`)
- Race detector output (if applicable)
- Coverage report (if relevant)

**Response Time:**
- **Bugs**: Typically reviewed within 48 hours
- **Security**: Acknowledged within 24 hours
- **Enhancements**: Reviewed as time permits

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/): <description>
- [Gomega Matchers](https://onsi.github.io/gomega/): <description>
- [Go Testing](https://pkg.go.dev/testing): <description>
- [Go Coverage](https://go.dev/blog/cover): <description>

**Testing References**
- [ISTQB concept](http://...): <description>

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector): <description>
- [Go Memory Model](https://go.dev/ref/mem): <description>
- [sync Package](https://pkg.go.dev/sync): <description>
- [sync/atomic Package](https://pkg.go.dev/sync/atomic): <description>

**Performance**
- [Go Profiling](https://go.dev/blog/pprof): <description>
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks): <description>
- [Execution Tracer](https://go.dev/doc/diagnostics#execution-tracer): <description>

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for test generation, debugging, and documentation under human supervision. All tests are validated and reviewed by humans.

---

## License
MIT License - Copyright (c) 2022-2025 Nicolas JUHEL
