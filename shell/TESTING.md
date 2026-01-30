# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-329%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-~60%25-orange)]()

Comprehensive testing documentation for the shell package, covering test execution, race detection, and quality assurance across all subpackages.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Subpackage Testing](#subpackage-testing)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The shell package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite Summary**
- Total Specs: 329
- Coverage: ~60% (average across all subpackages)
- Race Detection: ✅ Zero data races
- Execution Time: ~1s (without race), ~2.5s (with race)

**Package Breakdown**

| Package | Specs | Coverage | Skipped | Status |
|---------|-------|----------|---------|--------|
| `shell` | 120 | 48.1% | 0 | ✅ All pass |
| `shell/command` | 93 | 81.8% | 0 | ✅ All pass |
| `shell/tty` | 116 | 44.7% | 10 | ✅ Terminal-dependent |
| **Total** | **329** | **~60%** | **10** | ✅ **Zero races** |

**Coverage Areas**
- Shell interface (Add, Run, Get, Desc, Walk, RunPrompt, ExitRegister)
- Command definition and execution
- TTY state management and signal handling
- Interactive prompt functionality
- Thread-safe concurrent operations
- Error handling and edge cases

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
ginkgo -cover -race
```

**Expected Output**
```bash
ok  	github.com/nabbar/golib/shell         0.089s	coverage: 48.1%
ok  	github.com/nabbar/golib/shell/command 0.023s	coverage: 81.8%
ok  	github.com/nabbar/golib/shell/tty     0.322s	coverage: 44.7%
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Parallel execution support
- Rich CLI with filtering

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages

**gmeasure** - Performance measurement ([docs](https://onsi.github.io/gomega/#gmeasure-benchmarking-code))
- Integrated benchmarking
- Statistical analysis
- Report generation

---

## Running Tests

### Basic Commands

```bash
# Standard test run
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test ./shell
go test ./shell/command
go test ./shell/tty

# With coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Timeout for long-running tests
go test -timeout=10m ./...
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific package
ginkgo ./command

# Pattern matching
ginkgo --focus="TTYSaver"

# Skip certain tests
ginkgo --skip="interactive"

# Parallel execution (not recommended for shell tests)
ginkgo -p

# JUnit report for CI
ginkgo --junit-report=results.xml

# Coverage with Ginkgo
ginkgo -cover
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# With coverage and race
CGO_ENABLED=1 go test -race -cover ./...

# Specific package with race
CGO_ENABLED=1 go test -race ./shell/tty
```

**Validates**:
- Atomic operations (`atomic.MapTyped`, `atomic.Value`, `atomic.Bool`)
- Mutex protection (`sync.Mutex`)
- Channel synchronization
- Signal handler goroutines
- Concurrent command execution

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/shell         1.669s
ok  	github.com/nabbar/golib/shell/command 0.045s
ok  	github.com/nabbar/golib/shell/tty     1.483s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
Write at 0x... by goroutine ...
```

**Status**: Zero data races detected across all packages

### Performance & Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out

# Memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out

# View benchmarks (gmeasure reports)
go test -v ./shell | grep "Report Entries"
```

**Performance Expectations**

| Test Type | Duration | Notes |
|-----------|----------|-------|
| Full Suite (all packages) | ~0.5s | Without race |
| With `-race` | ~2.5s | 4-5x slower (normal) |
| shell package | ~0.09s | 120 specs |
| command package | ~0.02s | 48 specs |
| tty package | ~0.32s | 116 specs (10 skipped) |
| Individual Spec | <10ms | Most tests |

---

## Test Coverage

**Overall Target**: ≥60% (achieved: ~60%)

### Coverage By Package

| Package | Coverage | Target | Files Tested |
|---------|----------|--------|--------------|
| `shell` | 48.1% | ≥45% | interface.go, model.go, goprompt.go |
| `shell/command` | 81.8% | ≥80% | interface.go, model.go |
| `shell/tty` | 44.7% | ≥45% | interface.go, model.go |

**Note**: Lower coverage in `shell` and `tty` is due to:
- RunPrompt() requires actual terminal (untestable in CI)
- Signal handling requires system signals
- Terminal restoration fallbacks (ANSI sequences)

### Coverage By Component

| Component | Coverage | Test Files |
|-----------|----------|------------|
| **Shell Core** | | |
| Command registration (Add) | 100% | `add_test.go` |
| Command execution (Run) | 72.7% | `walk_run_test.go` |
| Command retrieval (Get, Desc) | 100% | `get_desc_test.go` |
| Command walking (Walk) | 71.4% | `walk_run_test.go` |
| Exit handling (ExitRegister) | 100% | `prompt_test.go` |
| **Shell Interactive** | | |
| RunPrompt() setup | 0% | Requires terminal |
| Executor | 0% | Requires go-prompt |
| Completer | 0% | Requires go-prompt |
| **command Subpackage** | | |
| Command creation | 100% | `command_test.go` |
| Name/Describe | 100% | `command_test.go` |
| Run execution | 75% | `command_test.go` |
| **tty Subpackage** | | |
| New() constructor | 100% | `input_test.go` |
| IsTerminal() | 100% | `terminal_test.go` |
| Restore() | 75% | `restore_advanced_test.go` |
| Signal() | 66% | `signal_handling_test.go` |
| SignalHandler() | 0% | Requires system signals |

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal (function-level)
go tool cover -func=coverage.out

# Generate HTML report (line-level)
go tool cover -html=coverage.out -o coverage.html

# Per-package coverage
go test -coverprofile=shell.out ./shell
go test -coverprofile=command.out ./shell/command
go test -coverprofile=tty.out ./shell/tty
```

---

## Thread Safety

Thread safety is verified across all packages with race detection.

### Concurrency Primitives

```go
// Shell package
atomic.MapTyped[string, Command]  // Lock-free command registry
atomic.Value[tty.TTYSaver]        // Thread-safe TTYSaver reference

// TTY package
atomic.Bool                       // State flags (closed, etc.)
sync.Mutex                        // Buffer protection
sync.WaitGroup                    // Goroutine synchronization
```

### Verified Components

| Component | Mechanism | Test File | Status |
|-----------|-----------|-----------|--------|
| Shell command registry | `atomic.MapTyped` | `integration_test.go` | ✅ Race-free |
| Shell TTYSaver access | `atomic.Value` | `tty_integration_test.go` | ✅ Race-free |
| TTY state management | `atomic.Bool` + `sync.Mutex` | `tty/restore_advanced_test.go` | ✅ Race-free |
| Signal handler | Goroutine + channel | `tty/signal_handling_test.go` | ✅ Race-free |
| Concurrent command execution | Independent calls | `integration_test.go` | ✅ Parallel-safe |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v -timeout=10m ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -v -run "concurrent" ./...
CGO_ENABLED=1 go test -race -v -run "Concurrent" ./shell

# Stress test (10 iterations)
for i in {1..10}; do 
    CGO_ENABLED=1 go test -race ./... || break
done
```

**Result**: Zero data races across 10+ consecutive runs

---

## Subpackage Testing

### `shell` Package Tests

**Files** (11 total, 2,599 lines):
- `shell_suite_test.go` - Suite setup with test helpers
- `add_test.go` - Command registration tests
- `walk_run_test.go` - Command execution and walking
- `get_desc_test.go` - Command retrieval tests
- `coverage_test.go` - Edge case coverage
- `integration_test.go` - Full workflow tests
- `performance_test.go` - gmeasure benchmarks
- `example_test.go` - GoDoc examples
- `constructor_test.go` - TTYSaver integration
- `prompt_test.go` - Executor and completer tests
- `tty_integration_test.go` - TTY integration tests

**Key Test Scenarios**:
```go
// Command registration
Describe("Add method")
  - Single command addition
  - Multiple commands at once
  - Commands with prefixes
  - Nil command handling
  - Concurrent additions

// Command execution
Describe("Run method")
  - Valid command execution
  - Invalid command errors
  - Empty args handling
  - Concurrent execution

// Interactive features
Describe("Prompt Functions")
  - Executor behavior
  - Command suggestions
  - Writer handling
```

**Coverage**: 48.1% (120 specs)

### `shell/command` Package Tests

**Files** (1 file, 196 lines):
- `command_test.go` - Complete command interface tests

**Key Test Scenarios**:
```go
Describe("Command Creation")
  - New command with all fields
  - Nil function handling
  - Name and description access

Describe("Command Execution")
  - Run with arguments
  - Nil writers handling
  - Multiple executions

Describe("Thread Safety")
  - Concurrent command creation
  - Concurrent execution
```

**Coverage**: 81.8% (93 specs)

### `shell/tty` Package Tests

**Files** (11 total, 2,678 lines):
- `tty_suite_test.go` - Suite with mock TTYSaver
- `tty_test.go` - Core TTYSaver tests
- `terminal_test.go` - IsTerminal() tests (some skipped)
- `input_test.go` - Constructor with various inputs
- `signal_handling_test.go` - Signal() and SignalHandler()
- `restore_advanced_test.go` - Advanced restore scenarios
- `errors_test.go` - Error handling and edge cases
- `benchmark_test.go` - Performance benchmarks
- Plus 4 more test files

**Key Test Scenarios**:
```go
Describe("TTYSaver")
  - Creation with nil/file/pipe input
  - IsTerminal detection (skipped if not terminal)
  - Restore operations
  - Signal handling setup
  - Concurrent restores
  - Error cases

Describe("Benchmarks")
  - Restore performance
  - Signal handler setup
  - Concurrent operations
  - Mock vs real operations
```

**Coverage**: 44.7% (116/126 specs, 10 skipped)

**Skipped Tests**: Terminal-dependent tests skip when stdin is not a terminal (CI environments, pipes, etc.)

---

## Test File Organization

### Shell Package Tests

| File | Specs | Purpose |
|------|-------|---------|
| `shell_suite_test.go` | 0 | Suite setup, test helpers (safeBuffer, etc.) |
| `add_test.go` | 15 | Command registration |
| `walk_run_test.go` | 18 | Command execution and walking |
| `get_desc_test.go` | 14 | Command retrieval |
| `coverage_test.go` | 12 | Edge cases and error handling |
| `integration_test.go` | 18 | Full workflow scenarios |
| `performance_test.go` | 11 | gmeasure benchmarks |
| `example_test.go` | 8 | GoDoc examples |
| `constructor_test.go` | 8 | TTYSaver constructor tests |
| `prompt_test.go` | 11 | Executor and completer |
| `tty_integration_test.go` | 5 | TTY integration |

### Command Package Tests

| File | Specs | Purpose |
|------|-------|---------|
| `command_test.go` | 93 | Complete command interface |

### TTY Package Tests

| File | Specs | Purpose |
|------|-------|---------|
| `tty_suite_test.go` | 0 | Suite setup with mock |
| `tty_test.go` | 12 | Core functionality |
| `terminal_test.go` | 15 | Terminal detection (10 skipped) |
| `input_test.go` | 25 | Constructor inputs |
| `signal_handling_test.go` | 18 | Signal handling |
| `restore_advanced_test.go` | 22 | Advanced restore |
| `errors_test.go` | 24 | Error scenarios |
| `benchmark_test.go` | 0 | Performance benchmarks |

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should register command with prefix", func() {
    // Test implementation
})

It("should handle concurrent command execution without race conditions", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should execute registered command", func() {
    // Arrange
    sh := shell.New(nil)
    outBuf := newSafeBuffer()
    errBuf := newSafeBuffer()
    sh.Add("", command.New("test", "Test", func(out, err io.Writer, args []string) {
        fmt.Fprint(out, "ok")
    }))
    
    // Act
    sh.Run(outBuf, errBuf, []string{"test"})
    
    // Assert
    Expect(outBuf.String()).To(Equal("ok"))
    Expect(errBuf.String()).To(BeEmpty())
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(ContainElement(item))
Expect(count).To(BeNumerically(">", 0))
Expect(sh).ToNot(BeNil())
```

**4. Always Cleanup Resources**
```go
var ttySaver tty.TTYSaver

BeforeEach(func() {
    ttySaver, _ = tty.New(nil, false)
})

AfterEach(func() {
    if ttySaver != nil {
        ttySaver.Restore()
    }
})
```

**5. Test Edge Cases** - nil inputs, empty args, concurrent access, etc.

**6. Skip Terminal-Dependent Tests** - Use conditional skips for terminal tests
```go
It("should detect terminal", func() {
    if !tty.IsTerminalFd(0) {
        Skip("Not running in terminal")
    }
    // Test terminal-specific behavior
})
```

### Test Template

```go
var _ = Describe("shell/new_feature", func() {
    var (
        sh      shell.Shell
        testCmd command.Command
    )

    BeforeEach(func() {
        sh = shell.New(nil)
        testCmd = command.New("test", "Test command", func(out, err io.Writer, args []string) {
            fmt.Fprint(out, "test output")
        })
    })

    Context("when using new feature", func() {
        It("should perform expected behavior", func() {
            // Arrange
            sh.Add("", testCmd)
            buf := newSafeBuffer()
            
            // Act
            sh.Run(buf, nil, []string{"test"})
            
            // Assert
            Expect(buf.String()).To(Equal("test output"))
        })

        It("should handle error case", func() {
            // Test error scenario
            Expect(err).To(HaveOccurred())
        })
    })
})
```

---

## Best Practices

**Test Independence**
- ✅ Each test should be independent
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Avoid global mutable state
- ✅ Create shells/commands on-demand
- ❌ Don't rely on test execution order

**Test Helpers**
```go
// Use provided helpers
buf := newSafeBuffer()          // Thread-safe buffer
counter := newCallCounter()     // Thread-safe counter
writer := newTestWriter()       // Configurable writer
```

**Concurrent Testing**
```go
It("should handle concurrent operations", func() {
    sh := shell.New(nil)
    done := make(chan bool, 10)
    
    for i := 0; i < 10; i++ {
        go func(id int) {
            defer GinkgoRecover()
            sh.Add("", command.New(fmt.Sprintf("cmd%d", id), "Test", nil))
            done <- true
        }(i)
    }
    
    for i := 0; i < 10; i++ {
        <-done
    }
})
```
- Always run with `-race` during development
- Test concurrent operations explicitly
- Use `GinkgoRecover()` in goroutines
- Verify cleanup with channels

**Performance**
- Keep tests fast (<1s total)
- Use small data sets
- Target: <10ms per spec
- Use gmeasure for benchmarks

**Error Handling**
```go
// ✅ Good
ttySaver, err := tty.New(nil, true)
Expect(err).ToNot(HaveOccurred())
Expect(ttySaver).ToNot(BeNil())

// ❌ Bad
ttySaver, _ := tty.New(nil, true) // Don't ignore errors!
```

**Conditional Skips**
```go
// Skip terminal-dependent tests
BeforeEach(func() {
    if !isTerminal() {
        Skip("Terminal required for this test")
    }
})
```

---

## Troubleshooting

**Terminal-Dependent Test Failures**
```bash
# Some tty tests require actual terminal
# Expected: 10 skipped in CI/non-terminal environments
# Run locally in terminal to execute all tests
```

**Stale Test Cache**
```bash
go clean -testcache
go test ./...
```

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected shared variable access
- Missing atomic operations
- Unsynchronized goroutines

Example fix:
```go
// ❌ Bad: Direct map access
cmd := s.commands[name]  // Race condition

// ✅ Good: Atomic operation
cmd, ok := s.c.Load(name)  // Thread-safe
```

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian: sudo apt-get install build-essential
# macOS: xcode-select --install
# Windows: Install MinGW-w64

export CGO_ENABLED=1
go test -race ./...
```

**Test Timeouts**
```bash
# Increase timeout for slow tests
go test -timeout=10m ./...

# Or use Ginkgo
ginkgo --timeout=10m
```

Check for:
- Goroutine leaks (missing done signals)
- Blocking signal handlers
- Unclosed channels

**Debugging Specific Tests**
```bash
# Single test
ginkgo --focus="should execute registered command"

# Specific file
go test -v -run TestTTY ./tty

# Verbose output with Ginkgo
ginkgo -v --trace
```

Use `GinkgoWriter` for debug output:
```go
fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
```

**Import Cycles**
- Tests use `package shell_test`, `package command_test`, `package tty_test` to avoid cycles
- Test files in the same directory can import the package being tested

---

## CI Integration

**GitHub Actions Example**
```yaml
name: Shell Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test -v -timeout=10m ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -timeout=10m ./...
      
      - name: Coverage
        run: go test -coverprofile=coverage.out -covermode=atomic ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

**Pre-commit Hook**
```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./... || exit 1

echo "Running race detector..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
go test -cover ./... | grep -E "coverage:" || exit 1

echo "All tests passed!"
```

**Makefile Integration**
```makefile
.PHONY: test test-race test-cover test-all

test:
	go test ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-all: test test-race test-cover
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained or improved (≥60%)
- [ ] New features have tests
- [ ] Edge cases tested
- [ ] Error cases tested
- [ ] Thread safety validated
- [ ] Test duration reasonable (<2s total)
- [ ] Documentation updated
- [ ] Examples added for new features

---

## Performance Metrics

**Test Execution Times** (reference)

| Suite | Without Race | With Race | Ratio |
|-------|--------------|-----------|-------|
| `shell` | 0.089s | 1.669s | 18.7x |
| `shell/command` | 0.023s | 0.045s | 2.0x |
| `shell/tty` | 0.322s | 1.483s | 4.6x |
| **Total** | **0.434s** | **3.197s** | **7.4x** |

**Note**: Race detector overhead is expected (2-10x slower)

**Benchmark Results** (from gmeasure):

Shell Package:
- Add (single): <1µs per operation
- Get: <1µs per operation  
- Walk (1000 commands): 100µs
- Run: <1µs + command execution time
- Concurrent operations: 0-200µs

TTY Package:
- Restore: 0-100µs
- Signal handler setup: <1µs
- Concurrent restore: 100µs median

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)
- [Go Coverage](https://go.dev/blog/cover)

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync Package](https://pkg.go.dev/sync)
- [atomic Package](https://pkg.go.dev/sync/atomic)

**Shell Testing**
- [go-prompt](https://github.com/c-bata/go-prompt)
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Shell Package Contributors
