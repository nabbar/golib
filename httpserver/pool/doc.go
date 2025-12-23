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

// Package pool provides unified management of multiple HTTP servers through a thread-safe
// pool abstraction. It enables simultaneous operation of multiple servers with different
// configurations, unified lifecycle control, and advanced filtering capabilities.
//
// # Overview
//
// The pool package extends github.com/nabbar/golib/httpserver by providing a container
// for managing multiple HTTP server instances as a cohesive unit. All servers in the pool
// can be started, stopped, or restarted together while maintaining individual configurations
// and bind addresses.
//
// Key capabilities:
//   - Unified lifecycle management (Start/Stop/Restart all servers)
//   - Thread-safe concurrent operations with sync.RWMutex
//   - Advanced filtering by name, bind address, or expose address
//   - Server merging and configuration validation
//   - Monitoring integration for all pooled servers
//   - Dynamic server addition and removal during operation
//
// # Design Philosophy
//
// 1. Unified Management: Control multiple heterogeneous servers as a single logical unit
// 2. Thread Safety First: All operations are protected by mutex for concurrent safety
// 3. Flexibility: Support for dynamic server addition, removal, and configuration updates
// 4. Observability: Built-in monitoring and health check integration
// 5. Error Aggregation: Collect and report errors from all servers systematically
//
// # Architecture
//
// The pool implementation uses a layered architecture:
//
//	┌─────────────────────────────────────────────────────────┐
//	│                         Pool                            │
//	├─────────────────────────────────────────────────────────┤
//	│                                                         │
//	│  ┌──────────────┐        ┌──────────────────────┐       │
//	│  │   Context    │        │   Handler Function   │       │
//	│  │   Provider   │        │   (shared optional)  │       │
//	│  └──────┬───────┘        └──────────┬───────────┘       │
//	│         │                           │                   │
//	│         ▼                           ▼                   │
//	│  ┌──────────────────────────────────────────────┐       │
//	│  │     Server Map (libctx.Config[string])       │       │
//	│  │     Key: Bind Address (e.g., "0.0.0.0:8080") │       │
//	│  │     Value: libhtp.Server instance            │       │
//	│  └──────────────────────────────────────────────┘       │
//	│         │                                               │
//	│         ▼                                               │
//	│  ┌─────────────────────────────────────────────┐        │
//	│  │  Individual Server Instances                │        │
//	│  │                                             │        │
//	│  │  Server 1 ──┐  Server 2 ──┐  Server N ──┐   │        │
//	│  │  :8080      │  :8443      │  :9000      │   │        │
//	│  │  HTTP       │  HTTPS      │  Custom     │   │        │
//	│  └─────────────────────────────────────────────┘        │
//	│                                                         │
//	│  ┌─────────────────────────────────────────────┐        │
//	│  │          Pool Operations                    │        │
//	│  │  - Walk: Iterate all servers                │        │
//	│  │  - WalkLimit: Iterate specific servers      │        │
//	│  │  - Filter: Query by criteria                │        │
//	│  │  - Start/Stop/Restart: Lifecycle            │        │
//	│  │  - Monitor: Health and metrics              │        │
//	│  └─────────────────────────────────────────────┘        │
//	│                                                         │
//	└─────────────────────────────────────────────────────────┘
//
// # Data Flow
//
// Server Lifecycle:
//  1. Configuration Phase: Servers defined via libhtp.Config
//  2. Pool Creation: New() creates empty pool or with initial servers
//  3. Server Addition: StoreNew() validates config and adds server
//  4. Lifecycle Control: Start() initiates all servers concurrently
//  5. Runtime Operations: Filter, Walk, Monitor during operation
//  6. Graceful Shutdown: Stop() drains and closes all servers
//
// Error Handling:
//  1. Validation errors collected during config validation
//  2. Startup errors aggregated during Start()
//  3. Shutdown errors collected during Stop()
//  4. All errors use liberr.Error with proper code hierarchy
//
// # Thread Safety
//
// All pool operations are thread-safe through sync.RWMutex:
//   - Read operations (Load, Walk, Filter, List) use RLock
//   - Write operations (Store, Delete, Clean) use Lock
//   - Atomic server map updates via libctx.Config
//   - Safe concurrent access to individual servers
//
// Synchronization guarantees:
//   - No data races during concurrent operations
//   - Consistent view of server collection
//   - Safe iteration during modifications
//   - Proper memory barriers for visibility
//
// # Basic Usage
//
// Creating and managing a simple pool:
//
//	// Create configurations
//	cfg1 := libhtp.Config{
//	    Name:   "api-server",
//	    Listen: "0.0.0.0:8080",
//	    Expose: "http://api.example.com",
//	}
//	cfg1.RegisterHandlerFunc(apiHandler)
//
//	cfg2 := libhtp.Config{
//	    Name:   "admin-server",
//	    Listen: "127.0.0.1:9000",
//	    Expose: "http://localhost:9000",
//	}
//	cfg2.RegisterHandlerFunc(adminHandler)
//
//	// Create pool and add servers
//	p := pool.New(nil, nil)
//	p.StoreNew(cfg1, nil)
//	p.StoreNew(cfg2, nil)
//
//	// Start all servers
//	ctx := context.Background()
//	if err := p.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Graceful shutdown
//	defer func() {
//	    stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	    defer cancel()
//	    p.Stop(stopCtx)
//	}()
//
// # Advanced Features
//
// ## Configuration Helpers
//
// The Config slice type provides bulk operations:
//
//	configs := pool.Config{cfg1, cfg2, cfg3}
//
//	// Validate all configurations
//	if err := configs.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Set shared handler for all
//	configs.SetHandlerFunc(sharedHandler)
//
//	// Set shared context
//	configs.SetContext(context.Background())
//
//	// Create pool from configs
//	p, err := configs.Pool(nil, nil, nil)
//
// ## Dynamic Server Management
//
// Add, remove, and update servers during operation:
//
//	// Add new server dynamically
//	newCfg := libhtp.Config{Name: "metrics", Listen: ":2112"}
//	p.StoreNew(newCfg, nil)
//
//	// Remove server
//	p.Delete(":2112")
//
//	// Atomic load and delete
//	if srv, ok := p.LoadAndDelete(":8080"); ok {
//	    srv.Stop(ctx)
//	}
//
//	// Clear all servers
//	p.Clean()
//
// ## Filtering and Querying
//
// Filter servers by various criteria:
//
//	// Filter by name pattern
//	apiServers := p.Filter(srvtps.FieldName, "", "^api-.*")
//
//	// Filter by bind address
//	localServers := p.Filter(srvtps.FieldBind, "", "^127\\.0\\.0\\.1:.*")
//
//	// Filter by expose address
//	prodServers := p.Filter(srvtps.FieldExpose, "", ".*\\.example\\.com.*")
//
//	// List specific fields
//	names := p.List(srvtps.FieldBind, srvtps.FieldName, "", ".*")
//
//	// Chain filters
//	filtered := p.Filter(srvtps.FieldBind, "", "^0\\.0\\.0\\.0:.*").
//	              Filter(srvtps.FieldName, "", "^api-.*")
//
// ## Pool Merging
//
// Combine multiple pools:
//
//	pool1 := New(nil, nil)
//	pool1.StoreNew(cfg1, nil)
//
//	pool2 := New(nil, nil)
//	pool2.StoreNew(cfg2, nil)
//
//	// Merge pool2 into pool1
//	if err := pool1.Merge(pool2, nil); err != nil {
//	    log.Printf("Merge error: %v", err)
//	}
//
// ## Iteration and Inspection
//
// Walk through servers with custom logic:
//
//	// Iterate all servers
//	p.Walk(func(bindAddress string, srv libhtp.Server) bool {
//	    log.Printf("Server %s at %s", srv.GetName(), bindAddress)
//	    return true // continue iteration
//	})
//
//	// Iterate specific servers
//	p.WalkLimit(func(bindAddress string, srv libhtp.Server) bool {
//	    if srv.IsRunning() {
//	        log.Printf("Running: %s", srv.GetName())
//	    }
//	    return true
//	}, ":8080", ":8443")
//
//	// Check server existence
//	if p.Has(":8080") {
//	    log.Println("Server on :8080 exists")
//	}
//
//	// Get server count
//	log.Printf("Pool has %d servers", p.Len())
//
// ## Monitoring Integration
//
// Access monitoring data for all servers:
//
//	version := libver.New("MyApp", "1.0.0")
//	monitors, err := p.Monitor(version)
//	if err != nil {
//	    log.Printf("Monitor errors: %v", err)
//	}
//
//	for _, mon := range monitors {
//	    log.Printf("Server %s status: %v", mon.Name, mon.Health)
//	}
//
//	// Get monitor identifiers
//	names := p.MonitorNames()
//
// # Use Cases
//
// ## Multi-Port HTTP Server
//
// Run HTTP and HTTPS servers simultaneously:
//
//	httpCfg := libhtp.Config{
//	    Name:   "http",
//	    Listen: ":80",
//	    Expose: "http://example.com",
//	}
//
//	httpsCfg := libhtp.Config{
//	    Name:   "https",
//	    Listen: ":443",
//	    Expose: "https://example.com",
//	    // TLS configuration...
//	}
//
//	p := pool.New(nil, sharedHandler)
//	p.StoreNew(httpCfg, nil)
//	p.StoreNew(httpsCfg, nil)
//	p.Start(context.Background())
//
// ## Microservices Gateway
//
// Route different services on different ports:
//
//	configs := pool.Config{
//	    makeConfig("users-api", ":8081", usersHandler),
//	    makeConfig("orders-api", ":8082", ordersHandler),
//	    makeConfig("payments-api", ":8083", paymentsHandler),
//	}
//
//	p, err := configs.Pool(nil, nil, logger)
//	p.Start(context.Background())
//
// ## Development vs Production
//
// Different configurations per environment:
//
//	var configs pool.Config
//	if isProd {
//	    configs = pool.Config{
//	        makeTLSConfig("api", ":443"),
//	        makeTLSConfig("admin", ":8443"),
//	    }
//	} else {
//	    configs = pool.Config{
//	        makeConfig("api", ":8080"),
//	        makeConfig("admin", ":9000"),
//	    }
//	}
//
//	p, _ := configs.Pool(ctx, handler, logger)
//
// ## Blue-Green Deployment
//
// Gradually switch traffic between server pools:
//
//	bluePool := createPoolFromConfigs(blueConfigs)
//	greenPool := createPoolFromConfigs(greenConfigs)
//
//	// Start green pool
//	greenPool.Start(ctx)
//
//	// Switch traffic...
//	time.Sleep(verificationPeriod)
//
//	// Shutdown blue pool
//	bluePool.Stop(ctx)
//
// ## Admin and Public Separation
//
// Isolate administrative interfaces:
//
//	publicCfg := libhtp.Config{
//	    Name:   "public",
//	    Listen: "0.0.0.0:8080",
//	    Expose: "https://api.example.com",
//	}
//
//	adminCfg := libhtp.Config{
//	    Name:   "admin",
//	    Listen: "127.0.0.1:9000", // localhost only
//	    Expose: "http://localhost:9000",
//	}
//
//	p := pool.New(nil, nil)
//	p.StoreNew(publicCfg, nil)
//	p.StoreNew(adminCfg, nil)
//
// # Performance Characteristics
//
// Pool operations have the following complexity:
//   - Store/Load/Delete: O(1) average, O(n) worst case (map operations)
//   - Walk/WalkLimit: O(n) where n is number of servers
//   - Filter: O(n) with regex matching overhead
//   - List: O(n) + O(m) where m is filtered result size
//   - Start/Stop/Restart: O(n) parallel server operations
//
// Memory usage:
//   - Base pool overhead: ~200 bytes
//   - Per-server overhead: ~100 bytes (map entry)
//   - Total: Base + (n × Server size) + (n × Overhead)
//   - Typical pool with 10 servers: ~50KB
//
// Concurrency:
//   - Read operations scale with goroutines (RLock)
//   - Write operations serialize (Lock)
//   - Server lifecycle operations run concurrently
//   - No goroutine leaks during normal operation
//
// # Error Handling
//
// The package defines error codes in the liberr hierarchy:
//   - ErrorParamEmpty: Invalid or empty parameters
//   - ErrorPoolAdd: Failed to add server to pool
//   - ErrorPoolValidate: Configuration validation failure
//   - ErrorPoolStart: One or more servers failed to start
//   - ErrorPoolStop: One or more servers failed to stop
//   - ErrorPoolRestart: One or more servers failed to restart
//   - ErrorPoolMonitor: Monitoring operation failure
//
// All errors implement liberr.Error interface with:
//   - Error code for programmatic handling
//   - Parent error chains for context
//   - Multiple error aggregation support
//
// # Limitations and Constraints
//
// Known limitations:
//
//  1. Bind Address Uniqueness: Each server must have a unique bind address.
//     Attempting to add servers with duplicate bind addresses will overwrite
//     the existing server.
//
//  2. No Automatic Port Allocation: The pool does not assign ports automatically.
//     All bind addresses must be explicitly configured.
//
//  3. Synchronous Lifecycle: Start/Stop/Restart operations are synchronous and
//     wait for all servers to complete. Use context timeouts for control.
//
//  4. No Load Balancing: The pool manages servers but does not distribute
//     traffic between them. Use external load balancers for traffic distribution.
//
//  5. Error Aggregation: Errors from multiple servers are collected but the
//     first error stops iteration in some operations (e.g., Merge).
//
//  6. No Health Checks: The pool does not perform automatic health checks.
//     Integrate with monitoring systems for health management.
//
// # Best Practices
//
// DO:
//   - Validate all configurations before pool creation
//   - Use unique bind addresses for each server
//   - Set appropriate context timeouts for Start/Stop operations
//   - Check error codes for specific failure types
//   - Use Filter operations to manage subsets of servers
//   - Clean up pools with defer Stop(ctx)
//   - Use monitoring integration for production observability
//
// DON'T:
//   - Don't assume all operations succeed (check errors)
//   - Don't use the same bind address for multiple servers
//   - Don't ignore validation errors
//   - Don't block indefinitely on Start/Stop (use context timeouts)
//   - Don't modify server configurations directly (use pool methods)
//   - Don't forget to handle partial failures during batch operations
//
// # Integration with golib Ecosystem
//
// The pool package integrates with:
//   - github.com/nabbar/golib/httpserver: Server implementation
//   - github.com/nabbar/golib/httpserver/types: Server types and interfaces
//   - github.com/nabbar/golib/context: Context management utilities
//   - github.com/nabbar/golib/logger: Logging integration
//   - github.com/nabbar/golib/errors: Error handling framework
//   - github.com/nabbar/golib/monitor/types: Monitoring abstractions
//   - github.com/nabbar/golib/runner: Lifecycle management interfaces
//
// # Testing
//
// The package includes comprehensive tests:
//   - Pool creation and initialization tests
//   - Server management operation tests (Store/Load/Delete)
//   - Filtering and querying tests
//   - Merge and clone operation tests
//   - Configuration validation tests
//   - Lifecycle management tests
//   - Error handling and edge case tests
//
// Run tests with:
//
//	go test -v ./httpserver/pool
//	CGO_ENABLED=1 go test -race -v ./httpserver/pool
//
// # Thread Safety Validation
//
// All operations are validated for thread safety:
//   - Zero race conditions with -race detector
//   - Concurrent read/write test scenarios
//   - Stress tests with multiple goroutines
//   - Safe server addition during iteration
//
// # Related Packages
//
// For single server management, see:
//   - github.com/nabbar/golib/httpserver: Individual HTTP server
//   - github.com/nabbar/golib/httpserver/types: Server type definitions
//
// For other pooling patterns, see:
//   - github.com/nabbar/golib/runner/startStop: Generic runner pool
package pool
