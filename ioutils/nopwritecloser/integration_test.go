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

package nopwritecloser_test

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	. "github.com/nabbar/golib/ioutils/nopwritecloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NopWriteCloser - Integration", func() {
	Context("JSON encoding", func() {
		It("should work with json.Encoder", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)
			encoder := json.NewEncoder(wc)

			data := map[string]interface{}{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
			}

			err := encoder.Encode(data)
			Expect(err).ToNot(HaveOccurred())

			// Decode to verify
			var decoded map[string]interface{}
			err = json.Unmarshal(buf.Bytes(), &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded["name"]).To(Equal("John Doe"))
			Expect(decoded["age"]).To(BeNumerically("==", 30))
		})

		It("should handle multiple JSON objects", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)
			encoder := json.NewEncoder(wc)

			objects := []map[string]string{
				{"type": "user", "name": "Alice"},
				{"type": "user", "name": "Bob"},
				{"type": "admin", "name": "Charlie"},
			}

			for _, obj := range objects {
				err := encoder.Encode(obj)
				Expect(err).ToNot(HaveOccurred())
			}

			wc.Close()

			// Verify data was written
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})
	})

	Context("Compression", func() {
		It("should work with gzip writer", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)
			gzipWriter := gzip.NewWriter(wc)

			originalData := []byte("This is test data that will be compressed using gzip")
			_, err := gzipWriter.Write(originalData)
			Expect(err).ToNot(HaveOccurred())

			err = gzipWriter.Close()
			Expect(err).ToNot(HaveOccurred())

			err = wc.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify data was compressed
			Expect(buf.Len()).To(BeNumerically(">", 0))

			// Decompress and verify
			gzipReader, err := gzip.NewReader(buf)
			Expect(err).ToNot(HaveOccurred())
			defer gzipReader.Close()

			decompressed, err := io.ReadAll(gzipReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(decompressed).To(Equal(originalData))
		})
	})

	Context("Hashing", func() {
		It("should work with hash writers", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Use MultiWriter to write to both buffer and hash
			hasher := md5.New()
			multiWriter := io.MultiWriter(wc, hasher)

			data := []byte("Hash this data")
			_, err := multiWriter.Write(data)
			Expect(err).ToNot(HaveOccurred())

			// Get hash
			hash := hex.EncodeToString(hasher.Sum(nil))
			Expect(hash).ToNot(BeEmpty())

			// Verify data was written
			Expect(buf.Bytes()).To(Equal(data))

			err = wc.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Formatted output", func() {
		It("should work with fmt.Fprintf", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			n, err := fmt.Fprintf(wc, "Hello, %s! You have %d new messages.", "Alice", 5)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(BeNumerically(">", 0))

			Expect(buf.String()).To(Equal("Hello, Alice! You have 5 new messages."))
		})

		It("should handle multiple format operations", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			fmt.Fprintf(wc, "Line 1\n")
			fmt.Fprintf(wc, "Line 2\n")
			fmt.Fprintf(wc, "Line 3\n")

			Expect(buf.String()).To(Equal("Line 1\nLine 2\nLine 3\n"))
		})
	})

	Context("Chained writers", func() {
		It("should work in a chain of writers", func() {
			finalBuf := &bytes.Buffer{}
			wc := New(finalBuf)

			// Create a chain: buffer -> nopwritecloser -> buffer
			intermediateBuf := &bytes.Buffer{}
			chainedWc := New(io.MultiWriter(wc, intermediateBuf))

			data := []byte("chained data")
			_, err := chainedWc.Write(data)
			Expect(err).ToNot(HaveOccurred())

			// Both buffers should have the data
			Expect(finalBuf.Bytes()).To(Equal(data))
			Expect(intermediateBuf.Bytes()).To(Equal(data))

			// Closing should not affect the buffers
			err = chainedWc.Close()
			Expect(err).ToNot(HaveOccurred())

			// Can still write after close
			_, err = chainedWc.Write([]byte(" more"))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Real-world scenarios", func() {
		It("should work as a log sink", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Simulate logging
			logger := func(level, message string) {
				fmt.Fprintf(wc, "[%s] %s\n", level, message)
			}

			logger("INFO", "Application started")
			logger("DEBUG", "Loading configuration")
			logger("INFO", "Server listening on :8080")
			logger("ERROR", "Database connection failed")

			wc.Close()

			logs := buf.String()
			Expect(logs).To(ContainSubstring("[INFO] Application started"))
			Expect(logs).To(ContainSubstring("[DEBUG] Loading configuration"))
			Expect(logs).To(ContainSubstring("[ERROR] Database connection failed"))
		})

		It("should work as a response writer wrapper", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Simulate HTTP response writing
			writeResponse := func(w io.Writer) {
				fmt.Fprintf(w, "HTTP/1.1 200 OK\r\n")
				fmt.Fprintf(w, "Content-Type: application/json\r\n")
				fmt.Fprintf(w, "\r\n")
				fmt.Fprintf(w, `{"status":"success","message":"Request processed"}`)
			}

			writeResponse(wc)
			wc.Close()

			response := buf.String()
			Expect(response).To(ContainSubstring("HTTP/1.1 200 OK"))
			Expect(response).To(ContainSubstring("Content-Type: application/json"))
			Expect(response).To(ContainSubstring(`"status":"success"`))
		})

		It("should work as a tee for data inspection", func() {
			mainBuf := &bytes.Buffer{}
			inspectionBuf := &bytes.Buffer{}

			// Create a tee that writes to both buffers
			tee := io.MultiWriter(mainBuf, inspectionBuf)
			wc := New(tee)

			// Write some data
			data := []byte("Important data that needs inspection")
			_, err := wc.Write(data)
			Expect(err).ToNot(HaveOccurred())

			// Both buffers should have identical data
			Expect(mainBuf.Bytes()).To(Equal(data))
			Expect(inspectionBuf.Bytes()).To(Equal(data))

			wc.Close()
		})

		It("should work with standard library io.Copy", func() {
			source := bytes.NewBufferString("Data to copy")
			dest := &bytes.Buffer{}
			wc := New(dest)

			n, err := io.Copy(wc, source)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(12)))
			Expect(dest.String()).To(Equal("Data to copy"))

			wc.Close()
		})

		It("should work with io.Pipe", func() {
			pr, pw := io.Pipe()
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Writer goroutine
			go func() {
				defer pw.Close()
				pw.Write([]byte("pipe data"))
			}()

			// Copy from pipe to nopwritecloser
			_, err := io.Copy(wc, pr)
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("pipe data"))
			wc.Close()
		})
	})

	Context("Performance scenarios", func() {
		It("should handle streaming writes efficiently", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Simulate streaming data
			chunkSize := 4096
			numChunks := 1000
			chunk := make([]byte, chunkSize)
			for i := range chunk {
				chunk[i] = byte(i % 256)
			}

			for i := 0; i < numChunks; i++ {
				n, err := wc.Write(chunk)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(chunkSize))
			}

			expectedSize := chunkSize * numChunks
			Expect(buf.Len()).To(Equal(expectedSize))

			wc.Close()
		})

		It("should handle buffered writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Write in various sizes
			sizes := []int{1, 10, 100, 1000, 10000}
			totalBytes := 0

			for _, size := range sizes {
				data := make([]byte, size)
				n, err := wc.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(size))
				totalBytes += size
			}

			Expect(buf.Len()).To(Equal(totalBytes))
			wc.Close()
		})
	})
})
