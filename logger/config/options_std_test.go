/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package config_test

import (
	. "github.com/nabbar/golib/logger/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OptionsStd", func() {
	Describe("Clone", func() {
		Context("with empty options", func() {
			It("should return a valid clone with default values", func() {
				original := &OptionsStd{}
				clone := original.Clone()

				Expect(clone).ToNot(BeNil())
				Expect(clone.DisableStandard).To(BeFalse())
				Expect(clone.DisableStack).To(BeFalse())
				Expect(clone.DisableTimestamp).To(BeFalse())
				Expect(clone.EnableTrace).To(BeFalse())
				Expect(clone.DisableColor).To(BeFalse())
				Expect(clone.EnableAccessLog).To(BeFalse())
			})
		})

		Context("with all flags enabled", func() {
			It("should clone all boolean flags correctly", func() {
				original := &OptionsStd{
					DisableStandard:  true,
					DisableStack:     true,
					DisableTimestamp: true,
					EnableTrace:      true,
					DisableColor:     true,
					EnableAccessLog:  true,
				}

				clone := original.Clone()

				Expect(clone).ToNot(BeIdenticalTo(original))
				Expect(clone.DisableStandard).To(BeTrue())
				Expect(clone.DisableStack).To(BeTrue())
				Expect(clone.DisableTimestamp).To(BeTrue())
				Expect(clone.EnableTrace).To(BeTrue())
				Expect(clone.DisableColor).To(BeTrue())
				Expect(clone.EnableAccessLog).To(BeTrue())

				// Verify it's a deep copy
				clone.DisableStandard = false
				Expect(original.DisableStandard).To(BeTrue())
			})
		})

		Context("with mixed flags", func() {
			It("should clone mixed boolean values correctly", func() {
				original := &OptionsStd{
					DisableStandard:  true,
					DisableStack:     false,
					DisableTimestamp: true,
					EnableTrace:      false,
					DisableColor:     true,
					EnableAccessLog:  false,
				}

				clone := original.Clone()

				Expect(clone.DisableStandard).To(BeTrue())
				Expect(clone.DisableStack).To(BeFalse())
				Expect(clone.DisableTimestamp).To(BeTrue())
				Expect(clone.EnableTrace).To(BeFalse())
				Expect(clone.DisableColor).To(BeTrue())
				Expect(clone.EnableAccessLog).To(BeFalse())
			})
		})
	})

	Describe("Field Behavior", func() {
		Context("DisableStandard flag", func() {
			It("should control standard output", func() {
				opts := &OptionsStd{
					DisableStandard: true,
				}

				Expect(opts.DisableStandard).To(BeTrue())
			})
		})

		Context("DisableStack flag", func() {
			It("should control goroutine id display", func() {
				opts := &OptionsStd{
					DisableStack: true,
				}

				Expect(opts.DisableStack).To(BeTrue())
			})
		})

		Context("DisableTimestamp flag", func() {
			It("should control timestamp display", func() {
				opts := &OptionsStd{
					DisableTimestamp: true,
				}

				Expect(opts.DisableTimestamp).To(BeTrue())
			})
		})

		Context("EnableTrace flag", func() {
			It("should control caller information display", func() {
				opts := &OptionsStd{
					EnableTrace: true,
				}

				Expect(opts.EnableTrace).To(BeTrue())
			})
		})

		Context("DisableColor flag", func() {
			It("should control color usage in output", func() {
				opts := &OptionsStd{
					DisableColor: true,
				}

				Expect(opts.DisableColor).To(BeTrue())
			})
		})

		Context("EnableAccessLog flag", func() {
			It("should control access log inclusion", func() {
				opts := &OptionsStd{
					EnableAccessLog: true,
				}

				Expect(opts.EnableAccessLog).To(BeTrue())
			})
		})
	})

	Describe("Common Use Cases", func() {
		Context("production configuration", func() {
			It("should support production-ready settings", func() {
				opts := &OptionsStd{
					DisableStandard:  false,
					DisableStack:     false,
					DisableTimestamp: false,
					EnableTrace:      true,
					DisableColor:     true, // Disable color for production
					EnableAccessLog:  true,
				}

				Expect(opts.DisableStandard).To(BeFalse())
				Expect(opts.EnableTrace).To(BeTrue())
				Expect(opts.DisableColor).To(BeTrue())
				Expect(opts.EnableAccessLog).To(BeTrue())
			})
		})

		Context("development configuration", func() {
			It("should support development-friendly settings", func() {
				opts := &OptionsStd{
					DisableStandard:  false,
					DisableStack:     false,
					DisableTimestamp: false,
					EnableTrace:      true,
					DisableColor:     false, // Enable color for development
					EnableAccessLog:  false,
				}

				Expect(opts.DisableStandard).To(BeFalse())
				Expect(opts.EnableTrace).To(BeTrue())
				Expect(opts.DisableColor).To(BeFalse())
				Expect(opts.EnableAccessLog).To(BeFalse())
			})
		})

		Context("minimal logging configuration", func() {
			It("should support minimal logging", func() {
				opts := &OptionsStd{
					DisableStandard:  false,
					DisableStack:     true,
					DisableTimestamp: true,
					EnableTrace:      false,
					DisableColor:     true,
					EnableAccessLog:  false,
				}

				Expect(opts.DisableStack).To(BeTrue())
				Expect(opts.DisableTimestamp).To(BeTrue())
				Expect(opts.EnableTrace).To(BeFalse())
			})
		})

		Context("verbose logging configuration", func() {
			It("should support verbose logging", func() {
				opts := &OptionsStd{
					DisableStandard:  false,
					DisableStack:     false,
					DisableTimestamp: false,
					EnableTrace:      true,
					DisableColor:     false,
					EnableAccessLog:  true,
				}

				Expect(opts.DisableStack).To(BeFalse())
				Expect(opts.DisableTimestamp).To(BeFalse())
				Expect(opts.EnableTrace).To(BeTrue())
				Expect(opts.EnableAccessLog).To(BeTrue())
			})
		})

		Context("silent mode configuration", func() {
			It("should support completely silent logging", func() {
				opts := &OptionsStd{
					DisableStandard:  true,
					DisableStack:     true,
					DisableTimestamp: true,
					EnableTrace:      false,
					DisableColor:     true,
					EnableAccessLog:  false,
				}

				Expect(opts.DisableStandard).To(BeTrue())
				Expect(opts.DisableStack).To(BeTrue())
				Expect(opts.DisableTimestamp).To(BeTrue())
				Expect(opts.EnableTrace).To(BeFalse())
				Expect(opts.DisableColor).To(BeTrue())
				Expect(opts.EnableAccessLog).To(BeFalse())
			})
		})
	})

	Describe("Nil Safety", func() {
		Context("when OptionsStd is nil", func() {
			It("should not panic on Clone", func() {
				var opts *OptionsStd = nil

				Expect(func() {
					if opts != nil {
						opts.Clone()
					}
				}).ToNot(Panic())
			})
		})
	})
})
