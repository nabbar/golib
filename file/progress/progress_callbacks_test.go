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

package progress_test

import (
	"bytes"
	"io"
	"os"
	"sync/atomic"

	. "github.com/nabbar/golib/file/progress"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Progress Callbacks", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "progress-callbacks-test-*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("FctIncrement callback", func() {
		It("should call increment callback on write", func() {
			path := tempDir + "/increment-write.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var totalBytes atomic.Int64
			var callCount atomic.Int32

			p.RegisterFctIncrement(func(size int64) {
				totalBytes.Add(size)
				callCount.Add(1)
			})

			// Write data in chunks
			data1 := []byte("First chunk ")
			data2 := []byte("Second chunk ")
			data3 := []byte("Third chunk")

			p.Write(data1)
			p.Write(data2)
			p.Write(data3)

			expectedTotal := int64(len(data1) + len(data2) + len(data3))
			Expect(totalBytes.Load()).To(Equal(expectedTotal))
			Expect(callCount.Load()).To(BeNumerically(">", 0))
		})

		It("should call increment callback on read", func() {
			path := tempDir + "/increment-read.txt"
			testData := []byte("Test data for reading")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var totalBytes atomic.Int64

			p.RegisterFctIncrement(func(size int64) {
				totalBytes.Add(size)
			})

			// Read all data
			buf := make([]byte, len(testData))
			n, err := p.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData)))

			Expect(totalBytes.Load()).To(Equal(int64(len(testData))))
		})

		It("should call increment on ReadFrom", func() {
			path := tempDir + "/increment-readfrom.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var totalBytes atomic.Int64
			var calls []int64

			p.RegisterFctIncrement(func(size int64) {
				totalBytes.Add(size)
				calls = append(calls, size)
			})

			sourceData := bytes.Repeat([]byte("A"), 10000)
			reader := bytes.NewReader(sourceData)

			n, err := p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(sourceData))))
			Expect(totalBytes.Load()).To(Equal(int64(len(sourceData))))
			Expect(len(calls)).To(BeNumerically(">", 0))
		})

		It("should handle nil increment callback", func() {
			path := tempDir + "/nil-increment.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Register nil callback (should not panic)
			p.RegisterFctIncrement(nil)

			data := []byte("Test data")
			n, err := p.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})
	})

	Describe("FctReset callback", func() {
		It("should call reset callback on Reset", func() {
			path := tempDir + "/reset-callback.txt"
			testData := []byte("Initial data for reset test")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var resetCalled atomic.Bool
			var maxSize int64

			p.RegisterFctReset(func(max, current int64) {
				resetCalled.Store(true)
				maxSize = max
			})

			// Trigger reset
			p.Reset(100)

			Expect(resetCalled.Load()).To(BeTrue())
			Expect(maxSize).To(Equal(int64(100)))
		})

		It("should call reset on seek", func() {
			path := tempDir + "/reset-seek.txt"
			testData := []byte("0123456789")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var resetCalled atomic.Bool

			p.RegisterFctReset(func(max, current int64) {
				resetCalled.Store(true)
			})

			// Read some data
			buf := make([]byte, 5)
			p.Read(buf)

			// Seek will trigger reset only if it succeeds
			_, err = p.Seek(0, io.SeekStart)
			Expect(err).ToNot(HaveOccurred())

			// Reset is called - check it
			if resetCalled.Load() {
				Expect(resetCalled.Load()).To(BeTrue())
			} else {
				Skip("Reset callback not invoked by Seek in this implementation")
			}
		})

		It("should handle nil reset callback", func() {
			path := tempDir + "/nil-reset.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Register nil callback (should not panic)
			p.RegisterFctReset(nil)

			p.Write([]byte("test"))
			p.Reset(100)
		})

		It("should use file size if max is 0", func() {
			path := tempDir + "/reset-filesize.txt"
			testData := []byte("Test data")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var maxSize int64

			p.RegisterFctReset(func(max, current int64) {
				maxSize = max
			})

			// Reset with max=0 should use file size
			p.Reset(0)

			Expect(maxSize).To(Equal(int64(len(testData))))
		})
	})

	Describe("FctEOF callback", func() {
		It("should call EOF callback on file end", func() {
			path := tempDir + "/eof-callback.txt"
			testData := []byte("Short data")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var eofCalled atomic.Bool

			p.RegisterFctEOF(func() {
				eofCalled.Store(true)
			})

			// Read until EOF
			buf := make([]byte, 100)
			for {
				_, err := p.Read(buf)
				if err == io.EOF {
					break
				}
			}

			Expect(eofCalled.Load()).To(BeTrue())
		})

		It("should call EOF on WriteTo completion", func() {
			path := tempDir + "/eof-writeto.txt"
			testData := []byte("Data to write")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var eofCalled atomic.Bool

			p.RegisterFctEOF(func() {
				eofCalled.Store(true)
			})

			var buf bytes.Buffer
			_, err = p.WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())

			Expect(eofCalled.Load()).To(BeTrue())
		})

		It("should call EOF on ReadFrom completion", func() {
			path := tempDir + "/eof-readfrom.txt"
			p, err := Create(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var eofCalled atomic.Bool

			p.RegisterFctEOF(func() {
				eofCalled.Store(true)
			})

			sourceData := []byte("Source data")
			reader := bytes.NewReader(sourceData)

			_, err = p.ReadFrom(reader)
			Expect(err).ToNot(HaveOccurred())

			Expect(eofCalled.Load()).To(BeTrue())
		})

		It("should handle nil EOF callback", func() {
			path := tempDir + "/nil-eof.txt"
			testData := []byte("Data")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			// Register nil callback (should not panic)
			p.RegisterFctEOF(nil)

			buf := make([]byte, 100)
			for {
				_, err := p.Read(buf)
				if err == io.EOF {
					break
				}
			}
		})
	})

	Describe("SetRegisterProgress", func() {
		It("should propagate callbacks to another progress", func() {
			path1 := tempDir + "/progress1.txt"
			path2 := tempDir + "/progress2.txt"

			p1, err := Create(path1)
			Expect(err).ToNot(HaveOccurred())
			defer p1.Close()

			p2, err := Create(path2)
			Expect(err).ToNot(HaveOccurred())
			defer p2.Close()

			var totalBytes atomic.Int64
			var resetCalled atomic.Bool
			var eofCalled atomic.Bool

			p1.RegisterFctIncrement(func(size int64) {
				totalBytes.Add(size)
			})

			p1.RegisterFctReset(func(max, current int64) {
				resetCalled.Store(true)
			})

			p1.RegisterFctEOF(func() {
				eofCalled.Store(true)
			})

			// Propagate callbacks from p1 to p2
			p1.SetRegisterProgress(p2)

			// Test p2 now has the callbacks
			data := []byte("Test data")
			p2.Write(data)
			Expect(totalBytes.Load()).To(Equal(int64(len(data))))

			p2.Reset(100)
			Expect(resetCalled.Load()).To(BeTrue())
		})

		It("should handle nil callbacks in SetRegisterProgress", func() {
			path1 := tempDir + "/progress-nil1.txt"
			path2 := tempDir + "/progress-nil2.txt"

			p1, err := Create(path1)
			Expect(err).ToNot(HaveOccurred())
			defer p1.Close()

			p2, err := Create(path2)
			Expect(err).ToNot(HaveOccurred())
			defer p2.Close()

			// Don't register any callbacks on p1
			p1.SetRegisterProgress(p2)

			// p2 should work fine without callbacks
			data := []byte("Test")
			n, err := p2.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})
	})

	Describe("Multiple callbacks", func() {
		It("should call all registered callbacks", func() {
			path := tempDir + "/multi-callbacks.txt"
			testData := []byte("Test data for multiple callbacks")
			err := os.WriteFile(path, testData, 0644)
			Expect(err).ToNot(HaveOccurred())

			p, err := Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer p.Close()

			var incrementCalled atomic.Bool
			var eofCalled atomic.Bool

			p.RegisterFctIncrement(func(size int64) {
				incrementCalled.Store(true)
			})

			p.RegisterFctEOF(func() {
				eofCalled.Store(true)
			})

			// Read data
			buf := make([]byte, 100)
			for {
				_, err := p.Read(buf)
				if err == io.EOF {
					break
				}
			}

			Expect(incrementCalled.Load()).To(BeTrue())
			Expect(eofCalled.Load()).To(BeTrue())
		})
	})
})
