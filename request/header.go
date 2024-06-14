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

package request

import (
	"fmt"
	"net/http"
)

func (r *request) ContentType(mime string) {
	r.SetHeader(_ContentType, mime)
}

func (r *request) ContentLength(size uint64) {
	r.SetHeader(__ContentLength, fmt.Sprintf("%d", size))
}

func (r *request) CleanHeader() {
	var k = make([]any, 0)

	r.hdr.Range(func(key, value any) bool {
		k = append(k, key)
		return true
	})

	for _, key := range k {
		r.hdr.Delete(key)
	}
}

func (r *request) DelHeader(key string) {
	r.hdr.Delete(key)
}

func (r *request) SetHeader(key, value string) {
	r.hdr.Store(key, append(make([]string, 0), value))
}

func (r *request) AddHeader(key, value string) {
	var val []string
	if i, l := r.hdr.Load(key); i != nil && l {
		if v, k := i.([]string); k && len(v) > 0 {
			val = v
		}
	}

	if len(val) < 1 {
		val = make([]string, 0)
	}

	r.hdr.Store(key, append(val, value))
}

func (r *request) GetHeader(key string) []string {
	if i, l := r.hdr.Load(key); i != nil && l {
		if v, k := i.([]string); k && len(v) > 0 {
			return v
		}
	}

	return make([]string, 0)
}

func (r *request) httpHeader() http.Header {
	var hdr = make(http.Header)

	r.hdr.Range(func(key, value any) bool {
		if u, k := key.(string); !k {
			return true
		} else if v, l := value.([]string); !l {
			return true
		} else {
			hdr[u] = v
			return true
		}
	})

	return hdr
}
