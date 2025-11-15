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

package head_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/head"
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

// Component lifecycle tests verify Init, Start, Reload, Stop operations
var _ = Describe("Component Lifecycle", func() {
	var (
		ctx context.Context
		cfg libcfg.Config
		cpt CptHead
		key string
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = libcfg.New(nil)
		cpt = New(ctx)
		key = "test-head"
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
		It("should return correct component type", func() {
			Expect(cpt.Type()).To(Equal("head"))
		})

		It("should return consistent type", func() {
			type1 := cpt.Type()
			type2 := cpt.Type()
			Expect(type1).To(Equal(type2))
		})
	})

	Describe("Init", func() {
		Context("with valid parameters", func() {
			It("should initialize successfully", func() {
				log := func() liblog.Logger { return nil }
				vpr := func() libvpr.Viper { return libvpr.New(ctx, log) }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				Expect(func() {
					cpt.Init(key, ctx, get, vpr, ver, log)
				}).NotTo(Panic())
			})

			It("should store component key", func() {
				log := func() liblog.Logger { return nil }
				vpr := func() libvpr.Viper { return libvpr.New(ctx, log) }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				cpt.Init(key, ctx, get, vpr, ver, log)
				// Type should still be accessible
				Expect(cpt.Type()).To(Equal("head"))
			})
		})

		Context("with nil parameters", func() {
			It("should handle nil viper function", func() {
				log := func() liblog.Logger { return nil }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				Expect(func() {
					cpt.Init(key, ctx, get, nil, ver, log)
				}).NotTo(Panic())
			})

			It("should handle nil logger function", func() {
				log := func() liblog.Logger { return nil }
				vpr := func() libvpr.Viper { return libvpr.New(ctx, log) }
				get := func(string) cfgtps.Component { return nil }
				ver := libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)

				Expect(func() {
					cpt.Init(key, ctx, get, vpr, ver, nil)
				}).NotTo(Panic())
			})
		})

		Context("with empty key", func() {
			It("should handle empty key", func() {
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
		Context("when no dependencies set", func() {
			It("should return empty list", func() {
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeNil())
				Expect(deps).To(BeEmpty())
			})
		})

		Context("when setting dependencies", func() {
			It("should set dependencies successfully", func() {
				deps := []string{"dep1", "dep2", "dep3"}
				err := cpt.SetDependencies(deps)
				Expect(err).To(BeNil())

				result := cpt.Dependencies()
				Expect(result).To(Equal(deps))
			})

			It("should handle empty dependencies", func() {
				deps := []string{}
				err := cpt.SetDependencies(deps)
				Expect(err).To(BeNil())

				result := cpt.Dependencies()
				Expect(result).To(BeEmpty())
			})

			It("should handle nil dependencies", func() {
				err := cpt.SetDependencies(nil)
				Expect(err).To(BeNil())
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
		It("should execute without error", func() {
			cmd := &spfcbr.Command{
				Use: "test",
			}
			err := cpt.RegisterFlag(cmd)
			Expect(err).To(BeNil())
		})

		It("should handle nil command", func() {
			err := cpt.RegisterFlag(nil)
			Expect(err).To(BeNil())
		})
	})

	Describe("RegisterFuncStart", func() {
		It("should register start callbacks", func() {
			before := func(c cfgtps.Component) error {
				return nil
			}
			after := func(c cfgtps.Component) error {
				return nil
			}

			Expect(func() {
				cpt.RegisterFuncStart(before, after)
			}).NotTo(Panic())
		})

		It("should handle nil callbacks", func() {
			Expect(func() {
				cpt.RegisterFuncStart(nil, nil)
			}).NotTo(Panic())
		})
	})

	Describe("RegisterFuncReload", func() {
		It("should register reload callbacks", func() {
			before := func(c cfgtps.Component) error {
				return nil
			}
			after := func(c cfgtps.Component) error {
				return nil
			}

			Expect(func() {
				cpt.RegisterFuncReload(before, after)
			}).NotTo(Panic())
		})

		It("should handle nil callbacks", func() {
			Expect(func() {
				cpt.RegisterFuncReload(nil, nil)
			}).NotTo(Panic())
		})
	})

	Describe("IsStarted", func() {
		Context("before start", func() {
			It("should return false", func() {
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

		Context("with headers set", func() {
			It("should return true when headers are set", func() {
				headers := cpt.GetHeaders()
				headers.Set("X-Test", "value")
				cpt.SetHeaders(headers)

				Expect(cpt.IsStarted()).To(BeTrue())
			})

			It("should return false with empty headers", func() {
				cpt.SetHeaders(nil)
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})
	})

	Describe("IsRunning", func() {
		It("should match IsStarted", func() {
			started := cpt.IsStarted()
			running := cpt.IsRunning()
			Expect(started).To(Equal(running))
		})

		It("should reflect component state", func() {
			Expect(cpt.IsRunning()).To(BeFalse())

			headers := cpt.GetHeaders()
			headers.Set("X-Test", "value")
			cpt.SetHeaders(headers)

			Expect(cpt.IsRunning()).To(BeTrue())
		})
	})

	Describe("Stop", func() {
		Context("when component is running", func() {
			It("should clear headers", func() {
				headers := cpt.GetHeaders()
				headers.Set("X-Test", "value")
				cpt.SetHeaders(headers)

				Expect(cpt.IsStarted()).To(BeTrue())

				cpt.Stop()

				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

		Context("when component is not started", func() {
			It("should not panic", func() {
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})
		})

		Context("multiple stops", func() {
			It("should handle multiple stop calls", func() {
				headers := cpt.GetHeaders()
				headers.Set("X-Test", "value")
				cpt.SetHeaders(headers)

				Expect(func() {
					cpt.Stop()
					cpt.Stop()
					cpt.Stop()
				}).NotTo(Panic())
			})
		})
	})

	Describe("RegisterMonitorPool", func() {
		It("should not panic with nil function", func() {
			Expect(func() {
				cpt.RegisterMonitorPool(nil)
			}).NotTo(Panic())
		})

		It("should accept monitor function", func() {
			monitorFunc := func() montps.Pool { return nil }

			Expect(func() {
				cpt.RegisterMonitorPool(monitorFunc)
			}).NotTo(Panic())
		})
	})
})

// Start and Reload tests with configuration
var _ = Describe("Start and Reload", func() {
	var (
		ctx context.Context
		cfg libcfg.Config
		vpr libvpr.Viper
		cpt CptHead
		key string
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = libcfg.New(nil)
		key = "test-head"

		// Create viper with test configuration
		log := func() liblog.Logger { return nil }
		vpr = libvpr.New(ctx, log)
		vpr.Viper().Set(key, map[string]string{
			"X-Frame-Options":        "DENY",
			"X-Content-Type-Options": "nosniff",
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
			It("should start successfully", func() {
				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(cpt.IsStarted()).To(BeTrue())
			})

			It("should load headers from config", func() {
				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Frame-Options")).To(Equal("DENY"))
				Expect(headers.Get("X-Content-Type-Options")).To(Equal("nosniff"))
			})
		})

		Context("without configuration", func() {
			It("should return error when config key is missing", func() {
				log := func() liblog.Logger { return nil }

				// Don't set any config for key
				vpr2 := libvpr.New(ctx, log)

				cpt2 := New(ctx)
				cpt2.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr2 }, libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0), log)

				err := cpt2.Start()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with callbacks", func() {
			It("should call before and after start callbacks", func() {
				beforeCalled := false
				afterCalled := false

				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						beforeCalled = true
						Expect(c).NotTo(BeNil())
						return nil
					},
					func(c cfgtps.Component) error {
						afterCalled = true
						Expect(c).NotTo(BeNil())
						return nil
					},
				)

				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())
			})

			It("should stop if before callback returns error", func() {
				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						return ErrorParamInvalid.Error(nil)
					},
					nil,
				)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Reload", func() {
		Context("when already started", func() {
			It("should reload configuration", func() {
				// Start first
				err := cpt.Start()
				Expect(err).To(BeNil())

				// Update config
				vpr.Viper().Set(key, map[string]string{
					"X-Frame-Options": "SAMEORIGIN",
					"X-New-Header":    "new-value",
				})

				// Reload
				err = cpt.Reload()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Frame-Options")).To(Equal("SAMEORIGIN"))
				Expect(headers.Get("X-New-Header")).To(Equal("new-value"))
			})

			It("should call reload callbacks", func() {
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

		Context("when not started", func() {
			It("should call start callbacks", func() {
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
			})
		})
	})
})
