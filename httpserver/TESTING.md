# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-194%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-60%25-brightgreen)]()

Comprehensive testing documentation for the httpserver package, covering test execution, organization, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety Testing](#thread-safety-testing)
- [Integration Tests](#integration-tests)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The `httpserver` package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions and organized test suites.

### Test Suite Summary

| Package | Tests | Coverage | Status |
|---------|-------|----------|--------|
| `httpserver` | 83/84 | 53.8% | ✅ 98.8% Pass (1 skipped) |
| `httpserver/pool` | 79/79 | 63.7% | ✅ All Pass |
| `httpserver/types` | 32/32 | 100.0% | ✅ All Pass |
| **Total** | **194/195** | **~60%** | ✅ 99.5% Pass |

### Coverage Areas

- **Configuration**: Validation, cloning, edge cases
- **Server Management**: Creation, lifecycle, info methods
- **Handler Management**: Registration, execution, replacement
- **Pool Operations**: CRUD, filtering, merging, cloning
- **Type Definitions**: Constants, field types, handlers
- **Monitoring**: Server and pool monitoring integration
- **Integration**: Actual HTTP servers (build tag: `integration`)

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all unit tests
go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run with race detection (recommended)
go test -race -v ./...

# Using Ginkgo CLI
ginkgo -v -r

# Run integration tests (starts actual servers)
go test -tags=integration -v -timeout 120s ./...
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([documentation](https://onsi.github.io/ginkgo/))
- Hierarchical test organization with `Describe`, `Context`, `It`
- Setup/teardown hooks: `BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`
- Parallel execution support
- Rich CLI with filtering and focusing

**Gomega** - Matcher library ([documentation](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages
- Custom matcher support

---

## Test Structure

### Main Package Tests (`httpserver`)

| File | Purpose | Specs | Focus |
|------|---------|-------|-------|
| `httpserver_suite_test.go` | Test suite initialization | - | Suite setup, GetFreePort helper |
| `config_test.go` | Configuration validation | 15+ | Required fields, formats, validation |
| `config_clone_test.go` | Config cloning and helpers | 10+ | Deep copy, independence |
| `server_test.go` | Server creation and info | 20+ | Creation, info methods, TLS |
| `server_lifecycle_test.go` | Server lifecycle operations | 7 | Start, stop, restart, port mgmt |
| `server_handlers_test.go` | Handler operations | 5 | Registration, HTTP methods, 404 |
| `server_monitor_test.go` | Monitoring and state | 5 | State tracking, uptime, config |
| `handler_test.go` | Handler management | 12+ | Handler functions, keys, execution |
| `monitoring_test.go` | Monitoring integration | 8+ | Monitor names, server info |

### Pool Package Tests (`httpserver/pool`)

| File | Purpose | Specs | Focus |
|------|---------|-------|-------|
| `pool/pool_suite_test.go` | Pool test suite setup | 1 | Suite initialization |
| `pool/pool_test.go` | Basic pool operations | 15+ | Creation, lifecycle |
| `pool/pool_manage_test.go` | Management operations | 20+ | Store, load, delete, walk |
| `pool/pool_filter_test.go` | Filtering and listing | 25+ | Name, bind, expose filtering |
| `pool/pool_config_test.go` | Config-based pool creation | 10+ | Validation, instantiation |
| `pool/pool_merge_test.go` | Pool merging and cloning | 8+ | Merge, clone operations |

### Types Package Tests (`httpserver/types`)

| File | Purpose | Specs | Focus |
|------|---------|-------|-------|
| `types/types_suite_test.go` | Types test suite setup | 1 | Suite initialization |
| `types/handler_test.go` | Handler types | 15+ | BadHandler, FuncHandler |
| `types/fields_test.go` | Field types and constants | 16+ | FieldType, timeouts |

### Test Helpers

| File | Purpose |
|------|---------|
| `testhelpers/certs.go` | Temporary TLS certificate generation for testing |

### Integration Tests

**Build Tag**: `integration`

These tests start actual HTTP servers and perform real network operations:

| File | Purpose | Requires |
|------|---------|----------|
| `integration_test.go` | Server lifecycle testing | Network ports |
| `tls_integration_test.go` | TLS configuration validation | TLS certificates |
| `pool/integration_test.go` | Multi-server coordination | Multiple ports |

**When to Run**:
- Before commits (recommended)
- CI/CD pipelines
- After network-related changes
- When troubleshooting server issues

### Coverage Summary

| Package | Files | Tests | Coverage | Status |
|---------|-------|-------|----------|--------|
| `httpserver` | 13 | 67 | 21.3% | ✅ All Pass |
| `httpserver/pool` | 6 | 79 | 60.3% | ✅ All Pass |
| `httpserver/types` | 3 | 32 | 100.0% | ✅ All Pass |
| **Total** | **22** | **178** | **42.0%** | **✅ All Pass** |

**Coverage Focus**:
- ✅ Configuration and validation (high coverage)
- ✅ Pool management (excellent coverage)
- ✅ Type definitions (complete coverage)
- ⚠️ Server lifecycle (requires integration tests)
- ⚠️ Network I/O (integration only)

---

## Running Tests

### Basic Commands

```bash
# Run all unit tests
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test -v ./pool

# Single test file
go test -v -run TestPool
```

### Coverage Testing

```bash
# Run with coverage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out -o coverage.html

# Coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Coverage for specific package
go test -coverprofile=pool_coverage.out ./pool
```

### Integration Tests

Integration tests start actual HTTP servers and require network access:

```bash
# Run integration tests (longer timeout needed)
go test -tags=integration -v -timeout 120s ./...

# Run specific integration test
go test -tags=integration -v -run TestServerLifecycle ./...

# Integration + unit tests
go test -tags=integration -v -timeout 120s ./...

# Only unit tests (default)
go test -v ./...
```

**Note**: Integration tests bind to random ports and may fail if ports are unavailable.

### Using Ginkgo CLI

```bash
# Run all tests
ginkgo -v -r

# Specific package
ginkgo -v ./pool

# With coverage
ginkgo -v -r --cover

# Integration tests
ginkgo -v -r --tags=integration --timeout=2m

# Parallel execution
ginkgo -v -r -p

# Focus on specific tests
ginkgo -v --focus="Config Validation"

# Generate JUnit report
ginkgo -v -r --junit-report=results.xml
```

### Race Detection

**Critical for verifying thread safety**:

```bash
# Run with race detector (requires CGO)
go test -race -v ./...

# With integration tests
go test -race -tags=integration -v -timeout 120s ./...

# Using Ginkgo
ginkgo -v -r -race

# Stress test for races
for i in {1..10}; do go test -race ./... || break; done
```

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/httpserver	2.123s
ok  	github.com/nabbar/golib/httpserver/pool	1.456s
ok  	github.com/nabbar/golib/httpserver/types	0.234s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

### Performance Testing

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out

# CPU profiling
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out

# Benchtime
go test -bench=. -benchtime=10s ./...
```

---

## Test Coverage

### Configuration Tests

**Files**: `config_test.go`, `config_clone_test.go`

**Coverage Areas**:
- ✅ Required field validation (Name, Listen, Expose)
- ✅ Address format validation (hostname:port)
- ✅ URL validation for expose field
- ✅ Optional fields (HandlerKey, Disabled, TLSMandatory)
- ✅ Deep cloning and independence
- ✅ Handler function registration
- ✅ Context provider setting
- ✅ Server instantiation from config
- ✅ Edge cases (empty, invalid, boundary values)

**Test Examples**:

```go
// Valid configuration
It("should validate complete config", func() {
    cfg := Config{
        Name:   "test-server",
        Listen: "127.0.0.1:8080",
        Expose: "http://localhost:8080",
    }
    Expect(cfg.Validate()).ToNot(HaveOccurred())
})

// Missing required field
It("should fail without name", func() {
    cfg := Config{
        Listen: "127.0.0.1:8080",
        Expose: "http://localhost:8080",
    }
    Expect(cfg.Validate()).To(HaveOccurred())
})

// Config cloning
It("should create independent clone", func() {
    original := Config{Name: "original", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"}
    clone := original.Clone()
    
    clone.Name = "modified"
    Expect(original.Name).To(Equal("original"))
})
```

### Server Management Tests

**Files**: `server_test.go`, `handler_test.go`, `monitoring_test.go`

**Coverage Areas**:
- ✅ Server creation and initialization
- ✅ Info methods (GetName, GetBindable, GetExpose)
- ✅ State methods (IsDisable, IsTLS, IsRunning)
- ✅ Configuration management (GetConfig, SetConfig)
- ✅ Handler registration and execution
- ✅ Handler key-based routing
- ✅ Server merging
- ✅ Monitoring integration
- ✅ Uptime tracking

**Test Examples**:

```go
// Server creation
It("should create server from config", func() {
    cfg := Config{Name: "test", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"}
    cfg.RegisterHandlerFunc(handlerFunc)
    
    srv, err := New(cfg, nil)
    Expect(err).ToNot(HaveOccurred())
    Expect(srv.GetName()).To(Equal("test"))
})

// Handler execution
It("should execute registered handler", func() {
    called := false
    handler := func() map[string]http.Handler {
        called = true
        return map[string]http.Handler{"": http.DefaultServeMux}
    }
    
    cfg.RegisterHandlerFunc(handler)
    srv, _ := New(cfg, nil)
    
    srv.HandlerLoadFct()  // Loads and executes handler
    Expect(called).To(BeTrue())
})
```

### 3. Pool Management Tests

**Files**: `pool/pool_test.go`, `pool/pool_manage_test.go`

**Scenarios Covered**:
- Pool creation (empty, with context, with servers)
- Store and load operations
- Delete and load-delete operations
- Walk operations with callbacks
- Walk limit with bind address filter
- Server existence checking (Has)
- Pool length tracking
- Pool cleaning
- Monitor name retrieval
- Pool cloning

**Coverage**:
- ✅ Pool creation with variations
- ✅ CRUD operations (store, load, delete)
- ✅ Walk and iteration
- ✅ Filtering by bind address
- ✅ Empty pool handling
- ✅ Pool state management

### 4. Pool Filtering Tests

**File**: `pool/pool_filter_test.go`

**Scenarios Covered**:
- Filter by exact name match
- Filter by name regex
- Filter by exact bind address
- Filter by bind address regex
- Filter by expose address (exact and regex)
- List operations with field selection
- Empty pattern and regex handling
- Invalid regex handling
- Filter on empty pool
- Chain filtering
- Case sensitivity

**Coverage**:
- ✅ Name-based filtering
- ✅ Bind address filtering
- ✅ Expose address filtering
- ✅ List operations
- ✅ Edge cases
- ✅ Complex filtering scenarios

### 5. Pool Configuration Tests

**File**: `pool/pool_config_test.go`

**Scenarios Covered**:
- Config validation (all valid, partial invalid, empty)
- Pool creation from configs
- Config walking and iteration
- Handler function setting
- Context function setting
- Multiple operations in sequence
- Partial validation errors

**Coverage**:
- ✅ Config array validation
- ✅ Pool instantiation from configs
- ✅ Walk operations
- ✅ Global handler registration
- ✅ Context management
- ✅ Error aggregation

### 6. Pool Merge Tests

**File**: `pool/pool_merge_test.go`

**Scenarios Covered**:
- Merge two pools
- Merge overlapping servers
- Merge empty pools
- Merge with handler functions
- Monitor name collection
- Pool creation with initial servers

**Coverage**:
- ✅ Pool merging logic
- ✅ Overlapping server handling
- ✅ Handler management
- ✅ Monitor aggregation
- ✅ Initial server handling

### 7. Types Tests

**Files**: `types/handler_test.go`, `types/fields_test.go`

**Scenarios Covered**:
- BadHandler creation and execution
- BadHandler with different HTTP methods
- BadHandler with different paths
- FuncHandler type definition
- FieldType constants (FieldName, FieldBind, FieldExpose)
- Timeout constants
- HandlerDefault constant
- BadHandlerName constant
- Constant usage in maps and switches

**Coverage**:
- ✅ 100% coverage of types package
- ✅ All constants tested
- ✅ Handler functionality verified
- ✅ Type assertions validated

### 8. Monitoring Tests

**File**: `monitoring_test.go`

**Scenarios Covered**:
- Monitor name generation and uniqueness
- Monitor interface availability
- Server info for monitoring (name, bind, expose, state)
- State change reflection in monitoring data

**Coverage**:
- ✅ Monitor name retrieval
- ✅ Server state tracking
- ✅ Info methods validation
- ✅ Configuration changes reflection

### 9. Integration Tests

**Files**: `integration_test.go`, `tls_integration_test.go`, `pool/integration_test.go`

**Build Tag**: `integration`

**Scenarios Covered**:
- **Server Lifecycle**: Start, stop, restart with actual HTTP servers
- **HTTP Handling**: GET/POST requests, different paths
- **TLS Configuration**: TLS mandatory flag, certificate validation
- **Pool Lifecycle**: Multiple servers start/stop, uptime tracking
- **Pool Operations**: Multiple handlers, partial failures, monitoring
- **Disabled Servers**: Graceful handling of disabled servers

**Coverage**:
- ✅ Real HTTP server operations
- ✅ Network request/response handling
- ✅ Multi-server coordination
- ✅ TLS configuration validation
- ✅ Dynamic port allocation
- ✅ Graceful shutdown

**Note**: Integration tests use build tags to avoid running during normal unit test execution. They require more time and actually bind to network ports.

## Testing Challenges

### No Server Startup in Unit Tests

**Challenge**: Starting actual HTTP servers requires port allocation and creates complexity.

**Solution**: Focus on configuration, structure, and lifecycle methods without actual network binding.

**Approach**:
- Test configuration validation
- Test data structure manipulation
- Test pool management logic
- Integration tests can be added separately with build tags

### TLS Configuration

**Challenge**: Testing TLS requires certificates.

**Solution**: Test configuration structure without actual TLS connections.

## Best Practices

### 1. Test Configuration, Not Network Operations

```go
// Good - tests config validation
It("should validate required fields", func() {
    cfg := Config{
        Name: "server",
        // Missing Listen and Expose
    }
    Expect(cfg.Validate()).To(HaveOccurred())
})

// Avoid in unit tests - requires network
It("should start server", func() {
    server.Start(context.Background())
})
```

### 2. Use Table-Driven Tests

```go
DescribeTable("Listen address formats",
    func(listen string, shouldPass bool) {
        cfg := Config{
            Name:   "test",
            Listen: listen,
            Expose: "http://localhost:8080",
        }
        
        err := cfg.Validate()
        if shouldPass {
            Expect(err).ToNot(HaveOccurred())
        } else {
            Expect(err).To(HaveOccurred())
        }
    },
    Entry("IPv4 with port", "192.168.1.1:8080", true),
    Entry("localhost", "localhost:8080", true),
    Entry("invalid format", "not-valid", false),
)
```

### 3. Clean Pool State

```go
var pool Pool

BeforeEach(func() {
    pool = New(nil, nil)
})

AfterEach(func() {
    pool.Clean()
})
```

### 4. Test Edge Cases

```go
It("should handle empty pool operations", func() {
    pool := New(nil, nil)
    
    Expect(pool.Len()).To(Equal(0))
    Expect(pool.Has("any")).To(BeFalse())
    Expect(pool.MonitorNames()).To(BeEmpty())
})
```

## Performance Considerations

### Test Performance

| Operation | Time/test | Notes |
|-----------|-----------|-------|
| Config validation | <1ms | Very fast |
| Pool creation | <1ms | Lightweight |
| Full test suite | ~20ms | All 27 specs |

### Benchmark Example

```go
Benchmark("Config validation", func(b *testing.B) {
    cfg := Config{
        Name:   "bench",
        Listen: "127.0.0.1:8080",
        Expose: "http://localhost:8080",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = cfg.Validate()
    }
})
```

## Debugging Tests

### Enable Verbose Output

```bash
ginkgo -v --trace
```

### Focus Specific Tests

```bash
ginkgo -v --focus "Config Validation"
ginkgo -v --focus "should validate"
```

### Check Pool State

```go
It("debug pool state", func() {
    fmt.Printf("Pool size: %d\n", pool.Len())
    fmt.Printf("Monitor names: %v\n", pool.MonitorNames())
})
```

## CI/CD Integration

### GitHub Actions Example

```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run Tests
      run: go test -v -race -cover ./httpserver/...
    
    - name: Coverage Report
      run: |
        go test -coverprofile=coverage.out ./httpserver/...
        go tool cover -html=coverage.out -o coverage.html
```

## Common Issues

### 1. Port Conflicts

Tests don't start servers, so no port conflicts occur in unit tests.

For integration tests:
```go
// Use dynamic port allocation
listener, _ := net.Listen("tcp", ":0")
port := listener.Addr().(*net.TCPAddr).Port
```

### 2. Validation Errors

```go
It("should show validation details", func() {
    cfg := Config{Name: "test"}
    err := cfg.Validate()
    
    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
    }
})
```

## Test Coverage Goals

### Current Coverage: 41.8% ✅ Target Achieved!

**Well Covered**:
- ✅ Configuration validation and cloning
- ✅ Server creation and info methods
- ✅ Handler registration and management
- ✅ Pool creation and CRUD operations
- ✅ Pool filtering (name, bind, expose)
- ✅ Pool merging and cloning
- ✅ Types package (100% coverage)
- ✅ Config helpers and edge cases

**Partially Covered**:
- ⚠️ Server lifecycle methods (requires actual server startup)
- ⚠️ Network I/O operations (integration tests)
- ⚠️ Monitoring integration (requires running servers)
- ⚠️ TLS connections (requires certificates)

**Not Covered** (by design):
- ❌ Actual HTTP server startup
- ❌ Network binding and listening
- ❌ TLS handshakes
- ❌ Request/response handling

### Coverage by Package:
- **httpserver**: 21.3% (67 tests) - Config, structure, and monitoring
- **httpserver/pool**: 60.3% (79 tests) - Excellent pool management coverage
- **httpserver/types**: 100.0% (32 tests) - Complete coverage
- **Integration tests**: Build tag separated for optional execution

### Additional Coverage Ideas:

1. **Add Server Structure Tests**:
```go
It("should get server info", func() {
    cfg := Config{
        Name:   "info-test",
        Listen: "127.0.0.1:8080",
        Expose: "http://localhost:8080",
    }
    
    // Test info methods
    Expect(cfg.Name).To(Equal("info-test"))
})
```

2. **Add Pool Filter Tests**:
```go
It("should filter servers", func() {
    pool := New(nil, nil)
    
    // Test filter operations
    filtered := pool.Filter(/* params */)
    Expect(filtered).ToNot(BeNil())
})
```

3. **Add Handler Tests**:
```go
It("should register handler", func() {
    pool := New(nil, nil)
    
    handler := func() map[string]http.Handler {
        return make(map[string]http.Handler)
    }
    
    pool.Handler(handler)
    // Verify handler is registered
})
```

## Integration Testing (Optional)

### With Build Tags

Create integration tests with build tags:

```go
// +build integration

package httpserver_test

var _ = Describe("Server Integration", func() {
    It("should start and stop server", func() {
        cfg := Config{
            Name:   "integration-test",
            Listen: "127.0.0.1:18080",
            Expose: "http://localhost:18080",
        }
        
        srv, err := New(cfg, nil)
        Expect(err).ToNot(HaveOccurred())
        
        ctx := context.Background()
        err = srv.Start(ctx)
        Expect(err).ToNot(HaveOccurred())
        
        time.Sleep(100 * time.Millisecond)
        
        err = srv.Stop(ctx)
        Expect(err).ToNot(HaveOccurred())
    })
})
```

Run integration tests:
```bash
go test -tags=integration -v
```

## Useful Commands

```bash
# Quick test
go test ./...

# Verbose
go test -v ./...

# Coverage
go test -cover ./...

# HTML coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detection
go test -race ./...

# Ginkgo
ginkgo -v -r
ginkgo -v -r --cover
ginkgo watch -r

# Specific package
go test -v ./pool
ginkgo -v ./pool

# Benchmarks
go test -bench=. -benchmem ./...
```

---

## Thread Safety Testing

Thread safety is critical for concurrent server management.

### Race Detection

```bash
# Standard race detection
go test -race -v ./...

# Extended stress test
for i in {1..20}; do 
    go test -race ./... || break
done

# Race detection with integration tests
go test -race -tags=integration -v -timeout 120s ./...
```

### Thread Safety Components

| Component | Protection | Verified |
|-----------|-----------|----------|
| Server State | `atomic.Value` | ✅ |
| Handler Registry | `atomic.Value` | ✅ |
| Logger | `atomic.Value` | ✅ |
| Pool Map | `sync.RWMutex` | ✅ |
| Runner | `atomic.Value` + `sync.WaitGroup` | ✅ |

---

## Integration Tests

Integration tests verify real HTTP server behavior with actual network operations.

### Running Integration Tests

```bash
# All integration tests
go test -tags=integration -v -timeout 120s ./...

# Specific integration test
go test -tags=integration -run TestServerLifecycle -v ./...

# With race detection
go test -race -tags=integration -v -timeout 180s ./...
```

### Integration Test Examples

```go
// +build integration

It("should start and handle HTTP requests", func() {
    cfg := Config{
        Name:   "integration-test",
        Listen: "127.0.0.1:0",  // Random port
        Expose: "http://localhost",
    }
    cfg.RegisterHandlerFunc(handlerFunc)
    
    srv, _ := New(cfg, nil)
    Expect(srv.Start(ctx)).ToNot(HaveOccurred())
    
    // Make HTTP request
    resp, err := http.Get("http://" + srv.GetBindable())
    Expect(err).ToNot(HaveOccurred())
    Expect(resp.StatusCode).To(Equal(http.StatusOK))
    
    srv.Stop(ctx)
})
```

---

## Writing Tests

### Test Template

```go
var _ = Describe("New Feature", func() {
    var cfg Config
    
    BeforeEach(func() {
        cfg = Config{
            Name:   "test",
            Listen: "127.0.0.1:8080",
            Expose: "http://localhost:8080",
        }
        cfg.RegisterHandlerFunc(defaultHandler)
    })
    
    It("should perform expected behavior", func() {
        // Arrange
        srv, err := New(cfg, nil)
        Expect(err).ToNot(HaveOccurred())
        
        // Act
        result := srv.GetName()
        
        // Assert
        Expect(result).To(Equal("test"))
    })
})
```

### Pool Test Template

```go
var _ = Describe("Pool Feature", func() {
    var p Pool
    
    BeforeEach(func() {
        p = pool.New(nil, defaultHandler)
    })
    
    AfterEach(func() {
        p.Clean()
    })
    
    It("should manage servers", func() {
        Expect(p.Len()).To(Equal(0))
        
        cfg := httpserver.Config{Name: "test", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"}
        Expect(p.StoreNew(cfg, nil)).ToNot(HaveOccurred())
        
        Expect(p.Len()).To(Equal(1))
    })
})
```

---

## Best Practices

### Test Organization

- ✅ One test file per feature area
- ✅ Use descriptive `Describe` and `Context` blocks
- ✅ One assertion per `It` block when possible
- ✅ Setup in `BeforeEach`, cleanup in `AfterEach`
- ✅ Use table-driven tests for multiple similar cases

### Test Quality

```go
// ✅ Good: Clear, focused test
It("should validate required name field", func() {
    cfg := Config{Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"}
    Expect(cfg.Validate()).To(HaveOccurred())
})

// ❌ Bad: Testing multiple things
It("should validate config", func() {
    // Tests multiple validations at once - split into separate tests
})
```

### Assertions

```go
// ✅ Good: Specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))
Expect(list).To(ContainElement(item))

// ❌ Bad: Generic assertions
Expect(err == nil).To(BeTrue())  // Use ToNot(HaveOccurred())
```

---

## Troubleshooting

### Test Failures

```bash
# Run failed test with verbose output
ginkgo -v --focus="failing test name"

# Check for race conditions
go test -race -run TestName ./...

# Debug with trace
ginkgo -v --trace --focus="test name"
```

### Common Issues

**Port Conflicts**
```go
// Use :0 for random port allocation in integration tests
cfg := Config{Listen: "127.0.0.1:0", ...}
```

**Race Conditions**
```go
// Always protect shared state
// Use atomic.Value or sync.Mutex
```

**Flaky Tests**
```bash
# Run multiple times to identify
for i in {1..50}; do go test ./... || break; done
```

---

## CI Integration

### GitHub Actions

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
      
      - name: Unit Tests
        run: go test -v -cover ./...
      
      - name: Race Detection
        run: go test -race -v ./...
      
      - name: Integration Tests
        run: go test -tags=integration -v -timeout 120s ./...
      
      - name: Coverage Report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload Coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage
          path: coverage.html
```

---

## Contributing

When contributing tests:

**Test Development Guidelines**
- **AI assistance is permitted** for test development, documentation, and bug fixes
- **Do not use AI** to generate package implementation code
- All test contributions must use Ginkgo v2 and Gomega
- Organize tests by scope (one test file per feature area)
- Keep test files readable and maintainable
- Use descriptive test names with `It("should ...")`

**Test Quality Standards**
- All tests must pass with race detector: `CGO_ENABLED=1 go test -race ./...`
- Maintain or improve coverage (target ≥60%)
- Test edge cases and error conditions
- Include both positive and negative test cases
- Add integration tests with `integration` build tag when needed

**Code Organization**
- Place tests in `*_test.go` files next to the code they test
- Use test suites (`*_suite_test.go`) for package-level setup
- Group related tests in `Describe` blocks
- Use `Context` for different scenarios
- Add helper functions to reduce test duplication

**Pull Request Requirements**
- Include test results in PR description
- Show coverage changes (before/after)
- Document any skipped tests with reason
- Update TESTING.md if test structure changes

---

## Quality Checklist

Before merging:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained or improved (target ≥60%)
- [ ] Integration tests pass: `go test -tags=integration ./...`
- [ ] New features have tests
- [ ] Edge cases tested
- [ ] Documentation updated
- [ ] Test files are organized and readable

---

## Resources

- **Testing Guide**: This document
- **Package Documentation**: [README.md](README.md)
- **Ginkgo Documentation**: [https://onsi.github.io/ginkgo/](https://onsi.github.io/ginkgo/)
- **Gomega Matchers**: [https://onsi.github.io/gomega/](https://onsi.github.io/gomega/)
- **Go Testing**: [https://pkg.go.dev/testing](https://pkg.go.dev/testing)
- **Race Detector**: [https://go.dev/doc/articles/race_detector](https://go.dev/doc/articles/race_detector)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: httpserver Package Contributors
