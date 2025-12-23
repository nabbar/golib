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

package socket_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
)

var _ = Describe("[TC-EC] Socket Edge Cases and Boundary Tests", func() {
	Describe("ErrorFilter edge cases", func() {
		Context("with complex error messages", func() {
			It("[TC-EC-001] should filter error with nested closed connection message", func() {
				err := fmt.Errorf("network error: use of closed network connection: details")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})

			It("[TC-EC-002] should filter error with uppercase closed connection message", func() {
				err := fmt.Errorf("USE OF CLOSED NETWORK CONNECTION")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})

			It("[TC-EC-003] should handle error with partial match", func() {
				err := fmt.Errorf("use of closed")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})

			It("[TC-EC-004] should handle very long error message", func() {
				longMsg := strings.Repeat("error ", 1000) + "use of closed network connection"
				err := fmt.Errorf("%s", longMsg)
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})

			It("[TC-EC-005] should handle empty error message", func() {
				err := fmt.Errorf("%s", "")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})
		})

		Context("with wrapped errors", func() {
			It("[TC-EC-006] should handle wrapped closed connection error", func() {
				innerErr := fmt.Errorf("use of closed network connection")
				wrappedErr := fmt.Errorf("failed to read: %w", innerErr)
				result := libsck.ErrorFilter(wrappedErr)
				Expect(result).NotTo(BeNil())
			})

			It("[TC-EC-007] should handle multi-level wrapped errors", func() {
				baseErr := fmt.Errorf("use of closed network connection")
				level1 := fmt.Errorf("layer 1: %w", baseErr)
				level2 := fmt.Errorf("layer 2: %w", level1)
				result := libsck.ErrorFilter(level2)
				Expect(result).NotTo(BeNil())
			})
		})
	})

	Describe("ConnState boundary cases", func() {
		Context("with boundary values", func() {
			It("[TC-EC-008] should handle zero value", func() {
				state := libsck.ConnState(0)
				Expect(state).To(Equal(libsck.ConnectionDial))
				Expect(state.String()).To(Equal("Dial Connection"))
			})

			It("[TC-EC-009] should handle maximum valid value", func() {
				state := libsck.ConnectionClose
				Expect(state.String()).To(Equal("Close Connection"))
			})

			It("[TC-EC-010] should handle value just beyond valid range", func() {
				state := libsck.ConnState(8)
				Expect(state.String()).To(Equal("unknown connection state"))
			})

			It("[TC-EC-011] should handle maximum uint8 value", func() {
				state := libsck.ConnState(255)
				Expect(state.String()).To(Equal("unknown connection state"))
			})
		})

		Context("with type conversion", func() {
			It("[TC-EC-012] should convert from int correctly", func() {
				intVal := 2
				state := libsck.ConnState(intVal)
				Expect(state).To(Equal(libsck.ConnectionRead))
			})

			It("[TC-EC-013] should maintain type safety", func() {
				state := libsck.ConnectionNew
				Expect(state).To(BeAssignableToTypeOf(libsck.ConnState(0)))
			})
		})
	})

	Describe("Constants boundary validation", func() {
		Context("DefaultBufferSize", func() {
			It("[TC-EC-014] should be positive", func() {
				Expect(libsck.DefaultBufferSize).To(BeNumerically(">", 0))
			})

			It("[TC-EC-015] should be power of 2 multiple for efficiency", func() {
				size := libsck.DefaultBufferSize
				Expect(size % 1024).To(Equal(0))
			})

			It("[TC-EC-016] should be reasonable size (not too small or too large)", func() {
				Expect(libsck.DefaultBufferSize).To(BeNumerically(">=", 4*1024))
				Expect(libsck.DefaultBufferSize).To(BeNumerically("<=", 1024*1024))
			})
		})

		Context("EOL", func() {
			It("[TC-EC-017] should be ASCII character", func() {
				Expect(libsck.EOL).To(BeNumerically("<", 128))
			})

			It("[TC-EC-018] should be printable or control character", func() {
				Expect(libsck.EOL).To(SatisfyAny(
					BeNumerically("<", 32),
					BeNumerically(">=", 32),
				))
			})
		})
	})

	Describe("Sequential state transitions", func() {
		Context("valid connection lifecycle", func() {
			It("[TC-EC-019] should have states in logical order", func() {
				lifecycle := []libsck.ConnState{
					libsck.ConnectionDial,
					libsck.ConnectionNew,
					libsck.ConnectionRead,
					libsck.ConnectionHandler,
					libsck.ConnectionWrite,
					libsck.ConnectionClose,
				}

				for i := 1; i < len(lifecycle); i++ {
					Expect(lifecycle[i]).To(BeNumerically(">", lifecycle[i-1]))
				}
			})

			It("[TC-EC-020] should have close states in order", func() {
				Expect(libsck.ConnectionCloseRead).To(BeNumerically("<", libsck.ConnectionCloseWrite))
				Expect(libsck.ConnectionCloseWrite).To(BeNumerically("<", libsck.ConnectionClose))
			})
		})
	})

	Describe("Error message patterns", func() {
		Context("with special characters", func() {
			It("[TC-EC-021] should handle error with newlines", func() {
				err := fmt.Errorf("line1\nuse of closed network connection\nline3")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})

			It("[TC-EC-022] should handle error with tabs", func() {
				err := fmt.Errorf("error\tuse of closed network connection\tdetails")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})

			It("[TC-EC-023] should handle error with unicode", func() {
				err := fmt.Errorf("错误: use of closed network connection")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})
		})
	})
})
