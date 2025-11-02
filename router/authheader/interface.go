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

// Package authheader provides HTTP authorization header constants and helper functions.
// It defines standard authorization response codes and functions for handling
// authentication failures in Gin-based applications.
//
// The package is used by github.com/nabbar/golib/router/auth for authorization middleware.
//
// Example usage:
//
//	func authHandler(c *gin.Context) {
//	    token := c.GetHeader(authheader.HeaderAuthSend)
//	    if token == "" {
//	        authheader.AuthRequire(c, errors.New("missing token"))
//	        return
//	    }
//	    // Validate token...
//	}
//
// See also: github.com/nabbar/golib/router/auth
package authheader

import (
	"net/http"

	ginsdk "github.com/gin-gonic/gin"
)

// AuthCode represents the result of an authorization check.
type AuthCode uint8

const (
	// AuthCodeSuccess indicates that authorization was successful.
	// The request should proceed to the protected handler.
	AuthCodeSuccess AuthCode = iota

	// AuthCodeRequire indicates that authorization failed or is missing.
	// This typically results in HTTP 401 Unauthorized response.
	AuthCodeRequire

	// AuthCodeForbidden indicates that authorization succeeded but access is denied.
	// This typically results in HTTP 403 Forbidden response.
	AuthCodeForbidden
)

const (
	// HeaderAuthRequire is the HTTP header name for authentication challenges.
	// Used in 401 Unauthorized responses to indicate the authentication scheme.
	HeaderAuthRequire = "WWW-Authenticate"

	// HeaderAuthSend is the HTTP header name for sending credentials.
	// Clients include this header with their authentication information.
	HeaderAuthSend = "Authorization"

	// HeaderAuthReal is the default realm value for Basic authentication.
	// This is sent in the WWW-Authenticate header for LDAP-based auth.
	HeaderAuthReal = "Basic realm=LDAP Authorization Required"
)

// AuthRequire sends an HTTP 401 Unauthorized response and aborts the handler chain.
// It sets the WWW-Authenticate header to challenge the client for credentials.
//
// If an error is provided, it is added to the Gin context's error list for logging.
// The handler chain is aborted, preventing any subsequent handlers from executing.
//
// Parameters:
//   - c: Gin context
//   - err: Optional error to attach to the context (can be nil)
//
// Example:
//
//	if token == "" {
//	    authheader.AuthRequire(c, errors.New("missing authorization header"))
//	    return
//	}
func AuthRequire(c *ginsdk.Context, err error) {
	if err != nil {
		c.Errors = append(c.Errors, &ginsdk.Error{
			Err:  err,
			Type: ginsdk.ErrorTypePrivate,
		})
	}
	// Credentials doesn't match, we return 401 and abort handlers chain.
	c.Header(HeaderAuthRequire, HeaderAuthReal)
	c.AbortWithStatus(http.StatusUnauthorized)
}

// AuthForbidden sends an HTTP 403 Forbidden response and aborts the handler chain.
// This is used when authentication succeeded but the user is not authorized to access the resource.
//
// If an error is provided, it is added to the Gin context's error list for logging.
// The handler chain is aborted, preventing any subsequent handlers from executing.
//
// Parameters:
//   - c: Gin context
//   - err: Optional error to attach to the context (can be nil)
//
// Example:
//
//	if !hasPermission(user, resource) {
//	    authheader.AuthForbidden(c, errors.New("insufficient permissions"))
//	    return
//	}
func AuthForbidden(c *ginsdk.Context, err error) {
	if err != nil {
		c.Errors = append(c.Errors, &ginsdk.Error{
			Err:  err,
			Type: ginsdk.ErrorTypePrivate,
		})
	}
	c.AbortWithStatus(http.StatusForbidden)
}
