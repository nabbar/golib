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

package static

import (
	"embed"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
)

const (
	urlPathSeparator = "/"
)

type staticHandler struct {
	l *atomic.Value // logger
	r *atomic.Value // router
	h *atomic.Value // monitor

	c embed.FS
	b *atomic.Value // base []string
	z *atomic.Int64 // size

	i libctx.Config[string] // index
	d libctx.Config[string] // download
	f libctx.Config[string] // follow
	s libctx.Config[string] // specific
}

func (s *staticHandler) _setLogger(fct liblog.FuncLog) {
	if fct == nil {
		fct = s._getDefaultLogger
	}

	s.l.Store(fct())
}

func (s *staticHandler) _getLogger() liblog.Logger {
	i := s.l.Load()

	if i == nil {
		return s._getDefaultLogger()
	} else if l, k := i.(liblog.FuncLog); !k {
		return s._getDefaultLogger()
	} else if log := l(); log == nil {
		return s._getDefaultLogger()
	} else {
		return log
	}
}

func (s *staticHandler) _getDefaultLogger() liblog.Logger {
	return liblog.New(s.d.GetContext)
}

func (s *staticHandler) RegisterLogger(log liblog.FuncLog) {
	s._setLogger(log)
}
