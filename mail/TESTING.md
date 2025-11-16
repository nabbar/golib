# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-967%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-85.5%25-brightgreen)]()

Comprehensive testing documentation for the mail package and all subpackages, covering test execution, race detection, performance benchmarks, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Statistics](#test-statistics)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Performance Benchmarks](#performance-benchmarks)
- [Subpackage Testing](#subpackage-testing)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The mail package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions across all subpackages.

**Test Suite Summary**
- **Total Specs**: 967 (966 passed, 1 skipped)
- **Average Coverage**: 85.5%
- **Race Detection**: ✅ Zero data races
- **Execution Time**: ~38.5s (without race), ~45s (with race)

**Quality Assurance**
- ✅ Thread-safe concurrent operations verified
- ✅ Zero memory leaks detected
- ✅ Goroutine synchronization validated
- ✅ Production-ready stability confirmed

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection (critical for concurrent operations)
CGO_ENABLED=1 go test -race ./...

# Run specific subpackage
go test ./smtp/...

# Using Ginkgo CLI (faster, better output)
ginkgo -cover -race
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Parallel execution support (`ginkgo -p`)
- Rich CLI with filtering and focus

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax (`Expect(...).To(...)`)
- Extensive built-in matchers
- Detailed failure messages
- Async assertions support

**gmeasure** - Performance measurement library
- Statistical benchmarks (min, max, mean, median, stddev)
- Used in smtp/tlsmode for parsing performance

---

## Running Tests

### Basic Commands

```bash
# Standard test run (all subpackages)
go test ./...

# Verbose output with test names
go test -v ./...

# With coverage report
go test -cover ./...

# With atomic coverage mode
go test -cover -covermode=atomic ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run with timeout (important for SMTP tests)
go test -timeout=10m ./...
```

### Ginkgo CLI Options

```bash
# Run all tests (faster than go test)
ginkgo

# Specific subpackage
ginkgo ./smtp

# Pattern matching (focus on specific tests)
ginkgo --focus="SMTP.*authentication"

# Parallel execution (careful with SMTP tests)
ginkgo -p -procs=4

# Skip specific tests
ginkgo --skip="slow tests"

# Generate JUnit report (for CI)
ginkgo --junit-report=results.xml

# Verbose with stack traces
ginkgo -v --trace
```

### Race Detection

**Critical for all mail package components due to concurrent operations**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# With timeout for SMTP tests
CGO_ENABLED=1 go test -race -timeout=10m ./...

# Specific subpackage
CGO_ENABLED=1 go test -race ./queuer
```

**Validates**:
- SMTP connection state management
- Queuer atomic counters and mutex locks
- Sender concurrent email composition
- Render template rendering in goroutines

**Expected Output**:
```bash
# ✅ Success - No races detected
ok  	github.com/nabbar/golib/mail/smtp	26.796s
ok  	github.com/nabbar/golib/mail/queuer	8.568s

# ❌ Failure - Race detected (would show)
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races across all 967 specs ✅

### Performance & Profiling

```bash
# Benchmarks (currently in smtp/tlsmode)
go test -bench=. -benchmem ./smtp/tlsmode

# Memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out

# CPU profiling
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out

# Block profiling (goroutine blocking)
go test -blockprofile=block.out ./...
go tool pprof block.out
```

---

## Test Statistics

### Summary by Subpackage

| Subpackage | Specs | Passed | Skipped | Coverage | Duration | Status |
|------------|-------|--------|---------|----------|----------|--------|
| `smtp` | 104 | 104 | 0 | 80.6% | 26.8s | ✅ |
| `smtp/config` | 222 | 222 | 0 | 92.7% | 0.2s | ✅ |
| `smtp/tlsmode` | 165 | 165 | 0 | 98.8% | 0.04s | ✅ |
| `sender` | 252 | 252 | 0 | 81.4% | 0.9s | ✅ |
| `render` | 123 | 123 | 0 | 89.6% | 2.0s | ✅ |
| `queuer` | 101 | 100 | 1 | 90.8% | 8.6s | ✅ |
| **Total** | **967** | **966** | **1** | **85.5%** | **38.5s** | ✅ |

### Race Detection Statistics

With `CGO_ENABLED=1 go test -race -timeout=10m ./...`:

| Subpackage | Duration (race) | Overhead | Data Races | Status |
|------------|-----------------|----------|------------|--------|
| `smtp` | ~32s | 1.2x | 0 | ✅ |
| `smtp/config` | ~0.3s | 1.5x | 0 | ✅ |
| `smtp/tlsmode` | ~0.3s | 7x* | 0 | ✅ |
| `sender` | ~1.3s | 1.4x | 0 | ✅ |
| `render` | ~2.4s | 1.2x | 0 | ✅ |
| `queuer` | ~10s | 1.2x | 0 | ✅ |
| **Total** | **~45s** | **1.2x** | **0** | ✅ |

*Higher overhead due to small absolute duration

---

## Test Coverage

### Coverage Goals

- **Minimum**: 80% statement coverage
- **Target**: 85-90% statement coverage
- **Current**: 85.5% average

### Coverage by Component

#### SMTP Package (80.6%)

| File | Coverage | Test Focus |
|------|----------|------------|
| `client.go` | ~85% | Client methods, connection mgmt |
| `dial.go` | ~90% | TLS modes, authentication |
| `monitor.go` | ~95% | Health checks |
| `model.go` | ~75% | Internal state management |

**High Coverage Areas**:
- TLS mode handling (STARTTLS, Strict TLS)
- Authentication mechanisms
- Error handling and validation

**Lower Coverage Areas**:
- Edge cases in connection failure scenarios
- Rarely used configuration combinations

#### SMTP Config Subpackage (92.7%)

Comprehensive configuration testing:
- DSN parsing and validation
- URL encoding/decoding
- Default value handling
- Configuration cloning
- Error conditions

#### SMTP TLSMode Subpackage (98.8%)

Nearly complete coverage:
- All TLS mode constants
- String parsing and validation
- JSON/YAML/TOML encoding/decoding
- Roundtrip conversions
- Performance benchmarks

#### Sender Package (81.4%)

| Feature | Coverage | Test Focus |
|---------|----------|------------|
| Email composition | ~90% | Headers, body, attachments |
| Recipient management | ~85% | To, CC, BCC, deduplication |
| File attachments | ~80% | Regular & inline files |
| Configuration | ~85% | JSON/YAML parsing, validation |
| SMTP integration | ~75% | Send operations |

**Well-Tested**:
- Multi-part content (HTML + text)
- Address parsing and validation
- Custom headers
- Priority levels

**Improvement Areas**:
- Edge cases in attachment handling
- Error recovery during sending

#### Render Package (89.6%)

| Feature | Coverage | Test Focus |
|---------|----------|------------|
| Template rendering | ~95% | HTML and text generation |
| Theme support | ~90% | Default and Flat themes |
| Variable parsing | ~85% | {{variable}} substitution |
| Configuration | ~90% | Validation, defaults |
| Body composition | ~88% | Actions, tables, dictionaries |

**High Quality**:
- All themes tested
- RTL/LTR direction handling
- Complex body structures

#### Queuer Package (90.8%)

| Feature | Coverage | Test Focus |
|---------|----------|------------|
| Rate limiting | ~95% | Throttling algorithm |
| Context handling | ~92% | Cancellation, timeout |
| SMTP wrapping | ~90% | Interface compliance |
| Concurrency | ~95% | Thread safety, race detection |
| Configuration | ~85% | Callback setup |

**Thoroughly Tested**:
- Concurrent sending scenarios
- Rate limit enforcement
- Context cancellation during throttle
- Atomic counter operations

**Note**: 1 spec skipped (intentional for specific test scenario)

### Viewing Coverage

```bash
# Generate coverage for all subpackages
go test -coverprofile=coverage.out ./...

# View in terminal (summary)
go tool cover -func=coverage.out

# View in terminal (by package)
go tool cover -func=coverage.out | grep -E "^github.com/nabbar/golib/mail/"

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows

# Per-subpackage coverage
go test -coverprofile=smtp.out ./smtp
go tool cover -html=smtp.out
```

---

## Thread Safety

Thread safety is critical for the mail package due to:
- Concurrent email sending in bulk operations
- Rate limiter shared across goroutines
- SMTP connection state management
- Template rendering in parallel

### Concurrency Primitives

```go
// SMTP - Connection protection
type client struct {
    mu sync.Mutex  // Protects connection state
    conn *smtp.Client
}

// Queuer - Rate limiting
type counter struct {
    mu sync.Mutex     // Protects counter and timer
    c  atomic.Int64   // Atomic counter for thread-safe reads
}

// Sender - Immutable after construction
// (Not thread-safe for modification, but safe for concurrent sending)

// Render - Stateless rendering
// (Thread-safe when using separate instances)
```

### Verified Components

| Component | Mechanism | Validation | Status |
|-----------|-----------|------------|--------|
| SMTP Client | `sync.Mutex` | Race detector | ✅ |
| Queuer Counter | `sync.Mutex` + `atomic.Int64` | Race detector + stress tests | ✅ |
| Queuer Reset | `sync.WaitGroup` | Lifecycle tests | ✅ |
| Sender Construction | Immutable | Concurrent send tests | ✅ |
| Render Cloning | Deep copy | Parallel render tests | ✅ |

### Testing Commands

```bash
# Full race detection
CGO_ENABLED=1 go test -race ./...

# Focus on concurrent components
CGO_ENABLED=1 go test -race ./queuer
CGO_ENABLED=1 go test -race ./smtp

# Stress test (multiple runs)
for i in {1..10}; do 
    CGO_ENABLED=1 go test -race ./... || break
done

# Specific concurrency tests
CGO_ENABLED=1 go test -race -run "Concurrent" ./...
CGO_ENABLED=1 go test -race -run "Parallel" ./...
```

**Result**: Zero data races across all test runs ✅

---

## Performance Benchmarks

### SMTP TLSMode Benchmarks

Located in `smtp/tlsmode/benchmark_test.go` using gmeasure:

| Benchmark | Mean Time | Notes |
|-----------|-----------|-------|
| String parsing | ~70ns | Parse "starttls" → TLSStartTLS |
| Int64 parsing | ~52ns | Parse int → TLSMode |
| JSON roundtrip | ~1.5µs | Marshal + Unmarshal |
| String roundtrip | ~70ns | String() + Parse() |

**Parsing Method Comparison**:
- Int parsing: **44ns** (fastest)
- Bytes parsing: **67ns**
- String parsing: **113ns**

**TLS Mode Comparison**:
- TLSNone: **72ns**
- TLSStartTLS: **75ns**
- TLSStrictTLS: **66ns**

**Stress Tests**:
- 100 rapid parses: **9.7µs** (~97ns each)
- 300 rapid conversions: **2.6µs** (~8.7ns each)

### Expected Performance

| Operation | Expected Time | Notes |
|-----------|---------------|-------|
| SMTP Connect | 50-500ms | Network dependent |
| SMTP Send | 100-2000ms | Network + server processing |
| Email Compose | <1ms | Memory operations |
| Template Render | 1-10ms | Complexity dependent |
| Queuer Throttle | 0-60s | Based on configuration |

---

## Subpackage Testing

### SMTP Tests

**Test Files** (104 specs):
- Connection and authentication
- TLS mode handling and fallback
- DSN parsing and configuration
- Health monitoring
- Error handling

**Key Scenarios**:
- ✅ STARTTLS upgrade (port 587)
- ✅ Strict TLS direct (port 465)
- ✅ Plain SMTP (port 25)
- ✅ TLS fallback (Strict → STARTTLS)
- ✅ CR/LF injection prevention
- ✅ Certificate validation

**External Dependencies**: Uses standalone SMTP server for testing (no external services)

**Documentation**: [smtp/TESTING.md](smtp/TESTING.md)

---

### SMTP Config Tests

**Test Files** (222 specs):
- DSN parsing with various formats
- URL encoding/decoding
- Configuration validation
- Default value handling
- Clone operations
- JSON/YAML/TOML encoding

**Coverage**: 92.7%

---

### SMTP TLSMode Tests

**Test Files** (165 specs):
- All TLS mode constants
- String/bytes/int parsing
- JSON/YAML/TOML encoding
- Roundtrip conversions
- Error handling
- Performance benchmarks (with gmeasure)

**Coverage**: 98.8% (highest in package)

---

### Sender Tests

**Test Files** (252 specs):
- Email composition and structure
- Multi-part content (HTML + text)
- File attachments (regular & inline)
- Recipient management (To, CC, BCC)
- Address parsing and validation
- Custom headers
- Priority levels
- Transfer encodings
- SMTP integration

**Key Scenarios**:
- ✅ RFC-compliant message generation
- ✅ Attachment encoding (Base64, QP)
- ✅ Inline image embedding
- ✅ Recipient deduplication
- ✅ Custom header handling
- ✅ Configuration via JSON/YAML

**Documentation**: [sender/TESTING.md](sender/TESTING.md)

---

### Render Tests

**Test Files** (123 specs):
- Theme rendering (Default, Flat)
- HTML and plain text generation
- Variable substitution (`{{var}}`)
- Body components (Intro, Outro, Actions, Tables, Dictionaries)
- Text direction (LTR, RTL)
- Configuration validation
- Clone operations
- Error handling

**Key Scenarios**:
- ✅ Complete email body rendering
- ✅ Theme-specific styling
- ✅ Complex nested structures
- ✅ Variable replacement
- ✅ Bidirectional text support

**Documentation**: [render/TESTING.md](render/TESTING.md)

---

### Queuer Tests

**Test Files** (101 specs, 1 skipped):
- Rate limiting algorithm
- Concurrent sending
- Context cancellation
- Throttle wait timing
- SMTP interface compliance
- Configuration handling
- Callback invocation
- Clone operations
- Health monitoring

**Key Scenarios**:
- ✅ Rate limit enforcement
- ✅ Context cancellation during wait
- ✅ Concurrent sender stress tests
- ✅ Counter overflow handling
- ✅ Thread-safe operations

**Coverage**: 90.8% (highest for main packages)

---

## Writing Tests

### Test Structure

Tests follow Ginkgo's BDD hierarchy:

```go
var _ = Describe("mail/component", func() {
    var (
        // Test variables
        client smtp.SMTP
        cfg    smtp.Config
    )

    BeforeEach(func() {
        // Per-test setup
        client = smtp.New()
        cfg = smtp.NewConfig()
    })

    AfterEach(func() {
        // Per-test cleanup
        if client != nil {
            client.Close()
        }
    })

    Context("Feature description", func() {
        It("should do something specific", func() {
            // Arrange
            cfg.SetHost("smtp.example.com")
            client.SetConfig(cfg)
            
            // Act
            err := client.Check()
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
        })
    })
})
```

### Guidelines

**1. Use Descriptive Names**
```go
It("should throttle after exceeding rate limit", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should send email with attachment", func() {
    // Arrange
    mail := sender.New()
    mail.SetFrom("sender@example.com", "")
    mail.AttachFile("test.pdf")
    
    // Act
    err := mail.Send(ctx, client)
    
    // Assert
    Expect(err).ToNot(HaveOccurred())
})
```

**3. Use Appropriate Matchers**
```go
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))
Expect(list).To(ContainElement(item))
Expect(number).To(BeNumerically(">", 0))
Expect(str).To(MatchRegexp("^[a-z]+$"))
```

**4. Always Cleanup Resources**
```go
defer client.Close()
defer os.Remove(tempFile)
defer cancel() // context cancellation
```

**5. Test Edge Cases**
- Empty/nil inputs
- Large data volumes
- Concurrent access
- Network failures
- Timeouts

**6. Avoid External Dependencies**
- Use mock SMTP servers (queuer tests)
- Generate test data in-memory
- No real email sending in tests
- No external API calls

### Test Template

```go
var _ = Describe("mail/new_feature", func() {
    Context("When using new feature", func() {
        var (
            testData []byte
            result   interface{}
        )

        BeforeEach(func() {
            testData = []byte("test data")
        })

        It("should perform expected behavior", func() {
            // Arrange
            input := prepareInput(testData)
            
            // Act
            result, err := newFeature(input)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result).ToNot(BeNil())
        })

        It("should handle error case", func() {
            // Act
            _, err := newFeature(invalidInput)
            
            // Assert
            Expect(err).To(HaveOccurred())
        })
    })
})
```

---

## Best Practices

### Test Independence

- ✅ Each test should run independently
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Avoid shared mutable state
- ✅ Generate test data on-demand
- ❌ Don't rely on test execution order

### Assertions

```go
// ✅ Good: Specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(email.From).To(Equal("sender@example.com"))
Expect(emails).To(HaveLen(10))

// ❌ Bad: Generic comparisons
Expect(err == nil).To(BeTrue())
Expect(email.From == "sender@example.com").To(BeTrue())
```

### Concurrency Testing

```go
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            if err := operation(id); err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    Expect(errors).To(BeEmpty())
})
```

**Always run with `-race` during development**

### Performance

- Keep tests fast (<100ms per spec typically)
- Use parallel execution when safe (`ginkgo -p`)
- Mock external dependencies
- Avoid unnecessary sleep statements

**Current Performance**:
- Target: <50ms per spec average
- Actual: ~40ms per spec average
- SMTP tests slower due to network operations (expected)

### Error Handling

```go
// ✅ Good: Check all errors
It("should handle errors properly", func() {
    result, err := operation()
    Expect(err).ToNot(HaveOccurred())
    
    err = result.Process()
    Expect(err).ToNot(HaveOccurred())
    
    defer result.Close()
})

// ❌ Bad: Ignore errors
It("should do something", func() {
    result, _ := operation()
    result.Process() // Return value ignored
})
```

---

## Troubleshooting

### Common Issues

**1. SMTP Connection Timeouts**
```bash
# Increase timeout
go test -timeout=15m ./smtp

# Issue: Network slowness or blocked ports
# Solution: Check firewall, use mock server
```

**2. Race Conditions Detected**
```bash
# Run with race detector
CGO_ENABLED=1 go test -race ./...

# Issue: Unprotected concurrent access
# Solution: Add mutex protection or atomic operations
```

**3. Coverage Report Generation Fails**
```bash
# Clean test cache
go clean -testcache

# Regenerate
go test -coverprofile=coverage.out ./...
```

**4. Ginkgo CLI Not Found**
```bash
# Install Ginkgo
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Verify installation
ginkgo version
```

**5. CGO Not Available (Race Detector)**
```bash
# Install build tools
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS
xcode-select --install

# Verify
export CGO_ENABLED=1
go test -race ./...
```

**6. Test Timeout**
```bash
# Issue: Long-running SMTP tests
# Solution: Increase timeout
go test -timeout=20m ./...
```

### Debugging Tests

```bash
# Run specific test
ginkgo --focus="should send email with STARTTLS"

# Run specific file
ginkgo --focus-file=client_test.go

# Verbose output with stack traces
ginkgo -v --trace

# Stop on first failure
ginkgo --fail-fast

# Run skipped tests
ginkgo --keep-going
```

**Debug Output in Tests**:
```go
fmt.Fprintf(GinkgoWriter, "Debug: value = %+v\n", value)
```

---

## CI Integration

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v -timeout=10m ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -timeout=10m ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out
```

### GitLab CI

```yaml
test:
  stage: test
  image: golang:1.21
  script:
    - go test -v -timeout=10m ./...
    - CGO_ENABLED=1 go test -race -timeout=10m ./...
    - go test -coverprofile=coverage.out ./...
  coverage: '/coverage: \d+\.\d+% of statements/'
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./... || exit 1

echo "Running race detector..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
go test -cover ./... | grep -E "coverage:" || exit 1

echo "All checks passed!"
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
- [ ] Coverage maintained: ≥85% overall, ≥80% per subpackage
- [ ] New features have tests (unit + integration)
- [ ] Error cases tested
- [ ] Thread safety validated (if applicable)
- [ ] Documentation updated (README, TESTING, GoDoc)
- [ ] Examples provided for new features
- [ ] Benchmarks added for performance-critical code
- [ ] No test flakiness (run 3+ times)
- [ ] Test duration reasonable (<1min total preferred)

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

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)
- [gmeasure Documentation](https://onsi.github.io/gomega/gmeasure/)

**Email Standards**
- [RFC 5321 - SMTP](https://tools.ietf.org/html/rfc5321)
- [RFC 822 - Email Format](https://tools.ietf.org/html/rfc822)
- [RFC 2045 - MIME](https://tools.ietf.org/html/rfc2045)

**Subpackage Testing Guides**
- [SMTP Testing Guide](smtp/TESTING.md)
- [Sender Testing Guide](sender/TESTING.md)
- [Render Testing Guide](render/TESTING.md)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Test Framework**: Ginkgo v2 + Gomega  
**Maintained By**: Mail Package Contributors
