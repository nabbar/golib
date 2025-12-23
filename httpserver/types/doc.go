/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

// Package types provides core type definitions and constants for HTTP server implementations.
//
// # Overview
//
// This package defines foundational types, constants, and utilities used across the
// httpserver ecosystem. It serves as a shared vocabulary for server configuration,
// handler registration, and field identification in server management operations.
//
// The package provides:
//   - Field type enumeration for server property filtering
//   - Handler registration function signatures
//   - Fallback error handlers for misconfigured servers
//   - Timeout constants for server lifecycle management
//
// # Design Philosophy
//
// The types package follows these design principles:
//
// 1. Minimal Dependencies: The package depends only on the standard library,
// making it suitable as a foundation for higher-level packages without circular
// dependencies.
//
// 2. Type Safety: Custom types like FieldType provide compile-time safety for
// server filtering operations, preventing invalid field specifications.
//
// 3. Fail-Safe Defaults: The BadHandler provides a safe fallback that always
// returns HTTP 500, ensuring misconfigured servers fail visibly rather than
// silently.
//
// 4. Constants Over Magic Values: Named constants like TimeoutWaitingStop and
// HandlerDefault improve code readability and maintainability.
//
// # Key Features
//
// Field Type System: FieldType enumeration enables type-safe server filtering
// by name, bind address, or expose URL. This is primarily used by the pool
// package for server management.
//
// Handler Registration: FuncHandler defines the contract for dynamic handler
// registration, allowing multiple named handlers per server instance.
//
// Fallback Handling: BadHandler provides a safe default when no valid handler
// is configured, preventing nil pointer panics and providing clear error signals.
//
// Timeout Management: Pre-defined timeout constants standardize server lifecycle
// operations like graceful shutdown and port availability checks.
//
// # Architecture
//
// The package structure is intentionally flat, providing building blocks for
// higher-level abstractions:
//
//	┌────────────────────────────────────────────────────┐
//	│                  httpserver/types                  │
//	├────────────────────────────────────────────────────┤
//	│                                                    │
//	│  ┌──────────────────┐        ┌──────────────────┐  │
//	│  │   Field Types    │        │   Handler Types  │  │
//	│  ├──────────────────┤        ├──────────────────┤  │
//	│  │ FieldType        │        │ FuncHandler      │  │
//	│  │  - FieldName     │        │ BadHandler       │  │
//	│  │  - FieldBind     │        │ NewBadHandler()  │  │
//	│  │  - FieldExpose   │        │                  │  │
//	│  └──────────────────┘        └──────────────────┘  │
//	│                                                    │
//	│  ┌──────────────────┐        ┌──────────────────┐  │
//	│  │    Constants     │        │   Handler Keys   │  │
//	│  ├──────────────────┤        ├──────────────────┤  │
//	│  │ TimeoutWaiting   │        │ HandlerDefault   │  │
//	│  │  PortFreeing     │        │ BadHandlerName   │  │
//	│  │ TimeoutWaiting   │        │                  │  │
//	│  │  Stop            │        │                  │  │
//	│  └──────────────────┘        └──────────────────┘  │
//	│                                                    │
//	└────────────────────────────────────────────────────┘
//	                          │
//	                          ▼
//	         Used by httpserver, httpserver/pool,
//	         and other HTTP server components
//
// # Usage Patterns
//
// ## Field Type Filtering
//
// FieldType enables type-safe filtering of servers by specific attributes:
//
//	switch filterField {
//	case types.FieldName:
//	    // Filter by server name
//	case types.FieldBind:
//	    // Filter by bind address
//	case types.FieldExpose:
//	    // Filter by expose URL
//	}
//
// ## Handler Registration
//
// FuncHandler defines how handlers are registered with servers:
//
//	handlerFunc := func() map[string]http.Handler {
//	    return map[string]http.Handler{
//	        types.HandlerDefault: myDefaultHandler,
//	        "api":                myAPIHandler,
//	        "admin":              myAdminHandler,
//	    }
//	}
//
// ## Fallback Handler
//
// BadHandler provides safe error handling for misconfigured servers:
//
//	handler := types.NewBadHandler()
//	// Returns HTTP 500 for all requests
//
// ## Timeout Constants
//
// Standard timeouts for server lifecycle operations:
//
//	ctx, cancel := context.WithTimeout(ctx, types.TimeoutWaitingStop)
//	defer cancel()
//	server.Shutdown(ctx)
//
// # Field Types
//
// FieldType is an enumeration for identifying server properties in filtering
// and listing operations:
//
//	FieldName:   Server name identifier
//	FieldBind:   Server listen address (e.g., ":8080", "127.0.0.1:9000")
//	FieldExpose: Server public expose URL (e.g., "https://example.com")
//
// These types are primarily used by the pool package to filter and retrieve
// servers based on specific criteria.
//
// # Handler Types
//
// ## FuncHandler
//
// FuncHandler defines the signature for handler registration functions. It
// returns a map of handler identifiers to http.Handler instances:
//
//	Key patterns:
//	  - "" or "default": Default handler for unmatched routes
//	  - "api": API-specific handler
//	  - "admin": Administrative interface handler
//	  - Custom keys: Application-specific handler identifiers
//
// ## BadHandler
//
// BadHandler is a fallback http.Handler that returns HTTP 500 Internal Server
// Error for all requests. It's used when:
//   - No handler is registered for a server
//   - Handler registration fails
//   - Configuration errors prevent proper handler setup
//
// The handler serves as a visible failure indicator rather than panicking or
// silently ignoring requests.
//
// # Constants
//
// ## Timeout Constants
//
// TimeoutWaitingPortFreeing (250µs):
//   - Duration for polling port availability before binding
//   - Used in port conflict detection and retry logic
//   - Short duration suitable for tight polling loops
//
// TimeoutWaitingStop (5s):
//   - Default timeout for graceful server shutdown
//   - Allows ongoing requests to complete before forced termination
//   - Balances graceful shutdown with reasonable wait times
//
// ## Handler Identifier Constants
//
// HandlerDefault ("default"):
//   - Standard key for default handler registration
//   - Used when no specific handler key is configured
//   - Fallback handler for unmatched routes
//
// BadHandlerName ("no handler"):
//   - Identifier string for BadHandler instances
//   - Used in logging and monitoring
//   - Indicates misconfigured or missing handler
//
// # Integration with httpserver Ecosystem
//
// This package integrates with:
//
//	httpserver:      Core server implementation uses these types
//	httpserver/pool: Server pooling uses FieldType for filtering
//	httpserver/config: Configuration uses timeout constants
//
// # Performance Considerations
//
// Type Overhead:
//   - FieldType is a uint8 (1 byte), minimal memory overhead
//   - Handler maps are created on-demand, not stored statically
//   - BadHandler is stateless, safe to create multiple instances
//
// Constant Access:
//   - All constants are compile-time values
//   - Zero runtime overhead for constant access
//   - No initialization required
//
// Handler Performance:
//   - BadHandler.ServeHTTP is a trivial function (single line)
//   - No allocations, no blocking operations
//   - Suitable for high-frequency error scenarios
//
// # Limitations
//
// 1. No Dynamic Field Types: FieldType is a closed enumeration. Adding new
// field types requires modifying this package.
//
// 2. No Handler Lifecycle Management: BadHandler does not implement graceful
// shutdown or resource cleanup. It's stateless by design.
//
// 3. Fixed Timeout Values: Timeout constants are not configurable at runtime.
// Applications needing custom timeouts should define their own.
//
// 4. No Validation: The package does not validate handler maps returned by
// FuncHandler. Validation is the responsibility of consuming packages.
//
// 5. Single Error Status: BadHandler always returns 500. It does not support
// custom error codes or messages.
//
// # Use Cases
//
// ## Server Pool Management
//
// Filter servers in a pool by specific attributes:
//
//	// Find all servers listening on a specific address
//	servers := pool.FilterByField(types.FieldBind, ":8080")
//
// ## Multi-Handler Server Configuration
//
// Register multiple handlers for different routes or purposes:
//
//	cfg.HandlerFunc = func() map[string]http.Handler {
//	    return map[string]http.Handler{
//	        types.HandlerDefault: webHandler,
//	        "api":                apiHandler,
//	        "metrics":            metricsHandler,
//	    }
//	}
//
// ## Graceful Shutdown
//
// Use standard timeout for server shutdown:
//
//	shutdownCtx, cancel := context.WithTimeout(
//	    context.Background(),
//	    types.TimeoutWaitingStop,
//	)
//	defer cancel()
//	server.Shutdown(shutdownCtx)
//
// ## Safe Default Handler
//
// Provide fallback when handler registration fails:
//
//	handler := getConfiguredHandler()
//	if handler == nil {
//	    handler = types.NewBadHandler()
//	}
//
// # Thread Safety
//
// All types in this package are safe for concurrent use:
//
//   - FieldType is a simple enumeration (immutable)
//   - Constants are immutable by definition
//   - BadHandler is stateless and safe for concurrent ServeHTTP calls
//   - FuncHandler signature does not enforce thread safety; implementations
//     must ensure their own thread safety if called concurrently
//
// # Error Handling
//
// The package provides minimal error handling as it defines types rather than
// implementing complex logic:
//
//   - BadHandler signals errors via HTTP 500 status code
//   - No errors are returned by package functions
//   - Invalid FieldType values are not validated at runtime
//
// Consumer packages are responsible for validating inputs and handling errors
// appropriately.
//
// # Best Practices
//
// DO:
//   - Use FieldType constants for server filtering to avoid typos
//   - Use HandlerDefault for primary handler registration
//   - Use NewBadHandler() as a safe fallback for missing handlers
//   - Use timeout constants for consistent server lifecycle management
//   - Document custom handler keys in application code
//
// DON'T:
//   - Don't cast arbitrary integers to FieldType
//   - Don't rely on BadHandler for production traffic
//   - Don't modify timeout constants (they're package-level)
//   - Don't assume FuncHandler implementations are thread-safe
//   - Don't use BadHandler for intentional error responses
//
// # Testing
//
// The package includes comprehensive testing:
//   - Field type enumeration and uniqueness
//   - Constant value verification
//   - BadHandler HTTP response validation
//   - FuncHandler signature compliance
//   - Integration with http.Handler interface
//
// See fields_test.go and handler_test.go for detailed test specifications.
//
// # Related Packages
//
//   - net/http: Standard library HTTP server interfaces
//   - github.com/nabbar/golib/httpserver: HTTP server implementation
//   - github.com/nabbar/golib/httpserver/pool: Server pool management
//
// For usage examples, see example_test.go in this package.
package types
