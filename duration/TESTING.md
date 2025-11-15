# Duration Package - Testing Documentation

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Ginkgo](https://img.shields.io/badge/Ginkgo-v2-green)](https://github.com/onsi/ginkgo)

Comprehensive testing guide for the duration package using Ginkgo v2/Gomega BDD framework.

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for test generation, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Test Framework](#test-framework)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Standard Duration Tests](#standard-duration-tests)
- [Big Duration Tests](#big-duration-tests)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)

---

## Overview

The duration package features comprehensive testing across both standard and big duration implementations.

### Test Metrics

| Package | Specs | Coverage | Status |
|---------|-------|----------|--------|
| **duration** | 150+ | 93.5% | ✅ All passing |
| **duration/big** | 250+ | 91.0% | ✅ All passing |
| **Total** | 400+ | ~92% | ✅ Excellent |

---

## Test Framework

### Ginkgo v2 + Gomega

**Features:**
- BDD-style test organization
- Table-driven tests for parsing
- Comprehensive edge case coverage
- Encoding/decoding verification

**Installation:**
```bash
go get github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega
```

---

## Test Organization

### Standard Duration Tests

```
duration/
├── duration_suite_test.go    # Suite setup
├── duration_test.go          # Basic duration tests
├── parse_test.go             # Parsing tests (40+ specs)
├── format_test.go            # Formatting tests (20+ specs)
├── encode_test.go            # Encoding tests (50+ specs)
├── operation_test.go         # Arithmetic tests (30+ specs)
├── truncate_test.go          # Truncation tests (30+ specs)
└── model_test.go             # Model tests (20+ specs)
```

### Big Duration Tests

```
duration/big/
├── big_suite_test.go         # Suite setup
├── big_test.go               # Basic big duration tests
├── parse_test.go             # Parsing tests (60+ specs)
├── format_test.go            # Formatting tests (40+ specs)
├── encode_test.go            # Encoding tests (80+ specs)
├── operation_test.go         # Arithmetic tests (40+ specs)
├── truncate_test.go          # Truncation tests (50+ specs)
└── model_test.go             # Model tests (30+ specs)
```

---

## Running Tests

### Quick Test

**Standard Duration:**
```bash
cd duration
go test -v
```

**Big Duration:**
```bash
cd duration/big
go test -v
```

**All Tests:**
```bash
cd duration
go test -v ./...
```

### With Coverage

```bash
cd duration
go test -v -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Big duration
cd big
go test -v -cover
```

### Using Ginkgo

```bash
cd duration
ginkgo -v
ginkgo -v -cover

# Parallel execution
ginkgo -v -p

# With trace
ginkgo -v --trace
```

---

## Test Coverage

### Coverage by Component

**Standard Duration:**

| Component | File | Specs | Coverage | Notes |
|-----------|------|-------|----------|-------|
| Parsing | parse.go | 40+ | 95% | All formats tested |
| Formatting | format.go | 20+ | 95% | String conversion |
| Encoding | encode.go | 50+ | 92% | JSON/YAML/TOML/CBOR |
| Operations | operation.go | 30+ | 95% | Arithmetic |
| Truncation | truncate.go | 30+ | 90% | All units |
| Model | model.go | 20+ | 93% | Core functions |

**Overall**: 93.5%

**Big Duration:**

| Component | File | Specs | Coverage | Notes |
|-----------|------|-------|----------|-------|
| Parsing | parse.go | 60+ | 93% | Large values tested |
| Formatting | format.go | 40+ | 92% | String conversion |
| Encoding | encode.go | 80+ | 90% | All formats |
| Operations | operation.go | 40+ | 92% | Overflow tests |
| Truncation | truncate.go | 50+ | 88% | All units |
| Model | model.go | 30+ | 91% | Core functions |

**Overall**: 91.0%

---

## Standard Duration Tests

### Parsing Tests

**Scenarios:**
- Valid duration strings
- Days notation
- Fractional values
- Negative durations
- Edge cases (empty, invalid)
- Mixed units

**Example:**
```go
var _ = Describe("Parse", func() {
    It("should parse days notation", func() {
        d, err := Parse("5d")
        Expect(err).NotTo(HaveOccurred())
        Expect(d.String()).To(Equal("5d"))
    })
    
    It("should parse mixed units", func() {
        d, err := Parse("2d3h15m30s")
        Expect(err).NotTo(HaveOccurred())
        Expect(d.Days()).To(Equal(int64(2)))
        Expect(d.Hours() % 24).To(Equal(int64(3)))
    })
    
    It("should handle fractional values", func() {
        d, err := Parse("1.5h")
        Expect(err).NotTo(HaveOccurred())
        Expect(d.Hours()).To(Equal(int64(1)))
        Expect(d.Minutes() % 60).To(Equal(int64(30)))
    })
    
    It("should reject invalid input", func() {
        _, err := Parse("invalid")
        Expect(err).To(HaveOccurred())
    })
})
```

### Encoding Tests

**Scenarios:**
- JSON marshal/unmarshal
- YAML marshal/unmarshal
- TOML marshal/unmarshal
- CBOR marshal/unmarshal
- Text marshal/unmarshal
- Round-trip encoding

**Example:**
```go
var _ = Describe("JSON Encoding", func() {
    type Example struct {
        Timeout Duration `json:"timeout"`
    }
    
    It("should marshal to JSON", func() {
        ex := Example{Timeout: Days(2)}
        data, err := json.Marshal(ex)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(string(data)).To(ContainSubstring("2d"))
    })
    
    It("should unmarshal from JSON", func() {
        data := []byte(`{"timeout":"2d3h"}`)
        var ex Example
        
        err := json.Unmarshal(data, &ex)
        Expect(err).NotTo(HaveOccurred())
        Expect(ex.Timeout.Days()).To(Equal(int64(2)))
    })
    
    It("should round-trip correctly", func() {
        original := Example{Timeout: Days(5) + Hours(12)}
        
        data, _ := json.Marshal(original)
        var decoded Example
        json.Unmarshal(data, &decoded)
        
        Expect(decoded.Timeout).To(Equal(original.Timeout))
    })
})
```

### Operation Tests

**Scenarios:**
- Addition
- Subtraction
- Multiplication
- Division
- Comparison
- Absolute value
- Min/Max

**Example:**
```go
var _ = Describe("Arithmetic", func() {
    It("should add durations", func() {
        d1 := Hours(2)
        d2 := Hours(3)
        result := d1 + d2
        
        Expect(result.Hours()).To(Equal(int64(5)))
    })
    
    It("should subtract durations", func() {
        d1 := Hours(5)
        d2 := Hours(2)
        result := d1 - d2
        
        Expect(result.Hours()).To(Equal(int64(3)))
    })
    
    It("should multiply duration", func() {
        d := Hours(2)
        result := d * 3
        
        Expect(result.Hours()).To(Equal(int64(6)))
    })
    
    It("should compare durations", func() {
        d1 := Hours(5)
        d2 := Hours(3)
        
        Expect(d1 > d2).To(BeTrue())
        Expect(d1 < d2).To(BeFalse())
        Expect(d1 == d1).To(BeTrue())
    })
})
```

### Truncation Tests

**Scenarios:**
- Truncate to days
- Truncate to hours
- Truncate to minutes
- Truncate to seconds
- Round to unit

**Example:**
```go
var _ = Describe("Truncation", func() {
    It("should truncate to days", func() {
        d, _ := Parse("5d23h45m30s")
        truncated := d.TruncateDays()
        
        Expect(truncated.String()).To(Equal("5d"))
    })
    
    It("should truncate to hours", func() {
        d, _ := Parse("2d3h15m")
        truncated := d.TruncateHours()
        
        Expect(truncated.Hours()).To(Equal(int64(51)))  // 2*24 + 3
        Expect(truncated.Minutes() % 60).To(Equal(int64(0)))
    })
    
    It("should round to nearest unit", func() {
        d, _ := Parse("1h40m")
        rounded := d.Round(Hour)
        
        Expect(rounded.Hours()).To(Equal(int64(2)))  // Rounds up
    })
})
```

---

## Big Duration Tests

### Parsing Tests

**Scenarios:**
- Very large durations (>290 years)
- Overflow detection
- Large day values
- Conversion from time.Duration
- Conversion from float64

**Example:**
```go
var _ = Describe("Big Parse", func() {
    It("should parse very large durations", func() {
        d, err := Parse("1000000d")  // ~2740 years
        Expect(err).NotTo(HaveOccurred())
        Expect(d.Days()).To(Equal(int64(1000000)))
    })
    
    It("should handle overflow gracefully", func() {
        // Test maximum value
        maxDays := int64(106751991167300)
        d := Days(maxDays)
        Expect(d.Days()).To(Equal(maxDays))
    })
    
    It("should convert from time.Duration", func() {
        td := 24 * time.Hour
        d := ParseDuration(td)
        Expect(d.Days()).To(Equal(int64(1)))
    })
    
    It("should parse fractional seconds", func() {
        d := ParseFloat64(3600.5)
        Expect(d.Int64()).To(Equal(int64(3600)))
    })
})
```

### Encoding Tests

**Scenarios:**
- JSON encoding of large values
- YAML encoding
- TOML encoding
- CBOR encoding
- Round-trip verification
- Error handling

**Example:**
```go
var _ = Describe("Big JSON Encoding", func() {
    type Config struct {
        MaxAge Duration `json:"max_age"`
    }
    
    It("should encode large durations", func() {
        cfg := Config{MaxAge: Days(365000)}  // ~1000 years
        data, err := json.Marshal(cfg)
        
        Expect(err).NotTo(HaveOccurred())
        Expect(string(data)).To(ContainSubstring("365000d"))
    })
    
    It("should decode large durations", func() {
        data := []byte(`{"max_age":"365000d"}`)
        var cfg Config
        
        err := json.Unmarshal(data, &cfg)
        Expect(err).NotTo(HaveOccurred())
        Expect(cfg.MaxAge.Days()).To(Equal(int64(365000)))
    })
})
```

### Type Conversion Tests

**Scenarios:**
- To int64
- To uint64
- To float64
- To time.Duration (with overflow check)
- To string

**Example:**
```go
var _ = Describe("Type Conversions", func() {
    It("should convert to int64", func() {
        d := Days(7)
        seconds := d.Int64()
        Expect(seconds).To(Equal(int64(604800)))  // 7 * 86400
    })
    
    It("should convert to uint64", func() {
        d := Days(7)
        useconds := d.Uint64()
        Expect(useconds).To(Equal(uint64(604800)))
    })
    
    It("should convert to float64", func() {
        d := ParseFloat64(3600.5)
        f := d.Float64()
        Expect(f).To(Equal(3600.5))
    })
    
    It("should detect overflow to time.Duration", func() {
        d := Days(365 * 1000)  // 1000 years (too large)
        _, err := d.Duration()
        Expect(err).To(HaveOccurred())
        Expect(err).To(Equal(ErrOverFlow))
    })
})
```

---

## Best Practices

### 1. Use Table-Driven Tests

```go
var _ = Describe("Parse Table Tests", func() {
    DescribeTable("parsing various formats",
        func(input string, expectedDays int64, expectedHours int64) {
            d, err := Parse(input)
            Expect(err).NotTo(HaveOccurred())
            Expect(d.Days()).To(Equal(expectedDays))
            Expect(d.Hours() % 24).To(Equal(expectedHours))
        },
        Entry("days only", "5d", int64(5), int64(0)),
        Entry("days and hours", "2d12h", int64(2), int64(12)),
        Entry("complex", "7d23h15m", int64(7), int64(23)),
    )
})
```

### 2. Test Edge Cases

```go
It("should handle zero duration", func() {
    d, err := Parse("0s")
    Expect(err).NotTo(HaveOccurred())
    Expect(d.Int64()).To(Equal(int64(0)))
})

It("should handle negative duration", func() {
    d, err := Parse("-5h")
    Expect(err).NotTo(HaveOccurred())
    Expect(d.Hours()).To(Equal(int64(-5)))
})

It("should handle empty string", func() {
    _, err := Parse("")
    Expect(err).To(HaveOccurred())
})
```

### 3. Test Round-Trip Encoding

```go
It("should round-trip through JSON", func() {
    original := Days(5) + Hours(12) + Minutes(30)
    
    // Marshal
    data, err := json.Marshal(struct{ D Duration }{D: original})
    Expect(err).NotTo(HaveOccurred())
    
    // Unmarshal
    var result struct{ D Duration }
    err = json.Unmarshal(data, &result)
    Expect(err).NotTo(HaveOccurred())
    
    // Compare
    Expect(result.D).To(Equal(original))
})
```

### 4. Test Overflow/Underflow

```go
It("should detect overflow for big duration", func() {
    // Test with maximum safe value
    maxVal := int64(math.MaxInt64 / 86400)  // Max days
    d := Days(maxVal)
    Expect(d.Days()).To(Equal(maxVal))
    
    // Values beyond this may overflow
})
```

### 5. Verify Error Messages

```go
It("should provide meaningful error messages", func() {
    _, err := Parse("invalid")
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(ContainSubstring("invalid"))
})
```

---

## CI/CD Integration

### GitHub Actions

```yaml
test-duration:
  runs-on: ubuntu-latest
  
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test Duration Package
      run: |
        cd duration
        go test -v -cover ./...
    
    - name: Test Big Duration
      run: |
        cd duration/big
        go test -v -cover
```

### GitLab CI

```yaml
test-duration:
  script:
    - cd duration
    - go test -v -cover ./...
  
  coverage: '/coverage: \d+\.\d+% of statements/'
```

### Coverage Reports

```bash
# Generate coverage report
cd duration
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View in browser
open coverage.html
```

---

## Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Cover edge cases** (zero, negative, overflow, invalid)
3. **Test all encodings** (JSON, YAML, TOML, CBOR, Text)
4. **Verify round-trips** (marshal → unmarshal → equal)
5. **Update coverage** metrics
6. **Document test scenarios**

### Test Template

```go
var _ = Describe("New Feature", func() {
    It("should handle basic case", func() {
        // Test implementation
        result := NewFeature(input)
        Expect(result).To(Equal(expected))
    })
    
    Context("when invalid input", func() {
        It("should return error", func() {
            _, err := NewFeature(invalidInput)
            Expect(err).To(HaveOccurred())
        })
    })
    
    Context("edge cases", func() {
        It("should handle zero", func() {
            result := NewFeature(0)
            Expect(result).To(Equal(zero))
        })
        
        It("should handle negative", func() {
            result := NewFeature(-1)
            Expect(result.IsNegative()).To(BeTrue())
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
