# Testing Documentation

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![Tests](https://img.shields.io/badge/Tests-120%20specs-success)](suite_test.go)
[![Assertions](https://img.shields.io/badge/Assertions-300+-blue)](suite_test.go)
[![Coverage](https://img.shields.io/badge/Coverage-80.8%25-brightgreen)](coverage.out)

Comprehensive testing guide for the `github.com/nabbar/golib/ioutils/multi` package using BDD methodology with Ginkgo v2 and Gomega.

---

## Table of Contents

- [Overview](#overview)
- [Test Architecture](#test-architecture)
- [Test Statistics](#test-statistics)
- [Framework & Tools](#framework--tools)
- [Quick Launch](#quick-launch)
- [Coverage](#coverage)
  - [Coverage Report](#coverage-report)
  - [Uncovered Code Analysis](#uncovered-code-analysis)
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

This test suite provides **comprehensive validation** of the `multi` package following **ISTQB** principles. It focuses on validating the **MultiWriter** behavior, adaptive strategies, and concurrency safety through:

1.  **Functional Testing**: Verification of all public APIs (AddWriter, SetInput, Write, Read, Copy).
2.  **Non-Functional Testing**: Performance benchmarking and concurrency safety validation.
3.  **Structural Testing**: Ensuring all code paths and logic branches are exercised, while acknowledging that coverage metrics are just one indicator of quality.

### Test Completeness

**Quality Indicators:**
-   **Code Coverage**: 80.8% of statements (Note: Used as a guide, not a guarantee of correctness).
-   **Race Conditions**: 0 detected across all scenarios.
-   **Flakiness**: 0 flaky tests detected.

**Test Distribution:**
-   ✅ **120 specifications** covering all major use cases
-   ✅ **300+ assertions** validating behavior
-   ✅ **15 performance benchmarks** measuring key metrics
-   ✅ **9 test files** organized by functional area
-   ✅ **Zero flaky tests** - all tests are deterministic

---

## Test Architecture

### Test Matrix

| Category | Files | Specs | Coverage | Priority | Dependencies |
|----------|-------|-------|----------|----------|-------------|
| **Basic** | constructor_test.go | 14 | 90%+ | Critical | None |
| **Reader Operations** | reader_test.go | 14 | 85%+ | Critical | Basic |
| **Writer Operations** | writer_test.go | 22 | 85%+ | Critical | Basic |
| **Copy Operations** | copy_test.go | 12 | 85%+ | Critical | Basic |
| **Concurrency** | concurrent_test.go | 11 | 90%+ | Critical | Implementation |
| **Adaptive Mode** | mode_test.go | 7 | 80%+ | High | Implementation |
| **Edge Cases** | edge_cases_test.go | 25 | 75%+ | High | Implementation |
| **Performance** | benchmark_test.go | 15 | N/A | Medium | Implementation |
| **Helpers** | helper_test.go | N/A | N/A | Low | All |
| **Examples** | example_test.go | N/A | N/A | Low | All |

### Detailed Test Inventory

**Test ID Pattern by File:**
- **TC-CT-xxx**: Constructor tests (constructor_test.go)
- **TC-RD-xxx**: Reader tests (reader_test.go)
- **TC-WR-xxx**: Writer tests (writer_test.go)
- **TC-CP-xxx**: Copy tests (copy_test.go)
- **TC-CC-xxx**: Concurrent tests (concurrent_test.go)
- **TC-MD-xxx**: Mode tests (mode_test.go)
- **TC-EC-xxx**: Edge case tests (edge_cases_test.go)
- **TC-BC-xxx**: Benchmark tests (benchmark_test.go)

| Test ID | File | Use Case | Priority | Expected Outcome |
|---------|------|----------|----------|------------------|
| **TC-CT-001** | constructor_test.go | **Initialization**: Create Multi with default configuration | Critical | Instance created with safe defaults (Discarders) |
| **TC-CT-002** | constructor_test.go | **Interface Compliance**: Verify Multi implements all interfaces | Critical | Implements Multi, io.ReadWriteCloser, io.StringWriter |
| **TC-CT-003** | constructor_test.go | **DiscardCloser**: Verify no-op reader/writer behavior | Medium | Read returns 0, Write accepts all data, Close is no-op |
| **TC-RD-001** | reader_test.go | **SetInput**: Set/Replace input reader atomically | Critical | Input source swapped safely without race conditions |
| **TC-RD-002** | reader_test.go | **Read Operations**: Read from input source | Critical | Data read correctly from underlying reader |
| **TC-RD-003** | reader_test.go | **Reader Access**: Get Reader() reference | High | Returns current reader instance |
| **TC-RD-004** | reader_test.go | **Close**: Close input reader | High | Underlying reader closed properly |
| **TC-WR-001** | writer_test.go | **AddWriter**: Add single/multiple writers atomically | Critical | Writer list updated safely, new writers receive data |
| **TC-WR-002** | writer_test.go | **Write Broadcasting**: Write data to multiple destinations | Critical | Data appears exactly once on all writers |
| **TC-WR-003** | writer_test.go | **WriteString**: String write operations | High | String data broadcast to all writers |
| **TC-WR-004** | writer_test.go | **Clean**: Remove all writers | Medium | Subsequent writes go to Discard until new writers added |
| **TC-WR-005** | writer_test.go | **Writer Access**: Get Writer() reference | High | Returns current writer instance |
| **TC-CP-001** | copy_test.go | **Copy Single**: Copy from Input to single Writer | Critical | Full stream replication to destination |
| **TC-CP-002** | copy_test.go | **Copy Multiple**: Copy from Input to multiple Writers | Critical | Full stream replication to all destinations |
| **TC-CP-003** | copy_test.go | **Copy Large Data**: Copy large data streams | High | Handles large data efficiently |
| **TC-CP-004** | copy_test.go | **Copy Errors**: Error handling in copy operations | High | Errors from reader/writer propagated correctly |
| **TC-CP-005** | copy_test.go | **Integration**: Mixed read/write/copy operations | High | Sequential operations work together |
| **TC-CC-001** | concurrent_test.go | **Concurrent AddWriter**: Concurrent writer additions | Critical | No race conditions during writer additions |
| **TC-CC-002** | concurrent_test.go | **Concurrent Write**: Concurrent write operations | Critical | No data races, no panics, consistent state |
| **TC-CC-003** | concurrent_test.go | **Concurrent SetInput**: Concurrent input changes | Critical | Safe input replacement during active operations |
| **TC-CC-004** | concurrent_test.go | **Concurrent Mixed**: Mixed Read/Write/Add operations | Critical | All operations thread-safe |
| **TC-MD-001** | mode_test.go | **Mode Flags**: IsParallel, IsSequential, IsAdaptive | High | Correct mode reported |
| **TC-MD-002** | mode_test.go | **Sequential Mode**: Force sequential write strategy | High | Uses sequential writes |
| **TC-MD-003** | mode_test.go | **Parallel Mode**: Force parallel write strategy | High | Uses parallel writes for large data |
| **TC-MD-004** | mode_test.go | **Adaptive Switch (S→P)**: High latency detected | High | Switches to Parallel mode automatically |
| **TC-MD-005** | mode_test.go | **Adaptive Switch (P→S)**: Low latency detected | High | Switches back to Sequential mode automatically |
| **TC-EC-001** | edge_cases_test.go | **Nil Handling**: Add nil writers or set nil input | High | Silently ignored or defaulted to DiscardCloser |
| **TC-EC-002** | edge_cases_test.go | **Empty Data**: Zero-length write/read operations | Medium | Handled gracefully without errors |
| **TC-EC-003** | edge_cases_test.go | **Large Data**: Very large write operations | Medium | Handles large buffers efficiently |
| **TC-EC-004** | edge_cases_test.go | **Error Propagation**: Write failure in one writer | High | Error returned to caller (first error encountered) |
| **TC-EC-005** | edge_cases_test.go | **Multiple Errors**: Errors from multiple writers | High | First error returned |
| **TC-BC-001** | benchmark_test.go | **Write Operations Benchmark**: 6 variations (writers × data sizes) | High | Sub-µs to 900µs mean, validates write efficiency |
| **TC-BC-002** | benchmark_test.go | **Read Operations Benchmark**: 3 variations (data sizes) | Medium | <1µs to 600µs mean, validates read delegation |
| **TC-BC-003** | benchmark_test.go | **Copy Operations Benchmark**: 6 variations (writers × data sizes) | High | Sub-µs to 900µs mean, validates broadcasting |
| **TC-BC-004** | benchmark_test.go | **Mode Comparison Benchmark**: Sequential vs Parallel vs Adaptive | High | Adaptive matches Sequential efficiency |
| **TC-BC-005** | benchmark_test.go | **Writer Management Benchmark**: 7 operations (constructor, add, clean, etc.) | Medium | All operations <1µs median |
| **TC-BC-006** | benchmark_test.go | **Log Broadcasting Scenario**: 10K lines to 3 destinations | High | 800µs mean validates real-world logging |
| **TC-BC-007** | benchmark_test.go | **Stream Replication Scenario**: 50K chunks to backups | High | 1.5ms mean validates replication |
| **TC-BC-008** | benchmark_test.go | **Adaptive Load Scenario**: Variable sizes with mode switching | Medium | 200µs mean validates adaptive behavior |

---

## Test Statistics

**Latest Test Run Results:**

```
Total Specs:         120
Passed:              119
Failed:              0
Skipped:             1
Execution Time:      ~2.03 seconds
Coverage:            80.8%
Race Conditions:     0
```

---

## Framework & Tools

### Testing Frameworks

#### Ginkgo v2 - BDD Testing Framework

**Why Ginkgo over standard Go testing:**
-   ✅ **Hierarchical organization**: `Describe`, `Context`, `It` for clear test structure.
-   ✅ **Better readability**: Tests read like specifications.
-   ✅ **Rich lifecycle hooks**: `BeforeEach`, `AfterEach` for setup/teardown.
-   ✅ **Async testing**: `Eventually`, `Consistently` for concurrent behavior.
-   ✅ **Parallel execution**: Built-in support for concurrent test runs.

#### Gomega - Matcher Library

**Advantages:**
-   ✅ **Expressive matchers**: `Equal`, `BeNumerically`, `HaveOccurred`.
-   ✅ **Async assertions**: `Eventually` polls for state changes.

#### gmeasure - Performance Measurement

Used for benchmarking throughput and latency within the BDD suite.

### Testing Concepts & Standards

#### ISTQB Alignment

This test suite follows **ISTQB (International Software Testing Qualifications Board)** principles:

1.  **Test Levels** (ISTQB Foundation Level):
    *   **Unit Testing**: Individual functions (`New`, `AddWriter`, `SetInput`).
    *   **Integration Testing**: Component interactions (`Write` broadcasting, `Copy`).
    *   **System Testing**: End-to-end scenarios (Concurrency, Examples).

2.  **Test Types** (ISTQB Advanced Level):
    *   **Functional Testing**: Verify behavior meets specifications (Broadcasting).
    *   **Non-Functional Testing**: Performance, concurrency, memory usage.
    *   **Structural Testing**: Code coverage (Branch coverage).

3.  **Test Design Techniques**:
    *   **Equivalence Partitioning**: Valid writers vs `nil` writers.
    *   **Boundary Value Analysis**: 0 writers, 1 writer, Threshold limits.
    *   **State Transition Testing**: Adaptive mode switching (Sequential <-> Parallel).
    *   **Error Guessing**: Concurrent access patterns.

#### Testing Pyramid

The suite follows the Testing Pyramid principle:

```
         /\
        /  \
       / E2E\       (System/Concurrency Tests)
      /______\
     /        \
    / Integr.  \    (Write/Copy/Mode Tests)
   /____________\
  /              \
 /   Unit Tests   \ (Constructor, Config, Helpers)
/__________________\
```

---

## Coverage

### Coverage Report

| Component | File | Coverage | Critical Paths |
|-----------|------|----------|----------------|
| **Interface** | interface.go | 100.0% | New(), configuration validation |
| **Core Logic** | model.go | 90.5% | AddWriter, SetInput, update logic |
| **Read** | read.go | 85.7% | Reader wrapper, Close handling |
| **Write** | writer.go | 78.5% | Sequential/Parallel write strategies |
| **Config** | config.go | 100.0% | DefaultConfig, validation |
| **Stats** | stat.go | 95.0% | Statistics tracking |

**Detailed Coverage:**

```
New()                100.0%  - All configuration paths tested
AddWriter()          100.0%  - Writer registration fully covered
SetInput()           100.0%  - Input source management
Write()               82.4%  - Standard and parallel paths
Read()                85.7%  - Wrapper delegation
Close()              100.0%  - Resource cleanup
Stats()               95.0%  - Metrics retrieval
IsParallel()         100.0%  - Mode checking
IsSequential()       100.0%  - Mode checking
IsAdaptive()         100.0%  - Mode checking
```

### Uncovered Code Analysis

**Uncovered Lines: 19.2% (target: <20%)**

#### 1. Parallel Write Error Paths (writer.go)

**Uncovered**: Lines handling concurrent write failures in parallel mode

```go
// UNCOVERED: Multiple concurrent writer failures
if e := <-errCh; e != nil {
    return len(p), e
}
```

**Reason**: Difficult to trigger specific error conditions with multiple goroutines failing simultaneously in integration tests.

**Impact**: Low - error paths are defensive, sequential fallback is well-tested

#### 2. Edge Cases in Adaptive Switching

**Uncovered**: Some state transition combinations in adaptive mode

**Reason**: Requires precise timing and latency conditions that are hard to reproduce consistently in tests.

**Impact**: Medium - core adaptive logic is tested, edge transitions are rare

### Thread Safety Assurance

**Race Detection Results:**

```bash
$ CGO_ENABLED=1 go test -race -v
Running Suite: IOUtils/Multi Package Suite
===========================================
Will run 120 of 120 specs

Ran 120 of 120 Specs in 2.03s
SUCCESS! -- 119 Passed | 0 Failed | 1 Skipped | 0 Pending

PASS
ok      github.com/nabbar/golib/ioutils/multi      2.123s
```

**Zero data races detected** across:
- ✅ Concurrent AddWriter operations
- ✅ Concurrent Write operations
- ✅ SetInput during active writes
- ✅ Stats() reads during writes
- ✅ Adaptive mode switching

**Synchronization Mechanisms:**

| Primitive | Usage | Thread-Safe Operations |
|-----------|-------|------------------------|
| `atomic.Value` | Reader/Writer storage | `Load()`, `Store()`, `Swap()` |
| `atomic.Int64` | Counters and stats | `Add()`, `Load()`, `Store()` |
| `atomic.Bool` | Mode flags | `Load()`, `Store()` |
| `sync.Map` | Writer registry | Thread-safe map operations |
| Channels | Error propagation | Buffered channels in parallel mode |

**Verified Thread-Safe:**
- All public methods can be called concurrently
- Dynamic writer addition during active writes
- Input source replacement without races
- Statistics reading without blocking writes


---

## Quick Launch

### Running All Tests

```bash
# Standard test run
go test -v

# With race detector (recommended)
CGO_ENABLED=1 go test -race -v

# With coverage
go test -cover -coverprofile=coverage.out

# Complete test suite (as used in CI)
go test -timeout=10m -v -cover -covermode=atomic ./...
```

### Expected Output

```
Running Suite: IOUtils/Multi Package Suite
===========================================
Random Seed: 1234567890

Will run 120 of 120 specs

••••••••••••••••••••••••••••••••••••••••••••••••••••••••••

Ran 120 of 120 Specs in 2.03 seconds
SUCCESS! -- 119 Passed | 0 Failed | 1 Skipped | 0 Pending

PASS
coverage: 80.8% of statements
ok      github.com/nabbar/golib/ioutils/multi   2.123s
```

---

## Performance

### Performance Report

**Benchmark Results (Aggregated Experiments):**

#### Write Operations

| Configuration | Sample Size | Median | Mean | Max | Notes |
|---------------|-------------|--------|------|-----|-------|
| Single writer, small data | 1000 | <1µs | <1µs | 300µs | Negligible overhead |
| 3 writers, small data | 1000 | <1µs | <1µs | 300µs | Broadcasting efficiency |
| Single writer, 1KB | 1000 | <1µs | <1µs | 400µs | Scales with data size |
| 3 writers, 1KB | 1000 | <1µs | <1µs | 300µs | Efficient multi-write |
| Single writer, 1MB | 100 | 400µs | 700µs | 3.4ms | Large data handling |
| 3 writers, 1MB | 100 | 600µs | 900µs | 3.4ms | Parallel write benefit |

#### Read Operations

| Configuration | Sample Size | Median | Mean | Max | Notes |
|---------------|-------------|--------|------|-----|-------|
| Read 100B | 1000 | <1µs | <1µs | 300µs | Minimal wrapper overhead |
| Read 1KB | 1000 | <1µs | <1µs | 400µs | Efficient delegation |
| Read 1MB | 100 | 400µs | 600µs | 3ms | Throughput maintained |

#### Mode Comparison

| Mode | Sample Size | Median | Mean | Max | Winner |
|------|-------------|--------|------|-----|--------|
| Sequential | 1000 | <1µs | <1µs | 100µs | ✅ Best for low latency |
| Parallel | 1000 | <1µs | <1µs | 600µs | Overhead visible |
| Adaptive | 1000 | <1µs | <1µs | 100µs | ✅ Smart switching |

#### Writer Management

| Operation | Sample Size | Median | Mean | Max | Notes |
|-----------|-------------|--------|------|-----|-------|
| Constructor (default) | 1000 | <1µs | <1µs | <1µs | Minimal overhead |
| Constructor (adaptive) | 1000 | <1µs | <1µs | <1µs | No config penalty |
| AddWriter (single) | 1000 | <1µs | <1µs | <1µs | Atomic operation |
| AddWriter (multiple) | 1000 | <1µs | <1µs | 100µs | Batch efficient |
| SetInput | 1000 | <1µs | <1µs | 100µs | Atomic swap |
| Clean | 1000 | <1µs | <1µs | <1µs | Fast cleanup |
| WriteString | 1000 | <1µs | <1µs | 200µs | String conversion |

#### Real-World Scenarios

| Scenario | Sample Size | Median | Mean | Max | Description |
|----------|-------------|--------|------|-----|-------------|
| Log broadcasting | 10 | 700µs | 800µs | 1.9ms | 10K log lines to 3 destinations |
| Stream replication | 10 | 1.3ms | 1.5ms | 2.8ms | 50K chunks to backup destinations |
| Adaptive under load | 10 | 200µs | 200µs | 500µs | Variable sizes with mode switching |

### Performance Analysis

**Key Findings:**

1.  **Sub-microsecond Operations**: Most operations complete in <1µs (median), demonstrating excellent efficiency
2.  **Large Data Handling**: 1MB writes scale predictably (600-900µs mean) with minimal degradation
3.  **Mode Efficiency**: Adaptive mode matches Sequential performance while enabling automatic optimization
4.  **Broadcasting Overhead**: Writing to 3 destinations adds minimal overhead vs single writer
5.  **Real-World Performance**: Log broadcasting (800µs) and stream replication (1.5ms) validate production readiness

**Test Conditions:**
-   **Hardware**: AMD64/ARM64 Multi-core, 8GB+ RAM
-   **Sample Sizes**: 1000 samples (micro-ops), 100 samples (large data), 10 samples (scenarios)
-   **Data Sizes**: Small (10B), Medium (1KB), Large (1MB)
-   **Writer Counts**: 1, 3, 5 concurrent destinations

### Performance Characteristics

**Strengths:**
-   ✅ **Atomic Operations**: Sub-microsecond writer/input management
-   ✅ **Efficient Broadcasting**: Multi-writer overhead <30% vs single writer
-   ✅ **Scalable**: 1MB data handled in <1ms mean latency
-   ✅ **Predictable**: Low standard deviation across all benchmarks
-   ✅ **Adaptive**: Mode switching without performance penalty

**Limitations:**
1.  **Goroutine Overhead**: Parallel mode adds latency for very small writes (<100B)
    -   *Observation*: Parallel max latency (600µs) > Sequential (100µs) in mode comparison
    -   *Mitigation*: Adaptive mode automatically avoids parallel for small payloads
2.  **Peak Latency**: Max latencies (P99) can reach 2-4ms under load
    -   *Context*: Acceptable for I/O-bound operations, GC-related spikes
3.  **Memory Allocation**: Parallel mode allocates 1 goroutine/writer + error channel per write
    -   *Impact*: Negligible for typical use cases (<10 writers)

### Memory Profile

-   **Sequential Write**: Zero allocations per operation
-   **Parallel Write**: ~1 allocation per writer (goroutine stack)
-   **Struct Overhead**: ~1KB base size (atomic values, maps)
-   **Real-World**: Log broadcasting (10K lines) = ~800µs, minimal GC pressure

---

## Test Writing

### File Organization

```
multi/
├── suite_test.go           # Test suite entry point (Ginkgo suite setup)
├── constructor_test.go     # Constructor and interface compliance tests (14 specs)
├── reader_test.go          # Read operations and input management tests (14 specs)
├── writer_test.go          # Write operations and output management tests (22 specs)
├── copy_test.go            # Copy operations and integration tests (12 specs)
├── concurrent_test.go      # Concurrent safety and race condition tests (11 specs)
├── mode_test.go            # Adaptive mode and strategy switching tests (7 specs)
├── edge_cases_test.go      # Edge cases and error handling tests (25 specs)
├── benchmark_test.go       # Performance benchmarks with gmeasure (8 aggregated experiments)
├── helper_test.go          # Shared test helpers and utilities
└── example_test.go         # Runnable examples for GoDoc
```

**File Purpose Alignment:**

Each test file has a **specific, non-overlapping scope** aligned with ISTQB test organization principles:

| File | Primary Responsibility | Unique Scope | Justification |
|------|------------------------|--------------|---------------|
| **suite_test.go** | Test suite bootstrap | Ginkgo suite initialization only | Required entry point for BDD tests |
| **constructor_test.go** | Object creation & interfaces | New(), DefaultConfig(), interface compliance, DiscardCloser | Unit tests for factory methods and type compliance |
| **reader_test.go** | Input operations | SetInput(), Read(), Reader(), Close() on input side | Isolated tests for read path and input lifecycle |
| **writer_test.go** | Output operations | AddWriter(), Write(), WriteString(), Clean(), Writer() | Isolated tests for write path and writer management |
| **copy_test.go** | Integration workflows | Copy() method and mixed read/write scenarios | Integration tests combining multiple operations |
| **concurrent_test.go** | Thread-safety | Race detection, concurrent access patterns | Validates atomicity and thread-safety guarantees |
| **mode_test.go** | Adaptive behavior | Mode switching, IsParallel(), IsSequential(), IsAdaptive() | Tests adaptive strategy and mode detection |
| **edge_cases_test.go** | Boundary & error cases | Nil handling, empty/large data, error propagation | Negative testing and boundary value analysis |
| **benchmark_test.go** | Performance metrics | **Aggregated experiments** with systematic variations | Non-functional performance validation using gmeasure |
| **helper_test.go** | Test infrastructure | errorReader, errorWriter, slowWriter utilities | Shared test doubles (not executable tests) |
| **example_test.go** | Documentation | 16 runnable GoDoc examples | Documentation via executable examples (not counted in 120 specs) |

**Benchmark Organization (Following ioutils/delim Pattern):**

The benchmark file uses **aggregated experiments** instead of fragmented individual tests:

| Experiment Group | Variations Tested | Sample Count | Purpose |
|------------------|-------------------|--------------|---------|
| **Write operations** | 6 variations (1/3 writers × small/1KB/1MB data) | 1000/100 | Measure write throughput with varying load |
| **Read operations** | 3 variations (100B/1KB/1MB data) | 1000/100 | Measure read performance across data sizes |
| **Copy operations** | 6 variations (1/3 writers × small/1KB/1MB data) | 1000/100 | Measure copy efficiency with broadcasting |
| **Mode comparison** | 3 variations (Sequential/Parallel/Adaptive) | 1000 | Compare strategy performance |
| **Writer management** | 7 operations (Constructor, AddWriter, Clean, SetInput, WriteString) | 1000 | Measure management overhead |
| **Real-world scenarios** | 3 scenarios (Log broadcasting, Stream replication, Adaptive load) | 10 | Validate real-world performance |

**Total**: **6 aggregated experiments** containing **28 systematic variations** (vs 15 fragmented tests before refactoring)

**Non-Redundancy Verification:**

- ✅ **No overlap** between reader_test.go (input) and writer_test.go (output) - separate I/O paths
- ✅ **copy_test.go is justified** - tests integration of read+write, not covered by isolated read/write tests
- ✅ **concurrent_test.go is unique** - only file using race detector, tests concurrent scenarios not tested elsewhere
- ✅ **mode_test.go is specific** - only file testing adaptive strategy switching logic
- ✅ **edge_cases_test.go is distinct** - focuses on error paths and boundaries, complementing happy-path tests
- ✅ **benchmark_test.go is non-functional** - performance testing is separate concern from correctness
- ✅ **helper_test.go is infrastructure** - provides test utilities, contains no executable tests
- ✅ **example_test.go is documentation** - GoDoc examples are separate from test specs

**Total Specs Distribution:**
- **Unit Tests** (constructor, reader, writer): 50 specs (42%)
- **Integration Tests** (copy): 12 specs (10%)
- **Concurrent Tests**: 11 specs (9%)
- **Mode/Adaptive Tests**: 7 specs (6%)
- **Edge/Boundary Tests**: 25 specs (21%)
- **Performance Tests**: 15 specs (12%)
- **Total**: **120 specs** across 8 test files

All test files are **necessary and justified** - no redundant files identified.

### Test Templates

**Basic Unit Test:**

```go
var _ = Describe("Multi", func() {
    var m multi.Multi

    BeforeEach(func() {
        m = multi.New(false, false, multi.DefaultConfig())
    })

    AfterEach(func() {
        m.Close()
    })

    It("should write data", func() {
        var buf bytes.Buffer
        m.AddWriter(&buf)
        n, err := m.Write([]byte("test"))
        Expect(err).ToNot(HaveOccurred())
        Expect(n).To(Equal(4))
        Expect(buf.String()).To(Equal("test"))
    })
})
```

### Running New Tests

```bash
# Focus on specific test
go test -ginkgo.focus="should write data" -v

# Run new test file
go test -v -run TestMulti/NewFeature
```

### Helper Functions

-   `newSlowWriter(delay)`: Creates a writer that sleeps to simulate latency.
-   `newErrorWriter(err)`: Creates a writer that always fails.
-   `newTestReader(data)`: Creates a reader with known data.

### Benchmark Template

**Aggregated Experiment Pattern (Recommended):**

```go
It("should benchmark operation with variations", func() {
    experiment := gmeasure.NewExperiment("Operation name")
    AddReportEntry(experiment.Name, experiment)

    // Variation 1
    experiment.SampleDuration("Small data", func(idx int) {
        // Test code here
    }, gmeasure.SamplingConfig{N: 1000, Duration: 0})

    // Variation 2
    experiment.SampleDuration("Large data", func(idx int) {
        // Test code here
    }, gmeasure.SamplingConfig{N: 100, Duration: 0})
})
```

**Real-world Scenario Pattern:**

```go
It("should benchmark real scenario", func() {
    experiment := gmeasure.NewExperiment("Scenario name")

    experiment.Sample(func(idx int) {
        // Setup
        experiment.MeasureDuration("operation", func() {
            // Actual operation
        })
    }, gmeasure.SamplingConfig{N: 10, Duration: 0})

    AddReportEntry(experiment.Name, experiment)
})
```

### Best Practices

-   ✅ **Use Atomic Helpers**: Verify state changes with `Eventually` in concurrent tests.
-   ✅ **Clean Up**: Always `Close()` the Multi instance.
-   ✅ **Test Both Modes**: Verify logic in both Sequential and Parallel modes.
-   ❌ **Avoid Sleep**: Use synchronization primitives or `Eventually` instead of `time.Sleep`.

---

## Troubleshooting

### Common Issues

**1. Race Conditions**
-   *Symptom*: `WARNING: DATA RACE`
-   *Fix*: Ensure all shared state access goes through the Multi API (which uses atomics) or is guarded by tests.

**2. Flaky Adaptive Tests**
-   *Symptom*: Mode doesn't switch when expected.
-   *Fix*: Increase the simulated latency gap between threshold and actual writer latency. Increase `SampleWrite` count in test config.

**3. Coverage Gaps**
-   *Symptom*: Low coverage in `writer.go`.
-   *Fix*: Add tests with `newErrorWriter` to trigger error paths in parallel execution.

---

## Reporting Bugs & Vulnerabilities

### Bug Report Template

When reporting a bug in the test suite or the multi package, please use this template:

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
[e.g., interface.go, model.go, specific function]

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

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for test generation, debugging, and documentation under human supervision. All tests are validated and reviewed by humans.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Test Suite Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)
**Package**: `github.com/nabbar/golib/ioutils/multi`

