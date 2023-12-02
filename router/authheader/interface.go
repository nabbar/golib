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

package authheader

import (
	"net/http"

	ginsdk "github.com/gin-gonic/gin"
)

type AuthCode uint8

const (
	AuthCodeSuccess = iota
	AuthCodeRequire
	AuthCodeForbidden
)

const (
	HeaderAuthRequire = "WWW-Authenticate"
	HeaderAuthSend    = "Authorization"
	HeaderAuthReal    = "Basic realm=LDAP Authorization Required"
)

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

func AuthForbidden(c *ginsdk.Context, err error) {
	if err != nil {
		c.Errors = append(c.Errors, &ginsdk.Error{
			Err:  err,
			Type: ginsdk.ErrorTypePrivate,
		})
	}
	c.AbortWithStatus(http.StatusForbidden)
}
