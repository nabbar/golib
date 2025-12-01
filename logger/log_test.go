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
	"errors"
	"time"

	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger Logging Methods", func() {
	var log liblog.Logger

	BeforeEach(func() {
		log = liblog.New(GetContext())
		log.SetLevel(loglvl.DebugLevel)
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

	Describe("Debug method", func() {
		Context("when logging debug message", func() {
			It("should not panic", func() {
				Expect(func() {
					log.Debug("Debug message", nil)
				}).ToNot(Panic())
			})

			It("should accept data and arguments", func() {
				Expect(func() {
					log.Debug("Debug with data", map[string]interface{}{"key": "value"})
				}).ToNot(Panic())
			})

			It("should handle formatted strings with args", func() {
				Expect(func() {
					log.Debug("Debug formatted: %s", nil, "test")
				}).ToNot(Panic())
			})
		})
	})

	Describe("Info method", func() {
		Context("when logging info message", func() {
			It("should not panic", func() {
				Expect(func() {
					log.Info("Info message", nil)
				}).ToNot(Panic())
			})

			It("should accept data and arguments", func() {
				Expect(func() {
					log.Info("Info with data", map[string]string{"status": "ok"})
				}).ToNot(Panic())
			})
		})
	})

	Describe("Warning method", func() {
		Context("when logging warning message", func() {
			It("should not panic", func() {
				Expect(func() {
					log.Warning("Warning message", nil)
				}).ToNot(Panic())
			})

			It("should accept data and arguments", func() {
				Expect(func() {
					log.Warning("Warning with data", map[string]interface{}{"severity": "high"})
				}).ToNot(Panic())
			})
		})
	})

	Describe("Error method", func() {
		Context("when logging error message", func() {
			It("should not panic", func() {
				Expect(func() {
					log.Error("Error message", nil)
				}).ToNot(Panic())
			})

			It("should accept data and arguments", func() {
				Expect(func() {
					log.Error("Error with data", map[string]interface{}{"code": 500})
				}).ToNot(Panic())
			})

			It("should handle actual errors as data", func() {
				testErr := errors.New("test error")
				Expect(func() {
					log.Error("Error occurred", testErr)
				}).ToNot(Panic())
			})
		})
	})

	Describe("CheckError method", func() {
		Context("when error is nil", func() {
			It("should return false and not log", func() {
				result := log.CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "Should not appear")
				Expect(result).To(BeFalse())
			})
		})

		Context("when error is not nil", func() {
			It("should return true and log the error", func() {
				testErr := errors.New("test error for logging")
				result := log.CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "Error occurred", testErr)
				Expect(result).To(BeTrue())
			})

			It("should handle multiple errors", func() {
				err1 := errors.New("error 1")
				err2 := errors.New("error 2")
				result := log.CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "Multiple errors", err1, err2)
				Expect(result).To(BeTrue())
			})
		})

		Context("with success level", func() {
			It("should log OK when no error and lvlOK is set", func() {
				result := log.CheckError(loglvl.ErrorLevel, loglvl.InfoLevel, "Operation completed")
				Expect(result).To(BeFalse())
			})
		})
	})

	Describe("Access method", func() {
		Context("when logging access message", func() {
			It("should create access entry", func() {
				entry := log.Access("192.168.1.1", "user1", time.Now(), 100*time.Millisecond,
					"GET", "/api/users", "HTTP/1.1", 200, 1024)
				Expect(entry).ToNot(BeNil())
			})

			It("should handle different status codes", func() {
				statuses := []int{200, 201, 404, 500}
				for _, status := range statuses {
					entry := log.Access("127.0.0.1", "", time.Now(), 50*time.Millisecond,
						"GET", "/test", "HTTP/1.1", status, 512)
					Expect(entry).ToNot(BeNil())
				}
			})

			It("should accept empty remote user", func() {
				entry := log.Access("10.0.0.1", "", time.Now(), 25*time.Millisecond,
					"POST", "/api/data", "HTTP/2.0", 201, 2048)
				Expect(entry).ToNot(BeNil())
			})
		})
	})

	Describe("Entry method", func() {
		Context("when creating entry", func() {
			It("should return valid entry for debug level", func() {
				entry := log.Entry(loglvl.DebugLevel, "Test message")
				Expect(entry).ToNot(BeNil())
			})

			It("should return valid entry for info level", func() {
				entry := log.Entry(loglvl.InfoLevel, "Test message")
				Expect(entry).ToNot(BeNil())
			})

			It("should return valid entry for warning level", func() {
				entry := log.Entry(loglvl.WarnLevel, "Test message")
				Expect(entry).ToNot(BeNil())
			})

			It("should return valid entry for error level", func() {
				entry := log.Entry(loglvl.ErrorLevel, "Test message")
				Expect(entry).ToNot(BeNil())
			})

			It("should allow chaining field additions", func() {
				entry := log.Entry(loglvl.InfoLevel, "Test message")
				Expect(func() {
					entry.FieldAdd("key", "value").Log()
				}).ToNot(Panic())
			})
		})
	})

	Describe("LogDetails method", func() {
		Context("with minimal parameters", func() {
			It("should log successfully", func() {
				Expect(func() {
					log.LogDetails(loglvl.InfoLevel, "Test message", nil, nil, nil)
				}).ToNot(Panic())
			})
		})

		Context("with error parameter", func() {
			It("should include error in log", func() {
				testErr := errors.New("test error")
				Expect(func() {
					log.LogDetails(loglvl.ErrorLevel, "Error occurred", testErr, nil, nil)
				}).ToNot(Panic())
			})
		})

		Context("with multiple errors", func() {
			It("should handle slice of errors", func() {
				errs := []error{
					errors.New("error 1"),
					errors.New("error 2"),
					errors.New("error 3"),
				}
				Expect(func() {
					log.LogDetails(loglvl.ErrorLevel, "Multiple errors", errs, nil, nil)
				}).ToNot(Panic())
			})
		})

		Context("with formatted message", func() {
			It("should handle format and arguments", func() {
				Expect(func() {
					log.LogDetails(loglvl.InfoLevel, "Formatted: %s, %d", nil, nil, nil, "test", 123)
				}).ToNot(Panic())
			})
		})
	})
})
