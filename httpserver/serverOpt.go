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

package httpserver

import (
	"net/http"
	"time"

	liberr "github.com/nabbar/golib/errors"
	"golang.org/x/net/http2"
)

type optServer struct {
	ReadTimeout                  time.Duration
	ReadHeaderTimeout            time.Duration
	WriteTimeout                 time.Duration
	MaxHeaderBytes               int
	MaxHandlers                  int
	MaxConcurrentStreams         uint32
	MaxReadFrameSize             uint32
	PermitProhibitedCipherSuites bool
	IdleTimeout                  time.Duration
	MaxUploadBufferPerConnection int32
	MaxUploadBufferPerStream     int32
	DisableKeepAlive             bool
}

func (o *optServer) initServer(s *http.Server) liberr.Error {
	if o.ReadTimeout > 0 {
		s.ReadTimeout = o.ReadTimeout
	}

	if o.ReadHeaderTimeout > 0 {
		s.ReadHeaderTimeout = o.ReadHeaderTimeout
	} else {
		s.ReadHeaderTimeout = 30 * time.Second
	}

	if o.WriteTimeout > 0 {
		s.WriteTimeout = o.WriteTimeout
	}

	if o.MaxHeaderBytes > 0 {
		s.MaxHeaderBytes = o.MaxHeaderBytes
	}

	if o.IdleTimeout > 0 {
		s.IdleTimeout = o.IdleTimeout
	}

	if o.DisableKeepAlive {
		s.SetKeepAlivesEnabled(false)
	} else {
		s.SetKeepAlivesEnabled(true)
	}

	s2 := &http2.Server{}

	if o.MaxHandlers > 0 {
		s2.MaxHandlers = o.MaxHandlers
	}

	if o.MaxConcurrentStreams > 0 {
		s2.MaxConcurrentStreams = o.MaxConcurrentStreams
	}

	if o.PermitProhibitedCipherSuites {
		s2.PermitProhibitedCipherSuites = true
	}

	if o.IdleTimeout > 0 {
		s2.IdleTimeout = o.IdleTimeout
	}

	if o.MaxUploadBufferPerConnection > 0 {
		s2.MaxUploadBufferPerConnection = o.MaxUploadBufferPerConnection
	}

	if o.MaxUploadBufferPerStream > 0 {
		s2.MaxUploadBufferPerStream = o.MaxUploadBufferPerStream
	}

	if e := http2.ConfigureServer(s, s2); e != nil {
		return ErrorHTTP2Configure.Error(e)
	}

	return nil
}
