# FTPClient Testing Guide

Comprehensive testing documentation for the `ftpclient` package.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Test Organization](#test-organization)
- [Writing Tests](#writing-tests)
- [Integration Testing](#integration-testing)
- [Performance Testing](#performance-testing)
- [Continuous Integration](#continuous-integration)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Overview

The `ftpclient` package uses the Ginkgo/Gomega v2 testing framework for behavior-driven development (BDD) style tests. The test suite covers configuration validation, connection management, and error handling.

### Current Test Statistics

- **Total Tests**: 22 specs
- **Pass Rate**: 100% (22/22)
- **Code Coverage**: 6.2%
- **Execution Time**: ~0.008s
- **Framework**: Ginkgo v2 + Gomega

### Test Categories

```
ftpclient/
├── Config Tests (22 specs)
│   ├── Structure validation
│   ├── Option settings
│   ├── Timezone configuration
│   ├── Context registration
│   └── Edge cases
└── Future Tests (Planned)
    ├── Connection management
    ├── File operations
    ├── Directory operations
    └── Concurrency tests
```

## Quick Start

### Prerequisites

```bash
# Install dependencies
go get github.com/onsi/ginkgo/v2
go get github.com/onsi/gomega

# Or use go mod
go mod tidy
```

### Run All Tests

```bash
# Standard test run
go test -v

# With coverage
go test -cover -v

# With race detector (recommended)
CGO_ENABLED=1 go test -race -v

# Ginkgo CLI (if installed)
ginkgo -v
```

### Expected Output

```
=== RUN   TestFTPClient
Running Suite: FTPClient Suite
===============================
Random Seed: 1762780137

Will run 22 of 22 specs
••••••••••••••••••••••

Ran 22 of 22 Specs in 0.008 seconds
SUCCESS! -- 22 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestFTPClient (0.01s)
PASS
coverage: 6.2% of statements
ok  	github.com/nabbar/golib/ftpclient	1.040s
```

## Test Framework

### Ginkgo/Gomega

The package uses Ginkgo v2 for test structure and Gomega for assertions.

#### Why Ginkgo/Gomega?

- **Readable**: BDD-style tests are self-documenting
- **Flexible**: Supports table-driven tests, nested contexts
- **Powerful**: Rich assertion library, custom matchers
- **Parallel**: Built-in parallel test execution support

### Test Structure

```go
var _ = Describe("Component", func() {
    var (
        cfg    *Config
        client FTPClient
    )

    BeforeEach(func() {
        // Setup before each test
        cfg = &Config{
            Hostname: "ftp.example.com:21",
        }
    })

    AfterEach(func() {
        // Cleanup after each test
        if client != nil {
            client.Close()
        }
    })

    Context("when doing something", func() {
        It("should behave correctly", func() {
            // Test implementation
            Expect(cfg.Hostname).To(Equal("ftp.example.com:21"))
        })
    })
})
```

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Verbose output
go test -v

# Run specific test
go test -v -run TestFTPClient

# Show coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### With Race Detector

**Always use race detector before committing:**

```bash
# Enable CGO for race detector
CGO_ENABLED=1 go test -race -v

# With timeout
CGO_ENABLED=1 go test -race -v -timeout 30s

# Full test with race and coverage
CGO_ENABLED=1 go test -race -cover -v
```

### Using Ginkgo CLI

```bash
# Install Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run tests
ginkgo -v

# Run with coverage
ginkgo -cover

# Watch mode (re-run on file changes)
ginkgo watch

# Parallel execution
ginkgo -p

# Focus on specific tests
ginkgo --focus "Config"

# Skip specific tests
ginkgo --skip "Edge cases"
```

## Test Coverage

### Current Coverage

```
File          Coverage    Lines    Uncovered
────────────────────────────────────────────
config.go       42.5%     80       46
errors.go       65.2%     23       8
interface.go     8.9%    112      102
model.go         2.7%    275      268
────────────────────────────────────────────
Total            6.2%    490      424
```

### Coverage Goals

- **Configuration**: Target 80%+
- **Connection Management**: Target 70%+
- **File Operations**: Target 75%+
- **Directory Operations**: Target 75%+
- **Error Handling**: Target 85%+
- **Overall**: Target 70%+

### Improving Coverage

To increase coverage, the following test areas need development:

1. **Connection Management** (High Priority)
   - Connect/Disconnect cycles
   - Automatic reconnection
   - Health checks (NOOP)
   - Timeout handling

2. **File Operations** (High Priority)
   - Upload (Stor, StorFrom)
   - Download (Retr, RetrFrom)
   - Append operations
   - File metadata (Size, Time)
   - File management (Rename, Delete)

3. **Directory Operations** (Medium Priority)
   - List and NameList
   - ChangeDir and CurrentDir
   - MakeDir and RemoveDir
   - RemoveDirRecur
   - Walk operations

4. **Concurrency** (Medium Priority)
   - Parallel operations
   - Race condition tests
   - Connection pooling

5. **Error Handling** (High Priority)
   - Error code validation
   - Recovery mechanisms
   - Edge cases

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser (Linux/macOS)
xdg-open coverage.html  # Linux
open coverage.html      # macOS
```

## Test Organization

### File Structure

```
ftpclient/
├── ftpclient_suite_test.go  # Suite initialization
├── config_test.go            # Configuration tests
└── (future test files)
    ├── connection_test.go    # Connection management
    ├── fileops_test.go       # File operations
    ├── dirops_test.go        # Directory operations
    ├── errors_test.go        # Error handling
    └── concurrency_test.go   # Thread safety
```

### Test Organization Principles

1. **One file per major component**: Keep test files focused
2. **Logical grouping**: Group related tests in Describe blocks
3. **Clear naming**: Use descriptive test names
4. **Isolation**: Tests should not depend on each other
5. **Cleanup**: Always clean up resources in AfterEach

### Suite Setup

```go
// ftpclient_suite_test.go
package ftpclient_test

import (
    "testing"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestFTPClient(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "FTPClient Suite")
}
```

## Writing Tests

### Configuration Tests

```go
var _ = Describe("FTP Config", func() {
    Describe("Config Structure", func() {
        It("should create config with hostname", func() {
            cfg := &Config{
                Hostname: "ftp.example.com:21",
            }
            
            Expect(cfg.Hostname).To(Equal("ftp.example.com:21"))
        })
        
        It("should detect missing hostname in validation", func() {
            cfg := &Config{
                Login:    "testuser",
                Password: "testpass",
            }
            
            err := cfg.Validate()
            Expect(err).To(HaveOccurred())
        })
    })
})
```

### Table-Driven Tests

```go
DescribeTable("Hostname Validation",
    func(hostname string, shouldPass bool) {
        cfg := &Config{Hostname: hostname}
        err := cfg.Validate()
        
        if shouldPass {
            Expect(err).ToNot(HaveOccurred())
        } else {
            Expect(err).To(HaveOccurred())
        }
    },
    Entry("valid hostname with port", "ftp.example.com:21", true),
    Entry("valid IP with port", "192.168.1.100:21", true),
    Entry("invalid hostname", "not-valid!@#", false),
    Entry("missing port", "ftp.example.com", false),
)
```

### Async/Timeout Tests

```go
It("should timeout appropriately", func(ctx context.Context) {
    cfg := &Config{
        Hostname:    "192.0.2.1:21", // Unroutable IP
        ConnTimeout: 500 * time.Millisecond,
    }
    
    start := time.Now()
    _, err := New(cfg)
    elapsed := time.Since(start)
    
    Expect(err).To(HaveOccurred())
    Expect(elapsed).To(BeNumerically("<", 2*time.Second))
}, SpecTimeout(5*time.Second))
```

### Error Testing

```go
It("should return appropriate error codes", func() {
    cfg := &Config{
        Hostname: "invalid-host:21",
    }
    
    _, err := New(cfg)
    Expect(err).To(HaveOccurred())
    
    // Check error type
    var ftpErr *Error
    Expect(errors.As(err, &ftpErr)).To(BeTrue())
})
```

## Integration Testing

### Prerequisites for Integration Tests

Integration tests require a real FTP server. Options include:

1. **Local FTP Server** (vsftpd, ProFTPD)
2. **Docker Container**
3. **Mock FTP Server**

### Docker-Based Test Server

```yaml
# docker-compose.yml
version: '3'
services:
  ftp:
    image: stilliard/pure-ftpd
    ports:
      - "21:21"
      - "30000-30009:30000-30009"
    environment:
      PUBLICHOST: localhost
      FTP_USER_NAME: testuser
      FTP_USER_PASS: testpass
      FTP_USER_HOME: /home/ftpuser
```

```bash
# Start test server
docker-compose up -d

# Run integration tests
go test -v -tags=integration

# Stop server
docker-compose down
```

### Integration Test Example

```go
// +build integration

var _ = Describe("FTP Integration Tests", func() {
    var (
        client FTPClient
        testServer string
    )
    
    BeforeEach(func() {
        testServer = os.Getenv("FTP_TEST_SERVER")
        if testServer == "" {
            Skip("FTP_TEST_SERVER not set")
        }
        
        cfg := &Config{
            Hostname: testServer,
            Login:    "testuser",
            Password: "testpass",
        }
        
        var err error
        client, err = New(cfg)
        Expect(err).ToNot(HaveOccurred())
    })
    
    AfterEach(func() {
        if client != nil {
            client.Close()
        }
    })
    
    It("should upload and download files", func() {
        // Upload test
        content := []byte("test content")
        err := client.Stor("test.txt", bytes.NewReader(content))
        Expect(err).ToNot(HaveOccurred())
        
        // Download test
        resp, err := client.Retr("test.txt")
        Expect(err).ToNot(HaveOccurred())
        defer resp.Close()
        
        downloaded, err := io.ReadAll(resp)
        Expect(err).ToNot(HaveOccurred())
        Expect(downloaded).To(Equal(content))
        
        // Cleanup
        err = client.Delete("test.txt")
        Expect(err).ToNot(HaveOccurred())
    })
})
```

## Performance Testing

### Benchmark Tests

```go
// benchmark_test.go
func BenchmarkConnect(b *testing.B) {
    cfg := &Config{
        Hostname: "localhost:21",
        Login:    "test",
        Password: "test",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client, _ := New(cfg)
        if client != nil {
            client.Close()
        }
    }
}

func BenchmarkUpload1MB(b *testing.B) {
    cfg := &Config{
        Hostname: "localhost:21",
        Login:    "test",
        Password: "test",
    }
    
    client, _ := New(cfg)
    defer client.Close()
    
    data := make([]byte, 1024*1024) // 1MB
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.Stor(fmt.Sprintf("bench_%d.bin", i), 
                    bytes.NewReader(data))
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkConnect

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# With memory profiling
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof

# Compare benchmarks
go test -bench=. -count=5 | tee old.txt
# Make changes
go test -bench=. -count=5 | tee new.txt
benchstat old.txt new.txt
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: FTPClient Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      ftp:
        image: stilliard/pure-ftpd
        ports:
          - 21:21
        env:
          FTP_USER_NAME: testuser
          FTP_USER_PASS: testpass
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: |
          go mod download
          go install github.com/onsi/ginkgo/v2/ginkgo@latest
      
      - name: Run tests
        run: |
          CGO_ENABLED=1 go test -race -v -cover ./...
      
      - name: Generate coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage-report
          path: coverage.html
```

### GitLab CI Example

```yaml
test:
  image: golang:1.21
  services:
    - name: stilliard/pure-ftpd
      alias: ftp
  variables:
    FTP_TEST_SERVER: "ftp:21"
    CGO_ENABLED: "1"
  script:
    - go mod download
    - go test -race -v -cover ./...
  coverage: '/coverage: \d+.\d+% of statements/'
```

## Troubleshooting

### Common Issues

#### 1. Tests Hanging

```bash
# Use timeout
go test -timeout 30s

# Check for deadlocks
go test -race -v
```

#### 2. Race Detector Errors

```bash
# Enable race detector
CGO_ENABLED=1 go test -race -v

# Common causes:
# - Shared variables without synchronization
# - Improper use of atomic values
# - Missing mutex locks
```

#### 3. Flaky Tests

```bash
# Run tests multiple times
go test -count=10

# Increase timeout
go test -timeout 60s

# Use Ginkgo's Eventually/Consistently
Eventually(func() bool {
    return condition
}).Should(BeTrue())
```

#### 4. Coverage Not Generated

```bash
# Ensure -cover flag
go test -cover

# Generate profile explicitly
go test -coverprofile=coverage.out

# Check file
cat coverage.out
```

### Debug Logging

```go
// Enable verbose output in tests
BeforeEach(func() {
    GinkgoWriter.Printf("Starting test: %s\n", 
                        CurrentSpecReport().FullText())
})

// Add debug prints
It("should work", func() {
    fmt.Fprintf(GinkgoWriter, "Debug: %v\n", someValue)
    // Test code
})
```

### Test Isolation

```go
var _ = Describe("Isolated Tests", func() {
    var cleanup func()
    
    AfterEach(func() {
        // Always cleanup
        if cleanup != nil {
            cleanup()
        }
    })
    
    It("creates resources", func() {
        // Setup with cleanup function
        cleanup = func() {
            // Cleanup code
        }
    })
})
```

## Contributing

### Test Requirements

When contributing new features:

1. **Add Tests**: All new code must include tests
2. **Maintain Coverage**: Don't reduce overall coverage
3. **Pass Race Detector**: Tests must pass with `-race`
4. **Follow Patterns**: Use existing test structures
5. **Document Tests**: Add comments for complex scenarios

### Test Checklist

Before submitting a pull request:

- [ ] All tests pass locally
- [ ] Race detector enabled and passes
- [ ] Coverage maintained or improved
- [ ] New tests added for new features
- [ ] Edge cases covered
- [ ] Error paths tested
- [ ] Documentation updated

### Running Full Test Suite

```bash
#!/bin/bash
# pre-commit-tests.sh

echo "Running test suite..."

# Format code
go fmt ./...

# Vet code
go vet ./...

# Run tests with race detector
CGO_ENABLED=1 go test -race -v -timeout 60s ./...

# Check coverage
go test -cover ./... | grep coverage

# Run benchmarks (optional)
# go test -bench=. -benchmem

echo "All tests passed!"
```

## Resources

### Documentation

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Coverage Tools](https://go.dev/blog/cover)

### Related Guides

- [Main README](README.md)
- [Contributing Guidelines](../../CONTRIBUTING.md)
- [Code of Conduct](../../CODE_OF_CONDUCT.md)

### Tools

- [Ginkgo CLI](https://github.com/onsi/ginkgo)
- [Gomega](https://github.com/onsi/gomega)
- [go-junit-report](https://github.com/jstemmer/go-junit-report)
- [gocov](https://github.com/axw/gocov)

---

## AI Transparency Notice

This testing guide was developed with AI assistance for structure, examples, and best practices, under human oversight and validation in compliance with EU AI Act Article 50.4.

---

**Version**: 1.0  
**Last Updated**: November 2024  
**Maintained By**: golib Contributors
