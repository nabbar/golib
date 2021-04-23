![Go](https://github.com/nabbar/golib/workflows/Go/badge.svg)

# golib : custom lib for go

snyk project : https://app.snyk.io/org/nabbar/project/2f55a2b8-6015-4db1-b859-c2bc3b7548a7


## using in source code 
first get the source dependancies
```shell script
go get -u github.com/nabbar/golib/...
```

second, import the needed lib in your code
```go
import "github.com/nabbar/golib/version"
```

## Details of packages :
* [package errors](errors/README.md)
* [package logger](logger/README.md)
* [package network](network/README.md)
* [package password](password/README.md)
* [package router](router/README.md)
* [package static](static/README.md)
* [package status](status/README.md)
* [package version](version/README.md)

# Build tags 
To build static, pure go, some packages need to use tags osusergo and netgo, like this
```bash
go build -a -tags osusergo,netgo -installsuffix cgo ...
```
