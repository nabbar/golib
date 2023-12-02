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
	"net/http"

	ginsdk "github.com/gin-gonic/gin"
)

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

func RoutersRegister(method string, relativePath string, router ...ginsdk.HandlerFunc) {
	defaultRouters.Register(method, relativePath, router...)
}

func RoutersRegisterInGroup(group, method string, relativePath string, router ...ginsdk.HandlerFunc) {
	defaultRouters.RegisterInGroup(group, method, relativePath, router...)
}

func RoutersHandler(engine *ginsdk.Engine) {
	defaultRouters.Handler(engine)
}

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

// SetGinHandler func that return given func as ginTonic HandlerFunc interface type.
func SetGinHandler(fct func(c *ginsdk.Context)) ginsdk.HandlerFunc {
	return fct
}

func Handler(routerList RouterList) http.Handler {
	e := routerList.Engine()

	if routerList == nil {
		RoutersHandler(e)
	} else {
		routerList.Handler(e)
	}

	return e
}
