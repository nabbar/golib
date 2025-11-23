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

// SetRedirect configures a permanent redirect (HTTP 301) from source to destination route.
// Useful for maintaining backward compatibility or reorganizing file structure.
func (s *staticHandler) SetRedirect(srcGroup, srcRoute, dstGroup, dstRoute string) {
	srcRoute = s.makeRoute(srcGroup, srcRoute)
	dstRoute = s.makeRoute(dstGroup, dstRoute)

	s.flw.Store(srcRoute, dstRoute)
}

// GetRedirect returns the destination route for a given source route.
// Returns empty string if no redirect is configured.
func (s *staticHandler) GetRedirect(srcGroup, srcRoute string) string {
	srcRoute = s.makeRoute(srcGroup, srcRoute)
	if i, l := s.flw.Load(srcRoute); !l {
		return ""
	} else if v, k := i.(string); !k {
		return ""
	} else {
		return v
	}
}

// IsRedirect checks if a route is configured as a redirect.
func (s *staticHandler) IsRedirect(group, route string) bool {
	route = s.makeRoute(group, route)

	if i, l := s.flw.Load(route); !l {
		return false
	} else {
		_, ok := i.(string)
		return ok
	}
}
