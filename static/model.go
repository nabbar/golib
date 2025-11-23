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
 */

package static

import (
	"context"
	"embed"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
)

const (
	urlPathSeparator = "/"
)

// staticHandler is the internal implementation of the Static interface.
// All fields use atomic operations for thread-safe access without mutexes.
//
// The structure is organized into logical groups:
//   - Core: logger, router, embedded filesystem
//   - File configuration: index, download, follow, specific
//   - Security: rate limiting, path security, suspicious detection
//   - HTTP: headers, caching
//   - Integration: WAF/IDS/EDR security backend
type staticHandler struct {
	log libatm.Value[liblog.Logger] // logger instance
	rtr libatm.Value[[]string]      // registered routes

	efs embed.FS               // embedded filesystem
	bph libatm.Value[[]string] // base paths within embed.FS
	siz *atomic.Int64          // total size counter

	idx libctx.Config[string] // index file configuration
	dwn libctx.Config[string] // download file configuration
	flw libctx.Config[string] // redirect configuration
	spc libctx.Config[string] // specific handler configuration

	// Rate limiting
	rlc libatm.Value[*RateLimitConfig]    // rate limit configuration
	rli libatm.MapTyped[string, *ipTrack] // IP tracking map (IP -> tracking data)
	rlx libatm.Value[context.CancelFunc]  // cleanup goroutine cancel function

	// Path security
	psc libatm.Value[*PathSecurityConfig] // path security configuration

	// Suspicious access detection
	sus libatm.Value[*SuspiciousConfig] // suspicious access configuration

	// HTTP Headers
	hdr libatm.Value[*HeadersConfig] // headers configuration (cache, etag, content-type)

	// Security backend integration (WAF/IDS/EDR)
	sec libatm.Value[*SecurityConfig] // security integration configuration
	seb libatm.Value[*evtBatch]       // event batch for batching security events
}

func (s *staticHandler) setLogger(log liblog.Logger) {
	if log == nil {
		log = liblog.New(s.dwn)
	}

	s.log.Store(log)
}

func (s *staticHandler) getLogger() liblog.Logger {
	i := s.log.Load()

	if i == nil {
		return s.defLogger()
	} else {
		return i
	}
}

func (s *staticHandler) defLogger() liblog.Logger {
	return liblog.New(s.dwn)
}

func (s *staticHandler) RegisterLogger(log liblog.Logger) {
	s.setLogger(log)
}
