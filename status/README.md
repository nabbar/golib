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

    "myapp/release"
    "myapp/bin/api/config"
    "myapp/bin/api/routers"
)

const (
	msgOk = "API is working"
	msgKO = "something is not well working"
)

func init() {
	sts := status.NewVersionStatus(release.GetVersion(), msgOk, msgKO, GetHealth, GetHeader)
	sts.Register("/status", routers.RouterList.Register)
	sts.AddComponent("AWS S3 Helper", msgOk, msgKO, "", "", true, GetAWSHealth)
	sts.AddComponent("OpenLDAP", msgOk, msgKO, "", "", true, GetLDAPHealth)
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

```

In some case, using init function could make mistake (specially if you need to read flag or config file).
In this case, you will have func "Later" to allow the init package on first call of `status` router.
 