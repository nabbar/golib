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
 *
 */

package gin

import (
	"context"
	"os"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	liblog "github.com/nabbar/golib/logger"
)

type GinTonic interface {
	context.Context

	//generic
	GinContext() *ginsdk.Context
	CancelOnSignal(s ...os.Signal)

	//gin context metadata
	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	MustGet(key string) interface{}
	GetString(key string) (s string)
	GetBool(key string) (b bool)
	GetInt(key string) (i int)
	GetInt64(key string) (i64 int64)
	GetFloat64(key string) (f64 float64)
	GetTime(key string) (t time.Time)
	GetDuration(key string) (d time.Duration)
	GetStringSlice(key string) (ss []string)
	GetStringMap(key string) (sm map[string]interface{})
	GetStringMapString(key string) (sms map[string]string)
	GetStringMapStringSlice(key string) (smss map[string][]string)

	SetLogger(log liblog.FuncLog)
}

func New(c *ginsdk.Context, log liblog.FuncLog) GinTonic {
	if c == nil {
		c = &ginsdk.Context{
			Request:  nil,
			Writer:   nil,
			Params:   make(ginsdk.Params, 0),
			Keys:     make(map[string]interface{}),
			Errors:   make([]*ginsdk.Error, 0),
			Accepted: make([]string, 0),
		}
	}

	var (
		x context.Context
		l context.CancelFunc
	)

	if c.Request != nil && c.Request.Context() != nil {
		x, l = context.WithCancel(c.Request.Context())
	} else {
		x, l = context.WithCancel(c)
	}

	return &ctxGinTonic{
		l: log,
		g: c,
		x: x,
		c: l,
	}
}
