# Package Router
This package help to manage routers in an API.
This package requires go Gin Tonic API Framework.

By default, Gin Tonic API need a main package to register all handler func into the gin engine. 
This way is not easy usable with many people who's can add routers or exploit them.
In add, middleware isn't defined as real middleware but more than cascaded call of all registered handler for a route.

This package allows to change this process : 
 - auth allow to register a real middleware who's can manage call of other handler for a route
 - register implement a global collector of handler with router / group and called into init func of yours routers directly
 - router implements the global registrar of all registered handler with route and group, ordered by group
 - ...

## Example of implementation
We will work on an example of file/folder tree like this : 
```bash
/
  bin/
    api/
      config/
      routers/
        routers.go
```

in the `routers.go` file, we will implement the router package call :
```go
package routers

import (
	"github.com/nabbar/golib/router"

    "myapp/bin/api/config"
)

var (
	RouterList = router.NewRouterList()
)

func Run() {
	config.GetConfig().ServerListen(router.Handler(RouterList))
}

```

This variable RouterList will be call by all routers.
Note you will just need to call your routers' packages into a main router like this :
```go
package main

import (
    _ "myapp/bin/api/routers/status"
    _ "myapp/bin/api/routers/static"
  // ... add all your packages with an init register
  // careful: do not add this import into your routers.go package to avoid circular import
)
```