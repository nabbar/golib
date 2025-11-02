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

package tls

import (
	"context"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	tlscas "github.com/nabbar/golib/certificates/ca"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
)

// CptTlS is the public interface of the TLS component.
//
// It exposes the common component lifecycle from `cfgtps.Component` and
// provides accessors to the underlying TLS settings constructed from the
// configuration system.
type CptTlS interface {
	cfgtps.Component
	Config() *libtls.Config
	GetTLS() libtls.TLSConfig
	SetTLS(tls libtls.TLSConfig)
}

// GetRootCaCert converts a root CA provider function into a consolidated
// certificate chain of type `tlscas.Cert`.
//
// The provided function typically returns a slice of PEM-encoded CA strings.
// This helper parses and appends them into a single certificate chain object.
// Invalid entries are ignored by the underlying parser implementation.
func GetRootCaCert(fct libtls.FctRootCA) tlscas.Cert {
	var res tlscas.Cert

	for _, c := range fct() {
		if res == nil {
			res, _ = tlscas.Parse(c)
		} else {
			_ = res.AppendString(c)
		}
	}

	return res
}

// New creates a new TLS component instance.
//
// The component relies on a context provider and optionally a default root CA
// provider. The returned component is not started; you must call `Init` then
// `Start` to build the runtime TLS configuration from Viper.
func New(ctx context.Context, defCARoot libtls.FctRootCACert) CptTlS {
	c := &mod{
		x: libctx.New[uint8](ctx),
		t: libatm.NewValue[libtls.TLSConfig](),
		c: libatm.NewValue[func() *libtls.Config](),
		f: defCARoot,
		r: new(atomic.Bool),
	}
	c.r.Store(false)
	return c
}

// Register registers the given TLS component in the provided global
// configuration registry under the specified key.
func Register(cfg libcfg.Config, key string, cpt CptTlS) {
	cfg.ComponentSet(key, cpt)
}

// RegisterNew instantiates a new TLS component and registers it in the
// configuration registry under the provided key.
func RegisterNew(ctx context.Context, cfg libcfg.Config, key string, defCARoot libtls.FctRootCACert) {
	cfg.ComponentSet(key, New(ctx, defCARoot))
}

// Load retrieves a TLS component from a component getter.
//
// It returns nil when the key is not found or the component under the key
// does not implement `CptTlS`.
func Load(getCpt cfgtps.FuncCptGet, key string) CptTlS {
	if getCpt == nil {
		return nil
	} else if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(CptTlS); !ok {
		return nil
	} else {
		return h
	}
}
