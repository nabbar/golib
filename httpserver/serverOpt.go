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

package httpserver

import (
	"net/http"
	"time"

	liberr "github.com/nabbar/golib/errors"
)

// optServer holds HTTP and HTTP/2 server configuration options.
// These options are applied to the underlying http.Server and http2.Server instances.
type optServer struct {
	ReadTimeout                   time.Duration // Maximum duration for reading entire request
	ReadHeaderTimeout             time.Duration // Maximum duration for reading request headers
	WriteTimeout                  time.Duration // Maximum duration for writing response
	MaxHeaderBytes                int           // Maximum header size in bytes
	MaxConcurrentStreams          int           // Maximum concurrent HTTP/2 streams per connection
	MaxReadFrameSize              uint32        // Maximum HTTP/2 frame size
	PermitProhibitedCipherSuites  bool          // Allow prohibited cipher suites for HTTP/2
	IdleTimeout                   time.Duration // Maximum idle time before closing connection
	MaxReceiveBufferPerConnection int           // HTTP/2 connection flow control window size
	MaxReceiveBufferPerStream     int           // HTTP/2 stream flow control window size
	DisableKeepAlive              bool          // Disable HTTP keep-alive connections
}

// initServer applies the configuration options to the http.Server and configures HTTP/2.
// It sets timeouts, header limits, keep-alive, and initializes HTTP/2 support.
// Returns an error if HTTP/2 configuration fails.
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

	if o.MaxConcurrentStreams > 0 {
		s.HTTP2.MaxConcurrentStreams = o.MaxConcurrentStreams
	}

	if o.PermitProhibitedCipherSuites {
		s.HTTP2.PermitProhibitedCipherSuites = true
	}

	if o.MaxReceiveBufferPerConnection > 0 {
		s.HTTP2.MaxReceiveBufferPerConnection = o.MaxReceiveBufferPerConnection
	}

	if o.MaxReceiveBufferPerStream > 0 {
		s.HTTP2.MaxReceiveBufferPerStream = o.MaxReceiveBufferPerStream
	}

	return nil
}
