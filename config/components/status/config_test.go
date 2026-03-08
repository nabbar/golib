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
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY, EXPRESS OR
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

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Config tests verify configuration parsing and validation.
var _ = Describe("Configuration Management", func() {
	var (
		ctx context.Context
		vpr libvpr.Viper
		cpt CptStatus
		key string
		log = func() liblog.Logger { return nil }
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)
	)

	BeforeEach(func() {
		ctx = context.Background()
		key = "test-status"
		vpr = libvpr.New(ctx, log)
		cpt = New(ctx)
	})

	Describe("Valid Configuration", func() {
		Context("with a simple, valid configuration", func() {
			It("should parse the configuration and start successfully", func() {
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 200,
					},
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Invalid Configuration", func() {
		Context("when the configuration key is missing", func() {
			It("should return an error", func() {
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})

			It("should return an error if the viper instance is nil", func() {
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return nil }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when the configuration has the wrong data type", func() {
			It("should return an error for a non-map structure", func() {
				vpr.Viper().Set(key, "not-a-map")
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})

			It("should return an error for an array instead of a map", func() {
				vpr.Viper().Set(key, []string{"value1", "value2"})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Configuration Reload", func() {
		Context("when updating a valid configuration", func() {
			It("should reload with the new configuration successfully", func() {
				// Initial config
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 200,
					},
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)
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
		})

		Context("when reloading with an invalid configuration", func() {
			BeforeEach(func() {
				// Initial valid config
				vpr.Viper().Set(key, map[string]interface{}{
					"return-code": map[string]int{
						"OK": 200,
					},
				})
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)
				err := cpt.Start()
				Expect(err).To(BeNil())
			})

			It("should return an error if the config key is removed", func() {
				// Remove config key
				vpr.Viper().Set(key, nil)

				err := cpt.Reload()
				Expect(err).NotTo(BeNil())
			})

			It("should return an error if the config type becomes invalid", func() {
				// Set invalid config
				vpr.Viper().Set(key, "invalid")

				err := cpt.Reload()
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
