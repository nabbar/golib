/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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
	"bytes"
	"encoding/json"
)

type requestError struct {
	c  int
	s  string
	se bool
	b  *bytes.Buffer
	be bool
	e  error
}

func (r *requestError) StatusCode() int {
	return r.c
}

func (r *requestError) Status() string {
	return r.s
}

func (r *requestError) Body() *bytes.Buffer {
	return r.b
}

func (r *requestError) Error() error {
	return r.e
}

func (r *requestError) IsError() bool {
	return r.se || r.be || r.e != nil
}

func (r *requestError) IsStatusError() bool {
	return r.se
}

func (r *requestError) IsBodyError() bool {
	return r.be
}

func (r *requestError) ParseBody(i interface{}) bool {
	if r.b != nil && r.b.Len() > 0 {
		if e := json.Unmarshal(r.b.Bytes(), i); e == nil {
			return true
		}
	}

	return false
}
