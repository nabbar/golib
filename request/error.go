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
	code      int
	status    string
	statusErr bool
	bufBody   []byte
	bodyErr   bool
	checksum  []byte
	isSame    bool
	err       error
}

func (r *requestError) StatusCode() int {
	return r.code
}

func (r *requestError) Status() string {
	return r.status
}

func (r *requestError) Body() *bytes.Buffer {
	if len(r.bufBody) > 0 {
		b := make([]byte, 0, len(r.bufBody))
		copy(b, r.bufBody)
		return bytes.NewBuffer(b)
	}
	return bytes.NewBuffer(make([]byte, 0))
}

func (r *requestError) CheckSum() []byte {
	return r.checksum
}

func (r *requestError) Error() error {
	return r.err
}

func (r *requestError) IsError() bool {
	return r.statusErr || r.bodyErr || r.err != nil
}

func (r *requestError) IsStatusError() bool {
	return r.statusErr
}

func (r *requestError) IsBodyError() bool {
	return r.bodyErr
}

func (r *requestError) IsBodySame() bool {
	return r.isSame
}

func (r *requestError) ParseBody(i interface{}) bool {
	if len(r.bufBody) > 0 {
		if e := json.Unmarshal(r.bufBody, i); e == nil {
			return true
		}
	}

	return false
}

func (r *requestError) Free() {
	if len(r.bufBody) > 0 {
		r.bufBody = r.bufBody[:0]
	}
	r.bufBody = nil
}

func (r *request) newError() {
	r.err.Store(&requestError{
		code:      0,
		status:    "",
		statusErr: false,
		bufBody:   nil,
		bodyErr:   false,
		checksum:  nil,
		isSame:    false,
		err:       nil,
	})
}

func (r *request) getError() *requestError {
	if i := r.err.Load(); i != nil {
		if v, k := i.(*requestError); k {
			return v
		}
	}

	return &requestError{
		code:      0,
		status:    "",
		statusErr: false,
		bufBody:   nil,
		bodyErr:   false,
		checksum:  nil,
		isSame:    false,
		err:       nil,
	}
}

func (r *request) setError(e *requestError) {
	if e == nil {
		e = &requestError{
			code:      0,
			status:    "",
			statusErr: false,
			bufBody:   nil,
			bodyErr:   false,
			checksum:  nil,
			isSame:    false,
			err:       nil,
		}
	}

	r.err.Store(e)
}
