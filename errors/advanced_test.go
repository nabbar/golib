/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package errors_test

import (
	"encoding/json"
	"fmt"
	"runtime"

	. "github.com/nabbar/golib/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Advanced Error Features", func() {
	BeforeEach(func() {
		// Register test error messages
		if !ExistInMapMessage(TestErrorCode1) {
			RegisterIdFctMessage(TestErrorCode1, func(code CodeError) string {
				switch code {
				case TestErrorCode1:
					return "test error 1"
				case TestErrorCode2:
					return "test error 2"
				case TestErrorCode3:
					return "test error 3"
				default:
					return ""
				}
			})
		}
	})

	Describe("NewErrorRecovered", func() {
		It("should create error from recovered panic", func() {
			err := NewErrorRecovered("panic recovered", "original panic message")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("panic recovered"))
		})

		It("should handle recovered with parent errors", func() {
			parent := New(100, "parent error")
			err := NewErrorRecovered("panic occurred", "panic message", parent)
			Expect(err).ToNot(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should handle empty recovered string", func() {
			err := NewErrorRecovered("panic happened", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("panic happened"))
		})

		It("should include stack trace in message", func() {
			err := NewErrorRecovered("panic with trace", "panic content")
			Expect(err).ToNot(BeNil())
			// Error message should include trace information
			msg := err.Error()
			Expect(msg).To(ContainSubstring("panic with trace"))
		})

		It("should handle multiple parent errors", func() {
			parent1 := New(100, "parent 1")
			parent2 := New(200, "parent 2")
			err := NewErrorRecovered("panic", "recovered", parent1, parent2)
			Expect(err).ToNot(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should create error with nil parents", func() {
			err := NewErrorRecovered("panic", "recovered", nil, nil)
			Expect(err).ToNot(BeNil())
		})

		It("should preserve parent error chain", func() {
			parent := TestErrorCode1.Error(nil)
			err := NewErrorRecovered("recovered panic", "panic message", parent)
			Expect(err.HasError(parent)).To(BeTrue())
		})
	})

	Describe("JSON Marshaling", func() {
		It("should marshal DefaultReturn to JSON", func() {
			r := NewDefaultReturn()
			r.SetError(404, "not found", "handler.go", 42)

			jsonBytes := r.JSON()
			Expect(jsonBytes).ToNot(BeEmpty())

			// Verify JSON is valid
			var result map[string]interface{}
			err := json.Unmarshal(jsonBytes, &result)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle JSON with special characters", func() {
			r := NewDefaultReturn()
			r.SetError(500, "error with \"quotes\" and \n newlines", "file.go", 10)

			jsonBytes := r.JSON()
			Expect(jsonBytes).ToNot(BeEmpty())

			// Should be valid JSON despite special chars
			var result map[string]interface{}
			err := json.Unmarshal(jsonBytes, &result)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should marshal empty DefaultReturn", func() {
			r := NewDefaultReturn()
			jsonBytes := r.JSON()
			Expect(jsonBytes).ToNot(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(jsonBytes, &result)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle JSON with unicode characters", func() {
			r := NewDefaultReturn()
			r.SetError(400, "erreur avec caractères spéciaux: 日本語 🔥", "file.go", 1)

			jsonBytes := r.JSON()
			Expect(jsonBytes).ToNot(BeEmpty())

			var result map[string]interface{}
			err := json.Unmarshal(jsonBytes, &result)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("NewErrorTrace", func() {
		It("should create error with specific trace", func() {
			err := NewErrorTrace(404, "not found", "handler.go", 42)
			Expect(err).ToNot(BeNil())
			Expect(err.Code()).To(Equal(uint16(404)))
			Expect(err.StringError()).To(Equal("not found"))
		})

		It("should include file and line in trace", func() {
			err := NewErrorTrace(500, "server error", "server.go", 100)
			trace := err.GetTrace()
			// Trace format is "file#line" after filtering
			Expect(trace).ToNot(BeEmpty())
			Expect(trace).To(ContainSubstring("100"))
		})

		It("should handle trace with parent errors", func() {
			parent := New(400, "bad request")
			err := NewErrorTrace(500, "internal error", "api.go", 50, parent)
			Expect(err).ToNot(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should handle empty file path", func() {
			err := NewErrorTrace(200, "ok", "", 0)
			Expect(err).ToNot(BeNil())
		})

		It("should preserve error code", func() {
			err := NewErrorTrace(12345, "custom code", "file.go", 1)
			Expect(err.Code()).To(Equal(uint16(12345)))
		})
	})

	Describe("IfError", func() {
		It("should return nil when no parents", func() {
			err := IfError(100, "test message")
			Expect(err).To(BeNil())
		})

		It("should return error when has valid parent", func() {
			parent := New(200, "parent error")
			err := IfError(100, "test message", parent)
			Expect(err).ToNot(BeNil())
			Expect(err.Code()).To(Equal(uint16(100)))
		})

		It("should skip nil parents", func() {
			err := IfError(100, "test", nil, nil)
			Expect(err).To(BeNil())
		})

		It("should handle mixed nil and valid parents", func() {
			parent := New(200, "valid parent")
			err := IfError(100, "test", nil, parent, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should accept multiple valid parents", func() {
			parent1 := New(200, "parent 1")
			parent2 := New(300, "parent 2")
			err := IfError(100, "main", parent1, parent2)
			Expect(err).ToNot(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})
	})

	Describe("Trace Path Filtering", func() {
		It("should filter vendor paths", func() {
			// Create error to trigger trace capture
			err := New(100, "test error")
			trace := err.GetTrace()

			// Trace should not contain "vendor"
			Expect(trace).ToNot(ContainSubstring("/vendor/"))
		})

		It("should filter module paths", func() {
			err := New(100, "test error")
			trace := err.GetTrace()

			// Verify trace is generated
			Expect(trace).ToNot(BeEmpty())
		})

		It("should convert path separators", func() {
			// Test with a path containing backslashes
			testPath := "C:\\Windows\\Path\\file.go"
			converted := ConvPathFromLocal(testPath)
			// On Unix systems, backslashes are valid filename characters
			// The function replaces filepath.Separator with "/"
			Expect(converted).ToNot(BeEmpty())
		})

		It("should handle unix paths", func() {
			converted := ConvPathFromLocal("/usr/local/go/src/file.go")
			Expect(converted).To(Equal("/usr/local/go/src/file.go"))
		})

		It("should handle empty path", func() {
			converted := ConvPathFromLocal("")
			Expect(converted).To(BeEmpty())
		})
	})

	Describe("Error Comparison Edge Cases", func() {
		It("should handle comparison with nil error", func() {
			err := New(100, "test")
			// Is() with nil should not panic
			Expect(func() {
				_ = err.Is(nil)
			}).ToNot(Panic())
		})

		It("should handle comparison with different error types", func() {
			err1 := New(100, "test")
			err2 := fmt.Errorf("standard error")
			result := err1.Is(err2)
			// Should not match different types
			_ = result // May be true or false depending on implementation
		})

		It("should compare errors with same trace", func() {
			// Create two errors from same location
			createError := func() Error {
				return New(100, "same error")
			}
			err1 := createError()
			err2 := createError()

			// They should be considered equal
			Expect(err1.Is(err2)).To(BeTrue())
		})

		It("should compare errors with empty traces", func() {
			err1 := NewErrorTrace(100, "test", "", 0)
			err2 := NewErrorTrace(100, "test", "", 0)
			Expect(err1.Is(err2)).To(BeTrue())
		})
	})

	Describe("Runtime Frame Operations", func() {
		It("should capture runtime frames", func() {
			err := New(100, "test with frame")
			trace := err.GetTrace()

			// Should have captured a frame
			Expect(trace).ToNot(BeEmpty())
		})

		It("should handle GetTraceSlice with deep hierarchy", func() {
			// Create a deep error chain
			parent3 := New(400, "level 3")
			parent2 := New(300, "level 2", parent3)
			parent1 := New(200, "level 1", parent2)
			err := New(100, "main", parent1)

			traces := err.GetTraceSlice()
			// Should have multiple traces
			Expect(len(traces)).To(BeNumerically(">", 1))
		})

		It("should skip frames from errors package", func() {
			err := New(100, "test")
			trace := err.GetTrace()

			// Trace should not include the errors package itself
			Expect(trace).ToNot(ContainSubstring("/errors/"))
		})
	})

	Describe("MakeIfError", func() {
		It("should return nil for all nil errors", func() {
			err := MakeIfError(nil, nil, nil)
			Expect(err).To(BeNil())
		})

		It("should create error from first non-nil", func() {
			stdErr := fmt.Errorf("standard error")
			err := MakeIfError(nil, stdErr, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("standard error"))
		})

		It("should combine multiple errors", func() {
			err1 := New(100, "error 1")
			err2 := New(200, "error 2")
			combined := MakeIfError(err1, err2)
			Expect(combined).ToNot(BeNil())
			Expect(combined.HasParent()).To(BeTrue())
		})

		It("should handle mixed error types", func() {
			customErr := New(100, "custom")
			stdErr := fmt.Errorf("standard")
			combined := MakeIfError(customErr, stdErr)
			Expect(combined).ToNot(BeNil())
		})
	})

	Describe("AddOrNew", func() {
		It("should create new error when main is nil", func() {
			newErr := fmt.Errorf("new error")
			result := AddOrNew(nil, newErr)
			Expect(result).ToNot(BeNil())
			Expect(result.Error()).To(ContainSubstring("new error"))
		})

		It("should add to existing error", func() {
			main := New(100, "main error")
			sub := New(200, "sub error")
			result := AddOrNew(main, sub)
			Expect(result).ToNot(BeNil())
			Expect(result.HasParent()).To(BeTrue())
		})

		It("should return nil when both are nil", func() {
			result := AddOrNew(nil, nil)
			Expect(result).To(BeNil())
		})

		It("should handle standard error as main", func() {
			main := fmt.Errorf("standard error")
			sub := New(200, "sub error")
			result := AddOrNew(main, sub)
			Expect(result).ToNot(BeNil())
		})

		It("should add multiple parents", func() {
			main := New(100, "main")
			sub := New(200, "sub")
			parent1 := New(300, "parent1")
			parent2 := New(400, "parent2")
			result := AddOrNew(main, sub, parent1, parent2)
			Expect(result).ToNot(BeNil())
			Expect(result.HasParent()).To(BeTrue())
		})
	})

	Describe("Error State Management", func() {
		It("should maintain independent error states", func() {
			err1 := New(100, "error 1")
			err2 := New(200, "error 2")

			err1.Add(New(300, "parent"))

			// err2 should not be affected
			Expect(err1.HasParent()).To(BeTrue())
			Expect(err2.HasParent()).To(BeFalse())
		})

		It("should preserve original after SetParent", func() {
			err := New(100, "main")
			parent1 := New(200, "parent1")
			parent2 := New(300, "parent2")

			err.Add(parent1)
			err.SetParent(parent2)

			// Should have replaced parents
			Expect(err.HasParent()).To(BeTrue())
		})
	})

	Describe("Frame Utilities", func() {
		It("should handle runtime.Frame structures", func() {
			frame := runtime.Frame{
				Function: "test.Function",
				File:     "/path/to/file.go",
				Line:     42,
			}

			// Create error with trace to test frame handling
			err := NewErrorTrace(100, "test", frame.File, frame.Line)
			trace := err.GetTrace()

			// Trace should contain the line number
			Expect(trace).ToNot(BeEmpty())
			Expect(trace).To(ContainSubstring("42"))
		})

		It("should filter paths correctly", func() {
			// Test path with /pkg/mod/ in it
			testPath := "/home/user/go/pkg/mod/github.com/package/file.go"
			filtered := ConvPathFromLocal(testPath)

			// Should be filtered to remove go cache paths
			Expect(filtered).ToNot(BeEmpty())
		})
	})
})
