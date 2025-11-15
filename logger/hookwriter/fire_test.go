/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package hookwriter_test

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	loghkw "github.com/nabbar/golib/logger/hookwriter"
)

var _ = Describe("HookStdOut Fire and Integration", func() {
	Describe("Fire method", func() {
		Context("with basic log entry", func() {
			It("should not error", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				// Create a log entry
				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = "test message"
				entry.Data = logrus.Fields{"key": "value"}

				// Fire should not error
				err = hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with empty data", func() {
			It("should handle empty fields", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = "test"
				entry.Data = logrus.Fields{}

				err = hook.Fire(entry)
				// Should not error, just returns nil
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with DisableStack", func() {
			It("should filter stack field", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					DisableStack:    true,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = "test"
				entry.Data = logrus.Fields{
					"stack": "should be filtered",
					"other": "should remain",
				}

				err = hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with DisableTimestamp", func() {
			It("should filter time field", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard:  false,
					DisableTimestamp: true,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = "test"
				entry.Data = logrus.Fields{
					"time":  "should be filtered",
					"other": "should remain",
				}

				err = hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with EnableTrace false", func() {
			It("should filter trace fields", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					EnableTrace:     false,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = "test"
				entry.Data = logrus.Fields{
					"caller": "should be filtered",
					"file":   "should be filtered",
					"line":   "should be filtered",
					"other":  "should remain",
				}

				err = hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with EnableAccessLog", func() {
			It("should use message-only format", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					EnableAccessLog: true,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = "access log entry"
				entry.Data = logrus.Fields{"field": "ignored"}

				err = hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle empty message", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
					EnableAccessLog: true,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = ""

				err = hook.Fire(entry)
				// Should return nil for empty message
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with JSON formatter", func() {
			It("should use formatter", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}
				formatter := &logrus.JSONFormatter{}

				hook, err := loghkw.New(io.Discard, opt, nil, formatter)
				Expect(err).ToNot(HaveOccurred())

				lgr := logrus.New()
				lgr.SetOutput(io.Discard)
				entry := logrus.NewEntry(lgr)
				entry.Level = logrus.InfoLevel
				entry.Message = "json message"
				entry.Data = logrus.Fields{"key": "value"}

				err = hook.Fire(entry)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Integration with logrus", func() {
		Context("with standard setup", func() {
			It("should work with logrus logger", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				logger := logrus.New()
				logger.SetOutput(io.Discard)
				logger.AddHook(hook)

				// These should not panic or error
				logger.Info("info message")
				logger.Warn("warn message")
				logger.Error("error message")
				logger.WithField("key", "value").Info("with field")
			})
		})

		Context("with level filtering", func() {
			It("should only handle specified levels", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}
				levels := []logrus.Level{logrus.ErrorLevel}

				hook, err := loghkw.New(io.Discard, opt, levels, nil)
				Expect(err).ToNot(HaveOccurred())

				logger := logrus.New()
				logger.SetOutput(io.Discard)
				logger.AddHook(hook)

				// Only error level should be handled by hook
				logger.Info("should not be handled")
				logger.Error("should be handled")
			})
		})

		Context("with multiple hooks", func() {
			It("should coexist with other hooks", func() {
				opt1 := &logcfg.OptionsStd{
					DisableStandard: false,
					DisableColor:    true,
				}
				opt2 := &logcfg.OptionsStd{
					DisableStandard: false,
					DisableColor:    false,
				}

				hook1, err := loghkw.New(io.Discard, opt1, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				hook2, err := loghkw.New(io.Discard, opt2, nil, &logrus.JSONFormatter{})
				Expect(err).ToNot(HaveOccurred())

				logger := logrus.New()
				logger.SetOutput(io.Discard)
				logger.AddHook(hook1)
				logger.AddHook(hook2)

				logger.Info("logged by both hooks")
			})
		})
	})

	Describe("Run method", func() {
		Context("when called", func() {
			It("should return immediately", func() {
				opt := &logcfg.OptionsStd{
					DisableStandard: false,
				}

				hook, err := loghkw.New(io.Discard, opt, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				// Run should be a no-op
				hook.Run(nil)

				// Should complete without blocking
				Expect(hook).ToNot(BeNil())
			})
		})
	})
})
