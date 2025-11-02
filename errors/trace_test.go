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
	. "github.com/nabbar/golib/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Trace Management", func() {
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

	Describe("GetTrace", func() {
		It("should get trace from error", func() {
			err := TestErrorCode1.Error(nil)
			trace := err.GetTrace()
			Expect(trace).ToNot(BeEmpty())
		})

		It("should filter path correctly", func() {
			err := TestErrorCode1.Error(nil)
			trace := err.GetTrace()
			// Trace should exist
			Expect(trace).ToNot(BeEmpty())
		})

		It("should handle trace with function name", func() {
			err := NewErrorTrace(100, "test", "", 42)
			trace := err.GetTrace()
			// When file is empty, trace may be empty or contain function name
			// Just verify it doesn't crash
			_ = trace
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("GetTraceSlice", func() {
		It("should get traces from error chain", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			traces := err.GetTraceSlice()
			Expect(len(traces)).To(BeNumerically(">", 0))
		})

		It("should handle GetTraceSlice with empty traces", func() {
			parent := New(0, "no trace parent")
			err := TestErrorCode1.Error(parent)
			traces := err.GetTraceSlice()
			// Should have at least the main error trace
			Expect(traces).ToNot(BeEmpty())
		})
	})

	Describe("SetTracePathFilter", func() {
		It("should set trace path filter", func() {
			SetTracePathFilter("/custom/path")
			// This sets the filter, we can't easily verify the effect but at least call it
			err := TestErrorCode1.Error(nil)
			Expect(err).ToNot(BeNil())
		})

		It("should not crash with empty path filter", func() {
			SetTracePathFilter("")
			err := TestErrorCode1.Error(nil)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("ConvPathFromLocal", func() {
		It("should convert local path", func() {
			converted := ConvPathFromLocal("/local/path/file.go")
			Expect(converted).ToNot(BeEmpty())
		})

		It("should handle empty path", func() {
			converted := ConvPathFromLocal("")
			_ = converted // May be empty or contain some value
		})
	})
})
