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

package status_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/status"
	montps "github.com/nabbar/golib/monitor/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

type Empty struct{}

// Component lifecycle tests verify Init, Start, Reload, and Stop operations.
var _ = Describe("Component Lifecycle", func() {
	var (
		ctx context.Context
		cfg libcfg.Config
		cpt CptStatus
		key string
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = libcfg.New(nil)
		cpt = New(ctx)
		key = "test-status"
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
		if cfg != nil {
			cfg.Stop()
		}
	})

	Describe("Type", func() {
		It("should return the correct component type 'status'", func() {
			Expect(cpt.Type()).To(Equal("status"))
		})

		It("should return a consistent type on multiple calls", func() {
			type1 := cpt.Type()
			type2 := cpt.Type()
			Expect(type1).To(Equal(type2))
		})
	})

	Describe("Init", func() {
		Context("with valid parameters", func() {
			It("should initialize successfully without panicking", func() {
				log := func() liblog.Logger { return nil }
				vpr := func() libvpr.Viper { return libvpr.New(ctx, log) }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				Expect(func() {
					cpt.Init(key, ctx, get, vpr, ver, log)
				}).NotTo(Panic())
			})
		})

		Context("with nil parameters", func() {
			It("should handle a nil viper function without panicking", func() {
				log := func() liblog.Logger { return nil }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				Expect(func() {
					cpt.Init(key, ctx, get, nil, ver, log)
				}).NotTo(Panic())
			})

			It("should handle a nil logger function without panicking", func() {
				log := func() liblog.Logger { return nil }
				vpr := func() libvpr.Viper { return libvpr.New(ctx, log) }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				Expect(func() {
					cpt.Init(key, ctx, get, vpr, ver, nil)
				}).NotTo(Panic())
			})
		})

		Context("with an empty key", func() {
			It("should handle an empty key without panicking", func() {
				log := func() liblog.Logger { return nil }
				vpr := func() libvpr.Viper { return libvpr.New(ctx, log) }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				Expect(func() {
					cpt.Init("", ctx, get, vpr, ver, log)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Dependencies", func() {
		Context("when no dependencies are set", func() {
			It("should return an empty list", func() {
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeNil())
				Expect(deps).To(BeEmpty())
			})
		})

		Context("when setting dependencies", func() {
			It("should store and return the dependencies successfully", func() {
				deps := []string{"dep1", "dep2", "dep3"}
				err := cpt.SetDependencies(deps)
				Expect(err).To(BeNil())

				result := cpt.Dependencies()
				Expect(result).To(Equal(deps))
			})

			It("should override previous dependencies", func() {
				deps1 := []string{"dep1", "dep2"}
				cpt.SetDependencies(deps1)

				deps2 := []string{"dep3", "dep4", "dep5"}
				cpt.SetDependencies(deps2)

				result := cpt.Dependencies()
				Expect(result).To(Equal(deps2))
			})
		})
	})

	Describe("RegisterFlag", func() {
		It("should execute without error for a valid command", func() {
			cmd := &spfcbr.Command{
				Use: "test",
			}
			err := cpt.RegisterFlag(cmd)
			Expect(err).To(BeNil())
		})

		It("should handle a nil command without error", func() {
			err := cpt.RegisterFlag(nil)
			Expect(err).To(BeNil())
		})
	})

	Describe("RegisterFuncStart", func() {
		It("should register start callbacks without panicking", func() {
			before := func(c cfgtps.Component) error { return nil }
			after := func(c cfgtps.Component) error { return nil }

			Expect(func() {
				cpt.RegisterFuncStart(before, after)
			}).NotTo(Panic())
		})
	})

	Describe("RegisterFuncReload", func() {
		It("should register reload callbacks without panicking", func() {
			before := func(c cfgtps.Component) error { return nil }
			after := func(c cfgtps.Component) error { return nil }

			Expect(func() {
				cpt.RegisterFuncReload(before, after)
			}).NotTo(Panic())
		})
	})

	Describe("IsStarted / IsRunning", func() {
		Context("before the component is started", func() {
			It("should return false", func() {
				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())
			})
		})
	})

	Describe("Stop", func() {
		Context("when the component is running", func() {
			It("should stop the component and update its state", func() {
				// Setup configuration to allow starting
				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]interface{}{"return-code": map[string]int{"OK": 200}})
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, nil, nil)

				err := cpt.Start()
				Expect(err).NotTo(HaveOccurred())
				Expect(cpt.IsStarted()).To(BeTrue())

				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

		Context("when the component is not started", func() {
			It("should not panic", func() {
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})
		})

		Context("when Stop is called multiple times", func() {
			It("should handle multiple stop calls gracefully", func() {
				cpt.Start()
				Expect(func() {
					cpt.Stop()
					cpt.Stop()
					cpt.Stop()
				}).NotTo(Panic())
			})
		})
	})

	Describe("RegisterMonitorPool", func() {
		It("should not panic with a nil function", func() {
			Expect(func() {
				cpt.RegisterMonitorPool(nil)
			}).NotTo(Panic())
		})

		It("should accept a valid monitor pool function", func() {
			monitorFunc := func() montps.Pool { return nil }
			Expect(func() {
				cpt.RegisterMonitorPool(monitorFunc)
			}).NotTo(Panic())
		})
	})
})

// Start and Reload tests with configuration.
var _ = Describe("Start and Reload", func() {
	var (
		ctx context.Context
		cfg libcfg.Config
		vpr libvpr.Viper
		cpt CptStatus
		key string
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = libcfg.New(nil)
		key = "test-status"

		// Create viper with test configuration
		log := func() liblog.Logger { return nil }
		vpr = libvpr.New(ctx, log)
		vpr.Viper().Set(key, map[string]interface{}{
			"return-code": map[string]int{
				"OK": 200,
			},
		})

		cpt = New(ctx)
		cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0), log)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
		if cfg != nil {
			cfg.Stop()
		}
	})

	Describe("Start", func() {
		Context("with valid configuration", func() {
			It("should start successfully and be in a running state", func() {
				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(cpt.IsStarted()).To(BeTrue())
			})
		})

		Context("without configuration", func() {
			It("should return an error if the config key is missing", func() {
				log := func() liblog.Logger { return nil }
				vpr2 := libvpr.New(ctx, log) // Fresh viper without the key
				cpt2 := New(ctx)
				cpt2.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr2 }, libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0), log)

				err := cpt2.Start()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with callbacks", func() {
			It("should call 'Before' and 'After' start callbacks in order", func() {
				beforeCalled := false
				afterCalled := false

				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						beforeCalled = true
						Expect(afterCalled).To(BeFalse()) // Ensure After is not yet called
						return nil
					},
					func(c cfgtps.Component) error {
						afterCalled = true
						Expect(beforeCalled).To(BeTrue()) // Ensure Before was called
						return nil
					},
				)

				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())
			})

			It("should prevent starting if the 'Before' callback returns an error", func() {
				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						return ErrorParamInvalid.Error(nil)
					},
					nil,
				)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})
	})

	Describe("Reload", func() {
		Context("when the component is already started", func() {
			It("should reload the configuration successfully", func() {
				err := cpt.Start()
				Expect(err).To(BeNil())

				// Update config
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 201,
					},
				})

				err = cpt.Reload()
				Expect(err).To(BeNil())
			})

			It("should call 'Before' and 'After' reload callbacks", func() {
				err := cpt.Start()
				Expect(err).To(BeNil())

				beforeCalled := false
				afterCalled := false

				cpt.RegisterFuncReload(
					func(c cfgtps.Component) error {
						beforeCalled = true
						return nil
					},
					func(c cfgtps.Component) error {
						afterCalled = true
						return nil
					},
				)

				err = cpt.Reload()
				Expect(err).To(BeNil())
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())
			})
		})

		Context("when the component is not started", func() {
			It("should effectively start the component and call start callbacks", func() {
				startBeforeCalled := false
				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						startBeforeCalled = true
						return nil
					},
					nil,
				)

				err := cpt.Reload()
				Expect(err).To(BeNil())
				Expect(startBeforeCalled).To(BeTrue())
				Expect(cpt.IsStarted()).To(BeTrue())
			})
		})
	})
})
