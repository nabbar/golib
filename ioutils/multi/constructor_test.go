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

package multi_test

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/ioutils/multi"
)

// Tests for Multi constructor and interface compliance.
// These tests verify that the New() function creates properly initialized
// instances and that the Multi type correctly implements expected interfaces.
var _ = Describe("Multi Constructor and Interface", func() {
	Describe("New constructor", func() {
		Context("creating a new Multi instance", func() {
			It("should create a new Multi instance successfully", func() {
				m := multi.New()
				Expect(m).NotTo(BeNil())
			})

			It("should implement Multi interface", func() {
				m := multi.New()
				var _ multi.Multi = m
				Expect(m).NotTo(BeNil())
			})

			It("should implement io.ReadWriteCloser", func() {
				m := multi.New()
				var _ io.ReadWriteCloser = m
				Expect(m).NotTo(BeNil())
			})

			It("should implement io.StringWriter", func() {
				m := multi.New()
				var _ io.StringWriter = m
				Expect(m).NotTo(BeNil())
			})
		})
	})

	Describe("DiscardCloser", func() {
		var d multi.DiscardCloser

		BeforeEach(func() {
			d = multi.DiscardCloser{}
		})

		Context("Read operation", func() {
			It("should read without error", func() {
				buf := make([]byte, 10)
				n, err := d.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle nil buffer gracefully", func() {
				n, err := d.Read(nil)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle empty buffer", func() {
				buf := make([]byte, 0)
				n, err := d.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})
		})

		Context("Write operation", func() {
			It("should write successfully and discard data", func() {
				data := []byte("test data")
				n, err := d.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(len(data)))
			})

			It("should handle empty write", func() {
				data := []byte{}
				n, err := d.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle nil write", func() {
				n, err := d.Write(nil)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle large write", func() {
				data := make([]byte, 1024*1024) // 1MB
				n, err := d.Write(data)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(len(data)))
			})
		})

		Context("Close operation", func() {
			It("should close without error", func() {
				err := d.Close()
				Expect(err).To(BeNil())
			})

			It("should allow multiple closes", func() {
				err1 := d.Close()
				err2 := d.Close()
				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
			})
		})

		Context("combined operations", func() {
			It("should work after close", func() {
				err := d.Close()
				Expect(err).To(BeNil())

				// Operations should still work after close
				data := []byte("test")
				n, writeErr := d.Write(data)
				Expect(writeErr).To(BeNil())
				Expect(n).To(Equal(len(data)))

				buf := make([]byte, 10)
				rn, readErr := d.Read(buf)
				Expect(readErr).To(BeNil())
				Expect(rn).To(Equal(0))
			})
		})
	})
})
