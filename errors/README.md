# Errors Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/errors)

**Advanced error handling with error codes, tracing, hierarchy, and collection management.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Sub-Packages](#sub-packages)
- [Quick Start](#quick-start)
- [Error Codes](#error-codes)
- [Error Hierarchy](#error-hierarchy)
- [Stack Tracing](#stack-tracing)
- [API Reference](#api-reference)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **errors** package provides advanced error handling capabilities beyond Go's standard library. It adds error codes, automatic stack tracing, error hierarchies, and utilities for error collection and management.

### Design Philosophy

- **Error Codes**: Assign numeric codes for programmatic error identification
- **Tracing**: Automatic file/line tracking for error origins
- **Hierarchy**: Chain errors with parent-child relationships
- **Type Safety**: Strongly typed error codes with constants
- **Compatibility**: Works with standard Go error handling patterns
- **Performance**: Minimal overhead with efficient error creation

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Error Codes** | Numeric error codes for classification |
| **Stack Traces** | Automatic file/line number tracking |
| **Error Hierarchy** | Parent-child error chains |
| **Error Pool** | Thread-safe error collection (sub-package) |
| **Code Constants** | Predefined HTTP-like error codes |
| **Pattern Matching** | Search errors by code or message |
| **Gin Integration** | Direct integration with Gin framework |
| **Type Safe** | Strong typing for error codes |

---

## Architecture

### Package Structure

```
errors/
├── interface.go        # Main Error interface
├── code.go            # Error code definitions and constants
├── errors.go          # Error creation and management
├── trace.go           # Stack trace functionality
├── return.go          # Error return helpers
├── compat.go          # Standard library compatibility
├── mode.go            # Error mode (dev/prod)
├── modules.go         # Module-specific errors
└── pool/              # Error collection sub-package
    ├── interface.go
    └── model.go
```

### Error Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Error Interface                     │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │         Error Code (uint16)                  │  │
│  │  - Predefined constants (404, 500, etc.)     │  │
│  │  - Custom codes                              │  │
│  └──────────────────────────────────────────────┘  │
│                       │                              │
│  ┌──────────────────────────────────────────────┐  │
│  │         Stack Trace                          │  │
│  │  - File path                                 │  │
│  │  - Line number                               │  │
│  │  - Function name                             │  │
│  └──────────────────────────────────────────────┘  │
│                       │                              │
│  ┌──────────────────────────────────────────────┐  │
│  │         Error Hierarchy                      │  │
│  │  - Parent errors (chain)                     │  │
│  │  - Child errors                              │  │
│  │  - Multi-error support                       │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Error Flow

```
Error Creation → Code Assignment → Stack Trace → Hierarchy
       │              │                │              │
       ▼              ▼                ▼              ▼
   New Error    HTTP/Custom      File:Line      Add Parents
                   Code           Automatic       Optional
```

---

## Sub-Packages

### Pool Package

**Purpose**: Thread-safe error collection with automatic indexing

**Features:**
- Concurrent error collection
- Automatic sequential indexing
- Sparse index support
- Query operations (Get, Set, Del)
- Combined error generation

**Use Cases:**
- Batch operation error collection
- Multi-goroutine error aggregation
- Error accumulation in loops
- Validation error collection

**Documentation**: [pool/README.md](pool/README.md)

**Quick Example:**
```go
import "github.com/nabbar/golib/errors/pool"

p := pool.New()

// Collect errors from multiple operations
for _, item := range items {
    if err := process(item); err != nil {
        p.Add(err)
    }
}

// Check if any errors occurred
if err := p.Error(); err != nil {
    return fmt.Errorf("processing failed: %w", err)
}
```

**Test Coverage**: 83 specs, 100% coverage, 0 race conditions

---

## Quick Start

### Basic Error Creation

```go
package main

import (
    "fmt"
    liberr "github.com/nabbar/golib/errors"
)

func main() {
    // Create error with code
    err := liberr.NotFoundError.Error(nil)
    fmt.Println(err) // 404: Not Found
    
    // Create custom error
    customErr := liberr.New(func(code liberr.CodeError) liberr.Error {
        return code.Error(nil)
    }, 1001)
    fmt.Println(customErr.Code()) // 1001
}
```

### With Stack Trace

```go
func processData() error {
    // Stack trace automatically captured
    return liberr.InternalError.Error(nil)
}

func main() {
    if err := processData(); err != nil {
        if e, ok := err.(liberr.Error); ok {
            fmt.Printf("Error at %s:%d\n", e.GetFile(), e.GetLine())
        }
    }
}
```

### Error Hierarchy

```go
func operation() error {
    err1 := liberr.ValidationError.Error(nil)
    err2 := liberr.InternalError.Error(nil)
    
    // Chain errors
    mainErr := liberr.UnknownError.Error(nil)
    mainErr.Add(err1, err2)
    
    return mainErr
}
```

---

## Error Codes

### Predefined HTTP Codes

| Code | Constant | Description |
|------|----------|-------------|
| 200 | SuccessCode | Success |
| 400 | BadRequestError | Bad Request |
| 401 | UnauthorizedError | Unauthorized |
| 403 | ForbiddenError | Forbidden |
| 404 | NotFoundError | Not Found |
| 408 | TimeoutError | Request Timeout |
| 409 | ConflictError | Conflict |
| 422 | ValidationError | Validation Error |
| 429 | TooManyRequestsError | Too Many Requests |
| 500 | InternalError | Internal Error |
| 501 | NotImplementedError | Not Implemented |
| 503 | ServiceUnavailableError | Service Unavailable |

### Custom Error Codes

```go
const (
    MyCustomError = liberr.CodeError(2000)
    DatabaseError = liberr.CodeError(2001)
    CacheError    = liberr.CodeError(2002)
)

func validateInput(data string) error {
    if len(data) == 0 {
        return MyCustomError.Error(nil)
    }
    return nil
}
```

### Code Checking

```go
err := liberr.NotFoundError.Error(nil)

// Check specific code
if err.(liberr.Error).IsCode(liberr.NotFoundError) {
    fmt.Println("Resource not found")
}

// Check in hierarchy
if err.(liberr.Error).HasCode(liberr.NotFoundError) {
    fmt.Println("Not found error in chain")
}

// Get code
code := err.(liberr.Error).GetCode()
fmt.Printf("Error code: %d\n", code)
```

---

## Error Hierarchy

### Creating Error Chains

```go
// Bottom level error
dbErr := liberr.InternalError.Error(errors.New("database connection failed"))

// Mid level error
serviceErr := liberr.ServiceUnavailableError.Error(nil)
serviceErr.Add(dbErr)

// Top level error
apiErr := liberr.BadRequestError.Error(nil)
apiErr.Add(serviceErr)

// apiErr now contains: BadRequest → ServiceUnavailable → Internal → db error
```

### Traversing Hierarchy

```go
err := createComplexError()

// Get all parent errors
parents := err.GetParent(true) // includes main error
for _, e := range parents {
    fmt.Println(e)
}

// Check for specific error
if err.HasError(specificErr) {
    fmt.Println("Found specific error in chain")
}

// Map over all errors
err.Map(func(e error) bool {
    fmt.Println(e.Error())
    return true // continue
})
```

### Error Unwrapping

```go
// Compatible with errors.Is and errors.As
err := liberr.NotFoundError.Error(baseErr)

if errors.Is(err, baseErr) {
    fmt.Println("Error matches")
}

var target *SpecificError
if errors.As(err, &target) {
    fmt.Println("Error is SpecificError")
}
```

---

## Stack Tracing

### Automatic Tracing

```go
func readFile(path string) error {
    // Error created here captures this location
    return liberr.NotFoundError.Error(nil)
}

func main() {
    if err := readFile("data.txt"); err != nil {
        if e, ok := err.(liberr.Error); ok {
            fmt.Printf("Error at %s:%d in %s\n", 
                e.GetFile(), 
                e.GetLine(),
                e.GetFunction())
        }
    }
}
```

### Manual Tracing

```go
err := liberr.New(func(code liberr.CodeError) liberr.Error {
    e := code.Error(nil)
    // Manually set trace
    e.SetTrace(runtime.Caller(0))
    return e
}, 1001)
```

### Trace Information

```go
err := liberr.InternalError.Error(nil)

if e, ok := err.(liberr.Error); ok {
    // Get trace details
    file := e.GetFile()      // "/path/to/file.go"
    line := e.GetLine()      // 42
    fn := e.GetFunction()    // "main.processData"
    
    // Format trace
    trace := e.GetTrace()    // "file.go:42"
    fullTrace := fmt.Sprintf("%s:%d %s", file, line, fn)
}
```

---

## API Reference

### Error Interface

**Core Methods:**
- `IsCode(code CodeError) bool` - Check error code match
- `HasCode(code CodeError) bool` - Check code in hierarchy
- `GetCode() CodeError` - Get error code
- `Code() uint16` - Get numeric code

**Hierarchy Methods:**
- `Add(parent ...error)` - Add parent errors
- `SetParent(parent ...error)` - Replace parents
- `HasParent() bool` - Check if has parents
- `GetParent(withMain bool) []error` - Get parent chain
- `HasError(err error) bool` - Find error in chain
- `Map(fct FuncMap) bool` - Iterate over hierarchy

**Trace Methods:**
- `GetFile() string` - Get source file
- `GetLine() int` - Get line number
- `GetFunction() string` - Get function name
- `GetTrace() string` - Get formatted trace
- `SetTrace(pc uintptr, file string, line int, ok bool)` - Set trace

**Utility Methods:**
- `Error() string` - Standard error message
- `Is(e error) bool` - Compatibility with errors.Is
- `ContainsString(s string) bool` - Search in messages
- `CodeError(pattern string) string` - Formatted code+message

### CodeError Type

```go
type CodeError uint16

// Create error from code
err := NotFoundError.Error(parentErr)

// Error creation
func (c CodeError) Error(parent error) Error

// Check if error exists
func (c CodeError) IfError(errs ...error) error
```

### Helper Functions

```go
// Create new error with function
func New(fct func(code CodeError) Error, code int) Error

// Create with return callback
func NewWithReturn(fct ReturnError, code int) Error

// Check if error contains code
func ErrorContainsCode(err error, codes ...CodeError) bool
```

---

## Use Cases

### HTTP API Error Handling

```go
func getUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, liberr.NotFoundError.Error(err)
        }
        return nil, liberr.InternalError.Error(err)
    }
    return user, nil
}

func handleGetUser(c *gin.Context) {
    user, err := getUser(c.Param("id"))
    if err != nil {
        if e, ok := err.(liberr.Error); ok {
            c.JSON(int(e.Code()), gin.H{
                "error": e.Error(),
                "code": e.Code(),
            })
            return
        }
        c.JSON(500, gin.H{"error": "Internal error"})
        return
    }
    c.JSON(200, user)
}
```

### Batch Operation Error Collection

```go
import "github.com/nabbar/golib/errors/pool"

func processBatch(items []Item) error {
    p := pool.New()
    
    for i, item := range items {
        if err := validateItem(item); err != nil {
            p.Add(fmt.Errorf("item %d: %w", i, err))
        }
    }
    
    if p.Len() > 0 {
        return liberr.ValidationError.Error(p.Error())
    }
    
    return nil
}
```

### Service Layer Error Translation

```go
type UserService struct{}

func (s *UserService) CreateUser(data UserData) error {
    // Validation errors
    if err := data.Validate(); err != nil {
        return liberr.ValidationError.Error(err)
    }
    
    // Database errors
    if err := db.Insert(data); err != nil {
        if isDuplicateKey(err) {
            return liberr.ConflictError.Error(err)
        }
        return liberr.InternalError.Error(err)
    }
    
    return nil
}
```

### Multi-Source Error Aggregation

```go
func syncData() error {
    p := pool.New()
    
    var wg sync.WaitGroup
    
    // Sync from multiple sources concurrently
    sources := []string{"api1", "api2", "api3"}
    for _, src := range sources {
        wg.Add(1)
        go func(source string) {
            defer wg.Done()
            if err := syncFrom(source); err != nil {
                p.Add(fmt.Errorf("%s: %w", source, err))
            }
        }(src)
    }
    
    wg.Wait()
    
    if err := p.Error(); err != nil {
        return liberr.ServiceUnavailableError.Error(err)
    }
    
    return nil
}
```

### Error Context Enrichment

```go
func processRequest(req *Request) error {
    err := validateRequest(req)
    if err != nil {
        // Add context to error
        contextErr := liberr.ValidationError.Error(err)
        contextErr.Add(fmt.Errorf("request_id: %s", req.ID))
        contextErr.Add(fmt.Errorf("user_id: %s", req.UserID))
        return contextErr
    }
    
    return handleRequest(req)
}
```

---

## Best Practices

### 1. Use Appropriate Error Codes

```go
// ✅ Good: Use semantic codes
if user == nil {
    return liberr.NotFoundError.Error(nil)
}

if !hasPermission {
    return liberr.ForbiddenError.Error(nil)
}

// ❌ Bad: Generic errors
return errors.New("something went wrong")
```

### 2. Preserve Error Context

```go
// ✅ Good: Wrap with context
if err := db.Query(); err != nil {
    return liberr.InternalError.Error(
        fmt.Errorf("query failed for user %s: %w", userID, err))
}

// ❌ Bad: Lose context
if err := db.Query(); err != nil {
    return liberr.InternalError.Error(nil)
}
```

### 3. Use Error Pool for Collections

```go
// ✅ Good: Thread-safe collection
p := pool.New()
for _, item := range items {
    if err := process(item); err != nil {
        p.Add(err)
    }
}

// ❌ Bad: Manual slice management
var errs []error
for _, item := range items {
    if err := process(item); err != nil {
        errs = append(errs, err) // not thread-safe
    }
}
```

### 4. Check Error Codes Properly

```go
// ✅ Good: Type assertion + check
if e, ok := err.(liberr.Error); ok {
    if e.IsCode(liberr.NotFoundError) {
        // Handle not found
    }
}

// ❌ Bad: String comparison
if strings.Contains(err.Error(), "not found") {
    // Fragile
}
```

### 5. Don't Ignore Stack Traces

```go
// ✅ Good: Use trace in logs
if e, ok := err.(liberr.Error); ok {
    log.Printf("Error at %s:%d: %v", 
        e.GetFile(), e.GetLine(), e)
}

// ⚠️ Missing: Trace available but unused
log.Printf("Error: %v", err)
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd errors
go test -v -cover
cd pool
go test -v -race -cover
```

**Test Metrics:**
- **Main Package**: Comprehensive test suite with Ginkgo/Gomega
- **Pool Sub-Package**: 83 specs, 100% coverage, 0 race conditions

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain backward compatibility
- Follow existing code style

**Testing**
- Write tests for all new features
- Test error code behavior
- Verify stack trace accuracy
- Test error hierarchy operations
- Include race detection tests for concurrent code

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Document all public APIs with GoDoc
- Keep TESTING.md synchronized

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

---

## Related Documentation

### Go Standard Library
- **[errors](https://pkg.go.dev/errors)** - Standard error handling
- **[fmt](https://pkg.go.dev/fmt)** - Error formatting
- **[runtime](https://pkg.go.dev/runtime)** - Stack trace information

### HTTP Status Codes
- **[RFC 7231](https://tools.ietf.org/html/rfc7231)** - HTTP status codes
- **[MDN HTTP Status](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)**

### Related Golib Packages
- **[atomic](../atomic/README.md)** - Used by error pool
- **[gin](https://github.com/gin-gonic/gin)** - Web framework integration

---

## License

MIT License - See [LICENSE](../LICENSE) file for details.

Copyright (c) 2020 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/errors)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../CONTRIBUTING.md)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
