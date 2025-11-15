# Testing Documentation - Router Package

[![Tests](https://img.shields.io/badge/Tests-113%20passed-success)](https://github.com/nabbar/golib)
[![Coverage](https://img.shields.io/badge/Coverage-91.4%25-brightgreen)](https://github.com/nabbar/golib)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Comprehensive testing documentation for the router package and its subpackages.

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

The router package has comprehensive test coverage across all components:

**Test Statistics**
- **Total Tests**: 113 specs
- **Overall Coverage**: 91.4%
- **Execution Time**: ~0.08s (without race detection)
- **Race Conditions**: 0 detected
- **Failed Tests**: 0

**Coverage by Package**
- `router`: 92.1% (61 tests)
- `router/auth`: 96.3% (12 tests)
- `router/authheader`: 100.0% (11 tests)
- `router/header`: 83.3% (29 tests)

**Coverage Areas**
- ✅ Route registration and management
- ✅ Middleware chain execution
- ✅ Authorization flows (success, failure, forbidden)
- ✅ Header manipulation and middleware
- ✅ Error handling and recovery
- ✅ Logging integration
- ✅ Context management
- ✅ Edge cases and error conditions

---

## Quick Start

```bash
# Run all tests
cd /path/to/golib/router
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector
CGO_ENABLED=1 go test -race ./...

# Run specific package
go test github.com/nabbar/golib/router/auth -v

# Run with verbose output
go test -v ./...
```

---

## Test Framework

### Ginkgo/Gomega v2

All tests use the Ginkgo BDD (Behavior-Driven Development) framework with Gomega matchers:

**Why Ginkgo?**
- Expressive BDD-style test organization
- Hierarchical test structure with `Describe` and `Context`
- Rich matcher library with Gomega
- Parallel test execution support
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

### Test Suites

Each package has a test suite file:
- `router_suite_test.go` - Main router suite
- `auth_suite_test.go` - Auth package suite
- `authheader_suite_test.go` - AuthHeader package suite
- `header_suite_test.go` - Header package suite

---

## Running Tests

### Basic Commands

```bash
# All packages
go test ./...

# Single package
go test github.com/nabbar/golib/router

# Verbose output
go test -v ./...

# Show coverage
go test -cover ./...
```

### Using Ginkgo CLI

```bash
# Install Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run with Ginkgo
ginkgo -r

# Run with coverage
ginkgo -r -cover

# Run specific suite
ginkgo router/auth

# Parallel execution
ginkgo -r -p

# Focus on specific tests
ginkgo -r --focus="Authorization"
```

### Race Detection

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -r -race

# Expected output
ok  	github.com/nabbar/golib/router	1.064s
ok  	github.com/nabbar/golib/router/auth	1.036s
ok  	github.com/nabbar/golib/router/authheader	1.034s
ok  	github.com/nabbar/golib/router/header	1.043s
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Benchmarks
go test -bench=. -benchmem ./...
```

---

## Test Coverage

### Coverage by Package

| Package | Coverage | Lines | Specs | Critical Paths |
|---------|----------|-------|-------|----------------|
| **router** | 92.1% | 152/165 | 61 | RouterList, Middleware |
| **router/auth** | 96.3% | 52/54 | 12 | Authorization flow |
| **router/authheader** | 100.0% | 20/20 | 11 | Auth helpers |
| **router/header** | 83.3% | 75/90 | 29 | Header operations |
| **Total** | **91.4%** | **299/329** | **113** | All components |

### Coverage Details

#### Router Core (92.1%)

**Covered**:
- ✅ RouterList creation and initialization
- ✅ Route registration (Register, RegisterInGroup, RegisterMergeInGroup)
- ✅ Handler application to Gin engine
- ✅ Engine creation with custom initializers
- ✅ Middleware execution (Latency, Request, Access, Error)
- ✅ Error recovery and panic handling
- ✅ Default configurations
- ✅ Error code messages

**Partially Covered**:
- ⚠️ GinRequestContext nil checks (92.3%)
- ⚠️ GinErrorLog broken pipe detection (81.2%)

#### Auth (96.3%)

**Covered**:
- ✅ Authorization creation
- ✅ Handler registration and appending
- ✅ Auth header parsing (Bearer, Basic)
- ✅ Auth check function execution
- ✅ Success flow (AuthCodeSuccess)
- ✅ Failure flows (AuthCodeRequire, AuthCodeForbidden)
- ✅ Unknown auth code handling
- ✅ Case-insensitive auth type matching

**Partially Covered**:
- ⚠️ Debug logging (only tested when logger present)

#### AuthHeader (100.0%)

**Covered**:
- ✅ AuthCode constants
- ✅ Header constants
- ✅ AuthRequire function (401 response)
- ✅ AuthForbidden function (403 response)
- ✅ Error attachment to context
- ✅ Handler chain abortion
- ✅ WWW-Authenticate header setting

#### Header (83.3%)

**Covered**:
- ✅ Headers creation
- ✅ Add/Set/Get/Del operations
- ✅ Header map export
- ✅ Clone functionality
- ✅ Handler middleware
- ✅ Register handler chain
- ✅ Configuration (HeadersConfig)
- ✅ Case-insensitive operations

**Partially Covered**:
- ⚠️ Nil header initialization in Get (66.7%)
- ⚠️ Multi-value header handling (75.0%)

### Viewing Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# View in terminal
go tool cover -func=coverage.out

# Coverage by function
go tool cover -func=coverage.out | grep -E "(router|auth|header)"
```

---

## Thread Safety

### Concurrency Testing

All packages are tested for thread safety using Go's race detector:

```bash
CGO_ENABLED=1 go test -race ./...
```

**Race Detection Results**: 0 race conditions detected

### Thread-Safe Components

**Router**
- ✅ Route registration (immutable after setup)
- ✅ Middleware execution (context-isolated)
- ✅ Engine creation (independent instances)

**Auth**
- ✅ Authorization checks (stateless)
- ✅ Handler execution (context-based)

**Header**
- ✅ Header operations (per-instance)
- ✅ Middleware application (context-isolated)

### Concurrency Primitives

The package uses standard Go concurrency patterns:
- **Context Isolation**: Each request has its own Gin context
- **Immutable Routes**: Routes registered once, never modified
- **Stateless Operations**: No shared mutable state during request handling

---

## Test File Organization

### File Structure

```
router/
├── router_suite_test.go         # Test suite entry point
├── router_test.go               # RouterList tests (32 specs)
├── middleware_test.go           # Middleware tests (13 specs)
├── error_test.go                # Error code tests (8 specs)
├── default_test.go              # Default config tests (8 specs)
├── auth/
│   ├── auth_suite_test.go       # Auth suite entry point
│   └── auth_test.go             # Authorization tests (12 specs)
├── authheader/
│   ├── authheader_suite_test.go # AuthHeader suite entry point
│   └── authheader_test.go       # Auth helper tests (11 specs)
└── header/
    ├── header_suite_test.go     # Header suite entry point
    ├── header_test.go           # Header tests (26 specs)
    └── config_test.go           # Config tests (3 specs)
```

### Test Organization

| File | Purpose | Specs | Lines |
|------|---------|-------|-------|
| `router_test.go` | RouterList operations | 32 | 380 |
| `middleware_test.go` | Middleware functionality | 13 | 350 |
| `error_test.go` | Error codes | 8 | 90 |
| `default_test.go` | Default configurations | 8 | 270 |
| `auth_test.go` | Authorization flow | 12 | 360 |
| `authheader_test.go` | Auth helpers | 11 | 200 |
| `header_test.go` | Header operations | 26 | 380 |
| `config_test.go` | Header configuration | 3 | 60 |
| **Total** | | **113** | **2,090** |

---

## Writing Tests

### Test Template

```go
package router_test

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
   It("should return 401 when authorization header is missing", func() {})
   
   // Bad
   It("should work", func() {})
   ```

2. **Test One Thing Per Spec**
   ```go
   // Good
   It("should set request path in context", func() {
       Expect(path).To(Equal("/test"))
   })
   
   // Bad
   It("should set path and user and query", func() {
       Expect(path).To(Equal("/test"))
       Expect(user).To(Equal("alice"))
       Expect(query).To(Equal("foo=bar"))
   })
   ```

3. **Use Appropriate Matchers**
   ```go
   Expect(value).To(Equal(expected))        // Exact match
   Expect(value).To(BeNumerically(">", 0))  // Numeric comparison
   Expect(value).To(ContainSubstring("x"))  // String contains
   Expect(value).To(HaveLen(5))             // Length check
   Expect(func()).ToNot(Panic())            // Panic check
   ```

4. **Test Edge Cases**
   ```go
   Context("when input is empty", func() {})
   Context("when input is nil", func() {})
   Context("when input is very large", func() {})
   ```

5. **Clean Up Resources**
   ```go
   AfterEach(func() {
       if server != nil {
           server.Close()
       }
   })
   ```

6. **Use Test Helpers**
   ```go
   func createTestEngine() *gin.Engine {
       gin.SetMode(gin.TestMode)
       return gin.New()
   }
   ```

---

## Best Practices

### Test Independence

**✅ DO**: Isolate test state
```go
BeforeEach(func() {
    engine = gin.New()  // Fresh instance per test
})
```

**❌ DON'T**: Share state between tests
```go
var engine = gin.New()  // Shared across all tests
```

### Test Data

**✅ DO**: Use realistic test data
```go
token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**❌ DON'T**: Use placeholder data
```go
token := "test"
```

### Assertions

**✅ DO**: Use specific matchers
```go
Expect(statusCode).To(Equal(http.StatusOK))
```

**❌ DON'T**: Use generic matchers
```go
Expect(statusCode).To(BeNumerically(">", 0))
```

### Concurrency Testing

**✅ DO**: Test concurrent access
```go
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        routerList.Register("GET", "/test", handler)
    }()
}
wg.Wait()
```

**❌ DON'T**: Assume thread safety
```go
routerList.Register("GET", "/test", handler)  // Only sequential test
```

### Error Handling

**✅ DO**: Test error paths
```go
Context("when authorization fails", func() {
    It("should return 401", func() {
        Expect(statusCode).To(Equal(401))
    })
})
```

**❌ DON'T**: Only test happy paths
```go
It("should work", func() {
    Expect(statusCode).To(Equal(200))
})
```

### File Operations

**✅ DO**: Use httptest for HTTP testing
```go
w := httptest.NewRecorder()
req, _ := http.NewRequest("GET", "/test", nil)
engine.ServeHTTP(w, req)
```

**❌ DON'T**: Start real HTTP servers in tests
```go
go engine.Run(":8080")  // Bad: real server
```

---

## Troubleshooting

### Common Issues

**1. Stale Coverage Data**
```bash
# Problem: Old coverage data
# Solution: Remove old files
rm -f coverage.out
go test -coverprofile=coverage.out ./...
```

**2. Parallel Test Failures**
```bash
# Problem: Tests fail when run in parallel
# Solution: Ensure test independence
ginkgo -r -p  # Run in parallel to detect issues
```

**3. Race Conditions**
```bash
# Problem: Race detector reports issues
# Solution: Fix shared state access
CGO_ENABLED=1 go test -race ./...
```

**4. CGO Not Enabled**
```bash
# Problem: Race detector requires CGO
# Solution: Enable CGO
export CGO_ENABLED=1
go test -race ./...
```

**5. Test Timeouts**
```bash
# Problem: Tests hang or timeout
# Solution: Increase timeout
go test -timeout 30s ./...
```

### Debugging Tests

```bash
# Run specific test
ginkgo --focus="should return 401" ./auth

# Skip tests
ginkgo --skip="slow tests" ./...

# Verbose output
go test -v ./...

# Print test names
ginkgo -r --dry-run

# Debug with delve
dlv test github.com/nabbar/golib/router
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
          cd router
          go test -v -coverprofile=coverage.out ./...
      
      - name: Race detection
        run: |
          cd router
          CGO_ENABLED=1 go test -race ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./router/coverage.out
```

### GitLab CI

```yaml
test:
  image: golang:1.21
  script:
    - cd router
    - go test -v -coverprofile=coverage.out ./...
    - CGO_ENABLED=1 go test -race ./...
  coverage: '/total:.*?(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: router/coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

cd router
echo "Running tests..."
go test ./...

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Running race detector..."
CGO_ENABLED=1 go test -race ./...

if [ $? -ne 0 ]; then
    echo "Race conditions detected. Commit aborted."
    exit 1
fi

echo "All tests passed!"
```

---

## Quality Checklist

Before submitting code, ensure:

- [ ] All tests pass: `go test ./...`
- [ ] Coverage maintained: `go test -cover ./...` (≥90%)
- [ ] No race conditions: `CGO_ENABLED=1 go test -race ./...`
- [ ] Code formatted: `go fmt ./...`
- [ ] Linter clean: `go vet ./...`
- [ ] Documentation updated: GoDoc comments
- [ ] Examples provided: For new features
- [ ] Edge cases tested: Nil, empty, invalid inputs
- [ ] Error paths tested: All error returns
- [ ] Integration tests: Component interactions

---

## Resources

**Testing Frameworks**
- [Ginkgo](https://onsi.github.io/ginkgo/) - BDD testing framework
- [Gomega](https://onsi.github.io/gomega/) - Matcher library
- [httptest](https://pkg.go.dev/net/http/httptest) - HTTP testing utilities

**Go Testing**
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Coverage](https://go.dev/blog/cover)

**Best Practices**
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.
