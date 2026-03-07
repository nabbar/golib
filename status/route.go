/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
 *
 */

package status

import (
	"context"
	"strconv"
	"strings"

	ginsdk "github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
)

const (
	// HeadVerbose is the HTTP header name for controlling verbosity.
	// Value should be "true" or "false". If "false", full component details are included.
	HeadVerbose = "X-Verbose"

	// HeadFormat is the HTTP header name for content negotiation (Accept header).
	// Supports "application/json" and "text/plain".
	HeadFormat = "Accept"

	// HeadMapMode is the HTTP header name for output mode map or not.
	// Value should be "true" or "false". If "false", full component details are included.
	HeadMapMode = "X-MapMode"

	// QueryVerbose is the query parameter name for controlling verbosity.
	// Value should be "true" or "1" for short output (no component details).
	QueryVerbose = "short"

	// QueryFormat is the query parameter name for output format.
	// Value should be "text" for plain text or "json" for JSON output.
	QueryFormat = "format"

	// QueryMapMode is the query parameter name for output mode map or not.
	// Value should be "true" or "1" for short output (no component details).
	QueryMapMode = "map"
)

// Expose handles the status endpoint request from a generic context.
// If the context is a Gin context (*ginsdk.Context), it delegates to MiddleWare.
// This method allows using the status handler with generic context.Context.
//
// This is useful when integrating with frameworks that use context.Context
// instead of directly passing *gin.Context.
func (o *sts) Expose(ctx context.Context) {
	if c, ok := ctx.(*ginsdk.Context); ok {
		o.MiddleWare(c)
	}
}

// MiddleWare is the Gin middleware handler that processes status requests.
// It determines the response format and content based on headers and query parameters.
//
// Verbosity (short vs full output):
//   - Header "X-Verbose: true" forces full output (disables short mode).
//   - Query parameter "short=true" or "short=1" enables short mode (only overall status).
//   - Default: full status with all component details.
//
// Map Mode (structured map vs list):
//   - Header "X-MapMode: true" enables map mode.
//   - Query parameter "map=true" or "map=1" enables map mode.
//   - Default: list mode.
//
// Format (JSON vs plain text):
//   - Query parameter "format=text" forces plain text.
//   - Header "Accept: text/plain" forces plain text.
//   - Default: JSON output.
//
// The response will include "X-Verbose" and "X-MapMode" headers indicating the effective mode.
// It also sets "Connection: Close".
func (o *sts) MiddleWare(c *ginsdk.Context) {
	var (
		err liberr.Error
		enc Encode
		shr = o.isParamInvBool(c.Request.URL.Query().Get(QueryVerbose), c.Request.Header.Get(HeadVerbose))
		txt = o.isText(c.Request.URL.Query().Get(QueryFormat), strings.Join(c.Request.Header.Values(HeadFormat), ","))
		mpm = o.isParamBool(c.Request.URL.Query().Get(QueryMapMode), c.Request.Header.Get(HeadMapMode))
	)

	if enc, err = o.getMarshal(mpm); err != nil {
		ret := o.getErrorReturn()
		err.Return(ret)
		ret.GinTonicErrorAbort(c, 0) // 0 = internal server error
		return
	}

	if shr { // if short if true0, so Verbose if false
		c.Header(HeadVerbose, "False")
	} else {
		c.Header(HeadVerbose, "True")
	}

	if mpm { // if short if true0, so Verbose if false
		c.Header(HeadMapMode, "True")
	} else {
		c.Header(HeadMapMode, "False")
	}

	c.Header("Connection", "Close")
	enc.GinRender(c, txt, shr)
}

// isText determines if the response should be in plain text format.
// It checks both query parameters and Accept headers.
//
// Priority:
//  1. Query parameter "format=text" forces text output
//  2. Accept header "application/json" forces JSON output
//  3. Accept header "text/plain" forces text output
//
// Parameters:
//   - query: value of the "format" query parameter
//   - header: value of the "Accept" header (may contain multiple MIME types)
//
// Returns true if text format should be used, false for JSON.
func (o *sts) isText(query, header string) bool {
	var txt = false

	if len(query) > 0 {
		if strings.EqualFold(query, "text") {
			txt = true
		}
	}

	if len(header) > 0 {
		val := strings.Split(header, ",")

		for _, m := range val {
			m = strings.TrimSpace(m)
			if m == ginsdk.MIMEJSON {
				txt = false
				break
			} else if m == ginsdk.MIMEPlain {
				txt = true
				break
			}
		}
	}

	return txt
}

// isParamBool checks if a boolean parameter is enabled based on header and query values.
// It returns true if either the header or the query parameter evaluates to true.
//
// Priority:
//  1. If header is present and "true", returns true.
//  2. If query is present and "true", returns true.
//  3. Otherwise, returns false.
//
// Parameters:
//   - query: value of the query parameter.
//   - header: value of the header.
func (o *sts) isParamBool(query, header string) bool {
	if len(header) > 0 {
		if b, e := strconv.ParseBool(header); e == nil && b {
			return true
		}
	}

	if len(query) > 0 {
		if b, e := strconv.ParseBool(query); e == nil && b {
			return true
		}
	}

	return false
}

// isParamInvBool checks if a boolean parameter is enabled with inverted logic for the header.
// It is used when the header implies the opposite of the query parameter (e.g. Verbose vs Short).
//
// Logic:
//  1. If header is present and "true", returns false (inverted).
//  2. If query is present and "true", returns true.
//  3. Otherwise, returns false.
//
// Parameters:
//   - query: value of the query parameter.
//   - header: value of the header.
func (o *sts) isParamInvBool(query, header string) bool {
	if len(header) > 0 {
		if b, e := strconv.ParseBool(header); e == nil {
			return !b
		}
	}

	if len(query) > 0 {
		if b, e := strconv.ParseBool(query); e == nil && b {
			return true
		}
	}

	return false
}

// getErrorReturn retrieves the configured error return formatter.
// If no custom formatter is set via SetErrorReturn, returns a default formatter.
// This method is thread-safe.
//
// Returns a ReturnGin instance for formatting error responses.
// See github.com/nabbar/golib/errors for ReturnGin interface details.
func (o *sts) getErrorReturn() liberr.ReturnGin {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.r == nil {
		return liberr.NewDefaultReturn()
	} else if r := o.r(); r == nil {
		return liberr.NewDefaultReturn()
	} else {
		return r
	}
}

// SetErrorReturn registers a custom error return model factory.
// The provided function should return a new ReturnGin instance for each call.
// This allows customizing how errors are formatted in HTTP responses.
//
// Parameters:
//   - f: factory function that creates a new ReturnGin instance
//
// If not set, a default return model from github.com/nabbar/golib/errors is used.
// This method is thread-safe.
func (o *sts) SetErrorReturn(f func() liberr.ReturnGin) {
	o.m.Lock()
	defer o.m.Unlock()

	o.r = f
}
