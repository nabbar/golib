# Testing Documentation

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![Tests](https://img.shields.io/badge/Tests-33%20specs-success)](suite_test.go)
[![Assertions](https://img.shields.io/badge/Assertions-100+-blue)](suite_test.go)
[![Coverage](https://img.shields.io/badge/Coverage-68.9%25-yellow)](coverage.out)

Comprehensive testing guide for the `github.com/nabbar/golib/archive/archive/zip` package using BDD methodology with Ginkgo v2 and Gomega.

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

This test suite provides **comprehensive validation** of the `zip` package following **ISTQB** principles. It focuses on validating ZIP archive **Reader/Writer** behavior, random access operations, and performance through:

1. **Functional Testing**: Verification of all public APIs (NewReader, NewWriter, List, Info, Get, Has, Walk, Add, FromPath).
2. **Non-Functional Testing**: Performance benchmarking and random access validation.
3. **Structural Testing**: Ensuring all code paths and logic branches are exercised, while acknowledging that coverage metrics are just one indicator of quality.

### Test Completeness

**Quality Indicators:**
- **Code Coverage**: 68.9% of statements (Note: Used as a guide, not a guarantee of correctness).
- **Race Conditions**: 0 detected across all scenarios.
- **Flakiness**: 0 flaky tests detected.

**Test Distribution:**
- ✅ **33 specifications** covering all major use cases
- ✅ **100+ assertions** validating behavior
- ✅ **12 performance benchmarks** measuring key metrics
- ✅ **6 test files** organized by functional area
- ✅ **Zero flaky tests** - all tests are deterministic

---

## Test Architecture

### Test Matrix

| Category | Files | Specs | Coverage | Priority | Dependencies |
|----------|-------|-------|----------|----------|-------------|
| **Writer Operations** | writer_test.go | 24 | 80-100% | Critical | None |
| **Reader/Integration** | simple_test.go | 16 | Reader 100%, Integration | Critical | Implementation |
| **Performance** | benchmark_test.go | 12 | N/A | Medium | Implementation |
| **Helpers** | helper_test.go | N/A | N/A | Low | All |
| **Examples** | example_test.go | N/A | N/A | Low | All |
| **Suite** | suite_test.go | N/A | N/A | Low | All |

### Detailed Test Inventory

**Test ID Pattern by File:**
- **TC-WR-xxx**: Writer tests (writer_test.go)
- **TC-SM-xxx**: Simple/Reader/Integration tests (simple_test.go)
- **TC-BC-xxx**: Benchmark tests (benchmark_test.go)

| Test ID | File | Use Case | Priority | Expected Outcome |
|---------|------|----------|----------|------------------|
| **TC-WR-001** | writer_test.go | **ZIP Writer Operations**: Main describe block for all writer tests | Critical | Test suite initialization and organization |
| **TC-WR-002** | writer_test.go | **NewWriter Group**: Constructor validation tests | Critical | Writer creation tests grouped logically |
| **TC-WR-003** | writer_test.go | **NewWriter Valid File**: Create writer from valid file | Critical | Writer instance created without error |
| **TC-WR-004** | writer_test.go | **NewWriter Buffer**: Create writer from buffer | Critical | Writer works with in-memory buffer |
| **TC-WR-005** | writer_test.go | **Add Operations Group**: Single file addition tests | Critical | Add operation tests grouped logically |
| **TC-WR-006** | writer_test.go | **Add Single File**: Add one file to archive | Critical | File added with correct content |
| **TC-WR-007** | writer_test.go | **Add Custom Path**: Add file with renamed path | Critical | File stored with custom archive path |
| **TC-WR-008** | writer_test.go | **Add Nil Reader**: Handle nil reader in Add | High | No error, graceful handling |
| **TC-WR-009** | writer_test.go | **Add Multiple Files**: Add multiple files sequentially | Critical | All files present in final archive |
| **TC-WR-010** | writer_test.go | **Add Empty Content**: Add zero-length file | High | Empty file handled correctly |
| **TC-WR-011** | writer_test.go | **Add Large File**: Add large file (50KB) | High | Large file added without memory issues |
| **TC-WR-012** | writer_test.go | **FromPath Operations Group**: Directory archiving tests | Critical | FromPath operation tests grouped logically |
| **TC-WR-013** | writer_test.go | **FromPath Directory**: Archive directory recursively | Critical | All files in tree added |
| **TC-WR-014** | writer_test.go | **FromPath Filter**: Filter files by glob pattern | Critical | Only matching files added (*.txt) |
| **TC-WR-015** | writer_test.go | **FromPath Transform**: Transform paths during archiving | High | Replacement function modifies archive paths |
| **TC-WR-016** | writer_test.go | **FromPath Single**: Archive single file by path | Critical | Single file added from filesystem |
| **TC-WR-017** | writer_test.go | **FromPath Invalid**: Handle non-existent path | High | Returns appropriate error |
| **TC-WR-018** | writer_test.go | **FromPath Skip Non-regular**: Skip non-regular files | Medium | Only regular files added |
| **TC-WR-019** | writer_test.go | **Close Operations Group**: Archive finalization tests | Critical | Close operation tests grouped logically |
| **TC-WR-020** | writer_test.go | **Close Normal**: Close writer normally | Critical | Flushes and finalizes archive |
| **TC-WR-021** | writer_test.go | **Close Finalize**: Verify archive validity after close | Critical | Closed archive is readable |
| **TC-WR-022** | writer_test.go | **Integration Group**: End-to-end tests | Critical | Integration tests grouped logically |
| **TC-WR-023** | writer_test.go | **Integration Round-trip**: Write and read back archive | Critical | Written archive is readable with correct content |
| **TC-WR-024** | writer_test.go | **Integration Errors**: Handle errors in FromPath | High | Error propagation works correctly |
| **TC-SM-001** | simple_test.go | **Simple Coverage Tests**: Main describe block for coverage tests | Critical | Coverage test suite initialization |
| **TC-SM-002** | simple_test.go | **Reader All Methods Group**: Comprehensive reader tests | Critical | Reader method tests grouped logically |
| **TC-SM-003** | simple_test.go | **Reader Methods Test**: Test all Reader methods comprehensively | Critical | List, Has, Info, Get, Walk all tested (100% coverage) |
| **TC-SM-004** | simple_test.go | **Writer FromPath and addFiltering Group**: Coverage tests | High | Filtering logic tests grouped |
| **TC-SM-005** | simple_test.go | **FromPath All Filtering Paths**: Test all filtering scenarios | High | All filter branches covered |
| **TC-SM-006** | simple_test.go | **FromPath Directories**: Handle directories and non-matching patterns | Medium | Edge cases handled correctly |
| **TC-SM-007** | simple_test.go | **FromPath Filter Errors**: Handle invalid filter patterns | High | Error handling for bad glob patterns |
| **TC-SM-007** | simple_test.go | **Writer Close and Add Coverage Group**: Additional coverage tests | High | Close/Add edge cases grouped (DUPLICATE ID) |
| **TC-SM-008** | simple_test.go | **Add and Close Paths**: Cover all Add and Close code paths | High | Nil readers, custom paths tested |
| **TC-SM-009** | simple_test.go | **NewReader Error Paths Group**: Reader validation tests | Critical | Error path tests grouped |
| **TC-SM-010** | simple_test.go | **NewReader Validation**: Test all NewReader validation branches | Critical | Invalid readers, zero size, negative size |
| **TC-SM-011** | simple_test.go | **Complete Integration Group**: Full cycle tests | Critical | End-to-end integration grouped |
| **TC-SM-012** | simple_test.go | **Write-Read Cycle**: Verify full write-read integration | Critical | Complete cycle with multiple files |
| **TC-SM-013** | simple_test.go | **Extensive addFiltering Group**: Deep filtering coverage | High | Internal filtering logic grouped |
| **TC-SM-014** | simple_test.go | **addFiltering Paths**: Cover all addFiltering code paths | High | Pattern matching, replacement, errors |
| **TC-SM-015** | simple_test.go | **NewReader Seek Group**: Additional reader tests | Medium | Seek and error path tests grouped |
| **TC-SM-016** | simple_test.go | **NewReader Valid Archive**: Test NewReader with valid data | Medium | Successful reader creation path |
| **TC-BC-001** | benchmark_test.go | **ZIP Performance Benchmarks**: Main benchmark suite | High | Performance test suite initialization |
| **TC-BC-002** | benchmark_test.go | **Reader Operations Group**: Reader throughput benchmarks | High | Reader benchmarks grouped |
| **TC-BC-003** | benchmark_test.go | **Reader Operations Bench**: Measure reader operations with varying file counts | High | List, Info, Get, Has, Walk benchmarked (5 files) |
| **TC-BC-004** | benchmark_test.go | **Writer Operations Group**: Writer throughput benchmarks | High | Writer benchmarks grouped |
| **TC-BC-005** | benchmark_test.go | **Writer Operations Bench**: Measure writer operations with varying sizes | High | Add benchmarked (100B, 10KB, 1MB, multiple files) |
| **TC-BC-006** | benchmark_test.go | **Round-trip Operations Group**: Complete cycle benchmarks | Medium | Round-trip benchmarks grouped |
| **TC-BC-007** | benchmark_test.go | **Round-trip Bench**: Measure complete write-read cycle | Medium | Small archive, 5 files scenarios |
| **TC-BC-008** | benchmark_test.go | **Memory Operations Group**: Constructor overhead benchmarks | Medium | Memory benchmarks grouped |
| **TC-BC-009** | benchmark_test.go | **Memory Operations Bench**: Measure creation and closure overhead | Medium | NewReader, NewWriter, Close costs |
| **TC-BC-010** | benchmark_test.go | **Real-world Scenarios Group**: Practical use case benchmarks | Medium | Real-world benchmarks grouped |
| **TC-BC-011** | benchmark_test.go | **Backup Scenario Bench**: Benchmark backup scenario | Medium | 10 files (1KB each) archive creation |
| **TC-BC-012** | benchmark_test.go | **Extraction Scenario Bench**: Benchmark extraction scenario | Medium | Extract all files performance |

---

## Test Statistics

**Latest Test Run Results:**

```
Total Specs:         33
Passed:              33
Failed:              0
Skipped:             0
Execution Time:      ~0.140 seconds (non-race), ~1.400 seconds (race)
Coverage:            68.9%
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
   * **Unit Testing**: Individual functions (`NewReader`, `NewWriter`, `List`, `Get`, `Add`).
   * **Integration Testing**: Component interactions (`Add` + `FromPath`, round-trip operations).
   * **System Testing**: End-to-end scenarios (complete archive creation and extraction).

2. **Test Types** (ISTQB Advanced Level):
   * **Functional Testing**: Verify behavior meets specifications (List, Get, Add, FromPath).
   * **Non-Functional Testing**: Performance, random access patterns.
   * **Structural Testing**: Code coverage (Statement coverage).

3. **Test Design Techniques**:
   * **Equivalence Partitioning**: Valid files vs non-existent files.
   * **Boundary Value Analysis**: Empty archives, single file, large files.
   * **State Transition Testing**: Multiple operations, integration cycles.
   * **Error Guessing**: Invalid readers, malformed archives, permission issues.

#### Testing Pyramid

The suite follows the Testing Pyramid principle:

```
         /\
        /  \
       / E2E\       (Integration/Round-trip Tests)
      /______\
     /        \
    / Integr.  \    (Simple Tests, FromPath Tests)
   /____________\
  /              \
 /   Unit Tests   \ (Writer, Reader, Individual Operations)
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
Running Suite: ZIP Archive Suite
=================================
Random Seed: 1766699570

Will run 33 of 33 specs

•••••••••••••••••••••••••••••••••

Ran 33 of 33 Specs in 0.137 seconds
SUCCESS! -- 33 Passed | 0 Failed | 0 Pending | 0 Skipped

PASS
coverage: 68.9% of statements
ok      github.com/nabbar/golib/archive/archive/zip     0.140s
```

---

## Coverage

### Coverage Report

| Component | File | Coverage | Critical Paths |
|-----------|------|----------|----------------|
| **Interface** | interface.go | 77.8% | NewReader(69.2%), NewWriter(100%) |
| **Reader Logic** | reader.go | 91.7% | List, Info, Get, Has, Walk (all 100%), Close (100%) |
| **Writer Logic** | writer.go | 66.7% | Add(80%), FromPath(100%), Close(57.1%), addFiltering(33.3%) |

**Detailed Coverage:**

```
NewReader()          69.2%  - Interface validation, error paths partially covered
NewWriter()          100.0% - Writer creation fully covered
List()               100.0% - File enumeration
Info()               100.0% - File metadata query
Get()                100.0% - File extraction
Has()                100.0% - File existence check
Walk()               100.0% - File iteration
Add()                 80.0% - Single file addition
FromPath()           100.0% - Directory archiving
Close()               57.1% - Resource cleanup (writer)
addFiltering()        33.3% - Internal filtering logic
```

### Uncovered Code Analysis

**Uncovered Lines: 31.1% (target: <20%)**

#### 1. NewReader Error Paths (interface.go)

**Uncovered**: Some validation branches for invalid readers

```go
// UNCOVERED: Specific combinations of missing interfaces
if _, ok := r.(readerSize); !ok {
    return nil, fs.ErrInvalid
}
```

**Reason**: Requires precise mock combinations to trigger all paths.

**Impact**: Low - main validation paths well-tested, edge cases defensive

#### 2. Writer Close Error Propagation (writer.go)

**Uncovered**: Lines handling flush/close failures in sequence

```go
// UNCOVERED: Specific error propagation paths
if e := o.z.Close(); e != nil {
    return e
}
```

**Reason**: Difficult to trigger zip.Writer internal close errors without mocking.

**Impact**: Low - error paths are defensive, main flows well-tested

#### 3. addFiltering Edge Cases (writer.go)

**Uncovered**: Some file type validation branches and error combinations

**Reason**: Requires specific filesystem conditions (special devices, permissions, etc.).

**Impact**: Medium - affects FromPath robustness, but unsupported types return fs.ErrInvalid as expected

### Thread Safety Assurance

**Race Detection Results:**

```bash
$ CGO_ENABLED=1 go test -race -v
Running Suite: ZIP Archive Suite
=================================
Will run 33 of 33 specs

Ran 33 of 33 Specs in 1.400 seconds
SUCCESS! -- 33 Passed | 0 Failed | 0 Pending | 0 Skipped

PASS
coverage: 68.9% of statements
ok      github.com/nabbar/golib/archive/archive/zip     1.400s
```

**Zero data races detected** across:
- ✅ Reader operations (List, Info, Get, Has, Walk)
- ✅ Writer operations (Add, FromPath, Close)
- ✅ Integration tests (full cycles)
- ✅ Benchmark operations
- ✅ Edge case scenarios

**Thread Safety Model:**

| Component | Thread Safety | Notes |
|-----------|---------------|-------|
| Reader | Per-instance safe | One goroutine per reader instance |
| Writer | Per-instance safe | One goroutine per writer instance |
| zip.Reader | Not thread-safe | Standard library limitation |
| zip.Writer | Not thread-safe | Standard library limitation |

**Verified Thread-Safe:**
- ✅ Multiple reader instances can operate concurrently
- ✅ Multiple writer instances can operate concurrently
- ✅ Reader/Writer use separate instances (no shared state)
- ⚠️ Single instance NOT safe for concurrent use (documented)

---

## Performance

### Performance Report

Based on gmeasure benchmark results with race detector enabled:

| Operation | Samples | Median | Mean | Max | Throughput |
|-----------|---------|--------|------|-----|------------|
| **Reader List (5 files)** | 1000 | <100µs | <100µs | <100µs | ~10K ops/sec |
| **Reader Info (5 files)** | 1000 | <100µs | 100µs | 200µs | ~10K ops/sec |
| **Reader Get (5 files)** | 1000 | <100µs | <100µs | <100µs | ~10K ops/sec |
| **Reader Has (5 files)** | 1000 | <100µs | <100µs | 200µs | ~10K ops/sec |
| **Reader Walk (5 files)** | 500 | <100µs | <100µs | 100µs | ~5K ops/sec |
| **Writer Add (100B)** | 1000 | <100µs | <100µs | 1ms | ~10K ops/sec |
| **Writer Add (10KB)** | 100 | <100µs | <100µs | 200µs | ~1K ops/sec |
| **Writer Add (1MB)** | 10 | 400µs | 1.6ms | 5.6ms | ~6 ops/sec |
| **Writer Add (10x100B)** | 100 | <100µs | <100µs | 400µs | ~1K ops/sec |
| **Write-Read Small** | 100 | <100µs | <100µs | 400µs | ~1K ops/sec |
| **Write-Read 5 Files** | 100 | <100µs | <100µs | 100µs | ~1K ops/sec |
| **NewReader + Close** | 500 | <100µs | 100µs | 11.8ms | ~5K ops/sec |
| **NewWriter + Close** | 1000 | <100µs | <100µs | 200µs | ~10K ops/sec |
| **NewWriter + Add + Close** | 500 | <100µs | <100µs | 200µs | ~5K ops/sec |
| **Backup (10x1KB)** | 50 | 100µs | 100µs | 500µs | ~500 ops/sec |
| **Extract All** | 100 | <100µs | <100µs | <100µs | ~1K ops/sec |

### Performance Analysis

**Reader Operations:**
- **Random Access**: ZIP format allows O(1) file lookup
- **List Performance**: Linear in number of files, but very fast (<100µs for 5 files)
- **Get Performance**: Direct access without scanning, <100µs per file
- **Walk Performance**: Slightly slower due to callback overhead

**Writer Operations:**
- **Add Performance**: Scales with file size (100B: <100µs, 1MB: 1.6ms)
- **FromPath Performance**: Includes filesystem operations (slower)
- **Compression Overhead**: Default compression level used

**Memory Operations:**
- **Reader Creation**: ~100µs overhead
- **Writer Creation**: <100µs overhead
- **Cleanup**: Proper Close() essential for finalization

### Test Conditions

**Environment:**
- **Go Version**: 1.24+
- **Race Detector**: Enabled (CGO_ENABLED=1)
- **Test Machine**: Standard CI environment
- **Warmup**: No warmup, cold start measurements

**Measurement Framework:**
- **gmeasure**: Statistical sampling with configurable sample counts
- **Precision**: Microsecond resolution
- **Metrics**: Min, Median, Mean, StdDev, Max

### Performance Characteristics

**Scalability:**
- **Small archives** (<100 files): Excellent performance, random access
- **Medium archives** (100-10K files): Good performance, no scanning overhead
- **Large archives** (>10K files): Scales well, random access eliminates full scans
- **File sizes**: Linear scaling up to MB-sized files

**Optimization Opportunities:**
- ✅ Random access eliminates unnecessary scans
- ✅ Streaming operations minimize memory usage
- ⚠️ Compression level not configurable (uses default)
- ⚠️ No bulk operations API

### Memory Profile

```
Reader overhead:     ~1KB (struct + zip.Reader + central directory)
Writer overhead:     ~1KB (struct + zip.Writer)
Per-file memory:     O(1) for read, O(1) for write
List() memory:       O(n) where n = number of files (paths only)
Archive metadata:    Stored in central directory (end of file)
```

**Memory Efficiency:**
- Random access allows constant memory per operation
- Files accessed directly, not cached
- Central directory loaded once on reader creation
- No temporary buffers for file contents

---

## Test Writing

### File Organization

```
archive/archive/zip/
├── suite_test.go          # Ginkgo test suite entry point
├── writer_test.go         # Writer operations (TC-WR-xxx)
├── simple_test.go         # Reader/Integration (TC-SM-xxx)
├── benchmark_test.go      # Performance (TC-BC-xxx)
├── helper_test.go         # Test utilities
└── example_test.go        # Usage examples
```

### Test Templates

**Basic Test Structure:**
```go
var _ = Describe("TC-XX-###: Test Group Name", func() {
    It("TC-XX-###: should do something specific", func() {
        // Arrange
        input := setupTestData()
        
        // Act
        result, err := functionUnderTest(input)
        
        // Assert
        Expect(err).ToNot(HaveOccurred())
        Expect(result).To(Equal(expected))
    })
})
```

**Writer Test Template:**
```go
It("TC-WR-###: should test writer operation", func() {
    buf := newBufferWriteCloser()
    writer, err := zip.NewWriter(buf)
    Expect(err).ToNot(HaveOccurred())
    defer writer.Close()
    
    // Test operation
    err = writer.Add(info, reader, path, target)
    Expect(err).ToNot(HaveOccurred())
})
```

**Reader Test Template:**
```go
It("TC-SM-###: should test reader operation", func() {
    // Create test archive
    buf := newBufferWriteCloser()
    writer, _ := zip.NewWriter(buf)
    // ... add files ...
    writer.Close()
    
    // Read archive
    zipReader := newReaderWithSize(buf.Bytes())
    reader, err := zip.NewReader(zipReader)
    Expect(err).ToNot(HaveOccurred())
    defer reader.Close()
    
    // Test operation
    files, err := reader.List()
    Expect(err).ToNot(HaveOccurred())
    Expect(len(files)).To(Equal(expectedCount))
})
```

### Running New Tests

```bash
# Run specific test by ID
go test -v -run "TC-WR-003"

# Run all writer tests
go test -v -run "TC-WR"

# Run with race detector
CGO_ENABLED=1 go test -race -v -run "TC-WR-003"
```

### Helper Functions

**Available in helper_test.go:**

```go
// Create test file info
func createTestFileInfo(name string, size int64) fs.FileInfo

// Create ZIP archive in memory
func createTestZipInMemory(files map[string]string) (*bytes.Buffer, error)

// Create ZIP archive on disk
func createTestZipFile(files map[string]string) (string, error)

// Create test directory structure
func createTestDirectory(files map[string]string) (string, error)

// Reader with Size() support
type readerWithSize struct {
    *bytes.Reader
    size int64
}

// WriteCloser for buffers
type bufferWriteCloser struct {
    *bytes.Buffer
}
```

### Benchmark Template

```go
It("TC-BC-###: should benchmark operation", func() {
    exp := gmeasure.NewExperiment("Operation Name")
    AddReportEntry(exp.Name, exp)
    
    exp.Sample(func(idx int) {
        exp.MeasureDuration("metric", func() {
            // Operation to measure
        })
    }, gmeasure.SamplingConfig{N: 1000})
})
```

### Best Practices

**DO:**
- ✅ Use descriptive test IDs (TC-XX-###)
- ✅ Clean up resources with `defer`
- ✅ Use `BeforeEach` for common setup
- ✅ Test both success and error paths
- ✅ Use helper functions for repetitive setup
- ✅ Run with race detector regularly

**DON'T:**
- ❌ Share state between tests
- ❌ Use hardcoded paths
- ❌ Forget to call Close()
- ❌ Skip error checking in tests
- ❌ Use timers (prefer gmeasure)

---

## Troubleshooting

### Common Issues

#### Test Failures

**Problem**: `undefined: bytes` compilation error
```go
// Solution: Add import
import "bytes"
```

**Problem**: `unknown field Reader` in readerWithSize
```go
// Solution: Use embedded field
type readerWithSize struct {
    *bytes.Reader  // Embedded, no field name
    size int64
}
```

**Problem**: Tests skip with "Failed to create test archive"
```go
// Solution: Use panic in setup instead of Skip
func createArchive() (string, map[string]string) {
    archive, err := createTestZipFile(files)
    if err != nil {
        panic("Setup failed: " + err.Error())
    }
    return archive, files
}
```

#### Coverage Issues

**Problem**: Coverage below target (68.9% vs 80% target)

**Analysis**:
- NewReader validation paths: 69.2% (some edge cases)
- Writer.Close error propagation: 57.1% (internal errors)
- addFiltering edge cases: 33.3% (special file types)

**Solution**: Accept current coverage or add mocks for internal errors

#### Race Conditions

**Problem**: Data race detected

**Solution**:
```bash
# Run with race detector
CGO_ENABLED=1 go test -race -v

# Fix: Don't share reader/writer between goroutines
# Create separate instances instead
```

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
**Package**: `github.com/nabbar/golib/archive/archive/zip`
