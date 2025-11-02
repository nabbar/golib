/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package artifact

import (
	"io"
	"os"
	"regexp"
	"strings"

	hscvrs "github.com/hashicorp/go-version"
	artcli "github.com/nabbar/golib/artifact/client"
	liberr "github.com/nabbar/golib/errors"
)

const subUp = 20

// Error code ranges for artifact package and subpackages.
// Each platform implementation has its own error code range to facilitate debugging.
const (
	MinArtifactGitlab = subUp + liberr.MinPkgArtifact // GitLab error code base
	MinArtifactGithub = subUp + MinArtifactGitlab     // GitHub error code base
	MinArtifactJfrog  = subUp + MinArtifactGithub     // JFrog error code base
	MinArtifactS3AWS  = subUp + MinArtifactJfrog      // AWS S3 error code base
)

// Client defines the unified interface for artifact management across different platforms.
// It embeds ArtHelper for version organization and adds platform-specific operations.
//
// Implementations:
//   - github.Github: GitHub Releases integration
//   - gitlab.Gitlab: GitLab Releases integration
//   - jfrog.Artifactory: JFrog Artifactory integration
//   - s3aws.S3: AWS S3 integration
//
// The interface provides:
//   - Version discovery and filtering (via ArtHelper)
//   - Artifact URL retrieval (GetArtifact)
//   - Direct streaming downloads (Download)
type Client interface {
	artcli.ArtHelper

	// ListReleases retrieves all stable versions from the artifact repository.
	// Pre-release versions (alpha, beta, rc, etc.) are automatically filtered out.
	// Returns a sorted collection of semantic versions.
	ListReleases() (releases hscvrs.Collection, err error)

	// GetArtifact retrieves the download URL for a specific artifact.
	// Matching is performed using either substring (containName) or regex (regexName).
	// If both are provided, regex takes precedence.
	//
	// Parameters:
	//   - containName: substring to match in artifact name (e.g., "linux-amd64")
	//   - regexName: regex pattern to match artifact name (e.g., `.*-linux-amd64\.tar\.gz$`)
	//   - release: target version to download
	//
	// Returns the download URL or an error if the artifact is not found.
	GetArtifact(containName string, regexName string, release *hscvrs.Version) (link string, err error)

	// Download streams the artifact content directly without intermediate storage.
	// Returns the content size, a ReadCloser for the content, and any error.
	// The caller is responsible for closing the ReadCloser.
	//
	// Parameters:
	//   - containName: substring to match in artifact name
	//   - regexName: regex pattern to match artifact name (takes precedence)
	//   - release: target version to download
	Download(containName string, regexName string, release *hscvrs.Version) (int64, io.ReadCloser, error)
}

// CheckRegex validates an artifact name against a regex pattern.
// This function is commonly used to match artifact filenames against expected patterns,
// such as platform-specific binary names or versioned release files.
//
// Example:
//
//	CheckRegex("myapp-1.2.3-linux-amd64.tar.gz", `myapp-\d+\.\d+\.\d+-linux-amd64\.tar\.gz`)
//	// Returns: true
func CheckRegex(name, regex string) bool {
	if ok, _ := regexp.MatchString(regex, name); ok {
		return ok
	}

	return false
}

// DownloadRelease downloads an artifact from a URL and saves it to a file.
// This function is not yet implemented and will panic if called.
//
// Deprecated: Use Client.Download() method instead for streaming downloads.
func DownloadRelease(link string) (file os.File, err error) {
	panic("not implemented")
}

// ValidatePreRelease filters out non-production release versions.
// Returns true only for GA (General Availability) versions or versions with
// custom prerelease tags that don't match development/testing patterns.
//
// Rejected patterns (returns false):
//   - alpha, beta, rc (release candidate)
//   - dev, test, draft
//   - master, main (branch tags)
//   - Single letter abbreviations: "a" or "b"
//
// Accepted patterns (returns true):
//   - GA versions (no prerelease): "1.2.3"
//   - Custom tags: "1.2.3-stable", "1.2.3-hotfix", "1.2.3-final"
//
// Example:
//
//	ValidatePreRelease(version.NewVersion("1.2.3"))        // true (GA)
//	ValidatePreRelease(version.NewVersion("1.2.3-beta"))   // false
//	ValidatePreRelease(version.NewVersion("1.2.3-stable")) // true
func ValidatePreRelease(version *hscvrs.Version) bool {
	var (
		p = strings.ToLower(version.Prerelease())
		s = []string{
			"alpha",
			"beta",
			"rc",
			"dev",
			"test",
			"draft",
			"master",
			"main",
		}
	)

	// Empty prerelease (GA version) is valid
	if p == "" {
		return true
	}

	// Check for single letter abbreviations (a for alpha, b for beta)
	if p == "a" || strings.HasPrefix(p, "a.") {
		return false
	}
	if p == "b" || strings.HasPrefix(p, "b.") {
		return false
	}

	// Check for blacklisted words
	for _, i := range s {
		if strings.Contains(p, i) {
			return false
		}
	}

	return true
}
