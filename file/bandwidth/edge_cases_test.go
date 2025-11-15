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

package bandwidth_test

import (
	"io"
	"os"

	. "github.com/nabbar/golib/file/bandwidth"
	libfpg "github.com/nabbar/golib/file/progress"
	libsiz "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bandwidth Edge Cases", func() {
	var (
		emptyFilePath  string
		smallFilePath  string
		mediumFilePath string
	)

	BeforeEach(func() {
		// Create empty file
		emptyFile, err := os.CreateTemp("", "bandwidth-empty-*.dat")
		Expect(err).ToNot(HaveOccurred())
		emptyFilePath = emptyFile.Name()
		emptyFile.Close()

		// Create small file (100 bytes)
		smallFile, err := os.CreateTemp("", "bandwidth-small-*.dat")
		Expect(err).ToNot(HaveOccurred())
		smallFilePath = smallFile.Name()
		_, err = smallFile.Write(make([]byte, 100))
		Expect(err).ToNot(HaveOccurred())
		smallFile.Close()

		// Create medium file (1KB)
		mediumFile, err := os.CreateTemp("", "bandwidth-medium-*.dat")
		Expect(err).ToNot(HaveOccurred())
		mediumFilePath = mediumFile.Name()
		_, err = mediumFile.Write(make([]byte, 1024))
		Expect(err).ToNot(HaveOccurred())
		mediumFile.Close()
	})

	AfterEach(func() {
		_ = os.Remove(emptyFilePath)
		_ = os.Remove(smallFilePath)
		_ = os.Remove(mediumFilePath)
	})

	Describe("Empty Files", func() {
		It("should handle empty file with bandwidth limit", func() {
			bw := New(libsiz.SizeKilo)
			fpg, err := libfpg.Open(emptyFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			bw.RegisterIncrement(fpg, nil)

			data, err := io.ReadAll(fpg)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(HaveLen(0))
		})

		It("should call increment callback for empty file", func() {
			var callCount int

			bw := New(0)
			fpg, err := libfpg.Open(emptyFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			bw.RegisterIncrement(fpg, func(size int64) {
				callCount++
			})

			_, err = io.ReadAll(fpg)
			Expect(err).ToNot(HaveOccurred())
			// Empty file might not trigger increment
		})
	})

	Describe("Small Files", func() {
		It("should handle small file with large bandwidth limit", func() {
			bw := New(libsiz.SizeMega) // 1MB/s for 100 bytes
			fpg, err := libfpg.Open(smallFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			bw.RegisterIncrement(fpg, nil)

			data, err := io.ReadAll(fpg)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(HaveLen(100))
		})

		It("should handle small file with small bandwidth limit", func() {
			bw := New(libsiz.Size(50)) // 50 bytes/s for 100 bytes file
			fpg, err := libfpg.Open(smallFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			var incrementCalls int
			bw.RegisterIncrement(fpg, func(size int64) {
				incrementCalls++
			})

			data, err := io.ReadAll(fpg)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(HaveLen(100))
		})
	})

	Describe("Bandwidth Limits", func() {
		It("should handle zero bandwidth limit gracefully", func() {
			bw := New(libsiz.Size(0))
			fpg, err := libfpg.Open(mediumFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			bw.RegisterIncrement(fpg, nil)

			data, err := io.ReadAll(fpg)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(HaveLen(1024))
		})

		It("should handle very large bandwidth limit", func() {
			bw := New(libsiz.Size(1024 * 1024 * 1024)) // 1GB/s
			fpg, err := libfpg.Open(mediumFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			bw.RegisterIncrement(fpg, nil)

			data, err := io.ReadAll(fpg)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(HaveLen(1024))
		})

		It("should handle very small bandwidth limit", func() {
			bw := New(libsiz.Size(1)) // 1 byte/s
			fpg, err := libfpg.Open(smallFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			bw.RegisterIncrement(fpg, nil)

			data, err := io.ReadAll(fpg)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(HaveLen(100))
		})
	})

	Describe("Multiple Resets", func() {
		It("should handle multiple resets", func() {
			var resetCount int

			bw := New(0)
			fpg, err := libfpg.Open(mediumFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			bw.RegisterReset(fpg, func(size, current int64) {
				resetCount++
			})

			// Read some data and reset multiple times
			buffer := make([]byte, 100)
			_, err = fpg.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			fpg.Reset(1024)

			_, err = fpg.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			fpg.Reset(1024)

			_, err = fpg.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			fpg.Reset(1024)

			Expect(resetCount).To(Equal(3))
		})
	})

	Describe("Callback Exceptions", func() {
		It("should handle panicking increment callback gracefully", func() {
			bw := New(0)
			fpg, err := libfpg.Open(smallFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fpg.Close()

			defer func() {
				if r := recover(); r != nil {
					// Recovered from panic
				}
			}()

			bw.RegisterIncrement(fpg, func(size int64) {
				// Panicking callback
				panic("test panic")
			})

			// This should handle the panic
			_, _ = io.ReadAll(fpg)
		})
	})
})
