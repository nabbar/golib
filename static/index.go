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

package static

import (
	"slices"
)

func (s *staticHandler) SetIndex(group, route, pathFile string) {
	if pathFile != "" && s.Has(pathFile) {
		var lst []string

		if i, l := s.i.Load(pathFile); !l {
			lst = make([]string, 0)
		} else if v, k := i.([]string); !k {
			lst = make([]string, 0)
		} else {
			lst = v
		}

		s.i.Store(pathFile, append(lst, s._makeRoute(group, route)))
	}
}

func (s *staticHandler) GetIndex(group, route string) string {
	route = s._makeRoute(group, route)
	var found string

	s.i.Walk(func(key string, val interface{}) bool {
		if v, k := val.([]string); !k {
			return true
		} else if !slices.Contains(v, route) {
			return true
		}

		found = key
		return false
	})

	return found
}

func (s *staticHandler) IsIndex(pathFile string) bool {
	if i, l := s.i.Load(pathFile); !l {
		return false
	} else {
		_, ok := i.([]string)
		return ok
	}
}

func (s *staticHandler) IsIndexForRoute(pathFile, group, route string) bool {
	if i, l := s.i.Load(pathFile); !l {
		return false
	} else if v, k := i.([]string); !k {
		return false
	} else {
		return slices.Contains(v, s._makeRoute(group, route))
	}
}
