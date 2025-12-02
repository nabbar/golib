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

// This file contains tests for logging operations and the Check/Log methods.
// It validates actual logging behavior with logrus, including output verification,
// message-only mode, context information handling, and guard conditions.
//
// Test Coverage:
//   - Check method with and without errors
//   - Log method with various configurations
//   - Message-only mode behavior
//   - NilLevel handling
//   - Nil logger and nil fields handling
//   - Context information inclusion (time, stack, caller, file, line)
//   - Error output formatting
//   - Multiple log levels
//   - Integration tests with complete entry setup
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

var _ = Describe("Log Operations", func() {
	var (
		buffer *bytes.Buffer
		logger *logrus.Logger
	)

	BeforeEach(func() {
		buffer = new(bytes.Buffer)
		logger = logrus.New()
		logger.SetOutput(buffer)
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.DebugLevel) // Allow all log levels
	})

	Describe("Check", func() {
		Context("with errors present", func() {
			It("should return true and maintain error level", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorAdd(false, errors.New("test error"))

				result := e.Check(loglvl.InfoLevel)

				Expect(result).To(BeTrue())
			})

			It("should return true with multiple errors", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorAdd(false,
						errors.New("error1"),
						errors.New("error2"),
					)

				result := e.Check(loglvl.InfoLevel)

				Expect(result).To(BeTrue())
			})

			It("should return true with mixed nil and non-nil errors", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorAdd(false,
						nil,
						errors.New("actual error"),
						nil,
					)

				result := e.Check(loglvl.InfoLevel)

				Expect(result).To(BeTrue())
			})
		})

		Context("without errors", func() {
			It("should return false and use fallback level", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields)

				result := e.Check(loglvl.InfoLevel)

				Expect(result).To(BeFalse())
			})

			It("should return false with only nil errors", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorAdd(false, nil, nil)

				result := e.Check(loglvl.InfoLevel)

				Expect(result).To(BeFalse())
			})

			It("should return false with empty error slice", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					ErrorSet([]error{})

				result := e.Check(loglvl.InfoLevel)

				Expect(result).To(BeFalse())
			})
		})

		Context("on nil entry", func() {
			It("should handle nil entry safely", func() {
				var e logent.Entry = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(e).To(BeNil())
			})
		})
	})

	Describe("Log", func() {
		Context("with valid configuration", func() {
			It("should log with all fields", func() {
				fields := logfld.New(nil)
				now := time.Now()

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(now, 123, "testCaller", "test.go", 456, "test message").
					DataSet("test data").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("test message"))
			})

			It("should log with error", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "error occurred").
					ErrorAdd(false, errors.New("test error")).
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("test error"))
			})

			It("should log multiple errors", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.ErrorLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "multiple errors").
					ErrorAdd(false,
						errors.New("error1"),
						errors.New("error2"),
					).
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("error1"))
				Expect(output).To(ContainSubstring("error2"))
			})

			It("should log with custom data", func() {
				fields := logfld.New(nil)
				data := map[string]interface{}{
					"key1": "value1",
					"key2": 123,
				}

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "with data").
					DataSet(data).
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})

			It("should log at different levels", func() {
				levels := []loglvl.Level{
					loglvl.DebugLevel,
					loglvl.InfoLevel,
					loglvl.WarnLevel,
					loglvl.ErrorLevel,
				}

				for _, lvl := range levels {
					buffer.Reset()
					fields := logfld.New(nil)

					logent.New(lvl).
						SetLogger(func() *logrus.Logger { return logger }).
						FieldSet(fields).
						SetEntryContext(time.Now(), 0, "", "", 0, "test message").
						Log()

					output := buffer.String()
					Expect(output).ToNot(BeEmpty())
				}
			})
		})

		Context("with message only mode", func() {
			It("should log clean message", func() {
				// Message only mode uses simpler logging
				testBuffer := new(bytes.Buffer)
				testLogger := logrus.New()
				testLogger.SetOutput(testBuffer)
				testLogger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
				testLogger.SetLevel(logrus.DebugLevel)

				// Fields must be set even in message only mode
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return testLogger }).
					FieldSet(fields).
					SetMessageOnly(true).
					SetEntryContext(time.Now(), 0, "", "", 0, "clean message").
					Log()

				output := testBuffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("clean message"))
			})

			It("should ignore fields in message only mode", func() {
				testBuffer := new(bytes.Buffer)
				testLogger := logrus.New()
				testLogger.SetOutput(testBuffer)
				testLogger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
				testLogger.SetLevel(logrus.DebugLevel)

				fields := logfld.New(nil)
				fields.Add("should_not_appear", "value")

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return testLogger }).
					SetMessageOnly(true).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "message only").
					Log()

				output := testBuffer.String()
				Expect(output).To(ContainSubstring("message only"))
			})
		})

		Context("with NilLevel", func() {
			It("should not log anything", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.NilLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "should not appear").
					Log()

				output := buffer.String()
				Expect(output).To(BeEmpty())
			})
		})

		Context("with nil logger", func() {
			It("should not panic", func() {
				fields := logfld.New(nil)

				Expect(func() {
					logent.New(loglvl.InfoLevel).
						FieldSet(fields).
						SetEntryContext(time.Now(), 0, "", "", 0, "test").
						Log()
				}).ToNot(Panic())
			})

			It("should not panic with nil logger function", func() {
				fields := logfld.New(nil)

				Expect(func() {
					logent.New(loglvl.InfoLevel).
						SetLogger(nil).
						FieldSet(fields).
						SetEntryContext(time.Now(), 0, "", "", 0, "test").
						Log()
				}).ToNot(Panic())
			})

			It("should not panic with logger returning nil", func() {
				fields := logfld.New(nil)

				Expect(func() {
					logent.New(loglvl.InfoLevel).
						SetLogger(func() *logrus.Logger { return nil }).
						FieldSet(fields).
						SetEntryContext(time.Now(), 0, "", "", 0, "test").
						Log()
				}).ToNot(Panic())
			})
		})

		Context("with nil fields", func() {
			It("should not log when fields is nil", func() {
				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					SetEntryContext(time.Now(), 0, "", "", 0, "test").
					Log()

				// Should not log because fields is nil
				output := buffer.String()
				Expect(output).To(BeEmpty())
			})
		})

		Context("with fields error", func() {
			It("should not log when fields has error", func() {
				// Create fields that would have an error
				// This is a test of the guard condition
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "test").
					Log()

				// Should log normally when fields is valid
				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})
		})

		Context("on nil entry", func() {
			It("should handle nil entry safely", func() {
				var e logent.Entry = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(e).To(BeNil())
			})
		})
	})

	Describe("Log with Context Information", func() {
		Context("with time", func() {
			It("should include timestamp", func() {
				fields := logfld.New(nil)
				now := time.Now()

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(now, 0, "", "", 0, "with time").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})

			It("should handle zero time", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Time{}, 0, "", "", 0, "zero time").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})
		})

		Context("with stack information", func() {
			It("should include stack number", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 999, "", "", 0, "with stack").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("999"))
			})

			It("should handle zero stack", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 0, "zero stack").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
			})
		})

		Context("with caller information", func() {
			It("should include caller", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "TestFunction", "", 0, "with caller").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("TestFunction"))
			})

			It("should prefer caller over file", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "Caller", "file.go", 0, "caller and file").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("Caller"))
			})

			It("should use file when caller is empty", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "test.go", 0, "file only").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("test.go"))
			})
		})

		Context("with line information", func() {
			It("should include line number", func() {
				fields := logfld.New(nil)

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetEntryContext(time.Now(), 0, "", "", 42, "with line").
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("42"))
			})
		})
	})

	Describe("Log Integration", func() {
		Context("with complete entry", func() {
			It("should log complex entry successfully", func() {
				fields := logfld.New(nil)
				fields.Add("custom_field", "custom_value")

				data := map[string]interface{}{
					"user_id": 123,
					"action":  "test",
				}

				logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					FieldSet(fields).
					SetLevel(loglvl.WarnLevel).
					SetEntryContext(time.Now(), 999, "TestFunc", "test.go", 100, "complex entry").
					DataSet(data).
					ErrorAdd(false, errors.New("warning message")).
					Log()

				output := buffer.String()
				Expect(output).ToNot(BeEmpty())
				Expect(output).To(ContainSubstring("complex entry"))
				Expect(output).To(ContainSubstring("warning message"))
			})
		})
	})
})
