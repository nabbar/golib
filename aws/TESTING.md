# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-208%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-S3%2080%25-brightgreen)]()

Comprehensive testing documentation for the AWS package with automatic MinIO integration for S3 operations and IAM API testing.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [MinIO Integration](#minio-integration)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)
- [Resources](#resources)

---

## Overview

The AWS package testing strategy focuses on:

- **MinIO Integration**: Automatic S3-compatible server lifecycle
- **Zero External Dependencies**: No AWS account required for S3 tests
- **Comprehensive Coverage**: 208+ test specifications
- **Fast Execution**: <15 minutes for full suite
- **Thread-Safe Testing**: Concurrent operation validation

### Test Statistics

- **Total Specs**: 208+
- **S3 Tests**: ~100 (100% pass with MinIO)
- **IAM Tests**: ~40 (AWS only, auto-skipped with MinIO)
- **Performance Tests**: ~8 stress/benchmark tests
- **Execution Time**: ~15 minutes (full suite)

### Test Philosophy

**Key Principles**:
- ✅ Real S3 operations (via MinIO)
- ✅ No mocks for S3 (use actual server)
- ✅ Automatic cleanup after tests
- ✅ Isolated test environments
- ✅ Seekable test data for AWS SDK compatibility

---

## Quick Start

### Automatic MinIO Mode (Recommended)

```bash
# MinIO binary must be present in aws/ directory
cd /sources/go/src/github.com/nabbar/golib/aws
chmod +x minio

# Run all tests (MinIO starts automatically)
go test -v -timeout=30m

# Run with coverage
go test -v -cover -timeout=30m

# Using Ginkgo
ginkgo -v -cover
```

### Manual Configuration Mode

Create `config.json`:
```json
{
    "bucket": "test-bucket",
    "region": "us-east-1",
    "access_key": "your-access-key",
    "secret_key": "your-secret-key",
    "endpoint": "http://localhost:9000"
}
```

Then run tests:
```bash
go test -v -timeout=30m
```

---

## Test Framework

**Framework**: Ginkgo v2 + Gomega  
**Style**: BDD (Behavior-Driven Development)  
**Language**: English

### Test File Organization

```
aws/
├── aws_suite_test.go              # Suite setup + MinIO lifecycle
├── client_test.go                 # Client management (47 specs)
├── config_test.go                 # Configuration (28 specs)
├── s3_bucket_basic_test.go        # Bucket basics (15+ specs)
├── s3_bucket_advanced_test.go     # Versioning, policies (10+ specs)
├── s3_bucket_cors_test.go         # CORS configuration (12+ specs)
├── s3_object_operations_test.go   # Object CRUD (20+ specs)
├── s3_object_tags_test.go         # Object tagging (15+ specs)
├── s3_pusher_upload_test.go       # Multipart upload (12+ specs)
├── s3_stress_test.go              # Performance (8+ specs)
├── iam_user_operations_test.go    # IAM users (10+ specs)
├── iam_group_operations_test.go   # IAM groups (10+ specs)
├── iam_role_operations_test.go    # IAM roles (8+ specs)
└── iam_policy_operations_test.go  # IAM policies (10+ specs)
```

---

## Running Tests

### Basic Commands

```bash
# All tests
go test -v -timeout=30m ./...

# Single package
go test -v -timeout=5m

# Specific test
go test -v -run "TestGolibAwsHelper"

# With coverage
go test -v -cover -coverprofile=coverage.out -timeout=30m ./...
```

### Ginkgo Commands

```bash
# Verbose output
ginkgo -v

# With coverage
ginkgo -v -cover

# Parallel execution
ginkgo -p -v

# Focus on specific tests
ginkgo -v -focus="bucket creation"

# Skip slow tests
ginkgo -v -skip="stress"
```

### Test Categories

```bash
# S3 bucket tests only
go test -v -run "S3 Bucket"

# S3 object tests only
go test -v -run "S3 Object"

# IAM tests only (requires AWS config)
go test -v -run "IAM"

# Performance/stress tests
go test -v -run "Stress"
```

---

## Test Coverage

### Coverage By Component

| Component | File | Specs | Status | Notes |
|-----------|------|-------|--------|-------|
| Client Management | client_test.go | 47 | ✅ Pass | Client creation, cloning, timeout |
| Configuration | config_test.go | 28 | ✅ Pass | AWS/Custom config, JSON serialization |
| S3 Buckets (Basic) | s3_bucket_basic_test.go | 15+ | ✅ Pass | Create, delete, list, exist |
| S3 Buckets (Advanced) | s3_bucket_advanced_test.go | 10+ | ✅ Pass | Versioning, policies |
| S3 CORS | s3_bucket_cors_test.go | 12+ | ✅ Pass | CORS rules, headers |
| S3 Objects | s3_object_operations_test.go | 20+ | ✅ Pass | Put, get, delete, copy |
| S3 Object Tags | s3_object_tags_test.go | 15+ | ✅ Pass | Tag management |
| S3 Pusher | s3_pusher_upload_test.go | 12+ | ✅ Pass | Multipart upload |
| S3 Stress | s3_stress_test.go | 8+ | ✅ Pass | Performance benchmarks |
| IAM Users | iam_user_operations_test.go | 10+ | ⚠️ Skip | MinIO incompatible |
| IAM Groups | iam_group_operations_test.go | 10+ | ⚠️ Skip | MinIO incompatible |
| IAM Roles | iam_role_operations_test.go | 8+ | ⚠️ Skip | MinIO incompatible |
| IAM Policies | iam_policy_operations_test.go | 10+ | ⚠️ Skip | MinIO incompatible |

### Coverage Metrics

```bash
# Generate coverage report
go test -coverprofile=coverage.out -timeout=30m ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

**Coverage Highlights**:
- S3 Operations: ~80%
- Configuration: ~90%
- Client Management: ~85%
- IAM Operations: ~60% (limited MinIO testing)

---

## MinIO Integration

### Automatic Lifecycle

The test suite automatically manages MinIO:

```
BeforeSuite:
1. Check for config.json
2. If not found:
   ├─ Generate random port (e.g., 38619)
   ├─ Generate random credentials
   ├─ Start MinIO process in background
   └─ Wait for readiness (max 60s)
3. Create AWS client
4. Enable path-style URLs (required for MinIO)
5. Create test bucket

AfterSuite:
1. Cancel context
2. MinIO process terminates
3. Cleanup temp data directory
```

### MinIO Compatibility Matrix

#### ✅ Fully Compatible

- Bucket operations (create, delete, list, exist, location)
- Object operations (put, get, delete, list, copy, head)
- Multipart uploads
- Object tagging
- Bucket versioning
- CORS configuration
- Object metadata
- Presigned URLs (basic)

#### ⚠️ Limited Compatibility

- IAM users (different API)
- IAM groups (different API)
- IAM roles (different API)
- IAM policies (different API)

#### ❌ Not Compatible

- AWS-specific IAM features
- STS temporary credentials
- Cross-region replication
- CloudWatch metrics
- S3 Glacier storage classes

### Test Mode Detection

Tests automatically detect environment:

```go
var minioMode = false  // Set during suite initialization

BeforeEach(func() {
    if minioMode {
        Skip("IAM not fully compatible with MinIO")
    }
})
```

---

## Writing Tests

### Test Template

```go
var _ = Describe("AWS Feature", func() {
    Context("When performing operation", func() {
        var testData []byte

        BeforeEach(func() {
            testData = []byte("test content")
        })

        It("should perform expected behavior", func() {
            // Arrange
            reader := bytes.NewReader(testData)
            
            // Act
            err := client.Object().Put("test.txt", reader)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
        })

        AfterEach(func() {
            // Cleanup
            _ = client.Object().Delete("test.txt")
        })
    })
})
```

### Seekable Test Data

AWS SDK requires seekable readers for checksum calculation:

```go
// randReader implements io.ReadSeeker
type randReader struct {
    data []byte
    pos  int64
}

func (r *randReader) Read(p []byte) (n int, err error) {
    if r.pos >= int64(len(r.data)) {
        return 0, io.EOF
    }
    n = copy(p, r.data[r.pos:])
    r.pos += int64(n)
    return n, nil
}

func (r *randReader) Seek(offset int64, whence int) (int64, error) {
    var newPos int64
    switch whence {
    case io.SeekStart:
        newPos = offset
    case io.SeekCurrent:
        newPos = r.pos + offset
    case io.SeekEnd:
        newPos = int64(len(r.data)) + offset
    }
    r.pos = newPos
    return newPos, nil
}

// Usage
func randContent(size int64) io.ReadSeeker {
    data := make([]byte, size)
    rand.Read(data)
    return &randReader{data: data, pos: 0}
}
```

---

## Best Practices

### Test Independence

```go
// Each test should be independent
It("should create bucket", func() {
    bucketName := GenerateUniqueName("test-bucket")
    err := client.Bucket().Create(bucketName)
    Expect(err).ToNot(HaveOccurred())
    
    // Clean up in AfterEach, not here
})
```

### Unique Names

```go
// Always use unique names to avoid conflicts
bucketName := GenerateUniqueName("test-bucket")
objectKey := fmt.Sprintf("test-%d.txt", time.Now().UnixNano())
```

### Resource Cleanup

```go
AfterEach(func() {
    // Always clean up resources
    _ = client.Object().Delete(testKey)
    _ = client.Bucket().Delete(testBucket)
})
```

### Error Assertions

```go
// Check errors properly
err := client.Object().Put(key, reader)
Expect(err).ToNot(HaveOccurred())

// Check specific errors
exists, err := client.Bucket().Exist("nonexistent")
Expect(err).ToNot(HaveOccurred())
Expect(exists).To(BeFalse())
```

### Realistic Data Sizes

```go
// Test with realistic file sizes
smallFile := randContent(10 * 1024)           // 10KB
mediumFile := randContent(5 * 1024 * 1024)    // 5MB
largeFile := randContent(100 * 1024 * 1024)   // 100MB
```

---

## Troubleshooting

### MinIO Won't Start

**Problem**: MinIO binary not found or not executable

**Solution**:
```bash
# Download MinIO
wget https://dl.min.io/server/minio/release/linux-amd64/minio
chmod +x minio
mv minio /sources/go/src/github.com/nabbar/golib/aws/
```

### Port Already in Use

**Problem**: Random port allocation fails

**Solution**: The suite automatically finds a free port. If issues persist:
```bash
# Kill any existing MinIO processes
pkill -9 minio

# Check port usage
netstat -tulpn | grep minio
```

### Seekable Reader Error

**Problem**:
```
failed to compute payload hash: failed to seek body to start, 
request stream is not seekable
```

**Solution**: Use `io.ReadSeeker` instead of `io.Reader`:
```go
// ✅ Good
reader := bytes.NewReader(data)
err := client.Object().Put(key, reader)

// ❌ Bad
reader := bytes.NewBuffer(data)
err := client.Object().Put(key, reader)
```

### IAM Tests Failing

**Problem**: IAM tests fail with MinIO

**Solution**: IAM tests are automatically skipped in MinIO mode. To run them:
```bash
# Create config.json with real AWS credentials
# Tests will use AWS IAM API instead of MinIO
```

---

## CI Integration

### GitHub Actions

```yaml
name: AWS Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Download MinIO
        run: |
          wget https://dl.min.io/server/minio/release/linux-amd64/minio
          chmod +x minio
          mv minio aws/
      
      - name: Run Tests
        run: |
          cd aws
          go test -v -timeout=30m -cover
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./aws/coverage.out
```

### GitLab CI

```yaml
test:aws:
  image: golang:1.21
  
  before_script:
    - wget https://dl.min.io/server/minio/release/linux-amd64/minio
    - chmod +x minio
    - mv minio aws/
  
  script:
    - cd aws
    - go test -v -timeout=30m -cover
  
  coverage: '/coverage: \d+\.\d+% of statements/'
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

cd aws || exit 1
go test -v -timeout=5m -run "TestGolibAwsHelper" || exit 1
echo "AWS tests passed"
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)

**AWS & MinIO**
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [MinIO Testing](https://min.io/docs/minio/linux/developers/go/minio-go.html)

**Support**
- [GitHub Issues](https://github.com/nabbar/golib/issues)
- [README](README.md)

---

**Test Suite Version**: Go 1.18+ on Linux, macOS, Windows  
**MinIO Tested**: v2023.11+  
**Maintained By**: AWS Package Contributors
