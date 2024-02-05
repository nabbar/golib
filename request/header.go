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
	"net/url"
)

func (r *request) ContentType(mime string) {
	r.SetHeader(_ContentType, mime)
}

func (r *request) ContentLength(size uint64) {
	r.SetHeader(__ContentLength, fmt.Sprintf("%d", size))
}

func (r *request) CleanHeader() {
	r.s.Lock()
	defer r.s.Unlock()

	r.h = make(url.Values)
}

func (r *request) DelHeader(key string) {
	r.s.Lock()
	defer r.s.Unlock()

	r.h.Del(key)
}

func (r *request) SetHeader(key, value string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.h) < 1 {
		r.h = make(url.Values)
	}

	r.h.Set(key, value)
}

func (r *request) AddHeader(key, value string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.h) < 1 {
		r.h = make(url.Values)
	}

	r.h.Add(key, value)
}
