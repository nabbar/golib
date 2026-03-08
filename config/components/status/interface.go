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

// Package status provides a configuration component that wraps the health check
// and status monitoring functionalities of the `golib/status` package.
//
// It allows managing status monitoring as a standard component within the `golib/config` framework,
// including lifecycle management (Init, Start, Reload, Stop) and configuration via Viper.
package status

import (
	"context"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
	libsts "github.com/nabbar/golib/status"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
)

// CptStatus defines the interface for the status component.
// It embeds the standard component interface and the status interface,
// combining lifecycle management with health check functionalities.
type CptStatus interface {
	cfgtps.Component
	libsts.Status
}

// New creates a new instance of the status component.
// It initializes the internal state and the underlying status object.
func New(ctx context.Context) CptStatus {
	return &mod{
		x: libctx.New[uint8](ctx),
		s: libsts.New(ctx),
		r: new(atomic.Bool),
	}
}

// Register is a helper function to register a status component instance
// with a given configuration manager.
func Register(cfg libcfg.Config, key string, cpt CptStatus) {
	cfg.ComponentSet(key, cpt)
}

// RegisterNew is a helper function to create and register a new status component
// in a single step.
func RegisterNew(ctx context.Context, cfg libcfg.Config, key string) {
	cfg.ComponentSet(key, New(ctx))
}

// Load is a helper function to safely retrieve a status component from a
// component getter function. It performs a type assertion and returns nil if
// the component is not found or has the wrong type.
func Load(getCpt cfgtps.FuncCptGet, key string) CptStatus {
	if getCpt == nil {
		return nil
	} else if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(CptStatus); !ok {
		return nil
	} else {
		return h
	}
}
