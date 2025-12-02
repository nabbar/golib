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
	"log"

	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger Go Standard Logger", func() {
	var logger liblog.Logger

	BeforeEach(func() {
		logger = liblog.New(GetContext())
		logger.SetLevel(loglvl.DebugLevel)
		err := logger.SetOptions(&logcfg.Options{
			Stdout: &logcfg.OptionsStd{
				DisableStandard: true,
			},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if logger != nil {
			Expect(logger.Close()).ToNot(HaveOccurred())
		}
	})

	Describe("GetStdLogger method", func() {
		Context("when getting standard logger", func() {
			It("should return valid standard logger", func() {
				stdLogger := logger.GetStdLogger(loglvl.InfoLevel, log.LstdFlags)
				Expect(stdLogger).ToNot(BeNil())
			})

			It("should return usable logger", func() {
				stdLogger := logger.GetStdLogger(loglvl.InfoLevel, log.LstdFlags)
				Expect(func() {
					stdLogger.Println("Test message from standard logger")
				}).ToNot(Panic())
			})

			It("should create logger with specified level", func() {
				stdLogger := logger.GetStdLogger(loglvl.WarnLevel, log.LstdFlags)
				Expect(stdLogger).ToNot(BeNil())
				Expect(logger.GetIOWriterLevel()).To(Equal(loglvl.WarnLevel))
			})

			It("should create logger with specified flags", func() {
				stdLogger := logger.GetStdLogger(loglvl.InfoLevel, log.Ldate|log.Ltime)
				Expect(stdLogger).ToNot(BeNil())
			})

			It("should work with different log levels", func() {
				levels := []loglvl.Level{
					loglvl.DebugLevel,
					loglvl.InfoLevel,
					loglvl.WarnLevel,
					loglvl.ErrorLevel,
				}

				for _, level := range levels {
					stdLogger := logger.GetStdLogger(level, log.LstdFlags)
					Expect(stdLogger).ToNot(BeNil())
				}
			})
		})
	})

	Describe("SetStdLogger method", func() {
		Context("when setting global standard logger", func() {
			It("should set global logger successfully", func() {
				Expect(func() {
					logger.SetStdLogger(loglvl.InfoLevel, log.LstdFlags)
				}).ToNot(Panic())
			})

			It("should set with different levels", func() {
				Expect(func() {
					logger.SetStdLogger(loglvl.WarnLevel, log.LstdFlags)
				}).ToNot(Panic())
				Expect(logger.GetIOWriterLevel()).To(Equal(loglvl.WarnLevel))
			})

			It("should set with custom flags", func() {
				Expect(func() {
					logger.SetStdLogger(loglvl.InfoLevel, log.Ldate|log.Ltime|log.Lshortfile)
				}).ToNot(Panic())
			})

			It("should allow logging after setting", func() {
				logger.SetStdLogger(loglvl.InfoLevel, log.LstdFlags)

				Expect(func() {
					log.Println("Message after setting global logger")
				}).ToNot(Panic())
			})
		})
	})

	Describe("Standard logger integration", func() {
		Context("when using Go standard logger interface", func() {
			It("should work with third-party code expecting log.Logger", func() {
				stdLogger := logger.GetStdLogger(loglvl.InfoLevel, log.LstdFlags)

				// Simulate third-party function that needs *log.Logger
				thirdPartyFunc := func(l *log.Logger) {
					l.Println("Called from third party")
					l.Printf("Formatted: %s", "message")
				}

				Expect(func() {
					thirdPartyFunc(stdLogger)
				}).ToNot(Panic())
			})

			It("should support all standard logger methods", func() {
				stdLogger := logger.GetStdLogger(loglvl.InfoLevel, log.LstdFlags)

				Expect(func() {
					stdLogger.Print("Print message")
					stdLogger.Println("Println message")
					stdLogger.Printf("Printf message: %d", 123)
				}).ToNot(Panic())
			})
		})
	})
})
