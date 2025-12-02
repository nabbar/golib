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

package hashicorp_test

import (
	"github.com/hashicorp/go-hclog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	loghc "github.com/nabbar/golib/logger/hashicorp"
	loglvl "github.com/nabbar/golib/logger/level"
)

var _ = Describe("HashiCorp Logger Level Operations", func() {
	var (
		mockLogger *MockLogger
		hcLogger   hclog.Logger
	)

	BeforeEach(func() {
		mockLogger = NewMockLogger()
		hcLogger = loghc.New(func() liblog.Logger { return mockLogger })
	})

	Describe("IsTrace", func() {
		Context("when trace is enabled in stdout", func() {
			It("should return true", func() {
				mockLogger.options = &logcfg.Options{
					Stdout: &logcfg.OptionsStd{
						EnableTrace: true,
					},
				}

				Expect(hcLogger.IsTrace()).To(BeTrue())
			})
		})

		Context("when trace is enabled in log file", func() {
			It("should return true", func() {
				mockLogger.options = &logcfg.Options{
					LogFile: []logcfg.OptionsFile{
						{EnableTrace: true},
					},
				}

				Expect(hcLogger.IsTrace()).To(BeTrue())
			})
		})

		Context("when trace is enabled in syslog", func() {
			It("should return true", func() {
				mockLogger.options = &logcfg.Options{
					LogSyslog: []logcfg.OptionsSyslog{
						{EnableTrace: true},
					},
				}

				Expect(hcLogger.IsTrace()).To(BeTrue())
			})
		})

		Context("when trace is not enabled", func() {
			It("should return false", func() {
				mockLogger.options = &logcfg.Options{
					Stdout: &logcfg.OptionsStd{
						EnableTrace: false,
					},
				}

				Expect(hcLogger.IsTrace()).To(BeFalse())
			})
		})

		Context("with nil options", func() {
			It("should return false", func() {
				mockLogger.options = nil

				Expect(hcLogger.IsTrace()).To(BeFalse())
			})
		})
	})

	Describe("IsDebug", func() {
		Context("when level is DebugLevel", func() {
			It("should return true", func() {
				mockLogger.SetLevel(loglvl.DebugLevel)

				Expect(hcLogger.IsDebug()).To(BeTrue())
			})
		})

		Context("when level is InfoLevel", func() {
			It("should return false", func() {
				mockLogger.SetLevel(loglvl.InfoLevel)

				Expect(hcLogger.IsDebug()).To(BeFalse())
			})
		})

		Context("when level is ErrorLevel", func() {
			It("should return false", func() {
				mockLogger.SetLevel(loglvl.ErrorLevel)

				Expect(hcLogger.IsDebug()).To(BeFalse())
			})
		})
	})

	Describe("IsInfo", func() {
		Context("when level is InfoLevel", func() {
			It("should return true", func() {
				mockLogger.SetLevel(loglvl.InfoLevel)

				Expect(hcLogger.IsInfo()).To(BeTrue())
			})
		})

		Context("when level is DebugLevel", func() {
			It("should return true", func() {
				mockLogger.SetLevel(loglvl.DebugLevel)

				Expect(hcLogger.IsInfo()).To(BeTrue())
			})
		})

		Context("when level is WarnLevel", func() {
			It("should return false", func() {
				mockLogger.SetLevel(loglvl.WarnLevel)

				Expect(hcLogger.IsInfo()).To(BeFalse())
			})
		})
	})

	Describe("IsWarn", func() {
		Context("when level is WarnLevel", func() {
			It("should return true", func() {
				mockLogger.SetLevel(loglvl.WarnLevel)

				Expect(hcLogger.IsWarn()).To(BeTrue())
			})
		})

		Context("when level is InfoLevel", func() {
			It("should return true", func() {
				mockLogger.SetLevel(loglvl.InfoLevel)

				Expect(hcLogger.IsWarn()).To(BeTrue())
			})
		})

		Context("when level is ErrorLevel", func() {
			It("should return false", func() {
				mockLogger.SetLevel(loglvl.ErrorLevel)

				Expect(hcLogger.IsWarn()).To(BeFalse())
			})
		})
	})

	Describe("IsError", func() {
		Context("when level is ErrorLevel", func() {
			It("should return true", func() {
				mockLogger.SetLevel(loglvl.ErrorLevel)

				Expect(hcLogger.IsError()).To(BeTrue())
			})
		})

		Context("when level is WarnLevel", func() {
			It("should return true", func() {
				mockLogger.SetLevel(loglvl.WarnLevel)

				Expect(hcLogger.IsError()).To(BeTrue())
			})
		})

		Context("when level is FatalLevel", func() {
			It("should return false", func() {
				mockLogger.SetLevel(loglvl.FatalLevel)

				Expect(hcLogger.IsError()).To(BeFalse())
			})
		})
	})

	Describe("SetLevel", func() {
		Context("with NoLevel", func() {
			It("should set to NilLevel", func() {
				hcLogger.SetLevel(hclog.NoLevel)

				Expect(mockLogger.GetLevel()).To(Equal(loglvl.NilLevel))
			})
		})

		Context("with Off", func() {
			It("should set to NilLevel", func() {
				hcLogger.SetLevel(hclog.Off)

				Expect(mockLogger.GetLevel()).To(Equal(loglvl.NilLevel))
			})
		})

		Context("with Trace", func() {
			It("should set to DebugLevel and enable trace", func() {
				mockLogger.options = &logcfg.Options{
					Stdout: &logcfg.OptionsStd{},
				}

				hcLogger.SetLevel(hclog.Trace)

				Expect(mockLogger.GetLevel()).To(Equal(loglvl.DebugLevel))
				Expect(mockLogger.options.Stdout.EnableTrace).To(BeTrue())
			})
		})

		Context("with Debug", func() {
			It("should set to DebugLevel", func() {
				hcLogger.SetLevel(hclog.Debug)

				Expect(mockLogger.GetLevel()).To(Equal(loglvl.DebugLevel))
			})
		})

		Context("with Info", func() {
			It("should set to InfoLevel", func() {
				hcLogger.SetLevel(hclog.Info)

				Expect(mockLogger.GetLevel()).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with Warn", func() {
			It("should set to WarnLevel", func() {
				hcLogger.SetLevel(hclog.Warn)

				Expect(mockLogger.GetLevel()).To(Equal(loglvl.WarnLevel))
			})
		})

		Context("with Error", func() {
			It("should set to ErrorLevel", func() {
				hcLogger.SetLevel(hclog.Error)

				Expect(mockLogger.GetLevel()).To(Equal(loglvl.ErrorLevel))
			})
		})
	})

	Describe("GetLevel", func() {
		Context("when level is NilLevel", func() {
			It("should return NoLevel", func() {
				mockLogger.SetLevel(loglvl.NilLevel)

				Expect(hcLogger.GetLevel()).To(Equal(hclog.NoLevel))
			})
		})

		Context("when level is DebugLevel with trace", func() {
			It("should return Trace", func() {
				mockLogger.SetLevel(loglvl.DebugLevel)
				mockLogger.options = &logcfg.Options{
					Stdout: &logcfg.OptionsStd{
						EnableTrace: true,
					},
				}

				Expect(hcLogger.GetLevel()).To(Equal(hclog.Trace))
			})
		})

		Context("when level is DebugLevel without trace", func() {
			It("should return Debug", func() {
				mockLogger.SetLevel(loglvl.DebugLevel)
				mockLogger.options = &logcfg.Options{
					Stdout: &logcfg.OptionsStd{
						EnableTrace: false,
					},
				}

				Expect(hcLogger.GetLevel()).To(Equal(hclog.Debug))
			})
		})

		Context("when level is InfoLevel", func() {
			It("should return Info", func() {
				mockLogger.SetLevel(loglvl.InfoLevel)

				Expect(hcLogger.GetLevel()).To(Equal(hclog.Info))
			})
		})

		Context("when level is WarnLevel", func() {
			It("should return Warn", func() {
				mockLogger.SetLevel(loglvl.WarnLevel)

				Expect(hcLogger.GetLevel()).To(Equal(hclog.Warn))
			})
		})

		Context("when level is ErrorLevel", func() {
			It("should return Error", func() {
				mockLogger.SetLevel(loglvl.ErrorLevel)

				Expect(hcLogger.GetLevel()).To(Equal(hclog.Error))
			})
		})

		Context("when level is unknown", func() {
			It("should return Off", func() {
				mockLogger.SetLevel(loglvl.Level(99))

				Expect(hcLogger.GetLevel()).To(Equal(hclog.Off))
			})
		})
	})

	Describe("StandardLogger", func() {
		Context("with valid options", func() {
			It("should return standard logger", func() {
				opts := &hclog.StandardLoggerOptions{
					ForceLevel: hclog.Info,
				}

				stdLogger := hcLogger.StandardLogger(opts)

				Expect(stdLogger).ToNot(BeNil())
			})
		})

		Context("with different force levels", func() {
			It("should handle all levels", func() {
				levels := []hclog.Level{
					hclog.NoLevel,
					hclog.Trace,
					hclog.Debug,
					hclog.Info,
					hclog.Warn,
					hclog.Error,
				}

				for _, level := range levels {
					opts := &hclog.StandardLoggerOptions{
						ForceLevel: level,
					}
					stdLogger := hcLogger.StandardLogger(opts)
					Expect(stdLogger).ToNot(BeNil())
				}
			})
		})
	})

	Describe("StandardWriter", func() {
		Context("with valid options", func() {
			It("should return writer", func() {
				opts := &hclog.StandardLoggerOptions{}

				writer := hcLogger.StandardWriter(opts)

				Expect(writer).ToNot(BeNil())
			})
		})

		Context("with nil logger", func() {
			It("should return os.Stdout", func() {
				nilLogger := loghc.New(nil)
				opts := &hclog.StandardLoggerOptions{}

				writer := nilLogger.StandardWriter(opts)

				Expect(writer).ToNot(BeNil())
			})
		})
	})
})
