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

import "net/url"

func (r *request) CleanParams() {
	r.prm = make(url.Values)
}

func (r *request) DelParams(key string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.prm.Del(key)
}

func (r *request) SetParams(key, val string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if len(r.prm) < 1 {
		r.prm = make(url.Values)
	}

	r.prm.Set(key, val)
}

func (r *request) AddParams(key, val string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if len(r.prm) < 1 {
		r.prm = make(url.Values)
	}

	r.prm.Set(key, val)
}

func (r *request) GetFullUrl() *url.URL {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.uri
}

func (r *request) SetFullUrl(u *url.URL) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.uri = u
}
