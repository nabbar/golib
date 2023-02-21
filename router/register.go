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
)

const (
	EmptyHandlerGroup           = "<nil>"
	GinContextStartUnixNanoTime = "gin-ctx-start-unix-nano-time"
	GinContextRequestPath       = "gin-ctx-request-path"
	GinContextRequestUser       = "gin-ctx-request-user"
)

var (
	defaultRouters = NewRouterList(DefaultGinInit)
)

func GinEngine(trustedPlatform string, trustyProxy ...string) (*ginsdk.Engine, error) {
	var err error

	engine := ginsdk.New()
	if len(trustyProxy) > 0 {
		err = engine.SetTrustedProxies(trustyProxy)
	}
	if len(trustedPlatform) > 0 {
		engine.TrustedPlatform = trustedPlatform
	}

	return engine, err
}

func GinAddGlobalMiddleware(eng *ginsdk.Engine, middleware ...ginsdk.HandlerFunc) *ginsdk.Engine {
	eng.Use(middleware...)
	return eng
}

func GinLatencyContext(c *ginsdk.Context) {
	// Start timer
	c.Set(GinContextStartUnixNanoTime, time.Now().UnixNano())

	// Process request
	c.Next()
}

func sanitizeString(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return s
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
			defer func() {
				if l != nil {
					_ = l.Close()
				}
			}()

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
				defer func() {
					if l != nil {
						_ = l.Close()
					}
				}()

				if len(c.Errors) > 0 {
					for _, e := range c.Errors {
						ent := l.Entry(liblog.ErrorLevel, "error on request \"%s %s %s\"", c.Request.Method, path, c.Request.Proto)
						ent.ErrorAdd(true, e)
						ent.Check(liblog.NilLevel)
					}
				}
				if rec != nil {
					ent := l.Entry(liblog.ErrorLevel, "error on request \"%s %s %s\"", c.Request.Method, path, c.Request.Proto)
					ent.ErrorAdd(true, rec)
					ent.Check(liblog.NilLevel)
				}
			}
		}()
		c.Next()
	}
}

func DefaultGinInit() *ginsdk.Engine {
	engine := ginsdk.New()
	engine.Use(ginsdk.Logger(), ginsdk.Recovery())

	return engine
}

func DefaultGinWithTrustyProxy(trustyProxy []string) *ginsdk.Engine {
	engine := ginsdk.New()
	engine.Use(ginsdk.Logger(), ginsdk.Recovery())

	if len(trustyProxy) > 0 {
		_ = engine.SetTrustedProxies(trustyProxy)
	}

	return engine
}

func DefaultGinWithTrustedPlatform(trustedPlatform string) *ginsdk.Engine {
	engine := ginsdk.New()
	engine.Use(ginsdk.Logger(), ginsdk.Recovery())

	if len(trustedPlatform) > 0 {
		engine.TrustedPlatform = trustedPlatform
	}

	return engine
}

type routerItem struct {
	method   string
	relative string
	router   []ginsdk.HandlerFunc
}

type routerList struct {
	init func() *ginsdk.Engine
	list map[string][]routerItem
}

type RegisterRouter func(method string, relativePath string, router ...ginsdk.HandlerFunc)
type RegisterRouterInGroup func(group, method string, relativePath string, router ...ginsdk.HandlerFunc)

type RouterList interface {
	Register(method string, relativePath string, router ...ginsdk.HandlerFunc)
	RegisterInGroup(group, method string, relativePath string, router ...ginsdk.HandlerFunc)
	Handler(engine *ginsdk.Engine)
	Engine() *ginsdk.Engine
}

func RoutersRegister(method string, relativePath string, router ...ginsdk.HandlerFunc) {
	defaultRouters.Register(method, relativePath, router...)
}

func RoutersRegisterInGroup(group, method string, relativePath string, router ...ginsdk.HandlerFunc) {
	defaultRouters.RegisterInGroup(group, method, relativePath, router...)
}

func RoutersHandler(engine *ginsdk.Engine) {
	defaultRouters.Handler(engine)
}

func NewRouterList(initGin func() *ginsdk.Engine) RouterList {
	return &routerList{
		init: initGin,
		list: make(map[string][]routerItem),
	}
}

func (l routerList) Handler(engine *ginsdk.Engine) {
	for grpRoute, grpList := range l.list {
		if grpRoute == EmptyHandlerGroup {
			for _, r := range grpList {
				engine.Handle(r.method, r.relative, r.router...)
			}
		} else {
			var grp = engine.Group(grpRoute)
			for _, r := range grpList {
				grp.Handle(r.method, r.relative, r.router...)
			}
		}
	}
}

func (l *routerList) RegisterInGroup(group, method string, relativePath string, router ...ginsdk.HandlerFunc) {
	if group == "" {
		group = EmptyHandlerGroup
	}

	if _, ok := l.list[group]; !ok {
		l.list[group] = make([]routerItem, 0)
	}

	l.list[group] = append(l.list[group], routerItem{
		method:   method,
		relative: relativePath,
		router:   router,
	})
}

func (l *routerList) Register(method string, relativePath string, router ...ginsdk.HandlerFunc) {
	l.RegisterInGroup("", method, relativePath, router...)
}

func (l routerList) Engine() *ginsdk.Engine {
	if l.init != nil {
		return l.init()
	} else {
		return DefaultGinInit()
	}
}
