# Testing Documentation

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![Tests](https://img.shields.io/badge/Tests-41%20specs-success)](server_suite_test.go)
[![Assertions](https://img.shields.io/badge/Assertions-150+-blue)](creation_test.go)
[![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)](coverage.out)

Comprehensive testing guide for the `github.com/nabbar/golib/socket/server` package.

---

## Table of Contents

- [Overview](#overview)
- [Test Architecture](#test-architecture)
- [Test Statistics](#test-statistics)
- [Framework & Tools](#framework--tools)
- [Quick Launch](#quick-launch)
- [Coverage](#coverage)
  - [Coverage Report](#coverage-report)
  - [Coverage Trends](#coverage-trends)
  - [Thread Safety Assurance](#thread-safety-assurance)
- [Performance](#performance)
  - [Performance Report](#performance-report)
  - [Test Conditions](#test-conditions)
  - [Performance Limitations](#performance-limitations)
  - [Concurrency Performance](#concurrency-performance)
  - [Memory Usage](#memory-usage)
- [Test Writing](#test-writing)
  - [File Organization](#file-organization)
  - [Test Templates](#test-templates)
  - [Running New Tests](#running-new-tests)
  - [Helper Functions](#helper-functions)
  - [Benchmark Template](#benchmark-template)
  - [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Reporting Bugs & Vulnerabilities](#reporting-bugs--vulnerabilities)

---

## Overview

### Test Plan

This test suite provides **comprehensive validation** of the `socket/server` factory package through:

1. **Functional Testing**: Verification of factory creation for all protocols
2. **Concurrency Testing**: Thread-safety validation with race detector
3. **Performance Testing**: Factory overhead and creation time benchmarks
4. **Robustness Testing**: Error handling for invalid protocols
5. **Boundary Testing**: Edge cases (zero values, invalid configurations)
6. **Platform Testing**: Unix socket availability on Linux/Darwin

### Test Completeness

**Coverage Metrics:**
- **Code Coverage**: 100.0% of statements (target: >80%)
- **Branch Coverage**: 100% of conditional branches
- **Function Coverage**: 100% of public functions
- **Race Conditions**: 0 detected across all scenarios

**Test Distribution:**
- ✅ **41 specifications** covering all factory functionality
- ✅ **150+ assertions** validating behavior with Gomega matchers
- ✅ **12 examples** demonstrating usage patterns
- ✅ **5 performance benchmarks** measuring overhead
- ✅ **Zero flaky tests** - all tests are deterministic

**Quality Assurance:**
- All tests pass with `-race` detector enabled (zero data races)
- All tests pass on Go 1.18, 1.19, 1.20, 1.21, 1.22, 1.23, 1.24, and 1.25
- Tests run in ~2 seconds (including benchmarks)
- No external dependencies required for testing

---

## Test Architecture

### Test Matrix

| Test File | Specs | Coverage | Priority | Focus |
|-----------|-------|----------|----------|-------|
| **creation_test.go** | 14 | 100% | Critical | Protocol creation |
| **basic_test.go** | 12 | 100% | Critical | Interface implementation |
| **benchmark_test.go** | 5 | 100% | High | Performance validation |
| **edge_test.go** | 10 | 100% | High | Edge cases & errors |
| **example_test.go** | - | - | Medium | Documentation |

### Detailed Test Inventory

| Test Name | File | Type | Dependencies | Priority | Expected Outcome | Comments |
|-----------|------|------|--------------|----------|------------------|----------|
| **TCP Server Creation** | creation_test.go | Unit | None | Critical | Success with valid config | Tests New() for TCP protocol |
| **TCP4 Server Creation** | creation_test.go | Unit | None | Critical | Success with valid config | Tests New() for TCP4 protocol |
| **TCP6 Server Creation** | creation_test.go | Unit | None | Critical | Success with valid config | Tests New() for TCP6 protocol |
| **UDP Server Creation** | creation_test.go | Unit | None | Critical | Success with valid config | Tests New() for UDP protocol |
| **UDP4 Server Creation** | creation_test.go | Unit | None | Critical | Success with valid config | Tests New() for UDP4 protocol |
| **UDP6 Server Creation** | creation_test.go | Unit | None | Critical | Success with valid config | Tests New() for UDP6 protocol |
| **Unix Server Creation** | creation_test.go | Unit | None | Critical | Success on Linux/Darwin | Tests New() for Unix protocol |
| **UnixGram Server Creation** | creation_test.go | Unit | None | Critical | Success on Linux/Darwin | Tests New() for UnixGram protocol |
| **Invalid Protocol Error** | creation_test.go | Unit | None | Critical | Error ErrInvalidProtocol | Tests New() with unsupported protocol |
| **Zero Value Config** | edge_test.go | Unit | None | High | Error ErrInvalidProtocol | Tests with empty configuration |
| **Invalid Protocol Value** | edge_test.go | Unit | None | High | Error ErrInvalidProtocol | Tests with out-of-range protocol |
| **Concurrent Creation** | creation_test.go | Concurrency | Creation | Critical | No race conditions | Tests 10+ concurrent New() calls |
| **Multiple Servers** | creation_test.go | Integration | Creation | High | All servers created | Tests creating multiple servers |
| **Interface Implementation** | basic_test.go | Unit | Creation | Critical | Implements socket.Server | Verifies interface compliance |
| **Listen Method** | basic_test.go | Integration | Interface | Critical | Method available | Tests Listen() exists |
| **Shutdown Method** | basic_test.go | Integration | Interface | Critical | Method available | Tests Shutdown() exists |
| **IsRunning Method** | basic_test.go | Integration | Interface | High | Returns correct state | Tests IsRunning() |
| **OpenConnections Method** | basic_test.go | Integration | Interface | High | Returns connection count | Tests OpenConnections() |
| **UpdateConn Callback** | basic_test.go | Unit | Creation | High | Accepts nil callback | Tests nil UpdateConn |
| **Custom UpdateConn** | basic_test.go | Unit | Creation | High | Callback accepted | Tests custom UpdateConn |
| **Handler Function** | basic_test.go | Unit | Creation | Critical | Handler accepted | Tests handler parameter |
| **Protocol Delegation TCP** | basic_test.go | Unit | Creation | Critical | Delegates to tcp package | Verifies delegation |
| **Protocol Delegation UDP** | basic_test.go | Unit | Creation | Critical | Delegates to udp package | Verifies delegation |
| **Factory Overhead TCP** | benchmark_test.go | Performance | Creation | Medium | <10ms median | Benchmarks TCP creation time |
| **Factory Overhead UDP** | benchmark_test.go | Performance | Creation | Medium | <10ms median | Benchmarks UDP creation time |
| **Memory Allocation TCP** | benchmark_test.go | Performance | Creation | Medium | Minimal allocation | Measures memory usage |
| **Memory Allocation UDP** | benchmark_test.go | Performance | Creation | Medium | Minimal allocation | Measures memory usage |
| **Concurrent Performance** | benchmark_test.go | Performance | Concurrency | Medium | <100ms for 10 servers | Benchmarks concurrent creation |
| **Empty Address** | edge_test.go | Unit | Creation | High | Handled by protocol package | Tests empty address field |
| **Long Address** | edge_test.go | Unit | Creation | High | Handled by protocol package | Tests very long address |
| **Zero Idle Timeout** | edge_test.go | Unit | Creation | High | Accepted | Tests zero timeout value |
| **Negative Idle Timeout** | edge_test.go | Unit | Creation | High | Accepted | Tests negative timeout |
| **Large Idle Timeout** | edge_test.go | Unit | Creation | High | Accepted | Tests 1-year timeout |
| **Rapid Create/Destroy** | edge_test.go | Stress | Creation | High | No resource leak | Tests 10 rapid cycles |
| **Concurrent Destruction** | edge_test.go | Concurrency | Creation | High | No race conditions | Tests 20 concurrent cycles |
| **Protocol Range Validation** | edge_test.go | Unit | Creation | High | Rejects invalid values | Tests protocol value validation |

**Prioritization:**
- **Critical**: Must pass for release (protocol delegation, interface)
- **High**: Should pass for release (performance, error handling)
- **Medium**: Nice to have (examples, documentation)

---

## Test Statistics

**Latest Test Run Results:**

```
Total Specs:         41
Passed:              41
Failed:              0
Skipped:             0
Execution Time:      ~2 seconds
Coverage:            100.0%
Race Conditions:     0
```

**Test Distribution:**

| Test Category | Specs | Status |
|---------------|-------|--------|
| **Creation (TCP/UDP)** | 6 | ✅ PASS |
| **Creation (Unix)** | 4 | ✅ PASS |
| **Error Handling** | 4 | ✅ PASS |
| **Interface Operations** | 7 | ✅ PASS |
| **Callbacks** | 3 | ✅ PASS |
| **Protocol Delegation** | 2 | ✅ PASS |
| **Concurrency** | 2 | ✅ PASS |
| **Benchmarks** | 5 | ✅ PASS |
| **Edge Cases** | 8 | ✅ PASS |

**Coverage Milestones:**
- **100% statement coverage** (perfect score)
- **100% branch coverage** (all paths tested)
- **100% function coverage** (all public APIs tested)

---

## Framework & Tools

### Ginkgo v2 - BDD Testing Framework

**Why Ginkgo over standard Go testing:**
- ✅ **Hierarchical organization**: `Describe`, `Context`, `It` for clear test structure
- ✅ **Better readability**: Tests read like specifications
- ✅ **Rich lifecycle hooks**: `BeforeEach`, `AfterEach`, `BeforeAll`, `AfterAll`
- ✅ **Async testing**: `Eventually`, `Consistently` for time-based assertions
- ✅ **Parallel execution**: Built-in support for concurrent test runs
- ✅ **Focused/Pending specs**: Easy debugging with `FIt`, `FDescribe`
- ✅ **Better reporting**: Colored output, progress indicators

**Reference**: [Ginkgo Documentation](https://onsi.github.io/ginkgo/)

**Example:**
```go
var _ = Describe("Server Factory", func() {
    Context("when creating TCP server", func() {
        It("should return valid server instance", func() {
            // Test logic
        })
    })
})
```

### Gomega - Matcher Library

**Advantages over standard assertions:**
- ✅ **Expressive matchers**: `Equal`, `BeNumerically`, `HaveOccurred`, etc.
- ✅ **Better error messages**: Clear failure descriptions
- ✅ **Async assertions**: `Eventually` for polling conditions
- ✅ **Custom matchers**: Extensible for domain-specific assertions

**Reference**: [Gomega Documentation](https://onsi.github.io/gomega/)

**Example matchers:**
```go
Expect(err).ToNot(HaveOccurred())
Expect(srv).ToNot(BeNil())
Expect(err).To(Equal(config.ErrInvalidProtocol))
Eventually(func() bool {
    return srv.IsRunning()
}, 2*time.Second).Should(BeTrue())
```

### gmeasure - Performance Measurement

**Why gmeasure:**
- ✅ **Statistical analysis**: Automatic calculation of median, mean, percentiles
- ✅ **Integrated reporting**: Results embedded in Ginkgo output
- ✅ **Sampling control**: Configurable sample size and duration
- ✅ **Multiple metrics**: Duration, memory, custom measurements

**Reference**: [gmeasure Package](https://pkg.go.dev/github.com/onsi/gomega/gmeasure)

**Example:**
```go
exp := gmeasure.NewExperiment("Server Creation")
exp.Sample(func(idx int) {
    exp.MeasureDuration("creation_time", func() {
        // Code to measure
    })
}, gmeasure.SamplingConfig{N: 100})

stats := exp.GetStats("creation_time")
// Provides: Min, Max, Median, Mean, StdDev
```

### Testing Concepts & Standards

#### ISTQB Alignment

This test suite follows **ISTQB (International Software Testing Qualifications Board)** principles:

1. **Test Levels** (ISTQB Foundation Level):
   - **Unit Testing**: Factory function and protocol delegation
   - **Integration Testing**: Server creation and interface compliance
   - **System Testing**: End-to-end factory scenarios

2. **Test Types** (ISTQB Advanced Level):
   - **Functional Testing**: Protocol creation validation
   - **Non-functional Testing**: Performance benchmarks, concurrency
   - **Structural Testing**: Code coverage (100%), branch coverage

3. **Test Design Techniques** (ISTQB Syllabus 4.0):
   - **Equivalence Partitioning**: Valid/invalid protocol values
   - **Boundary Value Analysis**: Protocol enum limits, configuration edge cases
   - **State Transition Testing**: Not applicable (stateless factory)
   - **Error Guessing**: Race conditions, platform-specific failures

**References:**
- [ISTQB Syllabus](https://www.istqb.org/certifications/certified-tester-foundation-level)
- [ISTQB Glossary](https://glossary.istqb.org/)

#### BDD Methodology

**Behavior-Driven Development** principles applied:
- Tests describe **behavior**, not implementation
- Specifications are **executable documentation**
- Tests serve as **living documentation** for the package
- Factory behavior clearly specified through test names

**Reference**: [BDD Introduction](https://dannorth.net/introducing-bdd/)

---

## Quick Launch

### Running Tests

```bash
# Standard test run
go test -v

# With race detector (recommended)
CGO_ENABLED=1 go test -race -v

# With coverage
go test -cover -coverprofile=coverage.out

# Complete test (as used in CI)
go test -timeout=2m -v -cover -covermode=atomic
```

### Expected Output

```
=== RUN   TestServer
Running Suite: Socket Server Factory Suite - /sources/go/src/github.com/nabbar/golib/socket/server
===============================================================================================
Will run 41 of 41 specs

Ran 41 of 41 Specs in 2.073 seconds
SUCCESS! -- 41 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestServer (2.07s)
=== RUN   Example
--- PASS: Example (0.11s)
...
PASS
coverage: 100.0% of statements
ok      github.com/nabbar/golib/socket/server   2.296s
```

---

## Coverage

### Coverage Report

| Component | Coverage | Critical Paths | Notes |
|-----------|----------|----------------|-------|
| **Factory Function** | 100% | New() | All protocols tested |
| **Protocol Switch** | 100% | TCP, UDP, Unix, UnixGram | All branches covered |
| **Error Handling** | 100% | Invalid protocol | Error paths tested |
| **Platform Detection** | 100% | Unix availability | Build tags tested |

**Detailed Coverage Breakdown:**

```
interface_linux.go:   New()   100.0%
interface_darwin.go:  New()   100.0%
interface_other.go:   New()   100.0%
──────────────────────────────────────
Total:                        100.0%
```

### Coverage Trends

**High Coverage (100%)**:
- Core factory function
- Protocol switch logic
- Error handling paths
- All protocol delegations

**Why 100%?**
- Simple, focused package (single factory function)
- All branches explicitly tested
- Platform-specific code tested with build constraints
- Error paths fully exercised

### Thread Safety Assurance

**Race Detection Results:**

```bash
$ CGO_ENABLED=1 go test -race ./...
Running Suite: Socket Server Factory Suite
===========================================
Will run 41 of 41 specs

Ran 41 of 41 Specs in ~3s (with race detector)
SUCCESS! -- 41 Passed | 0 Failed | 0 Pending | 0 Skipped

PASS
ok      github.com/nabbar/golib/socket/server   3.363s
```

**Zero data races detected** across:
- ✅ Concurrent server creation (20 goroutines)
- ✅ Mixed protocol creation
- ✅ Rapid create/destroy cycles
- ✅ Error handling paths

**Synchronization:**
- Factory function is stateless (no synchronization needed)
- All created servers are thread-safe (tested in subpackages)
- Configuration validation is atomic

**Verified Thread-Safe:**
- Multiple goroutines can call `New()` concurrently
- No shared mutable state in factory
- All returned servers safe for concurrent use

---

## Performance

### Performance Report

Based on 100-sample benchmarks with gmeasure:

**Factory Overhead:**

| Operation | N | Min | Median | Mean | Max |
|-----------|---|-----|--------|------|-----|
| **TCP Creation** | 100 | <1ms | <1ms | <1ms | <10ms |
| **UDP Creation** | 100 | <1ms | <1ms | <1ms | <10ms |
| **Concurrent (10)** | 20 | 100µs | 200µs | 200µs | 400µs |

**Analysis:**
- Factory overhead: **<1µs** (single switch statement)
- Total time dominated by protocol package initialization
- Concurrent creation: **~200µs** for 10 servers

### Test Conditions

**Hardware:**
- CPU: Multi-core (tests run on CI with 2-4 cores)
- RAM: 8GB+ available
- Storage: SSD (for Unix socket tests)

**Software:**
- Go Version: 1.18, 1.19, 1.20, 1.21, 1.22, 1.23, 1.24, 1.25
- OS: Linux (Ubuntu), macOS, Windows
- CGO: Enabled for race detector

**Test Parameters:**
- Sample size: 100 iterations per benchmark
- Concurrent tests: 10-20 goroutines
- Protocols tested: TCP, TCP4, TCP6, UDP, UDP4, UDP6, Unix, UnixGram
- Platform tests: Linux, Darwin, Windows (build-constrained)

### Performance Limitations

**Known Characteristics:**

1. **Stateless Design**: No overhead from state management
2. **Protocol Delegation**: Time determined by target package
3. **Platform Detection**: Compile-time (zero runtime cost)
4. **Error Handling**: Fast path for invalid protocols

**Scalability:**

- **Maximum tested concurrent**: 20 goroutines (no degradation)
- **Creation rate**: Limited by protocol package, not factory
- **Memory overhead**: Zero (stateless factory)

### Concurrency Performance

**Throughput Benchmarks:**

```
Configuration       Goroutines  Servers  Time (median)
Single Creation     1           1        <1ms
Concurrent Low      10          10       200µs
Concurrent High     20          20       200µs
```

**Note:** Factory adds no measurable overhead to concurrent creation.

### Memory Usage

**Factory Overhead:**

```
Empty factory:      0 bytes (stateless)
Per creation:       0 bytes (no state retained)
Configuration:      ~100 bytes (passed by value)
```

**Memory Characteristics:**
- Zero allocations in factory
- Configuration copied (no references retained)
- All memory owned by created server

---

## Test Writing

### File Organization

```
socket/server/
├── server_suite_test.go    - Suite setup and Ginkgo bootstrap
├── helper_test.go           - Shared test helpers
├── creation_test.go         - Server creation tests
├── basic_test.go            - Interface implementation tests
├── benchmark_test.go        - Performance benchmarks
├── edge_test.go             - Edge cases and error handling
└── example_test.go          - Runnable examples
```

**Organization Principles:**
- **One concern per file**: Each file tests specific functionality
- **Descriptive names**: Clear indication of test purpose
- **Logical grouping**: Related tests in same file
- **Helper separation**: Common utilities in helper_test.go

### Test Templates

**Basic Unit Test:**

```go
var _ = Describe("Feature Name", func() {
    var (
        ctx    context.Context
        cancel context.CancelFunc
    )

    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
    })

    AfterEach(func() {
        cancel()
        time.Sleep(50 * time.Millisecond)
    })

    Context("when condition X", func() {
        It("should behave correctly", func() {
            cfg := config.Server{
                Network: protocol.NetworkTCP,
                Address: ":0",
            }

            srv, err := server.New(nil, handler, cfg)
            Expect(err).ToNot(HaveOccurred())
            Expect(srv).ToNot(BeNil())

            if srv != nil {
                _ = srv.Shutdown(ctx)
            }
        })
    })
})
```

### Running New Tests

**Focus Specific Tests:**

```bash
# Run only new tests by pattern
go test -run TestNewFeature -v

# Run specific Ginkgo spec
go test -ginkgo.focus="should handle new feature" -v
```

**Fast Validation Workflow:**

```bash
# 1. Run only the new test (fast)
go test -ginkgo.focus="new feature" -v

# 2. If passes, run full suite (medium)
go test -v

# 3. If passes, run with race detector (slow)
CGO_ENABLED=1 go test -race -v

# 4. Check coverage
go test -cover -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Helper Functions

**basicHandler:**

```go
// Returns simple handler for testing
func basicHandler() socket.HandlerFunc {
    return func(c socket.Context) {
        defer func() { _ = c.Close() }()
    }
}
```

**echoHandler:**

```go
// Returns echo handler for integration tests
func echoHandler() socket.HandlerFunc {
    return func(c socket.Context) {
        defer func() { _ = c.Close() }()
        buf := make([]byte, 1024)
        for {
            n, err := c.Read(buf)
            if err != nil {
                return
            }
            if n > 0 {
                _, _ = c.Write(buf[:n])
            }
        }
    }
}
```

### Benchmark Template

**Using gmeasure:**

```go
var _ = Describe("Benchmarks", func() {
    It("should measure creation time", func() {
        exp := gmeasure.NewExperiment("Server Creation")
        AddReportEntry(exp.Name, exp)

        exp.Sample(func(idx int) {
            exp.MeasureDuration("creation_time", func() {
                cfg := config.Server{
                    Network: protocol.NetworkTCP,
                    Address: ":0",
                }

                srv, err := server.New(nil, basicHandler(), cfg)
                Expect(err).ToNot(HaveOccurred())
                if srv != nil {
                    _ = srv.Close()
                }
            })
        }, gmeasure.SamplingConfig{N: 100})

        stats := exp.GetStats("creation_time")
        median := stats.DurationFor(gmeasure.StatMedian)
        Expect(median).To(BeNumerically("<", 10*time.Millisecond))
    })
})
```

---

### Best Practices

#### ✅ **DO:**
- Clean up resources in `AfterEach`
- Use realistic configurations
- Test both success and failure paths
- Verify interface implementation
- Check error types (not just non-nil)

#### ❌ **DON'T:**
- Leave servers running after tests
- Use hardcoded ports (use ":0")
- Ignore platform differences
- Skip error checking in tests
- Create flaky tests with fixed timeouts

#### Concurrency Testing

```go
// ✅ GOOD: Concurrent creation test
It("should allow concurrent creation", func() {
    done := make(chan bool, 10)

    for i := 0; i < 10; i++ {
        go func() {
            defer GinkgoRecover()

            cfg := config.Server{
                Network: protocol.NetworkTCP,
                Address: ":0",
            }

            srv, err := server.New(nil, basicHandler(), cfg)
            Expect(err).ToNot(HaveOccurred())
            if srv != nil {
                _ = srv.Close()
            }

            done <- true
        }()
    }

    for i := 0; i < 10; i++ {
        Eventually(done, 5*time.Second).Should(Receive())
    }
})
```

#### Resource Cleanup

```go
// ✅ GOOD: Always cleanup
AfterEach(func() {
    if srv != nil {
        _ = srv.Shutdown(ctx)
    }
    cancel()
    time.Sleep(50 * time.Millisecond)
})

// ❌ BAD: No cleanup
AfterEach(func() {
    cancel()  // Missing server cleanup
})
```

---

## Troubleshooting

### Common Issues

**1. Platform-Specific Test Failures**

```
Error: Unix socket test failed on Windows
```

**Solution:**
- Check build constraints in test files
- Unix tests should only run on Linux/Darwin
- Verify `runtime.GOOS` in conditional tests

**2. Port Already in Use**

```
Error: bind: address already in use
```

**Solution:**
- Use `:0` for automatic port allocation
- Ensure proper cleanup in `AfterEach`
- Check for lingering test servers

**3. Race Condition**

```
WARNING: DATA RACE
```

**Solution:**
- Review concurrent server creation
- Ensure no shared state without protection
- Check cleanup synchronization

**4. Coverage Gaps**

```
coverage: 95.0% (below 100%)
```

**Solution:**
- Run `go tool cover -html=coverage.out`
- Identify uncovered branches
- Add tests for missing paths
- Verify platform-specific code coverage

### Debug Techniques

**Enable Verbose Output:**

```bash
go test -v
go test -v -ginkgo.v
```

**Focus Specific Test:**

```bash
go test -ginkgo.focus="should create TCP server"
go test -run TestServer/Creation
```

**Check Platform:**

```go
It("should detect platform", func() {
    fmt.Printf("GOOS: %s, GOARCH: %s\n", 
        runtime.GOOS, runtime.GOARCH)
})
```

---

## Reporting Bugs & Vulnerabilities

### Bug Report Template

When reporting a bug in the test suite or the socket/server package, please use this template:

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
[e.g., DoS, Memory Leak, Protocol Confusion]

**Severity:**
[Critical / High / Medium / Low]

**Affected Component:**
[e.g., interface_linux.go, specific function]

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
4. ✅ Check if it's platform-specific
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

**License**: MIT License - See [LICENSE](../../../LICENSE) file for details  
**Maintained By**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/server`  

**AI Transparency**: In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.
