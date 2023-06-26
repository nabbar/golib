/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package entry

import (
	liberr "github.com/nabbar/golib/errors"
)

func (e *entry) ErrorClean() Entry {
	e.Error = make([]error, 0)
	return e
}

func (e *entry) ErrorSet(err []error) Entry {
	if len(err) < 1 {
		err = make([]error, 0)
	}
	e.Error = err
	return e
}

func (e *entry) ErrorAdd(cleanNil bool, err ...error) Entry {
	if len(e.Error) < 1 {
		e.Error = make([]error, 0)
	}

	for _, er := range err {
		if cleanNil && er == nil {
			continue
		}
		e.Error = append(e.Error, er)
	}

	return e
}

func (e *entry) ErrorAddLib(cleanNil bool, err ...liberr.Error) Entry {
	for _, er := range err {
		e.ErrorAdd(cleanNil, er.GetErrorSlice()...)
	}

	return e
}
