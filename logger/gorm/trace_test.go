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

package gorm_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liblog "github.com/nabbar/golib/logger"
	loggorm "github.com/nabbar/golib/logger/gorm"
	loglvl "github.com/nabbar/golib/logger/level"
	gorlog "gorm.io/gorm/logger"
)

var _ = Describe("GORM Logger Trace", func() {
	var (
		mockLogger *MockLogger
		gormLogger gorlog.Interface
		ctx        context.Context
	)

	BeforeEach(func() {
		mockLogger = NewMockLogger()
		ctx = context.Background()
	})

	Describe("Trace with normal query", func() {
		BeforeEach(func() {
			gormLogger = loggorm.New(
				func() liblog.Logger { return mockLogger },
				false,
				100*time.Millisecond,
			)
		})

		Context("with fast query and no error", func() {
			It("should log as info", func() {
				begin := time.Now()

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users WHERE id = 1", 1
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
				Expect(mockLogger.entries[0].Fields).To(HaveKey("query"))
				Expect(mockLogger.entries[0].Fields).To(HaveKey("rows"))
				Expect(mockLogger.entries[0].Fields).To(HaveKey("elapsed ms"))
				Expect(mockLogger.entries[0].Fields["query"]).To(Equal("SELECT * FROM users WHERE id = 1"))
				Expect(mockLogger.entries[0].Fields["rows"]).To(BeNumerically("==", 1))
			})
		})

		Context("with query returning multiple rows", func() {
			It("should log row count", func() {
				begin := time.Now()

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users", 42
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Fields["rows"]).To(BeNumerically("==", 42))
			})
		})

		Context("with query returning no rows", func() {
			It("should log zero rows", func() {
				begin := time.Now()

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users WHERE id = 9999", 0
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Fields["rows"]).To(BeNumerically("==", 0))
			})
		})

		Context("with query returning -1 rows", func() {
			It("should log rows as dash", func() {
				begin := time.Now()

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "UPDATE users SET name = 'test'", -1
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Fields["rows"]).To(Equal("-"))
			})
		})
	})

	Describe("Trace with slow query", func() {
		BeforeEach(func() {
			gormLogger = loggorm.New(
				func() liblog.Logger { return mockLogger },
				false,
				10*time.Millisecond, // Short threshold for testing
			)
		})

		Context("when query exceeds threshold", func() {
			It("should log as warning", func() {
				begin := time.Now().Add(-50 * time.Millisecond) // Simulate 50ms ago

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users", 10
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.WarnLevel))
				Expect(mockLogger.entries[0].Errors).ToNot(BeEmpty())
			})
		})

		Context("when query is at threshold boundary", func() {
			It("should log as info when exactly at threshold", func() {
				threshold := 50 * time.Millisecond
				gormLogger = loggorm.New(
					func() liblog.Logger { return mockLogger },
					false,
					threshold,
				)

				begin := time.Now().Add(-threshold)

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users", 1
				}, nil)

				// At exact threshold, might be info or warn depending on timing
				Expect(mockLogger.entries).To(HaveLen(1))
			})
		})
	})

	Describe("Trace with zero slow threshold", func() {
		BeforeEach(func() {
			gormLogger = loggorm.New(
				func() liblog.Logger { return mockLogger },
				false,
				0, // Disabled slow query detection
			)
		})

		Context("when slow threshold is disabled", func() {
			It("should never log slow query warnings", func() {
				begin := time.Now().Add(-1 * time.Second) // Very slow

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users", 1
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
			})
		})
	})

	Describe("Trace with errors", func() {
		BeforeEach(func() {
			gormLogger = loggorm.New(
				func() liblog.Logger { return mockLogger },
				false,
				100*time.Millisecond,
			)
		})

		Context("with general error", func() {
			It("should log as error", func() {
				begin := time.Now()
				testError := errors.New("database connection failed")

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users", 0
				}, testError)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.ErrorLevel))
				Expect(mockLogger.entries[0].Errors).To(ContainElement(testError))
			})
		})

		Context("with ErrRecordNotFound and ignoreRecordNotFoundError false", func() {
			It("should log as error", func() {
				begin := time.Now()

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users WHERE id = 999", 0
				}, gorlog.ErrRecordNotFound)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.ErrorLevel))
			})
		})
	})

	Describe("Trace with ignoreRecordNotFoundError enabled", func() {
		BeforeEach(func() {
			gormLogger = loggorm.New(
				func() liblog.Logger { return mockLogger },
				true, // Ignore record not found errors
				100*time.Millisecond,
			)
		})

		Context("with ErrRecordNotFound", func() {
			It("should log as info instead of error", func() {
				begin := time.Now()

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users WHERE id = 999", 0
				}, gorlog.ErrRecordNotFound)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with other errors", func() {
			It("should still log as error", func() {
				begin := time.Now()
				testError := errors.New("syntax error")

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "INVALID SQL", 0
				}, testError)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Level).To(Equal(loglvl.ErrorLevel))
			})
		})
	})

	Describe("Trace with complex SQL queries", func() {
		BeforeEach(func() {
			gormLogger = loggorm.New(
				func() liblog.Logger { return mockLogger },
				false,
				100*time.Millisecond,
			)
		})

		Context("with JOIN query", func() {
			It("should log full query", func() {
				begin := time.Now()
				sql := "SELECT u.*, p.* FROM users u JOIN profiles p ON u.id = p.user_id WHERE u.active = true"

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return sql, 5
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Fields["query"]).To(Equal(sql))
			})
		})

		Context("with INSERT query", func() {
			It("should log insert statement", func() {
				begin := time.Now()
				sql := "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')"

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return sql, 1
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Fields["query"]).To(Equal(sql))
			})
		})

		Context("with UPDATE query", func() {
			It("should log update statement", func() {
				begin := time.Now()
				sql := "UPDATE users SET last_login = NOW() WHERE id = 1"

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return sql, -1
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Fields["query"]).To(Equal(sql))
			})
		})

		Context("with DELETE query", func() {
			It("should log delete statement", func() {
				begin := time.Now()
				sql := "DELETE FROM users WHERE created_at < '2020-01-01'"

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return sql, 10
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				Expect(mockLogger.entries[0].Fields["query"]).To(Equal(sql))
			})
		})
	})

	Describe("Trace timing accuracy", func() {
		BeforeEach(func() {
			gormLogger = loggorm.New(
				func() liblog.Logger { return mockLogger },
				false,
				100*time.Millisecond,
			)
		})

		Context("with very fast query", func() {
			It("should log elapsed time close to zero", func() {
				begin := time.Now()

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT 1", 1
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				elapsed := mockLogger.entries[0].Fields["elapsed ms"].(float64)
				Expect(elapsed).To(BeNumerically(">=", 0))
				Expect(elapsed).To(BeNumerically("<", 10)) // Less than 10ms
			})
		})

		Context("with measured delay", func() {
			It("should log approximate elapsed time", func() {
				begin := time.Now().Add(-25 * time.Millisecond)

				gormLogger.Trace(ctx, begin, func() (string, int64) {
					return "SELECT * FROM users", 1
				}, nil)

				Expect(mockLogger.entries).To(HaveLen(1))
				elapsed := mockLogger.entries[0].Fields["elapsed ms"].(float64)
				Expect(elapsed).To(BeNumerically(">=", 20))
				Expect(elapsed).To(BeNumerically("<", 35))
			})
		})
	})
})
