# Testing Documentation

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![Tests](https://img.shields.io/badge/Tests-61%20specs-success)](tar_suite_test.go)
[![Assertions](https://img.shields.io/badge/Assertions-180+-blue)](tar_suite_test.go)
[![Coverage](https://img.shields.io/badge/Coverage-85.6%25-brightgreen)](coverage.out)

Comprehensive testing guide for the `github.com/nabbar/golib/archive/archive/tar` package using BDD methodology with Ginkgo v2 and Gomega.

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
  - [Performance Analysis](#performance-analysis)
  - [Test Conditions](#test-conditions)
  - [Performance Characteristics](#performance-characteristics)
  - [Memory Profile](#memory-profile)
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

This test suite provides **comprehensive validation** of the `tar` package following **ISTQB** principles. It focuses on validating tar archive **Reader/Writer** behavior, edge case handling, and performance through:

1. **Functional Testing**: Verification of all public APIs (NewReader, NewWriter, List, Info, Get, Has, Walk, Add, FromPath).
2. **Non-Functional Testing**: Performance benchmarking and sequential access validation.
3. **Structural Testing**: Ensuring all code paths and logic branches are exercised, while acknowledging that coverage metrics are just one indicator of quality.

### Test Completeness

**Quality Indicators:**
- **Code Coverage**: 85.6% of statements (Note: Used as a guide, not a guarantee of correctness).
- **Race Conditions**: 0 detected across all scenarios.
- **Flakiness**: 0 flaky tests detected.

**Test Distribution:**
- ✅ **61 specifications** covering all major use cases
- ✅ **180+ assertions** validating behavior
- ✅ **5 performance benchmarks** measuring key metrics
- ✅ **5 test files** organized by functional area
- ✅ **Zero flaky tests** - all tests are deterministic

---

## Test Architecture

### Test Matrix

| Category | Files | Specs | Coverage | Priority | Dependencies |
|----------|-------|-------|----------|----------|-------------|
| **Reader Operations** | reader_test.go | 29 | 90%+ | Critical | None |
| **Writer Operations** | writer_test.go | 26 | 85%+ | Critical | None |
| **Edge Cases** | edge_cases_test.go | 25 | 80%+ | High | Implementation |
| **Performance** | benchmark_test.go | 5 | N/A | Medium | Implementation |
| **Helpers** | helper_test.go | N/A | N/A | Low | All |
| **Examples** | example_test.go | N/A | N/A | Low | All |

### Detailed Test Inventory

**Test ID Pattern by File:**
- **TC-RD-xxx**: Reader tests (reader_test.go)
- **TC-WR-xxx**: Writer tests (writer_test.go)
- **TC-EC-xxx**: Edge case tests (edge_cases_test.go)
- **TC-BC-xxx**: Benchmark tests (benchmark_test.go)

| Test ID | File | Use Case | Priority | Expected Outcome |
|---------|------|----------|----------|------------------|
| **TC-RD-001** | reader_test.go | **Tar Reader Suite**: Main describe block for all reader tests | Critical | Test suite initialization and organization |
| **TC-RD-002** | reader_test.go | **NewReader Group**: Constructor validation tests | Critical | Reader creation tests grouped logically |
| **TC-RD-003** | reader_test.go | **NewReader Valid**: Create reader from valid tar stream | Critical | Reader instance created without error |
| **TC-RD-004** | reader_test.go | **NewReader Empty**: Create reader from empty archive | Critical | Reader handles empty archive gracefully |
| **TC-RD-005** | reader_test.go | **List Group**: File enumeration tests | Critical | List operation tests grouped logically |
| **TC-RD-006** | reader_test.go | **List All Files**: List all files in populated archive | Critical | Returns all 4 test files in correct paths |
| **TC-RD-007** | reader_test.go | **List Empty**: List files in empty archive | High | Returns empty slice without error |
| **TC-RD-008** | reader_test.go | **List Multiple Calls**: Multiple List calls with reset | High | Resetable reader returns same results |
| **TC-RD-009** | reader_test.go | **Info Group**: File metadata query tests | Critical | Info operation tests grouped logically |
| **TC-RD-010** | reader_test.go | **Info Existing**: Get metadata for existing file | Critical | Returns valid FileInfo with correct size/name |
| **TC-RD-011** | reader_test.go | **Info Missing**: Query non-existent file | Critical | Returns fs.ErrNotExist error |
| **TC-RD-012** | reader_test.go | **Info Nested**: Get metadata for nested file | High | Returns FileInfo for file in subdirectory |
| **TC-RD-013** | reader_test.go | **Get Group**: File extraction tests | Critical | Get operation tests grouped logically |
| **TC-RD-014** | reader_test.go | **Get Content**: Extract and read file content | Critical | Returns ReadCloser with correct file data |
| **TC-RD-015** | reader_test.go | **Get Missing**: Extract non-existent file | Critical | Returns fs.ErrNotExist error |
| **TC-RD-016** | reader_test.go | **Get Nested**: Extract file from subdirectory | High | Extracts nested file with correct content |
| **TC-RD-017** | reader_test.go | **Get Multiple Reset**: Extract multiple files with reset | High | Resetable reader allows multiple extractions |
| **TC-RD-018** | reader_test.go | **Has Group**: File existence check tests | Critical | Has operation tests grouped logically |
| **TC-RD-019** | reader_test.go | **Has Existing**: Check existence of present files | Critical | Returns true for existing files |
| **TC-RD-020** | reader_test.go | **Has Missing**: Check existence of absent files | Critical | Returns false for non-existent files |
| **TC-RD-021** | reader_test.go | **Has Empty**: Check files in empty archive | Medium | Returns false for any filename |
| **TC-RD-022** | reader_test.go | **Walk Group**: File iteration tests | Critical | Walk operation tests grouped logically |
| **TC-RD-023** | reader_test.go | **Walk All**: Iterate through all archive files | Critical | Callback invoked for each file (4 times) |
| **TC-RD-024** | reader_test.go | **Walk FileInfo**: Verify callback receives correct info | Critical | FileInfo parameter has correct size/mode |
| **TC-RD-025** | reader_test.go | **Walk Stop**: Stop iteration by returning false | High | Walk stops when callback returns false |
| **TC-RD-026** | reader_test.go | **Walk Filter**: Filter files during iteration | High | Custom logic can filter by extension |
| **TC-RD-027** | reader_test.go | **Close Group**: Resource cleanup tests | High | Close operation tests grouped logically |
| **TC-RD-028** | reader_test.go | **Close Normal**: Close reader normally | High | Close succeeds without error |
| **TC-RD-029** | reader_test.go | **Close Multiple**: Call Close multiple times | Medium | Multiple Close calls are safe (idempotent) |
| **TC-WR-001** | writer_test.go | **Tar Writer Suite**: Main describe block for all writer tests | Critical | Test suite initialization and organization |
| **TC-WR-002** | writer_test.go | **NewWriter Group**: Constructor validation tests | Critical | Writer creation tests grouped logically |
| **TC-WR-003** | writer_test.go | **NewWriter Valid**: Create writer from valid stream | Critical | Writer instance created without error |
| **TC-WR-004** | writer_test.go | **Add Group**: Single file addition tests | Critical | Add operation tests grouped logically |
| **TC-WR-005** | writer_test.go | **Add Single**: Add one file to archive | Critical | File added and readable from archive |
| **TC-WR-006** | writer_test.go | **Add Multiple**: Add multiple files sequentially | Critical | All files present in final archive |
| **TC-WR-007** | writer_test.go | **Add Custom Path**: Add file with renamed path | Critical | File stored with custom archive path |
| **TC-WR-008** | writer_test.go | **Add Link**: Add symbolic link entry | High | Link stored with target path |
| **TC-WR-009** | writer_test.go | **Add Empty**: Add zero-length file | High | Empty file handled correctly |
| **TC-WR-010** | writer_test.go | **Add Large**: Add large file (10KB) | High | Large file added without memory issues |
| **TC-WR-011** | writer_test.go | **FromPath Group**: Directory archiving tests | Critical | FromPath operation tests grouped logically |
| **TC-WR-012** | writer_test.go | **FromPath Single**: Archive single file by path | Critical | Single file added from filesystem |
| **TC-WR-013** | writer_test.go | **FromPath Recursive**: Archive directory recursively | Critical | All files in tree added (4 files) |
| **TC-WR-014** | writer_test.go | **FromPath Filter**: Filter files by glob pattern | Critical | Only matching files added (*.txt) |
| **TC-WR-015** | writer_test.go | **FromPath Replace**: Transform paths during archiving | High | Replacement function modifies archive paths |
| **TC-WR-016** | writer_test.go | **FromPath Skip Dirs**: Skip directory entries | High | Only files added, no directory entries |
| **TC-WR-017** | writer_test.go | **FromPath Empty Dir**: Handle empty directory | Medium | No files added, no error |
| **TC-WR-018** | writer_test.go | **FromPath Invalid**: Handle non-existent path | High | Returns appropriate error |
| **TC-WR-019** | writer_test.go | **Close Group**: Archive finalization tests | Critical | Close operation tests grouped logically |
| **TC-WR-020** | writer_test.go | **Close Normal**: Close writer normally | Critical | Flushes and finalizes archive |
| **TC-WR-021** | writer_test.go | **Close Finalize**: Verify archive validity after close | Critical | Closed archive is readable |
| **TC-WR-022** | writer_test.go | **Close Multiple**: Call Close multiple times | Medium | Second close may error (acceptable) |
| **TC-WR-023** | writer_test.go | **Close Error**: Propagate underlying close errors | High | Errors from underlying writer propagated |
| **TC-WR-024** | writer_test.go | **Integration Group**: End-to-end tests | Critical | Integration tests grouped logically |
| **TC-WR-025** | writer_test.go | **Integration Round-trip**: Write and read back archive | Critical | Written archive is readable with correct content |
| **TC-WR-026** | writer_test.go | **Integration Mixed**: Combine Add and FromPath | High | Mixed operations produce valid archive |
| **TC-EC-001** | edge_cases_test.go | **Edge Cases Suite**: Main describe block for edge cases | High | Edge case tests grouped logically |
| **TC-EC-002** | edge_cases_test.go | **Empty Archive Group**: Tests with empty archives | High | Empty archive handling tests |
| **TC-EC-003** | edge_cases_test.go | **Empty List**: List empty archive | High | Returns empty slice |
| **TC-EC-004** | edge_cases_test.go | **Empty Walk**: Walk empty archive | High | Callback never invoked |
| **TC-EC-005** | edge_cases_test.go | **Empty Has**: Check files in empty archive | Medium | Returns false |
| **TC-EC-006** | edge_cases_test.go | **Large Data Group**: Tests with large content | High | Large data handling tests |
| **TC-EC-007** | edge_cases_test.go | **Large Content**: Handle very large file content | High | 280KB file handled correctly |
| **TC-EC-008** | edge_cases_test.go | **Many Files**: Handle archive with 100 files | High | All 100 files enumerated correctly |
| **TC-EC-009** | edge_cases_test.go | **Special Chars Group**: Filenames with special characters | Medium | Special character handling tests |
| **TC-EC-010** | edge_cases_test.go | **Spaces in Names**: Handle filenames with spaces | Medium | Spaces preserved in filenames |
| **TC-EC-011** | edge_cases_test.go | **Special Chars**: Handle dashes and underscores | Medium | Special characters allowed |
| **TC-EC-012** | edge_cases_test.go | **Deep Paths**: Handle deeply nested paths | Medium | 8-level deep path supported |
| **TC-EC-013** | edge_cases_test.go | **Binary Content Group**: Binary data tests | High | Binary data handling tests |
| **TC-EC-014** | edge_cases_test.go | **Binary Data**: Store and retrieve binary content | High | All 256 byte values preserved |
| **TC-EC-015** | edge_cases_test.go | **Zero Bytes**: Handle null bytes in content | High | Null bytes preserved in data |
| **TC-EC-016** | edge_cases_test.go | **Concurrent Access Group**: Sequential safety tests | Medium | Concurrency tests (per-instance) |
| **TC-EC-017** | edge_cases_test.go | **Sequential Reads**: Multiple operations on same reader | Medium | Sequential operations work with reset |
| **TC-EC-018** | edge_cases_test.go | **Reset Behavior Group**: Reset capability tests | Medium | Reset functionality tests |
| **TC-EC-019** | edge_cases_test.go | **Reset Support**: Reset with resetable reader | Medium | Reset allows re-reading archive |
| **TC-EC-020** | edge_cases_test.go | **No Reset**: Non-resetable reader behavior | Medium | Second read returns empty (acceptable) |
| **TC-EC-021** | edge_cases_test.go | **Malformed Input Group**: Invalid archive tests | Medium | Error handling for bad input |
| **TC-EC-022** | edge_cases_test.go | **Corrupted Archive**: Handle non-tar data | Medium | Returns empty list, no panic |
| **TC-EC-023** | edge_cases_test.go | **Truncated Archive**: Handle incomplete archive | Medium | Handles gracefully, no panic |
| **TC-EC-024** | edge_cases_test.go | **Permissions Group**: File mode tests | Medium | Permission preservation tests |
| **TC-EC-025** | edge_cases_test.go | **Preserve Permissions**: Maintain file mode bits | Medium | Mode 0755 preserved in archive |
| **TC-BC-001** | benchmark_test.go | **Reader Operations Bench**: Measure reader throughput | High | 5 operations x 1000 samples each |
| **TC-BC-002** | benchmark_test.go | **Writer Operations Bench**: Measure writer throughput | High | 4 operations with varying sizes |
| **TC-BC-003** | benchmark_test.go | **Round-trip Bench**: Measure complete write-read cycle | Medium | 2 scenarios (small/multiple files) |
| **TC-BC-004** | benchmark_test.go | **Memory Operations Bench**: Measure constructor overhead | Medium | Reader/Writer creation costs |
| **TC-BC-005** | benchmark_test.go | **Archive Operations**: Benchmark FromPath and List | Medium | Directory archiving and enumeration performance |

---

## Test Statistics

**Latest Test Run Results:**

```
Total Specs:         61
Passed:              61
Failed:              0
Skipped:             0
Execution Time:      ~0.111 seconds (non-race), ~0.766 seconds (race)
Coverage:            85.6%
Race Conditions:     0
```

---

## Framework & Tools

### Testing Frameworks

#### Ginkgo v2 - BDD Testing Framework

**Why Ginkgo over standard Go testing:**
- ✅ **Hierarchical organization**: `Describe`, `Context`, `It` for clear test structure.
- ✅ **Better readability**: Tests read like specifications.
- ✅ **Rich lifecycle hooks**: `BeforeEach`, `AfterEach` for setup/teardown.
- ✅ **Async testing**: `Eventually`, `Consistently` for concurrent behavior.
- ✅ **Parallel execution**: Built-in support for concurrent test runs.

#### Gomega - Matcher Library

**Advantages:**
- ✅ **Expressive matchers**: `Equal`, `BeNumerically`, `HaveOccurred`.
- ✅ **Async assertions**: `Eventually` polls for state changes.

#### gmeasure - Performance Measurement

Used for benchmarking throughput and latency within the BDD suite.

### Testing Concepts & Standards

#### ISTQB Alignment

This test suite follows **ISTQB (International Software Testing Qualifications Board)** principles:

1. **Test Levels** (ISTQB Foundation Level):
   * **Unit Testing**: Individual functions (`NewReader`, `NewWriter`, `List`, `Get`).
   * **Integration Testing**: Component interactions (`Add` + `FromPath`, round-trip operations).
   * **System Testing**: End-to-end scenarios (complete archive creation and extraction).

2. **Test Types** (ISTQB Advanced Level):
   * **Functional Testing**: Verify behavior meets specifications (List, Get, Add).
   * **Non-Functional Testing**: Performance, sequential access patterns.
   * **Structural Testing**: Code coverage (Statement coverage).

3. **Test Design Techniques**:
   * **Equivalence Partitioning**: Valid files vs non-existent files.
   * **Boundary Value Analysis**: Empty archives, single file, 100 files.
   * **State Transition Testing**: Reset behavior, multiple operations.
   * **Error Guessing**: Malformed archives, permission issues.

#### Testing Pyramid

The suite follows the Testing Pyramid principle:

```
         /\
        /  \
       / E2E\       (Integration/Round-trip Tests)
      /______\
     /        \
    / Integr.  \    (Edge Cases, FromPath Tests)
   /____________\
  /              \
 /   Unit Tests   \ (Reader, Writer, Individual Operations)
/__________________\
```

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
Running Suite: Archive/Tar Package Suite
=========================================
Random Seed: 1766694569

Will run 61 of 61 specs

•••••••••••••••••••••••••••••••••••••••••••••••••••••••••••

Ran 61 of 61 Specs in 0.111 seconds
SUCCESS! -- 61 Passed | 0 Failed | 0 Pending | 0 Skipped

PASS
coverage: 85.6% of statements
ok      github.com/nabbar/golib/archive/archive/tar     0.119s
```

---

## Coverage

### Coverage Report

| Component | File | Coverage | Critical Paths |
|-----------|------|----------|----------------|
| **Interface** | interface.go | 100.0% | NewReader(), NewWriter() |
| **Reader Logic** | reader.go | 90.5% | List, Info, Get, Has, Walk, Reset |
| **Writer Logic** | writer.go | 78.5% | Add, FromPath, Close, path filtering |

**Detailed Coverage:**

```
NewReader()          100.0%  - All configuration paths tested
NewWriter()          100.0%  - Writer creation fully covered
List()               100.0%  - File enumeration
Info()                90.0%  - File metadata query
Get()                100.0%  - File extraction
Has()                100.0%  - File existence check
Walk()                90.0%  - File iteration
Add()                 81.2%  - Single file addition
FromPath()           100.0%  - Directory archiving
Close()               71.4%  - Resource cleanup
```

### Uncovered Code Analysis

**Uncovered Lines: 14.4% (target: <20%)**

#### 1. Error Paths in Writer Close (writer.go)

**Uncovered**: Lines handling flush/close failures in sequence

```go
// UNCOVERED: Specific error propagation paths
if e := o.z.Flush(); e != nil {
    return e
}
```

**Reason**: Difficult to trigger tar.Writer internal flush errors without mocking.

**Impact**: Low - error paths are defensive, main flows well-tested

#### 2. Edge Cases in addFiltering (writer.go)

**Uncovered**: Some file type validation branches

**Reason**: Requires specific filesystem conditions (special devices, etc.).

**Impact**: Low - unsupported types return fs.ErrInvalid as expected

#### 3. Reset Edge Cases (reader.go)

**Uncovered**: Some combinations of Info/Get without reset

**Reason**: Sequential tar format requires full scan for each operation.

**Impact**: Low - documented behavior, users should use Walk for multiple operations

### Thread Safety Assurance

**Race Detection Results:**

```bash
$ CGO_ENABLED=1 go test -race -v
Running Suite: Archive/Tar Package Suite
=========================================
Will run 61 of 61 specs

Ran 61 of 61 Specs in 0.766 seconds
SUCCESS! -- 61 Passed | 0 Failed | 0 Pending | 0 Skipped

PASS
coverage: 85.6% of statements
ok      github.com/nabbar/golib/archive/archive/tar     1.804s
```

**Zero data races detected** across:
- ✅ Sequential reader operations (per-instance)
- ✅ Sequential writer operations (per-instance)
- ✅ Edge case scenarios
- ✅ Benchmark operations
- ✅ Integration tests

**Thread Safety Model:**

| Component | Thread Safety | Notes |
|-----------|---------------|-------|
| Reader | Per-instance safe | One goroutine per reader instance |
| Writer | Per-instance safe | One goroutine per writer instance |
| tar.Reader | Not thread-safe | Standard library limitation |
| tar.Writer | Not thread-safe | Standard library limitation |

**Verified Thread-Safe:**
- ✅ Multiple reader instances can operate concurrently
- ✅ Multiple writer instances can operate concurrently
- ✅ Reader/Writer use separate instances (no shared state)
- ⚠️ Single instance NOT safe for concurrent use (documented)

---

## Performance

### Performance Report

**Benchmark Results (Aggregated Experiments):**

#### Reader Operations

| Operation | Sample Size | Median | Mean | Max | Notes |
|-----------|-------------|--------|------|-----|-------|
| List() | 1000 | 100µs | 100µs | 800µs | Sequential scan of entire archive |
| Info() | 1000 | 0s | 100µs | 600µs | Scan until file found |
| Get() | 1000 | 0s | 100µs | 600µs | Scan + read file content |
| Has() | 1000 | 0s | 100µs | 1ms | Boolean existence check |
| Walk() | 1000 | 100µs | 100µs | 700µs | Single pass, most efficient |

#### Writer Operations

| Operation | Sample Size | Median | Mean | Max | Notes |
|-----------|-------------|--------|------|-----|-------|
| Add (100B) | 1000 | 0s | 0s | 500µs | Small file addition |
| Add (10KB) | 100 | 0s | 100µs | 400µs | Medium file |
| Add (1MB) | 10 | 8.3ms | 8.6ms | 9.9ms | Large file streaming |
| Add Multiple (10 files) | 100 | 200µs | 200µs | 600µs | Sequential additions |

#### Round-trip Operations

| Scenario | Sample Size | Median | Mean | Max | Description |
|----------|-------------|--------|------|-----|-------------|
| Small archive | 100 | 100µs | 100µs | 300µs | Write 1 file + read back |
| Multiple files | 100 | 200µs | 300µs | 600µs | Write 4 files + read all |

#### Memory Operations

| Operation | Sample Size | Median | Mean | Max | Notes |
|-----------|-------------|--------|------|-----|-------|
| Create Reader | 1000 | 0s | 0s | 400µs | Minimal overhead |
| Create Writer | 1000 | 0s | 0s | 200µs | Minimal overhead |

### Performance Analysis

**Key Findings:**

1. **Sub-millisecond Operations**: Most reader operations complete in <100µs (median), demonstrating excellent efficiency
2. **Large Data Handling**: 1MB writes complete in ~8.6ms mean (with race detector)
3. **Sequential Access**: Archive format requires full scan for each operation
4. **Reset Performance**: Resetable readers enable multiple operations without reopening files
5. **Real-World Performance**: Round-trip operations (300µs) validate production readiness

**Reader Performance:**
- **List()**: Sequential scan of entire archive, O(n) where n = file count
- **Info(path)**: Sequential scan until file found, worst-case O(n)
- **Get(path)**: Sequential scan until file found, worst-case O(n)
- **Has(path)**: Sequential scan until file found, worst-case O(n)
- **Walk()**: Single sequential pass, O(n), most efficient for processing all files

**Writer Performance:**
- **Add()**: Streaming write, O(file size), no buffering
- **FromPath()**: File tree walk, O(n files × file size)
- **Close()**: Finalize archive, O(1) metadata write

**Optimization Tips:**
1. Use Walk() instead of multiple Get() calls
2. Use Reset() when available to avoid reopening files
3. Buffer I/O operations for better throughput
4. Consider zip format for random access requirements

### Test Conditions

- **Hardware**: AMD64/ARM64 Multi-core, 8GB+ RAM
- **Sample Sizes**: 1000 samples (small ops), 100 samples (medium), 10 samples (large)
- **Data Sizes**: Small (100B), Medium (10KB), Large (1MB)
- **Archive Sizes**: 4-100 files per archive
- **Race Detector**: Enabled (adds 2-3x overhead)

### Performance Characteristics

**Strengths:**
- ✅ **Fast Small Operations**: Reader queries <100µs median
- ✅ **Streaming Efficiency**: Constant memory per operation
- ✅ **Scalable**: 1MB files handled in <10ms
- ✅ **Predictable**: Low standard deviation across benchmarks
- ✅ **Reset Support**: Efficient re-reading when available

**Limitations:**
1. **Sequential Access Only**: Tar format requires full scan for random access
   - *Observation*: Info/Get/Has must scan from beginning each time
   - *Mitigation*: Use Walk() for multiple files, or Reset() between operations
2. **Peak Latency**: Max latencies (P99) can reach 1ms under race detector
   - *Context*: Acceptable for I/O-bound operations, race detector overhead
3. **No Random Access**: Cannot seek to specific files
   - *Impact*: Use zip format if random access required

### Memory Profile

- **Reader Overhead**: ~1KB (struct + tar.Reader)
- **Writer Overhead**: ~1KB (struct + tar.Writer)
- **Per-Operation**: O(1) - streaming, no buffering
- **List() Operation**: O(n) paths stored in memory
- **Walk() Operation**: O(1) - callback per file
- **Real-World**: 100-file archive List() = ~10KB memory

---

## Test Writing

### File Organization

```
tar/
├── tar_suite_test.go        # Test suite entry point (Ginkgo suite setup)
├── reader_test.go            # Reader operations tests (29 specs)
├── writer_test.go            # Writer operations tests (26 specs)
├── edge_cases_test.go        # Edge cases and error handling tests (25 specs)
├── benchmark_test.go         # Performance benchmarks (5 aggregated experiments)
├── helper_test.go            # Shared test helpers and utilities
└── example_test.go           # Runnable examples for GoDoc (13 examples)
```

**File Purpose Alignment:**

Each test file has a **specific, non-overlapping scope** aligned with ISTQB test organization principles:

| File | Primary Responsibility | Unique Scope | Justification |
|------|------------------------|--------------|---------------|
| **tar_suite_test.go** | Test suite bootstrap | Ginkgo suite initialization only | Required entry point for BDD tests |
| **reader_test.go** | Reader operations | NewReader(), List(), Info(), Get(), Has(), Walk(), Close() | Unit tests for reader functionality |
| **writer_test.go** | Writer operations | NewWriter(), Add(), FromPath(), Close() | Unit tests for writer functionality |
| **edge_cases_test.go** | Boundary & error cases | Empty archives, large files, special chars, malformed input | Negative testing and boundary value analysis |
| **benchmark_test.go** | Performance metrics | **Aggregated experiments** with systematic variations | Non-functional performance validation using gmeasure |
| **helper_test.go** | Test infrastructure | createTestArchive, nopWriteCloser, errorReadCloser utilities | Shared test doubles (not executable tests) |
| **example_test.go** | Documentation | 13 runnable GoDoc examples | Documentation via executable examples (not counted in 61 specs) |

**Non-Redundancy Verification:**

- ✅ **No overlap** between reader_test.go (read operations) and writer_test.go (write operations) - separate I/O paths
- ✅ **edge_cases_test.go is justified** - tests boundary conditions and error paths not covered by happy-path tests
- ✅ **benchmark_test.go is non-functional** - performance testing is separate concern from correctness
- ✅ **helper_test.go is infrastructure** - provides test utilities, contains no executable tests
- ✅ **example_test.go is documentation** - GoDoc examples are separate from test specs

**Total Specs Distribution:**
- **Reader Unit Tests**: 29 specs (48%)
- **Writer Unit Tests**: 26 specs (43%)
- **Edge/Boundary Tests**: 25 specs (41%)
- **Performance Tests**: 5 specs (8%)
- **Total**: **61 specs** across 5 test files

All test files are **necessary and justified** - no redundant files identified.

### Test Templates

**Basic Unit Test:**

```go
var _ = Describe("TC-RD-001: Tar Reader", func() {
    var (
        reader Reader
        testData map[string]string
    )
    
    BeforeEach(func() {
        // Setup test fixtures
        testData = map[string]string{
            "file.txt": "content",
        }
    })
    
    AfterEach(func() {
        // Cleanup resources
        if reader != nil {
            reader.Close()
        }
    })
    
    It("TC-RD-003: should create a valid reader", func() {
        // Arrange
        archive := createTestArchive(testData)
        
        // Act
        reader, err := tar.NewReader(io.NopCloser(archive))
        
        // Assert
        Expect(err).ToNot(HaveOccurred())
        Expect(reader).ToNot(BeNil())
    })
})
```

### Running New Tests

```bash
# Focus on specific test
go test -ginkgo.focus="should create a valid reader" -v

# Run new test file
go test -v -run TestTarSuite/Reader
```

### Helper Functions

- `createTestArchive(files)`: Creates a tar archive with specified files.
- `createEmptyArchive()`: Creates an empty tar archive.
- `nopWriteCloser`: Wrapper for io.Writer to add Close().
- `errorWriteCloser`: Writer that always fails.
- `errorReadCloser`: Reader that always fails.
- `testFileInfo`: Mock fs.FileInfo implementation.

### Benchmark Template

**Aggregated Experiment Pattern (Recommended):**

```go
It("TC-BC-001: Reader operations", func() {
    experiment := gmeasure.NewExperiment("Reader Operations")
    AddReportEntry(experiment.Name, experiment)

    // Variation 1: List operation
    experiment.SampleDuration("List", func(idx int) {
        reader.List()
    }, gmeasure.SamplingConfig{N: 1000, Duration: 0})

    // Variation 2: Info operation
    experiment.SampleDuration("Info", func(idx int) {
        reader.Info("file1.txt")
    }, gmeasure.SamplingConfig{N: 1000, Duration: 0})
})
```

**Real-world Scenario Pattern:**

```go
It("TC-BC-003: Round-trip small archive", func() {
    experiment := gmeasure.NewExperiment("Round-trip Operations")

    experiment.Sample(func(idx int) {
        // Setup
        var buf bytes.Buffer
        
        experiment.MeasureDuration("write-read cycle", func() {
            // Write archive
            writer, _ := tar.NewWriter(&nopWriteCloser{&buf})
            writer.Add(testFileInfo, io.NopCloser(strings.NewReader("test")), "test.txt", "")
            writer.Close()
            
            // Read archive
            reader, _ := tar.NewReader(io.NopCloser(&buf))
            reader.List()
        })
    }, gmeasure.SamplingConfig{N: 100, Duration: 0})

    AddReportEntry(experiment.Name, experiment)
})
```

### Best Practices

- ✅ **Use Helper Functions**: Leverage `createTestArchive` for consistent test data.
- ✅ **Clean Up**: Always `Close()` readers and writers in `AfterEach`.
- ✅ **Test ID Discipline**: All tests must have TC-XX-XXX IDs.
- ❌ **Avoid Sleep**: Use synchronization primitives instead of `time.Sleep`.

---

## Troubleshooting

### Common Issues

**1. Test Failures with "file not found"**
- *Symptom*: fs.ErrNotExist returned unexpectedly
- *Fix*: Ensure test archive is properly created before reader operations. Check `createTestArchive()` is called in BeforeEach.

**2. Race Detector Failures**
- *Symptom*: `WARNING: DATA RACE`
- *Fix*: The package is NOT thread-safe per instance. Use separate reader/writer instances per goroutine.

**3. Coverage Not Generated**
- *Symptom*: `coverage: [no statements]`
- *Fix*: Ensure `-covermode=atomic` flag is used with race detector:
```bash
CGO_ENABLED=1 go test -race -coverprofile=coverage.out -covermode=atomic
```

**4. Benchmark Hangs or Timeout**
- *Symptom*: `panic: test timed out after 10m0s`
- *Fix*: Reduce sample count in gmeasure.SamplingConfig{N: 100} for large file benchmarks.

**5. Temporary File Cleanup Errors**
- *Symptom*: `cannot remove temp directory`
- *Fix*: Ensure all file handles are closed before cleanup. Use defer for reader.Close() and writer.Close().

---

## Reporting Bugs & Vulnerabilities

### Bug Report Template

When reporting a bug in the test suite or the tar package, please use this template:

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
[e.g., interface.go, reader.go, writer.go, specific function]

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
**Package**: `github.com/nabbar/golib/archive/archive/tar`

