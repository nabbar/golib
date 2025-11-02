# Hexadecimal Encoding Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the hexadecimal encoding package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Test Scenarios](#test-scenarios)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The hexa package features comprehensive testing covering encoding, decoding, streaming operations, and edge cases.

### Test Metrics

| Metric | Value | Status |
|--------|-------|--------|
| **Total Specs** | 97 | ✅ All passing |
| **Code Coverage** | 89.7% | ✅ Excellent |
| **Framework** | Ginkgo v2 + Gomega | ✅ BDD |
| **Edge Cases** | 25+ | ✅ Robust |

---

## Test Framework

### Ginkgo v2 + Gomega

**Features:**
- BDD-style test organization
- Table-driven tests for encoding/decoding
- Comprehensive edge case coverage
- Round-trip verification
- Streaming I/O testing

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Test Organization

### Test Files

```
encoding/hexa/
├── hexa_suite_test.go    # Suite setup
├── hexa_test.go          # Basic hexa tests (15+ specs)
├── encode_test.go        # Encoding tests (20+ specs)
├── reader_test.go        # Streaming read tests (15+ specs)
├── writer_test.go        # Streaming write tests (20+ specs)
└── edge_test.go          # Edge cases & errors (25+ specs)
```

### Test Categories

1. **Basic Operations** - Creation, basic encode/decode
2. **Encoding** - Various input sizes and patterns
3. **Decoding** - Valid/invalid hex, case sensitivity
4. **Streaming** - Reader/Writer interfaces
5. **Edge Cases** - Empty, invalid, large data
6. **Round-Trip** - Encode → Decode verification

---

## Running Tests

### Quick Test

```bash
cd encoding/hexa
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

# Parallel execution
ginkgo -v -p

# Focus on specific tests
ginkgo -v -focus="Encoding"

# Skip tests
ginkgo -v -skip="Edge cases"
```

### Verbose Output

```bash
go test -v -count=1
```

---

## Test Coverage

### Coverage by Component

| Component | File | Specs | Coverage | Notes |
|-----------|------|-------|----------|-------|
| Interface | interface.go | 5+ | 95% | New function |
| Encoding | model.go | 20+ | 92% | Encode operations |
| Decoding | model.go | 20+ | 92% | Decode operations |
| Streaming | model.go | 30+ | 85% | Reader/Writer |
| Edge Cases | all | 25+ | 88% | Error handling |

**Overall Coverage**: 89.7%

### Coverage Gaps

Minor gaps in:
- Some I/O error conditions
- Extreme memory scenarios
- Rare buffer edge cases

---

## Test Scenarios

### 1. Basic Operations

**Scenarios:**
- Create coder instance
- Reset (no-op)
- Instance reuse

**Example:**
```go
var _ = Describe("Basic Operations", func() {
    It("should create new coder", func() {
        coder := New()
        Expect(coder).NotTo(BeNil())
    })
    
    It("should be stateless", func() {
        coder1 := New()
        coder2 := New()
        
        data := []byte("test")
        enc1 := coder1.Encode(data)
        enc2 := coder2.Encode(data)
        
        Expect(enc1).To(Equal(enc2))
    })
    
    It("should handle reset (no-op)", func() {
        coder := New()
        data := []byte("test")
        
        enc1 := coder.Encode(data)
        coder.Reset()  // No-op
        enc2 := coder.Encode(data)
        
        Expect(enc1).To(Equal(enc2))
    })
})
```

### 2. Encoding Tests

**Scenarios:**
- Encode various data sizes
- Encode empty data
- Encode large data
- Verify output format (lowercase)
- Verify size (2× input)

**Example:**
```go
var _ = Describe("Encoding", func() {
    var coder Coder
    
    BeforeEach(func() {
        coder = New()
    })
    
    It("should encode to lowercase hex", func() {
        input := []byte("Hello")
        encoded := coder.Encode(input)
        
        Expect(string(encoded)).To(Equal("48656c6c6f"))
        // Verify lowercase (not "48656C6C6F")
    })
    
    It("should double the size", func() {
        input := []byte("Test")
        encoded := coder.Encode(input)
        
        Expect(len(encoded)).To(Equal(len(input) * 2))
    })
    
    It("should encode empty data", func() {
        input := []byte{}
        encoded := coder.Encode(input)
        
        Expect(len(encoded)).To(Equal(0))
    })
    
    It("should encode binary data", func() {
        input := []byte{0x00, 0xFF, 0xAA, 0x55}
        encoded := coder.Encode(input)
        
        Expect(string(encoded)).To(Equal("00ffaa55"))
    })
})
```

### 3. Decoding Tests

**Scenarios:**
- Decode valid hex
- Case-insensitive decoding
- Detect invalid characters
- Detect odd length
- Round-trip verification

**Example:**
```go
var _ = Describe("Decoding", func() {
    var coder Coder
    
    BeforeEach(func() {
        coder = New()
    })
    
    It("should decode valid hex", func() {
        hexData := []byte("48656c6c6f")
        decoded, err := coder.Decode(hexData)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(string(decoded)).To(Equal("Hello"))
    })
    
    It("should be case-insensitive", func() {
        lower := []byte("48656c6c6f")
        upper := []byte("48656C6C6F")
        mixed := []byte("48656C6c6F")
        
        dec1, _ := coder.Decode(lower)
        dec2, _ := coder.Decode(upper)
        dec3, _ := coder.Decode(mixed)
        
        Expect(dec1).To(Equal(dec2))
        Expect(dec2).To(Equal(dec3))
    })
    
    It("should reject invalid hex characters", func() {
        invalidHex := []byte("48656g6c6f")  // 'g' is invalid
        _, err := coder.Decode(invalidHex)
        
        Expect(err).To(HaveOccurred())
    })
    
    It("should reject odd length", func() {
        oddHex := []byte("48656c6c6")  // Odd length
        _, err := coder.Decode(oddHex)
        
        Expect(err).To(HaveOccurred())
    })
    
    It("should perform round-trip", func() {
        original := []byte("Round trip test")
        
        encoded := coder.Encode(original)
        decoded, err := coder.Decode(encoded)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(decoded).To(Equal(original))
    })
})
```

### 4. Streaming Tests

**Scenarios:**
- Encode via Reader
- Decode via Reader
- Handle large streams
- Handle empty streams
- Error propagation

**Example:**
```go
var _ = Describe("Streaming Operations", func() {
    var coder Coder
    
    BeforeEach(func() {
        coder = New()
    })
    
    It("should encode via reader", func() {
        input := []byte("Stream test")
        reader := bytes.NewReader(input)
        
        hexReader := coder.EncodeReader(reader)
        encoded, err := io.ReadAll(hexReader)
        
        Expect(err).NotTo(HaveOccurred())
        
        // Verify it's hex
        decoded, _ := coder.Decode(encoded)
        Expect(decoded).To(Equal(input))
    })
    
    It("should decode via reader", func() {
        hexData := []byte("53747265616d")
        reader := bytes.NewReader(hexData)
        
        binaryReader := coder.DecodeReader(reader)
        decoded, err := io.ReadAll(binaryReader)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(string(decoded)).To(Equal("Stream"))
    })
    
    It("should handle large streams", func() {
        // 1 MB of data
        largeData := make([]byte, 1024*1024)
        for i := range largeData {
            largeData[i] = byte(i % 256)
        }
        
        reader := bytes.NewReader(largeData)
        hexReader := coder.EncodeReader(reader)
        encoded, err := io.ReadAll(hexReader)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(len(encoded)).To(Equal(len(largeData) * 2))
    })
    
    It("should handle empty stream", func() {
        reader := bytes.NewReader([]byte{})
        hexReader := coder.EncodeReader(reader)
        encoded, err := io.ReadAll(hexReader)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(len(encoded)).To(Equal(0))
    })
})
```

### 5. Edge Case Tests

**Scenarios:**
- Nil input
- Zero-length input
- Very large input
- All zero bytes
- All 0xFF bytes
- Invalid UTF-8
- Special characters

**Example:**
```go
var _ = Describe("Edge Cases", func() {
    var coder Coder
    
    BeforeEach(func() {
        coder = New()
    })
    
    It("should handle nil input", func() {
        encoded := coder.Encode(nil)
        Expect(len(encoded)).To(Equal(0))
    })
    
    It("should handle zero bytes", func() {
        input := []byte{0x00, 0x00, 0x00}
        encoded := coder.Encode(input)
        
        Expect(string(encoded)).To(Equal("000000"))
    })
    
    It("should handle 0xFF bytes", func() {
        input := []byte{0xFF, 0xFF, 0xFF}
        encoded := coder.Encode(input)
        
        Expect(string(encoded)).To(Equal("ffffff"))
    })
    
    It("should handle large input", func() {
        // 10 MB
        largeData := make([]byte, 10*1024*1024)
        encoded := coder.Encode(largeData)
        
        Expect(len(encoded)).To(Equal(len(largeData) * 2))
    })
    
    It("should handle non-ASCII bytes", func() {
        input := []byte{0x80, 0x90, 0xA0, 0xF0}
        encoded := coder.Encode(input)
        decoded, err := coder.Decode(encoded)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(decoded).To(Equal(input))
    })
    
    It("should detect whitespace in hex", func() {
        hexWithSpace := []byte("48 65 6c 6c 6f")
        _, err := coder.Decode(hexWithSpace)
        
        Expect(err).To(HaveOccurred())
    })
})
```

### 6. Table-Driven Tests

**Example:**
```go
var _ = Describe("Encoding Table Tests", func() {
    var coder Coder
    
    BeforeEach(func() {
        coder = New()
    })
    
    DescribeTable("encoding various inputs",
        func(input []byte, expected string) {
            encoded := coder.Encode(input)
            Expect(string(encoded)).To(Equal(expected))
        },
        Entry("empty", []byte{}, ""),
        Entry("single byte", []byte{0x48}, "48"),
        Entry("Hello", []byte("Hello"), "48656c6c6f"),
        Entry("binary", []byte{0x00, 0xFF}, "00ff"),
        Entry("spaces", []byte("a b"), "612062"),
    )
    
    DescribeTable("decoding various inputs",
        func(hex string, expected []byte) {
            decoded, err := coder.Decode([]byte(hex))
            Expect(err).NotTo(HaveOccurred())
            Expect(decoded).To(Equal(expected))
        },
        Entry("empty", "", []byte{}),
        Entry("single byte", "48", []byte{0x48}),
        Entry("lowercase", "48656c6c6f", []byte("Hello")),
        Entry("uppercase", "48656C6C6F", []byte("Hello")),
        Entry("mixed case", "48656C6c6F", []byte("Hello")),
    )
})
```

---

## Best Practices

### 1. Use BeforeEach for Setup

```go
var _ = Describe("Test Suite", func() {
    var coder Coder
    
    BeforeEach(func() {
        coder = New()
    })
    
    It("test case", func() {
        // Use coder
    })
})
```

### 2. Test Round-Trips

```go
It("should preserve data through round-trip", func() {
    original := []byte("Test data")
    
    encoded := coder.Encode(original)
    decoded, err := coder.Decode(encoded)
    
    Expect(err).NotTo(HaveOccurred())
    Expect(decoded).To(Equal(original))
})
```

### 3. Test Error Conditions

```go
It("should handle errors gracefully", func() {
    invalidHex := []byte("zzzz")
    _, err := coder.Decode(invalidHex)
    
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(ContainSubstring("invalid"))
})
```

### 4. Test Size Relationships

```go
It("should verify size doubling", func() {
    input := []byte("test")
    encoded := coder.Encode(input)
    
    Expect(len(encoded)).To(Equal(len(input) * 2))
})

It("should verify size halving", func() {
    hexData := []byte("74657374")
    decoded, _ := coder.Decode(hexData)
    
    Expect(len(decoded)).To(Equal(len(hexData) / 2))
})
```

### 5. Test Case Insensitivity

```go
It("should accept both cases", func() {
    lower, _ := coder.Decode([]byte("48656c6c6f"))
    upper, _ := coder.Decode([]byte("48656C6C6F"))
    
    Expect(lower).To(Equal(upper))
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-hexa:
  runs-on: ubuntu-latest
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test Hexa Package
      run: |
        cd encoding/hexa
        go test -v -race -cover
    
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
```

### GitLab CI

```yaml
test-hexa:
  script:
    - cd encoding/hexa
    - go test -v -race -cover
  
  coverage: '/coverage: \d+\.\d+% of statements/'
```

### Coverage Reports

```bash
# Generate HTML coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# View in browser
open coverage.html
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Cover edge cases** (empty, invalid, large, special chars)
3. **Test round-trips** (encode → decode = original)
4. **Test case insensitivity** (uppercase, lowercase, mixed)
5. **Test streaming** (Reader/Writer interfaces)
6. **Update coverage** metrics
7. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    var coder Coder
    
    BeforeEach(func() {
        coder = New()
    })
    
    Describe("basic functionality", func() {
        It("should handle normal case", func() {
            result := operation(input)
            Expect(result).To(Equal(expected))
        })
    })
    
    Context("error conditions", func() {
        It("should handle invalid input", func() {
            _, err := operation(invalidInput)
            Expect(err).To(HaveOccurred())
        })
    })
    
    Context("edge cases", func() {
        It("should handle empty input", func() {
            result := operation([]byte{})
            Expect(len(result)).To(Equal(0))
        })
        
        It("should handle large input", func() {
            large := make([]byte, 1024*1024)
            result := operation(large)
            Expect(len(result)).To(BeNumerically(">", 0))
        })
    })
})
```

---

## Support

For issues or questions:

- **Test Failures**: Check output with `-v` flag
- **Coverage Gaps**: Run `go test -cover` to identify
- **Usage Examples**: Review test files for patterns
- **Feature Questions**: See [README.md](README.md)
- **Bug Reports**: [GitHub Issues](https://github.com/nabbar/golib/issues)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
