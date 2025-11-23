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
 */

package static

import (
	ginsdk "github.com/gin-gonic/gin"
)

// SetSpecific registers a custom Gin handler for a specific route.
// This overrides the default static file serving for this route.
func (s *staticHandler) SetSpecific(group, route string, router ginsdk.HandlerFunc) {
	route = s.makeRoute(group, route)
	s.spc.Store(route, router)
}

// GetSpecific returns the custom handler for a specific route.
// Returns nil if no custom handler is registered for this route.
func (s *staticHandler) GetSpecific(group, route string) ginsdk.HandlerFunc {
	route = s.makeRoute(group, route)

	if i, l := s.spc.Load(route); !l {
		return nil
	} else if v, k := i.(ginsdk.HandlerFunc); !k {
		return nil
	} else {
		return v
	}
}
