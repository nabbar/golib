/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package router

import (
	"net/http"

	ginsdk "github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
)

type HeadersConfig map[string]string

func (h HeadersConfig) New() Headers {
	var res = NewHeaders()

	for k, v := range h {
		res.Add(k, v)
	}

	return res
}

func (h HeadersConfig) Validate() liberr.Error {
	err := ErrorConfigValidator.Error(nil)

	if !err.HasParent() {
		err = nil
	}

	return err
}

type headers struct {
	head http.Header
}

type Headers interface {
	Add(key, value string)
	Set(key, value string)
	Get(key string) string
	Del(key string)

	Header() map[string]string
	Register(router ...ginsdk.HandlerFunc) []ginsdk.HandlerFunc
	Handler(c *ginsdk.Context)

	Clone() Headers
}

func NewHeaders() Headers {
	return &headers{
		head: make(http.Header),
	}
}

func (h headers) Clone() Headers {
	return &headers{
		head: h.head,
	}
}

func (h headers) Register(router ...ginsdk.HandlerFunc) []ginsdk.HandlerFunc {
	res := make([]ginsdk.HandlerFunc, 0)
	res = append(res, h.Handler)
	res = append(res, router...)

	return res
}

func (h headers) Header() map[string]string {
	res := make(map[string]string)

	if h.head == nil {
		return res
	}

	for k := range h.head {
		res[k] = h.head.Get(k)
	}

	return res
}

func (h headers) Handler(c *ginsdk.Context) {
	if h.head == nil {
		return
	}

	for k := range h.head {
		c.Header(k, h.head.Get(k))
	}
}

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
func (h *headers) Add(key, value string) {
	if h.head == nil {
		h.head = make(http.Header)
	}

	h.head.Add(key, value)
}

// Set sets the header entries associated with key to
// the single element value. It replaces any existing
// values associated with key.
func (h *headers) Set(key, value string) {
	if h.head == nil {
		h.head = make(http.Header)
	}

	h.head.Set(key, value)
}

// Get gets the first value associated with the given key.
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is used
// to canonicalize the provided key.
// If there are no values associated with the key, Get returns "".
// To access multiple values of a key, or to use non-canonical keys,
// access the map directly.
func (h headers) Get(key string) string {
	if h.head == nil {
		h.head = make(http.Header)
	}

	return h.head.Get(key)
}

// Del deletes the values associated with key.
func (h *headers) Del(key string) {
	if h.head == nil {
		h.head = make(http.Header)
	}

	h.head.Del(key)
}
