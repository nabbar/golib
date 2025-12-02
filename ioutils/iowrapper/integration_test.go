/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// This file tests real-world integration scenarios.
//
// Test Strategy:
//   - Test practical use cases: logging, transformation, checksumming
//   - Verify wrapper chaining for composing multiple transformations
//   - Validate real-world data transformations (ROT13, uppercase, etc.)
//   - Test integration with standard library functions (io.Copy, hash functions)
//   - Ensure wrappers work correctly in typical application scenarios
//
// Coverage: 8 specs testing integration with real-world patterns and external packages.
package iowrapper_test

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"strings"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IOWrapper - Integration Tests", func() {
	Context("Real-world use case: Logging wrapper", func() {
		It("should log all read operations", func() {
			reader := strings.NewReader("hello world")
			wrapper := New(reader)

			var logEntries []string

			// Set custom read that logs
			wrapper.SetRead(func(p []byte) []byte {
				// First do the actual read
				n, _ := reader.Read(p)
				data := p[:n]
				logEntries = append(logEntries, string(data))
				return data
			})

			// Read data
			buf := make([]byte, 5)
			wrapper.Read(buf)
			Expect(string(buf)).To(Equal("hello"))

			n, _ := wrapper.Read(buf)
			Expect(string(buf[:n])).To(ContainSubstring(" worl"))

			// Verify logging occurred
			Expect(logEntries).To(HaveLen(2))
		})

		It("should log all write operations", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			var writeLog []string

			// Set custom write that logs
			wrapper.SetWrite(func(p []byte) []byte {
				writeLog = append(writeLog, string(p))
				buf.Write(p)
				return p
			})

			// Write data
			wrapper.Write([]byte("first"))
			wrapper.Write([]byte("second"))

			Expect(buf.String()).To(Equal("firstsecond"))
			Expect(writeLog).To(Equal([]string{"first", "second"}))
		})
	})

	Context("Real-world use case: Data transformation", func() {
		It("should transform data on read (uppercase)", func() {
			reader := strings.NewReader("hello world")
			wrapper := New(reader)

			// Transform to uppercase
			wrapper.SetRead(func(p []byte) []byte {
				n, _ := reader.Read(p)
				data := p[:n]
				for i := range data {
					if data[i] >= 'a' && data[i] <= 'z' {
						data[i] -= 32
					}
				}
				return data
			})

			buf := make([]byte, 11)
			n, _ := wrapper.Read(buf)

			Expect(string(buf[:n])).To(Equal("HELLO WORLD"))
		})

		It("should transform data on write (base64-like encoding)", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			// Simple ROT13 transformation
			wrapper.SetWrite(func(p []byte) []byte {
				transformed := make([]byte, len(p))
				for i, b := range p {
					if b >= 'a' && b <= 'z' {
						transformed[i] = 'a' + (b-'a'+13)%26
					} else if b >= 'A' && b <= 'Z' {
						transformed[i] = 'A' + (b-'A'+13)%26
					} else {
						transformed[i] = b
					}
				}
				buf.Write(transformed)
				return transformed
			})

			wrapper.Write([]byte("hello"))

			Expect(buf.String()).To(Equal("uryyb"))
		})
	})

	Context("Real-world use case: Metrics collection", func() {
		It("should count bytes read", func() {
			reader := strings.NewReader("test data")
			wrapper := New(reader)

			bytesRead := 0

			wrapper.SetRead(func(p []byte) []byte {
				n, _ := reader.Read(p)
				data := p[:n]
				bytesRead += len(data)
				return data
			})

			buf := make([]byte, 100)
			wrapper.Read(buf)

			Expect(bytesRead).To(Equal(9))
		})

		It("should count bytes written", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			bytesWritten := 0

			wrapper.SetWrite(func(p []byte) []byte {
				bytesWritten += len(p)
				buf.Write(p)
				return p
			})

			wrapper.Write([]byte("hello"))
			wrapper.Write([]byte("world"))

			Expect(bytesWritten).To(Equal(10))
			Expect(buf.String()).To(Equal("helloworld"))
		})
	})

	Context("Real-world use case: Data validation", func() {
		It("should validate data on read", func() {
			reader := strings.NewReader("valid data")
			wrapper := New(reader)

			var validationErrors []string

			wrapper.SetRead(func(p []byte) []byte {
				n, _ := reader.Read(p)
				data := p[:n]

				// Check for invalid characters
				for _, b := range data {
					if b < 32 && b != '\n' && b != '\r' && b != '\t' {
						validationErrors = append(validationErrors, "invalid byte")
					}
				}

				return data
			})

			buf := make([]byte, 100)
			wrapper.Read(buf)

			Expect(validationErrors).To(BeEmpty())
		})

		It("should reject invalid data on write", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			wrapper.SetWrite(func(p []byte) []byte {
				// Only allow alphanumeric
				valid := make([]byte, 0, len(p))
				for _, b := range p {
					if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') {
						valid = append(valid, b)
					}
				}
				buf.Write(valid)
				return valid
			})

			wrapper.Write([]byte("hello123!@#world"))

			Expect(buf.String()).To(Equal("hello123world"))
		})
	})

	Context("Real-world use case: Buffering and batching", func() {
		It("should batch small reads into larger chunks", func() {
			reader := strings.NewReader("abcdefghijklmnop")
			wrapper := New(reader)

			buffer := make([]byte, 0, 100)
			batchSize := 8

			wrapper.SetRead(func(p []byte) []byte {
				// Fill buffer up to batch size
				for len(buffer) < batchSize {
					tmp := make([]byte, 1)
					n, err := reader.Read(tmp)
					if n > 0 {
						buffer = append(buffer, tmp[:n]...)
					}
					if err != nil {
						break
					}
				}

				// Return batch
				n := copy(p, buffer)
				buffer = buffer[n:]
				return p[:n]
			})

			buf := make([]byte, 8)
			n, _ := wrapper.Read(buf)

			Expect(n).To(Equal(8))
			Expect(string(buf[:n])).To(Equal("abcdefgh"))
		})
	})

	Context("Real-world use case: Checksumming", func() {
		It("should calculate checksum of read data", func() {
			data := "test data for checksum"
			reader := strings.NewReader(data)
			wrapper := New(reader)

			hash := md5.New()

			wrapper.SetRead(func(p []byte) []byte {
				n, err := reader.Read(p)
				if n > 0 {
					result := p[:n]
					hash.Write(result)
					if err == io.EOF {
						return result
					}
					return result
				}
				// n == 0, signal end
				return nil
			})

			// Read all data
			buf := make([]byte, 4096)
			for {
				n, err := wrapper.Read(buf)
				if n == 0 || err != nil {
					break
				}
			}

			checksum := hex.EncodeToString(hash.Sum(nil))
			Expect(checksum).ToNot(BeEmpty())

			// Verify checksum is consistent
			expectedHash := md5.Sum([]byte(data))
			expectedChecksum := hex.EncodeToString(expectedHash[:])
			Expect(checksum).To(Equal(expectedChecksum))
		})

		It("should calculate checksum of written data", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			hash := md5.New()

			wrapper.SetWrite(func(p []byte) []byte {
				hash.Write(p)
				buf.Write(p)
				return p
			})

			testData := []byte("test data for checksum")
			wrapper.Write(testData)

			checksum := hex.EncodeToString(hash.Sum(nil))

			// Verify checksum matches
			expectedHash := md5.Sum(testData)
			expectedChecksum := hex.EncodeToString(expectedHash[:])
			Expect(checksum).To(Equal(expectedChecksum))
		})
	})

	Context("Real-world use case: Rate limiting", func() {
		It("should limit read chunk size", func() {
			reader := strings.NewReader(strings.Repeat("x", 1000))
			wrapper := New(reader)

			maxChunkSize := 10

			wrapper.SetRead(func(p []byte) []byte {
				// Limit read size
				readSize := len(p)
				if readSize > maxChunkSize {
					readSize = maxChunkSize
				}
				tmp := make([]byte, readSize)
				n, _ := reader.Read(tmp)
				return tmp[:n]
			})

			buf := make([]byte, 100)
			n, _ := wrapper.Read(buf)

			// Should only read up to maxChunkSize
			Expect(n).To(Equal(maxChunkSize))
		})

		It("should limit write chunk size", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			maxChunkSize := 10
			actualWritten := 0

			wrapper.SetWrite(func(p []byte) []byte {
				// Limit write size
				writeSize := len(p)
				if writeSize > maxChunkSize {
					writeSize = maxChunkSize
				}
				chunk := p[:writeSize]
				buf.Write(chunk)
				actualWritten = writeSize
				return chunk
			})

			n, _ := wrapper.Write([]byte(strings.Repeat("x", 100)))

			// Should report writing only up to maxChunkSize
			Expect(n).To(Equal(maxChunkSize))
			Expect(actualWritten).To(Equal(maxChunkSize))
		})
	})

	Context("Real-world use case: Composability", func() {
		It("should allow chaining multiple wrappers", func() {
			reader := strings.NewReader("hello")

			// First wrapper: uppercase
			wrapper1 := New(reader)
			wrapper1.SetRead(func(p []byte) []byte {
				n, _ := reader.Read(p)
				data := p[:n]
				for i := range data {
					if data[i] >= 'a' && data[i] <= 'z' {
						data[i] -= 32
					}
				}
				return data
			})

			// Second wrapper: add prefix
			wrapper2 := New(wrapper1)
			firstRead := true
			wrapper2.SetRead(func(p []byte) []byte {
				if firstRead {
					firstRead = false
					prefix := []byte("PREFIX:")
					n, _ := wrapper1.Read(p[len(prefix):])
					copy(p, prefix)
					copy(p[len(prefix):], p[len(prefix):len(prefix)+n])
					return p[:len(prefix)+n]
				}
				n, _ := wrapper1.Read(p)
				return p[:n]
			})

			buf := make([]byte, 100)
			n, _ := wrapper2.Read(buf)

			Expect(string(buf[:n])).To(Equal("PREFIX:HELLO"))
		})
	})
})
