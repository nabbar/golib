# Testing Guide - Cobra Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://golang.org/)

Comprehensive testing documentation for the cobra package.

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
- [AI Transparency Notice](#ai-transparency-notice)
- [Resources](#resources)

---

## Overview

The cobra package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) to achieve comprehensive test coverage across all functionality.

### Test Philosophy

1. **Behavior-Driven** - Tests describe behavior in readable, hierarchical structures
2. **Integration Focus** - Test CLI behavior as users would experience it
3. **Instance Isolation** - Each test creates independent cobra instances
4. **Type Coverage** - Verify all 20+ flag types work correctly
5. **Command Testing** - Validate built-in and custom commands

### Coverage Scope

- ✅ CLI initialization and configuration
- ✅ Flag management (20+ types, persistent and local)
- ✅ Built-in commands (completion, configure, error printing)
- ✅ Custom command integration
- ✅ Version management
- ✅ Logger and viper integration
- ✅ Error handling and edge cases

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
# Run all tests
ginkgo

# Verbose output
ginkgo -v

# With coverage
ginkgo -cover

# Parallel execution
ginkgo -p

# Watch mode (re-run on changes)
ginkgo watch
```

### Coverage Reports

**Generate HTML Coverage Report**
```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

**View Coverage by Function**
```bash
go test -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Advanced Options

**Focus on Specific Tests**
```bash
# Run tests matching pattern
ginkgo --focus="Flag"

# Skip tests matching pattern
ginkgo --skip="Completion"
```

**Output Formats**
```bash
# JUnit XML report (CI integration)
ginkgo --junit-report=test-results.xml

# JSON output
ginkgo --json-report=test-results.json
```

---

## Test Coverage

### Coverage Metrics

| Test File | Specs | Coverage | Description |
|-----------|-------|----------|-------------|
| `commands_test.go` | 25 | ~80% | Built-in commands |
| `flags_test.go` | 30 | ~85% | Flag management |
| `configuration_test.go` | 20 | ~70% | Config generation |
| `completion_test.go` | 15 | ~65% | Shell completion |
| `initialization_test.go` | 10 | ~75% | CLI initialization |
| `error_printing_test.go` | 10 | ~70% | Error handling |
| **Total** | **110** | **~75%** | **All tests** |

### Coverage by Component

**Initialization Tests** (`initialization_test.go` - 10 specs)
- Cobra instance creation
- Version setting
- Logger injection
- Viper configuration
- Force no info mode
- Initialization function execution

**Flag Tests** (`flags_test.go` - 30 specs)
- Config flag (`SetFlagConfig`)
- Verbose flag (`SetFlagVerbose`)
- String flags
- Integer flags (int, int8, int16, int32, int64)
- Unsigned integer flags (uint, uint8, uint16, uint32, uint64)
- Floating-point flags (float32, float64)
- Boolean flags
- Duration flags
- Count flags
- IP address flags (IP, IPMask, IPNet)
- Slice flags (StringSlice, IntSlice)
- Persistent vs local flags

**Command Tests** (`commands_test.go` - 25 specs)
- Adding custom commands
- Command execution
- Subcommand hierarchies
- Command flags
- Pre/post run hooks
- Command context

**Built-in Commands** (`completion_test.go`, `configuration_test.go`, `error_printing_test.go`)
- Completion command generation (bash, zsh, fish, PowerShell)
- Configuration file generation (JSON, YAML, TOML)
- Error code printing
- Output file handling
- Format validation

---

## Test Organization

### File Structure

```
cobra/
├── cobra_suite_test.go          # Test suite setup
├── initialization_test.go       # Init and config tests (10 specs)
├── flags_test.go                # Flag management tests (30 specs)
├── commands_test.go             # Command tests (25 specs)
├── completion_test.go           # Completion tests (15 specs)
├── configuration_test.go        # Config generation tests (20 specs)
└── error_printing_test.go       # Error handling tests (10 specs)
```

### Test Structure Pattern

```go
// Hierarchical BDD structure
var _ = Describe("cobra/Feature", func() {
    var (
        app     libcbr.Cobra
        version libver.Version
    )
    
    BeforeEach(func() {
        // Setup for each test
        version = createTestVersion()
        app = libcbr.New()
        app.SetVersion(version)
    })
    
    Context("When configuring feature", func() {
        It("should configure successfully", func() {
            // Arrange
            var flagValue string
            
            // Act
            app.AddFlagString(true, &flagValue, "flag", "f", "default", "Usage")
            app.Init()
            
            // Assert
            Expect(app).ToNot(BeNil())
            Expect(flagValue).To(Equal("default"))
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
It("should add persistent string flag with default value", func() {
    // Test implementation
})

// ❌ Bad: Vague description
It("should work", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should set verbose flag", func() {
    // Arrange
    var verbose int
    app := libcbr.New()
    app.SetVersion(version)
    
    // Act
    app.SetFlagVerbose(true, &verbose)
    app.Init()
    
    // Assert
    Expect(app).ToNot(BeNil())
})
```

**3. Test Instance Isolation**
```go
// ✅ Good: Each test creates its own instance
It("test 1", func() {
    app1 := libcbr.New()
    app1.SetVersion(version)
    // Use app1
})

It("test 2", func() {
    app2 := libcbr.New()
    app2.SetVersion(version)
    // Use app2 - no interference with test 1
})
```

**4. Test All Flag Types**
```go
Describe("Flag Types", func() {
    It("should handle string flags", func() {
        var str string
        app.AddFlagString(true, &str, "str", "s", "default", "String flag")
        app.Init()
        Expect(str).To(Equal("default"))
    })
    
    It("should handle int flags", func() {
        var i int
        app.AddFlagInt(true, &i, "int", "i", 42, "Int flag")
        app.Init()
        Expect(i).To(Equal(42))
    })
    
    // Test all 20+ types...
})
```

**5. Test Built-in Commands**
```go
It("should generate bash completion", func() {
    app.AddCommandCompletion()
    app.Init()
    
    cmd := app.GetCommand()
    Expect(cmd).ToNot(BeNil())
    
    // Verify completion command exists
    hasCompletion := false
    for _, c := range cmd.Commands() {
        if c.Use == "completion" {
            hasCompletion = true
            break
        }
    }
    Expect(hasCompletion).To(BeTrue())
})
```

### Test Template

```go
var _ = Describe("cobra/NewFeature", func() {
    var (
        app     libcbr.Cobra
        version libver.Version
    )
    
    BeforeEach(func() {
        version = libver.NewVersion(
            libver.License_MIT,
            "testapp",
            "Test Application",
            "2024-01-01",
            "abc123",
            "v1.0.0",
            "Test Author",
            "testapp",
            struct{}{},
            0,
        )
        
        app = libcbr.New()
        app.SetVersion(version)
    })
    
    Context("When using feature", func() {
        It("should behave as expected", func() {
            // Arrange
            var config string
            
            // Act
            app.SetFlagConfig(true, &config)
            app.Init()
            
            // Assert
            Expect(app).ToNot(BeNil())
        })
        
        It("should handle edge case", func() {
            // Test edge case
            app.Init()
            cmd := app.GetCommand()
            Expect(cmd).ToNot(BeNil())
        })
    })
})
```

---

## Best Practices

### Test Independence

```go
// ✅ Good: Each test is independent
It("test 1", func() {
    app := libcbr.New()
    app.SetVersion(version)
    app.Init()
})

It("test 2", func() {
    app := libcbr.New()
    app.SetVersion(version)
    app.Init()
})

// ❌ Bad: Tests share state
var sharedApp libcbr.Cobra

It("test 1", func() {
    sharedApp.SetVersion(version)
})

It("test 2", func() {
    // Depends on test 1!
    sharedApp.Init()
})
```

### Version Creation

```go
// ✅ Good: Reusable helper
func createTestVersion() libver.Version {
    return libver.NewVersion(
        libver.License_MIT,
        "testapp",
        "Test App",
        "2024-01-01",
        "abc123",
        "v1.0.0",
        "Author",
        "testapp",
        struct{}{},
        0,
    )
}

// Use in BeforeEach
BeforeEach(func() {
    version = createTestVersion()
    app = libcbr.New()
    app.SetVersion(version)
})
```

### Testing Custom Commands

```go
It("should execute custom command", func() {
    // Arrange
    executed := false
    
    cmd := &spfcbr.Command{
        Use: "test",
        RunE: func(cmd *spfcbr.Command, args []string) error {
            executed = true
            return nil
        },
    }
    
    // Act
    app.AddCommand(cmd)
    app.Init()
    
    // Assert
    rootCmd := app.GetCommand()
    Expect(rootCmd.Commands()).ToNot(BeEmpty())
})
```

### Testing Logger Integration

```go
It("should integrate with logger", func() {
    // Arrange
    loggerCalled := false
    
    app.SetLogger(func() liblog.Logger {
        loggerCalled = true
        return liblog.New(context.Background)
    })
    
    // Act
    app.Init()
    
    // Assert - logger function should be stored
    Expect(app).ToNot(BeNil())
})
```

### Testing Error Handling

```go
It("should handle initialization errors gracefully", func() {
    // Arrange
    app := libcbr.New()
    // Don't set version
    
    // Act & Assert
    // Should not panic
    Expect(func() {
        app.Init()
    }).ToNot(Panic())
})
```

---

## Troubleshooting

### Common Issues

**Test Panics**

*Problem*: Tests panic during execution.

*Solution*: Ensure version is set before Init().
```go
// ✅ Correct
app.SetVersion(version)  // Set version first
app.Init()               // Then initialize
```

**Flag Not Set**

*Problem*: Flag value remains at zero value.

*Solution*: Initialize app before checking flag value.
```go
var flag string
app.AddFlagString(true, &flag, "flag", "f", "default", "Usage")
app.Init()  // ✅ Initialize first
Expect(flag).To(Equal("default"))
```

**Command Not Found**

*Problem*: Added command not found in command list.

*Solution*: Add commands after Init().
```go
app.Init()
app.AddCommand(myCmd)  // ✅ Add after Init
```

**Stale Coverage**

*Problem*: Coverage report doesn't reflect changes.

*Solution*:
```bash
go clean -testcache
go test -coverprofile=coverage.out
```

### Debugging Techniques

**Run Specific Tests**
```bash
# Focus on specific test
ginkgo --focus="Flag" -v

# Skip specific tests
ginkgo --skip="Completion"
```

**Verbose Output**
```bash
# Ginkgo verbose mode
ginkgo -v --trace

# Standard Go verbose
go test -v
```

**Debug Output in Tests**
```go
It("should do something", func() {
    app := libcbr.New()
    app.SetVersion(version)
    app.Init()
    
    // Debug output
    fmt.Fprintf(GinkgoWriter, "App: %+v\n", app)
    fmt.Fprintf(GinkgoWriter, "Command: %+v\n", app.GetCommand())
    
    Expect(app).ToNot(BeNil())
})
```

**Check Test Order**
```bash
# Run tests in random order to detect dependencies
ginkgo --randomize-all
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
4. Ensure test independence
5. Test all code paths
6. Verify error handling

**Test Review Checklist**
- [ ] Tests are independent
- [ ] Resources are properly managed
- [ ] Edge cases are covered
- [ ] All flag types tested
- [ ] Built-in commands verified
- [ ] Descriptions are clear
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
- [spf13/cobra Testing](https://cobra.dev/)

**Cobra Package**
- [Package GoDoc](https://pkg.go.dev/github.com/nabbar/golib/cobra)
- [README.md](README.md) - Package overview and examples
- [GitHub Repository](https://github.com/nabbar/golib)

**Testing Tools**
- [Go Test Command](https://pkg.go.dev/cmd/go#hdr-Test_packages)
- [Coverage Tool](https://go.dev/blog/cover)
- [Race Detector](https://go.dev/doc/articles/race_detector)

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.
