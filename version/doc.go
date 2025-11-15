/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

/*
Package version provides version management and license handling for Go applications.

# Overview

This package offers a comprehensive solution for managing application version information,
including build metadata, release versions, license information, and Go version constraints.
It supports multiple open-source licenses and provides formatted output for version and
license information.

# Features

  - Version information management (release, build, date, author)
  - Multiple license support (MIT, Apache, GPL, LGPL, AGPL, Mozilla, Creative Commons, etc.)
  - Go version constraint validation
  - Automatic package path extraction via reflection
  - Thread-safe operations
  - Formatted output for headers, info, and license text

# Basic Usage

Creating a version instance:

	import "github.com/nabbar/golib/version"

	type MyStruct struct{}

	v := version.NewVersion(
		version.License_MIT,           // License type
		"MyApp",                        // Package name
		"My Application Description",  // Description
		"2024-01-15T10:30:00Z",        // Build date (RFC3339)
		"abc123def",                    // Build hash
		"v1.2.3",                       // Release version
		"John Doe",                     // Author
		"MYAPP",                        // Prefix for environment variables
		MyStruct{},                     // Empty struct for reflection
		0,                              // Number of parent packages to traverse
	)

# Retrieving Version Information

	// Get formatted header
	fmt.Println(v.GetHeader())

	// Get detailed info
	fmt.Println(v.GetInfo())

	// Get individual fields
	release := v.GetRelease()
	build := v.GetBuild()
	date := v.GetDate()
	author := v.GetAuthor()

# License Management

The package supports multiple license types:

  - MIT License
  - Apache License v2.0
  - GNU GPL v3
  - GNU Affero GPL v3
  - GNU Lesser GPL v3
  - Mozilla Public License v2.0
  - Unlicense
  - Creative Commons Zero v1.0
  - Creative Commons Attribution v4.0
  - Creative Commons Attribution-ShareAlike v4.0
  - SIL Open Font License v1.1

Retrieving license information:

	// Get license name
	name := v.GetLicenseName()

	// Get full license text
	legal := v.GetLicenseLegal()

	// Get license boilerplate (for file headers)
	boiler := v.GetLicenseBoiler()

	// Get complete license (boilerplate + legal text)
	full := v.GetLicenseFull()

	// Support for multiple licenses
	combined := v.GetLicenseLegal(
		version.License_Apache_v2,
		version.License_MIT,
	)

# Go Version Constraints

Validate that the application is running with a compatible Go version:

	// Require Go >= 1.18
	if err := v.CheckGo("1.18", ">="); err != nil {
		log.Fatal(err)
	}

	// Require exact Go version
	if err := v.CheckGo("1.21.0", "=="); err != nil {
		log.Fatal(err)
	}

	// Pessimistic constraint (compatible with minor versions)
	if err := v.CheckGo("1.20", "~>"); err != nil {
		log.Fatal(err)
	}

Supported constraint operators:

  - "==" : Exact version match
  - "!=" : Not equal
  - ">"  : Greater than
  - ">=" : Greater than or equal
  - "<"  : Less than
  - "<=" : Less than or equal
  - "~>" : Pessimistic constraint (allows patch-level changes)

# Package Path Extraction

The package uses reflection to automatically extract the package path. The numSubPackage
parameter controls how many parent directories to traverse:

	// numSubPackage = 0: github.com/myorg/myapp/cmd
	// numSubPackage = 1: github.com/myorg/myapp
	// numSubPackage = 2: github.com/myorg

# Error Handling

The package uses the github.com/nabbar/golib/errors package for error handling.
Error codes:

  - ErrorParamEmpty: Required parameter is empty
  - ErrorGoVersionInit: Failed to initialize Go version constraint
  - ErrorGoVersionRuntime: Failed to extract runtime Go version
  - ErrorGoVersionConstraint: Go version constraint not satisfied

# Thread Safety

All methods are safe for concurrent use. The Version interface is read-only after creation,
making it inherently thread-safe.

# Testing

The package includes comprehensive tests with high code coverage (>93%). Run tests with:

	make test              # Run all tests
	make test-race         # Run with race detector
	make test-coverage     # Generate coverage report

# Dependencies

  - github.com/hashicorp/go-version: Version constraint parsing
  - github.com/nabbar/golib/errors: Error handling

# See Also

  - github.com/nabbar/golib/errors: Error handling package
  - github.com/hashicorp/go-version: Version constraint library

# License

This package is released under the MIT License. See the LICENSE file for details.
*/
package version
