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

package log_test

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	loglvl "github.com/nabbar/golib/logger/level"
)

// Integration tests verify end-to-end scenarios with full component lifecycle
// and real configuration.
var _ = Describe("Integration Tests", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		cpt CptLog
	)

	BeforeEach(func() {
		ctx, cnl = context.WithCancel(x)
		cpt = New(ctx, loglvl.NilLevel)
		cpt.Init(kd, ctx, nil, fv, vs, fl)

		v.Viper().SetConfigType("json")

		configData := map[string]interface{}{
			kd: map[string]interface{}{
				"stdout": map[string]interface{}{
					"disableStandard": true,
				},
			},
		}

		configJSON, err := json.Marshal(configData)
		Expect(err).To(BeNil())

		err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		if cpt != nil {
			cpt.Stop()
		}
		cnl()
	})

	Describe("Full registration and configuration cycle", func() {
		Context("with minimal valid config", func() {
			It("should register, initialize and start successfully", func() {
				configData := map[string]interface{}{
					kd: map[string]interface{}{
						"disableStandard":  true,
						"disableStack":     false,
						"disableTimestamp": false,
						"enableTrace":      false,
						"disableColor":     false,
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
				Expect(err).To(BeNil())

				// Start the component directly (already initialized in BeforeEach)
				err = cpt.Start()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with full config including options", func() {
			It("should handle complete logger configuration", func() {
				configData := map[string]interface{}{
					"log": map[string]interface{}{
						"disableStandard":  false,
						"disableStack":     true,
						"disableTimestamp": false,
						"enableTrace":      true,
						"traceFilter":      "/tmp",
						"disableColor":     true,
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
				Expect(err).ToNot(HaveOccurred())

				cfg := libcfg.New(vs)
				Register(cfg, kd, cpt)

				cpt := Load(cfg.ComponentGet, kd)
				Expect(cpt).NotTo(BeNil())

				// Component is initialized
				Expect(cpt.Type()).To(Equal("log"))
			})
		})
	})

	Describe("Component lifecycle with config", func() {
		Context("start, reload, stop cycle", func() {
			It("should handle full lifecycle", func() {
				// Test lifecycle methods are callable
				err := cpt.Start()
				Expect(err).ToNot(HaveOccurred())

				err = cpt.Reload()
				Expect(err).ToNot(HaveOccurred())

				cpt.Stop()
			})
		})

		Context("with callbacks", func() {
			It("should execute start callbacks", func() {
				// Register callbacks
				beforeCalled := false
				afterCalled := false

				before := func(c cfgtps.Component) error {
					beforeCalled = true
					return nil
				}

				after := func(c cfgtps.Component) error {
					afterCalled = true
					return nil
				}

				cpt.RegisterFuncStart(before, after)

				// Test lifecycle methods are callable
				err := cpt.Start()
				Expect(err).ToNot(HaveOccurred())

				// Callbacks should be called
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())

				err = cpt.Reload()
				Expect(err).ToNot(HaveOccurred())

				cpt.Stop()
			})

			It("should execute reload callbacks", func() {
				// Register reload callbacks
				beforeCalled := false
				afterCalled := false

				before := func(c cfgtps.Component) error {
					beforeCalled = true
					return nil
				}

				after := func(c cfgtps.Component) error {
					afterCalled = true
					return nil
				}

				cpt.RegisterFuncReload(before, after)

				// Test lifecycle methods are callable
				err := cpt.Start()
				Expect(err).ToNot(HaveOccurred())

				err = cpt.Reload()
				Expect(err).ToNot(HaveOccurred())

				// Reload callbacks should be called
				Expect(beforeCalled).To(BeTrue())
				Expect(afterCalled).To(BeTrue())

				cpt.Stop()
			})
		})
	})

	Describe("Multiple log components", func() {
		Context("managing multiple log configs", func() {
			It("should support multiple independent log components", func() {
				configData := map[string]interface{}{
					"log1": map[string]interface{}{
						"disableStandard": false,
					},
					"log2": map[string]interface{}{
						"disableStandard": true,
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).ToNot(HaveOccurred())

				err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
				Expect(err).ToNot(HaveOccurred())

				cfg := libcfg.New(vs)

				RegisterNew(ctx, cfg, "log1", loglvl.DebugLevel)
				RegisterNew(ctx, cfg, "log2", loglvl.InfoLevel)

				// Register two components
				cpt1 := Load(cfg.ComponentGet, "log1")
				cpt2 := Load(cfg.ComponentGet, "log2")

				Expect(cpt1).NotTo(BeNil())
				Expect(cpt2).NotTo(BeNil())

				cpt1.Init("log1", ctx, cfg.ComponentGet, fv, vs, fl)
				cpt2.Init("log2", ctx, cfg.ComponentGet, fv, vs, fl)

				// Verify different levels
				Expect(cpt1.GetLevel()).To(Equal(loglvl.DebugLevel))
				Expect(cpt2.GetLevel()).To(Equal(loglvl.InfoLevel))
			})
		})
	})

	Describe("Level management in integration", func() {
		Context("changing levels dynamically", func() {
			It("should allow level changes at runtime", func() {
				cfg := libcfg.New(vs)
				Register(cfg, kd, cpt)

				cpt := Load(cfg.ComponentGet, kd)
				Expect(cpt).NotTo(BeNil())

				// Change level dynamically
				cpt.SetLevel(loglvl.DebugLevel)
				Expect(cpt.GetLevel()).To(Equal(loglvl.DebugLevel))

				cpt.SetLevel(loglvl.ErrorLevel)
				Expect(cpt.GetLevel()).To(Equal(loglvl.ErrorLevel))
			})
		})
	})

	Describe("Concurrent operations", func() {
		Context("multiple goroutines", func() {
			It("should handle concurrent level changes", func() {
				cfg := libcfg.New(vs)
				Register(cfg, kd, cpt)

				cpt := Load(cfg.ComponentGet, kd)
				Expect(cpt).NotTo(BeNil())

				done := make(chan bool, 20)

				for i := 0; i < 20; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						switch idx % 3 {
						case 0:
							cpt.SetLevel(loglvl.DebugLevel)
						case 1:
							cpt.SetLevel(loglvl.InfoLevel)
						case 2:
							_ = cpt.GetLevel()
						}
						done <- true
					}(i)
				}

				for i := 0; i < 20; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent access to multiple components", func() {
				cfg := libcfg.New(vs)

				RegisterNew(ctx, cfg, "log1", DefaultLevel)
				RegisterNew(ctx, cfg, "log2", DefaultLevel)

				cpt1 := Load(cfg.ComponentGet, "log1")
				cpt2 := Load(cfg.ComponentGet, "log2")

				done := make(chan bool, 20)

				for i := 0; i < 20; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						if idx%2 == 0 {
							cpt1.SetLevel(loglvl.DebugLevel)
							_ = cpt1.GetLevel()
						} else {
							cpt2.SetLevel(loglvl.InfoLevel)
							_ = cpt2.GetLevel()
						}
						done <- true
					}(i)
				}

				for i := 0; i < 20; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})
})
