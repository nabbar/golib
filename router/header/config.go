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
	liberr "github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/router"
)

// HeadersConfig is a map-based configuration for HTTP headers.
// It provides a simple way to define headers in configuration files
// and convert them to a Headers instance.
//
// Example:
//
//	config := HeadersConfig{
//	    "X-API-Version": "v1",
//	    "Cache-Control": "no-cache",
//	    "X-Request-ID":  "12345",
//	}
//	headers := config.New()
type HeadersConfig map[string]string

// New creates a Headers instance from the configuration.
// Each key-value pair in the map is added as a header.
//
// Returns a Headers instance with all configured headers.
//
// Example:
//
//	config := HeadersConfig{"X-Custom": "value"}
//	headers := config.New()
//	engine.GET("/api", headers.Register(handler)...)
func (h HeadersConfig) New() Headers {
	var res = NewHeaders()

	for k, v := range h {
		res.Add(k, v)
	}

	return res
}

// Validate checks if the configuration is valid.
// Currently, this always returns nil as all string key-value pairs are valid headers.
//
// This method exists for consistency with other configuration types and may
// be extended in the future to validate header names or values.
//
// Returns nil if validation succeeds, or an error if validation fails.
func (h HeadersConfig) Validate() liberr.Error {
	err := router.ErrorConfigValidator.Error(nil)

	if !err.HasParent() {
		err = nil
	}

	return err
}
