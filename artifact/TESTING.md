# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-45%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-Core%20100%25-brightgreen)]()

Comprehensive testing documentation for the artifact package, focusing on zero external dependencies and fast execution.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The artifact package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for testing with a focus on zero external dependencies.

**Test Suite**
- Total Specs: 45
- Core Coverage: 100% (artifact, client)
- Platform Coverage: 2-14% (intentional - no external API calls)
- Execution Time: <5s

**Testing Philosophy**
- **Zero External Dependencies**: No real API calls
- **No Billing**: All tests run without cloud service costs  
- **Fast Execution**: Complete suite in seconds
- **CI/CD Ready**: No credentials or external services required

**Coverage Focus**
- ✅ Version parsing and validation (100%)
- ✅ Regex matching (100%)
- ✅ Pre-release filtering (100%)
- ✅ Client helper logic (98.6%)
- ⚠️ Platform implementations (8-14% - core logic only)

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Using Ginkgo
ginkgo -r -cover
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Parallel execution support
- Rich CLI with filtering

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages

---

## Running Tests

### Basic Commands

```bash
# Standard test run
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Ginkgo CLI Options

```bash
# Run all tests recursively
ginkgo -r

# Specific package
cd client && ginkgo

# Pattern matching
ginkgo --focus="version validation"

# Parallel execution
ginkgo -p -r

# JUnit report
ginkgo --junit-report=results.xml -r
```

---

## Test Coverage

**Coverage By Package**

| Package | Coverage | Specs | Notes |
|---------|----------|-------|-------|
| `artifact` | 100% | 19 | Helper functions |
| `artifact/client` | 98.6% | 21 | Version organization |
| `artifact/github` | 8.6% | 1 | No external API calls |
| `artifact/gitlab` | 14.4% | 2 | No external API calls |
| `artifact/jfrog` | 6.8% | 2 | No external API calls |
| `artifact/s3aws` | 2.0% | 0 | No external API calls |

**View Coverage**

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

**Note**: Lower coverage in platform packages is intentional to avoid external API calls and billing in CI/CD.

### Test File Organization

| Package | File | Purpose |
|---------|------|---------|
| `artifact` | `helpers_test.go` | Helper functions |
| `client` | `helper_test.go` | Version organization |
| `client` | `client_helper_advanced_test.go` | Advanced scenarios |
| `github` | `github_test.go` | GitHub integration |
| `gitlab` | `gitlab_test.go` | GitLab integration |
| `jfrog` | `jfrog_test.go` | JFrog integration |
| `s3aws` | `s3aws_suite_test.go` | S3/AWS setup |

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should reject beta versions in prerelease validation", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should match artifact names using regex", func() {
    // Arrange
    name := "myapp-1.2.3.tar.gz"
    pattern := `myapp-\d+\.\d+\.\d+\.tar\.gz`
    
    // Act
    result := artifact.CheckRegex(name, pattern)
    
    // Assert
    Expect(result).To(BeTrue())
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(collection).To(HaveLen(5))
Expect(version.String()).To(Equal("1.2.3"))
```

**4. Avoid External Dependencies** - No real API calls or external services

**5. Test Edge Cases** - Empty input, nil values, invalid data

### Test Template

```go
var _ = Describe("artifact/new_feature", func() {
    Context("When using new feature", func() {
        var testVersion *version.Version

        BeforeEach(func() {
            testVersion, _ = version.NewVersion("1.2.3")
        })

        It("should perform expected behavior", func() {
            result, err := newFeature(testVersion)
            
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(expectedResult))
        })

        It("should handle error case", func() {
            _, err := newFeature(invalidInput)
            Expect(err).To(HaveOccurred())
        })
    })
})
```

---

## Best Practices

**Test Independence**
- Each test should be independent
- Use `BeforeEach`/`AfterEach` for setup/cleanup
- Avoid global mutable state

**Test Data**
- Create version collections locally
- Use realistic version strings ("1.2.3", "2.0.0-beta")
- Clean up resources in `AfterEach`

**Assertions**
- Use specific matchers for better error messages
- One behavior per test
- Keep tests focused

**Performance**
- Keep tests fast (use small data)
- Use parallel execution (`ginkgo -p`)
- Mock external dependencies

**Documentation**
- Comment complex scenarios
- Document assumptions
- Keep descriptions clear

**Zero External Dependencies**
- No real API calls
- No credentials in tests
- Mock platform responses
- Fast CI/CD execution

---

## Troubleshooting

**Version Parsing Failures**
```bash
# Problem: Version strings cannot be parsed
# Solution: Use semantic versioning (e.g., "1.2.3", not "v1.2.3")
```

**Import Errors**
```bash
# Problem: Cannot import test packages
# Solution: Ensure dependencies are installed
go mod tidy
go mod download
```

**Stale Coverage**
```bash
# Problem: Coverage doesn't reflect recent changes
# Solution: Clean cache and regenerate
go clean -testcache
go test -coverprofile=coverage.out ./...
```

**Debugging**
```bash
# Run single test
ginkgo --focus="should reject beta versions"

# Verbose output
ginkgo -v --trace

# Debug logging in tests
It("should do something", func() {
    fmt.Fprintf(GinkgoWriter, "Debug: version = %v\n", version)
})
```

---

## CI Integration

**GitHub Actions Example**
```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test -v ./...
      
      - name: Coverage
        run: go test -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

**Pre-commit Hook**
```bash
#!/bin/bash
go test ./... || exit 1
go test -cover ./... | grep -E "coverage:" || exit 1
```

**Quality Checklist**

Before merging:
- [ ] All tests pass: `go test ./...`
- [ ] Coverage maintained: Core packages at 100%
- [ ] No external dependencies added
- [ ] Edge cases covered
- [ ] Documentation updated

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)

**Versioning**
- [Semantic Versioning](https://semver.org/)
- [hashicorp/go-version](https://github.com/hashicorp/go-version)

**Support**
- [GitHub Issues](https://github.com/nabbar/golib/issues)
- [README](README.md)

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Artifact Package Contributors
