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

package gorm_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liblog "github.com/nabbar/golib/logger"
	loggorm "github.com/nabbar/golib/logger/gorm"
	loglvl "github.com/nabbar/golib/logger/level"
	gorlog "gorm.io/gorm/logger"
)

var _ = Describe("GORM Logger Adapter", func() {
	var (
		mockLogger *MockLogger
		gormLogger gorlog.Interface
	)

	BeforeEach(func() {
		mockLogger = NewMockLogger()
		gormLogger = loggorm.New(
			func() liblog.Logger { return mockLogger },
			false,
			100*time.Millisecond,
		)
	})

	Describe("New", func() {
		Context("with valid parameters", func() {
			It("should create gorm logger adapter", func() {
				gl := loggorm.New(
					func() liblog.Logger { return mockLogger },
					false,
					100*time.Millisecond,
				)

				Expect(gl).ToNot(BeNil())
			})

			It("should create with ignoreRecordNotFoundError true", func() {
				gl := loggorm.New(
					func() liblog.Logger { return mockLogger },
					true,
					100*time.Millisecond,
				)

				Expect(gl).ToNot(BeNil())
			})

			It("should create with zero slowThreshold", func() {
				gl := loggorm.New(
					func() liblog.Logger { return mockLogger },
					false,
					0,
				)

				Expect(gl).ToNot(BeNil())
			})

			It("should create with large slowThreshold", func() {
				gl := loggorm.New(
					func() liblog.Logger { return mockLogger },
					false,
					10*time.Second,
				)

				Expect(gl).ToNot(BeNil())
			})
		})
	})

	Describe("LogMode", func() {
		Context("with Silent level", func() {
			It("should set logger to NilLevel", func() {
				result := gormLogger.LogMode(gorlog.Silent)

				Expect(result).ToNot(BeNil())
				Expect(mockLogger.level).To(Equal(loglvl.NilLevel))
			})
		})

		Context("with Info level", func() {
			It("should set logger to InfoLevel", func() {
				result := gormLogger.LogMode(gorlog.Info)

				Expect(result).ToNot(BeNil())
				Expect(mockLogger.level).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with Warn level", func() {
			It("should set logger to WarnLevel", func() {
				result := gormLogger.LogMode(gorlog.Warn)

				Expect(result).ToNot(BeNil())
				Expect(mockLogger.level).To(Equal(loglvl.WarnLevel))
			})
		})

		Context("with Error level", func() {
			It("should set logger to ErrorLevel", func() {
				result := gormLogger.LogMode(gorlog.Error)

				Expect(result).ToNot(BeNil())
				Expect(mockLogger.level).To(Equal(loglvl.ErrorLevel))
			})
		})

		Context("with multiple level changes", func() {
			It("should handle sequential level changes", func() {
				gormLogger.LogMode(gorlog.Info)
				Expect(mockLogger.level).To(Equal(loglvl.InfoLevel))

				gormLogger.LogMode(gorlog.Warn)
				Expect(mockLogger.level).To(Equal(loglvl.WarnLevel))

				gormLogger.LogMode(gorlog.Error)
				Expect(mockLogger.level).To(Equal(loglvl.ErrorLevel))

				gormLogger.LogMode(gorlog.Silent)
				Expect(mockLogger.level).To(Equal(loglvl.NilLevel))
			})
		})
	})

	Describe("Info", func() {
		Context("with simple message", func() {
			It("should log info message", func() {
				ctx := context.Background()
				gormLogger.Info(ctx, "test message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("test message"))
			})
		})

		Context("with formatted message", func() {
			It("should log formatted info message", func() {
				ctx := context.Background()
				gormLogger.Info(ctx, "test message: %s", "value")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with multiple arguments", func() {
			It("should log info with all arguments", func() {
				ctx := context.Background()
				gormLogger.Info(ctx, "test: %s, %d, %v", "str", 42, true)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with nil context", func() {
			It("should handle nil context gracefully", func() {
				gormLogger.Info(nil, "test message")

				Expect(mockLogger.entries).To(HaveLen(1))
			})
		})
	})

	Describe("Warn", func() {
		Context("with simple message", func() {
			It("should log warn message", func() {
				ctx := context.Background()
				gormLogger.Warn(ctx, "warning message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.WarnLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("warning message"))
			})
		})

		Context("with formatted message", func() {
			It("should log formatted warn message", func() {
				ctx := context.Background()
				gormLogger.Warn(ctx, "warning: %s", "something")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.WarnLevel))
			})
		})
	})

	Describe("Error", func() {
		Context("with simple message", func() {
			It("should log error message", func() {
				ctx := context.Background()
				gormLogger.Error(ctx, "error message")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.ErrorLevel))
				Expect(mockLogger.entries[0].Message).To(Equal("error message"))
			})
		})

		Context("with formatted message", func() {
			It("should log formatted error message", func() {
				ctx := context.Background()
				gormLogger.Error(ctx, "error: %s", "details")

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.ErrorLevel))
			})
		})
	})
})
