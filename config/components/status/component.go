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
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package status

import (
	"context"

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

const (
	// ComponentType is the unique identifier for this component type.
	ComponentType = "status"

	// keyCptKey is the context key for storing the component's configuration key.
	keyCptKey = iota + 1
	// keyCptDependencies is the context key for storing the component's dependencies.
	keyCptDependencies
	// keyFctViper is the context key for storing the Viper factory function.
	keyFctViper
	// keyFctGetCpt is the context key for storing the function to get other components.
	keyFctGetCpt
	// keyCptVersion is the context key for storing the application version info.
	keyCptVersion
	// keyCptLogger is the context key for storing the logger factory function.
	keyCptLogger
	// keyFctStaBef is the context key for the 'Before Start' callback.
	keyFctStaBef
	// keyFctStaAft is the context key for the 'After Start' callback.
	keyFctStaAft
	// keyFctRelBef is the context key for the 'Before Reload' callback.
	keyFctRelBef
	// keyFctRelAft is the context key for the 'After Reload' callback.
	keyFctRelAft
)

// Type returns the component's type identifier.
func (o *mod) Type() string {
	return ComponentType
}

// Init initializes the component with its key, context, and factory functions for dependencies.
func (o *mod) Init(key string, ctx context.Context, get cfgtps.FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog) {
	o.x.Store(keyCptKey, key)
	o.x.Store(keyFctGetCpt, get)
	o.x.Store(keyFctViper, vpr)
	o.x.Store(keyCptVersion, vrs)
	o.x.Store(keyCptLogger, log)

	o.s.RegisterGetConfigCpt(func(key string) cfgtps.ComponentMonitor {
		if get != nil {
			return get(key)
		}
		return nil
	})

	o.s.SetVersion(vrs)
}

// RegisterFuncStart registers the 'Before' and 'After' callbacks for the Start event.
func (o *mod) RegisterFuncStart(before, after cfgtps.FuncCptEvent) {
	o.x.Store(keyFctStaBef, before)
	o.x.Store(keyFctStaAft, after)
}

// RegisterFuncReload registers the 'Before' and 'After' callbacks for the Reload event.
func (o *mod) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {
	o.x.Store(keyFctRelBef, before)
	o.x.Store(keyFctRelAft, after)
}

// IsStarted checks if the component is currently running.
func (o *mod) IsStarted() bool {
	return o.r.Load()
}

// IsRunning is an alias for IsStarted, fulfilling the component interface.
func (o *mod) IsRunning() bool {
	return o.IsStarted()
}

// Start triggers the component's start sequence, which involves loading the configuration.
func (o *mod) Start() error {
	return o.run()
}

// Reload triggers the component's reload sequence, which involves reloading the configuration.
func (o *mod) Reload() error {
	return o.run()
}

// Stop marks the component as stopped.
func (o *mod) Stop() {
	o.r.Store(false)
}

// Dependencies returns a list of keys for components that this component depends on.
func (o *mod) Dependencies() []string {
	var def = make([]string, 0)

	if o == nil {
		return def
	} else if o.x == nil {
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

// SetDependencies sets the list of keys for components that this component depends on.
func (o *mod) SetDependencies(d []string) error {
	if o == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else {
		o.x.Store(keyCptDependencies, d)
		return nil
	}
}
