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

package router

import (
	"net/http"

	ginsdk "github.com/gin-gonic/gin"
)

// DefaultGinInit creates a new Gin engine with default middleware.
// The engine includes Logger and Recovery middleware from Gin.
//
// This is the default initializer used by NewRouterList when no custom
// initializer is provided.
//
// Returns a configured Gin engine ready to use.
func DefaultGinInit() *ginsdk.Engine {
	engine := ginsdk.New()
	engine.Use(ginsdk.Logger(), ginsdk.Recovery())

	return engine
}

// DefaultGinWithTrustyProxy creates a new Gin engine with trusted proxy configuration.
// The engine includes Logger and Recovery middleware.
//
// Trusted proxies are IP addresses or CIDR ranges that are allowed to set
// X-Forwarded-For, X-Real-IP, and other forwarding headers.
//
// Parameters:
//   - trustyProxy: List of trusted proxy IP addresses or CIDR ranges
//
// Returns a configured Gin engine with trusted proxies set.
func DefaultGinWithTrustyProxy(trustyProxy []string) *ginsdk.Engine {
	engine := ginsdk.New()
	engine.Use(ginsdk.Logger(), ginsdk.Recovery())

	if len(trustyProxy) > 0 {
		_ = engine.SetTrustedProxies(trustyProxy)
	}

	return engine
}

// DefaultGinWithTrustedPlatform creates a new Gin engine with trusted platform header.
// The engine includes Logger and Recovery middleware.
//
// The trusted platform header is used to determine the client's real IP address
// when behind a CDN or load balancer (e.g., "X-CDN-IP", "X-Real-IP").
//
// Parameters:
//   - trustedPlatform: Name of the header to trust for client IP
//
// Returns a configured Gin engine with trusted platform set.
func DefaultGinWithTrustedPlatform(trustedPlatform string) *ginsdk.Engine {
	engine := ginsdk.New()
	engine.Use(ginsdk.Logger(), ginsdk.Recovery())

	if len(trustedPlatform) > 0 {
		engine.TrustedPlatform = trustedPlatform
	}

	return engine
}

// RoutersRegister registers a route on the global default router list.
// The route is registered at root level (without a group).
//
// This is a convenience function for applications that use a single global router.
// For more complex applications, consider creating dedicated RouterList instances.
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, DELETE, etc.)
//   - relativePath: URL path for the route
//   - router: One or more handler functions
func RoutersRegister(method string, relativePath string, router ...ginsdk.HandlerFunc) {
	defaultRouters.Register(method, relativePath, router...)
}

// RoutersRegisterInGroup registers a route in a group on the global default router list.
//
// This is a convenience function for applications that use a single global router.
// For more complex applications, consider creating dedicated RouterList instances.
//
// Parameters:
//   - group: Group path prefix
//   - method: HTTP method (GET, POST, PUT, DELETE, etc.)
//   - relativePath: URL path for the route (will be prefixed with group)
//   - router: One or more handler functions
func RoutersRegisterInGroup(group, method string, relativePath string, router ...ginsdk.HandlerFunc) {
	defaultRouters.RegisterInGroup(group, method, relativePath, router...)
}

// RoutersHandler applies all routes from the global default router list to the engine.
//
// This is a convenience function for applications that use a single global router.
// For more complex applications, consider creating dedicated RouterList instances.
//
// Parameters:
//   - engine: Gin engine to register routes on
func RoutersHandler(engine *ginsdk.Engine) {
	defaultRouters.Handler(engine)
}

// GinEngine creates a new Gin engine with optional trusted platform and proxies.
// Unlike the Default* functions, this does not add any middleware by default.
//
// Parameters:
//   - trustedPlatform: Header name to trust for client IP (e.g., "X-Real-IP")
//   - trustyProxy: Optional list of trusted proxy IP addresses or CIDR ranges
//
// Returns:
//   - *gin.Engine: Configured Gin engine
//   - error: Error from SetTrustedProxies if proxy configuration fails
//
// Example:
//
//	engine, err := router.GinEngine("X-Forwarded-For", "127.0.0.1", "192.168.0.0/16")
//	if err != nil {
//	    log.Fatal(err)
//	}
func GinEngine(trustedPlatform string, trustyProxy ...string) (*ginsdk.Engine, error) {
	var err error

	engine := ginsdk.New()
	if len(trustyProxy) > 0 {
		err = engine.SetTrustedProxies(trustyProxy)
	}
	if len(trustedPlatform) > 0 {
		engine.TrustedPlatform = trustedPlatform
	}

	return engine, err
}

// GinAddGlobalMiddleware adds one or more middleware functions to the Gin engine.
// The middleware will be applied to all routes registered on the engine.
//
// Parameters:
//   - eng: Gin engine to add middleware to
//   - middleware: One or more middleware handler functions
//
// Returns the same engine for method chaining.
//
// Example:
//
//	engine := gin.New()
//	router.GinAddGlobalMiddleware(engine, router.GinLatencyContext, router.GinRequestContext)
func GinAddGlobalMiddleware(eng *ginsdk.Engine, middleware ...ginsdk.HandlerFunc) *ginsdk.Engine {
	eng.Use(middleware...)
	return eng
}

// SetGinHandler is a type conversion helper that converts a function to HandlerFunc.
// This is useful when you have a function that matches the HandlerFunc signature
// but needs explicit type conversion.
//
// Parameters:
//   - fct: Function with signature func(*gin.Context)
//
// Returns the function as a gin.HandlerFunc type.
func SetGinHandler(fct func(c *ginsdk.Context)) ginsdk.HandlerFunc {
	return fct
}

// Handler creates an http.Handler from a RouterList.
// It initializes a Gin engine and applies all routes from the RouterList.
//
// If routerList is nil, the global default router list is used instead.
//
// Parameters:
//   - routerList: RouterList containing routes to register
//
// Returns an http.Handler that can be used with http.Server.
//
// Example:
//
//	routerList := router.NewRouterList(router.DefaultGinInit)
//	routerList.Register(http.MethodGet, "/health", healthHandler)
//	handler := router.Handler(routerList)
//	http.ListenAndServe(":8080", handler)
func Handler(routerList RouterList) http.Handler {
	e := routerList.Engine()

	if routerList == nil {
		RoutersHandler(e)
	} else {
		routerList.Handler(e)
	}

	return e
}
