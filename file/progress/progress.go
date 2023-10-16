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

package progress

import (
	"errors"
	"io"
)

func (o *progress) RegisterFctIncrement(fct FctIncrement) {
	if fct == nil {
		fct = func(size int64) {}
	}

	o.fi.Store(fct)
}

func (o *progress) RegisterFctReset(fct FctReset) {
	if fct == nil {
		fct = func(size, current int64) {}
	}

	o.fr.Store(fct)
}

func (o *progress) RegisterFctEOF(fct FctEOF) {
	if fct == nil {
		fct = func() {}
	}

	o.fe.Store(fct)
}

func (o *progress) SetRegisterProgress(f Progress) {
	i := o.fi.Load()
	if i != nil {
		f.RegisterFctIncrement(i.(FctIncrement))
	}

	i = o.fr.Load()
	if i != nil {
		f.RegisterFctReset(i.(FctReset))
	}

	i = o.fe.Load()
	if i != nil {
		f.RegisterFctEOF(i.(FctEOF))
	}
}

func (o *progress) inc(n int64) {
	if o == nil {
		return
	}

	f := o.fi.Load()
	if f != nil {
		f.(FctIncrement)(n)
	}
}

func (o *progress) incN(n int64, s int) {
	if o == nil {
		return
	}

	f := o.fi.Load()
	if f != nil {
		f.(FctIncrement)(n + int64(s))
	}
}

func (o *progress) dec(n int64) {
	if o == nil {
		return
	}

	f := o.fi.Load()
	if f != nil {
		f.(FctIncrement)(0 - n)
	}
}

func (o *progress) decN(n int64, s int) {
	if o == nil {
		return
	}

	f := o.fi.Load()
	if f != nil {
		f.(FctIncrement)(0 - (n + int64(s)))
	}
}

func (o *progress) finish() {
	if o == nil {
		return
	}

	f := o.fe.Load()
	if f != nil {
		f.(FctEOF)()
	}
}

func (o *progress) reset() {
	o.Reset(0)
}

func (o *progress) Reset(max int64) {
	if o == nil {
		return
	}

	f := o.fr.Load()

	if f != nil {
		if max < 1 {
			if i, e := o.Stat(); e != nil {
				return
			} else {
				max = i.Size()
			}
		}

		if s, e := o.SizeBOF(); e != nil {
			return
		} else if s >= 0 {
			f.(FctReset)(max, s)
		}
	}
}

func (o *progress) analyze(i int, e error) (n int, err error) {
	if o == nil {
		return i, e
	}

	if i != 0 {
		o.inc(int64(i))
	}

	if e != nil && errors.Is(e, io.EOF) {
		o.finish()
	}

	return i, e
}

func (o *progress) analyze64(i int64, e error) (n int64, err error) {
	if o == nil {
		return i, e
	}

	if i != 0 {
		o.inc(i)
	}

	if e != nil && errors.Is(e, io.EOF) {
		o.finish()
	}

	return i, e
}
