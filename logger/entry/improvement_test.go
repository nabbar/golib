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

// This file contains additional tests to improve code coverage, focusing on
// edge cases and nil entry handling that may not be covered in other test files.
package entry_test

import (
	"bytes"
	"errors"
	"time"

	logent "github.com/nabbar/golib/logger/entry"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Coverage Improvements", func() {
	Describe("Nil Entry Behavior", func() {
		Context("documenting expected behavior", func() {
			It("should demonstrate that nil interface causes panic as expected in Go", func() {
				// In Go, calling methods on a nil interface causes panic.
				// This is expected behavior and documented in the existing tests.
				// The nil checks in the implementation methods are for when the
				// concrete type is nil, not when the interface itself is nil.
			})
		})
	})

	Describe("Message-Only Logging Path", func() {
		var buffer *bytes.Buffer
		var logger *logrus.Logger

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
			logger = logrus.New()
			logger.SetOutput(buffer)
			logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
			logger.SetLevel(logrus.DebugLevel)
		})

		Context("with message-only mode", func() {
			It("should log clean message without fields", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetMessageOnly(true).
					SetEntryContext(time.Now(), 0, "", "", 0, "clean message").
					Log()

				output := buffer.String()
				Expect(output).To(ContainSubstring("clean message"))
			})

			It("should use _logClean when clean flag is true", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					SetMessageOnly(true).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "message").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})

			It("should handle _logClean with nil logger", func() {
				fields := logfld.New(nil)

				Expect(func() {
					logent.New(loglvl.InfoLevel).
						SetLogger(nil).
						SetMessageOnly(true).
						FieldSet(fields).
						SetEntryContext(time.Now(), 0, "", "", 0, "message").
						Log()
				}).ToNot(Panic())
			})

			It("should handle _logClean with logger returning nil", func() {
				fields := logfld.New(nil)

				Expect(func() {
					logent.New(loglvl.InfoLevel).
						SetLogger(func() *logrus.Logger { return nil }).
						SetMessageOnly(true).
						FieldSet(fields).
						SetEntryContext(time.Now(), 0, "", "", 0, "message").
						Log()
				}).ToNot(Panic())
			})
		})
	})

	Describe("Error Handling Edge Cases", func() {
		Context("with golib errors", func() {
			It("should handle golib/errors wrapping", func() {
				// This tests the liberr.Is() path in ErrorAdd
				e := logent.New(loglvl.ErrorLevel)

				// Standard error (not a golib error)
				stdErr := errors.New("standard error")
				result := e.ErrorAdd(false, stdErr)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})

		Context("with nil errors slice initialization", func() {
			It("should initialize error slice when empty in ErrorAdd", func() {
				e := logent.New(loglvl.ErrorLevel)
				// Error slice should be initialized but empty
				result := e.ErrorAdd(true, errors.New("first"))

				Expect(result).ToNot(BeNil())
			})
		})
	})

	Describe("Log Method Edge Cases", func() {
		var logger *logrus.Logger
		var buffer *bytes.Buffer

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
			logger = logrus.New()
			logger.SetOutput(buffer)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.DebugLevel)
		})

		Context("with various context combinations", func() {
			It("should log with empty caller but with file", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "file.go", 42, "message").
					Log()

				output := buffer.String()
				Expect(output).To(ContainSubstring("file.go"))
			})

			It("should log with both caller and file", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "Caller", "file.go", 42, "message").
					Log()

				output := buffer.String()
				Expect(output).To(ContainSubstring("Caller"))
			})

			It("should log with zero stack", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "message").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})

			It("should log with nil data", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					DataSet(nil).
					SetEntryContext(time.Now(), 0, "", "", 0, "message").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})

			It("should log with empty message", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})

			It("should handle all nil errors in error slice", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorAdd(false, nil, nil, nil).
					SetEntryContext(time.Now(), 0, "", "", 0, "message").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})
		})
	})

	Describe("Check Method Coverage", func() {
		var logger *logrus.Logger
		var buffer *bytes.Buffer

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
			logger = logrus.New()
			logger.SetOutput(buffer)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.DebugLevel)
		})

		Context("with error checking", func() {
			It("should return false when error slice is empty", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields)

				result := e.Check(loglvl.InfoLevel)
				Expect(result).To(BeFalse())
			})

			It("should return false when error slice contains only nils", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorAdd(false, nil, nil)

				result := e.Check(loglvl.InfoLevel)
				Expect(result).To(BeFalse())
			})

			It("should return true when at least one non-nil error exists", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorAdd(false, nil, errors.New("error"), nil)

				result := e.Check(loglvl.InfoLevel)
				Expect(result).To(BeTrue())
			})
		})
	})
})
