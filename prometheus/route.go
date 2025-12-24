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
	"time"

	ginsdk "github.com/gin-gonic/gin"
	librtr "github.com/nabbar/golib/router"
)

// Expose handles the /metrics endpoint for Prometheus scraping.
// It serves the metrics in Prometheus text format.
//
// This method accepts a generic context and delegates to ExposeGin if
// the context is a Gin context. This allows for flexible integration
// with different routing frameworks.
//
// The metrics endpoint can be exposed on a different port or router than
// the main application for security or organizational reasons.
//
// Example:
//
//	router.GET("/metrics", func(c *gin.Context) {
//	    prm.Expose(c)
//	})
func (m *prom) Expose(ctx context.Context) {
	if c, ok := ctx.(*ginsdk.Context); ok {
		m.ExposeGin(c)
	}
}

// ExposeGin is the Gin-specific metrics endpoint handler.
// It serves metrics in Prometheus text format for scraping by Prometheus server.
//
// This handler uses the standard prometheus/promhttp handler to generate
// the metrics output in the format expected by Prometheus.
//
// Example:
//
//	// Simple setup
//	router.GET("/metrics", prm.ExposeGin)
//
//	// With authentication
//	metrics := router.Group("/")
//	metrics.Use(authMiddleware)
//	metrics.GET("/metrics", prm.ExposeGin)
func (m *prom) ExposeGin(c *ginsdk.Context) {
	m.hdl.ServeHTTP(c.Writer, c.Request)
}

// MiddleWareGin is the Gin-specific middleware for Prometheus metric collection.
//
// The middleware performs the following operations:
//  1. Captures request start time (if not already set by another middleware)
//  2. Captures request path including query parameters
//  3. Calls c.Next() to execute the remaining handlers in the chain
//  4. After the request completes, triggers metric collection (unless path is excluded)
//
// The start time and request path are stored in the Gin context using constants
// from the router package (GinContextStartUnixNanoTime and GinContextRequestPath).
//
// Thread-safe: Can be used concurrently by multiple goroutines handling different requests.
//
// Example:
//
//	router := gin.Default()
//	router.Use(gin.HandlerFunc(prm.MiddleWareGin))
//
//	// Or with method call
//	router.Use(prm.MiddleWareGin)
func (m *prom) MiddleWareGin(c *ginsdk.Context) {
	if c.GetInt64(librtr.GinContextStartUnixNanoTime) == 0 {
		c.Set(librtr.GinContextStartUnixNanoTime, time.Now().UnixNano())
	}

	path := c.GetString(librtr.GinContextRequestPath)
	if len(path) < 1 {
		path = c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; len(raw) > 0 {
			path += "?" + raw
		}
		c.Set(librtr.GinContextRequestPath, path)
	}

	// execute normal process.
	c.Next()

	if m.isExclude(c.Request.URL.Path) {
		return
	}

	// after request
	m.Collect(c)
}

// MiddleWare is a generic middleware handler for metric collection.
// It accepts a context and delegates to MiddleWareGin if the context is a Gin context.
//
// This method provides a context-agnostic interface for middleware integration,
// allowing the same code to work with different context types.
//
// Example:
//
//	router.Use(func(c *gin.Context) {
//	    prm.MiddleWare(c)
//	})
func (m *prom) MiddleWare(ctx context.Context) {
	if c, ok := ctx.(*ginsdk.Context); !ok {
		return
	} else {
		m.MiddleWareGin(c)
	}
}
