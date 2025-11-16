# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-123%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-89.6%25-brightgreen)]()

Comprehensive testing documentation for the mail/render package, covering test execution, benchmarks, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Benchmarks](#benchmarks)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The mail/render package uses **Ginkgo v2** (BDD testing framework), **Gomega** (matcher library), and **gmeasure** (performance measurement) for comprehensive testing with expressive assertions and accurate benchmarks.

**Test Suite Metrics**
- **Total Specs**: 123
- **Coverage**: 89.6% of statements
- **Race Detection**: ✅ Zero data races
- **Execution Time**: ~1.7s (without race), ~29.6s (with race)
- **Test Files**: 7 organized test files

**Coverage Areas**
- Mailer interface and configuration
- Theme and text direction parsing
- Email body management and cloning
- HTML and plain text generation
- Template variable replacement (ParseData)
- Concurrent operations and thread safety
- Error handling and validation

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

# Run benchmarks
go test -v ./... | grep -A 10 "Report Entries"

# Using Ginkgo CLI
ginkgo -v
ginkgo -cover -race
```

**Expected Output**:
```
Ran 123 of 123 Specs in 1.747 seconds
SUCCESS! -- 123 Passed | 0 Failed | 0 Pending | 0 Skipped
coverage: 89.6% of statements
```

---

## Test Framework

### Ginkgo v2

**BDD testing framework** ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Expressive spec descriptions
- Rich failure reporting
- Built-in parallelization support

### Gomega

**Matcher library** ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages
- Asynchronous assertions

### gmeasure

**Performance measurement** ([docs](https://onsi.github.io/gomega/#gmeasure-benchmarking-code))
- Statistical measurements (mean, median, stddev, max)
- Structured benchmark reporting
- Multiple measurement types
- Integration with Ginkgo Report Entries

---

## Running Tests

### Basic Commands

```bash
# Standard test run
go test ./...

# Verbose output
go test -v ./...

# With coverage report
go test -cover ./...

# Coverage with profile
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Specific timeout
go test -timeout=10m ./...
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Verbose output
ginkgo -v

# Focus on specific test
ginkgo --focus="should generate HTML"

# Focus on file
ginkgo --focus-file=render_test.go

# Parallel execution
ginkgo -p

# With coverage
ginkgo -cover -coverprofile=coverage.out

# Generate JUnit report
ginkgo --junit-report=results.xml
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Verbose race detection
CGO_ENABLED=1 go test -race -v ./...
```

**What Race Detection Validates**:
- Clone() method creates independent copies
- No shared references in deep copy
- Thread-safe concurrent generation
- Proper goroutine synchronization

**Expected Output**:
```bash
# ✅ Success (Zero races)
Ran 123 of 123 Specs in 29.554 seconds
SUCCESS! -- 123 Passed | 0 Failed | 0 Pending | 0 Skipped
coverage: 89.6% of statements
ok  	github.com/nabbar/golib/mail/render	30.634s

# ❌ Race detected (should not happen)
==================
WARNING: DATA RACE
Read at 0x... by goroutine ...
Previous write at 0x... by goroutine ...
==================
```

**Status**: Zero data races detected across all test scenarios

### Performance Impact

| Test Mode | Duration | Overhead | Notes |
|-----------|----------|----------|-------|
| Normal | ~1.7s | 1x | Standard execution |
| Race Detection | ~29.6s | 17x | Expected overhead |
| Coverage | ~1.8s | 1.06x | Minimal overhead |

*Race detection overhead is normal due to instrumentation*

---

## Test Coverage

### Current Coverage: 89.6%

**Coverage by File**:

| File | Coverage | Lines | Description |
|------|----------|-------|-------------|
| `interface.go` | ~95% | 167 | Mailer interface and Clone() |
| `email.go` | 100% | 105 | Getters and setters |
| `config.go` | ~90% | 85 | Configuration and validation |
| `render.go` | ~85% | 120 | HTML/text generation, ParseData |
| `themes.go` | 100% | 66 | Theme parsing and conversion |
| `direction.go` | ~95% | 88 | Text direction parsing |
| `error.go` | 100% | 63 | Error codes and messages |

**Uncovered Areas** (10.4%):
- Edge cases in error handling
- Rare validation failure paths
- Default fallback code (theme/direction)

### View Coverage Details

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal (function-level)
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Coverage Example Output

```
github.com/nabbar/golib/mail/render/config.go:49:    Validate        90.0%
github.com/nabbar/golib/mail/render/direction.go:63: ParseTextDirection 95.2%
github.com/nabbar/golib/mail/render/email.go:38:     SetTheme        100.0%
github.com/nabbar/golib/mail/render/interface.go:86: Clone           94.7%
github.com/nabbar/golib/mail/render/render.go:36:    ParseData       85.4%
github.com/nabbar/golib/mail/render/themes.go:56:    ParseTheme      100.0%
total:                                                (statements)    89.6%
```

---

## Thread Safety

Thread safety is critical for concurrent email generation using Clone().

### Concurrency Primitives

The package uses **deep copying** to ensure thread safety:

```go
// Clone() performs deep copy of:
// - Slices (Intros, Outros, Dictionary)
// - Tables with nested data
// - Actions and buttons
// - Maps (Table columns)
```

### Verified Components

| Component | Mechanism | Concurrent Ops | Status |
|-----------|-----------|----------------|--------|
| `Clone()` | Deep copy of all nested structures | ✅ Independent copies | Race-free |
| `ParseData()` | In-place modification | ❌ Not thread-safe | Use per-clone |
| `GenerateHTML()` | Stateless rendering | ✅ Via Clone() | Race-free |
| `Config.NewMailer()` | Immutable creation | ✅ No shared state | Race-free |

### Thread Safety Pattern

**✅ Correct Pattern**:
```go
baseMailer := render.New()
// Configure base template...

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        mailer := baseMailer.Clone()  // Independent copy
        // Safe to modify mailer in this goroutine
        body := &hermes.Body{Name: fmt.Sprintf("User %d", id)}
        mailer.SetBody(body)
        htmlBuf, _ := mailer.GenerateHTML()
    }(i)
}
wg.Wait()
```

**❌ Incorrect Pattern**:
```go
mailer := render.New()

for i := 0; i < 100; i++ {
    go func(id int) {
        // RACE CONDITION: Shared mailer
        body := &hermes.Body{Name: fmt.Sprintf("User %d", id)}
        mailer.SetBody(body)  // Concurrent writes!
        mailer.GenerateHTML()
    }(i)
}
```

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrency tests
CGO_ENABLED=1 go test -race -v -run "Concurrency" ./...

# Stress test (run multiple times)
for i in {1..10}; do 
    echo "Run $i"
    CGO_ENABLED=1 go test -race ./... || break
done
```

**Result**: Zero data races across all concurrent test scenarios

---

## Benchmarks

### Benchmark Results

All benchmarks use **gmeasure** for statistical accuracy with 100-1000 iterations.

#### Creation Benchmarks

| Operation | Mean | Median | StdDev | Notes |
|-----------|------|--------|--------|-------|
| `New()` | ~100 ns | ~100 ns | ~20 ns | Struct initialization |
| `Config.NewMailer()` | ~300 ns | ~300 ns | ~50 ns | With parsing |

#### Clone Benchmarks

| Operation | Mean | Median | StdDev | Notes |
|-----------|------|--------|--------|-------|
| Simple Clone | ~1 µs | ~1 µs | ~200 ns | Empty body |
| Complex Clone | ~10 µs | ~10 µs | ~2 µs | Tables + actions |

#### Generation Benchmarks (without race)

| Operation | Mean | Median | StdDev | Max | Notes |
|-----------|------|--------|--------|-----|-------|
| HTML (Simple) | 2.7 ms | 2.5 ms | 300 µs | 3.4 ms | Basic email |
| HTML (Complex) | 3.1 ms | 3.0 ms | 400 µs | 4.7 ms | Tables + actions |
| Plain Text | 3.6 ms | 3.5 ms | 400 µs | 5.6 ms | Text conversion |

#### Generation Benchmarks (with race detection)

| Operation | Mean | Median | StdDev | Max | Notes |
|-----------|------|--------|--------|-----|-------|
| HTML (Simple) | 43 ms | 41 ms | 2.5 ms | 52 ms | 16x overhead |
| HTML (Complex) | 51 ms | 50 ms | 3.7 ms | 66 ms | 16x overhead |
| Plain Text | 48 ms | 47 ms | 3.2 ms | 60 ms | 13x overhead |

*Race detection overhead is expected and acceptable for testing*

#### ParseData Benchmarks

| Operation | Mean | Median | Notes |
|-----------|------|--------|-------|
| Simple (few variables) | 350 ns | 300 ns | ~5 replacements |
| Complex (many variables) | 1.2 µs | 1.0 µs | ~20 replacements |

#### Parsing Benchmarks

| Operation | Mean | Median | Notes |
|-----------|------|--------|-------|
| `ParseTheme()` | 83 ns | 70 ns | String comparison |
| `ParseTextDirection()` | 54 ns | 50 ns | String parsing |

#### Validation Benchmarks

| Operation | Mean | Median | StdDev | Notes |
|-----------|------|--------|--------|-------|
| `Config.Validate()` | 31 µs | 30 µs | 5 µs | All validations |

#### Complete Workflow Benchmark

| Operation | Mean | Median | StdDev | Max | Notes |
|-----------|------|--------|--------|-----|-------|
| Config → HTML | 3.1 ms | 3.0 ms | 400 µs | 4.7 ms | Full pipeline |

*Workflow: Config validation → NewMailer() → ParseData → GenerateHTML*

### Running Benchmarks

```bash
# Run all benchmarks
go test -v ./... 2>&1 | grep -A 20 "Report Entries"

# Save benchmark results
go test -v ./... > benchmark-results.txt

# Compare benchmarks
go test -bench=. -benchmem ./...
```

### Benchmark Test Files

Benchmarks are integrated into test files using gmeasure:

- `benchmark_test.go` - Creation, generation, parsing, validation
- `concurrency_test.go` - Concurrent operations

### Performance Expectations

**Target Performance**:
- Email generation: <5ms (99th percentile)
- Clone operation: <20µs (complex)
- ParseData: <2µs (complex)
- Config validation: <50µs

**Throughput Estimates** (single-threaded):
- Email generation: ~300-350 emails/second
- With 10 goroutines: ~3000 emails/second
- ParseData only: ~1M operations/second

---

## Test File Organization

### Test Suite Structure

| File | Specs | Purpose | Focus |
|------|-------|---------|-------|
| `render_suite_test.go` | 1 | Suite initialization | Ginkgo entry point |
| `interface_test.go` | 28 | Mailer interface | CRUD, Clone, getters/setters |
| `config_test.go` | 20 | Configuration | Validation, NewMailer |
| `themes_test.go` | 21 | Themes & direction | Parsing, string conversion |
| `render_test.go` | 38 | Email generation | HTML, text, ParseData |
| `concurrency_test.go` | 5 | Thread safety | Concurrent operations |
| `benchmark_test.go` | 10 | Performance | gmeasure benchmarks |

**Total**: 123 test specifications

### Test Organization Pattern

Tests follow Ginkgo's hierarchical BDD structure:

```go
var _ = Describe("Component", func() {
    Context("When doing operation", func() {
        var (
            testData   []byte
            mailer     render.Mailer
        )

        BeforeEach(func() {
            // Per-test setup
            mailer = render.New()
        })

        It("should perform expected behavior", func() {
            // Arrange
            mailer.SetName("Test Company")
            
            // Act
            result, err := mailer.GenerateHTML()
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result.Len()).To(BeNumerically(">", 0))
        })
    })
})
```

---

## Writing Tests

### Test Guidelines

**1. Descriptive Names**
```go
It("should generate HTML email with proper structure", func() {
    // Clear, specific description
})

It("should handle empty body gracefully", func() {
    // Describes the edge case
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should replace template variables in all fields", func() {
    // Arrange
    mailer := render.New()
    mailer.SetName("{{company}}")
    body := &hermes.Body{
        Name: "{{user}}",
        Intros: []string{"Code: {{code}}"},
    }
    mailer.SetBody(body)
    
    // Act
    mailer.ParseData(map[string]string{
        "{{company}}": "Acme Inc",
        "{{user}}":    "John Doe",
        "{{code}}":    "123456",
    })
    
    // Assert
    result := mailer.GetName()
    Expect(result).To(Equal("Acme Inc"))
    body = mailer.GetBody()
    Expect(body.Name).To(Equal("John Doe"))
    Expect(body.Intros[0]).To(Equal("Code: 123456"))
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))           // Exact match
Expect(err).ToNot(HaveOccurred())          // No error
Expect(list).To(ContainElement(item))      // Contains
Expect(number).To(BeNumerically(">", 0))   // Numeric comparison
Expect(str).To(ContainSubstring("text"))   // Substring
Expect(buf.Len()).To(BeNumerically(">", 1000)) // Buffer length
```

**4. Test Error Cases**
```go
It("should return error for invalid configuration", func() {
    config := render.Config{
        Name: "", // Missing required field
        Link: "not-a-url", // Invalid URL
    }
    
    err := config.Validate()
    Expect(err).To(HaveOccurred())
    Expect(err.Code()).To(Equal(render.ErrorMailerConfigInvalid))
})
```

**5. Test Edge Cases**
```go
It("should handle empty body content", func() {
    mailer := render.New()
    body := &hermes.Body{} // Empty
    mailer.SetBody(body)
    
    htmlBuf, err := mailer.GenerateHTML()
    Expect(err).ToNot(HaveOccurred())
    Expect(htmlBuf.Len()).To(BeNumerically(">", 0))
})

It("should handle special characters in content", func() {
    mailer := render.New()
    body := &hermes.Body{
        Name: "Test & <User>",
        Intros: []string{"Special chars: & < > \" '"},
    }
    mailer.SetBody(body)
    
    htmlBuf, _ := mailer.GenerateHTML()
    html := htmlBuf.String()
    Expect(html).To(ContainSubstring("&amp;"))
    Expect(html).To(ContainSubstring("&lt;"))
})
```

**6. Test Concurrent Operations**
```go
It("should safely generate emails concurrently", func() {
    baseMailer := render.New()
    baseMailer.SetName("Test Company")
    
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            mailer := baseMailer.Clone()
            body := &hermes.Body{Name: fmt.Sprintf("User %d", id)}
            mailer.SetBody(body)
            _, err := mailer.GenerateHTML()
            if err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    Expect(errors).To(BeEmpty())
})
```

### Test Template

```go
var _ = Describe("render/new_feature", func() {
    Context("When using new feature", func() {
        var (
            mailer render.Mailer
        )

        BeforeEach(func() {
            mailer = render.New()
            // Setup common state
        })

        It("should perform expected behavior", func() {
            // Arrange
            mailer.SetTheme(render.ThemeFlat)
            
            // Act
            theme := mailer.GetTheme()
            
            // Assert
            Expect(theme).To(Equal(render.ThemeFlat))
        })

        It("should handle error case", func() {
            // Test error conditions
            config := render.Config{}
            err := config.Validate()
            Expect(err).To(HaveOccurred())
        })

        It("should work with edge cases", func() {
            // Test boundary conditions
        })
    })
})
```

---

## Best Practices

### Test Independence

**✅ Good Practices**:
- Each test should be independent
- Use `BeforeEach` for setup, `AfterEach` for cleanup
- Avoid global mutable state
- Create test data on-demand
- Don't rely on test execution order

```go
BeforeEach(func() {
    mailer = render.New()
    // Fresh instance for each test
})
```

**❌ Bad Practices**:
```go
// Don't share state across tests
var sharedMailer render.Mailer

It("test 1", func() {
    sharedMailer.SetName("Test")  // Affects other tests
})

It("test 2", func() {
    name := sharedMailer.GetName()  // Depends on test 1
})
```

### Test Data

**✅ Use Realistic Data**:
```go
body := &hermes.Body{
    Name:   "John Doe",
    Intros: []string{"Welcome to our service!"},
    Dictionary: []hermes.Entry{
        {Key: "Order ID", Value: "ORD-123456"},
        {Key: "Date", Value: "2024-01-15"},
    },
}
```

**✅ Test Data Helpers**:
```go
func createTestBody() *hermes.Body {
    return &hermes.Body{
        Name:   "Test User",
        Intros: []string{"Test intro"},
    }
}

func createTestConfig() render.Config {
    return render.Config{
        Theme:       "flat",
        Direction:   "ltr",
        Name:        "Test Company",
        Link:        "https://example.com",
        Logo:        "https://example.com/logo.png",
        Copyright:   "© 2024 Test",
        TroubleText: "Help",
        Body:        hermes.Body{},
    }
}
```

### Assertions

**✅ Specific Matchers**:
```go
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))
Expect(buf.Len()).To(BeNumerically(">", 1000))
Expect(html).To(ContainSubstring("<html>"))
```

**❌ Generic Comparisons**:
```go
Expect(value == expected).To(BeTrue())  // Less informative
Expect(err == nil).To(BeTrue())         // Use HaveOccurred()
```

### Performance Testing

**✅ Use gmeasure**:
```go
It("should measure generation performance", func() {
    experiment := gmeasure.NewExperiment("generation")
    AddReportEntry(experiment.Name, experiment)
    
    for i := 0; i < 100; i++ {
        experiment.Sample(func(idx int) {
            mailer := render.New()
            body := createTestBody()
            mailer.SetBody(body)
            _, _ = mailer.GenerateHTML()
        }, gmeasure.SamplingConfig{N: 1})
    }
    
    stats := experiment.GetStats("generation")
    AddReportEntry("Mean time", stats.DurationFor(gmeasure.StatMean))
})
```

### Cleanup

**✅ Explicit Cleanup**:
```go
AfterEach(func() {
    // Clean up if needed
    mailer = nil
})
```

**✅ Defer in Tests**:
```go
It("should cleanup resources", func() {
    file, err := os.CreateTemp("", "test-*.html")
    Expect(err).ToNot(HaveOccurred())
    defer os.Remove(file.Name())
    
    // Use file...
})
```

---

## Troubleshooting

### Stale Test Cache

```bash
# Clean test cache
go clean -testcache

# Run tests fresh
go test -count=1 ./...
```

### Race Conditions

```bash
# Debug races with verbose output
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt

# Search for race warnings
grep -A 30 "WARNING: DATA RACE" race-log.txt
```

**Common Race Patterns**:
- Shared Mailer across goroutines (use Clone())
- Concurrent ParseData calls (clone first)
- Shared configuration mutation

### Test Timeouts

```bash
# Increase timeout
go test -timeout=15m ./...

# Identify slow tests
go test -v ./... | grep -E "^--- (PASS|FAIL):"
```

### Import Cycles

Use `package render_test` to avoid import cycles:

```go
package render_test

import (
    "testing"
    "github.com/nabbar/golib/mail/render"
)
```

### CGO Not Available

```bash
# Install build tools
# Ubuntu/Debian:
sudo apt-get install build-essential

# macOS:
xcode-select --install

# Verify
export CGO_ENABLED=1
go test -race ./...
```

### Debugging Failed Tests

```bash
# Run specific test
ginkgo --focus="should generate HTML"

# Verbose output
ginkgo -v --trace

# Stop on first failure
ginkgo --fail-fast
```

**Use GinkgoWriter for debug output**:
```go
It("should do something", func() {
    fmt.Fprintf(GinkgoWriter, "Debug: mailer = %+v\n", mailer)
    // Test code...
})
```

---

## CI Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.20', '1.21']
    
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v -timeout=10m ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -timeout=10m ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### Pre-commit Hook

Save as `.git/hooks/pre-commit`:

```bash
#!/bin/bash
set -e

echo "Running tests..."
go test -timeout=5m ./mail/render/...

echo "Running race detection..."
CGO_ENABLED=1 go test -race -timeout=5m ./mail/render/...

echo "Checking coverage..."
go test -cover ./mail/render/... | grep -E "coverage:" | awk '{if ($2 < 85.0) exit 1}'

echo "✅ All checks passed"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: ≥85% (currently 89.6%)
- [ ] New features have tests
- [ ] Error cases tested
- [ ] Thread safety validated (if applicable)
- [ ] Benchmarks run successfully
- [ ] Documentation updated
- [ ] Test duration reasonable (<5s without race)

---

## Performance Targets

### Test Execution Targets

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Total specs | - | 123 | ✅ |
| Execution time | <3s | 1.7s | ✅ |
| Execution time (race) | <45s | 29.6s | ✅ |
| Coverage | ≥85% | 89.6% | ✅ |
| Data races | 0 | 0 | ✅ |

### Operation Performance Targets

| Operation | Target | Current | Status |
|-----------|--------|---------|--------|
| Email generation | <5ms | 2.7-3.6ms | ✅ |
| Clone (complex) | <20µs | ~10µs | ✅ |
| ParseData | <5µs | 0.35-1.2µs | ✅ |
| Config validation | <100µs | ~31µs | ✅ |

---

## Resources

### Testing Frameworks
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [gmeasure Guide](https://onsi.github.io/gomega/#gmeasure-benchmarking-code)
- [Go Testing](https://pkg.go.dev/testing)

### Concurrency
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync Package](https://pkg.go.dev/sync)

### Benchmarking
- [Go Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Performance Profiling](https://go.dev/blog/pprof)

### Related Documentation
- [README.md](README.md) - Package overview
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/mail/render) - API documentation

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Mail Render Package Contributors
