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
	"os/signal"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
)

type ctxGinTonic struct {
	l liblog.FuncLog
	g *ginsdk.Context
	x context.Context
	c context.CancelFunc
}

func (c *ctxGinTonic) SetLogger(fct liblog.FuncLog) {
	c.l = fct
}

func (c *ctxGinTonic) log(lvl loglvl.Level, msg string, args ...any) {
	if c.l != nil {
		c.l().Entry(lvl, msg, args...).Log()
	} else {
		l := liblog.New(func() context.Context {
			return c
		})

		l.Entry(lvl, msg, args...).Log()
	}
}

func (c *ctxGinTonic) CancelOnSignal(s ...os.Signal) {
	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, s...)

		select {
		case <-sc:
			c.log(loglvl.InfoLevel, "OS Signal received, calling context cancel !")
			c.c()
			return
		case <-c.Done():
			c.log(loglvl.InfoLevel, "Context has been closed !")
			return
		}
	}()
}

func (c *ctxGinTonic) Deadline() (deadline time.Time, ok bool) {
	return c.x.Deadline()
}

func (c *ctxGinTonic) Done() <-chan struct{} {
	return c.x.Done()
}

func (c *ctxGinTonic) Err() error {
	return c.x.Err()
}

func (c *ctxGinTonic) Value(key any) any {
	return c.g.Value(key)
}

func (c *ctxGinTonic) GinContext() *ginsdk.Context {
	return c.g
}

func (c *ctxGinTonic) Set(key any, value any) {
	c.g.Set(key, value)
}

func (c *ctxGinTonic) Get(key any) (value any, exists bool) {
	return c.g.Get(key)
}

func (c *ctxGinTonic) MustGet(key any) any {
	return c.g.MustGet(key)
}

func (c *ctxGinTonic) GetString(key any) (s string) {
	return c.g.GetString(key)
}

func (c *ctxGinTonic) GetBool(key any) (b bool) {
	return c.g.GetBool(key)
}

func (c *ctxGinTonic) GetInt(key any) (i int) {
	return c.g.GetInt(key)
}

func (c *ctxGinTonic) GetInt64(key any) (i64 int64) {
	return c.g.GetInt64(key)
}

func (c *ctxGinTonic) GetFloat64(key any) (f64 float64) {
	return c.g.GetFloat64(key)
}

func (c *ctxGinTonic) GetTime(key any) (t time.Time) {
	return c.g.GetTime(key)
}

func (c *ctxGinTonic) GetDuration(key any) (d time.Duration) {
	return c.g.GetDuration(key)
}

func (c *ctxGinTonic) GetStringSlice(key any) (ss []string) {
	return c.g.GetStringSlice(key)
}

func (c *ctxGinTonic) GetStringMap(key any) (sm map[string]any) {
	return c.g.GetStringMap(key)
}

func (c *ctxGinTonic) GetStringMapString(key any) (sms map[string]string) {
	return c.g.GetStringMapString(key)
}

func (c *ctxGinTonic) GetStringMapStringSlice(key any) (smss map[string][]string) {
	return c.g.GetStringMapStringSlice(key)
}
