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

package pool_test

import (
	"context"
	"fmt"
	"net/http"

	libhtp "github.com/nabbar/golib/httpserver"
	"github.com/nabbar/golib/httpserver/pool"
	srvtps "github.com/nabbar/golib/httpserver/types"
)

// ExampleNew demonstrates creating an empty pool.
// This is the simplest usage pattern - creating a pool container.
func ExampleNew() {
	p := pool.New(nil, nil)
	fmt.Printf("Pool created with %d servers\n", p.Len())
	// Output:
	// Pool created with 0 servers
}

// Example_simplePool demonstrates creating a pool with a single server.
// This shows basic server configuration and addition to the pool.
func Example_simplePool() {
	p := pool.New(nil, nil)

	cfg := libhtp.Config{
		Name:   "simple-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	err := p.StoreNew(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Pool has %d server(s)\n", p.Len())
	// Output:
	// Pool has 1 server(s)
}

// Example_multipleServers demonstrates managing multiple servers in a pool.
// Shows how to add several servers with different configurations.
func Example_multipleServers() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := []libhtp.Config{
		{Name: "api", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "admin", Listen: "127.0.0.1:9000", Expose: "http://localhost:9000"},
		{Name: "metrics", Listen: "127.0.0.1:2112", Expose: "http://localhost:2112"},
	}

	for _, cfg := range configs {
		cfg.RegisterHandlerFunc(handler)
		p.StoreNew(cfg, nil)
	}

	fmt.Printf("Pool contains %d servers\n", p.Len())
	// Output:
	// Pool contains 3 servers
}

// Example_configSlice demonstrates using Config slice for bulk operations.
// Shows validation and pool creation from multiple configurations.
func Example_configSlice() {
	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := pool.Config{
		{Name: "server1", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "server2", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
	}

	configs.SetHandlerFunc(handler)

	err := configs.Validate()
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	p, err := configs.Pool(nil, nil, nil)
	if err != nil {
		fmt.Printf("Pool creation error: %v\n", err)
		return
	}

	fmt.Printf("Created pool with %d servers\n", p.Len())
	// Output:
	// Created pool with 2 servers
}

// Example_loadAndCheck demonstrates loading servers and checking existence.
// Shows how to retrieve and verify servers in the pool.
func Example_loadAndCheck() {
	p := pool.New(nil, nil)

	cfg := libhtp.Config{
		Name:   "test-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	p.StoreNew(cfg, nil)

	if p.Has("127.0.0.1:8080") {
		srv := p.Load("127.0.0.1:8080")
		fmt.Printf("Found server: %s\n", srv.GetName())
	}

	if !p.Has("127.0.0.1:9999") {
		fmt.Println("Server on :9999 not found")
	}

	// Output:
	// Found server: test-server
	// Server on :9999 not found
}

// Example_walk demonstrates iterating over all servers in the pool.
// Shows how to execute logic for each server.
func Example_walk() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := []libhtp.Config{
		{Name: "server-a", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "server-b", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
	}

	for _, cfg := range configs {
		cfg.RegisterHandlerFunc(handler)
		p.StoreNew(cfg, nil)
	}

	var count int
	p.Walk(func(bindAddress string, srv libhtp.Server) bool {
		count++
		return true
	})

	fmt.Printf("Walked through %d servers\n", count)

	// Output:
	// Walked through 2 servers
}

// Example_walkLimit demonstrates iterating over specific servers.
// Shows how to process only selected servers by bind address.
func Example_walkLimit() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := []libhtp.Config{
		{Name: "api", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "web", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
		{Name: "admin", Listen: "127.0.0.1:9000", Expose: "http://localhost:9000"},
	}

	for _, cfg := range configs {
		cfg.RegisterHandlerFunc(handler)
		p.StoreNew(cfg, nil)
	}

	var names []string
	p.WalkLimit(func(bindAddress string, srv libhtp.Server) bool {
		names = append(names, srv.GetName())
		return true
	}, "127.0.0.1:8080", "127.0.0.1:9000")

	if len(names) == 2 {
		fmt.Printf("Selected %d servers\n", len(names))
	}

	// Output:
	// Selected 2 servers
}

// Example_filter demonstrates filtering servers by criteria.
// Shows how to create filtered subsets of the pool.
func Example_filter() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := []libhtp.Config{
		{Name: "api-server", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "api-v2-server", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
		{Name: "web-server", Listen: "127.0.0.1:9000", Expose: "http://localhost:9000"},
	}

	for _, cfg := range configs {
		cfg.RegisterHandlerFunc(handler)
		p.StoreNew(cfg, nil)
	}

	filtered := p.Filter(srvtps.FieldName, "", "^api-.*")

	fmt.Printf("Filtered pool has %d servers\n", filtered.Len())
	// Output:
	// Filtered pool has 2 servers
}

// Example_list demonstrates listing server attributes.
// Shows how to extract specific fields from servers.
func Example_list() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := []libhtp.Config{
		{Name: "server1", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "server2", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
	}

	for _, cfg := range configs {
		cfg.RegisterHandlerFunc(handler)
		p.StoreNew(cfg, nil)
	}

	names := p.List(srvtps.FieldName, srvtps.FieldName, "", ".*")

	fmt.Printf("Found %d servers\n", len(names))

	// Output:
	// Found 2 servers
}

// Example_delete demonstrates removing servers from the pool.
// Shows server deletion operations.
func Example_delete() {
	p := pool.New(nil, nil)

	cfg := libhtp.Config{
		Name:   "temp-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	p.StoreNew(cfg, nil)
	fmt.Printf("Before delete: %d servers\n", p.Len())

	p.Delete("127.0.0.1:8080")
	fmt.Printf("After delete: %d servers\n", p.Len())

	// Output:
	// Before delete: 1 servers
	// After delete: 0 servers
}

// Example_loadAndDelete demonstrates atomic load and delete.
// Shows how to retrieve and remove a server in one operation.
func Example_loadAndDelete() {
	p := pool.New(nil, nil)

	cfg := libhtp.Config{
		Name:   "remove-me",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	p.StoreNew(cfg, nil)

	srv, loaded := p.LoadAndDelete("127.0.0.1:8080")
	if loaded {
		fmt.Printf("Removed: %s\n", srv.GetName())
	}

	fmt.Printf("Pool now has %d servers\n", p.Len())

	// Output:
	// Removed: remove-me
	// Pool now has 0 servers
}

// Example_clean demonstrates clearing all servers.
// Shows how to empty the pool.
func Example_clean() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	cfg1 := libhtp.Config{Name: "s1", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"}
	cfg1.RegisterHandlerFunc(handler)
	p.StoreNew(cfg1, nil)

	cfg2 := libhtp.Config{Name: "s2", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"}
	cfg2.RegisterHandlerFunc(handler)
	p.StoreNew(cfg2, nil)

	fmt.Printf("Before clean: %d servers\n", p.Len())

	p.Clean()
	fmt.Printf("After clean: %d servers\n", p.Len())

	// Output:
	// Before clean: 2 servers
	// After clean: 0 servers
}

// Example_merge demonstrates merging two pools.
// Shows how to combine servers from different pools.
func Example_merge() {
	pool1 := pool.New(nil, nil)
	pool2 := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	cfg1 := libhtp.Config{Name: "server1", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"}
	cfg1.RegisterHandlerFunc(handler)
	pool1.StoreNew(cfg1, nil)

	cfg2 := libhtp.Config{Name: "server2", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"}
	cfg2.RegisterHandlerFunc(handler)
	pool2.StoreNew(cfg2, nil)

	err := pool1.Merge(pool2, nil)
	if err != nil {
		fmt.Printf("Merge error: %v\n", err)
		return
	}

	fmt.Printf("Merged pool has %d servers\n", pool1.Len())
	// Output:
	// Merged pool has 2 servers
}

// Example_clone demonstrates cloning a pool.
// Shows how to create an independent copy of a pool.
func Example_clone() {
	original := pool.New(nil, nil)

	cfg := libhtp.Config{
		Name:   "original-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	original.StoreNew(cfg, nil)

	cloned := original.Clone(context.Background())

	fmt.Printf("Original: %d servers\n", original.Len())
	fmt.Printf("Cloned: %d servers\n", cloned.Len())

	// Output:
	// Original: 1 servers
	// Cloned: 1 servers
}

// Example_monitorNames demonstrates accessing monitoring identifiers.
// Shows how to retrieve monitor names for all servers.
func Example_monitorNames() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := []libhtp.Config{
		{Name: "api", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "web", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
	}

	for _, cfg := range configs {
		cfg.RegisterHandlerFunc(handler)
		p.StoreNew(cfg, nil)
	}

	names := p.MonitorNames()
	fmt.Printf("Monitor count: %d\n", len(names))

	// Output:
	// Monitor count: 2
}

// Example_configWalk demonstrates walking through configuration slice.
// Shows how to iterate over configurations before pool creation.
func Example_configWalk() {
	configs := pool.Config{
		{Name: "config1", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "config2", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
	}

	var count int
	configs.Walk(func(cfg libhtp.Config) bool {
		count++
		fmt.Printf("Config: %s\n", cfg.Name)
		return true
	})

	fmt.Printf("Total: %d configs\n", count)

	// Output:
	// Config: config1
	// Config: config2
	// Total: 2 configs
}

// Example_handlerUpdate demonstrates updating handler function.
// Shows how to change the shared handler function.
func Example_handlerUpdate() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{
			"/api": http.NotFoundHandler(),
		}
	}

	p.Handler(handler)

	fmt.Println("Handler updated successfully")
	// Output:
	// Handler updated successfully
}

// Example_complexFiltering demonstrates chaining multiple filters.
// Shows advanced filtering techniques.
func Example_complexFiltering() {
	p := pool.New(nil, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := []libhtp.Config{
		{Name: "api-public", Listen: "0.0.0.0:8080", Expose: "http://api.example.com:8080"},
		{Name: "api-private", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "web-public", Listen: "0.0.0.0:80", Expose: "http://www.example.com"},
	}

	for _, cfg := range configs {
		cfg.RegisterHandlerFunc(handler)
		p.StoreNew(cfg, nil)
	}

	filtered := p.Filter(srvtps.FieldBind, "", "^0\\.0\\.0\\.0:.*").
		Filter(srvtps.FieldName, "", "^api-.*")

	fmt.Printf("Public API servers: %d\n", filtered.Len())
	// Output:
	// Public API servers: 1
}

// Example_multiStepPool demonstrates a complete pool lifecycle.
// Shows configuration, creation, management, and cleanup.
func Example_multiStepPool() {
	handler := func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	}

	configs := pool.Config{
		{Name: "primary", Listen: "127.0.0.1:8080", Expose: "http://localhost:8080"},
		{Name: "backup", Listen: "127.0.0.1:8081", Expose: "http://localhost:8081"},
	}

	configs.SetHandlerFunc(handler)

	if err := configs.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	p, err := configs.Pool(nil, nil, nil)
	if err != nil {
		fmt.Printf("Pool creation failed: %v\n", err)
		return
	}

	fmt.Printf("Phase 1: %d servers\n", p.Len())

	newCfg := libhtp.Config{
		Name:   "emergency",
		Listen: "127.0.0.1:9000",
		Expose: "http://localhost:9000",
	}
	newCfg.RegisterHandlerFunc(handler)
	p.StoreNew(newCfg, nil)

	fmt.Printf("Phase 2: %d servers\n", p.Len())

	p.Delete("127.0.0.1:8081")

	fmt.Printf("Phase 3: %d servers\n", p.Len())

	// Output:
	// Phase 1: 2 servers
	// Phase 2: 3 servers
	// Phase 3: 2 servers
}
