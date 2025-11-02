# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-57%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Comprehensive testing documentation for the bufferReadCloser package, covering test execution, coverage analysis, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Test Structure](#test-structure)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The bufferReadCloser package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite Statistics**
- Total Specs: 57
- Coverage: 100.0%
- Execution Time: ~10ms
- Success Rate: 100%

**Coverage Areas**
- Buffer wrapper (bytes.Buffer + io.Closer)
- Reader wrapper (bufio.Reader + io.Closer)
- Writer wrapper (bufio.Writer + io.Closer)
- ReadWriter wrapper (bufio.ReadWriter + io.Closer)
- Custom close functions and error handling
- All I/O operations and edge cases

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
- Expressive test descriptions
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Rich CLI with filtering and reporting

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax (`Expect(...).To(...)`)
- Extensive built-in matchers
- Detailed failure messages
- Type-safe assertions

---

## Running Tests

### Basic Commands

```bash
# Standard go test
go test .
go test -v .                                    # Verbose output
go test -cover .                                # With coverage

# Ginkgo CLI (recommended)
ginkgo                                          # Run all tests
ginkgo -v                                       # Verbose output
ginkgo -cover                                   # With coverage
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
# coverage: 100.0% of statements
```

### Advanced Options

```bash
# Focus on specific tests
ginkgo --focus="Buffer"                         # Only Buffer tests
ginkgo --focus="Close"                          # Only Close-related tests

# Exclude specific tests
ginkgo --skip="Edge cases"

# Parallel execution (not recommended for this package)
ginkgo -p                                       # May cause timing issues

# Generate JUnit XML report
ginkgo --junit-report=test-results.xml

# Watch mode for TDD
ginkgo watch
```

### Performance Testing

```bash
# Run with benchmarks (if available)
go test -bench=. -benchmem .

# Check test execution time
go test -v . 2>&1 | grep "Ran"
# Expected: Ran 57 of 57 Specs in 0.009-0.015 seconds
```

---

## Test Coverage

### Coverage Metrics

**Overall Coverage: 100.0%**

All statements in the package are covered by tests with no uncovered lines.

### Coverage by Component

| Component | File | Specs | Coverage | Functions Covered |
|-----------|------|-------|----------|-------------------|
| Buffer | `buffer.go` | 16 | 100.0% | 9/9 (Read, Write, ReadFrom, WriteTo, ReadByte, ReadRune, WriteByte, WriteString, Close) |
| Reader | `reader.go` | 13 | 100.0% | 3/3 (Read, WriteTo, Close) |
| Writer | `writer.go` | 14 | 100.0% | 4/4 (Write, WriteString, ReadFrom, Close) |
| ReadWriter | `readwriter.go` | 14 | 100.0% | 6/6 (Read, Write, WriteTo, ReadFrom, WriteString, Close) |
| Interface | `interface.go` | 14* | 100.0% | 5/5 (NewBuffer, New, NewReader, NewWriter, NewReadWriter) |
| **Total** | **5 files** | **57** | **100.0%** | **27/27** |

\* *Constructors are tested implicitly through component tests*

### Coverage by Source File

```
Function Coverage Report:
buffer.go:38:       Read         100.0%
buffer.go:42:       ReadFrom     100.0%
buffer.go:46:       ReadByte     100.0%
buffer.go:50:       ReadRune     100.0%
buffer.go:54:       Write        100.0%
buffer.go:58:       WriteString  100.0%
buffer.go:62:       WriteTo      100.0%
buffer.go:66:       WriteByte    100.0%
buffer.go:70:       Close        100.0%
reader.go:38:       Read         100.0%
reader.go:42:       WriteTo      100.0%
reader.go:46:       Close        100.0%
writer.go:38:       ReadFrom     100.0%
writer.go:42:       Write        100.0%
writer.go:46:       WriteString  100.0%
writer.go:50:       Close        100.0%
readwriter.go:38:   Read         100.0%
readwriter.go:42:   WriteTo      100.0%
readwriter.go:46:   ReadFrom     100.0%
readwriter.go:50:   Write        100.0%
readwriter.go:54:   WriteString  100.0%
readwriter.go:58:   Close        100.0%
interface.go:49:    New          100.0%
interface.go:56:    NewBuffer    100.0%
interface.go:72:    NewReader    100.0%
interface.go:89:    NewWriter    100.0%
interface.go:112:   NewReadWriter 100.0%
total:              (statements)  100.0%
```

### Viewing Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out .

# View function-level coverage
go tool cover -func=coverage.out

# Generate interactive HTML report
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser to see line-by-line coverage
```

---

## Test Structure

### Test Files

| File | Purpose | Specs | Description |
|------|---------|-------|-------------|
| `bufferreadcloser_suite_test.go` | Test suite entry point | - | Ginkgo test registration |
| `buffer_test.go` | Buffer wrapper tests | 16 | All Buffer operations and close behavior |
| `reader_test.go` | Reader wrapper tests | 13 | Reader operations and close behavior |
| `writer_test.go` | Writer wrapper tests | 14 | Writer operations, flush, and close |
| `readwriter_test.go` | ReadWriter wrapper tests | 14 | Bidirectional I/O and close behavior |

### Test Hierarchy

Each test file follows a consistent BDD structure:

```
Describe("ComponentName", func() {
    Context("Creation", func() {
        It("should create from underlying type", ...)
        It("should create with custom close function", ...)
        It("should create using deprecated constructor", ...) // Buffer only
    })
    
    Context("Read operations", func() {
        It("should read data", ...)
        It("should read byte/rune", ...)                      // Buffer only
        It("should read from reader", ...)
        It("should write to writer", ...)
    })
    
    Context("Write operations", func() {                      // Buffer, Writer, ReadWriter
        It("should write data", ...)
        It("should write string", ...)
        It("should write byte", ...)                          // Buffer only
        It("should read from source", ...)                    // Writer, ReadWriter
    })
    
    Context("Combined operations", func() {                   // Buffer, ReadWriter
        It("should support read and write", ...)
    })
    
    Context("Close operations", func() {
        It("should close and reset", ...)
        It("should call custom close function", ...)
        It("should return close function error", ...)
        It("should be safe to close multiple times", ...)
    })
    
    Context("Edge cases", func() {
        It("should handle empty buffer", ...)
        It("should handle large data", ...)
    })
})
```

### Test Naming Convention

Tests use descriptive "should" statements:

```go
It("should create buffer from bytes.Buffer", ...)
It("should read data", ...)
It("should flush on close", ...)
It("should return close function error", ...)
```

This makes test output self-documenting and easy to understand.

---

## Writing Tests

### Test Development Guidelines

**1. Follow the AAA Pattern** (Arrange, Act, Assert)

```go
It("should read data correctly", func() {
    // Arrange - Set up test conditions
    source := strings.NewReader("test data")
    br := bufio.NewReader(source)
    reader := NewReader(br, nil)
    
    // Act - Execute the operation
    data := make([]byte, 4)
    n, err := reader.Read(data)
    
    // Assert - Verify the results
    Expect(err).ToNot(HaveOccurred())
    Expect(n).To(Equal(4))
    Expect(string(data)).To(Equal("test"))
})
```

**2. Use Descriptive Test Names**

Good names explain what and why:

```go
// ✅ Good - Clear intent
It("should flush buffered data on close", ...)
It("should return error from custom close function", ...)

// ❌ Bad - Vague
It("should work", ...)
It("tests close", ...)
```

**3. Test Both Success and Error Paths**

```go
Context("Close operations", func() {
    It("should close successfully with nil close function", func() {
        buf := NewBuffer(bytes.NewBuffer(nil), nil)
        Expect(buf.Close()).ToNot(HaveOccurred())
    })
    
    It("should return error from custom close function", func() {
        expectedErr := errors.New("close failed")
        buf := NewBuffer(bytes.NewBuffer(nil), func() error {
            return expectedErr
        })
        Expect(buf.Close()).To(Equal(expectedErr))
    })
})
```

**4. Test Edge Cases**

```go
Context("Edge cases", func() {
    It("should handle empty buffer", func() {
        buf := NewBuffer(bytes.NewBuffer(nil), nil)
        data := make([]byte, 10)
        _, err := buf.Read(data)
        Expect(err).To(Equal(io.EOF))
    })
    
    It("should handle large data", func() {
        largeData := make([]byte, 1024*1024) // 1 MB
        buf := NewBuffer(bytes.NewBuffer(nil), nil)
        n, err := buf.Write(largeData)
        Expect(err).ToNot(HaveOccurred())
        Expect(n).To(Equal(len(largeData)))
    })
})
```

**5. Verify Custom Close Functions Are Called**

```go
It("should call custom close function", func() {
    closeCalled := false
    buf := NewBuffer(bytes.NewBuffer(nil), func() error {
        closeCalled = true
        return nil
    })
    
    buf.Close()
    Expect(closeCalled).To(BeTrue())
})
```

### Test Template

```go
var _ = Describe("NewFeature", func() {
    Context("Basic operations", func() {
        It("should perform expected operation", func() {
            // Arrange
            underlying := bytes.NewBuffer(nil)
            wrapper := NewFeature(underlying, nil)
            
            // Act
            result, err := wrapper.SomeOperation()
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(expectedValue))
        })
    })
    
    Context("Error handling", func() {
        It("should handle error condition", func() {
            // Test error scenario
        })
    })
    
    Context("Close behavior", func() {
        It("should clean up resources on close", func() {
            // Test close behavior
        })
    })
})
```

---

## Best Practices

### 1. Use In-Memory Buffers

Avoid external dependencies for fast, reliable tests:

```go
// ✅ Good - Self-contained
It("should handle data", func() {
    buf := bytes.NewBufferString("test data")
    wrapper := NewBuffer(buf, nil)
    // Test operations
})

// ❌ Bad - External dependency
It("should handle file", func() {
    file, _ := os.Open("/tmp/testfile") // Flaky, slow
    // ...
})
```

### 2. Test Close Behavior Explicitly

Verify that close performs expected cleanup:

```go
It("should reset buffer on close", func() {
    b := bytes.NewBufferString("data to clear")
    buf := NewBuffer(b, nil)
    
    Expect(b.Len()).To(BeNumerically(">", 0)) // Has data
    buf.Close()
    Expect(b.Len()).To(Equal(0))              // Cleared
})
```

### 3. Verify Flush Behavior

Writers buffer data - test that it's flushed on close:

```go
It("should flush on close", func() {
    dest := &bytes.Buffer{}
    bw := bufio.NewWriter(dest)
    writer := NewWriter(bw, nil)
    
    writer.WriteString("test")
    Expect(dest.Len()).To(Equal(0))        // Buffered, not visible
    
    writer.Close()
    Expect(dest.String()).To(Equal("test")) // Flushed and visible
})
```

### 4. Test Custom Close Functions

Ensure callbacks are invoked and errors are propagated:

```go
It("should call custom close function", func() {
    callCount := 0
    buf := NewBuffer(bytes.NewBuffer(nil), func() error {
        callCount++
        return nil
    })
    
    buf.Close()
    Expect(callCount).To(Equal(1))
})

It("should propagate close errors", func() {
    expectedErr := errors.New("cleanup failed")
    buf := NewBuffer(bytes.NewBuffer(nil), func() error {
        return expectedErr
    })
    
    err := buf.Close()
    Expect(err).To(Equal(expectedErr))
})
```

### 5. Test Idempotent Close

Verify that multiple Close() calls are safe:

```go
It("should be safe to close multiple times", func() {
    buf := NewBuffer(bytes.NewBuffer(nil), nil)
    
    err1 := buf.Close()
    err2 := buf.Close()
    err3 := buf.Close()
    
    Expect(err1).ToNot(HaveOccurred())
    Expect(err2).ToNot(HaveOccurred())
    Expect(err3).ToNot(HaveOccurred())
})
```

### 6. Keep Tests Fast

Each spec should complete in < 1ms:

```go
// ✅ Good - Fast in-memory operation
It("should write data", func() {
    buf := NewBuffer(bytes.NewBuffer(nil), nil)
    buf.WriteString("test")
})

// ❌ Bad - Slow operation
It("should wait", func() {
    time.Sleep(time.Second) // Don't do this
})
```

---

## Troubleshooting

### Common Issues

**Problem: Data not visible after write**

```go
// ❌ Wrong - Data is buffered
writer.WriteString("test")
Expect(dest.String()).To(Equal("test")) // Fails!

// ✅ Correct - Flush first
writer.WriteString("test")
writer.Close() // Flushes
Expect(dest.String()).To(Equal("test")) // Passes
```

**Problem: EOF handling**

```go
// Handle EOF gracefully
n, err := reader.Read(data)
if err != nil && err != io.EOF {
    Expect(err).ToNot(HaveOccurred())
}
```

**Problem: Coverage not updating**

```bash
# Clear test cache
go clean -testcache
go test -cover .
```

**Problem: Cannot import test dependencies**

```bash
# Download dependencies
go mod tidy
go mod download
```

### Debugging Tests

**Run specific test:**
```bash
ginkgo --focus="should flush on close"
```

**Verbose output:**
```bash
ginkgo -v --trace
```

**Debug logging in tests:**
```go
It("should debug", func() {
    fmt.Fprintf(GinkgoWriter, "Debug: value=%v\n", someValue)
    // GinkgoWriter only shows output on failure
})
```

**Check for race conditions:**
```bash
go test -race .
# Note: This package is not thread-safe by design
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
          cd ioutils/bufferReadCloser
          go test -v -cover .
      
      - name: Check coverage
        run: |
          cd ioutils/bufferReadCloser
          go test -coverprofile=coverage.out .
          go tool cover -func=coverage.out | grep total | awk '{print $3}' | grep -E '^100'
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

cd ioutils/bufferReadCloser || exit 1
go test ./... || exit 1
go test -cover . | grep "100.0%" || {
    echo "Coverage must be 100%"
    exit 1
}
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)
- [Go Coverage](https://go.dev/blog/cover)

**Standard Library**
- [io Package](https://pkg.go.dev/io)
- [bufio Package](https://pkg.go.dev/bufio)
- [bytes Package](https://pkg.go.dev/bytes)

**Related Documentation**
- [README.md](README.md) - Package overview and usage
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/bufferReadCloser)

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Test Execution Time**: ~10ms  
**Test Success Rate**: 100%  
**Maintained By**: bufferReadCloser Package Contributors
