# Package Static
This package help to manage static file router in an API to embedded files into the go binary api.
This package requires `packr` tools, `golib/router` & go Gin Tonic API Framework.

## Example of implementation
We will work on an example of file/folder tree like this : 
```bash
/
  bin/
    api/
      config/
      routers/
        static/
          get.go
  static/
    static/
      ...some_static_files...
```

in the `get.go` file, we will implement the static package call :
```go
package static

import (
    "github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
	"github.com/nabbar/golib/static"

    "myapp/release"
    "myapp/bin/api/config"
    "myapp/bin/api/routers"
)

const UrlPrefix = "/static"

func init() {
	staticStcFile := static.NewStatic(false, UrlPrefix, packr.NewBox("../../../../static/static"), GetHeader)

	staticStcFile.SetDownloadAll()
	staticStcFile.Register(routers.RouterList.Register)
}

func GetHeader(c *gin.Context) {
    // any function to return global & generic header (like CSP, HSTS, ...)
}

```
