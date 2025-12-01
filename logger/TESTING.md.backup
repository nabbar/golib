# Testing Guide

Comprehensive testing documentation for the logger package and its subpackages.

> **AI Disclaimer**: AI tools are used solely to assist with testing, documentation, and bug fixes under human supervision, in compliance with EU AI Act Article 50.4.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety Testing](#thread-safety-testing)
- [Testing Strategies](#testing-strategies)
- [Package-Specific Tests](#package-specific-tests)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The logger package provides structured logging with multiple output destinations, field injection, and extensive integration capabilities. Testing requires careful validation of output formatting, hook execution, level filtering, and thread safety.

### Test Characteristics

- **Framework**: Ginkgo v2 + Gomega
- **Execution**: I/O dependent (file operations, network)
- **Concurrency**: Thread-safe operations validated with race detector
- **Dependencies**: logrus, context, ioutils

### Coverage Areas

Latest test results (705 total specs, ~77% average coverage):

| Package | Specs | Coverage | Key Areas |
|---------|-------|----------|-----------|
| **logger** | 81 | 75.0% | Core logging, io.Writer, cloning |
| **config** | 127 | 85.3% | Options, validation, serialization |
| **entry** | 119 | 85.1% | Entry creation, formatting, fields |
| **fields** | 49 | 78.4% | Field operations, merging, cloning |
| **gorm** | 34 | 100.0% | GORM adapter (perfect coverage) |
| **hashicorp** | 89 | 96.6% | Hashicorp adapter |
| **hookfile** | 22 | 20.1% | File output, rotation |
| **hookstderr** | 30 | 100.0% | Stderr output (perfect coverage) |
| **hookstdout** | 30 | 100.0% | Stdout output (perfect coverage) |
| **hooksyslog** | 20 | 53.5% | Syslog protocol, network |
| **hookwriter** | 31 | 90.2% | Custom writer integration |
| **level** | 42 | 65.9% | Level parsing, comparison |
| **types** | 32 | N/A | Type definitions |
| **TOTAL** | **705** | **~77%** | **All tests passing** |

---

## Quick Start

```bash
# Install test dependencies (if not already installed)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (critical!)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
ginkgo -v -race -cover

# Package-specific
go test ./logger/config
go test ./logger/entry
go test ./logger/fields
```

---

## Test Framework

### Ginkgo v2

**BDD testing framework** - [Documentation](https://onsi.github.io/ginkgo/)

Features used:
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Focused tests (`FDescribe`, `FIt`) and skip (`XDescribe`, `XIt`)
- Parallel execution support
- Rich reporting

### Gomega

**Matcher library** - [Documentation](https://onsi.github.io/gomega/)

Critical matchers for this package:
- `HaveOccurred()`: Error checking
- `Equal()`: Value equality
- `BeNil()`: Nil checks
- `ContainSubstring()`: String matching
- `HaveLen()`: Collection size
- `BeEmpty()`: Empty checks

---

## Running Tests

### Basic Commands

```bash
# All packages
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test ./logger
go test ./logger/config
go test ./logger/entry

# With short flag (skip long-running tests)
go test -short ./...
```

### Coverage

```bash
# Basic coverage
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Per-package coverage
go test -coverprofile=logger_coverage.out ./logger
go test -coverprofile=config_coverage.out ./logger/config
go test -coverprofile=entry_coverage.out ./logger/entry

# Coverage summary
go test -cover ./... 2>&1 | grep -E "coverage:|ok"
```

### Race Detection

**Critical for this package** - Always run before commits:

```bash
# Enable race detector
CGO_ENABLED=1 go test -race ./...

# With timeout
CGO_ENABLED=1 go test -race -timeout=5m ./...

# Specific package
CGO_ENABLED=1 go test -race ./logger/config
```

### Ginkgo CLI

```bash
# Install
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
ginkgo

# With options
ginkgo -v -race -cover

# Specific package
ginkgo ./logger/config

# Focused tests only
ginkgo -focus="Level"

# Skip tests
ginkgo -skip="Integration.*"

# Parallel execution
ginkgo -p
```

---

## Test Coverage

### Current Coverage

Detailed coverage by package (705 total specs):

| Package | Specs | Coverage | Key Areas |
|---------|-------|----------|-----------|
| **logger** | 81 | 75.0% | Core logging, io.Writer, cloning |
| **config** | 127 | 85.3% | Options, validation, serialization |
| **entry** | 119 | 85.1% | Entry creation, formatting, fields |
| **fields** | 49 | 78.4% | Field operations, merging, cloning |
| **gorm** | 34 | 100.0% | GORM adapter (perfect coverage) |
| **hashicorp** | 89 | 96.6% | Hashicorp adapter |
| **hookfile** | 22 | 20.1% | File output, rotation |
| **hookstderr** | 30 | 100.0% | Stderr output (perfect coverage) |
| **hookstdout** | 30 | 100.0% | Stdout output (perfect coverage) |
| **hooksyslog** | 20 | 53.5% | Syslog protocol, network |
| **hookwriter** | 31 | 90.2% | Custom writer integration |
| **level** | 42 | 65.9% | Level parsing, comparison |
| **types** | 32 | N/A | Type definitions |

**Overall**: 705 specs, ~77% average coverage, all tests passing

### Test Distribution

```
Total Specs: 705
├── Core Logger: 81 specs (75.0%)
├── Configuration: 127 specs (85.3%)
├── Entry Management: 119 specs (85.1%)
├── Fields: 49 specs (78.4%)
├── Hooks: 133 specs (varied coverage)
└── Integrations: 196 specs (98.3%)
```

### Coverage Highlights

**Perfect Coverage (100%)**:
- `gorm` - GORM integration adapter
- `hookstderr` - Stderr output hook
- `hookstdout` - Stdout output hook

**Excellent Coverage (>90%)**:
- `hashicorp` - 96.6% - Hashicorp tools adapter
- `hookwriter` - 90.2% - Custom writer hook

**Good Coverage (75-85%)**:
- `config` - 85.3% - Configuration management
- `entry` - 85.1% - Log entry handling
- `fields` - 78.4% - Field operations
- `logger` - 75.0% - Core logging

**Areas for Improvement**:
- `hookfile` - 20.1% - File rotation (complex I/O scenarios)
- `hooksyslog` - 53.5% - Syslog protocol (network edge cases)
- `level` - 65.9% - Level parsing (edge cases)

---

## Thread Safety Testing

### Race Detector

**Always run with `-race` flag**:

```bash
CGO_ENABLED=1 go test -race ./...
```

### Common Race Conditions to Test

**1. Concurrent Logging**

```go
It("should handle concurrent logging", func() {
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            log.Info(fmt.Sprintf("Message %d", idx), nil)
        }(i)
    }
    wg.Wait()
})
```

**2. Concurrent Level Changes**

```go
It("should handle concurrent level changes", func() {
    var wg sync.WaitGroup
    
    // Writers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            log.SetLevel(level.InfoLevel)
            log.SetLevel(level.DebugLevel)
        }()
    }
    
    // Readers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = log.GetLevel()
        }()
    }
    
    wg.Wait()
})
```

**3. Concurrent Field Operations**

```go
It("should handle concurrent field operations", func() {
    flds := fields.New()
    
    var wg sync.WaitGroup
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            key := fmt.Sprintf("key%d", idx)
            flds.Add(key, idx)
            _ = flds.Get(key)
        }(i)
    }
    wg.Wait()
})
```

---

## Testing Strategies

### 1. Output Validation

Test that logs are written correctly:

```go
Describe("File Output", func() {
    var (
        log      Logger
        tempFile string
    )
    
    BeforeEach(func() {
        tempFile = filepath.Join(os.TempDir(), "test.log")
        opts := &config.Options{
            LogFile: &config.OptionsFile{
                LogFileName: tempFile,
            },
        }
        
        log, _ = New(context.Background)
        log.SetOptions(opts)
    })
    
    AfterEach(func() {
        log.Close()
        os.Remove(tempFile)
    })
    
    It("should write log to file", func() {
        log.Info("Test message", map[string]interface{}{
            "key": "value",
        })
        
        // Ensure flush
        log.Close()
        
        // Read file
        content, err := os.ReadFile(tempFile)
        Expect(err).ToNot(HaveOccurred())
        Expect(string(content)).To(ContainSubstring("Test message"))
        Expect(string(content)).To(ContainSubstring("key"))
        Expect(string(content)).To(ContainSubstring("value"))
    })
})
```

### 2. Level Filtering

Test that log levels are respected:

```go
Describe("Level Filtering", func() {
    It("should filter by level", func() {
        var buf bytes.Buffer
        
        log, _ := New(context.Background)
        // Configure to write to buffer
        
        log.SetLevel(level.InfoLevel)
        
        log.Trace("Trace message", nil)  // Filtered
        log.Debug("Debug message", nil)  // Filtered
        log.Info("Info message", nil)    // Logged
        log.Warning("Warning", nil)      // Logged
        
        output := buf.String()
        Expect(output).ToNot(ContainSubstring("Trace message"))
        Expect(output).ToNot(ContainSubstring("Debug message"))
        Expect(output).To(ContainSubstring("Info message"))
        Expect(output).To(ContainSubstring("Warning"))
    })
})
```

### 3. Field Merging

Test field merging behavior:

```go
Describe("Field Merging", func() {
    It("should merge persistent and per-entry fields", func() {
        // Persistent fields
        persistFields := fields.NewFromMap(map[string]interface{}{
            "app": "test",
            "version": "1.0",
        })
        log.SetFields(persistFields)
        
        // Per-entry fields
        log.Info("Message", map[string]interface{}{
            "request_id": "123",
            "version": "2.0",  // Override
        })
        
        // Verify output contains both sets
        // Per-entry fields should override persistent
    })
})
```

### 4. Configuration Validation

Test configuration validation:

```go
Describe("Configuration Validation", func() {
    It("should validate file configuration", func() {
        opts := &config.Options{
            LogFile: &config.OptionsFile{
                LogFileName: "",  // Invalid
            },
        }
        
        err := opts.Validate()
        Expect(err).To(HaveOccurred())
    })
    
    It("should accept valid configuration", func() {
        opts := &config.Options{
            LogLevel: level.InfoLevel,
            LogFormatter: config.FormatJSON,
            LogFile: &config.OptionsFile{
                LogFileName: "/tmp/test.log",
                LogFileMaxSize: 100,
            },
        }
        
        err := opts.Validate()
        Expect(err).ToNot(HaveOccurred())
    })
})
```

### 5. Integration Testing

Test third-party integrations:

```go
Describe("GORM Integration", func() {
    It("should log GORM queries", func() {
        var buf bytes.Buffer
        
        gormLogger := gorm.New(log, gorm.Config{
            SlowThreshold: 100 * time.Millisecond,
        })
        
        // Simulate GORM query
        gormLogger.Info(context.Background(), "SELECT * FROM users")
        
        output := buf.String()
        Expect(output).To(ContainSubstring("SELECT"))
        Expect(output).To(ContainSubstring("users"))
    })
})
```

---

## Package-Specific Tests

### Logger Core Package

**Focus Areas**:
- Log method execution (Debug, Info, Warning, Error, etc.)
- Level filtering
- Field injection
- io.Writer interface
- Clone functionality
- Standard library integration

**Test Files**:
- `logger_suite_test.go`: Test suite setup
- `golog_test.go`: Standard library integration
- `interface_test.go`: Interface compliance
- `log_test.go`: Core logging methods
- `manage_test.go`: Configuration management
- `iowriter_test.go`: io.Writer interface
- `spf13_test.go`: spf13 integration

### Config Package

**Focus Areas**:
- Options structure validation
- Serialization (JSON, YAML, TOML)
- Default configuration
- File rotation settings
- Syslog configuration
- Format enumeration

**Test Files**:
- `config_test.go`: Configuration testing (127 specs)
- Validation testing
- Serialization testing
- Default values testing

### Entry Package

**Focus Areas**:
- Entry creation
- Field association
- Level setting
- Formatting
- Timestamp management

**Test Files**:
- `entry_test.go`: Entry testing (119 specs)
- Field merging
- Level association
- Format conversion

### Fields Package

**Focus Areas**:
- Field CRUD operations
- Merging
- Cloning
- Thread-safe access
- logrus.Fields conversion

**Test Files**:
- `fields_test.go`: Fields testing (49 specs)
- Add/Get/Del operations
- Merge functionality
- Clone behavior
- Concurrent access

---

## Writing Tests

### Test Structure

Follow Ginkgo BDD style:

```go
var _ = Describe("Logger", func() {
    var (
        log Logger
        ctx context.Context
    )
    
    BeforeEach(func() {
        ctx = context.Background()
        var err error
        log, err = New(ctx)
        Expect(err).ToNot(HaveOccurred())
    })
    
    AfterEach(func() {
        if log != nil {
            log.Close()
        }
    })
    
    Describe("Logging Methods", func() {
        Context("when logging info messages", func() {
            It("should log the message", func() {
                // Test implementation
                log.Info("Test message", nil)
                // Assertions
            })
        })
    })
})
```

### Test Naming

Use descriptive names:

```go
// Good
It("should write logs to file when file output is configured")
It("should filter messages below minimum level")
It("should merge persistent and per-entry fields")

// Bad
It("works")
It("test1")
It("check logging")
```

### Assertions

Use appropriate matchers:

```go
// Errors
Expect(err).ToNot(HaveOccurred())
Expect(err).To(HaveOccurred())
Expect(err).To(MatchError("expected error"))

// Values
Expect(level).To(Equal(level.InfoLevel))
Expect(message).To(ContainSubstring("expected"))
Expect(fields).To(HaveLen(3))
Expect(list).To(BeEmpty())

// Nil checks
Expect(value).ToNot(BeNil())
Expect(value).To(BeNil())

// Type assertions
Expect(value).To(BeAssignableToTypeOf(&Logger{}))
```

---

## Best Practices

### 1. Test Independence

Each test should be independent:

```go
// Good - Independent
It("test A", func() {
    log, _ := New(ctx)
    defer log.Close()
    // test logic
})

It("test B", func() {
    log, _ := New(ctx)  // Fresh instance
    defer log.Close()
    // test logic
})

// Bad - Shared state
var log Logger
It("test A", func() {
    log, _ = New(ctx)
    log.SetLevel(level.InfoLevel)
})
It("test B", func() {
    // Assumes state from test A
    Expect(log.GetLevel()).To(Equal(level.InfoLevel))
})
```

### 2. Cleanup

Always cleanup resources:

```go
AfterEach(func() {
    if log != nil {
        log.Close()
    }
    if tempFile != "" {
        os.Remove(tempFile)
    }
})

// Or use defer in test
It("should...", func() {
    log, _ := New(ctx)
    defer log.Close()
    // test logic
})
```

### 3. File System Tests

Handle temporary files properly:

```go
var tempFile string

BeforeEach(func() {
    f, _ := os.CreateTemp("", "logger-test-*.log")
    tempFile = f.Name()
    f.Close()
})

AfterEach(func() {
    if tempFile != "" {
        os.Remove(tempFile)
        // Also remove backup files from rotation
        os.Remove(tempFile + ".1")
        os.Remove(tempFile + ".2.gz")
    }
})
```

### 4. Capture Output

Test output correctly:

```go
It("should log to custom writer", func() {
    var buf bytes.Buffer
    
    // Configure logger with buffer
    hook := hookwriter.New(&buf)
    // Add hook to logger
    
    log.Info("Test message", nil)
    
    output := buf.String()
    Expect(output).To(ContainSubstring("Test message"))
})
```

### 5. Mock Integrations

Mock external dependencies:

```go
type mockWriter struct {
    messages []string
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
    m.messages = append(m.messages, string(p))
    return len(p), nil
}

It("should write to mock writer", func() {
    mock := &mockWriter{}
    // Configure logger with mock
    
    log.Info("Message", nil)
    
    Expect(mock.messages).To(HaveLen(1))
    Expect(mock.messages[0]).To(ContainSubstring("Message"))
})
```

---

## Troubleshooting

### Common Issues

**1. File Permission Errors**

```go
// Problem: Tests fail with permission denied
opts.LogFile = &config.OptionsFile{
    LogFileName: "/var/log/app.log",  // Not writable in tests
}

// Solution: Use temp directory
tempDir := os.TempDir()
opts.LogFile = &config.OptionsFile{
    LogFileName: filepath.Join(tempDir, "test.log"),
}
```

**2. Race Conditions**

```bash
# Run with race detector to find issues
CGO_ENABLED=1 go test -race ./...

# Common fixes:
# - Protect shared state with mutexes
# - Use proper synchronization
# - Avoid sharing logger instances across goroutines unsafely
```

**3. File Not Flushed**

```go
// Problem: Log file empty in test
log.Info("Message", nil)
content, _ := os.ReadFile(tempFile)
// content is empty!

// Solution: Close logger to flush
log.Info("Message", nil)
log.Close()  // Flushes buffers
content, _ := os.ReadFile(tempFile)
// content now has the message
```

**4. Level Not Working**

```go
// Problem: Debug messages still logged
log.SetLevel(level.InfoLevel)
log.Debug("Should not appear", nil)
// But it does!

// Check: Ensure level is set before logging
// Check: Verify io.Writer level separately
log.SetIOWriterLevel(level.InfoLevel)
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
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Check coverage
        run: |
          go tool cover -func=coverage.out | grep total
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### Makefile Example

```makefile
.PHONY: test test-race test-cover test-all

test:
	go test ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-all: test-race test-cover
	@echo "All tests passed!"

test-ci:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
```

---

## Summary

- **370+ test specs** covering core logging, configuration, entry management, and fields
- **Always run with `-race`** to detect concurrency issues
- **Cleanup resources** (files, loggers) in `AfterEach` or with `defer`
- **Test independence** - don't share state between tests
- **Use temp directories** for file-based tests
- **Flush buffers** by closing logger before asserting file contents
- **Mock external dependencies** for unit tests

For detailed API documentation, see [README.md](./README.md).
