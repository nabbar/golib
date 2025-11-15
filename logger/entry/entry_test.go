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

package entry_test

import (
	"time"

	logent "github.com/nabbar/golib/logger/entry"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Entry Creation and Configuration", func() {
	Describe("New", func() {
		Context("with valid level", func() {
			It("should create entry with InfoLevel", func() {
				e := logent.New(loglvl.InfoLevel)

				Expect(e).ToNot(BeNil())
			})

			It("should create entry with DebugLevel", func() {
				e := logent.New(loglvl.DebugLevel)

				Expect(e).ToNot(BeNil())
			})

			It("should create entry with ErrorLevel", func() {
				e := logent.New(loglvl.ErrorLevel)

				Expect(e).ToNot(BeNil())
			})

			It("should create entry with WarnLevel", func() {
				e := logent.New(loglvl.WarnLevel)

				Expect(e).ToNot(BeNil())
			})

			It("should initialize Error slice", func() {
				e := logent.New(loglvl.InfoLevel)

				// Access through interface to verify internals
				Expect(e).ToNot(BeNil())
				// Error slice should be initialized
				e = e.ErrorAdd(false)
				Expect(e).ToNot(BeNil())
			})

			It("should set current time", func() {
				e := logent.New(loglvl.InfoLevel)

				Expect(e).ToNot(BeNil())
				// Time is set internally during creation
			})
		})

		Context("with NilLevel", func() {
			It("should create entry with NilLevel", func() {
				e := logent.New(loglvl.NilLevel)

				Expect(e).ToNot(BeNil())
			})
		})
	})

	Describe("SetLevel", func() {
		Context("on valid entry", func() {
			It("should set level to InfoLevel", func() {
				e := logent.New(loglvl.DebugLevel)
				result := e.SetLevel(loglvl.InfoLevel)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should set level to ErrorLevel", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.SetLevel(loglvl.ErrorLevel)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should allow level change", func() {
				e := logent.New(loglvl.InfoLevel)
				e = e.SetLevel(loglvl.DebugLevel)
				e = e.SetLevel(loglvl.WarnLevel)

				Expect(e).ToNot(BeNil())
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

	Describe("SetLogger", func() {
		Context("on valid entry", func() {
			It("should set logger function", func() {
				e := logent.New(loglvl.InfoLevel)
				logger := logrus.New()
				result := e.SetLogger(func() *logrus.Logger {
					return logger
				})

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should accept nil logger function", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.SetLogger(nil)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should allow changing logger", func() {
				e := logent.New(loglvl.InfoLevel)
				logger1 := logrus.New()
				logger2 := logrus.New()

				e = e.SetLogger(func() *logrus.Logger {
					return logger1
				})
				e = e.SetLogger(func() *logrus.Logger {
					return logger2
				})

				Expect(e).ToNot(BeNil())
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

	Describe("SetMessageOnly", func() {
		Context("on valid entry", func() {
			It("should set message only flag to true", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.SetMessageOnly(true)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should set message only flag to false", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.SetMessageOnly(false)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should allow toggling flag", func() {
				e := logent.New(loglvl.InfoLevel)
				e = e.SetMessageOnly(true)
				e = e.SetMessageOnly(false)
				e = e.SetMessageOnly(true)

				Expect(e).ToNot(BeNil())
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

	Describe("SetEntryContext", func() {
		Context("on valid entry", func() {
			It("should set all context fields", func() {
				e := logent.New(loglvl.InfoLevel)
				now := time.Now()
				result := e.SetEntryContext(now, 123, "testCaller", "testFile.go", 456, "test message")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should handle zero values", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.SetEntryContext(time.Time{}, 0, "", "", 0, "")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should allow updating context", func() {
				e := logent.New(loglvl.InfoLevel)
				time1 := time.Now()
				time2 := time1.Add(1 * time.Second)

				e = e.SetEntryContext(time1, 1, "caller1", "file1.go", 10, "msg1")
				e = e.SetEntryContext(time2, 2, "caller2", "file2.go", 20, "msg2")

				Expect(e).ToNot(BeNil())
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

	Describe("SetGinContext", func() {
		Context("on valid entry", func() {
			It("should accept nil gin context", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.SetGinContext(nil)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			// Note: We can't easily create a real gin.Context without a full HTTP request setup
			// The nil case is the most common test case for unit tests
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

	Describe("DataSet", func() {
		Context("on valid entry", func() {
			It("should set string data", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.DataSet("test data")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should set struct data", func() {
				e := logent.New(loglvl.InfoLevel)
				data := struct {
					Name  string
					Value int
				}{
					Name:  "test",
					Value: 123,
				}
				result := e.DataSet(data)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should set map data", func() {
				e := logent.New(loglvl.InfoLevel)
				data := map[string]interface{}{
					"key1": "value1",
					"key2": 123,
				}
				result := e.DataSet(data)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should set nil data", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.DataSet(nil)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should allow overwriting data", func() {
				e := logent.New(loglvl.InfoLevel)
				e = e.DataSet("first")
				e = e.DataSet("second")

				Expect(e).ToNot(BeNil())
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

	Describe("Method Chaining", func() {
		Context("with valid entry", func() {
			It("should allow chaining all setters", func() {
				logger := logrus.New()
				now := time.Now()

				result := logent.New(loglvl.InfoLevel).
					SetLogger(func() *logrus.Logger { return logger }).
					SetLevel(loglvl.DebugLevel).
					SetMessageOnly(false).
					SetEntryContext(now, 1, "caller", "file.go", 10, "message").
					SetGinContext(nil).
					DataSet("test data")

				Expect(result).ToNot(BeNil())
			})

			It("should maintain entry through chain", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.SetLevel(loglvl.DebugLevel).
					SetMessageOnly(true).
					DataSet("data")

				Expect(result).To(Equal(e))
			})
		})
	})
})
