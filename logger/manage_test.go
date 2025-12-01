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

package logger_test

import (
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger Management", func() {
	var log liblog.Logger

	BeforeEach(func() {
		log = liblog.New(GetContext())
		log.SetLevel(loglvl.InfoLevel)
		err := log.SetOptions(&logcfg.Options{
			Stdout: &logcfg.OptionsStd{
				DisableStandard: true,
			},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if log != nil {
			Expect(log.Close()).ToNot(HaveOccurred())
		}
	})

	Describe("Clone method", func() {
		Context("when cloning logger", func() {
			It("should create independent copy", func() {
				clone, err := log.Clone()
				Expect(err).ToNot(HaveOccurred())
				Expect(clone).ToNot(BeNil())
				Expect(clone).ToNot(BeIdenticalTo(log))

				// Modifications to clone shouldn't affect original
				clone.SetLevel(loglvl.DebugLevel)
				Expect(log.GetLevel()).To(Equal(loglvl.InfoLevel))
			})

			It("should copy fields", func() {
				log.SetFields(logfld.New(GetContext()).Add("original", "value"))
				clone, err := log.Clone()
				Expect(err).ToNot(HaveOccurred())

				fields := clone.GetFields()
				Expect(fields).ToNot(BeNil())
				// Fields should contain at least one entry
				logrusFields := fields.Logrus()
				Expect(len(logrusFields)).To(BeNumerically(">", 0))
			})

			It("should allow multiple clones", func() {
				var (
					clone1, clone2, clone3 liblog.Logger
					err                    error
				)

				clone1, err = log.Clone()
				Expect(err).ToNot(HaveOccurred())
				Expect(clone1).ToNot(BeNil())

				clone2, err = log.Clone()
				Expect(err).ToNot(HaveOccurred())
				Expect(clone2).ToNot(BeNil())

				clone3, err = log.Clone()
				Expect(err).ToNot(HaveOccurred())
				Expect(clone3).ToNot(BeNil())

				// All should be different instances
				Expect(clone1).ToNot(BeIdenticalTo(clone2))
				Expect(clone2).ToNot(BeIdenticalTo(clone3))
				Expect(clone1).ToNot(BeIdenticalTo(clone3))
			})

			It("should close clones independently", func() {
				clone, err := log.Clone()
				Expect(err).ToNot(HaveOccurred())
				Expect(clone.Close()).ToNot(HaveOccurred())

				// Original should still be usable
				Expect(func() {
					log.Info("Still working", nil)
				}).ToNot(Panic())
			})
		})
	})

	Describe("SetLevel and GetLevel methods", func() {
		Context("when managing log level", func() {
			It("should set and get level correctly", func() {
				log.SetLevel(loglvl.DebugLevel)
				Expect(log.GetLevel()).To(Equal(loglvl.DebugLevel))

				log.SetLevel(loglvl.WarnLevel)
				Expect(log.GetLevel()).To(Equal(loglvl.WarnLevel))

				log.SetLevel(loglvl.ErrorLevel)
				Expect(log.GetLevel()).To(Equal(loglvl.ErrorLevel))
			})

			It("should handle all log levels", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
				}

				for _, level := range levels {
					log.SetLevel(level)
					Expect(log.GetLevel()).To(Equal(level))
				}
			})
		})
	})

	Describe("SetFields and GetFields methods", func() {
		Context("when managing fields", func() {
			It("should set and get fields correctly", func() {
				fields := logfld.New(GetContext())
				fields.Add("key1", "value1")
				fields.Add("key2", 123)

				log.SetFields(fields)
				retrieved := log.GetFields()

				Expect(retrieved).ToNot(BeNil())
				logrusFields := retrieved.Logrus()
				Expect(len(logrusFields)).To(Equal(2))
			})

			It("should handle empty fields", func() {
				emptyFields := logfld.New(GetContext())
				log.SetFields(emptyFields)

				retrieved := log.GetFields()
				Expect(retrieved).ToNot(BeNil())
			})

			It("should replace existing fields", func() {
				fields1 := logfld.New(GetContext()).Add("key1", "value1")
				log.SetFields(fields1)

				fields2 := logfld.New(GetContext()).Add("key2", "value2")
				log.SetFields(fields2)

				retrieved := log.GetFields()
				logrusFields := retrieved.Logrus()
				Expect(len(logrusFields)).To(Equal(1))
			})

			It("should handle nil fields", func() {
				Expect(func() {
					log.SetFields(nil)
				}).ToNot(Panic())
			})
		})
	})

	Describe("SetOptions and GetOptions methods", func() {
		Context("when managing options", func() {
			It("should set stdout options", func() {
				opts := &logcfg.Options{
					Stdout: &logcfg.OptionsStd{
						EnableTrace: true,
					},
				}

				err := log.SetOptions(opts)
				Expect(err).ToNot(HaveOccurred())

				retrieved := log.GetOptions()
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.Stdout).ToNot(BeNil())
				Expect(retrieved.Stdout.EnableTrace).To(BeTrue())
			})

			It("should set multiple stdout options", func() {
				opts := &logcfg.Options{
					Stdout: &logcfg.OptionsStd{
						EnableTrace: true,
					},
				}

				err := log.SetOptions(opts)
				Expect(err).ToNot(HaveOccurred())

				retrieved := log.GetOptions()
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.Stdout).ToNot(BeNil())
				Expect(retrieved.Stdout.EnableTrace).To(BeTrue())
			})

			It("should handle file options", func() {
				opts := &logcfg.Options{
					Stdout:  &logcfg.OptionsStd{},
					LogFile: []logcfg.OptionsFile{},
				}

				err := log.SetOptions(opts)
				Expect(err).ToNot(HaveOccurred())

				retrieved := log.GetOptions()
				Expect(retrieved.Stdout).ToNot(BeNil())
			})

			It("should update options multiple times", func() {
				opts1 := &logcfg.Options{
					Stdout: &logcfg.OptionsStd{},
				}
				err := log.SetOptions(opts1)
				Expect(err).ToNot(HaveOccurred())

				opts2 := &logcfg.Options{
					Stdout: &logcfg.OptionsStd{
						EnableTrace: true,
					},
				}
				err = log.SetOptions(opts2)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
