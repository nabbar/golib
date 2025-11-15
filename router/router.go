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

import ginsdk "github.com/gin-gonic/gin"

// itm represents a single route item with its method, path, and handlers.
// It is used internally by RouterList to store route information.
type itm struct {
	method   string               // HTTP method (GET, POST, etc.)
	relative string               // Relative path for the route
	router   []ginsdk.HandlerFunc // Chain of handler functions
}

// Same checks if this route item matches the given method and relative path.
// Returns true if both method and path are identical, false otherwise.
// This is used by RegisterMergeInGroup to find existing routes.
func (o *itm) Same(method, relative string) bool {
	if o.method != method {
		return false
	}
	if o.relative != relative {
		return false
	}
	return true
}

// Merge replaces the current handler chain with the provided handlers.
// This is used by RegisterMergeInGroup to update route handlers.
func (o *itm) Merge(rtr ...ginsdk.HandlerFunc) {
	o.router = rtr
}
