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

	liberr "github.com/nabbar/golib/errors"

	"github.com/gin-gonic/gin"
	liblog "github.com/nabbar/golib/logger"
)

const (
	EmptyHandlerGroup           = "<nil>"
	GinContextStartUnixNanoTime = "gin-ctx-start-unix-nano-time"
	GinContextRequestPath       = "gin-ctx-request-path"
)

var (
	defaultRouters = NewRouterList(DefaultGinInit)
)

func GinEngine(trustedPlatform string, trustyProxy ...string) (*gin.Engine, error) {
	var err error

	engine := gin.New()
	if len(trustyProxy) > 0 {
		err = engine.SetTrustedProxies(trustyProxy)
	}
	if len(trustedPlatform) > 0 {
		engine.TrustedPlatform = trustedPlatform
	}

	return engine, err
}

func GinAddGlobalMiddleware(eng *gin.Engine, middleware ...gin.HandlerFunc) *gin.Engine {
	eng.Use(middleware...)
	return eng
}

func GinLatencyContext(c *gin.Context) {
	// Start timer
	c.Set(GinContextStartUnixNanoTime, time.Now().UnixNano())

	// Process request
	c.Next()
}

func GinRequestContext(c *gin.Context) {
	// Set Path
	path := c.Request.URL.Path

	if raw := c.Request.URL.RawQuery; len(raw) > 0 {
		path += "?" + raw
	}

	c.Set(GinContextRequestPath, path)

	// Process request
	c.Next()
}

func GinAccessLog(log liblog.FuncLog) gin.HandlerFunc {
	return func(c *gin.Context) {
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

			ent := l.Access(
				c.ClientIP(),
				c.Request.URL.User.Username(),
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

func GinErrorLog(log liblog.FuncLog) gin.HandlerFunc {
	return func(c *gin.Context) {
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

func DefaultGinInit() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	return engine
}

func DefaultGinWithTrustyProxy(trustyProxy []string) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	if len(trustyProxy) > 0 {
		_ = engine.SetTrustedProxies(trustyProxy)
	}

	return engine
}

func DefaultGinWithTrustedPlatform(trustedPlatform string) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	if len(trustedPlatform) > 0 {
		engine.TrustedPlatform = trustedPlatform
	}

	return engine
}

type routerItem struct {
	method   string
	relative string
	router   []gin.HandlerFunc
}

type routerList struct {
	init func() *gin.Engine
	list map[string][]routerItem
}

type RegisterRouter func(method string, relativePath string, router ...gin.HandlerFunc)
type RegisterRouterInGroup func(group, method string, relativePath string, router ...gin.HandlerFunc)

type RouterList interface {
	Register(method string, relativePath string, router ...gin.HandlerFunc)
	RegisterInGroup(group, method string, relativePath string, router ...gin.HandlerFunc)
	Handler(engine *gin.Engine)
	Engine() *gin.Engine
}

func RoutersRegister(method string, relativePath string, router ...gin.HandlerFunc) {
	defaultRouters.Register(method, relativePath, router...)
}

func RoutersRegisterInGroup(group, method string, relativePath string, router ...gin.HandlerFunc) {
	defaultRouters.RegisterInGroup(group, method, relativePath, router...)
}

func RoutersHandler(engine *gin.Engine) {
	defaultRouters.Handler(engine)
}

func NewRouterList(initGin func() *gin.Engine) RouterList {
	return &routerList{
		init: initGin,
		list: make(map[string][]routerItem),
	}
}

func (l routerList) Handler(engine *gin.Engine) {
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

func (l *routerList) RegisterInGroup(group, method string, relativePath string, router ...gin.HandlerFunc) {
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

func (l *routerList) Register(method string, relativePath string, router ...gin.HandlerFunc) {
	l.RegisterInGroup("", method, relativePath, router...)
}

func (l routerList) Engine() *gin.Engine {
	if l.init != nil {
		return l.init()
	} else {
		return DefaultGinInit()
	}
}
