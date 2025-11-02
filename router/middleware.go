/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package router

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
)

// GinLatencyContext is a middleware that records the request start time.
// It stores the current time in nanoseconds in the Gin context under the key
// GinContextStartUnixNanoTime. This allows subsequent middleware or handlers
// to calculate request latency.
//
// Usage:
//
//	engine.Use(router.GinLatencyContext)
//
// To calculate latency in a handler:
//
//	startTime := time.Unix(0, c.GetInt64(router.GinContextStartUnixNanoTime))
//	latency := time.Since(startTime)
func GinLatencyContext(c *ginsdk.Context) {
	// Start timer
	c.Set(GinContextStartUnixNanoTime, time.Now().UnixNano())

	// Process request
	c.Next()
}

// GinRequestContext is a middleware that extracts and stores request information.
// It sanitizes and stores the request path (with query parameters) and username
// (if present in the URL) in the Gin context.
//
// Stored context keys:
//   - GinContextRequestPath: Sanitized request path with query string
//   - GinContextRequestUser: Username from URL (if present)
//
// The sanitization prevents log injection attacks by removing newlines, tabs, etc.
//
// Usage:
//
//	engine.Use(router.GinRequestContext)
func GinRequestContext(c *ginsdk.Context) {
	// Set Path
	if c != nil {
		if req := c.Request; req != nil {
			if u := req.URL; u != nil {
				// Extract and sanitize request path
				if p := u.Path; len(p) > 0 {
					if q := u.Query(); q != nil {
						if enc := q.Encode(); len(enc) > 0 {
							c.Set(GinContextRequestPath, sanitizeString(p+"?"+enc))
						} else {
							c.Set(GinContextRequestPath, sanitizeString(p))
						}
					} else {
						c.Set(GinContextRequestPath, sanitizeString(p))
					}
				}
				// Extract and sanitize username from URL
				if r := u.User; r != nil {
					if nm := r.Username(); len(nm) > 0 {
						c.Set(GinContextRequestUser, sanitizeString(nm))
					}
				}
			}
		}
	}

	// Process request
	c.Next()
}

// GinAccessLog is a middleware that logs HTTP access information.
// It should be used after GinLatencyContext and GinRequestContext to have
// access to timing and request path information.
//
// The middleware logs:
//   - Client IP address
//   - Authenticated user (if any)
//   - Request timestamp
//   - Request duration
//   - HTTP method
//   - Request path with query parameters
//   - HTTP protocol version
//   - Response status code
//   - Response size in bytes
//
// If log is nil or returns nil, no logging is performed.
//
// Usage:
//
//	logFunc := func() logger.Logger { return myLogger }
//	engine.Use(router.GinLatencyContext)
//	engine.Use(router.GinRequestContext)
//	engine.Use(router.GinAccessLog(logFunc))
//
// See also: github.com/nabbar/golib/logger
func GinAccessLog(log liblog.FuncLog) ginsdk.HandlerFunc {
	return func(c *ginsdk.Context) {
		// Process request
		c.Next()

		// Log only when logger is available
		if log == nil {
			return
		} else if l := log(); l == nil {
			return
		} else {
			// Retrieve timing and request information from context
			sttm := time.Unix(0, c.GetInt64(GinContextStartUnixNanoTime))
			path := c.GetString(GinContextRequestPath)
			user := c.GetString(GinContextRequestUser)

			// Create and log access entry
			ent := l.Access(
				c.ClientIP(),
				user,
				time.Now(),
				time.Since(sttm),
				c.Request.Method,
				path,
				c.Request.Proto,
				c.Writer.Status(),
				int64(c.Writer.Size()),
			)
			ent.Log()
		}
	}
}

// GinErrorLog is a middleware that handles panic recovery and error logging.
// It catches panics, logs errors from the Gin context, and sends appropriate
// HTTP responses. Should be used with GinRequestContext to have access to the
// request path.
//
// The middleware:
//   - Recovers from panics and converts them to errors
//   - Detects broken pipe errors (client disconnected)
//   - Logs all errors attached to the Gin context
//   - Logs recovered panics
//   - Returns 500 Internal Server Error for panics (except broken pipes)
//
// Special handling:
//   - Broken pipe errors: Connection is aborted without writing status
//   - Other panics: Returns HTTP 500 status
//
// If log is nil or returns nil, errors are recovered but not logged.
//
// Usage:
//
//	logFunc := func() logger.Logger { return myLogger }
//	engine.Use(router.GinRequestContext)
//	engine.Use(router.GinErrorLog(logFunc))
//
// See also:
//   - github.com/nabbar/golib/logger for logging
//   - github.com/nabbar/golib/errors for error types
func GinErrorLog(log liblog.FuncLog) ginsdk.HandlerFunc {
	return func(c *ginsdk.Context) {
		defer func() {
			var rec liberr.Error

			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// Create appropriate error based on type
				if brokenPipe {
					rec = liberr.NewErrorRecovered("[Recovery] connection error", fmt.Sprintf("%s", err))
				} else {
					rec = liberr.NewErrorRecovered("[Recovery] panic recovered", fmt.Sprintf("%s", err))
				}

				// Abort request with appropriate status
				if !c.IsAborted() {
					if brokenPipe {
						// If the connection is dead, we can't write a status to it.
						c.Abort()
					} else {
						c.AbortWithStatus(http.StatusInternalServerError)
					}
				}
			}

			path := c.GetString(GinContextRequestPath)

			// Log errors if logger is available
			if log == nil {
				return
			} else if l := log(); l == nil {
				return
			} else {
				// Log all errors from context
				if len(c.Errors) > 0 {
					for _, e := range c.Errors {
						ent := l.Entry(loglvl.ErrorLevel, "error on request \"%s %s %s\"", c.Request.Method, path, c.Request.Proto)
						ent.ErrorAdd(true, e)
						ent.Check(loglvl.NilLevel)
					}
				}
				// Log recovered panic
				if rec != nil {
					ent := l.Entry(loglvl.ErrorLevel, "error on request \"%s %s %s\"", c.Request.Method, path, c.Request.Proto)
					ent.ErrorAdd(true, rec)
					ent.Check(loglvl.NilLevel)
				}
			}
		}()
		c.Next()
	}
}
