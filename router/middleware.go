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

func GinLatencyContext(c *ginsdk.Context) {
	// Start timer
	c.Set(GinContextStartUnixNanoTime, time.Now().UnixNano())

	// Process request
	c.Next()
}

func GinRequestContext(c *ginsdk.Context) {
	// Set Path
	if c != nil {
		if req := c.Request; req != nil {
			if u := req.URL; u != nil {
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

func GinAccessLog(log liblog.FuncLog) ginsdk.HandlerFunc {
	return func(c *ginsdk.Context) {
		// Process request
		c.Next()

		// Log only when path is not being skipped
		if log == nil {
			return
		} else if l := log(); l == nil {
			return
		} else {
			sttm := time.Unix(0, c.GetInt64(GinContextStartUnixNanoTime))
			path := c.GetString(GinContextRequestPath)
			user := c.GetString(GinContextRequestUser)

			ent := l.Access(
				c.ClientIP(),
				user,
				time.Now(),
				time.Now().Sub(sttm),
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

				if brokenPipe {
					rec = liberr.NewErrorRecovered("[Recovery] connection error", fmt.Sprintf("%s", err))
				} else {
					rec = liberr.NewErrorRecovered("[Recovery] panic recovered", fmt.Sprintf("%s", err))
				}

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

			// Log only when path is not being skipped
			if log == nil {
				return
			} else if l := log(); l == nil {
				return
			} else {
				if len(c.Errors) > 0 {
					for _, e := range c.Errors {
						ent := l.Entry(loglvl.ErrorLevel, "error on request \"%s %s %s\"", c.Request.Method, path, c.Request.Proto)
						ent.ErrorAdd(true, e)
						ent.Check(loglvl.NilLevel)
					}
				}
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
