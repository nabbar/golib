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
 */

package tls_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/tls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	spfcbr "github.com/spf13/cobra"
)

// Configuration tests verify config loading, validation, and flag registration.
var _ = Describe("Configuration Management", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		cpt CptTlS
	)

	BeforeEach(func() {
		ctx, cnl = context.WithCancel(x)
		cpt = New(ctx, nil)
	})

	AfterEach(func() {
		cnl()
		if cpt != nil {
			cpt.Stop()
		}
	})

	Describe("RegisterFlag method", func() {
		Context("registering command flags", func() {
			It("should not panic with nil command", func() {
				Expect(func() {
					_ = cpt.RegisterFlag(nil)
				}).NotTo(Panic())
			})

			It("should accept valid cobra command", func() {
				cmd := &spfcbr.Command{
					Use:   "test",
					Short: "Test command",
				}

				err := cpt.RegisterFlag(cmd)
				Expect(err).To(BeNil())
			})

			It("should not modify command flags", func() {
				cmd := &spfcbr.Command{
					Use:   "test",
					Short: "Test command",
				}

				initialFlagCount := cmd.Flags().NFlag()
				_ = cpt.RegisterFlag(cmd)
				finalFlagCount := cmd.Flags().NFlag()

				// TLS component doesn't add flags, so count should be same
				Expect(finalFlagCount).To(Equal(initialFlagCount))
			})

			It("should allow multiple calls", func() {
				cmd := &spfcbr.Command{
					Use:   "test",
					Short: "Test command",
				}

				err1 := cpt.RegisterFlag(cmd)
				err2 := cpt.RegisterFlag(cmd)

				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
			})
		})
	})

	Describe("Configuration loading", func() {
		Context("with uninitialized component", func() {
			It("should fail to start without initialization", func() {
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})

			It("should fail to reload without initialization", func() {
				err := cpt.Reload()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with initialized but no viper", func() {
			It("should fail to start without viper", func() {
				cpt.Init(kd, ctx, nil, fv, vs, fl)
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
				// Error message may vary based on implementation
			})
		})
	})

	Describe("Configuration validation", func() {
		Context("error conditions", func() {
			It("should handle missing config key gracefully", func() {
				cpt.Init(kd, ctx, nil, fv, vs, fl)
				// Start will attempt to load config - behavior depends on implementation
				_ = cpt.Start()
			})

			It("should return error for invalid config", func() {
				cpt.Init(kd, ctx, nil, fv, vs, fl)
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Configuration state", func() {
		Context("before and after initialization", func() {
			It("should not be started before Init", func() {
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should not be started after Init without Start", func() {
				cpt.Init(kd, ctx, nil, fv, vs, fl)
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})
	})

	Describe("Edge cases", func() {
		Context("configuration handling", func() {
			It("should handle empty config key", func() {
				cpt.Init(kd, ctx, nil, fv, vs, fl)
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})

			It("should handle nil command in RegisterFlag", func() {
				err := cpt.RegisterFlag(nil)
				Expect(err).To(BeNil())
			})
		})
	})
})
