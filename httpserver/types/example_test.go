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

package types_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/nabbar/golib/httpserver/types"
)

// ExampleFieldType demonstrates basic usage of FieldType enumeration.
// This is the simplest use case for field identification.
func ExampleFieldType() {
	field := types.FieldName

	fmt.Printf("Field type: %d\n", field)
	// Output:
	// Field type: 0
}

// Example_fieldTypeSwitch demonstrates using FieldType in switch statements.
// This shows how to handle different field types in filtering operations.
func Example_fieldTypeSwitch() {
	fields := []types.FieldType{
		types.FieldName,
		types.FieldBind,
		types.FieldExpose,
	}

	for _, field := range fields {
		switch field {
		case types.FieldName:
			fmt.Println("Filtering by name")
		case types.FieldBind:
			fmt.Println("Filtering by bind address")
		case types.FieldExpose:
			fmt.Println("Filtering by expose URL")
		}
	}
	// Output:
	// Filtering by name
	// Filtering by bind address
	// Filtering by expose URL
}

// Example_fieldTypeMap demonstrates using FieldType as map keys.
// This pattern is useful for storing field-specific configurations.
func Example_fieldTypeMap() {
	fieldNames := map[types.FieldType]string{
		types.FieldName:   "server-name",
		types.FieldBind:   "bind-address",
		types.FieldExpose: "expose-url",
	}

	fields := []types.FieldType{types.FieldName, types.FieldBind, types.FieldExpose}
	for _, field := range fields {
		fmt.Printf("Field %d: %s\n", field, fieldNames[field])
	}
	// Output:
	// Field 0: server-name
	// Field 1: bind-address
	// Field 2: expose-url
}

// ExampleNewBadHandler demonstrates creating a fallback error handler.
// This is the simplest way to create a safe default handler.
func ExampleNewBadHandler() {
	handler := types.NewBadHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 500
}

// Example_badHandlerMultipleRequests demonstrates BadHandler with multiple requests.
// Shows that BadHandler consistently returns 500 for all requests.
func Example_badHandlerMultipleRequests() {
	handler := types.NewBadHandler()

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		fmt.Printf("%s: %d\n", method, w.Code)
	}
	// Output:
	// GET: 500
	// POST: 500
	// PUT: 500
}

// Example_handlerDefault demonstrates using the HandlerDefault constant.
// This shows the standard key for default handler registration.
func Example_handlerDefault() {
	handlers := map[string]http.Handler{
		types.HandlerDefault: http.NotFoundHandler(),
	}

	if handler, exists := handlers[types.HandlerDefault]; exists {
		fmt.Printf("Default handler registered: %t\n", handler != nil)
	}
	// Output:
	// Default handler registered: true
}

// Example_funcHandler demonstrates implementing FuncHandler.
// This is a simple example of handler registration function.
func Example_funcHandler() {
	var handlerFunc types.FuncHandler

	handlerFunc = func() map[string]http.Handler {
		return map[string]http.Handler{
			types.HandlerDefault: http.NotFoundHandler(),
		}
	}

	handlers := handlerFunc()
	fmt.Printf("Handlers returned: %d\n", len(handlers))
	// Output:
	// Handlers returned: 1
}

// Example_funcHandlerMultiple demonstrates multiple handler registration.
// This shows how to register multiple named handlers.
func Example_funcHandlerMultiple() {
	var handlerFunc types.FuncHandler

	handlerFunc = func() map[string]http.Handler {
		return map[string]http.Handler{
			types.HandlerDefault: types.NewBadHandler(),
			"api":                http.NotFoundHandler(),
			"admin":              http.NotFoundHandler(),
		}
	}

	handlers := handlerFunc()
	fmt.Printf("Total handlers: %d\n", len(handlers))

	keys := []string{types.HandlerDefault, "api", "admin"}
	for _, key := range keys {
		if _, exists := handlers[key]; exists {
			fmt.Printf("Handler key: %s\n", key)
		}
	}
	// Output:
	// Total handlers: 3
	// Handler key: default
	// Handler key: api
	// Handler key: admin
}

// Example_timeoutConstants demonstrates using timeout constants.
// This shows how to access and use predefined timeout values.
func Example_timeoutConstants() {
	portTimeout := types.TimeoutWaitingPortFreeing
	stopTimeout := types.TimeoutWaitingStop

	fmt.Printf("Port freeing timeout: %v\n", portTimeout)
	fmt.Printf("Stop timeout: %v\n", stopTimeout)
	// Output:
	// Port freeing timeout: 250µs
	// Stop timeout: 5s
}

// Example_badHandlerName demonstrates the BadHandlerName constant.
// This shows the identifier used for BadHandler instances.
func Example_badHandlerName() {
	handlerName := types.BadHandlerName

	fmt.Printf("Bad handler identifier: %s\n", handlerName)
	// Output:
	// Bad handler identifier: no handler
}

// Example_fieldTypeComparison demonstrates comparing FieldType values.
// This is useful for validation and conditional logic.
func Example_fieldTypeComparison() {
	field1 := types.FieldName
	field2 := types.FieldName
	field3 := types.FieldBind

	fmt.Printf("field1 == field2: %t\n", field1 == field2)
	fmt.Printf("field1 == field3: %t\n", field1 == field3)
	// Output:
	// field1 == field2: true
	// field1 == field3: false
}

// Example_handlerWithFallback demonstrates using BadHandler as fallback.
// This pattern provides safe defaults when handler registration fails.
func Example_handlerWithFallback() {
	var handlerFunc types.FuncHandler

	handlerFunc = func() map[string]http.Handler {
		return nil
	}

	handlers := handlerFunc()
	var handler http.Handler

	if handlers == nil || handlers[types.HandlerDefault] == nil {
		handler = types.NewBadHandler()
		fmt.Println("Using fallback handler")
	} else {
		handler = handlers[types.HandlerDefault]
		fmt.Println("Using configured handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Using fallback handler
	// Status: 500
}

// Example_serverFiltering demonstrates filtering pattern with FieldType.
// This shows a realistic server filtering scenario.
func Example_serverFiltering() {
	type srv struct {
		name   string
		bind   string
		expose string
	}

	servers := []srv{
		{name: "api-server", bind: ":8080", expose: "https://api.example.com"},
		{name: "web-server", bind: ":8081", expose: "https://www.example.com"},
		{name: "admin-server", bind: ":8082", expose: "https://admin.example.com"},
	}

	filterByField := func(field types.FieldType, value string) []srv {
		var result []srv
		for _, s := range servers {
			var match bool
			switch field {
			case types.FieldName:
				match = s.name == value
			case types.FieldBind:
				match = s.bind == value
			case types.FieldExpose:
				match = s.expose == value
			}
			if match {
				result = append(result, s)
			}
		}
		return result
	}

	results := filterByField(types.FieldBind, ":8080")
	for _, s := range results {
		fmt.Printf("Found: %s\n", s.name)
	}
	// Output:
	// Found: api-server
}

// Example_handlerRegistration demonstrates complete handler registration.
// This shows a realistic multi-handler server configuration.
func Example_handlerRegistration() {
	createHandlers := func() types.FuncHandler {
		return func() map[string]http.Handler {
			apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			webHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			return map[string]http.Handler{
				types.HandlerDefault: webHandler,
				"api":                apiHandler,
			}
		}
	}

	handlerFunc := createHandlers()
	handlers := handlerFunc()

	fmt.Printf("Registered handlers: %d\n", len(handlers))

	if _, exists := handlers[types.HandlerDefault]; exists {
		fmt.Println("Default handler: configured")
	}

	if _, exists := handlers["api"]; exists {
		fmt.Println("API handler: configured")
	}
	// Output:
	// Registered handlers: 2
	// Default handler: configured
	// API handler: configured
}

// Example_handlerValidation demonstrates validating handler registration.
// This shows a pattern for checking handler configuration.
func Example_handlerValidation() {
	validateHandlers := func(handlerFunc types.FuncHandler) error {
		if handlerFunc == nil {
			return fmt.Errorf("handler function is nil")
		}

		handlers := handlerFunc()
		if handlers == nil {
			return fmt.Errorf("handlers map is nil")
		}

		if handlers[types.HandlerDefault] == nil {
			return fmt.Errorf("default handler not configured")
		}

		return nil
	}

	validFunc := func() map[string]http.Handler {
		return map[string]http.Handler{
			types.HandlerDefault: http.NotFoundHandler(),
		}
	}

	if err := validateHandlers(validFunc); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Validation passed")
	}
	// Output:
	// Validation passed
}

// Example_complete demonstrates combining all types package features.
// This is the most complex example showing realistic integration.
func Example_complete() {
	type srvCfg struct {
		name    string
		bind    string
		expose  string
		handler http.Handler
	}

	createServerConfig := func(name, bind, expose string, handlerFunc types.FuncHandler) srvCfg {
		handlers := handlerFunc()
		handler := handlers[types.HandlerDefault]

		if handler == nil {
			handler = types.NewBadHandler()
		}

		return srvCfg{
			name:    name,
			bind:    bind,
			expose:  expose,
			handler: handler,
		}
	}

	handlerFunc := func() map[string]http.Handler {
		return map[string]http.Handler{
			types.HandlerDefault: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		}
	}

	cfg := createServerConfig("api-server", ":8080", "https://api.example.com", handlerFunc)

	fmt.Printf("Server: %s\n", cfg.name)
	fmt.Printf("Bind: %s\n", cfg.bind)
	fmt.Printf("Expose: %s\n", cfg.expose)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	cfg.handler.ServeHTTP(w, req)

	fmt.Printf("Handler status: %d\n", w.Code)
	// Output:
	// Server: api-server
	// Bind: :8080
	// Expose: https://api.example.com
	// Handler status: 200
}
