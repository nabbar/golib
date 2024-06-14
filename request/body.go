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
	"bytes"
	"encoding/json"
	"io"
)

const (
	contentTypeJson = "application/json"
)

type bds struct {
	b []byte
}

func (b *bds) Retry() io.Reader {
	var p = make([]byte, len(b.b), cap(b.b))
	copy(p, b.b)
	return bytes.NewBuffer(p)
}

type bdf struct {
	f io.ReadSeeker
	p int64
}

func (b *bdf) Retry() io.Reader {
	if _, e := b.f.Seek(b.p, io.SeekStart); e != nil {
		return nil
	} else {
		return b.f
	}
}

func (r *request) BodyJson(body interface{}) error {
	if p, e := json.Marshal(body); e != nil {
		return e
	} else if len(p) < 1 {
		return ErrorInvalidBody.Error()
	} else {
		r.ContentType(contentTypeJson)
		r.ContentLength(uint64(len(p)))
		r._BodyReader(&bds{
			b: p,
		})
	}

	return nil
}

func (r *request) BodyReader(body io.Reader, contentType string) error {
	if f, k := body.(io.ReadSeeker); k {
		if p, e := f.Seek(0, io.SeekStart); e != nil {
			return e
		} else {
			r._BodyReader(&bdf{
				f: f,
				p: p,
			})
		}
	} else {
		if p, e := io.ReadAll(body); e != nil {
			return e
		} else {
			r._BodyReader(&bds{
				b: p,
			})
		}
	}

	if contentType != "" {
		r.ContentType(contentType)
	}

	return nil
}

func (r *request) _BodyReader(body BodyRetryer) {
	r.bdr = body
}
