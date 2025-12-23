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

package httpserver_test

import (
	"fmt"
	"net/http"

	"github.com/nabbar/golib/httpserver"
)

// ExampleNew demonstrates the simplest way to create an HTTP server.
// This is the most basic use case for server creation.
func ExampleNew() {
	cfg := httpserver.Config{
		Name:   "example-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}

	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{
			"": http.NotFoundHandler(),
		}
	})

	srv, err := httpserver.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Server created: %s\n", srv.GetName())
	// Output:
	// Server created: example-server
}

// Example_basicServer demonstrates creating and configuring a basic HTTP server.
// Shows the essential steps: configuration, validation, and creation.
func Example_basicServer() {
	cfg := httpserver.Config{
		Name:   "basic-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}

	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		return map[string]http.Handler{"": mux}
	})

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	srv, err := httpserver.New(cfg, nil)
	if err != nil {
		fmt.Printf("Creation failed: %v\n", err)
		return
	}

	fmt.Printf("Server: %s on %s\n", srv.GetName(), srv.GetBindable())
	// Output:
	// Server: basic-server on 127.0.0.1:8080
}

// Example_serverInfo demonstrates accessing server information.
// Shows how to retrieve server properties after creation.
func Example_serverInfo() {
	cfg := httpserver.Config{
		Name:   "info-server",
		Listen: "127.0.0.1:9000",
		Expose: "http://localhost:9000",
	}

	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, _ := httpserver.New(cfg, nil)

	fmt.Printf("Name: %s\n", srv.GetName())
	fmt.Printf("Bind: %s\n", srv.GetBindable())
	fmt.Printf("Expose: %s\n", srv.GetExpose())
	fmt.Printf("Disabled: %t\n", srv.IsDisable())
	fmt.Printf("TLS: %t\n", srv.IsTLS())
	// Output:
	// Name: info-server
	// Bind: 127.0.0.1:9000
	// Expose: localhost:9000
	// Disabled: false
	// TLS: false
}

// Example_configValidation demonstrates configuration validation.
// Shows how validation catches configuration errors early.
func Example_configValidation() {
	validCfg := httpserver.Config{
		Name:   "valid-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}

	if err := validCfg.Validate(); err != nil {
		fmt.Println("Valid config failed")
	} else {
		fmt.Println("Valid config passed")
	}

	invalidCfg := httpserver.Config{
		Name: "invalid-server",
		// Missing Listen and Expose
	}

	if err := invalidCfg.Validate(); err != nil {
		fmt.Println("Invalid config failed as expected")
	} else {
		fmt.Println("Invalid config unexpectedly passed")
	}
	// Output:
	// Valid config passed
	// Invalid config failed as expected
}

// Example_handlerRegistration demonstrates handler registration.
// Shows how to register custom HTTP handlers.
func Example_handlerRegistration() {
	cfg := httpserver.Config{
		Name:   "handler-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}

	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World"))
		})
		return map[string]http.Handler{"": mux}
	})

	srv, err := httpserver.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Handler registered for %s\n", srv.GetName())
	// Output:
	// Handler registered for handler-server
}

// Example_multipleHandlers demonstrates multiple named handlers.
// Shows how to use handler keys for different handlers.
func Example_multipleHandlers() {
	handlerFunc := func() map[string]http.Handler {
		return map[string]http.Handler{
			"api":   http.NotFoundHandler(),
			"admin": http.NotFoundHandler(),
			"web":   http.NotFoundHandler(),
		}
	}

	apiCfg := httpserver.Config{
		Name:       "api-server",
		Listen:     "127.0.0.1:8080",
		Expose:     "http://localhost:8080",
		HandlerKey: "api",
	}
	apiCfg.RegisterHandlerFunc(handlerFunc)

	adminCfg := httpserver.Config{
		Name:       "admin-server",
		Listen:     "127.0.0.1:8081",
		Expose:     "http://localhost:8081",
		HandlerKey: "admin",
	}
	adminCfg.RegisterHandlerFunc(handlerFunc)

	apiSrv, _ := httpserver.New(apiCfg, nil)
	adminSrv, _ := httpserver.New(adminCfg, nil)

	fmt.Printf("API server: %s\n", apiSrv.GetName())
	fmt.Printf("Admin server: %s\n", adminSrv.GetName())
	// Output:
	// API server: api-server
	// Admin server: admin-server
}

// Example_disabledServer demonstrates the disabled flag.
// Shows how to create a server that won't start.
func Example_disabledServer() {
	cfg := httpserver.Config{
		Name:     "disabled-server",
		Listen:   "127.0.0.1:8080",
		Expose:   "http://localhost:8080",
		Disabled: true,
	}

	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, _ := httpserver.New(cfg, nil)

	fmt.Printf("Server disabled: %t\n", srv.IsDisable())
	// Output:
	// Server disabled: true
}

// Example_configClone demonstrates configuration cloning.
// Shows how to create independent configuration copies.
func Example_configClone() {
	original := httpserver.Config{
		Name:   "original",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}

	cloned := original.Clone()
	cloned.Name = "cloned"

	fmt.Printf("Original: %s\n", original.Name)
	fmt.Printf("Cloned: %s\n", cloned.Name)
	// Output:
	// Original: original
	// Cloned: cloned
}

// Example_serverMerge demonstrates merging server configurations.
// Shows how to update one server with another's configuration.
func Example_serverMerge() {
	cfg1 := httpserver.Config{
		Name:   "server1",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg1.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	cfg2 := httpserver.Config{
		Name:   "server2",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg2.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv1, _ := httpserver.New(cfg1, nil)
	srv2, _ := httpserver.New(cfg2, nil)

	fmt.Printf("Before merge: %s\n", srv1.GetName())
	srv1.Merge(srv2, nil)
	fmt.Printf("After merge: %s\n", srv1.GetName())
	// Output:
	// Before merge: server1
	// After merge: server2
}

// Example_setConfig demonstrates updating server configuration.
// Shows how to change server settings after creation.
func Example_setConfig() {
	originalCfg := httpserver.Config{
		Name:   "original-name",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	originalCfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, _ := httpserver.New(originalCfg, nil)
	fmt.Printf("Initial: %s\n", srv.GetName())

	updatedCfg := httpserver.Config{
		Name:   "updated-name",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	updatedCfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv.SetConfig(updatedCfg, nil)
	fmt.Printf("Updated: %s\n", srv.GetName())
	// Output:
	// Initial: original-name
	// Updated: updated-name
}

// Example_getConfig demonstrates retrieving server configuration.
// Shows how to access the current configuration.
func Example_getConfig() {
	cfg := httpserver.Config{
		Name:   "config-test",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, _ := httpserver.New(cfg, nil)
	retrievedCfg := srv.GetConfig()

	fmt.Printf("Name: %s\n", retrievedCfg.Name)
	fmt.Printf("Listen: %s\n", retrievedCfg.Listen)
	// Output:
	// Name: config-test
	// Listen: 127.0.0.1:8080
}

// Example_monitorName demonstrates monitoring identifier.
// Shows how to get the unique monitoring name.
func Example_monitorName() {
	cfg := httpserver.Config{
		Name:   "monitor-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, _ := httpserver.New(cfg, nil)
	monitorName := srv.MonitorName()

	fmt.Printf("Monitor name contains bind: %t\n", len(monitorName) > 0)
	// Output:
	// Monitor name contains bind: true
}

// Example_lifecycleState demonstrates checking server state.
// Shows how to query if a server is running.
func Example_lifecycleState() {
	cfg := httpserver.Config{
		Name:   "state-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, _ := httpserver.New(cfg, nil)

	fmt.Printf("Running initially: %t\n", srv.IsRunning())
	// Note: Not actually starting to keep test simple
	// Output:
	// Running initially: false
}

// Example_serverFromConfig demonstrates creating server from config method.
// Shows the convenience method on Config.
func Example_serverFromConfig() {
	cfg := httpserver.Config{
		Name:   "convenience-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, err := cfg.Server(nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Created via config method: %s\n", srv.GetName())
	// Output:
	// Created via config method: convenience-server
}

// Example_dynamicHandler demonstrates dynamic handler replacement.
// Shows how to update handlers after server creation.
func Example_dynamicHandler() {
	cfg := httpserver.Config{
		Name:   "dynamic-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, _ := httpserver.New(cfg, nil)

	newHandler := func() map[string]http.Handler {
		mux := http.NewServeMux()
		mux.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("New Handler"))
		})
		return map[string]http.Handler{"": mux}
	}

	srv.Handler(newHandler)
	fmt.Println("Handler updated dynamically")
	// Output:
	// Handler updated dynamically
}

// Example_portBinding demonstrates different bind address formats.
// Shows various ways to specify the listen address.
func Example_portBinding() {
	configs := []httpserver.Config{
		{
			Name:   "localhost",
			Listen: "localhost:8080",
			Expose: "http://localhost:8080",
		},
		{
			Name:   "specific-ip",
			Listen: "192.168.1.100:8080",
			Expose: "http://192.168.1.100:8080",
		},
		{
			Name:   "all-interfaces",
			Listen: "0.0.0.0:8080",
			Expose: "http://localhost:8080",
		},
	}

	for _, cfg := range configs {
		if err := cfg.Validate(); err != nil {
			fmt.Printf("%s: invalid\n", cfg.Name)
		} else {
			fmt.Printf("%s: valid\n", cfg.Name)
		}
	}
	// Output:
	// localhost: valid
	// specific-ip: valid
	// all-interfaces: valid
}

// Example_httpVersions demonstrates HTTP and HTTPS configuration.
// Shows both HTTP and HTTPS expose URLs.
func Example_httpVersions() {
	httpCfg := httpserver.Config{
		Name:   "http-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}

	httpsCfg := httpserver.Config{
		Name:   "https-server",
		Listen: "127.0.0.1:8443",
		Expose: "https://localhost:8443",
	}

	fmt.Printf("HTTP valid: %v\n", httpCfg.Validate() == nil)
	fmt.Printf("HTTPS valid: %v\n", httpsCfg.Validate() == nil)
	// Output:
	// HTTP valid: true
	// HTTPS valid: true
}

// Example_complete demonstrates a complete server setup workflow.
// Shows a realistic scenario with all components.
func Example_complete() {
	cfg := httpserver.Config{
		Name:   "complete-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}

	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello"))
		})
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		return map[string]http.Handler{"": mux}
	})

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	srv, err := httpserver.New(cfg, nil)
	if err != nil {
		fmt.Printf("Creation error: %v\n", err)
		return
	}

	fmt.Printf("Server: %s\n", srv.GetName())
	fmt.Printf("Binding: %s\n", srv.GetBindable())
	fmt.Printf("Expose: %s\n", srv.GetExpose())
	fmt.Printf("Ready: %t\n", !srv.IsDisable())
	// Output:
	// Server: complete-server
	// Binding: 127.0.0.1:8080
	// Expose: localhost:8080
	// Ready: true
}

// Example_gracefulPattern demonstrates a graceful shutdown pattern.
// Shows the recommended way to handle server lifecycle.
func Example_gracefulPattern() {
	cfg := httpserver.Config{
		Name:   "graceful-server",
		Listen: "127.0.0.1:8080",
		Expose: "http://localhost:8080",
	}
	cfg.RegisterHandlerFunc(func() map[string]http.Handler {
		return map[string]http.Handler{"": http.NotFoundHandler()}
	})

	srv, err := httpserver.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// In real code, you would call:
	// ctx := context.Background()
	// err = srv.Start(ctx)
	// defer srv.Stop(ctx)

	fmt.Printf("Server ready for lifecycle: %s\n", srv.GetName())
	// Output:
	// Server ready for lifecycle: graceful-server
}
