/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type GinTonic interface {
	context.Context

	//generic
	GinContext() *gin.Context

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
}

type ctxGinTonic struct {
	gin.Context
}

func NewGinTonic(c *gin.Context) GinTonic {
	return &ctxGinTonic{
		*c.Copy(),
	}
}

func (c *ctxGinTonic) Deadline() (deadline time.Time, ok bool) {
	return c.Request.Context().Deadline()
}

func (c *ctxGinTonic) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

func (c *ctxGinTonic) Err() error {
	return c.Request.Context().Err()
}

func (c *ctxGinTonic) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

func (c *ctxGinTonic) GinContext() *gin.Context {
	return &c.Context
}
