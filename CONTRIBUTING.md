# Contributing to golib

Thank you for your interest in contributing to golib! This document provides guidelines and instructions for contributing to the project.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Guidelines](#development-guidelines)
- [AI Usage Policy](#ai-usage-policy)
- [Testing Requirements](#testing-requirements)
- [Documentation Standards](#documentation-standards)
- [Pull Request Process](#pull-request-process)
- [Code Review](#code-review)
- [Release Process](#release-process)

---

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Focus on constructive feedback
- Accept criticism gracefully
- Prioritize project and community benefit
- Show empathy towards other contributors

### Unacceptable Behavior

- Harassment or discriminatory language
- Trolling or inflammatory comments
- Personal or political attacks
- Publishing private information
- Other conduct inappropriate in a professional setting

---

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Git
- CGO enabled (for race detection tests)
- Basic understanding of Go best practices

### Fork and Clone

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/golib.git
cd golib

# Add upstream remote
git remote add upstream https://github.com/nabbar/golib.git

# Verify remotes
git remote -v
```

### Set Up Development Environment

```bash
# Install dependencies
go mod download

# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run tests to verify setup
go test ./...

# Run with race detection
CGO_ENABLED=1 go test -race ./...
```

---

## Development Guidelines

### Code Style

Follow standard Go conventions:

- Use `gofmt` to format code
- Use `golint` or `golangci-lint` for linting
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Keep functions small and focused
- Use meaningful variable names
- Add comments for exported functions and types

```bash
# Format code
gofmt -w .

# Run linter
golangci-lint run ./...
```

### Package Organization

- Each package should be self-contained
- Avoid circular dependencies
- Keep internal implementation details private
- Export only necessary interfaces and types
- Place interfaces close to where they're used

### Error Handling

```go
// ✅ Good - Wrap errors with context
func Process(input string) error {
    data, err := Parse(input)
    if err != nil {
        return fmt.Errorf("failed to parse input: %w", err)
    }
    return nil
}

// ❌ Bad - Lose error context
func Process(input string) error {
    data, err := Parse(input)
    if err != nil {
        return err  // Lost context
    }
    return nil
}

// ❌ Bad - Ignore errors
func Process(input string) error {
    data, _ := Parse(input)  // Never ignore errors
    return nil
}
```

### Concurrency

- Always test concurrent code with `-race`
- Use proper synchronization (mutex, channels, atomic)
- Document thread-safety guarantees
- Avoid shared mutable state when possible

```go
// ✅ Good - Proper synchronization
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// ❌ Bad - Race condition
type UnsafeCounter struct {
    count int
}

func (c *UnsafeCounter) Increment() {
    c.count++  // Data race!
}
```

---

## AI Usage Policy

### ⚠️ Important: Restricted AI Usage

In accordance with project guidelines and EU AI Act Article 50.4:

#### ❌ AI Must NOT Be Used For:

- **Package implementation code**
- **Core functionality**
- **Algorithm implementation**
- **Business logic**
- **Security-critical code**

#### ✅ AI May Assist With:

- **Writing tests** - Unit tests, integration tests, test cases
- **Documentation** - README files, code comments, API documentation
- **Bug fixing** - Analyzing bugs, suggesting fixes (under human review)
- **Code examples** - Usage examples in documentation
- **Test data generation** - Creating test fixtures

#### Human Supervision Required

- All AI-assisted content must be reviewed by a human
- Contributors are responsible for AI-generated content
- AI suggestions must be validated for correctness
- Security implications must be manually verified

#### Disclosure

When using AI assistance for allowed tasks:
- Document AI usage in commit messages if significant
- Review and validate all AI-generated content
- Take responsibility for the final output

**Rationale**: This policy ensures code quality, security, and compliance with the EU AI Act while allowing AI to assist with non-critical tasks.

---

## Testing Requirements

### Minimum Requirements

All contributions must:

1. **Pass all existing tests**
   ```bash
   go test ./...
   ```

2. **Pass race detection**
   ```bash
   CGO_ENABLED=1 go test -race ./...
   ```

3. **Maintain or improve coverage** (target: ≥80%)
   ```bash
   go test -cover ./...
   ```

4. **Include tests for new features**
   - Unit tests for all new functions
   - Integration tests for complex features
   - Edge case testing

### Test Quality Standards

```go
// ✅ Good test
var _ = Describe("NewFeature", func() {
    var subject FeatureType
    
    BeforeEach(func() {
        subject = NewFeature()
    })
    
    AfterEach(func() {
        subject.Close()
    })
    
    Context("When input is valid", func() {
        It("should process successfully", func() {
            result, err := subject.Process(validInput)
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(expected))
        })
    })
    
    Context("When input is invalid", func() {
        It("should return error", func() {
            _, err := subject.Process(invalidInput)
            Expect(err).To(HaveOccurred())
        })
    })
    
    Context("When used concurrently", func() {
        It("should be thread-safe", func() {
            var wg sync.WaitGroup
            for i := 0; i < 100; i++ {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    subject.Process(validInput)
                }()
            }
            wg.Wait()
        })
    })
})
```

### Test Documentation

- Use descriptive test names
- Follow BDD style with Ginkgo
- Test edge cases and error conditions
- Document complex test scenarios
- Keep tests maintainable

See [TESTING.md](TESTING.md) for complete testing guidelines.

---

## Documentation Standards

### Required Documentation

#### Package Documentation

Every package must have:
- `README.md` with overview, examples, and API reference
- `TESTING.md` if testing is complex
- GoDoc comments for all exported types and functions

#### README.md Structure

```markdown
# Package Name

Brief description

## Features
## Installation
## Quick Start
## API Reference
## Examples
## Testing
## Contributing
## License
```

#### Code Comments

```go
// ✅ Good - Explain why and what
// NewClient creates a client with the provided configuration.
// It returns an error if the configuration is invalid or if
// the connection cannot be established.
func NewClient(config Config) (*Client, error) {
    // Validate configuration before proceeding
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    // ...
}

// ❌ Bad - State the obvious
// NewClient creates a new client
func NewClient(config Config) (*Client, error) {
    // Create client
    c := &Client{}
    // Return client
    return c, nil
}
```

### Documentation Language

- **Use English** for all documentation
- **Use English** for all code comments
- **Use English** for commit messages
- Write clear, concise descriptions
- Use proper grammar and spelling

### Examples

Provide working code examples:

```go
// Example of basic usage
func ExampleClient() {
    client := NewClient(DefaultConfig())
    defer client.Close()
    
    result, err := client.Query("SELECT * FROM users")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d users\n", len(result))
}
```

---

## Pull Request Process

### Before Submitting

1. **Update your fork**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

3. **Make your changes**
   - Write code following guidelines
   - Add tests
   - Update documentation

4. **Run tests**
   ```bash
   go test ./...
   CGO_ENABLED=1 go test -race ./...
   go test -cover ./...
   ```

5. **Commit changes**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or fixes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Examples**:
```
feat(logger): add syslog hook support

Add support for logging to syslog with configurable
facility and severity levels.

Closes #123
```

```
fix(archive): handle empty tar archives

Previously crashed on empty archives. Now returns
appropriate error message.
```

### Submitting Pull Request

1. **Push to your fork**
   ```bash
   git push origin feature/my-feature
   ```

2. **Create Pull Request on GitHub**
   - Use descriptive title
   - Reference related issues
   - Describe changes clearly
   - Include test results

3. **PR Description Template**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Motivation
   Why this change is needed
   
   ## Changes
   - List of changes made
   - Impact on existing functionality
   
   ## Testing
   - [ ] All tests pass
   - [ ] Race detection clean
   - [ ] Coverage maintained
   - [ ] New tests added
   
   ## Documentation
   - [ ] README updated
   - [ ] TESTING.md updated if needed
   - [ ] Code comments added
   
   ## Related Issues
   Closes #123
   Related to #456
   ```

---

## Code Review

### Review Process

1. **Automated checks** run on PR submission
2. **Maintainer review** of code quality and design
3. **Request changes** if necessary
4. **Approval** when requirements met
5. **Merge** by maintainer

### Review Criteria

- Code follows guidelines
- Tests are comprehensive
- Documentation is complete
- No race conditions
- Coverage maintained
- Changes are minimal and focused
- Commit messages are clear

### Responding to Feedback

- Address all comments
- Make requested changes
- Push updates to same branch
- Notify reviewers when ready
- Be open to suggestions

---

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features, backward compatible
- **Patch** (0.0.X): Bug fixes, backward compatible

### Release Checklist

1. All tests pass
2. Documentation updated
3. CHANGELOG updated
4. Version tagged
5. GitHub release created
6. Go module proxy updated

---

## Getting Help

### Resources

- **Documentation**: [README.md](README.md)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Security Policy**: [SECURITY.md](SECURITY.md)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Discussions**: [GitHub Discussions](https://github.com/nabbar/golib/discussions)

### Asking Questions

When asking for help:

1. Search existing issues first
2. Provide minimal reproducible example
3. Include Go version and OS
4. Share relevant error messages
5. Describe expected vs actual behavior

### Reporting Bugs

Use the bug report template:

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step one
2. Step two
3. ...

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- Go version: 1.22
- OS: Ubuntu 22.04
- Package version: v1.2.3

## Additional Context
Any other relevant information
```

---

## License

By contributing to golib, you agree that your contributions will be licensed under the MIT License.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision. Contributors using AI tools must comply with the [AI Usage Policy](#ai-usage-policy).

---

## Thank You!

Thank you for contributing to golib! Your efforts help make this library better for everyone.

**Maintained by**: golib Contributors
