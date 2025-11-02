# Console Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the console package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Structure](#test-structure)
- [Test Coverage](#test-coverage)
- [Test Categories](#test-categories)
- [UTF-8 Testing](#utf-8-testing)
- [Race Detection](#race-detection)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The console package features a comprehensive test suite with **182 specifications** covering colored output, text formatting, user prompts, UTF-8 support, and more. All tests follow BDD (Behavior-Driven Development) principles for maximum readability and maintainability.

### Test Metrics

| Metric | Value | Status |
|--------|-------|--------|
| **Total Specs** | 182 | ✅ All passing |
| **Test Files** | 7 | ✅ Organized by feature |
| **Code Coverage** | 60.9% | ⚠️ Prompts require interactive input |
| **Execution Time** | ~0.2s | ✅ Fast |
| **Framework** | Ginkgo v2 + Gomega | ✅ BDD style |
| **UTF-8 Coverage** | 100% | ✅ Full multi-byte support |

---

## Test Framework

### Ginkgo v2

Ginkgo is a BDD-style testing framework for Go, providing:

- **Structured Organization**: `Describe`, `Context`, `It` blocks for hierarchical test structure
- **Setup/Teardown**: `BeforeEach`, `AfterEach` hooks
- **Focus/Skip**: `FDescribe`, `FIt`, `XDescribe`, `XIt` for debugging
- **Parallel Execution**: Built-in support for concurrent test execution
- **Rich Reporting**: Detailed failure messages and stack traces
- **Table-Driven Tests**: Native support for data-driven testing

### Gomega

Gomega is an assertion/matcher library providing:

- **Expressive Matchers**: `Expect(x).To(Equal(y))`, `Expect(err).NotTo(HaveOccurred())`
- **Type-Safe**: Strong typing for assertions
- **Custom Matchers**: Domain-specific assertions
- **Clear Failure Messages**: Detailed error output with actual vs expected values

### Installation

```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Test Organization

### Test Files Structure

```
console/
├── console_suite_test.go      # Test suite entry point
├── color_test.go              # Color management tests (29 specs)
├── formatting_test.go         # Print formatting tests (40 specs)
├── buffer_test.go             # Buffer operations tests (38 specs)
├── padding_test.go            # Text padding tests (47 specs)
├── error_test.go              # Error handling tests (14 specs)
└── integration_test.go        # Integration tests (14 specs)
```

---

## Running Tests

### Quick Test

Run all tests with standard output:

```bash
cd /sources/go/src/github.com/nabbar/golib/console
go test -v
```

### With Coverage

Generate and view coverage report:

```bash
# Run tests with coverage
go test -v -cover

# Generate HTML coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Coverage by function
go tool cover -func=coverage.out
```

### With Race Detection

Enable race detector:

```bash
CGO_ENABLED=1 go test -race -v
```

### Using Ginkgo CLI

Run tests with enhanced Ginkgo output:

```bash
# Install Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run tests with verbose output
ginkgo -v

# Run with coverage
ginkgo -v -cover

# Run with trace for debugging
ginkgo -v --trace

# Run specific test file
ginkgo -v --focus-file padding_test.go
```

### Parallel Execution

Run tests in parallel:

```bash
ginkgo -v -p

# Custom parallelism
ginkgo -v -procs=8

# Parallel with race detection
CGO_ENABLED=1 ginkgo -v -p -race
```

### Focus and Skip

Debug specific tests:

```bash
# Focus on specific file
ginkgo -v --focus-file padding_test.go

# Focus on specific spec
ginkgo -v --focus "UTF-8"

# Skip specific specs
ginkgo -v --skip "Integration"

# Fail fast (stop on first failure)
ginkgo -v --fail-fast
```

---

## Test Structure

Tests are organized by functionality for maximum maintainability and readability:

### 1. **console_suite_test.go** (Suite Setup)

- Ginkgo test suite configuration
- Test runner initialization
- Package-level test setup

### 2. **color_test.go** (29 specs)

**Scenarios Covered:**
- Color type constants and conversions
- Color setting with single and multiple attributes
- Color retrieval and deletion
- Thread-safe color operations
- Color persistence across operations

**Example:**
```go
Describe("Color Management", func() {
    It("should set and retrieve colors", func() {
        console.SetColor(console.ColorPrint, int(color.FgRed))
        c := console.GetColor(console.ColorPrint)
        Expect(c).NotTo(BeNil())
    })
})
```

### 3. **formatting_test.go** (40 specs)

**Scenarios Covered:**
- Print operations (Print, Println, Printf, PrintLnf)
- Sprintf string formatting
- Color formatting with attributes
- Edge cases (empty strings, special characters)
- Formatting preservation

**Example:**
```go
Describe("Print Formatting", func() {
    It("should format with Printf", func() {
        console.ColorPrint.Printf("Hello %s", "World")
        // Output verified through buffer capture
    })
})
```

### 4. **buffer_test.go** (38 specs)

**Scenarios Covered:**
- Buffer creation and writing
- BuffPrintf operations
- Error handling (nil buffers, write failures)
- Multiple buffer support
- Thread-safe buffer operations

**Example:**
```go
Describe("Buffer Operations", func() {
    It("should write to buffer", func() {
        var buf bytes.Buffer
        n, err := console.ColorPrint.BuffPrintf(&buf, "test")
        Expect(err).NotTo(HaveOccurred())
        Expect(n).To(BeNumerically(">", 0))
    })
})
```

### 5. **padding_test.go** (47 specs)

**Scenarios Covered:**
- PadLeft (right-align text)
- PadRight (left-align text)
- PadCenter (center-align text)
- UTF-8 multi-byte character handling
- Emoji support
- CJK character support (Chinese, Japanese, Korean)
- Arabic/RTL text support
- Hierarchical tabbed output (PrintTabf)
- Edge cases (empty strings, zero length, negative length)

**Example:**
```go
Describe("Text Padding", func() {
    It("should pad left correctly", func() {
        result := console.PadLeft("test", 10, " ")
        Expect(result).To(Equal("      test"))
        Expect(len([]rune(result))).To(Equal(10))
    })
    
    It("should handle UTF-8 correctly", func() {
        result := console.PadCenter("你好", 10, " ")
        // Correctly counts visual width, not byte length
        Expect(len([]rune(result))).To(Equal(10))
    })
})
```

### 6. **error_test.go** (14 specs)

**Scenarios Covered:**
- Error type definitions
- Error formatting
- Buffer error handling
- Error messages and codes

**Example:**
```go
Describe("Error Handling", func() {
    It("should return error for nil buffer", func() {
        _, err := console.ColorPrint.BuffPrintf(nil, "test")
        Expect(err).To(HaveOccurred())
    })
})
```

### 7. **integration_test.go** (14 specs)

**Scenarios Covered:**
- Complete workflows combining multiple features
- Table generation with padding and colors
- Progress indicators
- Multi-language output
- Configuration display
- Real-world use cases

**Example:**
```go
Describe("Integration Tests", func() {
    It("should create formatted table", func() {
        // Combines padding, colors, and formatting
        header := console.PadRight("Name", 20, " ")
        console.ColorPrint.Println(header)
        // Complete table generation...
    })
})
```

---

## Test Coverage

### Coverage by Component

| Component | File | Specs | Coverage |
|-----------|------|-------|----------|
| Color Management | color_test.go | 29 | ~95% |
| Formatting | formatting_test.go | 40 | ~90% |
| Buffer Operations | buffer_test.go | 38 | ~85% |
| Text Padding | padding_test.go | 47 | ~100% |
| Error Handling | error_test.go | 14 | ~80% |
| Integration | integration_test.go | 14 | ~70% |

**Overall Coverage**: 60.9%

### Coverage Notes

- **Prompt functions** (`PromptString`, `PromptInt`, etc.) require interactive terminal input and are difficult to test automatically
- **Terminal-specific features** may behave differently in CI/CD environments without TTY
- **Color output** depends on terminal capabilities (ANSI support)

---

## UTF-8 Testing

The console package has extensive UTF-8 test coverage:

### Character Sets Tested

1. **ASCII**: Basic Latin characters
2. **Latin Extended**: Accented characters (é, ñ, ü)
3. **CJK**: Chinese, Japanese, Korean characters
4. **Arabic**: Right-to-left script
5. **Emoji**: Single and multi-codepoint emojis
6. **Mathematical**: Special symbols (∑, ∫, √)

### UTF-8 Edge Cases

```go
// Multi-byte character tests
"你好世界"     // Chinese (4 characters, 12 bytes)
"こんにちは"   // Japanese (5 characters, 15 bytes)
"مرحبا"       // Arabic (5 characters, 10 bytes)
"Hello 🌍"    // Mixed ASCII + emoji

// Wide character handling
"你" // 1 visual cell, 3 bytes
"🌍" // 1 visual cell, 4 bytes (can be 2 cells in some terminals)
```

### Padding Accuracy

Tests verify that padding correctly counts **visual width**, not byte length:

```go
// Should all produce 10-character wide output
PadCenter("hello", 10, " ")     // ASCII
PadCenter("你好", 10, " ")       // CJK (2 visual cells each)
PadCenter("🌍", 10, " ")          // Emoji
```

---

## Race Detection

Run tests with race detector to verify thread-safety:

```bash
CGO_ENABLED=1 go test -race -v
```

### Thread-Safe Components

- **Color Storage**: Uses `libatm.NewMapTyped` for atomic map operations
- **Concurrent Reads**: Multiple goroutines can read colors simultaneously
- **Concurrent Writes**: Color updates are atomic

---

## Best Practices

### 1. Use Descriptive Test Names

```go
It("should pad left with spaces correctly", func() {
    // Test implementation
})
```

### 2. Test Both Success and Failure Cases

```go
Context("when buffer is nil", func() {
    It("should return error", func() { /* ... */ })
})

Context("when buffer is valid", func() {
    It("should write successfully", func() { /* ... */ })
})
```

### 3. Clean Up Resources

```go
BeforeEach(func() {
    console.SetColor(console.ColorPrint, int(color.FgWhite))
})

AfterEach(func() {
    console.DelColor(console.ColorPrint)
})
```

### 4. Test Edge Cases

- Empty strings
- Very long strings
- Nil pointers
- Zero/negative lengths
- Special characters
- Multi-byte UTF-8

### 5. Verify Output Format

```go
It("should format correctly", func() {
    var buf bytes.Buffer
    console.ColorPrint.BuffPrintf(&buf, "test")
    
    output := buf.String()
    Expect(output).To(ContainSubstring("test"))
})
```

---

## Troubleshooting

### Verbose Output

Enable detailed test output:

```bash
go test -v
ginkgo -v --trace
```

### Focus on Specific Test

Isolate failing tests:

```bash
ginkgo -focus "should pad left"
```

### Skip Tests

Temporarily skip tests:

```bash
ginkgo -skip "integration"
```

### Check for Race Conditions

Always run with race detector during development:

```bash
CGO_ENABLED=1 go test -race
```

### Debug a Failing Test

Steps to debug:

1. Run with increased verbosity: `ginkgo -v -vv`
2. Add trace for stack traces: `ginkgo -v --trace`
3. Run single test file: `ginkgo -v --focus-file padding_test.go`
4. Add temporary `fmt.Printf` or use debugger

---

## CI/CD Integration

### GitHub Actions

```yaml
test-console:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Test console package
      run: |
        cd console
        go test -v -race -cover
```

### GitLab CI

```yaml
test-console:
  script:
    - cd console
    - go test -v -race -cover
  coverage: '/coverage: \d+\.\d+% of statements/'
```

---

## Contributing

When adding new features to the console package:

1. **Write tests first** (TDD approach)
2. **Cover edge cases** (nil, empty, UTF-8)
3. **Test with race detector**
4. **Verify UTF-8 handling** for text operations
5. **Update coverage** metrics in this document
6. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    BeforeEach(func() {
        // Setup
    })

    Describe("Feature behavior", func() {
        It("should handle basic case", func() {
            // Test implementation
            Expect(result).To(Equal(expected))
        })

        It("should handle UTF-8", func() {
            // Test with multi-byte characters
            Expect(result).NotTo(BeNil())
        })

        Context("when error occurs", func() {
            It("should return error", func() {
                // Test error handling
                Expect(err).To(HaveOccurred())
            })
        })
    })
})
```

---

## Support

For issues or questions about tests:

- **Test Failures**: Check test output for detailed error messages and stack traces
- **Usage Examples**: Review specific test files for implementation patterns
- **Feature Questions**: Consult [README.md](README.md) for feature documentation
- **Bug Reports**: Open an issue on [GitHub](https://github.com/nabbar/golib/issues) with:
  - Test output (with `-v` flag)
  - Go version and OS
  - Steps to reproduce
  - Expected vs actual behavior

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
