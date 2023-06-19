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
	"sync"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
)

const (
	urlPathSeparator = "/"
)

type staticHandler struct {
	m sync.RWMutex
	l func() liblog.Logger

	c embed.FS
	b []string // base
	z int64    // size

	i libctx.Config[string] // index
	d libctx.Config[string] // download
	f libctx.Config[string] // follow
	s libctx.Config[string] // specific
	r *atomic.Value         // router
	h *atomic.Value         // monitor
}

func (s *staticHandler) _setLogger(fct func() liblog.Logger) {
	s.m.Lock()
	defer s.m.Unlock()

	s.l = fct
}

func (s *staticHandler) _getLogger() liblog.Logger {
	s.m.RLock()
	defer s.m.RUnlock()

	var log liblog.Logger

	if s.l == nil {
		s.m.RUnlock()
		log = s._getDefaultLogger()
		s.m.RLock()
		return log
	} else if log = s.l(); log == nil {
		s.m.RUnlock()
		log = s._getDefaultLogger()
		s.m.RLock()
		return log
	} else {
		return log
	}
}

func (s *staticHandler) _getDefaultLogger() liblog.Logger {
	s.m.Lock()
	defer s.m.Unlock()

	var log = liblog.New(s.d.GetContext)

	s.l = func() liblog.Logger {
		return log
	}

	return log
}

func (s *staticHandler) RegisterLogger(log func() liblog.Logger) {
	s._setLogger(log)
}
