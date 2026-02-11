/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package version

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	govers "github.com/hashicorp/go-version"
	liberr "github.com/nabbar/golib/errors"
)

type versionModel struct {
	versionRelease     string
	versionBuild       string
	versionTime        time.Time
	versionPackage     string
	versionDescription string
	versionAuthor      string
	versionPrefix      string
	versionSource      string
	licenceType        license
}

// Version is the main interface for accessing version and license information.
// All methods are safe for concurrent use.
type Version interface {
	// CheckGo validates that the current Go runtime version satisfies the given constraint.
	// RequireGoVersion is the version to check against (e.g., "1.18", "1.21.0").
	// RequireGoContraint is the operator (e.g., ">=", "==", "~>").
	// Returns an error if the constraint is not satisfied or if parsing fails.
	CheckGo(RequireGoVersion, RequireGoContraint string) liberr.Error

	// GetAppId returns a unique application identifier including version, OS, and architecture.
	GetAppId() string
	// GetAuthor returns the author information with source package path.
	GetAuthor() string
	// GetBuild returns the build identifier (typically a git commit hash).
	GetBuild() string
	// GetDate returns the build date as a formatted string.
	GetDate() string
	// GetTime returns the build date as a time.Time object.
	GetTime() time.Time
	// GetDescription returns the application description.
	GetDescription() string
	// GetHeader returns a formatted header string with package, release, and build info.
	GetHeader() string
	// GetInfo returns detailed version information including release, build, and date.
	GetInfo() string
	// GetPackage returns the package name.
	GetPackage() string
	// GetRootPackagePath returns the root package path extracted via reflection.
	GetRootPackagePath() string
	// GetPrefix returns the uppercase prefix for environment variables.
	GetPrefix() string
	// GetRelease returns the release version string.
	GetRelease() string

	// GetLicenseName returns the name of the primary license.
	GetLicenseName() string
	// GetLicenseLegal returns the full legal text of the license(s).
	// Additional licenses can be provided to combine multiple license texts.
	GetLicenseLegal(addMoreLicence ...license) string
	// GetLicenseFull returns the complete license including boilerplate and legal text.
	// Additional licenses can be provided to combine multiple licenses.
	GetLicenseFull(addMoreLicence ...license) string
	// GetLicenseBoiler returns the license boilerplate suitable for file headers.
	// Additional licenses can be provided to combine multiple boilerplates.
	GetLicenseBoiler(addMoreLicence ...license) string

	// PrintInfo prints version information to stderr.
	PrintInfo()
	// PrintLicense prints license boilerplate to stderr.
	// Additional licenses can be provided to print multiple licenses.
	PrintLicense(addlicence ...license)
}

// NewVersion creates a new Version instance with the provided information.
//
// Parameters:
//   - License: The primary license type (e.g., License_MIT, License_Apache_v2)
//   - Package: The package/application name (use "" or "noname" to auto-detect from reflection)
//   - Description: A description of the application
//   - Date: Build date in RFC3339 format (e.g., "2024-01-15T10:30:00Z")
//   - Build: Build identifier, typically a git commit hash
//   - Release: Release version string (e.g., "v1.2.3")
//   - Author: Author name or organization
//   - Prefix: Prefix for environment variables (will be converted to uppercase)
//   - emptyInterface: An empty struct instance used for reflection to extract package path
//   - numSubPackage: Number of parent directories to traverse from the reflected package path
//     (0 = current package, 1 = parent, 2 = grandparent, etc.)
//
// Returns a Version interface for accessing version and license information.
//
// Example:
//
//	type MyApp struct{}
//	v := NewVersion(
//		License_MIT,
//		"MyApp",
//		"My Application",
//		"2024-01-15T10:30:00Z",
//		"abc123",
//		"v1.0.0",
//		"John Doe",
//		"MYAPP",
//		MyApp{},
//		0,
//	)
func NewVersion(License license, Package, Description, Date, Build, Release, Author, Prefix string, emptyInterface interface{}, numSubPackage int) Version {
	rfl := reflect.TypeOf(emptyInterface)
	//println("reflect typeOf name : " + rfl.Name())
	//println("reflect typeOf package path : " + rfl.PkgPath())
	Source := rfl.PkgPath()

	for i := 1; i <= numSubPackage; i++ {
		Source = filepath.Dir(Source)
	}

	if Package == "" || Package == "noname" {
		Package = filepath.Base(Source)
	}

	var timeBuild time.Time
	if ts, err := time.Parse(time.RFC3339, Date); err != nil {
		timeBuild = time.Now()
	} else {
		timeBuild = ts
	}

	return &versionModel{
		versionRelease:     Release,
		versionBuild:       Build,
		versionTime:        timeBuild,
		versionPackage:     Package,
		versionDescription: Description,
		versionAuthor:      Author,
		versionPrefix:      Prefix,
		versionSource:      Source,
		licenceType:        License,
	}
}

func (v versionModel) CheckGo(RequireGoVersion, RequireGoContraint string) liberr.Error {
	constraint, err := govers.NewConstraint(RequireGoContraint + RequireGoVersion)
	if err != nil {
		return ErrorGoVersionInit.Error(err)
	}

	goVersion, err := govers.NewVersion(runtime.Version()[2:])
	if err != nil {
		return ErrorGoVersionRuntime.Error(err)
	}

	if !constraint.Check(goVersion) {
		//nolint #goerr113
		return ErrorGoVersionConstraint.Error(fmt.Errorf("must be compiled with Go version %s (instead of %s)", RequireGoVersion, goVersion))
	}

	return nil
}

func (v versionModel) getYearOfDate() string {
	return fmt.Sprintf("%d", v.versionTime.Year())
}

// Info print all information about current build and version.
func (v versionModel) PrintInfo() {
	println(fmt.Sprintf("Running %s", v.GetHeader()))
}

// GetInfo return string about current build and version.
func (v versionModel) GetInfo() string {
	return fmt.Sprintf("Release: %s, Build: %s, Date: %s", v.versionRelease, v.versionBuild, v.GetDate())
}

// GetAppId return string about package name, release and runtime info.
func (v versionModel) GetAppId() string {
	return fmt.Sprintf("%s (OS: %s; Arch: %s; Runtime: %s)", v.versionRelease, runtime.GOOS, runtime.GOARCH, runtime.Version()[2:])
}

// GetAuthor return string about author name and repository info.
func (v versionModel) GetAuthor() string {
	return fmt.Sprintf("by %s (source : %s)", v.versionAuthor, v.versionSource)
}

func (v versionModel) GetDescription() string {
	return v.versionDescription
}

// GetAuthor return string about author name and repository info.
func (v versionModel) GetHeader() string {
	return fmt.Sprintf("%s (%s)", v.versionPackage, v.GetInfo())
}

func (v versionModel) GetDate() string {
	return v.versionTime.Format(time.RFC1123)
}

func (v versionModel) GetTime() time.Time {
	return v.versionTime
}

func (v versionModel) GetBuild() string {
	return v.versionBuild
}

func (v versionModel) GetPackage() string {
	return v.versionPackage
}

func (v versionModel) GetRootPackagePath() string {
	return v.versionSource
}

func (v versionModel) GetPrefix() string {
	return strings.ToUpper(v.versionPrefix)
}

func (v versionModel) GetRelease() string {
	return v.versionRelease
}

func (v versionModel) GetLicenseName() string {
	return v.licenceType.GetLicenseName()
}

func (v versionModel) GetLicenseLegal(addMoreLicence ...license) string {
	if len(addMoreLicence) == 0 {
		return v.licenceType.GetLicense()
	}

	buff := bytes.NewBufferString(v.licenceType.GetLicense())

	for _, l := range addMoreLicence {
		// #nosec nolint
		/* #nosec */
		_, _ = buff.WriteString("\n\n")
		// #nosec nolint
		/* #nosec */
		_, _ = buff.WriteString(strings.Repeat("*", 80))
		// #nosec nolint
		/* #nosec */
		_, _ = buff.WriteString(strings.Repeat("*", 80))
		// #nosec nolint
		/* #nosec */
		_, _ = buff.WriteString("\n\n")
		// #nosec nolint
		/* #nosec */
		_, _ = buff.WriteString(l.GetLicense())
	}

	return buff.String()
}

func (v versionModel) GetLicenseFull(addMoreLicence ...license) string {
	buff := bytes.NewBufferString(v.GetLicenseBoiler(addMoreLicence...))

	// #nosec nolint
	/* #nosec */
	_, _ = buff.WriteString("\n\n")
	// #nosec nolint
	/* #nosec */
	_, _ = buff.WriteString(strings.Repeat("*", 80))
	// #nosec nolint
	/* #nosec */
	_, _ = buff.WriteString(strings.Repeat("*", 80))
	// #nosec nolint
	/* #nosec */
	_, _ = buff.WriteString("\n\n")
	// #nosec nolint
	/* #nosec */
	_, _ = buff.WriteString(v.GetLicenseLegal(addMoreLicence...))

	return buff.String()
}

func (v versionModel) GetLicenseBoiler(addMoreLicence ...license) string {
	if len(addMoreLicence) == 0 {
		return v.licenceType.GetBoilerPlate(v.versionPackage, v.versionDescription, v.getYearOfDate(), v.versionAuthor)
	}

	year := v.getYearOfDate()
	buff := bytes.NewBufferString(v.licenceType.GetBoilerPlate(v.versionPackage, v.versionDescription, year, v.versionAuthor))

	for _, l := range addMoreLicence {
		// #nosec nolint
		/* #nosec */
		_, _ = buff.WriteString("\n\n")
		// #nosec nolint
		/* #nosec */
		_, _ = buff.WriteString(l.GetBoilerPlate(v.versionPackage, v.versionDescription, year, v.versionAuthor))
	}

	return buff.String()
}

func (v versionModel) PrintLicense(addMoreLicence ...license) {
	println(v.GetLicenseBoiler(addMoreLicence...))
}
