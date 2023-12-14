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

package startStop

func (o *run) getError() []error {
	if i := o.e.Load(); i == nil {
		return make([]error, 0)
	} else if d, k := i.([]error); !k {
		return make([]error, 0)
	} else {
		return d
	}
}

func (o *run) setError(err ...error) {
	if len(err) < 1 {
		err = make([]error, 0)
	}

	o.e.Store(err)
}

func (o *run) ErrorsLast() error {
	if o == nil {
		return ErrInvalid
	}

	e := o.getError()

	if len(e) > 0 {
		return e[len(e)-1]
	}

	return nil
}

func (o *run) ErrorsList() []error {
	if o == nil {
		return append(make([]error, 0), ErrInvalid)
	}

	return o.getError()
}

func (o *run) errorsAdd(e error) {
	if o == nil {
		return
	}

	o.setError(append(o.getError(), e)...)
}

func (o *run) errorsClean() {
	if o == nil {
		return
	}

	o.setError(nil)
}
