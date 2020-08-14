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
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/logger"
)

type AuthCode uint8

const (
	AUTH_CODE_SUCCESS = iota
	AUTH_CODE_REQUIRE
	AUTH_CODE_FORBIDDEN
)

const (
	HEAD_AUTH_REQR = "WWW-Authenticate"
	HEAD_AUTH_SEND = "Authorization"
	HEAD_AUTH_REAL = "Basic realm=LDAP Authorization Required"
)

func AuthRequire(c *gin.Context, err error) {
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{
			Err:  err,
			Type: gin.ErrorTypePrivate,
		})
	}
	// Credentials doesn't match, we return 401 and abort handlers chain.
	c.Header(HEAD_AUTH_REQR, HEAD_AUTH_REAL)
	c.AbortWithStatus(http.StatusUnauthorized)
}

func AuthForbidden(c *gin.Context, err error) {
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{
			Err:  err,
			Type: gin.ErrorTypePrivate,
		})
	}
	c.AbortWithStatus(http.StatusForbidden)
}

type authorization struct {
	check    func(AuthHeader string) (AuthCode, errors.Error)
	router   []gin.HandlerFunc
	authType string
}

type Authorization interface {
	Handler(c *gin.Context)
	Register(router ...gin.HandlerFunc) gin.HandlerFunc
	Append(router ...gin.HandlerFunc)
}

func NewAuthorization(HeadAuthType string, authCheckFunc func(AuthHeader string) (AuthCode, errors.Error)) Authorization {
	return &authorization{
		check:    authCheckFunc,
		authType: HeadAuthType,
		router:   make([]gin.HandlerFunc, 0),
	}
}

func (a *authorization) Register(router ...gin.HandlerFunc) gin.HandlerFunc {
	a.router = router
	return a.Handler
}

func (a *authorization) Append(router ...gin.HandlerFunc) {
	a.router = append(a.router, router...)
}

func (a authorization) Handler(c *gin.Context) {
	// Search user in the slice of allowed credentials
	auth := c.Request.Header.Get(HEAD_AUTH_SEND)

	if auth == "" {
		AuthRequire(c, HEADER_AUTH_MISSING.Error(nil).GetErrorFull(""))
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
		AuthRequire(c, HEADER_AUTH_EMPTY.Error(nil).GetErrorFull(""))
		return
	} else {
		code, err := a.check(authValue)

		switch code {
		case AUTH_CODE_SUCCESS:
			for _, r := range a.router {
				logger.DebugLevel.Logf("Calling router '%s=%s'", c.Request.Method, c.Request.URL.RawPath)
				r(c)
			}
		case AUTH_CODE_REQUIRE:
			AuthRequire(c, HEADER_AUTH_REQUIRE.Error(err).GetErrorFull(""))
		case AUTH_CODE_FORBIDDEN:
			AuthForbidden(c, HEADER_AUTH_FORBIDDEN.Error(err).GetErrorFull(""))
		default:
			c.Errors = append(c.Errors, &gin.Error{
				Err:  HEADER_AUTH_ERROR.Error(err).GetErrorFull(""),
				Type: gin.ErrorTypePrivate,
			})
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
