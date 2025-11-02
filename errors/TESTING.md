# Errors Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the errors package and pool sub-package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Main Package Testing](#main-package-testing)
- [Pool Sub-Package Testing](#pool-sub-package-testing)
- [Test Scenarios](#test-scenarios)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The errors package features comprehensive testing covering error creation, code management, hierarchy operations, stack tracing, and error pool functionality.

### Test Metrics Summary

| Package | Test Files | Key Areas | Framework |
|---------|------------|-----------|-----------|
| **errors** | 10+ files | Codes, Hierarchy, Trace, Modes | Ginkgo v2 |
| **errors/pool** | 3 files | Collection, Concurrency | Ginkgo v2 |

**Pool Sub-Package Metrics:**
- **83 specs** - All passing
- **100% coverage** - Complete code coverage
- **0 race conditions** - Verified with -race flag

---

## Test Framework

### Ginkgo v2 + Gomega

**Features:**
- BDD-style test organization
- Table-driven tests
- Concurrent execution support
- Rich assertion library

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Running Tests

### All Tests

```bash
# Main package
cd errors
go test -v

# Pool sub-package
cd errors/pool
CGO_ENABLED=1 go test -v -race -cover
```

### With Coverage

```bash
# Main package
cd errors
go test -v -cover

# Pool with race detection
cd errors/pool
CGO_ENABLED=1 go test -v -race -cover
```

### Using Ginkgo

```bash
# Run all
ginkgo -r -v

# Specific package
cd errors
ginkgo -v

cd errors/pool
ginkgo -v -race
```

---

## Main Package Testing

### Test Files Organization

```
errors/
├── errors_suite_test.go    # Suite setup
├── creation_test.go        # Error creation tests
├── code_test.go           # Error code tests
├── hierarchy_test.go      # Parent-child tests
├── trace_test.go          # Stack trace tests
├── return_test.go         # Return helper tests
├── mode_test.go           # Mode (dev/prod) tests
├── interface_test.go      # Interface tests
├── advanced_test.go       # Advanced scenarios
├── pattern_test.go        # Pattern matching tests
├── gin_test.go           # Gin integration tests
└── common_test.go         # Common utilities
```

### Key Test Areas

**Error Creation:**
- Basic error creation
- With parent errors
- Custom error codes
- Nil handling

**Error Codes:**
- Predefined HTTP codes
- Custom codes
- Code checking (IsCode, HasCode)
- Code retrieval

**Hierarchy:**
- Adding parents
- Setting parents
- Parent traversal
- Error chains

**Stack Tracing:**
- Automatic capture
- Manual setting
- Trace retrieval
- File/line accuracy

**Compatibility:**
- errors.Is integration
- errors.As integration
- Standard error interface

---

## Pool Sub-Package Testing

### Test Coverage Details

**Specs: 83 (All Passing)**
- Basic Operations: 45 specs
- Concurrent Operations: 22 specs
- Edge Cases: 16 specs

**Coverage: 100%**
- All functions tested
- All code paths covered
- Race conditions: 0

### Test File Organization

```
pool/
├── pool_suite_test.go    # Suite setup
├── basic_test.go         # Basic operations (45 specs)
├── concurrent_test.go    # Concurrency (22 specs)
└── edge_test.go          # Edge cases (16 specs)
```

### Test Scenarios

#### Basic Operations (45 specs)

**Pool Creation:**
```go
It("should create a new pool", func() {
    p := pool.New()
    Expect(p).NotTo(BeNil())
    Expect(p.Len()).To(Equal(uint64(0)))
})
```

**Add Operation:**
```go
It("should add errors", func() {
    p := pool.New()
    err1 := errors.New("error 1")
    err2 := errors.New("error 2")
    
    p.Add(err1, err2)
    
    Expect(p.Len()).To(Equal(uint64(2)))
    Expect(p.Get(1)).To(Equal(err1))
    Expect(p.Get(2)).To(Equal(err2))
})
```

**Get/Set/Del:**
```go
It("should manage errors by index", func() {
    p := pool.New()
    p.Add(errors.New("test"))
    
    // Get
    err := p.Get(1)
    Expect(err).NotTo(BeNil())
    
    // Set
    newErr := errors.New("updated")
    p.Set(1, newErr)
    Expect(p.Get(1)).To(Equal(newErr))
    
    // Del
    p.Del(1)
    Expect(p.Get(1)).To(BeNil())
})
```

#### Concurrent Operations (22 specs)

**Concurrent Add:**
```go
It("should handle concurrent additions", func() {
    p := pool.New()
    const goroutines = 100
    
    var wg sync.WaitGroup
    wg.Add(goroutines)
    
    for i := 0; i < goroutines; i++ {
        go func(id int) {
            defer wg.Done()
            p.Add(fmt.Errorf("error %d", id))
        }(i)
    }
    
    wg.Wait()
    Expect(p.Len()).To(Equal(uint64(goroutines)))
})
```

**Mixed Operations:**
```go
It("should handle all operations concurrently", func() {
    p := pool.New()
    var wg sync.WaitGroup
    
    // Add, Get, Set, Del operations in parallel
    // All operations should be thread-safe
})
```

#### Edge Cases (16 specs)

**Large Scale:**
```go
It("should handle many errors", func() {
    p := pool.New()
    const count = 10000
    
    for i := 0; i < count; i++ {
        p.Add(fmt.Errorf("error %d", i))
    }
    
    Expect(p.Len()).To(Equal(uint64(count)))
})
```

**Sparse Indices:**
```go
It("should handle sparse indices", func() {
    p := pool.New()
    p.Set(1, errors.New("error 1"))
    p.Set(100, errors.New("error 100"))
    p.Set(1000, errors.New("error 1000"))
    
    Expect(p.MaxId()).To(Equal(uint64(1000)))
    Expect(p.Len()).To(Equal(uint64(3)))
})
```

---

## Test Scenarios

### Error Code Testing

```go
var _ = Describe("Error Codes", func() {
    It("should create error with code", func() {
        err := liberr.NotFoundError.Error(nil)
        
        Expect(err).To(MatchError(MatchRegexp("404")))
        Expect(err.(liberr.Error).Code()).To(Equal(uint16(404)))
    })
    
    It("should check error codes", func() {
        err := liberr.InternalError.Error(nil)
        
        Expect(err.(liberr.Error).IsCode(liberr.InternalError)).To(BeTrue())
        Expect(err.(liberr.Error).IsCode(liberr.NotFoundError)).To(BeFalse())
    })
})
```

### Hierarchy Testing

```go
var _ = Describe("Error Hierarchy", func() {
    It("should chain errors", func() {
        err1 := errors.New("base error")
        err2 := liberr.InternalError.Error(err1)
        err3 := liberr.BadRequestError.Error(nil)
        err3.Add(err2)
        
        Expect(err3.HasParent()).To(BeTrue())
        Expect(err3.HasError(err1)).To(BeTrue())
    })
})
```

### Stack Trace Testing

```go
var _ = Describe("Stack Trace", func() {
    It("should capture stack trace", func() {
        err := liberr.InternalError.Error(nil)
        
        Expect(err.(liberr.Error).GetFile()).NotTo(BeEmpty())
        Expect(err.(liberr.Error).GetLine()).To(BeNumerically(">", 0))
    })
})
```

---

## Best Practices

### 1. Test Error Codes

```go
It("should have correct error code", func() {
    err := liberr.NotFoundError.Error(nil)
    
    Expect(err.(liberr.Error).Code()).To(Equal(uint16(404)))
    Expect(err.(liberr.Error).IsCode(liberr.NotFoundError)).To(BeTrue())
})
```

### 2. Test Hierarchy

```go
It("should maintain error hierarchy", func() {
    baseErr := errors.New("base")
    parentErr := liberr.InternalError.Error(baseErr)
    mainErr := liberr.BadRequestError.Error(nil)
    mainErr.Add(parentErr)
    
    parents := mainErr.GetParent(false)
    Expect(parents).To(HaveLen(1))
    Expect(mainErr.HasError(baseErr)).To(BeTrue())
})
```

### 3. Test Concurrency

```go
It("should be thread-safe", func() {
    p := pool.New()
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            p.Add(fmt.Errorf("error %d", id))
        }(i)
    }
    wg.Wait()
    
    Expect(p.Len()).To(Equal(uint64(100)))
})
```

### 4. Test Edge Cases

```go
It("should handle nil errors", func() {
    p := pool.New()
    p.Add(nil, nil, nil)
    Expect(p.Len()).To(Equal(uint64(0)))
})

It("should handle empty pool", func() {
    p := pool.New()
    Expect(p.Error()).To(BeNil())
    Expect(p.Last()).To(BeNil())
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-errors:
  runs-on: ubuntu-latest
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test Errors Package
      run: |
        cd errors
        go test -v -cover
    
    - name: Test Pool Package
      run: |
        cd errors/pool
        CGO_ENABLED=1 go test -v -race -cover
```

### GitLab CI

```yaml
test-errors:
  script:
    - cd errors
    - go test -v -cover
    - cd pool
    - CGO_ENABLED=1 go test -v -race -cover
  
  coverage: '/coverage: \d+\.\d+% of statements/'
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Test error codes** and their behavior
3. **Test hierarchy** operations
4. **Test concurrency** (use -race flag)
5. **Test edge cases** (nil, empty, large scale)
6. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    var err liberr.Error
    
    BeforeEach(func() {
        err = liberr.InternalError.Error(nil)
    })
    
    Describe("basic functionality", func() {
        It("should work correctly", func() {
            // Test implementation
        })
    })
    
    Context("error conditions", func() {
        It("should handle errors", func() {
            // Error test
        })
    })
})
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
