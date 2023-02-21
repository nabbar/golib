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

package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	librtr "github.com/nabbar/golib/router"
	sdkpht "github.com/prometheus/client_golang/prometheus/promhttp"
)

func (m *prom) getHandler() http.Handler {
	m.m.RLock()
	defer m.m.RUnlock()

	if m.handle != nil {
		return m.handle
	}

	return nil
}

func (m *prom) setHandler() {
	m.m.Lock()
	defer m.m.Unlock()

	m.handle = sdkpht.Handler()
}

// Expose adds metric path to a given router.
// The router can be different with the one passed to UseWithoutExposingEndpoint.
// This allows to expose ginMet on different port.
func (m *prom) Expose(ctx context.Context) {
	if c, ok := ctx.(*ginsdk.Context); ok {
		m.ExposeGin(c)
	}
}

// ExposeGin is like Expose but dedicated to gin context struct.
func (m *prom) ExposeGin(c *ginsdk.Context) {
	if h := m.getHandler(); h != nil {
		h.ServeHTTP(c.Writer, c.Request)
		return
	}

	m.setHandler()
	if h := m.getHandler(); h != nil {
		h.ServeHTTP(c.Writer, c.Request)
		return
	}

	_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cannot initiate prometheus handler"))
	return
}

// GinHandlerFunc adds metric path to a given router.
// The router can be different with the one passed to UseWithoutExposingEndpoint.
// This allows to expose ginMet on different port.
func (m *prom) MiddleWareGin(c *ginsdk.Context) {
	if c.GetInt64(librtr.GinContextStartUnixNanoTime) == 0 {
		c.Set(librtr.GinContextStartUnixNanoTime, time.Now().UnixNano())
	}

	path := c.GetString(librtr.GinContextRequestPath)
	if path == "" {
		path = c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; len(raw) > 0 {
			path += "?" + raw
		}
	}

	if m.isExclude(c.Request.URL.Path) {
		return
	}

	// execute normal process.
	c.Next()

	// after request
	m.Collect(c)
}

// MiddleWare as gin monitor middleware.
func (m *prom) MiddleWare(ctx context.Context) {
	if c, ok := ctx.(*ginsdk.Context); !ok {
		return
	} else {
		m.MiddleWareGin(c)
	}
}
