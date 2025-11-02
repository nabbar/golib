# Version Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-173%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-93.8%25-brightgreen)]()

Production-ready version and license management for Go applications with support for 11 open-source licenses, Go version constraints, and automatic package path extraction.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The version package provides comprehensive version and license management for Go applications. It handles build metadata, release versions, license information, and Go version constraints with thread-safe operations and minimal memory overhead.

### Design Philosophy

1. **Immutable After Creation**: Version instances are read-only after construction
2. **Thread-Safe**: All methods are safe for concurrent use
3. **Zero Dependencies**: Only depends on `github.com/nabbar/golib/errors` and `github.com/hashicorp/go-version`
4. **Reflection-Based**: Automatic package path extraction using Go reflection
5. **License Compliant**: Built-in support for major open-source licenses

---

## Key Features

- **Version Management**: Build hash, release tag, date, author, and description
- **License Support**: 11 open-source licenses with full legal text and boilerplate
  - MIT, Apache 2.0, GPL v3, AGPL v3, LGPL v3
  - Mozilla PL v2.0, Unlicense, CC0, CC BY, CC BY-SA
  - SIL Open Font License v1.1
- **Go Version Constraints**: Validate runtime Go version with flexible operators
- **Package Path Extraction**: Automatic detection via reflection
- **Formatted Output**: Headers, info strings, and license text generation
- **Thread-Safe**: Immutable design ensures safe concurrent access
- **High Performance**: ~94ns for version creation, 0 allocations for license retrieval

---

## Installation

```bash
go get github.com/nabbar/golib/version
```

**Requirements**:
- Go ≥ 1.18
- CGO enabled for race detection (optional)

---

## Architecture

### Package Structure

```
version/
├── version.go           # Core Version interface and NewVersion()
├── license.go           # License type and constants
├── license_*.go         # Individual license implementations
├── error.go             # Error codes and messages
└── doc.go              # Package documentation
```

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                   Version Interface                      │
│  GetRelease(), GetBuild(), GetDate(), GetAuthor()       │
│  GetLicense*(), CheckGo(), PrintInfo()                  │
└──────────────┬──────────────┬──────────────┬────────────┘
               │              │              │
      ┌────────▼─────┐  ┌────▼─────┐  ┌────▼────────┐
      │   Metadata   │  │ Licenses │  │ Validation  │
      │              │  │          │  │             │
      │ Build info   │  │ 11 types │  │ Go version  │
      │ Package path │  │ Legal    │  │ constraints │
      └──────────────┘  └──────────┘  └─────────────┘
```

### Data Flow

```
NewVersion() → Reflection → Package Path Extraction
     │              │
     ├──────────────┴──→ Parse Date (RFC3339)
     │
     └──→ Create immutable Version instance
              │
              ├──→ GetHeader() / GetInfo()
              ├──→ GetLicense*()
              └──→ CheckGo()
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/version"
)

// Define a struct for reflection-based package path extraction
type MyApp struct{}

func main() {
    // Create version instance
    v := version.NewVersion(
        version.License_MIT,           // License type
        "MyApp",                        // Package name
        "My Application Description",  // Description
        "2024-01-15T10:30:00Z",        // Build date (RFC3339)
        "abc123def",                    // Build hash (git commit)
        "v1.2.3",                       // Release version
        "John Doe",                     // Author
        "MYAPP",                        // Prefix for env vars
        MyApp{},                        // Empty struct for reflection
        0,                              // Number of parent packages
    )

    // Display version information
    fmt.Println(v.GetHeader())
    // Output: MyApp (Release: v1.2.3, Build: abc123def, Date: Mon, 15 Jan 2024 10:30:00 UTC)

    // Check Go version constraint
    if err := v.CheckGo("1.18", ">="); err != nil {
        panic(err)
    }

    // Get license information
    fmt.Println(v.GetLicenseName())    // "MIT License"
    fmt.Println(v.GetLicenseBoiler())  // Boilerplate for file headers
}
```

### Build Integration

Create a `release` package to inject build-time variables:

**release/release.go**:
```go
package release

type EmptyStruct struct{}

var (
    // Injected at build time with -ldflags
    Release     = "v0.0.0"
    Build       = "dev"
    Date        = "2024-01-01T00:00:00Z"
    Package     = "myapp"
    Description = "My Application"
    Author      = "Author Name"
    Prefix      = "MYAPP"
)
```

**release/version.go**:
```go
package release

import "github.com/nabbar/golib/version"

var vers version.Version

func init() {
    vers = version.NewVersion(
        version.License_MIT,
        Package,
        Description,
        Date,
        Build,
        Release,
        Author,
        Prefix,
        EmptyStruct{},
        1, // Go up one directory from release/ to app root
    )
}

func GetVersion() version.Version {
    return vers
}
```

**Build command**:
```bash
go build -ldflags "\
  -X release.Release=$(git describe --tags HEAD) \
  -X release.Build=$(git rev-parse --short HEAD) \
  -X release.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -X release.Package=$(basename $(pwd)) \
  -X release.Description='My Application' \
  -X release.Author='John Doe'"
```

---

## Performance

### Benchmark Results

```
Operation                    Time/op    Memory/op   Allocs/op
─────────────────────────────────────────────────────────────
NewVersion                   94.00 ns   144 B       1
GetHeader                   383.3 ns    272 B       8
GetInfo                     296.1 ns    160 B       5
GetLicenseLegal              19.50 ns     0 B       0
GetLicenseBoiler            817.0 ns   1201 B       5
GetLicenseFull              4.088 µs  12384 B      11
CheckGo                     4.749 µs   1982 B      27
ConcurrentAccess            331.3 ns    432 B      13
```

### Performance Characteristics

- **Version Creation**: Constant time O(1), single allocation
- **License Retrieval**: Zero allocations for legal text (pre-compiled strings)
- **Thread-Safe**: No locks needed due to immutable design
- **Memory Efficient**: ~144 bytes per version instance
- **Concurrent Access**: Linear scaling with CPU cores

---

## Use Cases

### 1. Application Version Display

```go
v := release.GetVersion()
fmt.Printf("Starting %s\n", v.GetHeader())
// Output: Starting MyApp (Release: v1.2.3, Build: abc123, Date: ...)
```

### 2. License File Generation

```go
v := release.GetVersion()
licenseText := v.GetLicenseFull()
os.WriteFile("LICENSE", []byte(licenseText), 0644)
```

### 3. Go Version Validation

```go
v := release.GetVersion()

// Require Go >= 1.18
if err := v.CheckGo("1.18", ">="); err != nil {
    log.Fatal("This application requires Go 1.18 or later")
}

// Pessimistic constraint (allow patch updates)
if err := v.CheckGo("1.20", "~>"); err != nil {
    log.Fatal("This application requires Go 1.20.x")
}
```

### 4. Multi-License Projects

```go
v := version.NewVersion(version.License_MIT, ...)

// Combine multiple licenses
combined := v.GetLicenseFull(
    version.License_Apache_v2,
    version.License_Mozilla_PL_v2,
)
```

### 5. File Header Boilerplate

```go
v := release.GetVersion()
header := v.GetLicenseBoiler()

// Add to source files
fileContent := header + "\n\npackage main\n\n// ..."
```

---

## API Reference

### Version Interface

```go
type Version interface {
    // Version information
    GetPackage() string
    GetDescription() string
    GetRelease() string
    GetBuild() string
    GetDate() string
    GetTime() time.Time
    GetAuthor() string
    GetPrefix() string
    GetRootPackagePath() string
    
    // Formatted output
    GetHeader() string
    GetInfo() string
    GetAppId() string
    
    // License methods
    GetLicenseName() string
    GetLicenseLegal(addMoreLicence ...license) string
    GetLicenseBoiler(addMoreLicence ...license) string
    GetLicenseFull(addMoreLicence ...license) string
    
    // Validation
    CheckGo(RequireGoVersion, RequireGoContraint string) liberr.Error
    
    // Output methods
    PrintInfo()
    PrintLicense(addlicence ...license)
}
```

### License Constants

```go
const (
    License_MIT
    License_Apache_v2
    License_GNU_GPL_v3
    License_GNU_Affero_GPL_v3
    License_GNU_Lesser_GPL_v3
    License_Mozilla_PL_v2
    License_Unlicense
    License_Creative_Common_Zero_v1
    License_Creative_Common_Attribution_v4_int
    License_Creative_Common_Attribution_Share_Alike_v4_int
    License_SIL_Open_Font_1_1
)
```

### Go Version Constraint Operators

- `==` : Exact version match
- `!=` : Not equal
- `>` : Greater than
- `>=` : Greater than or equal
- `<` : Less than
- `<=` : Less than or equal
- `~>` : Pessimistic constraint (allows patch-level changes)

### Error Codes

```go
const (
    ErrorParamEmpty          // Required parameter is empty
    ErrorGoVersionInit       // Failed to parse version constraint
    ErrorGoVersionRuntime    // Failed to extract runtime Go version
    ErrorGoVersionConstraint // Go version constraint not satisfied
)
```

---

## Best Practices

### 1. Use Build-Time Injection

Always inject version information at build time rather than hardcoding:

```bash
go build -ldflags "-X package.Variable=value"
```

### 2. Validate Go Version Early

Check Go version constraints during application initialization:

```go
func init() {
    if err := release.GetVersion().CheckGo("1.18", ">="); err != nil {
        log.Fatal(err)
    }
}
```

### 3. Single Version Instance

Create one version instance per application (singleton pattern):

```go
var version = version.NewVersion(...)  // Package-level variable

func GetVersion() version.Version {
    return version
}
```

### 4. Package Path Extraction

Use `numSubPackage` to navigate from nested packages:

```go
// In myapp/internal/release/version.go
v := version.NewVersion(..., EmptyStruct{}, 2)
// Goes up 2 levels: internal/release → internal → myapp
```

### 5. License Selection

Choose the appropriate license for your project:

- **Permissive**: MIT, Apache 2.0, Unlicense
- **Copyleft (Strong)**: GPL v3, AGPL v3
- **Copyleft (Weak)**: LGPL v3, Mozilla PL v2.0
- **Creative Commons**: CC0, CC BY, CC BY-SA (for content/docs)
- **Fonts**: SIL Open Font License

---

## Testing

**Test Suite**: 173 specs using Ginkgo v2 and Gomega (93.8% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Coverage Areas**
- Version creation and metadata management
- License operations (all 11 licenses)
- Go version constraint validation
- Error handling and edge cases
- Thread safety and concurrent access

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Immutable design pattern
- ✅ High test coverage (93.8%)

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥90%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Additional Licenses**
- BSD 2-Clause / 3-Clause
- ISC License
- Boost Software License

**Enhanced Metadata**
- Git branch information
- CI/CD build numbers
- Semantic version parsing

**Output Formats**
- JSON/YAML version info export
- SPDX license identifiers
- Machine-readable metadata

**Validation**
- Semantic version validation
- License compatibility checking
- Dependency version tracking

**Integration**
- GitHub Actions integration
- Docker label generation
- Kubernetes annotation support

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/version)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Error Handling**: [github.com/nabbar/golib/errors](../errors)
- **Version Constraints**: [github.com/hashicorp/go-version](https://github.com/hashicorp/go-version)
- **License Information**: [choosealicense.com](https://choosealicense.com/)
- **Semantic Versioning**: [semver.org](https://semver.org/)
