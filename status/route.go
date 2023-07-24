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
	headVerbose = "X-Verbose"
	headFormat  = "Accept"
	keyVerbose  = "short"
	keyFormat   = "format"
)

// Expose adds metric path to a given router.
// The router can be different with the one passed to UseWithoutExposingEndpoint.
// This allows to expose ginMet on different port.
func (o *sts) Expose(ctx context.Context) {
	if c, ok := ctx.(*ginsdk.Context); ok {
		o.MiddleWare(c)
	}
}

// MiddleWare as gin monitor middleware HandleFunc.
func (o *sts) MiddleWare(c *ginsdk.Context) {
	var (
		err liberr.Error
		enc Encode
		shr = o.isShort(c.Request.URL.Query().Get(keyVerbose), c.Request.Header.Get(headVerbose))
		txt = o.isText(c.Request.URL.Query().Get(keyFormat), strings.Join(c.Request.Header.Values(headFormat), ","))
	)

	if enc, err = o.getMarshal(); err != nil {
		ret := o.getErrorReturn()
		err.Return(ret)
		ret.GinTonicErrorAbort(c, 0) // 0 = internal server error
		return
	}

	if shr {
		c.Header("X-Verbose", "False")
	} else {
		c.Header("X-Verbose", "True")
	}

	c.Header("Connection", "Close")
	enc.GinRender(c, txt, shr)
}

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

func (o *sts) isShort(query, header string) bool {
	var shr = false

	if len(query) > 0 {
		if strings.EqualFold(query, "true") {
			shr = true
		} else if strings.EqualFold(query, "1") {
			shr = true
		}
	}

	if len(header) > 0 {
		if b, e := strconv.ParseBool(header); e == nil {
			shr = !b
		}
	}

	return shr
}

// MiddleWare as gin monitor middleware HandleFunc.
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

// SetErrorReturn allow to register a return model use to export error to output.
func (o *sts) SetErrorReturn(f func() liberr.ReturnGin) {
	o.m.Lock()
	defer o.m.Unlock()

	o.r = f
}
