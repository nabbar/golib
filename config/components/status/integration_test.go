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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Integration tests verify end-to-end scenarios, including the full component lifecycle.
var _ = Describe("Integration Tests", func() {
	var (
		ctx context.Context
		cfg libcfg.Config
	)

	BeforeEach(func() {
		ctx = context.Background()
		cfg = libcfg.New(nil)
	})

	AfterEach(func() {
		if cfg != nil {
			cfg.Stop()
		}
	})

	Describe("Complete Component Lifecycle", func() {
		Context("with a standard workflow", func() {
			It("should handle the complete lifecycle: create, init, start, and stop", func() {
				// 1. Create
				cpt := New(ctx)
				Expect(cpt).NotTo(BeNil())

				// 2. Register
				key := "integration-status"
				Register(cfg, key, cpt)
				Expect(cfg.ComponentHas(key)).To(BeTrue())

				// 3. Setup configuration
				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 200,
					},
				})

				// 4. Initialize
				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// 5. Start
				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(cpt.IsStarted()).To(BeTrue())

				// 6. Stop
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should handle multiple start-stop cycles correctly", func() {
				cpt := New(ctx)
				key := "cycle-status"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 200,
					},
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// Cycle 1
				Expect(cpt.Start()).To(Succeed())
				Expect(cpt.IsStarted()).To(BeTrue())
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())

				// Cycle 2
				Expect(cpt.Start()).To(Succeed())
				Expect(cpt.IsStarted()).To(BeTrue())
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

		Context("with lifecycle callbacks", func() {
			It("should execute all callbacks in the correct order", func() {
				cpt := New(ctx)
				key := "callback-status"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 200,
					},
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				var callOrder []string

				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "start-before")
						return nil
					},
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "start-after")
						return nil
					},
				)

				cpt.RegisterFuncReload(
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "reload-before")
						return nil
					},
					func(c cfgtps.Component) error {
						callOrder = append(callOrder, "reload-after")
						return nil
					},
				)

				// Start
				err := cpt.Start()
				Expect(err).To(BeNil())
				Expect(callOrder).To(Equal([]string{"start-before", "start-after"}))

				// Reload
				callOrder = []string{} // Reset for reload
				err = cpt.Reload()
				Expect(err).To(BeNil())
				Expect(callOrder).To(Equal([]string{"reload-before", "reload-after"}))
			})

			It("should handle callback errors gracefully", func() {
				cpt := New(ctx)
				key := "error-callback-status"

				vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 200,
					},
				})

				cpt.Init(key, ctx,
					func(string) cfgtps.Component { return nil },
					func() libvpr.Viper { return vpr },
					libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
					func() liblog.Logger { return nil })

				// Register a 'Before Start' callback that returns an error
				cpt.RegisterFuncStart(
					func(c cfgtps.Component) error {
						return ErrorParamInvalid.Error(nil)
					},
					nil,
				)

				// Start should fail
				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Multi-Component Scenarios", func() {
		Context("with multiple status components", func() {
			It("should manage multiple independent components", func() {
				keys := []string{"status-api", "status-web", "status-admin"}

				for _, key := range keys {
					RegisterNew(ctx, cfg, key)

					vpr := libvpr.New(ctx, func() liblog.Logger { return nil })
					vpr.Viper().Set(key, map[string]interface{}{
						"return-code": map[string]int{
							"OK": 200,
						},
					})

					getCpt := func(k string) cfgtps.Component {
						return cfg.ComponentGet(k)
					}

					cpt := Load(getCpt, key)
					cpt.Init(key, ctx,
						getCpt,
						func() libvpr.Viper { return vpr },
						libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0),
						func() liblog.Logger { return nil })

					err := cpt.Start()
					Expect(err).To(BeNil())
				}

				// Verify all are started and independent
				for _, key := range keys {
					getCpt := func(k string) cfgtps.Component {
						return cfg.ComponentGet(k)
					}
					cpt := Load(getCpt, key)
					Expect(cpt).NotTo(BeNil())
					Expect(cpt.IsStarted()).To(BeTrue())
				}
			})

			It("should handle independent start/stop of multiple components", func() {
				cpt1 := New(ctx)
				cpt2 := New(ctx)
				key1, key2 := "status-1", "status-2"

				vpr1 := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr1.Viper().Set(key1, map[string]interface{}{"return-code": map[string]int{"OK": 200}})
				cpt1.Init(key1, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr1 }, nil, nil)

				vpr2 := libvpr.New(ctx, func() liblog.Logger { return nil })
				vpr2.Viper().Set(key2, map[string]interface{}{"return-code": map[string]int{"OK": 200}})
				cpt2.Init(key2, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr2 }, nil, nil)

				// Start both
				Expect(cpt1.Start()).To(Succeed())
				Expect(cpt2.Start()).To(Succeed())
				Expect(cpt1.IsStarted()).To(BeTrue())
				Expect(cpt2.IsStarted()).To(BeTrue())

				// Stop only the first one
				cpt1.Stop()
				Expect(cpt1.IsStarted()).To(BeFalse())
				Expect(cpt2.IsStarted()).To(BeTrue())

				// Stop the second one
				cpt2.Stop()
				Expect(cpt1.IsStarted()).To(BeFalse())
				Expect(cpt2.IsStarted()).To(BeFalse())
			})
		})
	})
})
