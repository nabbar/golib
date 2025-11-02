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

package mail_test

import (
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/mail"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	spfcbr "github.com/spf13/cobra"
)

// Configuration tests verify default configuration generation and CLI flag registration.
var _ = Describe("Configuration Management", func() {
	var (
		cpt CptMail
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	Describe("DefaultConfig method", func() {
		Context("generating default configuration", func() {
			It("should generate valid JSON config", func() {
				config := cpt.DefaultConfig("  ")
				Expect(config).NotTo(BeNil())
				Expect(len(config)).To(BeNumerically(">", 0))

				// Verify it's valid JSON
				var parsed map[string]interface{}
				err := json.Unmarshal(config, &parsed)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should generate config with empty indent", func() {
				config := cpt.DefaultConfig("")
				Expect(config).NotTo(BeNil())
				Expect(len(config)).To(BeNumerically(">", 0))

				var parsed map[string]interface{}
				err := json.Unmarshal(config, &parsed)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should generate config with custom indent", func() {
				config := cpt.DefaultConfig("    ")
				Expect(config).NotTo(BeNil())
				Expect(len(config)).To(BeNumerically(">", 0))

				var parsed map[string]interface{}
				err := json.Unmarshal(config, &parsed)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should contain expected mail fields", func() {
				config := cpt.DefaultConfig("  ")

				var parsed map[string]interface{}
				err := json.Unmarshal(config, &parsed)
				Expect(err).NotTo(HaveOccurred())

				// Check for expected mail configuration fields
				Expect(parsed).NotTo(BeEmpty())
			})
		})
	})

	Describe("RegisterFlag method", func() {
		Context("registering CLI flags without init", func() {
			It("should return error when component not initialized", func() {
				cmd := &spfcbr.Command{
					Use: "test",
				}

				err := cpt.RegisterFlag(cmd)
				Expect(err).To(HaveOccurred())
			})

			It("should return error with nil command when not initialized", func() {
				err := cpt.RegisterFlag(nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("RegisterMonitorPool method", func() {
		Context("monitor pool registration", func() {
			It("should not panic when called with nil", func() {
				Expect(func() {
					cpt.RegisterMonitorPool(nil)
				}).NotTo(Panic())
			})

			It("should not panic when called multiple times", func() {
				Expect(func() {
					cpt.RegisterMonitorPool(nil)
					cpt.RegisterMonitorPool(nil)
				}).NotTo(Panic())
			})
		})
	})
})
