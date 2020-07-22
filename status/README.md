# Package Status
This package help to manage status router in a API to respond a standard response for status of API and his component.
This package requires `golib/router` + go Gin Tonic API Framework.

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
	msgKO = "something is not well working"
)

type EmptyStruct struct{}

var vers version.Version

func init() {
    // optional, get or create a new version interface
    vers = version.NewVersion(version.License_MIT, "Package", "Description", "2017-10-21T00:00:00+0200", "0123456789abcdef", "v0.0-dev", "Author Name", "pfx", EmptyStruct{}, 1)

    // to create new status, you will need some small function to give data, this is type func : 
    // FctMessagesAll   func() (ok string, ko string, cpt string)
    // FctMessageItem   func() (ok string, ko string)
    // FctHealth        func() error
    // FctInfo          func() (name, release, build string)
    // FctVersion       func() version.Version

    // create new status not as later
	sts := status.NewVersionStatus(getVersion, getMessageAll, GetHealth, GetHeader, false)
    // add a new component  
	sts.AddComponent(infoAws, getMessageItem, GetAWSHealth, true, false)
    // add a new component 
	sts.AddComponent(infoLDAP, getMessageItem, GetLDAPHealth, true, false) 
	
    // register to the router list
	sts.Register("/status", routers.RouterList.Register)
    
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

func infoAws() (name, release, build string) {
	return "AWS S3 Helper", "v0.1.2.3.4", ""
}

func infoLDAP() (name, release, build string) {
	return "OpenLDAP Lib", "v0.1.2.3.4", ""
}

func getVersion() version.Version {
    return vers
}

func getMessageItem() (ok string, ko string) {
  return "all is ok", "there is a mistake somewhere"
}

func getMessageAll() (ok string, ko string, cptErr string) {
  ok, ko = getMessageItem()
  return ok, ko, "at least one component is in failed"
}
```

In some case, using init function could make mistake (specially if you need to read flag or config file).
In this case, you will have func "Later" to allow the init package on first call of `status` router.
 