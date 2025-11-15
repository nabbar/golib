# Artifact Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Unified Go library for artifact version management across multiple platforms (GitHub, GitLab, JFrog Artifactory, AWS S3).

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Supported Platforms](#supported-platforms)
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

This package provides a unified interface for managing software artifacts across different platforms. It abstracts version discovery, filtering, and retrieval operations into a consistent API.

### Design Philosophy

1. **Platform Abstraction**: Single interface for multiple repositories
2. **Semantic Versioning**: Built on `hashicorp/go-version`
3. **Production Focus**: Automatic pre-release filtering
4. **Flexible Matching**: Substring and regex-based artifact search
5. **Test-Friendly**: Zero external API calls in tests

---

## Key Features

- **Multi-Platform Support**: GitHub, GitLab, JFrog Artifactory, AWS S3
- **Semantic Versioning**: Robust version parsing and comparison
- **Pre-Release Filtering**: Automatic exclusion of alpha/beta/rc/dev versions
- **Flexible Matching**: Substring and regex-based artifact search
- **Version Organization**: Query by major/minor version numbers
- **Streaming Downloads**: Direct artifact retrieval without intermediate storage
- **Pagination**: Automatic handling of large result sets
- **Context Support**: Cancellation and timeout management

---

## Installation

```bash
go get github.com/nabbar/golib/artifact
```

---

## Architecture

The package follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────┐
│              artifact.Client Interface              │
│  (ListReleases, GetArtifact, Download, etc.)        │
└─────────────────────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌────────────────┐ ┌──────────────┐ ┌──────────────┐
│   github/      │ │   gitlab/    │ │   jfrog/     │
│   s3aws/       │ │              │ │              │
└────────────────┘ └──────────────┘ └──────────────┘
         │               │               │
         └───────────────┼───────────────┘
                         ▼
         ┌───────────────────────────────┐
         │    client.ArtHelper           │
         │ (Helper - version organizing) │
         └───────────────────────────────┘
```

### Core Components

#### 1. **Client Interface** (`artifact.Client`)
The main interface exposing:
- `ListReleases()`: Retrieve all available versions
- `GetArtifact()`: Get download URL for specific artifact
- `Download()`: Stream artifact content directly
- Version management methods (via embedded `client.ArtHelper`)

#### 2. **Helper** (`client.Helper`)
Internal component providing version organization:
- Groups versions by major/minor numbers
- Retrieves latest versions by criteria
- Maintains sorted collections for efficient lookups

#### 3. **Platform Implementations**
- **github**: Uses `google/go-github` SDK
- **gitlab**: Uses `gitlab-org/api/client-go` SDK
- **jfrog**: Custom HTTP client with Artifactory Storage API
- **s3aws**: Integration with AWS S3 via custom AWS wrapper

#### 4. **Helper Functions** (`artifact` package)
- `CheckRegex()`: Validate artifact names against patterns
- `ValidatePreRelease()`: Filter out alpha/beta/rc versions

---

## Quick Start

### GitHub Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    "github.com/nabbar/golib/artifact/github"
)

func main() {
    ctx := context.Background()
    
    // Create GitHub artifact client
    client, err := github.NewGithub(ctx, &http.Client{}, "owner/repository")
    if err != nil {
        panic(err)
    }

    // List all stable versions
    versions, err := client.ListReleases()
    if err != nil {
        panic(err)
    }

    // Get latest version
    latest, err := client.GetLatest()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Latest version: %s\n", latest.String())

    // Get download URL for specific artifact
    url, err := client.GetArtifact("linux-amd64", "", latest)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Download URL: %s\n", url)
}
```

### GitLab Example

```go
package main

import (
    "context"
    "github.com/nabbar/golib/artifact/gitlab"
)

func main() {
    ctx := context.Background()
    
    // Create GitLab artifact client (projectID is an integer)
    client, err := gitlab.NewGitlab(ctx, "https://gitlab.com", "your-token", 12345)
    if err != nil {
        panic(err)
    }

    // Get latest version from major version 2
    latest, err := client.GetLatestMajor(2)
    if err != nil {
        panic(err)
    }
    
    // Download artifact
    size, reader, err := client.Download("darwin", "", latest)
    if err != nil {
        panic(err)
    }
    defer reader.Close()
    
    // Process artifact (size in bytes, reader is io.ReadCloser)
    fmt.Printf("Downloading %d bytes\n", size)
}
```

### JFrog Artifactory Example

```go
package main

import (
    "context"
    "net/http"
    "github.com/nabbar/golib/artifact/jfrog"
)

func main() {
    ctx := context.Background()
    client := &http.Client{}
    
    // Create JFrog client with regex to extract version
    // regex must have at least one capturing group for version
    art, err := jfrog.New(
        ctx,
        client.Do,
        "https://artifactory.example.com",
        []string{"repo-name", "path", "to", "artifacts"},
        `artifact-(\d+\.\d+\.\d+)\.tar\.gz`,  // Regex with version capture group
        1,  // Group index for version
    )
    if err != nil {
        panic(err)
    }

    // List versions matching the regex
    versions, err := art.ListReleases()
    if err != nil {
        panic(err)
    }
}
```

### AWS S3 Example

```go
package main

import (
    "context"
    "github.com/nabbar/golib/artifact/s3aws"
    "github.com/nabbar/golib/aws"  // Custom AWS wrapper
)

func main() {
    ctx := context.Background()
    
    // Configure AWS client (using custom golib/aws package)
    awsClient, err := aws.New(/* AWS configuration */)
    if err != nil {
        panic(err)
    }
    
    // Create S3 artifact client
    s3Client, err := s3aws.New(
        ctx,
        awsClient,
        `releases/myapp-v(\d+\.\d+\.\d+)\.zip`,  // Regex with version capture
        1,  // Group index for version
    )
    if err != nil {
        panic(err)
    }

    // Get latest artifact
    latest, err := s3Client.GetLatest()
    if err != nil {
        panic(err)
    }
}
```

---

## Supported Platforms

| Platform | Package | Authentication | Notes |
|----------|---------|----------------|-------|
| **GitHub** | `artifact/github` | Optional token via HTTP client | Public repos work without auth |
| **GitLab** | `artifact/gitlab` | Token required | Self-hosted and GitLab.com |
| **JFrog Artifactory** | `artifact/jfrog` | Credentials via HTTP client | Uses Storage API |
| **AWS S3** | `artifact/s3aws` | AWS credentials | Requires `golib/aws` wrapper |

---

## Use Cases

### 1. Automated Deployment Systems
Check for new versions and deploy automatically:

```go
latest, _ := client.GetLatest()
deployed, _ := version.NewVersion(getCurrentDeployedVersion())

if latest.GreaterThan(deployed) {
    url, _ := client.GetArtifact("linux-amd64", "", latest)
    downloadAndDeploy(url)
}
```

### 2. Version Pinning
Lock to specific major/minor versions:

```go
// Always use latest patch of v2.1.x
v21Latest, _ := client.GetLatestMinor(2, 1)
```

### 3. Multi-Platform Binary Distribution
Retrieve platform-specific artifacts:

```go
platforms := []string{"linux-amd64", "darwin-arm64", "windows-amd64.exe"}
for _, platform := range platforms {
    url, err := client.GetArtifact(platform, "", targetVersion)
    if err != nil {
        // Handle missing platform
        continue
    }
    downloadBinary(url, platform)
}
```

### 4. Artifact Mirroring
Download from one platform and upload to another:

```go
size, reader, _ := sourceClient.Download("", `.*\.tar\.gz$`, version)
defer reader.Close()

// Upload to internal storage
uploadToInternalRepository(reader, size)
```

### 5. Version Auditing
Generate reports of available versions:

```go
releases, _ := client.ListReleases()
for _, v := range releases {
    fmt.Printf("%s - %s\n", v.String(), getReleaseDat(v))
}
```

---

## API Reference

### artifact.Client Interface

```go
type Client interface {
    // ListReleases returns all stable versions (pre-releases filtered)
    ListReleases() (hscvrs.Collection, error)
    
    // GetArtifact retrieves download URL
    // containName: substring match in artifact name
    // regexName: regex pattern match (takes precedence)
    GetArtifact(containName, regexName string, release *hscvrs.Version) (string, error)
    
    // Download streams artifact content
    Download(containName, regexName string, release *hscvrs.Version) (int64, io.ReadCloser, error)
    
    // Version management methods (from client.ArtHelper)
    ListReleasesOrder() (map[int]map[int]hscvrs.Collection, error)
    ListReleasesMajor(major int) (hscvrs.Collection, error)
    ListReleasesMinor(major, minor int) (hscvrs.Collection, error)
    GetLatest() (*hscvrs.Version, error)
    GetLatestMajor(major int) (*hscvrs.Version, error)
    GetLatestMinor(major, minor int) (*hscvrs.Version, error)
}
```

### Helper Functions

```go
// CheckRegex validates artifact name against regex pattern
func CheckRegex(name, regex string) bool

// ValidatePreRelease filters out non-production versions
// Returns false for: alpha, beta, rc, dev, test, draft, master, main
func ValidatePreRelease(version *hscvrs.Version) bool
```

---

## Performance

### API Rate Limits

- **GitHub**: 60 req/hour (unauthenticated), 5000 req/hour (authenticated)
- **GitLab**: Varies by plan (typically 10 req/second)
- **JFrog**: Depends on server configuration
- **S3**: No specific rate limits, but cost per request applies

### Pagination

Platform implementations automatically handle pagination:
- GitHub: 100 releases per page
- GitLab: 100 releases per page
- JFrog: Single request retrieves file list
- S3: Configurable page size

### Caching Strategy

The package does NOT implement caching. For production use:

```go
// Example caching wrapper
type CachedClient struct {
    client artifact.Client
    cache  map[string]interface{}
    ttl    time.Duration
}

func (c *CachedClient) ListReleases() (hscvrs.Collection, error) {
    if cached, ok := c.cache["releases"]; ok {
        return cached.(hscvrs.Collection), nil
    }
    
    releases, err := c.client.ListReleases()
    if err == nil {
        c.cache["releases"] = releases
        go c.expireAfter("releases", c.ttl)
    }
    return releases, err
}
```

### Memory Usage

Version collections use `hashicorp/go-version` which is memory-efficient:
- ~100 bytes per version entry
- 1000 versions ≈ 100KB memory

---

## Best Practices

**Use Context for Timeouts**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

versions, err := client.ListReleases()
```

**Cache Version Lists**
```go
// Avoid repeated API calls - cache results
type CachedClient struct {
    client artifact.Client
    cache  map[string]interface{}
    ttl    time.Duration
}
```

**Handle Pre-releases Appropriately**
```go
// Pre-releases are automatically filtered by ListReleases()
// To include them, use platform-specific methods if available
```

**Check for Multiple Platforms**
```go
platforms := []string{"linux-amd64", "darwin-arm64", "windows-amd64.exe"}
for _, platform := range platforms {
    url, err := client.GetArtifact(platform, "", version)
    if err != nil {
        continue // Handle missing platform gracefully
    }
}
```

---

## Testing

**Test Suite**: 45 specs using Ginkgo v2 and Gomega

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# Using Ginkgo
ginkgo -r -cover
```

**Coverage**

| Package | Coverage | Notes |
|---------|----------|-------|
| `artifact` | 100% | Helper functions |
| `artifact/client` | 98.6% | Version organization |
| `artifact/github` | 8.6% | No external API calls |
| `artifact/gitlab` | 14.4% | No external API calls |
| `artifact/jfrog` | 6.8% | No external API calls |
| `artifact/s3aws` | 2.0% | No external API calls |

**Note**: Lower coverage in platform packages is intentional to avoid billing and external dependencies in CI/CD.

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `ginkgo -r`
- Maintain or improve test coverage
- Follow existing code style and patterns

**Pull Request Process**
1. Fork the repository
2. Create a feature branch
3. Write tests for your changes
4. Ensure all tests pass
5. Update documentation
6. Submit PR with clear description

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Caching & Performance**
- Built-in TTL-based caching layer
- Parallel downloads for multiple artifacts
- GraphQL support for GitHub (better performance)

**Reliability**
- Checksum verification
- Retry logic with exponential backoff
- Circuit breaker pattern

**Observability**
- OpenTelemetry integration
- Metrics and tracing
- Structured logging

**Platform Support**
- Container registries (Docker Hub, GHCR, GCR)
- Maven/NPM repositories
- Cloud storage providers (Azure Blob, GCS)

Suggestions welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/artifact)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Semantic Versioning**: [semver.org](https://semver.org/)
- **hashicorp/go-version**: [Documentation](https://github.com/hashicorp/go-version)
- **GitHub API**: [REST API Docs](https://docs.github.com/en/rest)
- **GitLab API**: [API Docs](https://docs.gitlab.com/ee/api/)
- **JFrog API**: [REST APIs](https://jfrog.com/help/r/jfrog-rest-apis)
- **AWS S3 API**: [Reference](https://docs.aws.amazon.com/s3/)