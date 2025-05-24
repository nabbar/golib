# golib : custom lib for go

[![Known Vulnerabilities](https://snyk.io/test/github/nabbar/golib/badge.svg)](https://snyk.io/test/github/nabbar/golib)
[![Go](https://github.com/nabbar/golib/workflows/Go/badge.svg)](https://github.com/nabbar/golib)
[![GoDoc](https://pkg.go.dev/badge/github.com/nabbar/golib)](https://pkg.go.dev/github.com/nabbar/golib)
[![Go Report Card](https://goreportcard.com/badge/github.com/nabbar/golib)](https://goreportcard.com/report/github.com/nabbar/golib)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)


## using in source code 
first get the source dependancies
```shell script
go get github.com/nabbar/golib/...
```

second, import the needed lib in your code
```go
import "github.com/nabbar/golib/version"
```

## Details of packages :
* [package archive](archive/README.md)
* [package artifact](artifact/README.md)
* [package atomic](atomic/README.md)
* [package aws](aws/README.md)
* [package certificates](certificates/README.md)
* [pacakge cobra](cobra/README.md)
* [package config](config/README.md)
* [package console](console/README.md)
* [package context](context/README.md)
* [package database](database/README.md)
* [pacakge duration](duration/README.md)
* [package encoding](encoding/README.md)
* [package errors](errors/README.md)
* [package file](file/README.md)
* [package ftpclient](ftpclient/README.md)
* [package httpcli](httpcli/README.md)
* [package httpserver](httpserver/README.md)
* [package ioutil](ioutil/README.md)
* [package ldap](ldap/README.md)
* [package logger](logger/README.md)
* [package mail](mail/README.md)
* [package mailer](mailer/README.md)
* [package mailPooler](mailPooler/README.md)
* [package monitor](monitor/README.md)
* [package network](network/README.md)
* [package password](password/README.md)
* [package router](router/README.md)
* [package static](static/README.md)
* [package status](status/README.md)
* [package version](version/README.md)

# Build tags 
To build static, pure go, some packages need to use tags osusergo and netgo, like this
```bash
go build -a -tags "osusergo netgo" -installsuffix cgo ...
```
