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

package header

import (
	"net/http"

	ginsdk "github.com/gin-gonic/gin"
)

// headers is the internal implementation of the Headers interface.
type headers struct {
	head http.Header // Underlying HTTP header map
}

// Clone creates a shallow copy of the headers instance.
// Note: The underlying http.Header map is shared, so modifications
// to one instance will affect the cloned instance.
func (h headers) Clone() Headers {
	return &headers{
		head: h.head,
	}
}

// Register creates a handler chain with the Header middleware first.
// This ensures that headers are applied before any route handlers execute.
//
// Returns a slice of handlers that can be used with Gin's routing methods.
func (h headers) Register(router ...ginsdk.HandlerFunc) []ginsdk.HandlerFunc {
	res := make([]ginsdk.HandlerFunc, 0)
	res = append(res, h.Handler)
	res = append(res, router...)

	return res
}

// Header returns a map of all configured headers.
// Each header key is mapped to its first value.
// Returns an empty map if no headers are configured.
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

// Handler is the Gin middleware that applies headers to the response.
// It iterates through all configured headers and sets them on the Gin context.
// If no headers are configured, this is a no-op.
func (h headers) Handler(c *ginsdk.Context) {
	if h.head == nil {
		return
	}

	for k := range h.head {
		c.Header(k, h.head.Get(k))
	}
}

// Add adds a value to the header.
// If the header already exists, the value is appended to the existing values.
// Header names are canonicalized using textproto.CanonicalMIMEHeaderKey.
//
// Example:
//
//	headers.Add("Set-Cookie", "session=abc")
//	headers.Add("Set-Cookie", "token=xyz")
func (h *headers) Add(key, value string) {
	if h.head == nil {
		h.head = make(http.Header)
	}

	h.head.Add(key, value)
}

// Set replaces all values for the header with the given value.
// Any existing values are discarded.
// Header names are canonicalized using textproto.CanonicalMIMEHeaderKey.
//
// Example:
//
//	headers.Set("Content-Type", "application/json")
func (h *headers) Set(key, value string) {
	if h.head == nil {
		h.head = make(http.Header)
	}

	h.head.Set(key, value)
}

// Get returns the first value associated with the given header key.
// Header names are case-insensitive.
// Returns empty string if the header doesn't exist.
//
// To access multiple values or use non-canonical keys, access the
// underlying http.Header map directly.
//
// Example:
//
//	contentType := headers.Get("Content-Type")
func (h headers) Get(key string) string {
	if h.head == nil {
		h.head = make(http.Header)
	}

	return h.head.Get(key)
}

// Del removes all values associated with the given header key.
// Header names are canonicalized using textproto.CanonicalMIMEHeaderKey.
//
// Example:
//
//	headers.Del("X-Deprecated-Header")
func (h *headers) Del(key string) {
	if h.head == nil {
		h.head = make(http.Header)
	}

	h.head.Del(key)
}
