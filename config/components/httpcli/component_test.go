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

// Component lifecycle tests
package httpcli_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/httpcli"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Component Lifecycle", func() {
	var (
		cpt CptHTTPClient
		ctx context.Context
		vrs libver.Version
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
		cpt = New(ctx, nil, false, nil)
	})

	Describe("Initialization", func() {
		It("should initialize component with key", func() {
			key := "test-httpcli"
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }

			cpt.Init(key, ctx, getCpt, vpr, vrs, log)
			Expect(cpt.IsStarted()).To(BeFalse())
		})
	})

	Describe("State management", func() {
		It("should return false before start", func() {
			Expect(cpt.IsStarted()).To(BeFalse())
			Expect(cpt.IsRunning()).To(BeFalse())
		})
	})

	Describe("Dependencies", func() {
		It("should return empty slice by default", func() {
			deps := cpt.Dependencies()
			Expect(deps).NotTo(BeNil())
			Expect(deps).To(BeEmpty())
		})

		It("should set and get dependencies", func() {
			key := "test-httpcli"
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }

			cpt.Init(key, ctx, getCpt, vpr, vrs, log)

			expectedDeps := []string{"dep1", "dep2"}
			err := cpt.SetDependencies(expectedDeps)
			Expect(err).To(BeNil())

			deps := cpt.Dependencies()
			Expect(deps).To(Equal(expectedDeps))
		})
	})

	Describe("Callbacks", func() {
		It("should register start callbacks", func() {
			before := func(c cfgtps.Component) error { return nil }
			after := func(c cfgtps.Component) error { return nil }

			Expect(func() {
				cpt.RegisterFuncStart(before, after)
			}).NotTo(Panic())
		})

		It("should register reload callbacks", func() {
			before := func(c cfgtps.Component) error { return nil }
			after := func(c cfgtps.Component) error { return nil }

			Expect(func() {
				cpt.RegisterFuncReload(before, after)
			}).NotTo(Panic())
		})
	})

	Describe("Stop method", func() {
		It("should not panic", func() {
			Expect(func() {
				cpt.Stop()
			}).NotTo(Panic())
		})
	})
})
