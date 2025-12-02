/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package entry

import liberr "github.com/nabbar/golib/errors"

// ErrorClean removes all errors from the entry by reinitializing the error slice to an empty slice.
// This method is useful when you want to reuse an entry for a new log statement without the
// previous errors.
//
// Returns:
//   - The entry itself for method chaining
//
// Example:
//
//	e := New(loglvl.ErrorLevel)
//	e.ErrorAdd(false, err1, err2)
//	e.ErrorClean() // Errors are now empty
func (e *entry) ErrorClean() Entry {
	e.Error = make([]error, 0)
	return e
}

// ErrorSet replaces the entire error slice of the entry with the provided slice. If the provided
// slice is empty or nil, the entry's error slice is set to an empty slice.
//
// This method is useful when you have a pre-existing slice of errors that you want to log.
//
// Parameters:
//   - err: The slice of errors to set. Can be nil or empty.
//
// Returns:
//   - The entry itself for method chaining
//
// Example:
//
//	errs := []error{err1, err2, err3}
//	e := New(loglvl.ErrorLevel).ErrorSet(errs)
func (e *entry) ErrorSet(err []error) Entry {
	if len(err) < 1 {
		err = make([]error, 0)
	}
	e.Error = err
	return e
}

// ErrorAdd appends one or more errors to the entry's error slice. If cleanNil is true, nil errors
// are filtered out and not added. If the error implements github.com/nabbar/golib/errors interface,
// its error slice is extracted and appended.
//
// This method is useful for accumulating errors from multiple operations before logging.
//
// Parameters:
//   - cleanNil: If true, nil errors are skipped. If false, all errors including nil are added.
//   - err: Variable number of errors to add to the entry
//
// Returns:
//   - The entry itself for method chaining
//
// Example:
//
//	e := New(loglvl.ErrorLevel)
//	e.ErrorAdd(true, err1, nil, err2) // Only err1 and err2 are added
//	e.ErrorAdd(false, err3, nil)      // Both err3 and nil are added
func (e *entry) ErrorAdd(cleanNil bool, err ...error) Entry {
	if len(e.Error) < 1 {
		e.Error = make([]error, 0)
	}

	for _, er := range err {
		if cleanNil && er == nil {
			continue
		}
		if liberr.Is(er) {
			e.Error = append(e.Error, liberr.Get(er).GetErrorSlice()...)
		} else {
			e.Error = append(e.Error, er)
		}
	}

	return e
}
