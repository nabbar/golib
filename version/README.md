# Version package
Help manage package, version, build hash, tag ....

## Example of implement
Create a package release :
```bash
/
...
  release/
    release.go
    version.go
```

In `release.go` file, we will create the public variable to be overwrite in build : 
```go
package release

type EmptyStruct struct{}

var (
	// Release the git tag of the current build, used with -X release.Release=$(git describe --tags HEAD || git describe --all HEAD)
	Release = "0.0"

	// Build the git commit of the current build, used with -X release.Build=$(git rev-parse --short HEAD)
	Build = "00000"

	// Date the current datetime RFC like for the build, used with -X release.Date=$(date +%FT%T%z)
	Date = "2017-10-21T00:00:00+0200"

	// Package the current package name of the build directory, used with -X release.Package=$(basename $(pwd))
	Package = ""

	// Package the current package name of the build directory, used with -X release.Description=...
	Description = "example of dexscription ..."

	// Author the name of the author for the current package, used with -X release.Author=...
	Author = "placeholder"

	// Prefix the package prefix could be used example for env var, used with -X config.Prefix=...
	Prefix = "EXPL"
)
``` 

In `version.go` file, we will implement the call to package `golib/version` :
```go
package release

import "github.com/nabbar/golib/version"

var (
	vers version.Version
)

func init() {
	vers = version.NewVersion(version.License_MIT, release.Package, release.Description, release.Date, release.Build, release.Release, release.Author, release.Prefix, release.EmptyStruct{}, 1)
}

func GetVersion() version.Version {
	return vers
}
```

