# Encoding Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the encoding package and its sub-packages using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Sub-Package Testing](#sub-package-testing)
- [Interface Testing](#interface-testing)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The encoding package features comprehensive testing across all sub-packages with a unified testing approach using Ginkgo v2/Gomega.

### Test Metrics Summary

| Package | Specs | Coverage | Status |
|---------|-------|----------|--------|
| **aes** | 126 | 91.5% | ✅ All passing |
| **hexa** | 97 | 89.7% | ✅ All passing |
| **mux** | 59 | 81.7% | ✅ All passing |
| **randRead** | 41 | 81.4% | ✅ All passing |
| **sha256** | 61 | 84.8% | ✅ All passing |
| **Total** | 384 | 85.8% | ✅ Excellent |

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

### All Packages

```bash
cd encoding
go test -v ./...
```

### With Coverage

```bash
go test -v -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Using Ginkgo

```bash
# Run all tests
ginkgo -r -v

# Parallel execution
ginkgo -r -v -p

# With coverage
ginkgo -r -v -cover

# Specific package
ginkgo -v ./aes
```

---

## Sub-Package Testing

### AES Package

**Test Coverage**: 126 specs, 91.5%

**Key Test Areas:**
- Key/nonce generation
- Encryption/decryption
- Authentication verification
- Streaming operations
- Error handling

**Documentation**: [aes/TESTING.md](aes/TESTING.md)

**Quick Test:**
```bash
cd aes
CGO_ENABLED=1 go test -v -cover
```

---

### Hexa Package

**Test Coverage**: 97 specs, 89.7%

**Key Test Areas:**
- Hex encoding/decoding
- Case insensitivity
- Round-trip verification
- Streaming operations
- Edge cases

**Documentation**: [hexa/TESTING.md](hexa/TESTING.md)

**Quick Test:**
```bash
cd hexa
go test -v -cover
```

---

### Mux Package

**Test Coverage**: 59 specs, 81.7%

**Key Test Areas:**
- Multiplexing operations
- Demultiplexing operations
- Channel routing
- Concurrent access
- Round-trip verification

**Documentation**: [mux/TESTING.md](mux/TESTING.md)

**Quick Test:**
```bash
cd mux
go test -v -race
```

---

### RandRead Package

**Test Coverage**: 41 specs, 81.4%

**Key Test Areas:**
- Remote source management
- Buffering behavior
- Error handling
- Reconnection logic
- Nil handling

**Documentation**: [randRead/TESTING.md](randRead/TESTING.md)

**Quick Test:**
```bash
cd randRead
go test -v -cover
```

---

### SHA256 Package

**Test Coverage**: 61 specs, 84.8%

**Key Test Areas:**
- Hash computation
- Known test vectors (NIST)
- Streaming operations
- Reset behavior
- Determinism

**Documentation**: [sha256/TESTING.md](sha256/TESTING.md)

**Quick Test:**
```bash
cd sha256
go test -v -cover
```

---

## Interface Testing

### Coder Interface Compliance

All implementations must satisfy the `Coder` interface:

```go
var _ = Describe("Coder Interface", func() {
    var coder encoding.Coder
    
    BeforeEach(func() {
        coder = NewImplementation()
    })
    
    It("should implement Encode", func() {
        result := coder.Encode([]byte("test"))
        Expect(result).NotTo(BeNil())
    })
    
    It("should implement Decode", func() {
        encoded := coder.Encode([]byte("test"))
        decoded, err := coder.Decode(encoded)
        Expect(err).NotTo(HaveOccurred())
        Expect(decoded).To(Equal([]byte("test")))
    })
    
    It("should implement streaming", func() {
        reader := bytes.NewReader([]byte("test"))
        encReader := coder.EncodeReader(reader)
        Expect(encReader).NotTo(BeNil())
    })
})
```

### Round-Trip Testing Pattern

```go
It("should preserve data through round-trip", func() {
    original := []byte("Round trip test data")
    
    // Encode
    encoded := coder.Encode(original)
    Expect(encoded).NotTo(BeNil())
    
    // Decode
    decoded, err := coder.Decode(encoded)
    Expect(err).NotTo(HaveOccurred())
    Expect(decoded).To(Equal(original))
})
```

---

## Best Practices

### 1. Test All Interface Methods

```go
var _ = Describe("Full Interface", func() {
    It("should implement all methods", func() {
        var _ encoding.Coder = implementation
    })
})
```

### 2. Use Table-Driven Tests

```go
DescribeTable("encoding various inputs",
    func(input []byte, shouldError bool) {
        encoded := coder.Encode(input)
        decoded, err := coder.Decode(encoded)
        
        if shouldError {
            Expect(err).To(HaveOccurred())
        } else {
            Expect(err).NotTo(HaveOccurred())
            Expect(decoded).To(Equal(input))
        }
    },
    Entry("empty", []byte{}, false),
    Entry("small", []byte("test"), false),
    Entry("large", make([]byte, 1024*1024), false),
)
```

### 3. Test Concurrency

```go
It("should be safe for concurrent use", func() {
    done := make(chan bool)
    
    for i := 0; i < 10; i++ {
        go func(id int) {
            data := []byte(fmt.Sprintf("data%d", id))
            encoded := coder.Encode(data)
            decoded, _ := coder.Decode(encoded)
            Expect(decoded).To(Equal(data))
            done <- true
        }(i)
    }
    
    for i := 0; i < 10; i++ {
        <-done
    }
})
```

### 4. Test Error Paths

```go
It("should handle errors gracefully", func() {
    _, err := coder.Decode([]byte("invalid"))
    Expect(err).To(HaveOccurred())
})
```

### 5. Test Resource Cleanup

```go
It("should clean up resources", func() {
    coder.Reset()
    // Verify cleanup
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-encoding:
  runs-on: ubuntu-latest
  
  strategy:
    matrix:
      package: [aes, hexa, mux, randRead, sha256]
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test ${{ matrix.package }}
      run: |
        cd encoding/${{ matrix.package }}
        go test -v -race -cover
```

### GitLab CI

```yaml
test-encoding:
  script:
    - cd encoding
    - go test -v -race -cover ./...
  
  coverage: '/coverage: \d+\.\d+% of statements/'
```

### Coverage Report

```bash
# Generate combined coverage
cd encoding
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Contributing

When adding new features or implementations:

1. **Write tests first** (TDD approach)
2. **Implement full interface** (all Coder methods)
3. **Test round-trips** (encode → decode)
4. **Test concurrency** (race detection)
5. **Test error cases** thoroughly
6. **Document test scenarios**
7. **Update coverage** metrics

### Test Template

```go
var _ = Describe("New Implementation", func() {
    var coder encoding.Coder
    
    BeforeEach(func() {
        coder = NewImplementation()
    })
    
    AfterEach(func() {
        coder.Reset()
    })
    
    Describe("basic operations", func() {
        It("should encode", func() {
            // Test
        })
        
        It("should decode", func() {
            // Test
        })
    })
    
    Describe("streaming", func() {
        It("should support readers", func() {
            // Test
        })
    })
    
    Describe("error handling", func() {
        It("should handle errors", func() {
            // Test
        })
    })
})
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
