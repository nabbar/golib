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
	// HeadVerbose is the HTTP header for controlling response verbosity.
	// A value of "false" enables short mode (equivalent to `short=true` query param).
	HeadVerbose = "X-Verbose"

	// HeadFormat is the standard "Accept" HTTP header used for content negotiation.
	// It supports "application/json" and "text/plain".
	HeadFormat = "Accept"

	// HeadMapMode is the HTTP header for enabling map mode for component output.
	// A value of "true" enables map mode.
	HeadMapMode = "X-MapMode"

	// QueryVerbose is the query parameter for enabling short output mode.
	// A value of "true" or "1" will omit component details from the response.
	QueryVerbose = "short"

	// QueryFormat is the query parameter for selecting the output format.
	// Supported values are "text" or "json".
	QueryFormat = "format"

	// QueryMapMode is the query parameter for enabling map mode for component output.
	// A value of "true" or "1" will format the component list as a map.
	QueryMapMode = "map"
)

// Expose handles a status request from a generic `context.Context`.
// If the context is a Gin context (`*ginsdk.Context`), it delegates to `MiddleWare`.
// This method provides a generic entry point for frameworks that do not directly
// expose the underlying `*gin.Context`.
func (o *sts) Expose(ctx context.Context) {
	if c, ok := ctx.(*ginsdk.Context); ok {
		o.MiddleWare(c)
	}
}

// MiddleWare is the Gin middleware handler that processes status requests.
// It determines the response format and content based on a combination of HTTP
// headers and query parameters.
//
// The handler orchestrates the following:
//  1. Parses request parameters to determine verbosity, format, and map mode.
//  2. Retrieves the appropriate encoder.
//  3. Renders the response using the selected format and verbosity.
//
// It sets the "Connection: Close" header to ensure the connection is not kept alive.
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
		ret.GinTonicErrorAbort(c, 0) // 0 defaults to internal server error
		return
	}

	if shr {
		c.Header(HeadVerbose, "False")
	} else {
		c.Header(HeadVerbose, "True")
	}

	if mpm {
		c.Header(HeadMapMode, "True")
	} else {
		c.Header(HeadMapMode, "False")
	}

	c.Header("Connection", "Close")
	enc.GinRender(c, txt, shr)
}

// isText determines if the response should be in plain text format by checking
// query parameters and `Accept` headers.
//
// The precedence is as follows:
//  1. `format=text` query parameter forces text output.
//  2. `Accept: application/json` header forces JSON output.
//  3. `Accept: text/plain` header requests text output.
//
// Returns `true` if text format should be used, otherwise `false` for JSON.
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

// isParamBool checks if a boolean parameter is enabled based on its query and
// header values. It returns `true` if either the query or header value is "true".
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

// isParamInvBool checks if a boolean parameter is enabled, but with inverted logic
// for the header value. This is used for cases like `X-Verbose` vs. `short`, where
// `X-Verbose: false` is equivalent to `short=true`.
//
// The logic is:
//  1. If the header is "true", returns `false`.
//  2. If the query is "true", returns `true`.
//  3. Otherwise, returns `false`.
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
// If no custom formatter is registered via `SetErrorReturn`, it returns a default
// formatter from the `golib/errors` package.
// This method is thread-safe.
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

// SetErrorReturn registers a factory function for creating custom error formatters.
// The provided function should return a new `liberr.ReturnGin` instance on each call.
// This allows for customizing how errors are formatted in HTTP responses.
// This method is thread-safe.
func (o *sts) SetErrorReturn(f func() liberr.ReturnGin) {
	o.m.Lock()
	defer o.m.Unlock()

	o.r = f
}
