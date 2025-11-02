# Testing Guide - Cache Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Comprehensive testing documentation for the cache package.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Test Organization](#test-organization)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [Resources](#resources)

---

## Overview

The cache package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) to achieve 96.7% test coverage across 64 test specifications.

### Test Philosophy

1. **Behavior-Driven**: Tests describe behavior in readable, hierarchical structures
2. **Thread-Safe Verification**: All tests pass race detection (`go test -race`)
3. **Comprehensive Coverage**: Core operations, edge cases, concurrency, and context integration
4. **Fast Execution**: No external dependencies, tests run in milliseconds
5. **Maintainable**: Clear test structure with reusable patterns

### Coverage Scope

- ✅ Core operations (Store, Load, Delete, LoadOrStore, LoadAndDelete, Swap)
- ✅ Advanced features (Clone, Merge, Walk, Expire, Clean)
- ✅ Context integration (Deadline, Done, Err, Value)
- ✅ Automatic expiration mechanics
- ✅ Thread safety and concurrent access
- ✅ Edge cases (expired items, cancelled contexts, zero values)
- ✅ Cache item internals

---

## Test Framework

### Ginkgo v2

[Ginkgo](https://onsi.github.io/ginkgo/) is a BDD-style testing framework providing:
- Hierarchical test organization (`Describe`, `Context`, `It` blocks)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `DeferCleanup`)
- Parallel test execution
- Rich CLI with filtering and reporting
- Excellent failure diagnostics

### Gomega

[Gomega](https://onsi.github.io/gomega/) is the matcher library offering:
- Readable assertion syntax: `Expect(value).To(Equal(expected))`
- Extensive built-in matchers
- Asynchronous assertions: `Eventually(fn).Should(BeClosed())`
- Detailed failure messages
- Custom matcher support

---

## Running Tests

### Quick Start

**Standard Go Testing**
```bash
# Run all tests
go test ./...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Ginkgo CLI** (install: `go install github.com/onsi/ginkgo/v2/ginkgo@latest`)
```bash
# Run all tests recursively
ginkgo -r

# Verbose output
ginkgo -v -r

# With coverage
ginkgo -cover -r

# Parallel execution
ginkgo -p -r

# Watch mode (re-run on file changes)
ginkgo watch -r
```

### Coverage Reports

**Generate HTML Coverage Report**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**View Coverage by Function**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Advanced Options

**Focus on Specific Tests**
```bash
# Run tests matching pattern
ginkgo --focus="LoadOrStore" -r

# Skip tests matching pattern
ginkgo --skip="Context" -r
```

**Output Formats**
```bash
# JUnit XML report (CI integration)
ginkgo --junit-report=test-results.xml -r

# JSON output
ginkgo --json-report=test-results.json -r
```

**Specific Packages**
```bash
# Test only cache package
cd /path/to/golib/cache
go test .

# Test only item subpackage
cd item
go test .
```

---

## Test Coverage

### Coverage Metrics

| Package | Coverage | Specs | Lines Covered |
|---------|----------|-------|---------------|
| `cache` | 96.7% | 43 | ~350 |
| `cache/item` | 96.7% | 21 | ~120 |
| **Total** | **96.7%** | **64** | **~470** |

### Coverage by Component

**Core Cache Operations** (`cache_test.go` - 13 specs)
- Cache creation (with/without expiration)
- Store, Load, Delete operations
- LoadOrStore, LoadAndDelete, Swap atomic operations
- Walk iteration
- Merge and Clone operations
- Clean, Expire, Close lifecycle methods

**Context Integration** (`context_test.go` - 11 specs)
- Deadline, Done, Err methods
- Value method (cache keys + parent context fallback)
- Context cancellation behavior

**Advanced Operations** (`advanced_operations_test.go` - 19 specs)
- Merge with various scenarios
- LoadOrStore edge cases
- LoadAndDelete edge cases
- Clone with cancelled context
- Walk with early termination
- Operations with cancelled contexts

**Cache Item Basics** (`item/item_test.go` - 5 specs)
- Item initialization
- Store and Load (with/without expiration)
- Expiration mechanics
- Clean and Remain operations

**Cache Item Edge Cases** (`item/item_edge_cases_test.go` - 16 specs)
- LoadRemain with various states
- Check and Duration methods
- Different data types (string, struct, pointer)
- Concurrent access patterns
- Remain method accuracy

---

## Test Organization

### File Structure

```
cache/
├── cache_suite_test.go              # Main test suite setup
├── cache_test.go                    # Core operations (13 specs)
├── context_test.go                  # Context integration (11 specs)
├── advanced_operations_test.go      # Advanced features (19 specs)
└── item/
    ├── item_suite_test.go          # Item test suite setup
    ├── item_test.go                # Basic operations (5 specs)
    └── item_edge_cases_test.go     # Edge cases (16 specs)
```

### Test Structure Pattern

```go
// Hierarchical BDD structure
Describe("cache/Feature", func() {
    var c cache.Cache[string, int]
    
    BeforeEach(func() {
        c = cache.New[string, int](context.Background(), 0)
    })
    
    AfterEach(func() {
        _ = c.Close()
    })
    
    Context("When specific condition", func() {
        It("should exhibit expected behavior", func() {
            // Arrange
            c.Store("key", 42)
            
            // Act
            value, _, ok := c.Load("key")
            
            // Assert
            Expect(ok).To(BeTrue())
            Expect(value).To(Equal(42))
        })
    })
})
```

---

## Writing Tests

### Test Guidelines

**1. Descriptive Test Names**
```go
// ✅ Good: Clear, specific description
It("should return false when loading expired items", func() {
    // Test implementation
})

// ❌ Bad: Vague description
It("should work", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should store and retrieve values", func() {
    // Arrange
    c := cache.New[string, int](context.Background(), 0)
    DeferCleanup(func() { _ = c.Close() })
    
    // Act
    c.Store("key", 42)
    value, _, ok := c.Load("key")
    
    // Assert
    Expect(ok).To(BeTrue())
    Expect(value).To(Equal(42))
})
```

**3. Use Appropriate Matchers**
```go
// Common Gomega matchers
Expect(value).To(Equal(42))
Expect(err).ToNot(HaveOccurred())
Expect(slice).To(HaveLen(5))
Expect(duration).To(BeNumerically(">=", 0))
Expect(ok).To(BeTrue())
Eventually(channel).Should(BeClosed())
```

**4. Clean Up Resources**
```go
// ✅ Good: Use DeferCleanup
It("should work correctly", func() {
    c := cache.New[string, int](context.Background(), 0)
    DeferCleanup(func() { _ = c.Close() })
    // Test logic
})

// ✅ Good: Use BeforeEach/AfterEach for shared setup
var c cache.Cache[string, int]

BeforeEach(func() {
    c = cache.New[string, int](context.Background(), 0)
})

AfterEach(func() {
    _ = c.Close()
})
```

**5. Test Edge Cases**
```go
// Test with zero values, nil, expired items, cancelled contexts
It("should handle expired items", func() {
    c := cache.New[string, int](ctx, 20*time.Millisecond)
    DeferCleanup(func() { _ = c.Close() })
    
    c.Store("key", 1)
    time.Sleep(30 * time.Millisecond)
    
    _, _, ok := c.Load("key")
    Expect(ok).To(BeFalse())
})
```

### Test Template

```go
var _ = Describe("cache/NewFeature", func() {
    var c cache.Cache[string, int]
    
    BeforeEach(func() {
        c = cache.New[string, int](context.Background(), 0)
    })
    
    AfterEach(func() {
        _ = c.Close()
    })
    
    Context("When feature is used", func() {
        It("should perform expected behavior", func() {
            // Arrange
            expectedValue := 42
            
            // Act
            c.Store("key", expectedValue)
            value, remaining, ok := c.Load("key")
            
            // Assert
            Expect(ok).To(BeTrue())
            Expect(value).To(Equal(expectedValue))
            Expect(remaining).To(Equal(time.Duration(0)))
        })
        
        It("should handle edge case", func() {
            // Test edge case
            _, _, ok := c.Load("nonexistent")
            Expect(ok).To(BeFalse())
        })
    })
})
```

---

## Best Practices

### Test Independence
```go
// ✅ Good: Each test is independent
It("should store value", func() {
    c := cache.New[string, int](context.Background(), 0)
    DeferCleanup(func() { _ = c.Close() })
    c.Store("key", 1)
    Expect(c.Load("key")).To(/* ... */)
})

// ❌ Bad: Tests depend on execution order
var sharedCache cache.Cache[string, int]

It("test1", func() {
    sharedCache.Store("key", 1)
})

It("test2", func() {
    // Depends on test1!
    value, _, _ := sharedCache.Load("key")
})
```

### Timing Considerations
```go
// ✅ Good: Use Eventually for async operations
cancel()
Eventually(c.Err).ShouldNot(BeNil())

// ✅ Good: Realistic sleep durations
c := cache.New[string, int](ctx, 20*time.Millisecond)
c.Store("key", 1)
time.Sleep(30 * time.Millisecond) // 50% buffer

// ❌ Bad: Race-prone timing
c := cache.New[string, int](ctx, 10*time.Millisecond)
time.Sleep(10 * time.Millisecond) // Might not expire yet!
```

### Focused Assertions
```go
// ✅ Good: One behavior per test
It("should return true when value exists", func() {
    c.Store("key", 42)
    _, _, ok := c.Load("key")
    Expect(ok).To(BeTrue())
})

It("should return correct value", func() {
    c.Store("key", 42)
    value, _, _ := c.Load("key")
    Expect(value).To(Equal(42))
})

// ❌ Bad: Testing multiple behaviors
It("should work", func() {
    c.Store("key", 42)
    value, remaining, ok := c.Load("key")
    Expect(ok).To(BeTrue())
    Expect(value).To(Equal(42))
    Expect(remaining).To(BeZero())
    // Too many assertions!
})
```

### Resource Management
```go
// ✅ Good: Proper cleanup
BeforeEach(func() {
    c = cache.New[string, int](context.Background(), 0)
})

AfterEach(func() {
    _ = c.Close() // Always clean up
})

// ✅ Good: DeferCleanup for one-off resources
It("should handle context cancellation", func() {
    ctx, cancel := context.WithCancel(context.Background())
    DeferCleanup(cancel)
    
    c := cache.New[string, int](ctx, 0)
    DeferCleanup(func() { _ = c.Close() })
})
```

### Documentation
```go
// ✅ Good: Clear description
It("should remove expired items during Walk operation", func() {
    c := cache.New[string, int](ctx, 20*time.Millisecond)
    DeferCleanup(func() { _ = c.Close() })
    
    c.Store("key1", 1)
    time.Sleep(30 * time.Millisecond) // Allow expiration
    
    // Walk should skip expired items
    walked := false
    c.Walk(func(k string, v int, r time.Duration) bool {
        walked = true
        return true
    })
    
    Expect(walked).To(BeFalse(), "expired item should not be walked")
})
```

---

## Troubleshooting

### Common Issues

**Flaky Timing Tests**

*Problem*: Tests fail intermittently due to timing.

*Solution*: Use `Eventually` for asynchronous operations:
```bash
// ✅ Good
cancel()
Eventually(c.Err).ShouldNot(BeNil())

// ✅ Good: Buffer for timing variations
c := cache.New[string, int](ctx, 20*time.Millisecond)
c.Store("key", 1)
time.Sleep(30 * time.Millisecond) // 50% buffer
```

**Import Errors**

*Problem*: Cannot import test packages.

*Solution*:
```bash
go mod tidy
go mod download
```

**Stale Coverage**

*Problem*: Coverage report doesn't reflect recent changes.

*Solution*:
```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

**Race Conditions**

*Problem*: Race detector reports data races.

*Solution*:
```bash
# Run tests with race detection
CGO_ENABLED=1 go test -race ./...

# Fix any reported races by reviewing thread safety
```

### Debugging Techniques

**Run Specific Tests**
```bash
# Focus on specific test
ginkgo --focus="LoadOrStore" -r

# Skip specific tests
ginkgo --skip="Context" -r

# Run single file
go test -run TestCacheSuite/cache/Store
```

**Verbose Output**
```bash
# Ginkgo verbose mode
ginkgo -v --trace -r

# Standard Go verbose
go test -v ./...
```

**Debug Logging**
```go
It("should do something", func() {
    // Use GinkgoWriter for test output
    fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
    
    // Or use custom logging
    By("storing value in cache")
    c.Store("key", 42)
    
    By("loading value from cache")
    value, _, ok := c.Load("key")
    
    Expect(ok).To(BeTrue())
})
```

**Check Test Coverage**
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# View by function
go tool cover -func=coverage.out
```

---

## Contributing

### Test Contributions

**Guidelines**
- Do not use AI to generate test implementation code
- AI may assist with test documentation and bug fixing
- All tests must pass race detection (`go test -race`)
- Maintain or improve coverage (≥96%)
- Follow existing test patterns

**Adding New Tests**
1. Choose appropriate test file based on feature
2. Use descriptive test names
3. Follow AAA pattern (Arrange, Act, Assert)
4. Add cleanup with `DeferCleanup` or `AfterEach`
5. Test edge cases and error conditions
6. Verify thread safety

**Test Review Checklist**
- [ ] Tests are independent
- [ ] Resources are properly cleaned up
- [ ] Edge cases are covered
- [ ] Timing is realistic (no races)
- [ ] Descriptions are clear
- [ ] Passes race detection
- [ ] Coverage maintained or improved

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Documentation**
- [Ginkgo v2 Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matcher Reference](https://onsi.github.io/gomega/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Context Package](https://pkg.go.dev/context)

**Cache Package**
- [Cache Package GoDoc](https://pkg.go.dev/github.com/nabbar/golib/cache)
- [README.md](README.md) - Package overview and examples
- [GitHub Repository](https://github.com/nabbar/golib)

**Testing Tools**
- [Go Test Command](https://pkg.go.dev/cmd/go#hdr-Test_packages)
- [Race Detector](https://go.dev/doc/articles/race_detector)
- [Coverage Tool](https://go.dev/blog/cover)

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.
