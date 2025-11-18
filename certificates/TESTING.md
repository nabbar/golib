# Testing Guide - Certificates Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://golang.org/)

Comprehensive testing documentation for the certificates package and subpackages.

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

The certificates package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) to achieve comprehensive test coverage across all subpackages.

### Test Philosophy

1. **Behavior-Driven**: Tests describe behavior in readable, hierarchical structures
2. **Security-Focused**: Verify secure defaults and TLS configuration correctness
3. **Format Coverage**: Test multiple encoding formats (JSON, YAML, TOML, CBOR)
4. **Parser Robustness**: Test parsing with valid and invalid inputs
5. **Maintainable**: Clear test structure with reusable patterns

### Coverage Scope

- ✅ TLS configuration management and generation
- ✅ Certificate parsing and validation
- ✅ CA certificate management (root and client)
- ✅ TLS version, cipher suite, and curve configuration
- ✅ Client authentication modes
- ✅ Multiple encoding formats (JSON, YAML, TOML, CBOR)
- ✅ Parser edge cases and error handling

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
ginkgo --focus="Parse" -r

# Skip tests matching pattern
ginkgo --skip="Encoding" -r
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
# Test only main package
cd /path/to/golib/certificates
go test .

# Test specific subpackage
cd auth
go test .
```

---

## Test Coverage

### Coverage Metrics

| Package | Coverage | Specs | Description |
|---------|----------|-------|-------------|
| `certificates` | ~70% | 15 | Main TLS configuration |
| `auth` | 73.0% | 12 | Client authentication modes |
| `ca` | 68.5% | 18 | CA certificate management |
| `certs` | 47.8% | 9 | Certificate pair management |
| `cipher` | 50.6% | 12 | Cipher suite configuration |
| `curves` | 50.5% | 9 | Elliptic curve configuration |
| `tlsversion` | 54.5% | 9 | TLS version management |
| **Total** | **~60%** | **84** | **All packages** |

### Coverage by Component

**Main Package** (`certificates_config_test.go` - 15 specs)
- TLS configuration creation and validation
- Certificate pair management
- Root CA and Client CA operations
- Version, cipher, and curve configuration
- TLS config generation

**auth Package** (`auth_test.go` - 12 specs)
- Parse client authentication modes from strings
- Parse from integer values
- Encoding/decoding (JSON, YAML, TOML, CBOR)
- String representation and validation
- TLS type conversion

**ca Package** (`ca_test.go` - 18 specs)
- Parse CA certificates from PEM strings
- Parse certificate chains
- File path handling
- Certificate pool operations
- Encoding/decoding in multiple formats
- Error handling for invalid certificates

**certs Package** (`certs_test.go` - 9 specs)
- Parse certificate pairs from PEM
- Parse from ConfigPair and ConfigChain
- Certificate validation
- TLS certificate generation
- Encoding/decoding

**cipher Package** (`cipher_test.go` - 12 specs)
- Parse cipher suites from strings
- List available cipher suites
- Validate cipher suite IDs
- String representation
- Encoding/decoding

**curves Package** (`curves_test.go` - 9 specs)
- Parse elliptic curves from strings
- List available curves
- Validate curve IDs
- String representation
- Encoding/decoding

**tlsversion Package** (`tlsversion_test.go` - 9 specs)
- Parse TLS versions from strings
- Parse from integer values
- List available versions
- String representation
- Encoding/decoding

---

## Test Organization

### File Structure

```
certificates/
├── certificates_suite_test.go       # Main test suite setup
├── certificates_config_test.go      # TLS configuration tests (15 specs)
└── Subpackages/
    ├── auth/
    │   ├── auth_suite_test.go      # Auth test suite setup
    │   └── auth_test.go            # Auth mode tests (12 specs)
    ├── ca/
    │   ├── ca_suite_test.go        # CA test suite setup
    │   └── ca_test.go              # CA certificate tests (18 specs)
    ├── certs/
    │   ├── certs_suite_test.go     # Certs test suite setup
    │   └── certs_test.go           # Certificate pair tests (9 specs)
    ├── cipher/
    │   ├── cipher_suite_test.go    # Cipher test suite setup
    │   └── cipher_test.go          # Cipher suite tests (12 specs)
    ├── curves/
    │   ├── curves_suite_test.go    # Curves test suite setup
    │   └── curves_test.go          # Elliptic curve tests (9 specs)
    └── tlsversion/
        ├── tlsversion_suite_test.go # Version test suite setup
        └── tlsversion_test.go       # TLS version tests (9 specs)
```

### Test Structure Pattern

```go
// Hierarchical BDD structure
Describe("certificates/Feature", func() {
    Context("When parsing input", func() {
        It("should parse valid PEM string", func() {
            // Arrange
            pemData := `-----BEGIN CERTIFICATE-----...`
            
            // Act
            cert, err := ca.Parse(pemData)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(cert).ToNot(BeNil())
            Expect(cert.Len()).To(Equal(1))
        })
        
        It("should return error for invalid input", func() {
            // Arrange
            invalid := "not a certificate"
            
            // Act
            _, err := ca.Parse(invalid)
            
            // Assert
            Expect(err).To(HaveOccurred())
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
It("should parse ECDHE-RSA-AES128-GCM-SHA256 cipher suite", func() {
    // Test implementation
})

// ❌ Bad: Vague description
It("should work", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should parse TLS version from string", func() {
    // Arrange
    input := "1.2"
    
    // Act
    version := tlsversion.Parse(input)
    
    // Assert
    Expect(version).To(Equal(tlsversion.VersionTLS12))
})
```

**3. Test Multiple Formats**
```go
// Test encoding/decoding for all supported formats
It("should marshal to JSON", func() {
    auth := auth.RequireAndVerifyClientCert
    data, err := json.Marshal(auth)
    Expect(err).ToNot(HaveOccurred())
    Expect(string(data)).To(ContainSubstring("require"))
})

It("should unmarshal from JSON", func() {
    var auth auth.ClientAuth
    err := json.Unmarshal([]byte(`"verify"`), &auth)
    Expect(err).ToNot(HaveOccurred())
    Expect(auth).To(Equal(auth.VerifyClientCertIfGiven))
})
```

**4. Test Edge Cases**
```go
// Test with empty, nil, invalid, and boundary inputs
It("should handle empty string", func() {
    version := tlsversion.Parse("")
    Expect(version).To(Equal(tlsversion.VersionUnknown))
})

It("should handle invalid cipher suite ID", func() {
    valid := cipher.Check(0xFFFF)
    Expect(valid).To(BeFalse())
})
```

**5. Test Security Properties**
```go
It("should default to secure TLS version", func() {
    cfg := certificates.New()
    min := cfg.GetVersionMin()
    Expect(min).To(Or(
        Equal(tlsversion.VersionTLS12),
        Equal(tlsversion.VersionTLS13),
    ))
})
```

### Test Template

```go
var _ = Describe("certificates/NewFeature", func() {
    Context("When feature is used", func() {
        It("should perform expected behavior", func() {
            // Arrange
            cfg := certificates.New()
            
            // Act
            cfg.SetVersionMin(tlsversion.VersionTLS12)
            version := cfg.GetVersionMin()
            
            // Assert
            Expect(version).To(Equal(tlsversion.VersionTLS12))
        })
        
        It("should handle edge case", func() {
            // Test edge case
            cfg := certificates.New()
            certs := cfg.GetCertificatePair()
            Expect(certs).To(BeEmpty())
        })
    })
})
```

---

## Best Practices

### Test Independence
```go
// ✅ Good: Each test creates its own config
It("should add certificate", func() {
    cfg := certificates.New()
    err := cfg.AddCertificatePairFile("key.pem", "cert.pem")
    Expect(err).ToNot(HaveOccurred())
})

// ❌ Bad: Tests share state
var sharedCfg certificates.TLSConfig

It("test1", func() {
    sharedCfg.SetVersionMin(tlsversion.VersionTLS12)
})

It("test2", func() {
    // Depends on test1!
    version := sharedCfg.GetVersionMin()
})
```

### Parser Testing
```go
// ✅ Good: Test various input formats
It("should parse flexible string formats", func() {
    inputs := []string{"1.2", "TLS 1.2", "tls_1_2", "tls12"}
    for _, input := range inputs {
        version := tlsversion.Parse(input)
        Expect(version).To(Equal(tlsversion.VersionTLS12))
    }
})
```

### Encoding Testing
```go
// ✅ Good: Test round-trip encoding
It("should round-trip through JSON", func() {
    original := cipher.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    
    // Marshal
    data, err := json.Marshal(original)
    Expect(err).ToNot(HaveOccurred())
    
    // Unmarshal
    var decoded cipher.Cipher
    err = json.Unmarshal(data, &decoded)
    Expect(err).ToNot(HaveOccurred())
    
    // Verify
    Expect(decoded).To(Equal(original))
})
```

### Security Testing
```go
// ✅ Good: Verify secure defaults
It("should use secure cipher suites by default", func() {
    cfg := certificates.New()
    ciphers := cfg.GetCiphers()
    
    for _, c := range ciphers {
        // No weak ciphers
        Expect(c).ToNot(Equal(cipher.Unknown))
    }
})
```

---

## Troubleshooting

### Common Issues

**File Not Found Errors**

*Problem*: Tests fail when loading certificate files.

*Solution*:
```go
// Use test fixtures or embedded test data
const testCertPEM = `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHHCgVZU...
-----END CERTIFICATE-----`

It("should parse certificate", func() {
    cert, err := ca.Parse(testCertPEM)
    Expect(err).ToNot(HaveOccurred())
})
```

**Encoding Mismatches**

*Problem*: Encoding tests fail due to format differences.

*Solution*:
```bash
# Verify encoding with actual output
ginkgo -v --focus="Encoding"
```

**Stale Coverage**

*Problem*: Coverage report doesn't reflect recent changes.

*Solution*:
```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

### Debugging Techniques

**Run Specific Tests**
```bash
# Focus on specific package
cd auth
go test -v

# Focus on specific test
ginkgo --focus="Parse" -v
```

**Verbose Output**
```bash
# Ginkgo verbose mode
ginkgo -v --trace -r

# Standard Go verbose
go test -v ./...
```

**Check Test Data**
```go
It("should parse certificate", func() {
    cert, err := ca.Parse(pemData)
    
    // Debug output
    fmt.Fprintf(GinkgoWriter, "Cert: %+v\n", cert)
    fmt.Fprintf(GinkgoWriter, "Error: %v\n", err)
    
    Expect(err).ToNot(HaveOccurred())
})
```

---

## Contributing

### Test Contributions

**Guidelines**
- Do not use AI to generate test implementation code
- AI may assist with test documentation and bug fixing
- Follow existing test patterns and structure
- Add tests for new features
- Test edge cases and error conditions

**Adding New Tests**
1. Choose appropriate test file based on feature
2. Use descriptive test names
3. Follow AAA pattern (Arrange, Act, Assert)
4. Test multiple input formats
5. Test edge cases and error conditions
6. Verify security properties

**Test Review Checklist**
- [ ] Tests are independent
- [ ] Resources are properly managed
- [ ] Edge cases are covered
- [ ] Multiple formats tested (JSON, YAML, TOML, CBOR)
- [ ] Descriptions are clear
- [ ] Security properties verified
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
- [Go crypto/tls](https://pkg.go.dev/crypto/tls)

**Certificates Package**
- [Package GoDoc](https://pkg.go.dev/github.com/nabbar/golib/certificates)
- [README.md](README.md) - Package overview and examples
- [GitHub Repository](https://github.com/nabbar/golib)

**Testing Tools**
- [Go Test Command](https://pkg.go.dev/cmd/go#hdr-Test_packages)
- [Coverage Tool](https://go.dev/blog/cover)

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.
