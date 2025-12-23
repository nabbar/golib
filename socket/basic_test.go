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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
)

var _ = Describe("[TC-BS] Socket Basic Tests", func() {
	Describe("Constants", func() {
		Context("DefaultBufferSize", func() {
			It("[TC-BS-001] should have correct default buffer size", func() {
				expected := 32 * 1024
				Expect(libsck.DefaultBufferSize).To(Equal(expected))
			})
		})

		Context("EOL", func() {
			It("[TC-BS-002] should be newline character", func() {
				Expect(libsck.EOL).To(Equal(byte('\n')))
			})
		})
	})

	Describe("ErrorFilter function", func() {
		Context("with nil error", func() {
			It("[TC-BS-003] should return nil", func() {
				result := libsck.ErrorFilter(nil)
				Expect(result).To(BeNil())
			})
		})

		Context("with closed connection error", func() {
			It("[TC-BS-004] should filter out closed network connection error", func() {
				err := fmt.Errorf("use of closed network connection")
				result := libsck.ErrorFilter(err)
				Expect(result).To(BeNil())
			})

			It("should filter error with closed connection message in context", func() {
				err := fmt.Errorf("read tcp 127.0.0.1:8080->127.0.0.1:54321: use of closed network connection")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
			})
		})

		Context("with normal error", func() {
			It("[TC-BS-005] should return connection timeout error", func() {
				err := fmt.Errorf("connection timeout")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
				Expect(result.Error()).To(Equal("connection timeout"))
			})

			It("should return connection refused error", func() {
				err := fmt.Errorf("connection refused")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
				Expect(result.Error()).To(Equal("connection refused"))
			})

			It("should return broken pipe error", func() {
				err := fmt.Errorf("broken pipe")
				result := libsck.ErrorFilter(err)
				Expect(result).NotTo(BeNil())
				Expect(result.Error()).To(Equal("broken pipe"))
			})
		})
	})

	Describe("ConnState enumeration", func() {
		Context("String method", func() {
			It("[TC-BS-006] should return correct string for ConnectionDial", func() {
				state := libsck.ConnectionDial
				Expect(state.String()).To(Equal("Dial Connection"))
			})

			It("[TC-BS-007] should return correct string for ConnectionNew", func() {
				state := libsck.ConnectionNew
				Expect(state.String()).To(Equal("New Connection"))
			})

			It("[TC-BS-008] should return correct string for ConnectionRead", func() {
				state := libsck.ConnectionRead
				Expect(state.String()).To(Equal("Read Incoming Stream"))
			})

			It("[TC-BS-009] should return correct string for ConnectionCloseRead", func() {
				state := libsck.ConnectionCloseRead
				Expect(state.String()).To(Equal("Close Incoming Stream"))
			})

			It("[TC-BS-010] should return correct string for ConnectionHandler", func() {
				state := libsck.ConnectionHandler
				Expect(state.String()).To(Equal("Run HandlerFunc"))
			})

			It("[TC-BS-011] should return correct string for ConnectionWrite", func() {
				state := libsck.ConnectionWrite
				Expect(state.String()).To(Equal("Write Outgoing Steam"))
			})

			It("[TC-BS-012] should return correct string for ConnectionCloseWrite", func() {
				state := libsck.ConnectionCloseWrite
				Expect(state.String()).To(Equal("Close Outgoing Stream"))
			})

			It("[TC-BS-013] should return correct string for ConnectionClose", func() {
				state := libsck.ConnectionClose
				Expect(state.String()).To(Equal("Close Connection"))
			})

			It("[TC-BS-014] should return unknown for invalid state", func() {
				state := libsck.ConnState(255)
				Expect(state.String()).To(Equal("unknown connection state"))
			})
		})

		Context("Values", func() {
			It("[TC-BS-015] should have correct value for ConnectionDial", func() {
				Expect(libsck.ConnectionDial).To(Equal(libsck.ConnState(0)))
			})

			It("[TC-BS-016] should have correct value for ConnectionNew", func() {
				Expect(libsck.ConnectionNew).To(Equal(libsck.ConnState(1)))
			})

			It("[TC-BS-017] should have correct value for ConnectionRead", func() {
				Expect(libsck.ConnectionRead).To(Equal(libsck.ConnState(2)))
			})

			It("[TC-BS-018] should have correct value for ConnectionCloseRead", func() {
				Expect(libsck.ConnectionCloseRead).To(Equal(libsck.ConnState(3)))
			})

			It("[TC-BS-019] should have correct value for ConnectionHandler", func() {
				Expect(libsck.ConnectionHandler).To(Equal(libsck.ConnState(4)))
			})

			It("[TC-BS-020] should have correct value for ConnectionWrite", func() {
				Expect(libsck.ConnectionWrite).To(Equal(libsck.ConnState(5)))
			})

			It("[TC-BS-021] should have correct value for ConnectionCloseWrite", func() {
				Expect(libsck.ConnectionCloseWrite).To(Equal(libsck.ConnState(6)))
			})

			It("[TC-BS-022] should have correct value for ConnectionClose", func() {
				Expect(libsck.ConnectionClose).To(Equal(libsck.ConnState(7)))
			})
		})

		Context("All states iteration", func() {
			It("[TC-BS-023] should have valid string representation for all standard states", func() {
				states := []libsck.ConnState{
					libsck.ConnectionDial,
					libsck.ConnectionNew,
					libsck.ConnectionRead,
					libsck.ConnectionCloseRead,
					libsck.ConnectionHandler,
					libsck.ConnectionWrite,
					libsck.ConnectionCloseWrite,
					libsck.ConnectionClose,
				}

				for _, state := range states {
					str := state.String()
					Expect(str).NotTo(BeEmpty())
					Expect(str).NotTo(Equal("unknown connection state"))
				}
			})
		})
	})
})
