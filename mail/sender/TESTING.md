# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-252%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-81.4%25-brightgreen)]()

Comprehensive testing documentation for the mail/sender package, covering test execution, race detection, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Test File Organization](#test-file-organization)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The mail/sender package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions and real SMTP server integration.

**Test Suite Statistics**
- Total Specs: 252
- Coverage: 81.4%
- Race Detection: ✅ Zero data races
- Execution Time: ~0.8s (without race), ~1.5s (with race)

**Coverage Areas**
- Mail interface operations (creation, properties, headers)
- Email address management (From, To, CC, BCC)
- Body and attachment handling
- Configuration validation and creation
- Sender creation and lifecycle
- Error handling and edge cases
- Type parsing (encoding, priority, content type)
- Real SMTP server integration for send operations
- Performance benchmarks with gmeasure

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
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

**Expected Output:**
```
Ran 252 of 252 Specs in 0.808 seconds
SUCCESS! -- 252 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS
coverage: 81.4% of statements
ok  	github.com/nabbar/golib/mail/sender	0.839s
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Rich CLI with filtering and focus
- Detailed failure reporting

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages
- Custom matchers support

**gmeasure** - Performance benchmarking ([docs](https://onsi.github.io/gomega/#gmeasure-benchmarking-code))
- Built on top of Gomega
- Statistical analysis (mean, median, stddev)
- Performance regression detection
- Experiment-based benchmarking

---

## Running Tests

### Basic Commands

```bash
# Standard test run
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Specific timeout
go test -timeout=10m ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=mail_test.go

# Pattern matching
ginkgo --focus="Email address"

# Skip specific tests
ginkgo --skip="Performance"

# Parallel execution (use with caution - SMTP server conflicts)
ginkgo -p --procs=4

# JUnit report
ginkgo --junit-report=results.xml

# JSON report
ginkgo --json-report=results.json
```

### Race Detection

**Critical for validating thread-safety assumptions**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# With coverage and race detection
CGO_ENABLED=1 go test -race -cover -covermode=atomic ./...
```

**What it validates:**
- Mail object cloning safety
- Concurrent sender creation
- SMTP server state management
- No data races in test helpers

**Expected Output:**
```bash
# ✅ Success
Ran 252 of 252 Specs in 1.553 seconds
SUCCESS! -- 252 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS
ok  	github.com/nabbar/golib/mail/sender	2.738s

# ❌ Race detected (example - should never occur)
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: ✅ Zero data races detected

### Performance & Profiling

```bash
# View benchmark results during tests
go test -v ./...

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

**Performance Expectations**

| Test Type | Duration | Notes |
|-----------|----------|-------|
| Full Suite | ~0.8s | Without race |
| With `-race` | ~1.5s | 2x slower (expected) |
| Individual Spec | <50ms | Most tests |
| SMTP Tests | 50-200ms | Server startup overhead |
| Benchmarks | Included | gmeasure reports |

---

## Test Coverage

**Target**: ≥80% statement coverage  
**Current**: 81.4%

### Coverage By Category

| Category | Files | Specs | Description |
|----------|-------|-------|-------------|
| **Mail Operations** | `mail_test.go` | 54 | Create, clone, properties, headers, body, attachments |
| **Email Addresses** | `email_test.go` | 37 | From, sender, replyTo, recipients, fallback behavior |
| **Sending** | `send_test.go` | 25 | Sender creation, SMTP integration, context handling |
| **Configuration** | `config_test.go` | 43 | Validation, mailer creation, file attachments |
| **Types** | `types_test.go` | 29 | Encoding, priority, content type, recipient type |
| **Errors** | `errors_test.go` | 32 | Error codes, wrapping, validation, edge cases |
| **Edge Cases** | `edge_cases_test.go` | 20 | Empty values, large data, special chars, limits |
| **Benchmarks** | `benchmark_test.go` | 12 | Performance measurements with gmeasure |

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out -covermode=atomic ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

### Coverage Analysis

**Well-Covered Areas (>85%)**
- Mail interface implementation
- Email address management
- Configuration validation
- Type parsing and conversion
- Error handling

**Areas for Improvement (<75%)**
- Some edge cases in sender.go
- Complex error scenarios
- Date/time parsing edge cases

### Test Structure

Tests follow Ginkgo's hierarchical BDD structure:

```go
var _ = Describe("mail/sender Component", func() {
    var (
        ctx context.Context
        cnl context.CancelFunc
    )
    
    BeforeEach(func() {
        ctx, cnl = context.WithTimeout(context.Background(), 30*time.Second)
    })
    
    AfterEach(func() {
        if cnl != nil {
            cnl()
        }
    })
    
    Context("Feature or scenario", func() {
        It("should do something specific", func() {
            mail := sender.New()
            mail.SetSubject("Test")
            
            Expect(mail.GetSubject()).To(Equal("Test"))
        })
    })
})
```

---

## Thread Safety

The mail/sender package has specific thread-safety characteristics that are validated through testing.

### Thread-Safety Model

**NOT Thread-Safe (by design):**
- `Mail` objects - Clone for concurrent use
- `Email` interface - Part of Mail
- `Sender` objects - One-time use

**Thread-Safe:**
- `Config` objects - Immutable after creation
- Type parsing functions - Pure functions
- Error codes - Constants

### Concurrent Usage Pattern

```go
// ✅ Correct: Clone for concurrency
template := sender.New()
template.SetSubject("Notification")

for i := 0; i < 10; i++ {
    go func(index int) {
        mail := template.Clone() // Independent copy
        mail.Email().AddRecipients(sender.RecipientTo, fmt.Sprintf("user%d@example.com", index))
        // Use mail...
    }(i)
}

// ❌ Incorrect: Shared Mail object
mail := sender.New()
for i := 0; i < 10; i++ {
    go func() {
        mail.SetSubject("Test") // DATA RACE!
    }()
}
```

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent scenarios
CGO_ENABLED=1 go test -race -v -run "Clone\|Concurrent" ./...

# Stress test (10 iterations)
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done
```

**Result**: ✅ Zero data races across all test runs

---

## Test File Organization

| File | Purpose | Specs | Notes |
|------|---------|-------|-------|
| `sender_suite_test.go` | Suite initialization, test helpers | 1 | SMTP server setup |
| `mail_test.go` | Mail interface operations | 54 | Core functionality |
| `email_test.go` | Email address management | 37 | Recipients, addresses |
| `send_test.go` | Sending operations | 25 | Real SMTP integration |
| `config_test.go` | Configuration operations | 43 | Validation, creation |
| `types_test.go` | Type definitions | 29 | Enums, parsing |
| `errors_test.go` | Error handling | 32 | Error codes, wrapping |
| `edge_cases_test.go` | Edge cases and limits | 20 | Boundary conditions |
| `benchmark_test.go` | Performance benchmarks | 12 | gmeasure stats |

### Test Helper Infrastructure

**`sender_suite_test.go` provides:**

```go
// Context management
var testCtx context.Context
var testCtxCancel context.CancelFunc

// SMTP test server backend
type testBackend struct {
    requireAuth bool
    messages    []testMessage
}

// SMTP session for testing
type testSession struct {
    backend *testBackend
    from    string
    to      []string
}

// Helper functions
func startTestSMTPServer(backend *testBackend, useTLS bool) (*smtpsv.Server, string, int, error)
func newTestConfig(host string, port int, tlsMode smtptp.TLSMode) *smtpcfg.Config
func newTestSMTPClient(host string, port int) libsmtp.SMTP
func newReadCloser(s string) io.ReadCloser
```

---

## Writing Tests

### Test File Template

```go
package sender_test

import (
    "context"
    "time"
    
    libsnd "github.com/nabbar/golib/mail/sender"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("Component Name", func() {
    var (
        ctx  context.Context
        cnl  context.CancelFunc
        mail libsnd.Mail
    )
    
    BeforeEach(func() {
        ctx, cnl = context.WithTimeout(testCtx, 10*time.Second)
        mail = libsnd.New()
    })
    
    AfterEach(func() {
        if cnl != nil {
            cnl()
        }
    })
    
    Context("Scenario description", func() {
        It("should have expected behavior", func() {
            // Arrange
            mail.SetSubject("Test")
            
            // Act
            result := mail.GetSubject()
            
            // Assert
            Expect(result).To(Equal("Test"))
        })
    })
})
```

### Common Gomega Matchers

```go
// Equality
Expect(value).To(Equal(expected))
Expect(value).ToNot(Equal(unexpected))

// Nil checks
Expect(err).ToNot(HaveOccurred())
Expect(ptr).To(BeNil())

// Strings
Expect(str).To(ContainSubstring("substring"))
Expect(str).To(HavePrefix("prefix"))
Expect(str).To(MatchRegexp("pattern"))

// Collections
Expect(slice).To(HaveLen(5))
Expect(slice).To(ContainElement("item"))
Expect(slice).To(BeEmpty())

// Numeric
Expect(num).To(BeNumerically(">", 0))
Expect(num).To(BeNumerically("~", 1.5, 0.1))

// Boolean
Expect(condition).To(BeTrue())
Expect(condition).To(BeFalse())

// Types
Expect(obj).To(BeAssignableToTypeOf(&Type{}))
```

### Benchmark Tests with gmeasure

```go
It("should measure performance", func() {
    experiment := gmeasure.NewExperiment("Operation Name")
    AddReportEntry(experiment.Name, experiment)
    
    experiment.Sample(func(idx int) {
        experiment.MeasureDuration("operation", func() {
            // Code to benchmark
            mail := sender.New()
            mail.SetSubject("Test")
        })
    }, gmeasure.SamplingConfig{N: 1000})
    
    stats := experiment.GetStats("operation")
    AddReportEntry(experiment.Name+" Stats", 
        fmt.Sprintf("Mean: %v, StdDev: %v", stats.DurationFor(gmeasure.StatMean), 
        stats.DurationFor(gmeasure.StatStdDev)))
})
```

### SMTP Integration Tests

```go
It("should send email via SMTP", func() {
    // Setup SMTP test server
    backend := &testBackend{requireAuth: false, messages: make([]testMessage, 0)}
    smtpServer, host, port, err := startTestSMTPServer(backend, false)
    Expect(err).ToNot(HaveOccurred())
    defer smtpServer.Close()
    
    // Create SMTP client
    smtpClient := newTestSMTPClient(host, port)
    defer smtpClient.Close()
    
    // Create and configure email
    mail := sender.New()
    mail.SetSubject("Test")
    mail.Email().SetFrom("sender@test.com")
    mail.Email().AddRecipients(sender.RecipientTo, "recipient@test.com")
    body := newReadCloser("Test body")
    mail.SetBody(sender.ContentPlainText, body)
    
    // Send
    snd, err := mail.Sender()
    Expect(err).ToNot(HaveOccurred())
    defer snd.Close()
    
    err = snd.Send(ctx, smtpClient)
    Expect(err).ToNot(HaveOccurred())
    
    // Verify
    Expect(backend.messages).To(HaveLen(1))
    Expect(backend.messages[0].from).To(Equal("sender@test.com"))
})
```

---

## Best Practices

**Organize Tests Logically**
```go
// ✅ Good: Clear hierarchy
Describe("Mail Interface", func() {
    Context("Subject Operations", func() {
        It("should set subject", func() { /* ... */ })
        It("should get subject", func() { /* ... */ })
    })
    
    Context("Body Operations", func() {
        It("should set plain text body", func() { /* ... */ })
        It("should add HTML body", func() { /* ... */ })
    })
})

// ❌ Bad: Flat structure
Describe("Tests", func() {
    It("test 1", func() { /* ... */ })
    It("test 2", func() { /* ... */ })
    It("test 3", func() { /* ... */ })
})
```

**Use Descriptive Names**
```go
// ✅ Good: Clear intent
It("should return error when from address is invalid", func() { /* ... */ })

// ❌ Bad: Vague
It("should fail", func() { /* ... */ })
```

**Clean Up Resources**
```go
// ✅ Good: Proper cleanup
It("should send with attachment", func() {
    file, _ := os.Open("test.pdf")
    // Don't defer here - let sender close it
    mail.AddAttachment("test.pdf", "application/pdf", file, false)
    
    snd, _ := mail.Sender()
    defer snd.Close() // Closes file too
})

// ❌ Bad: Resource leak
It("should send", func() {
    file, _ := os.Open("test.pdf")
    mail.AddAttachment("test.pdf", "application/pdf", file, false)
    // File never closed if test fails
})
```

**Test Edge Cases**
```go
Context("Edge Cases", func() {
    It("should handle empty subject", func() {
        mail.SetSubject("")
        Expect(mail.GetSubject()).To(Equal(""))
    })
    
    It("should handle very long subject", func() {
        longSubject := strings.Repeat("A", 10000)
        mail.SetSubject(longSubject)
        Expect(mail.GetSubject()).To(Equal(longSubject))
    })
    
    It("should handle special characters", func() {
        mail.SetSubject("Test 日本語 émojis 🎉")
        Expect(mail.GetSubject()).To(ContainSubstring("🎉"))
    })
})
```

**Use Table-Driven Tests for Patterns**
```go
DescribeTable("Encoding parsing",
    func(input string, expected sender.Encoding) {
        result := sender.ParseEncoding(input)
        Expect(result).To(Equal(expected))
    },
    Entry("Base64", "Base 64", sender.EncodingBase64),
    Entry("Quoted-Printable", "Quoted Printable", sender.EncodingQuotedPrintable),
    Entry("Case insensitive", "base 64", sender.EncodingBase64),
    Entry("Invalid defaults to None", "invalid", sender.EncodingNone),
)
```

---

## Troubleshooting

### Common Issues

**1. Race Detector Not Working**

```bash
# Problem: Race detector requires CGO
go test -race ./...
# Error: -race requires cgo

# Solution: Enable CGO
CGO_ENABLED=1 go test -race ./...
```

**2. Tests Timeout**

```bash
# Problem: Default 10m timeout exceeded
go test ./...
# FAIL: timeout after 10m0s

# Solution: Increase timeout
go test -timeout=30m ./...
```

**3. SMTP Server Port Conflicts**

```bash
# Problem: Port already in use
# Error: bind: address already in use

# Solution 1: Kill conflicting process
lsof -i :25 | grep LISTEN | awk '{print $2}' | xargs kill

# Solution 2: Tests use random free ports (already implemented)
# The test suite automatically finds free ports
```

**4. Coverage Report Not Generated**

```bash
# Problem: Missing coverage file
go tool cover -html=coverage.out
# Error: coverage.out: no such file

# Solution: Generate coverage first
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**5. Ginkgo Not Found**

```bash
# Problem: Ginkgo CLI not installed
ginkgo
# Error: command not found

# Solution: Install Ginkgo
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Verify installation
ginkgo version
```

### Debug Failing Tests

```bash
# Run specific test
go test -v -run "TestSender/Mail_Operations/Subject"

# With Ginkgo focus
ginkgo --focus="should set subject"

# Print debug output (add to test)
It("should work", func() {
    GinkgoWriter.Println("Debug:", value)
    // Test code...
})

# Fail fast on first error
go test -failfast ./...
```

### Performance Issues

```bash
# Identify slow tests
go test -v ./... | grep "seconds"

# Profile CPU usage
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out
# In pprof: top10, list <function>

# Profile memory
go test -memprofile=mem.out ./...
go tool pprof mem.out
# In pprof: top10, list <function>
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
        go-version: ['1.18', '1.19', '1.20', '1.21']
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v -cover ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -v ./...
      
      - name: Generate coverage
        run: go test -coverprofile=coverage.out -covermode=atomic ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
```

### GitLab CI Example

```yaml
test:
  image: golang:1.21
  stage: test
  script:
    - go mod download
    - go test -v -cover ./...
    - CGO_ENABLED=1 go test -race ./...
  coverage: '/coverage: \d+\.\d+% of statements/'
  artifacts:
    reports:
      junit: report.xml
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any
    
    stages {
        stage('Test') {
            steps {
                sh 'go mod download'
                sh 'go test -v -cover ./...'
            }
        }
        
        stage('Race Detection') {
            steps {
                sh 'CGO_ENABLED=1 go test -race ./...'
            }
        }
        
        stage('Coverage') {
            steps {
                sh 'go test -coverprofile=coverage.out ./...'
                sh 'go tool cover -html=coverage.out -o coverage.html'
                publishHTML([reportDir: '.', reportFiles: 'coverage.html', reportName: 'Coverage Report'])
            }
        }
    }
}
```

### Pre-Commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./...
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Running race detector..."
CGO_ENABLED=1 go test -race ./...
if [ $? -ne 0 ]; then
    echo "Race detector found issues. Commit aborted."
    exit 1
fi

echo "All checks passed!"
exit 0
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

- **Ginkgo Documentation**: [https://onsi.github.io/ginkgo/](https://onsi.github.io/ginkgo/)
- **Gomega Documentation**: [https://onsi.github.io/gomega/](https://onsi.github.io/gomega/)
- **Go Testing**: [https://golang.org/pkg/testing/](https://golang.org/pkg/testing/)
- **Race Detector**: [https://go.dev/doc/articles/race_detector](https://go.dev/doc/articles/race_detector)
- **Package Documentation**: [README.md](README.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
