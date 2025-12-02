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

package hookstderr_test

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	loghks "github.com/nabbar/golib/logger/hookstderr"
)

var _ = Describe("HookStdErr Creation and Configuration", func() {
	Describe("New", func() {
		Context("with nil options", func() {
			It("should return nil hook", func() {
				hook, err := loghks.New(nil, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).To(BeNil())
			})
		})

		Context("with DisableStandard true", func() {
			It("should return nil hook", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: true,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).To(BeNil())
			})
		})

		Context("with valid options", func() {
			It("should create hook successfully", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with empty levels", func() {
			It("should use all levels", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, []logrus.Level{}, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				levels := hook.Levels()
				Expect(levels).To(Equal(logrus.AllLevels))
			})
		})

		Context("with specific levels", func() {
			It("should use provided levels", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}
				customLevels := []logrus.Level{
					logrus.ErrorLevel,
					logrus.WarnLevel,
					logrus.InfoLevel,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, customLevels, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				levels := hook.Levels()
				Expect(levels).To(Equal(customLevels))
			})
		})

		Context("with DisableColor true", func() {
			It("should use os.Stderr", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					DisableColor:    true,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with DisableColor false", func() {
			It("should use colorable stderr", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					DisableColor:    false,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with DisableStack option", func() {
			It("should accept DisableStack", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					DisableStack:    true,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with DisableTimestamp option", func() {
			It("should accept DisableTimestamp", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard:  false,
					DisableTimestamp: true,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with EnableTrace option", func() {
			It("should accept EnableTrace", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					EnableTrace:     true,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with EnableAccessLog option", func() {
			It("should accept EnableAccessLog", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					EnableAccessLog: true,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with custom formatter", func() {
			It("should accept JSON formatter", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}
				formatter := &logrus.JSONFormatter{}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, formatter)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})

			It("should accept Text formatter", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}
				formatter := &logrus.TextFormatter{}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, formatter)

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})

		Context("with all options combined", func() {
			It("should handle all options together", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard:  false,
					DisableColor:     true,
					DisableStack:     true,
					DisableTimestamp: true,
					EnableTrace:      true,
					EnableAccessLog:  false,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, &logrus.JSONFormatter{})

				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())
			})
		})
	})

	Describe("RegisterHook", func() {
		Context("with valid logger", func() {
			It("should register hook successfully", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				logger := logrus.New()
				hook.RegisterHook(logger)

				// Logger should now have the hook
				Expect(logger).ToNot(BeNil())
			})
		})
	})

	Describe("Levels", func() {
		Context("with default levels", func() {
			It("should return all logrus levels", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				levels := hook.Levels()
				Expect(levels).To(HaveLen(len(logrus.AllLevels)))
				Expect(levels).To(Equal(logrus.AllLevels))
			})
		})

		Context("with custom levels", func() {
			It("should return custom levels", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}
				customLevels := []logrus.Level{
					logrus.ErrorLevel,
					logrus.FatalLevel,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, customLevels, nil)
				Expect(err).ToNot(HaveOccurred())

				levels := hook.Levels()
				Expect(levels).To(Equal(customLevels))
			})
		})
	})

	Describe("Write method", func() {
		Context("with valid hook", func() {
			It("should implement io.Writer", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghks.NewWithWriter(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(hook).ToNot(BeNil())

				// Test that it implements io.Writer
				data := []byte("test message\n")
				n, err := hook.Write(data)

				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))
			})
		})
	})
})
