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

package njs_version

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"time"

	"strings"

	govers "github.com/hashicorp/go-version"
	. "github.com/nabbar/golib/njs-errors"
)

type versionModel struct {
	versionRelease     string
	versionBuild       string
	versionDate        string
	versionPackage     string
	versionDescription string
	versionAuthor      string
	versionPrefix      string
	versionSource      string
	licenceType        license
}

type Version interface {
	CheckGo(RequireGoVersion, RequireGoContraint string) Error

	GetAppId() string
	GetAuthor() string
	GetBuild() string
	GetDescription() string
	GetHeader() string
	GetInfo() string
	GetPackage() string
	GetPrefix() string
	GetRelease() string

	GetLicenseLegal(addMoreLicence ...license) string
	GetLicenseFull(addMoreLicence ...license) string
	GetLicenseBoiler(addMoreLicence ...license) string

	PrintInfo()
	PrintLicense(addlicence ...license)
}

func NewVersion(License license, Package, Description, Date, Build, Release, Author, Prefix string, emptyInterface interface{}, numSubPackage int) Version {
	rfl := reflect.TypeOf(emptyInterface)
	//println("reflect typeOf name : " + rfl.Name())
	//println("reflect typeOf package path : " + rfl.PkgPath())
	Source := rfl.PkgPath()

	for i := 1; i <= numSubPackage; i++ {
		Source = path.Dir(Source)
	}

	if Package == "" || Package == "noname" {
		Package = path.Base(Source)
	}

	return &versionModel{
		versionRelease:     Release,
		versionBuild:       Build,
		versionDate:        Date,
		versionPackage:     Package,
		versionDescription: Description,
		versionAuthor:      Author,
		versionPrefix:      Prefix,
		versionSource:      Source,
		licenceType:        License,
	}
}

func (vers versionModel) CheckGo(RequireGoVersion, RequireGoContraint string) Error {
	constraint, err := govers.NewConstraint(RequireGoContraint + RequireGoVersion)
	if err != nil {
		return GOVERSION_INIT.ErrorParent(err)
	}

	goVersion, err := govers.NewVersion(runtime.Version()[2:])
	if err != nil {
		return GOVERSION_RUNTIME.ErrorParent(err)
	}

	if !constraint.Check(goVersion) {
		return GOVERSION_CONTRAINT.ErrorParent(fmt.Errorf("must be compiled with Go version %s (instead of %s)", RequireGoVersion, goVersion))
	}

	return nil
}

func (vers versionModel) getYearOfDate() string {
	dt, err := time.Parse(time.RFC3339, vers.versionDate)

	if err != nil {
		dt = time.Now()
	}

	return fmt.Sprintf("%d", dt.Year())
}

// Info print all information about current build and version
func (vers versionModel) PrintInfo() {
	println(fmt.Sprintf("Running %s", vers.GetHeader()))
}

// GetInfo return string about current build and version
func (vers versionModel) GetInfo() string {
	return fmt.Sprintf("Release: %s, Build: %s, Date: %s", vers.versionRelease, vers.versionBuild, vers.versionDate)
}

// GetAppId return string about package name, release and runtime info
func (vers versionModel) GetAppId() string {
	return fmt.Sprintf("%s (OS: %s; Arch: %s; Runtime: %s)", vers.versionRelease, runtime.GOOS, runtime.GOARCH, runtime.Version()[2:])
}

// GetAuthor return string about author name and repository info
func (vers versionModel) GetAuthor() string {
	return fmt.Sprintf("by %s (source : %s)", vers.versionAuthor, vers.versionSource)
}

func (vers versionModel) GetDescription() string {
	return vers.versionDescription
}

// GetAuthor return string about author name and repository info
func (vers versionModel) GetHeader() string {
	return fmt.Sprintf("%s (%s)", vers.versionPackage, vers.GetInfo())
}

func (vers versionModel) GetBuild() string {
	return vers.versionBuild
}

func (vers versionModel) GetPackage() string {
	return vers.versionPackage
}

func (vers versionModel) GetPrefix() string {
	return strings.ToUpper(vers.versionPrefix)
}

func (vers versionModel) GetRelease() string {
	return vers.versionRelease
}

func (vers versionModel) GetLicenseLegal(addMoreLicence ...license) string {
	if len(addMoreLicence) == 0 {
		return vers.licenceType.GetLicense()
	}

	buff := bytes.NewBufferString(vers.licenceType.GetLicense())

	for _, l := range addMoreLicence {
		_, _ = buff.WriteString("\n\n")                  // #nosec
		_, _ = buff.WriteString(strings.Repeat("*", 80)) // #nosec
		_, _ = buff.WriteString(strings.Repeat("*", 80)) // #nosec
		_, _ = buff.WriteString("\n\n")                  // #nosec
		_, _ = buff.WriteString(l.GetLicense())          // #nosec
	}

	return buff.String()
}

func (vers versionModel) GetLicenseFull(addMoreLicence ...license) string {
	buff := bytes.NewBufferString(vers.GetLicenseBoiler(addMoreLicence...))

	_, _ = buff.WriteString("\n\n")                                  // #nosec
	_, _ = buff.WriteString(strings.Repeat("*", 80))                 // #nosec
	_, _ = buff.WriteString(strings.Repeat("*", 80))                 // #nosec
	_, _ = buff.WriteString("\n\n")                                  // #nosec
	_, _ = buff.WriteString(vers.GetLicenseLegal(addMoreLicence...)) // #nosec

	return buff.String()
}

func (vers versionModel) GetLicenseBoiler(addMoreLicence ...license) string {
	if len(addMoreLicence) == 0 {
		return vers.licenceType.GetBoilerPlate(vers.versionPackage, vers.versionDescription, vers.getYearOfDate(), vers.versionAuthor)
	}

	year := vers.getYearOfDate()
	buff := bytes.NewBufferString(vers.licenceType.GetBoilerPlate(vers.versionPackage, vers.versionDescription, year, vers.versionAuthor))

	for _, l := range addMoreLicence {
		_, _ = buff.WriteString("\n\n")                                                                                   // #nosec
		_, _ = buff.WriteString(l.GetBoilerPlate(vers.versionPackage, vers.versionDescription, year, vers.versionAuthor)) // #nosec
	}

	return buff.String()
}

func (vers versionModel) PrintLicense(addMoreLicence ...license) {
	println(vers.GetLicenseBoiler(addMoreLicence...))
}
