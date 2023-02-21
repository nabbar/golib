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

package logger

import "fmt"

func (o *logger) Close() error {
	if o == nil {
		return fmt.Errorf("not initialized")
	} else if o.c == nil {
		return fmt.Errorf("not initialized")
	}

	_ = o.c.Close()
	o.c.Clean()

	return nil
}

func (o *logger) Write(p []byte) (n int, err error) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	o.newEntry(o.GetIOWriterLevel(), string(p), nil, o.GetFields(), nil).Log()
	return len(p), nil
}

func (o *logger) SetIOWriterLevel(lvl Level) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	o.x.Store(keyWriter, lvl)
}

func (o *logger) GetIOWriterLevel() Level {
	if o == nil {
		return NilLevel
	} else if o.x == nil {
		return NilLevel
	} else if i, l := o.x.Load(keyWriter); !l {
		return NilLevel
	} else if v, k := i.(Level); !k {
		return NilLevel
	} else {
		return v
	}
}
