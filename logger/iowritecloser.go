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

import "sync/atomic"

func (l *logger) Close() error {

	for _, c := range l.closeGetMutex() {
		if c != nil {
			_ = c.Close()
		}
	}

	if l.n != nil {
		l.n()
	}

	l.closeClean()

	return nil
}

func (l *logger) Write(p []byte) (n int, err error) {
	l.newEntry(l.GetIOWriterLevel(), string(p), nil, l.GetFields(), nil).Log()
	return len(p), nil
}

func (l *logger) SetIOWriterLevel(lvl Level) {
	if l == nil {
		return
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.w == nil {
		l.w = new(atomic.Value)
	}

	l.w.Store(lvl)
}

func (l *logger) GetIOWriterLevel() Level {
	if l == nil {
		return NilLevel
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.w == nil {
		l.w = new(atomic.Value)
	}

	if i := l.w.Load(); i == nil {
		return NilLevel
	} else if o, ok := i.(Level); ok {
		return o
	}

	return NilLevel
}
