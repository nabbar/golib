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

// authorization is the internal implementation of the Authorization interface.
type authorization struct {
	log      liblog.FuncLog                                          // Logger function
	check    func(AuthHeader string) (rtrhdr.AuthCode, liberr.Error) // Authorization check function
	router   []ginsdk.HandlerFunc                                    // Handler chain to execute on success
	authType string                                                  // Expected authorization type (e.g., "BEARER")
}

// Register sets the handler chain and returns the Handler function.
// The provided handlers will only be executed if authorization succeeds.
func (a *authorization) Register(router ...ginsdk.HandlerFunc) ginsdk.HandlerFunc {
	a.router = router
	return a.Handler
}

// Append adds handlers to the existing handler chain.
// This allows extending the protected route with additional middleware or handlers.
func (a *authorization) Append(router ...ginsdk.HandlerFunc) {
	a.router = append(a.router, router...)
}

// logDebug logs a debug message if a logger is configured.
func (a *authorization) logDebug(msg string, args ...interface{}) {
	if a.log != nil {
		a.log().Entry(loglvl.DebugLevel, msg, args...)
	}
}

// Handler is the main authorization middleware function.
// It performs the following steps:
//  1. Extracts the Authorization header
//  2. Validates the authorization type matches the expected type
//  3. Calls the custom check function with the authorization value
//  4. Executes registered handlers if authorized, or returns appropriate error
//
// HTTP responses:
//   - 401 Unauthorized: Missing or invalid authorization header
//   - 403 Forbidden: Authorization check returned AuthCodeForbidden
//   - 500 Internal Server Error: Unknown auth code returned
//   - Success: Executes registered handler chain
func (a *authorization) Handler(c *ginsdk.Context) {
	// Extract Authorization header
	auth := c.Request.Header.Get(rtrhdr.HeaderAuthSend)

	if auth == "" {
		// No authorization header provided
		rtrhdr.AuthRequire(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthMissing.Error(nil).GetErrorSlice()))
		return
	}

	authValue := ""

	// Parse "Type Value" format (e.g., "Bearer token123")
	if strings.ContainsAny(auth, " ") {
		sAuth := strings.SplitN(auth, " ", 2)
		if len(sAuth) == 2 && strings.ToUpper(sAuth[0]) == a.authType {
			authValue = sAuth[1]
		}
	}

	if authValue == "" {
		// Authorization type doesn't match or value is empty
		rtrhdr.AuthRequire(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthEmpty.Error(nil).GetErrorSlice()))
		return
	} else {
		// Call custom authorization check function
		code, err := a.check(authValue)

		switch code {
		case rtrhdr.AuthCodeSuccess:
			// Authorization successful, execute handler chain
			for _, r := range a.router {
				a.logDebug("Calling router '%s=%s'", c.Request.Method, c.Request.URL.RawPath)
				r(c)
			}
		case rtrhdr.AuthCodeRequire:
			// Authorization failed, require authentication
			rtrhdr.AuthRequire(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthRequire.Error(err).GetErrorSlice()))
		case rtrhdr.AuthCodeForbidden:
			// Authorization succeeded but access is forbidden
			rtrhdr.AuthForbidden(c, fmt.Errorf("%v", librtr.ErrorHeaderAuthForbidden.Error(err).GetErrorSlice()))
		default:
			// Unknown authorization code
			c.Errors = append(c.Errors, &ginsdk.Error{
				Err:  fmt.Errorf("%v", librtr.ErrorHeaderAuth.Error(err).GetErrorSlice()),
				Type: ginsdk.ErrorTypePrivate,
			})
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
