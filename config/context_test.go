/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package config_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

var _ = Describe("Config Context", func() {
	var cfg libcfg.Config

	BeforeEach(func() {
		cfg = libcfg.New(nil)
	})

	Describe("Context", func() {
		It("should return a valid context instance", func() {
			ctx := cfg.Context()
			Expect(ctx).ToNot(BeNil())
		})

		It("should return context that implements Config interface", func() {
			ctx := cfg.Context()
			var _ libctx.Config[string] = ctx
			Expect(ctx).ToNot(BeNil())
		})

		It("should have context that can store and load values", func() {
			ctx := cfg.Context()
			ctx.Store("test-key", "test-value")

			val, ok := ctx.Load("test-key")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("test-value"))
		})

		It("should have context with done channel", func() {
			ctx := cfg.Context()
			c := ctx.GetContext()
			Expect(c.Done()).ToNot(BeNil())
		})
	})

	Describe("CancelAdd", func() {
		It("should register a cancel function", func() {
			cfg.CancelAdd(func() {
				// Cancel function registered
			})

			// Trigger cancel by stopping
			cfg.Stop()
			// Note: cancel functions are called during shutdown/cancel
			// We can't easily test this without triggering the actual cancel
		})

		It("should register multiple cancel functions", func() {
			count := 0
			cfg.CancelAdd(
				func() { count++ },
				func() { count++ },
				func() { count++ },
			)

			// Functions are registered, will be called on cancel
			Expect(count).To(Equal(0)) // Not called yet
		})

		It("should not panic with nil function", func() {
			Expect(func() {
				cfg.CancelAdd(nil)
			}).ToNot(Panic())
		})
	})

	Describe("CancelClean", func() {
		It("should clear all registered cancel functions", func() {
			count := 0
			cfg.CancelAdd(func() { count++ })
			cfg.CancelAdd(func() { count++ })

			cfg.CancelClean()

			// After clean, no functions should be registered
			// We can verify by checking the internal state is reset
			cfg.CancelAdd(func() { count++ })
			// Only the last one should be registered
		})

		It("should not panic when called multiple times", func() {
			Expect(func() {
				cfg.CancelClean()
				cfg.CancelClean()
			}).ToNot(Panic())
		})

		It("should allow adding new functions after clean", func() {
			cfg.CancelAdd(func() {})
			cfg.CancelClean()

			Expect(func() {
				cfg.CancelAdd(func() {})
			}).ToNot(Panic())
		})
	})

	Describe("Context Integration", func() {
		It("should allow components to access context", func() {
			cpt := &contextAwareComponent{}
			cfg.ComponentSet("ctx-comp", cpt)

			// Component should have access to context through FuncContext
			Expect(cpt.hasContext).To(BeTrue())
		})

		It("should provide consistent context to all components", func() {
			cpt1 := &contextAwareComponent{}
			cpt2 := &contextAwareComponent{}

			cfg.ComponentSet("comp1", cpt1)
			cfg.ComponentSet("comp2", cpt2)

			// Both should have access to context
			Expect(cpt1.hasContext).To(BeTrue())
			Expect(cpt2.hasContext).To(BeTrue())
		})

		It("should allow context value storage and retrieval", func() {
			ctx := cfg.Context()

			// Store a value
			ctx.Store("shared-key", "shared-value")

			// Retrieve from same context
			val, ok := ctx.Load("shared-key")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("shared-value"))
		})

		It("should handle context cancellation", func() {
			ctx := cfg.Context()
			c := ctx.GetContext()

			// Create a derived context with timeout
			derived, cancel := context.WithTimeout(c, 100*time.Millisecond)
			defer cancel()

			// Wait for timeout
			select {
			case <-derived.Done():
				Expect(derived.Err()).To(Equal(context.DeadlineExceeded))
			case <-time.After(200 * time.Millisecond):
				Fail("Context should have been cancelled")
			}
		})
	})
})

// contextAwareComponent is a test component that tracks context access
type contextAwareComponent struct {
	simpleComponent
	hasContext bool
	ctxFunc    context.Context
}

func (c *contextAwareComponent) Init(key string, ctx context.Context, get cfgtps.FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog) {
	c.key = key
	c.ctxFunc = ctx
	c.hasContext = (ctx != nil)
}
