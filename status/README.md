# Package Status
This package help to manage status router in a API to respond a standard response for status of API and his component.
This package requires `golib/router` + go Gin Tonic API Framework.

This package also include 2 option of call that can be passed into query string : 
- `short` : if use, the response will only include the main status and no one component, but all health are still check
- `online` : if use, the response will be into a list of text line composed as `status: name (release - build) - message`, instead of a JSON output
This 2 options call be use together. 

## Example of implementation
We will work on an example of file/folder tree like this : 
```bash
/
  bin/
    api/
      config/
      routers/
        status/
          get.go
```

in the `get.go` file, we will implement the status package call :
```go
package status

import (
    "github.com/gin-gonic/gin"
    "github.com/nabbar/golib/status"
    "github.com/nabbar/golib/version"
    "net/http"
)

const (
	msgOk = "API is working"
	msgKO = "API is not working"
	msgWarn = "something is not well working"
)

type EmptyStruct struct{}

var vers version.Version

func init() {
    // optional, get or create a new version interface
    vers = version.NewVersion(version.License_MIT, "Package", "Description", "2017-10-21T00:00:00+0200", "0123456789abcdef", "v0.0-dev", "Author Name", "pfx", EmptyStruct{}, 1)

    // to create new status, you will need some small function to give data, this is type func : 
    // FctHealth        func() error
    // FctInfo          func() (name, release, build string)

    // create new status not as later
    sts := status.NewVersion(vers, msgOk, msgKO, msgWarn)

	// add some middleware before router
	sts.MiddlewareAdd(func(context *gin.Context) {
	    // add here your middleware need to be run before the status route
	})

    // register to the router list
	sts.Register("/status", routers.RouterList.Register)

    // register to the router list with a group
	sts.Register("/v1", "/status", routers.RouterList.Register)

    // add a new component mandatory
	sts.ComponentNew(
	    "myComponentMandatory",
	    NewComponent(true, infoMandatory, healthMandatory,
	        func() (msgOk string, msgKo string) {
                return msgOk, msgKO
	        },
	        24 * time.Hour, 5 * time.second,
        ),
    )

    // add a new component mandatory
	sts.ComponentNew(
	    "myComponentNotMandatory",
	    NewComponent(true, infoNotMandatory, healthNotMandatory,
	        func() (msgOk string, msgKo string) {
                return msgOk, msgKO
	        },
	        24 * time.Hour, 5 * time.second,
        ),
    )

    // use this func to customize return code for each status
    sts.SetErrorCode(http.StatusOK, http.StatusInternalServerError, http.StatusAccepted)
}

func GetHealth() error {
    // any function to return check global API is up or not 
    return nil
}

func GetHeader(c *gin.Context) {
    // any function to return global & generic header (like CSP, HSTS, ...)
}

func GetAWSHealth() error {
    // any function to return the check of the component AWS is up or not 
    return nil
}

func GetLDAPHealth() error {
    // any function to return the check of the component LDAP is up or not 
    return nil
}

func infoMandatory() (name, release, build string) {
	return "Name of my component mandatory", "v0.1.2.3.4", "abcd1234abcd1234"
}

func healthMandatory() error {
	return nil
}

func infoNotMandatory() (name, release, build string) {
	return "Name of my component not mandatory", "v0.1.2.3.4", "abcd1234abcd1234"
}

func healthNotMandatory() error {
	return nil
}

```

In some case, using init function could make mistake (specially if you need to read flag or config file).
In this case, you will have func "Later" to allow the init package on first call of `status` router.
 