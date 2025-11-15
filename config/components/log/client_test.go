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
	"sync/atomic"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	loglvl "github.com/nabbar/golib/logger/level"
)

// Client tests verify internal client functions and started logger behavior.
var _ = Describe("Client Operations", func() {
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
		cnl()
		if cpt != nil {
			cpt.Stop()
		}
	})

	Describe("Internal helper methods", func() {
		Context("with proper initialization", func() {
			It("should handle component with valid viper config", func() {
				// Component is initialized
				Expect(cpt).NotTo(BeNil())
			})

			It("should handle config missing key", func() {
				configData := map[string]interface{}{
					"other": map[string]interface{}{
						"value": "test",
					},
				}

				configJSON, err := json.Marshal(configData)
				Expect(err).To(BeNil())

				err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
				Expect(err).To(BeNil())

				// Start should fail with missing key
				err = cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("callback execution", func() {
			It("should execute callbacks during start", func() {
				cnt := new(atomic.Uint32)
				before := func(c cfgtps.Component) error {
					cnt.Add(1)
					return nil
				}
				after := func(c cfgtps.Component) error {
					cnt.Add(1)
					return nil
				}
				cpt.RegisterFuncStart(before, after)

				// Start the component - may fail due to config validation
				err := cpt.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(cnt.Load()).To(Equal(uint32(2)))

				err = cpt.Reload()
				Expect(err).ToNot(HaveOccurred())
				Expect(cnt.Load()).To(Equal(uint32(2)))
				// Callbacks are registered and will be called if start succeeds
			})

			It("should execute callbacks during reload", func() {
				cnt := new(atomic.Uint32)
				before := func(c cfgtps.Component) error {
					cnt.Add(1)
					return nil
				}
				after := func(c cfgtps.Component) error {
					cnt.Add(1)
					return nil
				}

				cpt.RegisterFuncReload(before, after)

				err := cpt.Start()
				Expect(err).ToNot(HaveOccurred())
				Expect(cnt.Load()).To(Equal(uint32(0)))

				err = cpt.Reload()
				Expect(err).ToNot(HaveOccurred())
				Expect(cnt.Load()).To(Equal(uint32(2)))

				// Callbacks are registered and will be called if reload succeeds
			})
		})

		Context("error propagation", func() {
			It("should propagate callback errors during start", func() {
				callbackErr := ErrorParamInvalid.Error(nil)
				before := func(c cfgtps.Component) error {
					return callbackErr
				}

				cpt.RegisterFuncStart(before, nil)

				// Start should fail with callback error
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(ErrorParamInvalid.Error().Error()))
			})

			It("should propagate callback errors during reload", func() {
				// Try to start first
				_ = cpt.Start()

				callbackErr := ErrorParamInvalid.Error(nil)
				before := func(c cfgtps.Component) error {
					return callbackErr
				}

				cpt.RegisterFuncReload(before, nil)

				// Reload should fail with callback error
				err := cpt.Reload()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(ErrorParamInvalid.Error().Error()))
			})
		})
	})

	Describe("Started logger operations", func() {
		Context("with started logger", func() {
			It("should return logger instance when started", func() {
				// Start should succeed with valid config
				err := cpt.Start()
				Expect(err).ToNot(HaveOccurred())

				// If started, Log() should return a logger
				logger := cpt.Log()
				Expect(logger).NotTo(BeNil())

				// Stop to cleanup
				cpt.Stop()
			})
		})
	})
})
