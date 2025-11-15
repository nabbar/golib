# Testing Documentation - Retro Package

[![Tests](https://img.shields.io/badge/Tests-156%20passed-success)](https://github.com/nabbar/golib)
[![Coverage](https://img.shields.io/badge/Coverage-84.2%25-brightgreen)](https://github.com/nabbar/golib)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Comprehensive testing documentation for the retro package.

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
- [Quality Checklist](#quality-checklist)
- [Resources](#resources)
- [AI Transparency Notice](#ai-transparency-notice)

---

## Overview

The retro package has comprehensive test coverage across all components:

**Test Statistics**
- **Total Tests**: 156 specs
- **Overall Coverage**: 84.2%
- **Execution Time**: ~0.02s (without race detection)
- **Race Conditions**: 0 detected
- **Failed Tests**: 0

**Coverage by Component**
- Version logic: 100.0% (50 tests)
- Encoding: 85.4% (48 tests)
- Model operations: 74.4% (35 tests)
- Utilities: 100.0% (20 tests)
- Format validation: 100.0% (3 tests)

**Coverage Areas**
- ✅ Semantic version parsing and comparison
- ✅ Version constraint evaluation (all operators)
- ✅ JSON/YAML/TOML marshaling and unmarshaling
- ✅ Field filtering based on version
- ✅ Standard mode (bypass filtering)
- ✅ Omitempty behavior
- ✅ Custom unmarshalers
- ✅ Error handling and edge cases

---

## Quick Start

```bash
# Run all tests
cd /path/to/golib/retro
go test

# Run with coverage
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with race detector
CGO_ENABLED=1 go test -race

# Run with verbose output
go test -v

# Run specific test
go test -run TestModel_MarshalJSON
```

---

## Test Framework

### Ginkgo/Gomega v2

All tests use the Ginkgo BDD (Behavior-Driven Development) framework with Gomega matchers:

**Why Ginkgo?**
- Expressive BDD-style test organization
- Hierarchical test structure with `Describe` and `Context`
- Rich matcher library with Gomega
- Table-driven test support
- Detailed failure reporting

**Test Structure**
```go
var _ = Describe("Component", func() {
    var (
        // Test variables
    )
    
    BeforeEach(func() {
        // Setup before each test
    })
    
    Describe("Feature", func() {
        Context("when condition", func() {
            It("should behave correctly", func() {
                // Test assertions
                Expect(result).To(Equal(expected))
            })
        })
    })
})
```

### Test Suite

The package has a single test suite file:
- `suite_test.go` - Ginkgo test suite entry point

---

## Running Tests

### Basic Commands

```bash
# All tests
go test

# Verbose output
go test -v

# Show coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Using Ginkgo CLI

```bash
# Install Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run with Ginkgo
ginkgo

# Run with coverage
ginkgo -cover

# Focus on specific tests
ginkgo --focus="MarshalJSON"

# Skip tests
ginkgo --skip="YAML"

# Verbose output
ginkgo -v
```

### Race Detection

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Expected output
Ran 156 of 156 Specs in 0.037 seconds
SUCCESS! -- 156 Passed | 0 Failed | 0 Pending | 0 Skipped
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof
go tool pprof mem.prof

# Benchmarks
go test -bench=. -benchmem
```

---

## Test Coverage

### Coverage by Component

| Component | Coverage | Lines | Specs | Critical Paths |
|-----------|----------|-------|-------|----------------|
| **version.go** | 100.0% | 115/115 | 50 | Version comparison, constraints |
| **encoding.go** | 85.4% | 82/96 | 48 | Marshal/Unmarshal |
| **format.go** | 100.0% | 8/8 | 3 | Format validation |
| **utils.go** | 100.0% | 20/20 | 20 | Empty value detection |
| **model.go** | 74.4% | 95/128 | 35 | Field filtering |
| **Total** | **84.2%** | **320/380** | **156** | All components |

### Coverage Details

#### Version Logic (100.0%)

**Covered**:
- ✅ isVersionSupported - All constraint types
- ✅ parseOperator - All operators (>=, <=, >, <)
- ✅ isValidVersion - Valid and invalid formats
- ✅ checkCondition - All comparison operators
- ✅ compareVersions - Major, minor, patch comparison
- ✅ validRetroTag - Tag validation rules
- ✅ detectedBoundaries - Dual boundary detection

**Test Cases**:
- Single version constraints
- Range constraints (dual boundaries)
- Multiple version matching
- Invalid version formats
- Edge cases (default version, empty tags)

#### Encoding (85.4%)

**Covered**:
- ✅ MarshalJSON - 100%
- ✅ UnmarshalJSON - 100%
- ✅ MarshalTOML - 100%
- ⚠️ MarshalYAML - 71.4%
- ⚠️ UnmarshalYAML - 66.7%
- ⚠️ UnmarshalTOML - 75.0%

**Partially Covered**:
- ⚠️ YAML error paths (marshaling intermediate data)
- ⚠️ TOML type assertion failures
- ⚠️ Custom unmarshaler error handling

#### Model Operations (74.4%)

**Covered**:
- ✅ marshal - Field filtering, omitempty, format switching
- ⚠️ unmarshal - Basic unmarshaling, version extraction

**Partially Covered**:
- ⚠️ unmarshal (60.6%) - Complex error scenarios
  - Custom unmarshaler failures
  - Invalid field types
  - Non-addressable fields
  - Marshaling errors in unmarshal flow

#### Utilities (100.0%)

**Covered**:
- ✅ isEmptyValue - All Go types
  - Strings, arrays, slices, maps
  - Booleans
  - Integer types (int, int8, int16, int32, int64)
  - Unsigned integers (uint, uint8, uint16, uint32, uint64)
  - Floating point (float32, float64)
  - Interfaces and pointers

### Viewing Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out

# View in browser
go tool cover -html=coverage.out

# View in terminal
go tool cover -func=coverage.out

# Coverage by function
go tool cover -func=coverage.out | grep -E "(version|encoding|model)"
```

---

## Thread Safety

### Concurrency Testing

All operations are tested for thread safety using Go's race detector:

```bash
CGO_ENABLED=1 go test -race
```

**Race Detection Results**: 0 race conditions detected

### Thread-Safe Components

**Model Operations**
- ✅ marshal - Stateless, read-only reflection
- ✅ unmarshal - Modifies target struct only
- ✅ No shared mutable state

**Version Comparison**
- ✅ Pure functions, no side effects
- ✅ Regex compilation cached globally (read-only)
- ✅ No synchronization needed

**Encoding**
- ✅ Delegates to standard libraries (thread-safe)
- ✅ No package-level mutable state
- ✅ Each operation independent

### Concurrency Primitives

The package uses:
- **Stateless Operations**: All functions are pure or have isolated state
- **Read-Only Globals**: Only `versionRegex` and `SupportedFormats` (immutable)
- **No Synchronization**: Not needed due to stateless design

---

## Test File Organization

### File Structure

```
retro/
├── suite_test.go           # Test suite entry point (39 lines)
├── format_test.go          # Format validation (96 lines, ~15 specs)
├── utils_test.go           # Utility functions (256 lines, ~20 specs)
├── version_test.go         # Version logic (468 lines, ~50 specs)
├── model_test.go           # Model operations (530 lines, ~35 specs)
└── retro_test.go           # Integration tests (578 lines, ~36 specs)
```

### Test Organization

| File | Purpose | Specs | Lines |
|------|---------|-------|-------|
| `suite_test.go` | Ginkgo suite setup | 1 | 39 |
| `format_test.go` | Format validation | ~15 | 96 |
| `utils_test.go` | Empty value detection | ~20 | 256 |
| `version_test.go` | Version comparison | ~50 | 468 |
| `model_test.go` | Marshal/Unmarshal | ~35 | 530 |
| `retro_test.go` | Integration tests | ~36 | 578 |
| **Total** | | **156** | **1,967** |

---

## Writing Tests

### Test Template

```go
package retro_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("Component", func() {
    var (
        component ComponentType
    )
    
    BeforeEach(func() {
        // Setup
        component = NewComponent()
    })
    
    Describe("Method", func() {
        Context("when valid input", func() {
            It("should return expected result", func() {
                result := component.Method(input)
                Expect(result).To(Equal(expected))
            })
        })
        
        Context("when invalid input", func() {
            It("should return error", func() {
                result := component.Method(invalidInput)
                Expect(result).To(BeNil())
            })
        })
    })
})
```

### Guidelines

1. **Use Descriptive Names**
   ```go
   // Good
   It("should include field when version matches constraint", func() {})
   
   // Bad
   It("should work", func() {})
   ```

2. **Test One Thing Per Spec**
   ```go
   // Good
   It("should parse >= operator correctly", func() {
       op, ver := parseOperator(">=v1.0.0")
       Expect(op).To(Equal(">="))
   })
   
   // Bad
   It("should parse all operators", func() {
       // Tests multiple operators in one spec
   })
   ```

3. **Use Appropriate Matchers**
   ```go
   Expect(value).To(Equal(expected))        // Exact match
   Expect(value).To(BeTrue())               // Boolean
   Expect(value).To(HaveLen(5))             // Length check
   Expect(func()).ToNot(Panic())            // Panic check
   Expect(err).To(HaveOccurred())           // Error check
   ```

4. **Test Edge Cases**
   ```go
   Context("when version is empty", func() {})
   Context("when version is 'default'", func() {})
   Context("when constraint has invalid operator", func() {})
   ```

5. **Use Table-Driven Tests**
   ```go
   DescribeTable("version comparison",
       func(v1, v2 string, expected int) {
           result := compareVersions(v1, v2)
           Expect(result).To(Equal(expected))
       },
       Entry("v1 < v2", "1.0.0", "2.0.0", -1),
       Entry("v1 == v2", "1.0.0", "1.0.0", 0),
       Entry("v1 > v2", "2.0.0", "1.0.0", 1),
   )
   ```

---

## Best Practices

### Test Independence

**✅ DO**: Isolate test state
```go
BeforeEach(func() {
    model = retro.Model[TestStruct]{}  // Fresh instance per test
})
```

**❌ DON'T**: Share state between tests
```go
var model = retro.Model[TestStruct]{}  // Shared across all tests
```

### Test Data

**✅ DO**: Use realistic test data
```go
user := User{
    Version: "v1.5.0",
    Name:    "Alice",
    Email:   "alice@example.com",
}
```

**❌ DON'T**: Use placeholder data
```go
user := User{
    Version: "v1",
    Name:    "test",
}
```

### Assertions

**✅ DO**: Use specific matchers
```go
Expect(version).To(Equal("v1.0.0"))
```

**❌ DON'T**: Use generic matchers
```go
Expect(version).ToNot(BeEmpty())
```

### Version Testing

**✅ DO**: Test all constraint types
```go
Context("with >= constraint", func() {})
Context("with < constraint", func() {})
Context("with range constraint", func() {})
```

**❌ DON'T**: Only test happy paths
```go
It("should work", func() {
    Expect(isVersionSupported("v1.0.0", ">=v1.0.0")).To(BeTrue())
})
```

### Error Handling

**✅ DO**: Test error paths
```go
Context("when version format is invalid", func() {
    It("should return false", func() {
        result := isValidVersion("invalid")
        Expect(result).To(BeFalse())
    })
})
```

**❌ DON'T**: Only test success cases
```go
It("should validate version", func() {
    Expect(isValidVersion("v1.0.0")).To(BeTrue())
})
```

---

## Troubleshooting

### Common Issues

**1. Stale Coverage Data**
```bash
# Problem: Old coverage data
# Solution: Remove old files
rm -f coverage.out
go test -coverprofile=coverage.out
```

**2. Test Failures**
```bash
# Problem: Tests fail unexpectedly
# Solution: Run with verbose output
go test -v

# Check specific test
go test -run TestModel_MarshalJSON -v
```

**3. Race Conditions**
```bash
# Problem: Race detector reports issues
# Solution: Fix shared state access
CGO_ENABLED=1 go test -race -v
```

**4. CGO Not Enabled**
```bash
# Problem: Race detector requires CGO
# Solution: Enable CGO
export CGO_ENABLED=1
go test -race
```

**5. Test Timeouts**
```bash
# Problem: Tests hang or timeout
# Solution: Increase timeout
go test -timeout 30s
```

### Debugging Tests

```bash
# Run specific test
ginkgo --focus="should parse operator" -v

# Skip tests
ginkgo --skip="YAML" -v

# Print test names
ginkgo --dry-run

# Debug with delve
dlv test github.com/nabbar/golib/retro
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
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          cd retro
          go test -v -coverprofile=coverage.out
      
      - name: Race detection
        run: |
          cd retro
          CGO_ENABLED=1 go test -race
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./retro/coverage.out
```

### GitLab CI

```yaml
test:
  image: golang:1.21
  script:
    - cd retro
    - go test -v -coverprofile=coverage.out
    - CGO_ENABLED=1 go test -race
  coverage: '/total:.*?(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: retro/coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

cd retro
echo "Running tests..."
go test

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Running race detector..."
CGO_ENABLED=1 go test -race

if [ $? -ne 0 ]; then
    echo "Race conditions detected. Commit aborted."
    exit 1
fi

echo "All tests passed!"
```

---

## Quality Checklist

Before submitting code, ensure:

- [ ] All tests pass: `go test`
- [ ] Coverage maintained: `go test -cover` (≥80%)
- [ ] No race conditions: `CGO_ENABLED=1 go test -race`
- [ ] Code formatted: `go fmt ./...`
- [ ] Linter clean: `go vet ./...`
- [ ] Documentation updated: GoDoc comments
- [ ] Examples provided: For new features
- [ ] Edge cases tested: Nil, empty, invalid inputs
- [ ] Error paths tested: All error returns
- [ ] Version constraints tested: All operators

---

## Resources

**Testing Frameworks**
- [Ginkgo](https://onsi.github.io/ginkgo/) - BDD testing framework
- [Gomega](https://onsi.github.io/gomega/) - Matcher library
- [Go Testing Package](https://pkg.go.dev/testing)

**Go Testing**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Coverage](https://go.dev/blog/cover)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

**Best Practices**
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.
