# Context Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the context package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Concurrency Testing](#concurrency-testing)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The context package features **60+ test specifications** covering thread-safe operations, context management, cloning, merging, and concurrent access patterns.

### Test Metrics

| Metric | Value | Status |
|--------|-------|--------|
| **Total Specs** | 60+ | ✅ All passing |
| **Test Files** | 4 | ✅ Organized by feature |
| **Code Coverage** | >90% | ✅ High coverage |
| **Execution Time** | ~0.3s | ✅ Fast |
| **Framework** | Ginkgo v2 + Gomega | ✅ BDD style |
| **Race Detection** | Enabled | ✅ No races found |

---

## Test Framework

### Ginkgo v2 + Gomega

**Features:**
- Structured organization with `Describe`, `Context`, `It` blocks
- Setup/teardown with `BeforeEach`, `AfterEach`
- Parallel test execution support
- Rich failure reporting
- Table-driven tests

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Test Organization

### Test Files

```
context/
├── context_suite_test.go       # Suite entry point
├── model_test.go               # Core model tests (20+ specs)
├── operations_test.go          # Operations tests (25+ specs)
├── utilities_test.go           # Utility functions (15+ specs)
└── context_integration_test.go # Integration tests (10+ specs)
```

---

## Running Tests

### Quick Test

```bash
cd context
go test -v
```

### With Coverage

```bash
go test -v -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### With Race Detection

```bash
CGO_ENABLED=1 go test -race -v
```

### Using Ginkgo

```bash
ginkgo -v
ginkgo -v -cover
ginkgo -v --trace
```

### Parallel Execution

```bash
ginkgo -v -p
CGO_ENABLED=1 ginkgo -v -p -race
```

---

## Test Coverage

### Coverage by Component

| Component | File | Specs | Coverage |
|-----------|------|-------|----------|
| Core Model | model_test.go | 20+ | ~95% |
| Operations | operations_test.go | 25+ | ~95% |
| Utilities | utilities_test.go | 15+ | ~90% |
| Integration | context_integration_test.go | 10+ | ~90% |

**Overall Coverage**: >90%

### Test Categories

#### 1. **model_test.go** - Core Functionality

**Scenarios:**
- Config creation and initialization
- Store and Load operations
- Thread-safe concurrent access
- Context function management
- Clean operation

**Example:**
```go
Describe("Config Creation", func() {
    It("should create config with context", func() {
        cfg := New[string](nil)
        Expect(cfg).NotTo(BeNil())
        Expect(cfg.GetContext()).NotTo(BeNil())
    })
})
```

#### 2. **operations_test.go** - Advanced Operations

**Scenarios:**
- Clone operation with independence
- Merge operation
- Walk and WalkLimit
- LoadOrStore atomic operation
- LoadAndDelete atomic operation
- Delete operation

**Example:**
```go
Describe("Clone", func() {
    It("should create independent copy", func() {
        original := New[string](nil)
        original.Store("key", "value")
        
        clone := original.Clone(nil)
        clone.Store("new", "data")
        
        _, ok := original.Load("new")
        Expect(ok).To(BeFalse())
    })
})
```

#### 3. **utilities_test.go** - Helper Functions

**Scenarios:**
- Walk operation
- WalkLimit with specific keys
- SetContext operation
- Context compatibility
- Edge cases

#### 4. **context_integration_test.go** - Integration Tests

**Scenarios:**
- Multi-goroutine scenarios
- Context cancellation propagation
- Complex merge operations
- Real-world usage patterns

---

## Concurrency Testing

### Race Detection

All tests pass with race detector:

```bash
CGO_ENABLED=1 go test -race -v
```

### Concurrent Operations Tested

1. **Concurrent Stores**
```go
It("should handle concurrent stores", func() {
    cfg := New[int](nil)
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            cfg.Store(n, n*2)
        }(i)
    }
    
    wg.Wait()
    // Verify all stored
})
```

2. **Concurrent Reads**
```go
It("should handle concurrent reads", func() {
    cfg := New[string](nil)
    cfg.Store("key", "value")
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            val, ok := cfg.Load("key")
            Expect(ok).To(BeTrue())
            Expect(val).To(Equal("value"))
        }()
    }
    
    wg.Wait()
})
```

3. **Mixed Read/Write**
```go
It("should handle mixed operations", func() {
    cfg := New[string](nil)
    var wg sync.WaitGroup
    
    // Writers
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            cfg.Store(fmt.Sprintf("key%d", n), n)
        }(i)
    }
    
    // Readers
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            cfg.Load(fmt.Sprintf("key%d", n))
        }(i)
    }
    
    wg.Wait()
})
```

---

## Best Practices

### 1. Use BeforeEach for Setup

```go
var cfg Config[string]

BeforeEach(func() {
    cfg = New[string](nil)
})
```

### 2. Test Both Success and Failure

```go
Context("when key exists", func() {
    It("should load value", func() { /* ... */ })
})

Context("when key doesn't exist", func() {
    It("should return false", func() { /* ... */ })
})
```

### 3. Test Edge Cases

- Nil values
- Empty strings
- Concurrent operations
- Context cancellation
- Large datasets

### 4. Verify Thread Safety

```go
It("should be thread-safe", func() {
    cfg := New[int](nil)
    
    done := make(chan bool)
    go func() {
        for i := 0; i < 1000; i++ {
            cfg.Store(i, i)
        }
        done <- true
    }()
    
    go func() {
        for i := 0; i < 1000; i++ {
            cfg.Load(i)
        }
        done <- true
    }()
    
    <-done
    <-done
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-context:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Test context package
      run: |
        cd context
        go test -v -race -cover
```

### GitLab CI

```yaml
test-context:
  script:
    - cd context
    - go test -v -race -cover
  coverage: '/coverage: \d+\.\d+% of statements/'
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Cover edge cases** (nil, empty, concurrent)
3. **Test with race detector**
4. **Verify thread safety**
5. **Update coverage** metrics
6. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    var cfg Config[string]
    
    BeforeEach(func() {
        cfg = New[string](nil)
    })
    
    Describe("Feature behavior", func() {
        It("should handle basic case", func() {
            // Test implementation
            Expect(result).To(Equal(expected))
        })
        
        Context("when concurrent", func() {
            It("should be thread-safe", func() {
                // Concurrent test
            })
        })
    })
})
```

---

## Support

For issues or questions:

- **Test Failures**: Check output with `-v` flag
- **Usage Examples**: Review test files
- **Feature Questions**: See [README.md](README.md)
- **Bug Reports**: [GitHub Issues](https://github.com/nabbar/golib/issues)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
