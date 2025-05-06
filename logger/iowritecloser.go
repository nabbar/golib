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

import (
	"bytes"
	"strings"

	loglvl "github.com/nabbar/golib/logger/level"
)

func (o *logger) Close() error {
	if o != nil && o.hasCloser() {
		o.switchCloser(nil)
	}

	return nil
}

func (o *logger) Write(p []byte) (n int, err error) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	val := strings.TrimSpace(string(o.IOWriterFilter(p)))

	if len(val) < 1 {
		return len(p), nil
	}

	o.newEntry(o.GetIOWriterLevel(), val, nil, o.GetFields(), nil).Log()
	return len(p), nil
}

func (o *logger) SetIOWriterLevel(lvl loglvl.Level) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	o.x.Store(keyWriter, lvl)
}
func (o *logger) GetIOWriterLevel() loglvl.Level {
	if o == nil {
		return loglvl.NilLevel
	} else if o.x == nil {
		return loglvl.NilLevel
	} else if i, l := o.x.Load(keyWriter); !l {
		return loglvl.NilLevel
	} else if v, k := i.(loglvl.Level); !k {
		return loglvl.NilLevel
	} else {
		return v
	}
}

func (o *logger) SetIOWriterFilter(pattern ...string) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	var p = make([][]byte, 0, len(pattern))
	for _, s := range pattern {
		p = append(p, []byte(s))
	}

	o.x.Store(keyFilter, p)
}

func (o *logger) AddIOWriterFilter(pattern ...string) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	var p = make([][]byte, 0, len(pattern))

	if i, l := o.x.Load(keyFilter); !l {
		// nothing
	} else if v, k := i.([][]byte); !k {
		// nothing
	} else {
		p = append(make([][]byte, 0, len(pattern)+len(v)), v...)
	}

	for _, s := range pattern {
		p = append(p, []byte(s))
	}

	o.x.Store(keyFilter, p)
}

func (o *logger) IOWriterFilter(p []byte) []byte {
	if o == nil {
		return p
	} else if o.x == nil {
		return p
	} else if i, l := o.x.Load(keyFilter); !l {
		return p
	} else if v, k := i.([][]byte); !k {
		return p
	} else {
		for _, b := range v {
			if bytes.Contains(p, b) {
				return make([]byte, 0)
			}
		}

		return p
	}
}
