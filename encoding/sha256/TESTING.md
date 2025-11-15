# SHA-256 Encoding Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the SHA-256 encoding package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Scenarios](#test-scenarios)
- [Test Vectors](#test-vectors)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The sha256 package features comprehensive testing covering hashing operations, streaming, and verification against known test vectors.

---

## Test Framework

### Ginkgo v2 + Gomega

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Test Organization

### Test Files

```
encoding/sha256/
├── sha256_suite_test.go  # Suite setup
├── sha256_test.go        # Basic tests
├── hash_test.go          # Hash verification
├── reader_test.go        # Streaming read
└── writer_test.go        # Streaming write
```

---

## Running Tests

### Quick Test

```bash
cd encoding/sha256
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
ginkgo -v
ginkgo -v -cover
```

---

## Test Scenarios

### 1. Basic Hashing

**Example:**
```go
var _ = Describe("SHA256", func() {
    var hasher Coder
    
    BeforeEach(func() {
        hasher = New()
    })
    
    It("should hash data", func() {
        data := []byte("Hello")
        hash := hasher.Encode(data)
        
        Expect(hash).To(HaveLen(64))  // Hex-encoded
        Expect(hash).NotTo(BeEmpty())
    })
    
    It("should be deterministic", func() {
        data := []byte("test")
        hash1 := hasher.Encode(data)
        hasher.Reset()
        hash2 := hasher.Encode(data)
        
        Expect(hash1).To(Equal(hash2))
    })
})
```

### 2. Known Test Vectors

**Example:**
```go
It("should match known SHA-256 hashes", func() {
    testVectors := []struct{
        input string
        expected string
    }{
        {
            input: "",
            expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
        },
        {
            input: "abc",
            expected: "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
        },
        {
            input: "Hello, World!",
            expected: "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
        },
    }
    
    hasher := New()
    for _, tv := range testVectors {
        hash := hasher.Encode([]byte(tv.input))
        Expect(string(hash)).To(Equal(tv.expected))
        hasher.Reset()
    }
})
```

### 3. Streaming Operations

**Example:**
```go
It("should hash via reader", func() {
    data := []byte("Stream test data")
    reader := bytes.NewReader(data)
    
    hasher := New()
    hashReader := hasher.EncodeReader(reader)
    
    // Read all data
    _, err := io.Copy(io.Discard, hashReader)
    Expect(err).NotTo(HaveOccurred())
    
    // Verify hash was computed
    hash := hasher.Encode(nil)
    Expect(hash).To(HaveLen(64))
})
```

### 4. Reset Behavior

**Example:**
```go
It("should reset state", func() {
    hasher := New()
    
    hash1 := hasher.Encode([]byte("first"))
    hasher.Reset()
    hash2 := hasher.Encode([]byte("second"))
    
    Expect(hash1).NotTo(Equal(hash2))
})

It("should not carry state without reset", func() {
    hasher := New()
    
    // Hash without reset between
    hash1 := hasher.Encode([]byte("data1"))
    hash2 := hasher.Encode([]byte("data2"))
    
    // hash2 includes data1 state
    Expect(hash1).NotTo(Equal(hash2))
})
```

---

## Test Vectors

### Standard Test Vectors

From NIST FIPS 180-4:

| Input | SHA-256 Hash |
|-------|--------------|
| "" | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| "abc" | ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad |
| "abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq" | 248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1 |

### Custom Test Vectors

```go
var customVectors = []struct{
    name string
    input []byte
    expected string
}{
    {
        name: "empty",
        input: []byte{},
        expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    },
    {
        name: "single byte",
        input: []byte{0x00},
        expected: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
    },
    {
        name: "binary data",
        input: []byte{0xFF, 0xFF, 0xFF},
        expected: "7d38b5cd25a2baf85ad3bb5b9311383e671a8a142eb302b324d4a5fba8748c69",
    },
}
```

---

## Best Practices

### 1. Use Test Vectors

```go
It("should verify against NIST test vectors", func() {
    // Use official test vectors
    hasher := New()
    hash := hasher.Encode([]byte("abc"))
    
    expected := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
    Expect(string(hash)).To(Equal(expected))
})
```

### 2. Test Reset

```go
It("should reset properly", func() {
    hasher := New()
    
    hasher.Encode([]byte("data"))
    hasher.Reset()
    hash := hasher.Encode([]byte("data"))
    
    // Should match fresh hash
    freshHash := New().Encode([]byte("data"))
    Expect(hash).To(Equal(freshHash))
})
```

### 3. Test Edge Cases

```go
It("should handle empty input", func() {
    hash := hasher.Encode([]byte{})
    Expect(hash).To(HaveLen(64))
})

It("should handle large input", func() {
    large := make([]byte, 1024*1024)
    hash := hasher.Encode(large)
    Expect(hash).To(HaveLen(64))
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-sha256:
  runs-on: ubuntu-latest
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test SHA256 Package
      run: |
        cd encoding/sha256
        go test -v -cover
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD)
2. **Use test vectors** from standards
3. **Test edge cases** (empty, large, binary)
4. **Test streaming** operations
5. **Document test scenarios**

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
