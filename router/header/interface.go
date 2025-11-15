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

// Package header provides HTTP header management for Gin-based applications.
// It offers a convenient interface for setting, getting, and managing HTTP headers
// across multiple routes and handlers.
//
// Key features:
//   - Add/Set/Get/Del operations for headers
//   - Middleware integration with Gin
//   - Handler chain registration
//   - Header cloning for reuse
//
// Example usage:
//
//	headers := header.NewHeaders()
//	headers.Set("X-API-Version", "v1")
//	headers.Set("Cache-Control", "no-cache")
//	engine.GET("/api/data", headers.Register(dataHandler)...)
//
// See also: github.com/nabbar/golib/router
package header

import (
	"net/http"

	ginsdk "github.com/gin-gonic/gin"
)

// Headers manages HTTP headers for routes and handlers.
// It provides methods to manipulate headers and integrate them into Gin middleware chains.
//
// All methods are safe for concurrent use when used with separate instances.
type Headers interface {
	// Add appends a value to the header.
	// If the header already exists, the value is added to the existing values.
	Add(key, value string)

	// Set replaces all values for the header with the given value.
	// Any existing values are discarded.
	Set(key, value string)

	// Get returns the first value for the given header key.
	// Returns empty string if the header doesn't exist.
	// Header names are case-insensitive.
	Get(key string) string

	// Del removes all values for the given header key.
	Del(key string)

	// Header returns a map of all headers with their first values.
	// Useful for inspecting or serializing header configuration.
	Header() map[string]string

	// Register creates a handler chain with the Header middleware first,
	// followed by the provided handlers. This ensures headers are set
	// before the route handlers execute.
	Register(router ...ginsdk.HandlerFunc) []ginsdk.HandlerFunc

	// Handler is the Gin middleware function that applies headers to the response.
	// It sets all configured headers on the Gin context.
	Handler(c *ginsdk.Context)

	// Clone creates a shallow copy of the Headers instance.
	// Note: The underlying header map is shared between instances.
	Clone() Headers
}

// NewHeaders creates a new Headers instance with an empty header map.
//
// Example:
//
//	headers := NewHeaders()
//	headers.Set("X-Custom-Header", "value")
func NewHeaders() Headers {
	return &headers{
		head: make(http.Header),
	}
}
