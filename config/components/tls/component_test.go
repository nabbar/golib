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

package tls_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/tls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Component lifecycle tests verify Init, Start, Reload, Stop behavior
// for the TLS component.
var _ = Describe("Component Lifecycle", func() {
	var (
		ctx context.Context
		cpt CptTlS
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil)
	})

	Describe("Type method", func() {
		Context("component type", func() {
			It("should return correct component type", func() {
				Expect(cpt.Type()).To(Equal(ComponentType))
			})

			It("should return consistent type", func() {
				type1 := cpt.Type()
				type2 := cpt.Type()
				Expect(type1).To(Equal(type2))
			})

			It("should match expected value", func() {
				Expect(cpt.Type()).To(Equal("tls"))
			})
		})
	})

	Describe("Init method", func() {
		var (
			key    string
			getCpt cfgtps.FuncCptGet
			vpr    libvpr.FuncViper
			vrs    libver.Version
			log    liblog.FuncLog
		)

		BeforeEach(func() {
			key = "tls-config"
			getCpt = func(k string) cfgtps.Component { return nil }
			vpr = func() libvpr.Viper { return nil }
			vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
			log = func() liblog.Logger { return nil }
		})

		Context("initializing component", func() {
			It("should accept valid initialization parameters", func() {
				Expect(func() {
					cpt.Init(key, ctx, getCpt, vpr, vrs, log)
				}).NotTo(Panic())
			})

			It("should initialize with empty key", func() {
				Expect(func() {
					cpt.Init("", ctx, getCpt, vpr, vrs, log)
				}).NotTo(Panic())
			})

			It("should initialize with nil getCpt", func() {
				Expect(func() {
					cpt.Init(key, ctx, nil, vpr, vrs, log)
				}).NotTo(Panic())
			})

			It("should initialize with nil viper", func() {
				Expect(func() {
					cpt.Init(key, ctx, getCpt, nil, vrs, log)
				}).NotTo(Panic())
			})

			It("should initialize with nil version", func() {
				Expect(func() {
					cpt.Init(key, ctx, getCpt, vpr, nil, log)
				}).NotTo(Panic())
			})

			It("should initialize with nil logger", func() {
				Expect(func() {
					cpt.Init(key, ctx, getCpt, vpr, vrs, nil)
				}).NotTo(Panic())
			})
		})

		Context("multiple initializations", func() {
			It("should allow re-initialization", func() {
				cpt.Init(key, ctx, getCpt, vpr, vrs, log)
				cpt.Init("new-key", ctx, getCpt, vpr, vrs, log)
				// Should not panic
			})
		})
	})

	Describe("RegisterFuncStart method", func() {
		Context("registering start callbacks", func() {
			It("should accept before and after callbacks", func() {
				before := func(c cfgtps.Component) error { return nil }
				after := func(c cfgtps.Component) error { return nil }

				Expect(func() {
					cpt.RegisterFuncStart(before, after)
				}).NotTo(Panic())
			})

			It("should accept nil callbacks", func() {
				Expect(func() {
					cpt.RegisterFuncStart(nil, nil)
				}).NotTo(Panic())
			})

			It("should accept only before callback", func() {
				before := func(c cfgtps.Component) error { return nil }
				Expect(func() {
					cpt.RegisterFuncStart(before, nil)
				}).NotTo(Panic())
			})

			It("should accept only after callback", func() {
				after := func(c cfgtps.Component) error { return nil }
				Expect(func() {
					cpt.RegisterFuncStart(nil, after)
				}).NotTo(Panic())
			})
		})
	})

	Describe("RegisterFuncReload method", func() {
		Context("registering reload callbacks", func() {
			It("should accept before and after callbacks", func() {
				before := func(c cfgtps.Component) error { return nil }
				after := func(c cfgtps.Component) error { return nil }

				Expect(func() {
					cpt.RegisterFuncReload(before, after)
				}).NotTo(Panic())
			})

			It("should accept nil callbacks", func() {
				Expect(func() {
					cpt.RegisterFuncReload(nil, nil)
				}).NotTo(Panic())
			})

			It("should accept only before callback", func() {
				before := func(c cfgtps.Component) error { return nil }
				Expect(func() {
					cpt.RegisterFuncReload(before, nil)
				}).NotTo(Panic())
			})

			It("should accept only after callback", func() {
				after := func(c cfgtps.Component) error { return nil }
				Expect(func() {
					cpt.RegisterFuncReload(nil, after)
				}).NotTo(Panic())
			})
		})
	})

	Describe("IsStarted method", func() {
		Context("initial state", func() {
			It("should not be started initially", func() {
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should remain not started without Start call", func() {
				// Without proper initialization and config, should not be started
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

	})

	Describe("IsRunning method", func() {
		Context("initial state", func() {
			It("should not be running initially", func() {
				Expect(cpt.IsRunning()).To(BeFalse())
			})

			It("should return false when not started", func() {
				Expect(cpt.IsRunning()).To(BeFalse())
			})
		})

		Context("dependency on IsStarted", func() {
			It("should match IsStarted", func() {
				// Component is not started, so IsRunning should be false
				Expect(cpt.IsRunning()).To(Equal(cpt.IsStarted()))
			})
		})
	})

	Describe("Start method", func() {
		Context("starting component", func() {
			It("should return error when not initialized", func() {
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})

			It("should return error without proper configuration", func() {
				key := "tls-config"
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Reload method", func() {
		Context("reloading component", func() {
			It("should return error when not initialized", func() {
				err := cpt.Reload()
				Expect(err).To(HaveOccurred())
			})

			It("should return error without proper configuration", func() {
				key := "tls-config"
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				err := cpt.Reload()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Stop method", func() {
		Context("stopping component", func() {
			It("should not panic on uninitialized component", func() {
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})

			It("should not panic on already stopped component", func() {
				cpt.Stop()
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})

			It("should clear TLS configuration", func() {
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})
	})

	Describe("Dependencies method", func() {
		Context("default dependencies", func() {
			It("should return empty dependencies by default", func() {
				deps := cpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})

			It("should return consistent dependencies", func() {
				deps1 := cpt.Dependencies()
				deps2 := cpt.Dependencies()
				Expect(deps1).To(Equal(deps2))
			})
		})

		Context("custom dependencies", func() {
			It("should allow setting custom dependencies", func() {
				customDeps := []string{"dep1", "dep2"}
				err := cpt.SetDependencies(customDeps)
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(Equal(customDeps))
			})

			It("should replace default dependencies", func() {
				customDeps := []string{"custom-dep"}
				cpt.SetDependencies(customDeps)

				deps := cpt.Dependencies()
				Expect(deps).To(Equal(customDeps))
			})

			It("should handle empty dependencies", func() {
				emptyDeps := []string{}
				err := cpt.SetDependencies(emptyDeps)
				Expect(err).To(BeNil())

				// When empty, should return empty
				deps := cpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})
		})
	})

	Describe("SetDependencies method", func() {
		Context("setting dependencies", func() {
			It("should accept valid dependencies", func() {
				deps := []string{"dep1", "dep2", "dep3"}
				err := cpt.SetDependencies(deps)
				Expect(err).To(BeNil())
			})

			It("should accept empty slice", func() {
				err := cpt.SetDependencies([]string{})
				Expect(err).To(BeNil())
			})

			It("should accept nil slice", func() {
				err := cpt.SetDependencies(nil)
				Expect(err).To(BeNil())
			})

			It("should allow multiple updates", func() {
				err1 := cpt.SetDependencies([]string{"dep1"})
				err2 := cpt.SetDependencies([]string{"dep2"})
				err3 := cpt.SetDependencies([]string{"dep3"})

				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
				Expect(err3).To(BeNil())
			})
		})
	})

	Describe("Component lifecycle flow", func() {
		Context("typical lifecycle", func() {
			It("should follow expected state transitions", func() {
				// Initial state
				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())

				// After Init (without proper config, won't start)
				key := "tls-config"
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				vrs := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				// Still not started
				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())

				// Stop should not panic
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())
			})
		})
	})

	Describe("Concurrent operations", func() {
		Context("concurrent lifecycle calls", func() {
			It("should handle concurrent IsStarted calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.IsStarted()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent IsRunning calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.IsRunning()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent Dependencies calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.Dependencies()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})
})
