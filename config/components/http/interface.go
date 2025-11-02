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

package http

import (
	"context"
	"net/http"

	libatm "github.com/nabbar/golib/atomic"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	htpool "github.com/nabbar/golib/httpserver/pool"
	srvtps "github.com/nabbar/golib/httpserver/types"
)

const (
	// DefaultTlsKey is the default key used to reference the TLS component.
	// This key is used when no custom TLS key is provided during component creation.
	DefaultTlsKey = "t"
)

// CptHttp represents an HTTP server component that manages HTTP/HTTPS server pools.
// It extends the base Component interface with HTTP-specific functionality including
// TLS configuration, handler management, and server pool operations.
//
// The component supports:
//   - Multiple HTTP/HTTPS servers with different configurations
//   - Dynamic TLS configuration through component dependencies
//   - Custom HTTP handlers for different routes
//   - Server pool lifecycle management (start, stop, reload)
//   - Health monitoring integration
//
// Usage:
//
//	cpt := http.New(ctx, "tls-key", handlerFunc)
//	cpt.Init("http-server", ctx, getCpt, vpr, vrs, log)
//	err := cpt.Start()
//	if err != nil {
//	    // Handle error
//	}
//	defer cpt.Stop()
type CptHttp interface {
	cfgtps.Component

	// SetTLSKey sets the key used to reference the TLS component.
	// This key is used to load TLS configuration from the component registry.
	SetTLSKey(tlsKey string)

	// SetHandler sets the function that returns HTTP handlers for different routes.
	// The handler function is called when building the server pool.
	SetHandler(fct srvtps.FuncHandler)

	// GetPool returns the current HTTP server pool.
	// Returns nil if the pool has not been initialized.
	GetPool() htpool.Pool

	// SetPool sets the HTTP server pool.
	// If nil is passed, a new pool is created automatically.
	SetPool(pool htpool.Pool)
}

// New creates a new HTTP component instance with the specified context, TLS key, and handler function.
//
// Parameters:
//   - ctx: Function that returns the context for the component
//   - tlsKey: Key to reference the TLS component (uses DefaultTlsKey if empty)
//   - hdl: Function that returns HTTP handlers for different routes (can be nil)
//
// Returns:
//   - CptHttp: A new HTTP component instance
//
// Example:
//
//	ctx := func() context.Context { return context.Background() }
//	handler := func() map[string]http.Handler {
//	    return map[string]http.Handler{
//	        "api": apiHandler,
//	        "status": statusHandler,
//	    }
//	}
//	cpt := http.New(ctx, "my-tls", handler)
func New(ctx context.Context, tlsKey string, hdl srvtps.FuncHandler) CptHttp {
	if tlsKey == "" {
		tlsKey = DefaultTlsKey
	}

	fdh := func() map[string]http.Handler {
		return map[string]http.Handler{}
	}

	fdp := htpool.New(ctx, fdh)

	c := &mod{
		x: libctx.New[uint8](ctx),
		t: libatm.NewValue[string](),
		h: libatm.NewValueDefault[srvtps.FuncHandler](fdh, fdh),
		s: libatm.NewValueDefault[htpool.Pool](fdp, fdp),
	}

	c.t.Store(tlsKey)
	c.h.Store(hdl)

	return c
}

// Register registers an existing HTTP component with the configuration system.
//
// Parameters:
//   - cfg: The configuration instance
//   - key: The key to use for this component
//   - cpt: The HTTP component to register
//
// This function is typically used when you have created a component with New()
// and want to register it in the configuration system.
func Register(cfg libcfg.Config, key string, cpt CptHttp) {
	cfg.ComponentSet(key, cpt)
}

// RegisterNew creates a new HTTP component and registers it with the configuration system.
//
// Parameters:
//   - ctx: Function that returns the context for the component
//   - cfg: The configuration instance
//   - key: The key to use for this component
//   - tlsKey: Key to reference the TLS component
//   - hdl: Function that returns HTTP handlers for different routes
//
// This is a convenience function that combines New() and Register().
func RegisterNew(ctx context.Context, cfg libcfg.Config, key string, tlsKey string, hdl srvtps.FuncHandler) {
	cfg.ComponentSet(key, New(ctx, tlsKey, hdl))
}

// Load retrieves an HTTP component from the configuration system by its key.
//
// Parameters:
//   - getCpt: Function to retrieve components by key
//   - key: The key of the component to load
//
// Returns:
//   - CptHttp: The HTTP component if found and of correct type, nil otherwise
//
// This function performs type checking and returns nil if the component
// is not found or is not of type CptHttp.
func Load(getCpt cfgtps.FuncCptGet, key string) CptHttp {
	if getCpt == nil {
		return nil
	} else if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(CptHttp); !ok {
		return nil
	} else {
		return h
	}
}
