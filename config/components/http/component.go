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

	cpttls "github.com/nabbar/golib/config/components/tls"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

const (
	// ComponentType is the identifier for the HTTP component type.
	// This constant is used by the configuration system to identify HTTP components.
	ComponentType = "http"

	// Internal context keys for storing component state
	keyCptKey          = iota + 1 // Component key in configuration
	keyCptDependencies            // Custom dependencies list
	keyFctViper                   // Viper configuration function
	keyFctGetCpt                  // Component getter function
	keyCptVersion                 // Version information
	keyCptLogger                  // Logger function
	keyFctStaBef                  // Before-start callback
	keyFctStaAft                  // After-start callback
	keyFctRelBef                  // Before-reload callback
	keyFctRelAft                  // After-reload callback
	keyFctMonitorPool             // Monitor pool function
)

// Type returns the component type identifier.
// This implements the cfgtps.Component interface.
func (o *mod) Type() string {
	return ComponentType
}

// Init initializes the HTTP component with required dependencies.
// This implements the cfgtps.Component interface.
//
// Parameters:
//   - key: The configuration key for this component
//   - ctx: Function returning the component's context
//   - get: Function to retrieve other components
//   - vpr: Function to access Viper configuration
//   - vrs: Version information
//   - log: Function to access the logger
func (o *mod) Init(key string, ctx context.Context, get cfgtps.FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog) {
	o.x.Store(keyCptKey, key)
	o.x.Store(keyFctGetCpt, get)
	o.x.Store(keyFctViper, vpr)
	o.x.Store(keyCptVersion, vrs)
	o.x.Store(keyCptLogger, log)
}

// RegisterFuncStart registers callbacks to be executed before and after component start.
// This implements the cfgtps.Component interface.
//
// Parameters:
//   - before: Function called before starting the component (can be nil)
//   - after: Function called after starting the component (can be nil)
func (o *mod) RegisterFuncStart(before, after cfgtps.FuncCptEvent) {
	o.x.Store(keyFctStaBef, before)
	o.x.Store(keyFctStaAft, after)
}

// RegisterFuncReload registers callbacks to be executed before and after component reload.
// This implements the cfgtps.Component interface.
//
// Parameters:
//   - before: Function called before reloading the component (can be nil)
//   - after: Function called after reloading the component (can be nil)
func (o *mod) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {
	o.x.Store(keyFctRelBef, before)
	o.x.Store(keyFctRelAft, after)
}

// IsStarted returns true if the component has been successfully started.
// A component is considered started if it has a valid server pool, TLS configuration,
// and at least one HTTP handler.
// This implements the cfgtps.Component interface.
func (o *mod) IsStarted() bool {
	if o == nil || o.s == nil {
		return false
	} else {
		return o.s.Load() != nil && o._GetTLS() != nil && len(o._GetHandler()) > 0
	}
}

// IsRunning returns true if the component is started and its server pool is running.
// This implements the cfgtps.Component interface.
func (o *mod) IsRunning() bool {
	if !o.IsStarted() {
		return false
	}

	if p := o.GetPool(); p == nil {
		return false
	} else {
		return p.IsRunning()
	}
}

// Start starts the HTTP component and all configured servers.
// This implements the cfgtps.Component interface.
//
// Returns an error if:
//   - The component is not properly initialized
//   - The configuration is invalid
//   - The TLS component is not available
//   - No handlers are configured
//   - Server startup fails
func (o *mod) Start() error {
	return o._run()
}

// Reload reloads the component configuration and restarts all servers with new settings.
// This implements the cfgtps.Component interface.
//
// The reload process:
//  1. Executes before-reload callbacks
//  2. Loads new configuration
//  3. Merges or replaces server pool
//  4. Restarts servers
//  5. Updates monitoring
//  6. Executes after-reload callbacks
func (o *mod) Reload() error {
	return o._run()
}

// Stop stops all HTTP servers and releases resources.
// This implements the cfgtps.Component interface.
//
// This method is safe to call multiple times and on nil components.
func (o *mod) Stop() {
	if o == nil {
		return
	} else if p := o.GetPool(); p == nil {
		return
	} else {
		_ = p.Stop(o.x.GetContext())
		o.SetPool(nil)
	}
}

// Dependencies returns the list of component keys that this component depends on.
// By default, returns the TLS component key. Can be overridden with SetDependencies.
// This implements the cfgtps.Component interface.
func (o *mod) Dependencies() []string {
	var def = []string{cpttls.ComponentType}

	if o == nil {
		return def
	} else if t := o.t.Load(); len(t) > 0 {
		def = []string{t}
	}

	if o.x == nil {
		return def
	} else if i, l := o.x.Load(keyCptDependencies); !l {
		return def
	} else if v, k := i.([]string); !k {
		return def
	} else if len(v) > 0 {
		return v
	} else {
		return def
	}
}

// SetDependencies sets custom dependencies for this component.
// This implements the cfgtps.Component interface.
//
// Parameters:
//   - d: Slice of component keys this component depends on
//
// If empty, the component reverts to default dependencies (TLS component).
func (o *mod) SetDependencies(d []string) error {
	if o == nil || o.x == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else {
		o.x.Store(keyCptDependencies, d)
		return nil
	}
}

func (o *mod) getLogger() liblog.Logger {
	if i, l := o.x.Load(keyCptLogger); !l {
		return nil
	} else if v, k := i.(liblog.FuncLog); !k {
		return nil
	} else {
		return v()
	}
}
