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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Config tests verify configuration parsing and validation
var _ = Describe("Configuration Management", func() {
	var (
		ctx context.Context
		vpr libvpr.Viper
		cpt CptHead
		key string
		log = func() liblog.Logger { return nil }
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "abcd1234", "1.0.0", "maintainer", "", Empty{}, 0)
	)

	BeforeEach(func() {
		ctx = context.Background()
		key = "test-head"
		vpr = libvpr.New(ctx, log)
		cpt = New(ctx)
	})

	Describe("Valid Configuration", func() {
		Context("with simple headers", func() {
			It("should parse basic headers", func() {
				vpr.Viper().Set(key, map[string]string{
					"X-Frame-Options":  "DENY",
					"X-XSS-Protection": "1; mode=block",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Frame-Options")).To(Equal("DENY"))
				Expect(headers.Get("X-XSS-Protection")).To(Equal("1; mode=block"))
			})

			It("should parse multiple headers", func() {
				vpr.Viper().Set(key, map[string]string{
					"Header-1": "value-1",
					"Header-2": "value-2",
					"Header-3": "value-3",
					"Header-4": "value-4",
					"Header-5": "value-5",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("Header-1")).To(Equal("value-1"))
				Expect(headers.Get("Header-2")).To(Equal("value-2"))
				Expect(headers.Get("Header-3")).To(Equal("value-3"))
				Expect(headers.Get("Header-4")).To(Equal("value-4"))
				Expect(headers.Get("Header-5")).To(Equal("value-5"))
			})
		})

		Context("with security headers", func() {
			It("should parse Content-Security-Policy", func() {
				vpr.Viper().Set(key, map[string]string{
					"Content-Security-Policy": "default-src 'self'; script-src 'self' 'unsafe-inline'",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("Content-Security-Policy")).To(Equal("default-src 'self'; script-src 'self' 'unsafe-inline'"))
			})

			It("should parse Strict-Transport-Security", func() {
				vpr.Viper().Set(key, map[string]string{
					"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("Strict-Transport-Security")).To(Equal("max-age=31536000; includeSubDomains; preload"))
			})

			It("should parse all common security headers", func() {
				vpr.Viper().Set(key, map[string]string{
					"X-Frame-Options":        "SAMEORIGIN",
					"X-Content-Type-Options": "nosniff",
					"X-XSS-Protection":       "1; mode=block",
					"Referrer-Policy":        "no-referrer",
					"Permissions-Policy":     "geolocation=(), microphone=()",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Frame-Options")).To(Equal("SAMEORIGIN"))
				Expect(headers.Get("X-Content-Type-Options")).To(Equal("nosniff"))
				Expect(headers.Get("X-XSS-Protection")).To(Equal("1; mode=block"))
				Expect(headers.Get("Referrer-Policy")).To(Equal("no-referrer"))
				Expect(headers.Get("Permissions-Policy")).To(Equal("geolocation=(), microphone=()"))
			})
		})

		Context("with CORS headers", func() {
			It("should parse CORS configuration", func() {
				vpr.Viper().Set(key, map[string]string{
					"Access-Control-Allow-Origin":      "https://example.com",
					"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
					"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-Requested-With",
					"Access-Control-Allow-Credentials": "true",
					"Access-Control-Max-Age":           "86400",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("Access-Control-Allow-Origin")).To(Equal("https://example.com"))
				Expect(headers.Get("Access-Control-Allow-Methods")).To(Equal("GET, POST, PUT, DELETE, OPTIONS"))
				Expect(headers.Get("Access-Control-Allow-Headers")).To(Equal("Content-Type, Authorization, X-Requested-With"))
				Expect(headers.Get("Access-Control-Allow-Credentials")).To(Equal("true"))
				Expect(headers.Get("Access-Control-Max-Age")).To(Equal("86400"))
			})
		})

		Context("with cache headers", func() {
			It("should parse cache control headers", func() {
				vpr.Viper().Set(key, map[string]string{
					"Cache-Control": "public, max-age=3600",
					"ETag":          "\"33a64df551425fcc55e4d42a148795d9f25f89d4\"",
					"Vary":          "Accept-Encoding",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("Cache-Control")).To(Equal("public, max-age=3600"))
				Expect(headers.Get("ETag")).To(Equal("\"33a64df551425fcc55e4d42a148795d9f25f89d4\""))
				Expect(headers.Get("Vary")).To(Equal("Accept-Encoding"))
			})
		})

		Context("with custom headers", func() {
			It("should parse custom headers", func() {
				vpr.Viper().Set(key, map[string]string{
					"X-Custom-Header":  "custom-value",
					"X-API-Version":    "v1.0.0",
					"X-Request-ID":     "12345",
					"X-Custom-Feature": "enabled",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Custom-Header")).To(Equal("custom-value"))
				Expect(headers.Get("X-API-Version")).To(Equal("v1.0.0"))
				Expect(headers.Get("X-Request-ID")).To(Equal("12345"))
				Expect(headers.Get("X-Custom-Feature")).To(Equal("enabled"))
			})
		})

		Context("with empty values", func() {
			It("should handle empty header values", func() {
				vpr.Viper().Set(key, map[string]string{
					"X-Empty-Header":  "",
					"X-Normal-Header": "value",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Empty-Header")).To(Equal(""))
				Expect(headers.Get("X-Normal-Header")).To(Equal("value"))
			})
		})
	})

	Describe("Invalid Configuration", func() {
		Context("missing configuration", func() {
			It("should return error when key is not set", func() {
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})

			It("should return error with empty viper", func() {
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return nil }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with wrong configuration type", func() {
			It("should handle invalid config structure", func() {
				vpr.Viper().Set(key, "not-a-map")
				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})

			It("should handle array instead of map", func() {
				vpr.Viper().Set(key, []string{"value1", "value2"})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Configuration Reload", func() {
		Context("updating configuration", func() {
			It("should reload with new configuration", func() {
				// Initial config
				vpr.Viper().Set(key, map[string]string{
					"X-Header": "old-value",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Header")).To(Equal("old-value"))

				// Update config
				vpr.Viper().Set(key, map[string]string{
					"X-Header": "new-value",
				})

				err = cpt.Reload()
				Expect(err).To(BeNil())

				headers = cpt.GetHeaders()
				Expect(headers.Get("X-Header")).To(Equal("new-value"))
			})

			It("should add new headers on reload", func() {
				// Initial config
				vpr.Viper().Set(key, map[string]string{
					"X-Header-1": "value-1",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				// Update config with additional headers
				vpr.Viper().Set(key, map[string]string{
					"X-Header-1": "value-1",
					"X-Header-2": "value-2",
					"X-Header-3": "value-3",
				})

				err = cpt.Reload()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Header-1")).To(Equal("value-1"))
				Expect(headers.Get("X-Header-2")).To(Equal("value-2"))
				Expect(headers.Get("X-Header-3")).To(Equal("value-3"))
			})

			It("should remove headers on reload", func() {
				// Initial config
				vpr.Viper().Set(key, map[string]string{
					"X-Header-1": "value-1",
					"X-Header-2": "value-2",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				// Update config with fewer headers
				vpr.Viper().Set(key, map[string]string{
					"X-Header-1": "value-1",
				})

				err = cpt.Reload()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("X-Header-1")).To(Equal("value-1"))
				Expect(headers.Get("X-Header-2")).To(BeEmpty())
			})

			It("should handle complete replacement on reload", func() {
				// Initial config
				vpr.Viper().Set(key, map[string]string{
					"Old-Header-1": "old-value-1",
					"Old-Header-2": "old-value-2",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				// Completely new config
				vpr.Viper().Set(key, map[string]string{
					"New-Header-1": "new-value-1",
					"New-Header-2": "new-value-2",
				})

				err = cpt.Reload()
				Expect(err).To(BeNil())

				headers := cpt.GetHeaders()
				Expect(headers.Get("Old-Header-1")).To(BeEmpty())
				Expect(headers.Get("Old-Header-2")).To(BeEmpty())
				Expect(headers.Get("New-Header-1")).To(Equal("new-value-1"))
				Expect(headers.Get("New-Header-2")).To(Equal("new-value-2"))
			})
		})

		Context("reload error handling", func() {
			It("should handle reload with missing config", func() {
				// Initial config
				vpr.Viper().Set(key, map[string]string{
					"X-Header": "value",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				// Remove config
				vpr.Viper().Set(key, nil)

				err = cpt.Reload()
				Expect(err).NotTo(BeNil())
			})

			It("should handle reload with invalid config", func() {
				// Initial config
				vpr.Viper().Set(key, map[string]string{
					"X-Header": "value",
				})

				cpt.Init(key, ctx, func(string) cfgtps.Component { return nil }, func() libvpr.Viper { return vpr }, vrs, log)

				err := cpt.Start()
				Expect(err).To(BeNil())

				// Invalid config
				vpr.Viper().Set(key, "invalid")

				err = cpt.Reload()
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
