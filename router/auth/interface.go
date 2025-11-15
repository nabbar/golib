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

// Package auth provides HTTP authorization middleware for Gin-based applications.
// It supports custom authorization schemes (Bearer, Basic, API Key, etc.) with
// flexible authentication check functions.
//
// The package handles:
//   - Authorization header parsing and validation
//   - Custom authentication check functions
//   - HTTP 401 (Unauthorized) and 403 (Forbidden) responses
//   - Handler chain management
//
// Example usage:
//
//	checkFunc := func(token string) (authheader.AuthCode, error) {
//	    if validateToken(token) {
//	        return authheader.AuthCodeSuccess, nil
//	    }
//	    return authheader.AuthCodeForbidden, nil
//	}
//	auth := auth.NewAuthorization(logFunc, "BEARER", checkFunc)
//	engine.GET("/protected", auth.Register(protectedHandler))
//
// See also:
//   - github.com/nabbar/golib/router/authheader for auth response codes
//   - github.com/nabbar/golib/errors for error handling
//   - github.com/nabbar/golib/logger for logging
package auth

import (
	ginsdk "github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	rtrhdr "github.com/nabbar/golib/router/authheader"
)

// Authorization manages HTTP authorization for routes.
// It validates the Authorization header and executes registered handlers
// only if authorization succeeds.
type Authorization interface {
	// Handler is the main authorization middleware.
	// It checks the Authorization header and calls registered handlers if authorized.
	// Returns HTTP 401 if auth is missing/invalid, 403 if forbidden, 500 for other errors.
	Handler(c *ginsdk.Context)

	// Register sets the handlers to execute after successful authorization
	// and returns the Handler function for use as Gin middleware.
	Register(router ...ginsdk.HandlerFunc) ginsdk.HandlerFunc

	// Append adds additional handlers to the existing handler chain.
	// Useful for dynamically extending authorization-protected routes.
	Append(router ...ginsdk.HandlerFunc)
}

// NewAuthorization creates a new Authorization instance.
//
// Parameters:
//   - log: Logger function (can be nil for no logging)
//   - HeadAuthType: Expected authorization type (e.g., "BEARER", "BASIC")
//     Case-insensitive comparison is performed.
//   - authCheckFunc: Function to validate the authorization value
//     Returns AuthCodeSuccess, AuthCodeRequire, or AuthCodeForbidden
//
// Example:
//
//	auth := NewAuthorization(
//	    func() logger.Logger { return myLogger },
//	    "BEARER",
//	    func(token string) (authheader.AuthCode, error) {
//	        if isValidToken(token) {
//	            return authheader.AuthCodeSuccess, nil
//	        }
//	        return authheader.AuthCodeForbidden, errors.New("invalid token")
//	    },
//	)
func NewAuthorization(log liblog.FuncLog, HeadAuthType string, authCheckFunc func(AuthHeader string) (rtrhdr.AuthCode, liberr.Error)) Authorization {
	return &authorization{
		log:      log,
		check:    authCheckFunc,
		authType: HeadAuthType,
		router:   make([]ginsdk.HandlerFunc, 0),
	}
}
