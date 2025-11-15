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

// Package router provides HTTP routing functionality built on top of the Gin web framework.
// It offers a flexible router list system, middleware support, authentication helpers,
// and header management utilities.
//
// Key features:
//   - RouterList: Organize routes with optional grouping
//   - Middleware: Latency tracking, request context, access logging, error recovery
//   - Authentication: Authorization handlers with customizable check functions
//   - Headers: Manage HTTP headers across routes
//
// Example usage:
//
//	routerList := router.NewRouterList(router.DefaultGinInit)
//	routerList.Register(http.MethodGet, "/api/health", healthHandler)
//	routerList.RegisterInGroup("/api/v1", http.MethodGet, "/users", usersHandler)
//	engine := routerList.Engine()
//	routerList.Handler(engine)
//	engine.Run(":8080")
//
// See also:
//   - github.com/gin-gonic/gin for the underlying web framework
//   - github.com/nabbar/golib/logger for logging integration
//   - github.com/nabbar/golib/errors for error handling
package router

import (
	"os"

	ginsdk "github.com/gin-gonic/gin"
)

const (
	// EmptyHandlerGroup is the identifier used for routes registered without a group.
	// Routes with this group are registered directly on the engine root.
	EmptyHandlerGroup = "<nil>"

	// GinContextStartUnixNanoTime is the context key for storing request start time in nanoseconds.
	// Used by GinLatencyContext middleware to calculate request duration.
	GinContextStartUnixNanoTime = "gin-ctx-start-unix-nano-time"

	// GinContextRequestPath is the context key for storing the sanitized request path.
	// Includes query parameters if present. Set by GinRequestContext middleware.
	GinContextRequestPath = "gin-ctx-request-path"

	// GinContextRequestUser is the context key for storing the authenticated user from URL.
	// Set by GinRequestContext middleware when user info is present in the request URL.
	GinContextRequestUser = "gin-ctx-request-user"
)

var (
	defaultRouters = NewRouterList(DefaultGinInit)
)

func init() {
	if os.Getenv("GIN_MODE") == "" {
		ginsdk.SetMode(ginsdk.ReleaseMode)
	}
}

// RegisterRouter is a function type for registering routes without a group.
// It takes an HTTP method, relative path, and one or more handler functions.
type RegisterRouter func(method string, relativePath string, router ...ginsdk.HandlerFunc)

// RegisterRouterInGroup is a function type for registering routes within a group.
// It takes a group path, HTTP method, relative path, and one or more handler functions.
type RegisterRouterInGroup func(group, method string, relativePath string, router ...ginsdk.HandlerFunc)

// RouterList manages a collection of HTTP routes with optional grouping.
// It provides methods to register routes, organize them into groups, and apply them to a Gin engine.
//
// All methods are safe for concurrent use.
type RouterList interface {
	// Register adds a route without a group (registered at root level).
	// Multiple handlers can be provided and will be executed in order.
	Register(method string, relativePath string, router ...ginsdk.HandlerFunc)

	// RegisterInGroup adds a route within a specified group.
	// The group path is prefixed to the relative path.
	// Multiple routes can be registered in the same group.
	RegisterInGroup(group, method string, relativePath string, router ...ginsdk.HandlerFunc)

	// RegisterMergeInGroup adds or replaces a route in a group.
	// If a route with the same method and path already exists in the group,
	// its handlers are replaced with the new ones.
	RegisterMergeInGroup(group, method string, relativePath string, router ...ginsdk.HandlerFunc)

	// Handler applies all registered routes to the given Gin engine.
	// Routes are organized by group and registered accordingly.
	Handler(engine *ginsdk.Engine)

	// Engine returns a new Gin engine instance using the configured init function.
	// If no init function was provided, DefaultGinInit is used.
	Engine() *ginsdk.Engine
}

// NewRouterList creates a new RouterList instance with the specified Gin engine initializer.
// If initGin is nil, DefaultGinInit will be used when Engine() is called.
//
// Example:
//
//	routerList := NewRouterList(func() *gin.Engine {
//	    engine := gin.New()
//	    engine.Use(gin.Logger(), gin.Recovery())
//	    return engine
//	})
func NewRouterList(initGin func() *ginsdk.Engine) RouterList {
	return &rtr{
		init: initGin,
		list: make(map[string][]itm),
	}
}
