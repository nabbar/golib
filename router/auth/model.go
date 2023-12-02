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

package auth

import (
	"fmt"
	"net/http"
	"strings"

	ginsdk "github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
	librtr "github.com/nabbar/golib/router"
	rtrhdr "github.com/nabbar/golib/router/authheader"
)

type authorization struct {
	log      liblog.FuncLog
	check    func(AuthHeader string) (rtrhdr.AuthCode, liberr.Error)
	router   []ginsdk.HandlerFunc
	authType string
}

func (a *authorization) Register(router ...ginsdk.HandlerFunc) ginsdk.HandlerFunc {
	a.router = router
	return a.Handler
}

func (a *authorization) Append(router ...ginsdk.HandlerFunc) {
	a.router = append(a.router, router...)
}

func (a *authorization) logDebug(msg string, args ...interface{}) {
	if a.log != nil {
		a.log().Entry(loglvl.DebugLevel, msg, args...)
	}
}

func (a *authorization) Handler(c *ginsdk.Context) {
	// Search user in the slice of allowed credentials
	auth := c.Request.Header.Get(rtrhdr.HeaderAuthSend)

	if auth == "" {
		rtrhdr.AuthRequire(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthMissing.Error(nil).GetErrorSlice()))
		return
	}

	authValue := ""

	if strings.ContainsAny(auth, " ") {
		sAuth := strings.SplitN(auth, " ", 2)
		if len(sAuth) == 2 && strings.ToUpper(sAuth[0]) == a.authType {
			authValue = sAuth[1]
		}
	}

	if authValue == "" {
		rtrhdr.AuthRequire(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthEmpty.Error(nil).GetErrorSlice()))
		return
	} else {
		code, err := a.check(authValue)

		switch code {
		case rtrhdr.AuthCodeSuccess:
			for _, r := range a.router {
				a.logDebug("Calling router '%s=%s'", c.Request.Method, c.Request.URL.RawPath)
				r(c)
			}
		case rtrhdr.AuthCodeRequire:
			rtrhdr.AuthRequire(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthRequire.Error(err).GetErrorSlice()))
		case rtrhdr.AuthCodeForbidden:
			rtrhdr.AuthForbidden(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthForbidden.Error(err).GetErrorSlice()))
		default:
			c.Errors = append(c.Errors, &ginsdk.Error{
				Err:  fmt.Errorf("%v", librtr.ErrorHeaderAuth.Error(err).GetErrorSlice()),
				Type: ginsdk.ErrorTypePrivate,
			})
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
