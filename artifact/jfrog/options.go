package jfrog

import (
	"net/url"
	"strings"
)

type Options interface {
	GetPattern() string
	SetPattern(pattern string)

	GetExcludePattern() string
	SetExcludePattern(pattern string)

	GetRecursive() bool
	SetRecursive(enabled bool)

	LenProps() int
	EncProps() string
	GetProps() url.Values
	SetProps(props url.Values)
	GetPropKey(key string) string
	SetPropKey(key, value string)
	DelPropKey(key string)

	LenExcludeProps() int
	EncExcludeProps() string
	GetExcludeProps() url.Values
	SetExcludeProps(props url.Values)
	GetExcludePropKey(key string) string
	SetExcludePropKey(key, value string)
	DelExcludePropKey(key string)
}

type artifactoryOptions struct {
	r string
	e string
	p url.Values
	x url.Values
	a bool
}

func (a artifactoryOptions) GetPattern() string {
	return a.r
}

func (a *artifactoryOptions) SetPattern(pattern string) {
	a.r = pattern
}

func (a artifactoryOptions) GetExcludePattern() string {
	return a.e
}

func (a *artifactoryOptions) SetExcludePattern(pattern string) {
	a.e = pattern
}

func (a artifactoryOptions) GetRecursive() bool {
	return a.a
}

func (a *artifactoryOptions) SetRecursive(enabled bool) {
	a.a = enabled
}

func (a artifactoryOptions) LenProps() int {
	return len(a.p)
}

func (a artifactoryOptions) EncProps() string {
	return strings.ReplaceAll(a.p.Encode(), "&", ";")
}

func (a artifactoryOptions) GetProps() url.Values {
	return a.p
}

func (a *artifactoryOptions) SetProps(props url.Values) {
	a.p = props
}

func (a *artifactoryOptions) GetPropKey(key string) string {
	return a.p.Get(key)
}

func (a *artifactoryOptions) SetPropKey(key, value string) {
	a.p.Set(key, value)
}

func (a *artifactoryOptions) DelPropKey(key string) {
	a.p.Del(key)
}

func (a artifactoryOptions) LenExcludeProps() int {
	return len(a.x)
}

func (a artifactoryOptions) EncExcludeProps() string {
	return strings.ReplaceAll(a.x.Encode(), "&", ";")
}

func (a artifactoryOptions) GetExcludeProps() url.Values {
	return a.x
}

func (a *artifactoryOptions) SetExcludeProps(props url.Values) {
	a.x = props
}

func (a *artifactoryOptions) GetExcludePropKey(key string) string {
	return a.x.Get(key)
}

func (a *artifactoryOptions) SetExcludePropKey(key, value string) {
	a.x.Set(key, value)
}

func (a *artifactoryOptions) DelExcludePropKey(key string) {
	a.x.Del(key)
}
