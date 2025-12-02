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
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger SPF13 Integration", func() {
	var log liblog.Logger

	BeforeEach(func() {
		log = liblog.New(GetContext())
		log.SetLevel(loglvl.InfoLevel)
		err := log.SetOptions(&logcfg.Options{
			Stdout: &logcfg.OptionsStd{},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if log != nil {
			Expect(log.Close()).ToNot(HaveOccurred())
		}
	})

	Describe("SetSPF13Level method", func() {
		Context("when setting SPF13 log level", func() {
			It("should set panic level", func() {
				Expect(func() {
					log.SetSPF13Level(loglvl.PanicLevel, nil)
				}).ToNot(Panic())
			})

			It("should set fatal level", func() {
				Expect(func() {
					log.SetSPF13Level(loglvl.FatalLevel, nil)
				}).ToNot(Panic())
			})

			It("should set error level", func() {
				Expect(func() {
					log.SetSPF13Level(loglvl.ErrorLevel, nil)
				}).ToNot(Panic())
			})

			It("should set warn level", func() {
				Expect(func() {
					log.SetSPF13Level(loglvl.WarnLevel, nil)
				}).ToNot(Panic())
			})

			It("should set info level", func() {
				Expect(func() {
					log.SetSPF13Level(loglvl.InfoLevel, nil)
				}).ToNot(Panic())
			})

			It("should set debug level", func() {
				Expect(func() {
					log.SetSPF13Level(loglvl.DebugLevel, nil)
				}).ToNot(Panic())
			})

			It("should handle all valid level values", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
				}

				for _, level := range levels {
					Expect(func() {
						log.SetSPF13Level(level, nil)
					}).ToNot(Panic())
				}
			})

			It("should not affect logger level", func() {
				// SetSPF13Level configures the SPF13 logger, not the main logger level
				originalLevel := log.GetLevel()

				log.SetSPF13Level(loglvl.ErrorLevel, nil)
				Expect(log.GetLevel()).To(Equal(originalLevel))

				log.SetSPF13Level(loglvl.InfoLevel, nil)
				Expect(log.GetLevel()).To(Equal(originalLevel))
			})
		})
	})
})
