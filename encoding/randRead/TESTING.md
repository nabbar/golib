# Random Reader Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the random reader package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Scenarios](#test-scenarios)
- [Error Testing](#error-testing)
- [Reconnection Testing](#reconnection-testing)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The randRead package features comprehensive testing covering remote source management, buffering, error handling, and reconnection scenarios.

---

## Test Framework

### Ginkgo v2 + Gomega

**Features:**
- BDD-style test organization
- Error scenario testing
- Reconnection behavior verification
- Mock remote sources

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Test Organization

### Test Files

```
encoding/randRead/
├── randRead_suite_test.go   # Suite setup
├── randRead_test.go         # Basic tests
├── reader_test.go           # Reader behavior tests
└── error_test.go            # Error handling tests
```

---

## Running Tests

### Quick Test

```bash
cd encoding/randRead
go test -v
```

### With Coverage

```bash
go test -v -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Using Ginkgo

```bash
# Run all tests
ginkgo -v

# With coverage
ginkgo -v -cover
```

---

## Test Scenarios

### 1. Basic Operations

**Example:**
```go
var _ = Describe("RandRead", func() {
    It("should create reader", func() {
        source := func() (io.ReadCloser, error) {
            return io.NopCloser(bytes.NewReader([]byte("test"))), nil
        }
        
        reader := New(source)
        Expect(reader).NotTo(BeNil())
    })
    
    It("should read data", func() {
        data := []byte("Hello World")
        source := func() (io.ReadCloser, error) {
            return io.NopCloser(bytes.NewReader(data)), nil
        }
        
        reader := New(source)
        defer reader.Close()
        
        buffer := make([]byte, len(data))
        n, err := reader.Read(buffer)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(n).To(Equal(len(data)))
        Expect(buffer).To(Equal(data))
    })
})
```

### 2. Error Handling

**Example:**
```go
It("should handle source errors", func() {
    callCount := 0
    source := func() (io.ReadCloser, error) {
        callCount++
        if callCount == 1 {
            return nil, errors.New("connection failed")
        }
        return io.NopCloser(bytes.NewReader([]byte("success"))), nil
    }
    
    reader := New(source)
    defer reader.Close()
    
    buffer := make([]byte, 100)
    n, err := reader.Read(buffer)
    
    // Should succeed after retry
    Expect(err).NotTo(HaveOccurred())
    Expect(n).To(BeNumerically(">", 0))
    Expect(callCount).To(Equal(2))  // Called twice
})
```

### 3. Reconnection

**Example:**
```go
It("should reconnect on read error", func() {
    callCount := 0
    source := func() (io.ReadCloser, error) {
        callCount++
        if callCount == 1 {
            // First connection returns reader that fails
            return &failingReader{}, nil
        }
        // Second connection succeeds
        return io.NopCloser(bytes.NewReader([]byte("recovered"))), nil
    }
    
    reader := New(source)
    defer reader.Close()
    
    buffer := make([]byte, 100)
    n, err := reader.Read(buffer)
    
    Expect(err).NotTo(HaveOccurred())
    Expect(string(buffer[:n])).To(Equal("recovered"))
    Expect(callCount).To(Equal(2))
})
```

### 4. Nil Handling

**Example:**
```go
It("should handle nil function", func() {
    reader := New(nil)
    Expect(reader).To(BeNil())
})

It("should handle nil reader from source", func() {
    source := func() (io.ReadCloser, error) {
        return nil, errors.New("no reader available")
    }
    
    reader := New(source)
    defer reader.Close()
    
    buffer := make([]byte, 100)
    _, err := reader.Read(buffer)
    
    Expect(err).To(HaveOccurred())
})
```

---

## Error Testing

### Mock Failing Reader

```go
type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
    return 0, errors.New("read failed")
}

func (f *failingReader) Close() error {
    return nil
}

// Use in tests
It("should handle reader failure", func() {
    source := func() (io.ReadCloser, error) {
        return &failingReader{}, nil
    }
    
    reader := New(source)
    defer reader.Close()
    
    buffer := make([]byte, 100)
    _, err := reader.Read(buffer)
    
    Expect(err).To(HaveOccurred())
})
```

### Intermittent Failures

```go
It("should recover from intermittent failures", func() {
    attemptCount := 0
    source := func() (io.ReadCloser, error) {
        attemptCount++
        if attemptCount <= 3 {
            return nil, fmt.Errorf("attempt %d failed", attemptCount)
        }
        return io.NopCloser(bytes.NewReader([]byte("success"))), nil
    }
    
    reader := New(source)
    defer reader.Close()
    
    buffer := make([]byte, 100)
    n, err := reader.Read(buffer)
    
    Expect(err).NotTo(HaveOccurred())
    Expect(attemptCount).To(Equal(4))
})
```

---

## Reconnection Testing

### Verify Reconnection Count

```go
It("should limit reconnection attempts", func() {
    attempts := 0
    source := func() (io.ReadCloser, error) {
        attempts++
        return nil, errors.New("always fails")
    }
    
    reader := New(source)
    defer reader.Close()
    
    buffer := make([]byte, 100)
    reader.Read(buffer)
    
    // Check attempts were made
    Expect(attempts).To(BeNumerically(">=", 1))
})
```

### Test Connection Reuse

```go
It("should reuse connection", func() {
    connectionCount := 0
    source := func() (io.ReadCloser, error) {
        connectionCount++
        data := bytes.Repeat([]byte("data"), 1000)
        return io.NopCloser(bytes.NewReader(data)), nil
    }
    
    reader := New(source)
    defer reader.Close()
    
    // Multiple reads
    for i := 0; i < 100; i++ {
        buffer := make([]byte, 10)
        reader.Read(buffer)
    }
    
    // Should only connect once
    Expect(connectionCount).To(Equal(1))
})
```

---

## Best Practices

### 1. Use BeforeEach/AfterEach

```go
var _ = Describe("Test Suite", func() {
    var reader io.ReadCloser
    
    BeforeEach(func() {
        source := func() (io.ReadCloser, error) {
            return io.NopCloser(bytes.NewReader([]byte("test"))), nil
        }
        reader = New(source)
    })
    
    AfterEach(func() {
        if reader != nil {
            reader.Close()
        }
    })
    
    It("test case", func() {
        // Use reader
    })
})
```

### 2. Test Error Paths

```go
It("should handle all error scenarios", func() {
    // Test nil source
    // Test connection errors
    // Test read errors
    // Test close errors
})
```

### 3. Verify Cleanup

```go
It("should close underlying connection", func() {
    closed := false
    source := func() (io.ReadCloser, error) {
        return &mockCloser{onClose: func() { closed = true }}, nil
    }
    
    reader := New(source)
    reader.Close()
    
    Expect(closed).To(BeTrue())
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-randRead:
  runs-on: ubuntu-latest
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test RandRead Package
      run: |
        cd encoding/randRead
        go test -v -cover
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Test error scenarios** thoroughly
3. **Test reconnection** behavior
4. **Mock external dependencies**
5. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    var reader io.ReadCloser
    
    BeforeEach(func() {
        source := func() (io.ReadCloser, error) {
            return mockSource(), nil
        }
        reader = New(source)
    })
    
    AfterEach(func() {
        reader.Close()
    })
    
    It("should work correctly", func() {
        // Test implementation
    })
    
    Context("error cases", func() {
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
