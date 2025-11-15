# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-23%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-85.7%25-brightgreen)]()

Comprehensive testing documentation for the fileDescriptor package, covering test execution, platform-specific testing, and privilege-aware scenarios.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Platform-Specific Testing](#platform-specific-testing)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The fileDescriptor package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with platform-aware and privilege-aware scenarios.

**Test Suite Statistics**
- Total Specs: 23
- Passed: 20
- Skipped: 3 (permission/state dependent)
- Coverage: 85.7%
- Execution Time: ~2ms
- Success Rate: 100% (passed + skipped)

**Coverage Areas**
- Current limit queries (100% coverage)
- Limit increases (with privilege handling)
- Platform-specific implementations
- Error conditions and edge cases
- System state consistency
- Realistic application scenarios

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Using Ginkgo CLI
ginkgo -v -cover
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Conditional test execution (Skip when needed)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Rich CLI with filtering and reporting

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax (`Expect(...).To(...)`)
- Numerical comparators (`BeNumerically`)
- Detailed failure messages
- Type-safe assertions

---

## Running Tests

### Basic Commands

```bash
# Standard go test
go test .
go test -v .                    # Verbose output
go test -cover .                # With coverage

# Ginkgo CLI (recommended)
ginkgo                          # Run all tests
ginkgo -v                       # Verbose output
ginkgo -cover                   # With coverage
```

### Coverage Reports

```bash
# Generate coverage profile
go test -coverprofile=coverage.out .

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Expected output:
# coverage: 85.7% of statements
```

### Advanced Options

```bash
# Focus on specific tests
ginkgo --focus="Reading current limits"
ginkgo --focus="increase"

# Skip specific tests
ginkgo --skip="permission"

# Run without privilege-dependent tests
ginkgo --skip="increase.*hard"

# Verbose with trace
ginkgo -v --trace
```

### Privilege-Aware Testing

Some tests may require elevated privileges:

```bash
# Run as regular user (some tests will skip)
go test .

# Run with elevated privileges (Unix)
sudo go test .

# Output will show skipped tests:
# S = Skipped (expected for certain system states)
```

---

## Test Coverage

### Coverage Metrics

**Overall Coverage: 85.7%**

All critical paths are tested with proper handling of platform and privilege constraints.

### Coverage by Component

| Component | File | Specs | Coverage | Functions Covered |
|-----------|------|-------|----------|-------------------|
| Public API | `fileDescriptor.go` | 23 | 100.0% | SystemFileDescriptor |
| Unix impl | `fileDescriptor_ok.go` | 15 | 89.5% | systemFileDescriptor, getCurMax |
| Windows impl | `fileDescriptor_ko.go` | 8 | N/A | Platform-specific |
| **Total** | **3 files** | **23** | **85.7%** | **3/3 public** |

### Coverage by Source File

```
Function Coverage Report:
fileDescriptor.go:49:       SystemFileDescriptor     100.0%
fileDescriptor_ok.go:38:    systemFileDescriptor      89.5%
fileDescriptor_ok.go:75:    getCurMax                 75.0%
total:                      (statements)              85.7%
```

### Test Results Breakdown

| Status | Count | Percentage | Reason |
|--------|-------|------------|--------|
| Passed | 20 | 87% | Core functionality |
| Skipped | 3 | 13% | Permission/state dependent |
| Failed | 0 | 0% | All tests pass |

**Why Tests Are Skipped:**
1. System already at maximum limit (cannot test increase)
2. No room between soft and hard limit
3. Requires specific privilege levels not available

This is **normal and expected** - tests adapt to system state.

### Viewing Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out .

# View function-level coverage
go tool cover -func=coverage.out

# Generate interactive HTML report
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser
```

---

## Test Structure

### Test Files

| File | Purpose | Specs | Description |
|------|---------|-------|-------------|
| `filedescriptor_suite_test.go` | Test suite entry point | - | Ginkgo test registration |
| `filedescriptor_test.go` | Basic functionality | 15 | Query limits, basic increases, edge cases |
| `filedescriptor_increase_test.go` | Advanced scenarios | 8 | Server configs, conditional increases, platform awareness |

### Test Hierarchy

The test suite follows a consistent BDD structure:

```
Describe("SystemFileDescriptor", func() {
    Context("Reading current limits", func() {
        It("should return current and max limits", ...)
        It("should not require privileges for reading", ...)
    })
    
    Context("Query mode", func() {
        It("should return limits without modification when newValue is zero", ...)
        It("should not modify when newValue is below current", ...)
    })
    
    Context("Increasing limits", func() {
        It("should increase within soft limit", ...)
        It("should increase to hard limit", ...)      // May skip
        It("should handle permission errors", ...)
    })
    
    Context("Edge cases", func() {
        It("should handle negative values", ...)
        It("should handle very large values", ...)
    })
    
    Context("Application scenarios", func() {
        It("should support server initialization", ...)
        It("should support conditional increases", ...)
    })
})
```

### Test Naming Convention

Tests use descriptive "should" statements:

```go
It("should return current and max file descriptor limits", ...)
It("should increase limit when newValue is higher than current", ...)
It("should skip when already at maximum", ...)
```

---

## Platform-Specific Testing

### Unix/Linux/macOS

**Implementation Tested**: `fileDescriptor_ok.go` (build tag: `!windows`)

**Test Coverage:**
- ✅ Query operations via `syscall.Getrlimit` (100%)
- ✅ Limit increases via `syscall.Setrlimit` (89.5%)
- ✅ Soft limit increases (no privileges needed)
- ⚠️ Hard limit increases (requires root - may skip)
- ✅ uint64 to int conversion with overflow handling

**Typical System Limits:**
- Soft: 1024-4096 (varies by distro)
- Hard: 4096-unlimited (often 65536)

**Privilege Requirements:**
```bash
# Regular user: Can increase soft limit up to hard limit
go test .                    # May skip hard limit tests

# Root: Can increase both soft and hard limits
sudo go test .              # All tests should pass
```

### Windows

**Implementation Tested**: `fileDescriptor_ko.go` (build tag: `windows`)

**Test Coverage:**
- ✅ Query via `maxstdio.GetMaxStdio`
- ✅ Limit modification via `maxstdio.SetMaxStdio`
- ✅ Default value (512) handling
- ✅ Hard limit (8192) enforcement
- ✅ Auto-capping at maximum

**Windows-Specific Behavior:**
- No privileges required (within 8192 limit)
- Automatic capping at 8192 (cannot exceed)
- C runtime limits, not OS-level

**Expected Test Output:**
```
Windows limits:
  Default:   512
  Maximum:  8192
  Requested: 16384 → Capped to: 8192
```

### Cross-Platform Test Design

Tests adapt to platform capabilities:

```go
BeforeEach(func() {
    originalCurrent, originalMax, _ = SystemFileDescriptor(0)
    
    // Platform-aware expectations
    if runtime.GOOS == "windows" {
        Expect(originalMax).To(Equal(8192))
    } else {
        Expect(originalMax).To(BeNumerically(">=", originalCurrent))
    }
})

It("should respect platform limits", func() {
    excessive := 100000
    current, max, _ := SystemFileDescriptor(excessive)
    
    if runtime.GOOS == "windows" {
        Expect(current).To(BeNumerically("<=", 8192))
    } else {
        Expect(current).To(BeNumerically("<=", max))
    }
})
```

---

## Writing Tests

### Test Development Guidelines

**1. Always Capture Initial State**

```go
var originalCurrent, originalMax int

BeforeEach(func() {
    var err error
    originalCurrent, originalMax, err = SystemFileDescriptor(0)
    Expect(err).ToNot(HaveOccurred())
    GinkgoWriter.Printf("System: current=%d, max=%d\n", originalCurrent, originalMax)
})
```

**2. Skip When Conditions Aren't Met**

```go
It("should increase to hard limit", func() {
    if originalMax <= originalCurrent {
        Skip("Already at maximum - cannot test increase")
    }
    
    // Test increase
})
```

**3. Handle Both Success and Permission Errors**

```go
It("should attempt to increase limit", func() {
    desired := originalCurrent + 1000
    current, max, err := SystemFileDescriptor(desired)
    
    if err == nil {
        // Success - verify increase
        Expect(current).To(BeNumerically(">=", originalCurrent))
        Expect(max).To(BeNumerically(">=", current))
    } else {
        // Permission error is acceptable
        GinkgoWriter.Printf("Expected permission error: %v\n", err)
    }
})
```

**4. Verify System State Consistency**

```go
It("should maintain valid state", func() {
    // Any operation
    SystemFileDescriptor(someValue)
    
    // State should remain valid
    current, max, err := SystemFileDescriptor(0)
    Expect(err).ToNot(HaveOccurred())
    Expect(max).To(BeNumerically(">=", current))
    Expect(current).To(BeNumerically(">", 0))
})
```

**5. Platform-Aware Testing**

```go
It("should handle platform-specific limits", func() {
    huge := 1000000
    current, max, _ := SystemFileDescriptor(huge)
    
    if runtime.GOOS == "windows" {
        Expect(current).To(BeNumerically("<=", 8192))
    } else {
        Expect(current).To(BeNumerically("<=", max))
    }
})
```

### Test Template

```go
var _ = Describe("NewFeature", func() {
    var initialCurrent, initialMax int

    BeforeEach(func() {
        initialCurrent, initialMax, _ = SystemFileDescriptor(0)
    })

    Context("Feature behavior", func() {
        It("should work correctly", func() {
            // Arrange
            testValue := initialCurrent + 100
            
            if testValue > initialMax {
                Skip("Test value exceeds system maximum")
            }
            
            // Act
            current, max, err := SystemFileDescriptor(testValue)
            
            // Assert
            if err == nil {
                Expect(current).To(BeNumerically(">=", initialCurrent))
            }
            // Permission error is acceptable
        })
    })
})
```

---

## Best Practices

### 1. Always Capture Initial State

Understand the starting point before testing:

```go
BeforeEach(func() {
    originalCurrent, originalMax, _ = SystemFileDescriptor(0)
    GinkgoWriter.Printf("Starting with: current=%d, max=%d\n", originalCurrent, originalMax)
})
```

### 2. Skip Gracefully

Don't fail when conditions aren't met:

```go
It("should test increase", func() {
    if originalMax <= originalCurrent {
        Skip("System already at maximum")
    }
    // Test code
})
```

### 3. Accept Permission Errors

Not all tests can run without privileges:

```go
_, _, err := SystemFileDescriptor(highValue)
if err != nil {
    GinkgoWriter.Printf("Permission denied (expected): %v\n", err)
    // This is OK - test passes
}
```

### 4. Use Realistic Test Values

Test with real-world application values:

```go
const (
    TypicalWebServer   = 4096
    HighTrafficServer  = 16384
    WindowsMaximum     = 8192
)
```

### 5. Keep Tests Fast

File descriptor tests should complete in microseconds:

```go
// ✅ Good - Fast query
It("should query quickly", func() {
    start := time.Now()
    SystemFileDescriptor(0)
    Expect(time.Since(start)).To(BeNumerically("<", time.Millisecond))
})
```

### 6. Don't Leave System Modified

Tests should be non-destructive:

```go
// State changes are process-local and temporary
// No cleanup needed - limits persist only for test process
```

---

## Troubleshooting

### Common Issues

**Problem: Tests show "S" (skipped)**

This is **normal and expected**. Tests skip when:
- Already at maximum limit (cannot test increase)
- No room between soft and hard limit
- Specific privileges not available

**Solution**: This is correct behavior. Skipped tests don't indicate failure.

**Problem: "operation not permitted" errors**

```bash
# Run without privilege tests
ginkgo --skip="hard limit"

# Or run with privileges (Unix)
sudo -E go test .
```

**Problem: Different results on different systems**

This is expected - limits vary by:
- Operating system (Unix vs Windows)
- Distribution (Ubuntu vs RHEL)
- System configuration
- Current user privileges

**Problem: Coverage not 100%**

Some code paths require:
- Root/admin privileges
- Specific system states
- Edge conditions hard to trigger

85.7% coverage is excellent for system-level code.

### Debugging Commands

```bash
# Run specific test
ginkgo --focus="should return current"

# Verbose output
ginkgo -v --trace

# Show all output (including GinkgoWriter)
ginkgo -v

# Check for race conditions
go test -race .
```

---

## CI Integration

### GitHub Actions Example

```yaml
name: Test

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
        run: |
          cd ioutils/fileDescriptor
          go test -v -cover .
      
      - name: Check coverage
        run: |
          cd ioutils/fileDescriptor
          go test -coverprofile=coverage.out .
          go tool cover -func=coverage.out | grep total | awk '{print $3}' | grep -E '^(8[5-9]|9[0-9])'
```

**Note**: Some tests may skip in CI environments due to limited privileges.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)

**System Documentation**
- [Unix getrlimit](https://man7.org/linux/man-pages/man2/getrlimit.2.html)
- [Windows SetMaxStdio](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio)
- [Go syscall Package](https://pkg.go.dev/syscall)

**Related Documentation**
- [README.md](README.md) - Package overview and usage
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/fileDescriptor)

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Test Execution Time**: ~2ms  
**Test Success Rate**: 100% (20 passed + 3 appropriately skipped)  
**Maintained By**: fileDescriptor Package Contributors
