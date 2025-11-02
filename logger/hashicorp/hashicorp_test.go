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

package hashicorp_test

import (
	"github.com/hashicorp/go-hclog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liblog "github.com/nabbar/golib/logger"
	loghc "github.com/nabbar/golib/logger/hashicorp"
	loglvl "github.com/nabbar/golib/logger/level"
)

var _ = Describe("HashiCorp Logger Adapter", func() {
	var (
		mockLogger *MockLogger
		hcLogger   hclog.Logger
	)

	BeforeEach(func() {
		mockLogger = NewMockLogger()
		hcLogger = loghc.New(func() liblog.Logger { return mockLogger })
	})

	Describe("New", func() {
		Context("with valid logger function", func() {
			It("should create hclog logger adapter", func() {
				logger := loghc.New(func() liblog.Logger { return mockLogger })

				Expect(logger).ToNot(BeNil())
			})
		})

		Context("with nil logger function", func() {
			It("should create adapter that handles nil gracefully", func() {
				logger := loghc.New(nil)

				Expect(logger).ToNot(BeNil())
				// Should not panic when calling methods
				logger.Info("test")
			})
		})
	})

	Describe("SetDefault", func() {
		Context("with valid logger function", func() {
			It("should set default hclog logger", func() {
				loghc.SetDefault(func() liblog.Logger { return mockLogger })

				// This should not panic
				defaultLogger := hclog.Default()
				Expect(defaultLogger).ToNot(BeNil())
			})
		})
	})

	Describe("Trace", func() {
		Context("with simple message", func() {
			It("should log at debug level", func() {
				hcLogger.Trace("trace message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.DebugLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("trace message"))
			})
		})

		Context("with arguments", func() {
			It("should log with arguments", func() {
				hcLogger.Trace("trace message", "key", "value")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.DebugLevel))
			})
		})
	})

	Describe("Debug", func() {
		Context("with simple message", func() {
			It("should log at debug level", func() {
				hcLogger.Debug("debug message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.DebugLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("debug message"))
			})
		})

		Context("with arguments", func() {
			It("should log with arguments", func() {
				hcLogger.Debug("debug message", "key", "value")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Args).To(HaveLen(2))
			})
		})
	})

	Describe("Info", func() {
		Context("with simple message", func() {
			It("should log at info level", func() {
				hcLogger.Info("info message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("info message"))
			})
		})

		Context("with arguments", func() {
			It("should log with arguments", func() {
				hcLogger.Info("info message", "key", "value")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Args).To(HaveLen(2))
			})
		})
	})

	Describe("Warn", func() {
		Context("with simple message", func() {
			It("should log at warn level", func() {
				hcLogger.Warn("warn message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.WarnLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("warn message"))
			})
		})

		Context("with arguments", func() {
			It("should log with arguments", func() {
				hcLogger.Warn("warn message", "key", "value")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Args).To(HaveLen(2))
			})
		})
	})

	Describe("Error", func() {
		Context("with simple message", func() {
			It("should log at error level", func() {
				hcLogger.Error("error message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.ErrorLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("error message"))
			})
		})

		Context("with arguments", func() {
			It("should log with arguments", func() {
				hcLogger.Error("error message", "key", "value")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Args).To(HaveLen(2))
			})
		})
	})

	Describe("Log", func() {
		Context("with NoLevel", func() {
			It("should not log anything", func() {
				hcLogger.Log(hclog.NoLevel, "message")

				Expect(mockLogger.entries).To(HaveLen(0))
			})
		})

		Context("with Off level", func() {
			It("should not log anything", func() {
				hcLogger.Log(hclog.Off, "message")

				Expect(mockLogger.entries).To(HaveLen(0))
			})
		})

		Context("with Trace level", func() {
			It("should log at debug level", func() {
				hcLogger.Log(hclog.Trace, "trace message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.DebugLevel))
			})
		})

		Context("with Debug level", func() {
			It("should log at debug level", func() {
				hcLogger.Log(hclog.Debug, "debug message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.DebugLevel))
			})
		})

		Context("with Info level", func() {
			It("should log at info level", func() {
				hcLogger.Log(hclog.Info, "info message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with Warn level", func() {
			It("should log at warn level", func() {
				hcLogger.Log(hclog.Warn, "warn message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.WarnLevel))
			})
		})

		Context("with Error level", func() {
			It("should log at error level", func() {
				hcLogger.Log(hclog.Error, "error message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.ErrorLevel))
			})
		})
	})
})
