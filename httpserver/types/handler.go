/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
 *
 */

package types

import "net/http"

// FuncHandler is the function signature for handler registration.
// It returns a map where keys are handler identifiers and values are http.Handler instances.
// The "default" key or empty string "" is used when no specific handler key is configured.
//
// Example:
//
//	func() map[string]http.Handler {
//	    return map[string]http.Handler{
//	        "":      defaultHandler,
//	        "api":   apiHandler,
//	        "admin": adminHandler,
//	    }
//	}
type FuncHandler func() map[string]http.Handler

// NewBadHandler creates a default error handler that returns HTTP 500 Internal Server Error.
// This handler is used as a fallback when no valid handler is registered for a server.
//
// Returns:
//   - http.Handler: A handler that always returns 500 status code
func NewBadHandler() http.Handler {
	return &BadHandler{}
}

// BadHandler is a default HTTP handler that returns 500 Internal Server Error for all requests.
// It's used as a fallback when no proper handler is configured for a server instance.
type BadHandler struct{}

// ServeHTTP implements http.Handler interface, returning HTTP 500 for all requests.
// This indicates that no valid handler was configured for the server.
func (o BadHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusInternalServerError)
}
