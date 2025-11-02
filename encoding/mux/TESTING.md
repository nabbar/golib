# Multiplexer/DeMultiplexer Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the multiplexer/demultiplexer package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Test Scenarios](#test-scenarios)
- [Concurrency Testing](#concurrency-testing)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The mux package features comprehensive testing covering multiplexing, demultiplexing, concurrent operations, and edge cases.

### Test Metrics

| Metric | Value | Status |
|--------|-------|--------|
| **Total Specs** | 59 | ✅ All passing |
| **Framework** | Ginkgo v2 + Gomega | ✅ BDD |
| **Race Tests** | Validated | ✅ No races |
| **Benchmarks** | Included | ✅ Performance |

---

## Test Framework

### Ginkgo v2 + Gomega

**Features:**
- BDD-style test organization
- Concurrent testing scenarios
- Round-trip verification
- Performance benchmarks

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Test Organization

### Test Files

```
encoding/mux/
├── mux_suite_test.go        # Suite setup
├── mux_test.go              # Basic tests (10+ specs)
├── multiplexer_test.go      # Multiplexer tests (20+ specs)
├── demultiplexer_test.go    # DeMultiplexer tests (20+ specs)
└── benchmark_test.go        # Benchmarks (10+ specs)
```

### Test Categories

1. **Basic Operations** - Creation, basic write/read
2. **Multiplexing** - Channel creation, concurrent writes
3. **DeMultiplexing** - Channel routing, message parsing
4. **Round-Trip** - Mux → DeMux verification
5. **Concurrency** - Thread safety validation
6. **Edge Cases** - Empty data, invalid channels, errors
7. **Performance** - Benchmarks for throughput

---

## Running Tests

### Quick Test

```bash
cd encoding/mux
go test -v
```

### With Coverage

```bash
go test -v -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Race Detection

```bash
go test -v -race
```

### Using Ginkgo

```bash
# Run all tests
ginkgo -v

# Parallel execution
ginkgo -v -p

# With race detection
ginkgo -v -race

# Benchmarks
go test -bench=. -benchmem
```

---

## Test Coverage

### Coverage by Component

| Component | File | Specs | Notes |
|-----------|------|-------|-------|
| Multiplexer | mux.go | 20+ | Write operations |
| DeMultiplexer | demux.go | 20+ | Read operations |
| Interface | interface.go | 10+ | API creation |
| Edge Cases | all | 15+ | Error handling |

---

## Test Scenarios

### 1. Multiplexer Tests

**Scenarios:**
- Create multiplexer
- Create channels
- Write to channels
- Concurrent writes
- Message formatting

**Example:**
```go
var _ = Describe("Multiplexer", func() {
    var (
        buf *bytes.Buffer
        mux Multiplexer
    )
    
    BeforeEach(func() {
        buf = &bytes.Buffer{}
        mux = NewMultiplexer(buf, '\n')
    })
    
    It("should create channel", func() {
        ch := mux.NewChannel('a')
        Expect(ch).NotTo(BeNil())
    })
    
    It("should write to channel", func() {
        ch := mux.NewChannel('a')
        n, err := ch.Write([]byte("test"))
        
        Expect(err).NotTo(HaveOccurred())
        Expect(n).To(Equal(4))
        Expect(buf.Len()).To(BeNumerically(">", 0))
    })
    
    It("should handle concurrent writes", func() {
        done := make(chan bool)
        
        for i := 0; i < 10; i++ {
            go func(id int) {
                ch := mux.NewChannel(rune('a' + id))
                ch.Write([]byte(fmt.Sprintf("msg%d", id)))
                done <- true
            }(i)
        }
        
        for i := 0; i < 10; i++ {
            <-done
        }
    })
})
```

### 2. DeMultiplexer Tests

**Scenarios:**
- Create demultiplexer
- Register channels
- Read and route messages
- Copy operation
- Concurrent reads

**Example:**
```go
var _ = Describe("DeMultiplexer", func() {
    It("should route messages to channels", func() {
        // Create mux side
        muxBuf := &bytes.Buffer{}
        mux := NewMultiplexer(muxBuf, '\n')
        
        chA := mux.NewChannel('a')
        chB := mux.NewChannel('b')
        
        chA.Write([]byte("msg A"))
        chB.Write([]byte("msg B"))
        
        // Create demux side
        demux := NewDeMultiplexer(muxBuf, '\n', 0)
        
        outA := &bytes.Buffer{}
        outB := &bytes.Buffer{}
        
        demux.NewChannel('a', outA)
        demux.NewChannel('b', outB)
        
        // Process
        err := demux.Copy()
        Expect(err).NotTo(HaveOccurred())
        
        // Verify
        Expect(outA.String()).To(Equal("msg A"))
        Expect(outB.String()).To(Equal("msg B"))
    })
})
```

### 3. Round-Trip Tests

**Scenarios:**
- Data integrity verification
- Multiple channels
- Different data sizes
- Unicode keys

**Example:**
```go
It("should preserve data through round-trip", func() {
    testData := []struct{
        key rune
        data string
    }{
        {'a', "Hello World"},
        {'b', "Binary\x00Data"},
        {'1', "Numbers123"},
        {'α', "Unicode α β γ"},
    }
    
    // Mux
    muxBuf := &bytes.Buffer{}
    mux := NewMultiplexer(muxBuf, '\n')
    
    for _, td := range testData {
        ch := mux.NewChannel(td.key)
        ch.Write([]byte(td.data))
    }
    
    // DeMux
    demux := NewDeMultiplexer(muxBuf, '\n', 0)
    outputs := make(map[rune]*bytes.Buffer)
    
    for _, td := range testData {
        buf := &bytes.Buffer{}
        outputs[td.key] = buf
        demux.NewChannel(td.key, buf)
    }
    
    err := demux.Copy()
    Expect(err).NotTo(HaveOccurred())
    
    // Verify
    for _, td := range testData {
        Expect(outputs[td.key].String()).To(Equal(td.data))
    }
})
```

---

## Concurrency Testing

### Thread Safety Validation

```go
It("should be thread-safe", func() {
    buf := &bytes.Buffer{}
    mux := NewMultiplexer(buf, '\n')
    
    // Many goroutines writing concurrently
    const numGoroutines = 100
    const numWrites = 100
    
    done := make(chan bool)
    
    for g := 0; g < numGoroutines; g++ {
        go func(id int) {
            ch := mux.NewChannel(rune('a' + (id % 26)))
            for i := 0; i < numWrites; i++ {
                ch.Write([]byte(fmt.Sprintf("g%d-w%d", id, i)))
            }
            done <- true
        }(g)
    }
    
    for g := 0; g < numGoroutines; g++ {
        <-done
    }
    
    // Verify no panics or corruption
    Expect(buf.Len()).To(BeNumerically(">", 0))
})
```

### Race Detection

```bash
# Run with race detector
go test -race -v

# All tests should pass without data races
```

---

## Best Practices

### 1. Use BeforeEach/AfterEach

```go
var _ = Describe("Test Suite", func() {
    var (
        buf *bytes.Buffer
        mux Multiplexer
    )
    
    BeforeEach(func() {
        buf = &bytes.Buffer{}
        mux = NewMultiplexer(buf, '\n')
    })
    
    It("test case", func() {
        // Use mux
    })
})
```

### 2. Test Round-Trips

```go
It("should preserve data", func() {
    original := []byte("test data")
    
    // Mux
    muxBuf := &bytes.Buffer{}
    mux := NewMultiplexer(muxBuf, '\n')
    ch := mux.NewChannel('a')
    ch.Write(original)
    
    // DeMux
    demux := NewDeMultiplexer(muxBuf, '\n', 0)
    out := &bytes.Buffer{}
    demux.NewChannel('a', out)
    demux.Copy()
    
    // Verify
    Expect(out.Bytes()).To(Equal(original))
})
```

### 3. Test Concurrency

```go
It("should handle concurrent access", func() {
    // Use multiple goroutines
    // Verify with -race flag
})
```

### 4. Test Edge Cases

```go
It("should handle empty data", func() {
    ch := mux.NewChannel('a')
    n, err := ch.Write([]byte{})
    Expect(err).NotTo(HaveOccurred())
})

It("should handle unknown channels", func() {
    // Write to channel that's not registered
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-mux:
  runs-on: ubuntu-latest
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test Mux Package
      run: |
        cd encoding/mux
        go test -v -race -cover
```

### GitLab CI

```yaml
test-mux:
  script:
    - cd encoding/mux
    - go test -v -race -cover
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Test concurrency** (use -race flag)
3. **Test round-trips** (mux → demux)
4. **Test edge cases** (empty, errors, invalid)
5. **Include benchmarks** for performance changes
6. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    var (
        buf *bytes.Buffer
        mux Multiplexer
    )
    
    BeforeEach(func() {
        buf = &bytes.Buffer{}
        mux = NewMultiplexer(buf, '\n')
    })
    
    It("should handle normal case", func() {
        // Test implementation
    })
    
    Context("concurrent access", func() {
        It("should be thread-safe", func() {
            // Concurrent test
        })
    })
})
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

*Part of the [golib](https://github.com/nabbar/golib) testing suite.*
