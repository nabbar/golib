/*
MIT License

Copyright (c) 2023 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package hexa_test

import (
	"bytes"
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libenc "github.com/nabbar/golib/encoding"
	enchex "github.com/nabbar/golib/encoding/hexa"
)

// errorReader always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

func (e *errorReader) Close() error {
	return fmt.Errorf("simulated close error")
}

// errorWriter always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write error")
}

func (e *errorWriter) Close() error {
	return fmt.Errorf("simulated close error")
}

var _ = Describe("Hexadecimal Edge Cases and Error Handling", func() {
	Describe("Error Handling", func() {
		It("should export ErrInvalidBufferSize error", func() {
			Expect(enchex.ErrInvalidBufferSize).ToNot(BeNil())
			Expect(enchex.ErrInvalidBufferSize.Error()).To(ContainSubstring("buffer"))
		})
	})

	Describe("Boundary Conditions", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should handle single byte", func() {
			data := []byte{0x42}
			encoded := coder.Encode(data)
			decoded, err := coder.Decode(encoded)

			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(data))
		})

		It("should handle all zero bytes", func() {
			data := make([]byte, 100)
			encoded := coder.Encode(data)
			decoded, err := coder.Decode(encoded)

			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(data))
		})

		It("should handle all 0xFF bytes", func() {
			data := make([]byte, 100)
			for i := range data {
				data[i] = 0xFF
			}
			encoded := coder.Encode(data)
			decoded, err := coder.Decode(encoded)

			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(data))
		})

		It("should handle alternating pattern", func() {
			data := make([]byte, 1000)
			for i := range data {
				data[i] = byte(i % 2 * 255)
			}
			encoded := coder.Encode(data)
			decoded, err := coder.Decode(encoded)

			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(data))
		})

		It("should handle very large data", func() {
			// 10MB of data
			largeData := make([]byte, 10*1024*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			encoded := coder.Encode(largeData)
			decoded, err := coder.Decode(encoded)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(decoded)).To(Equal(len(largeData)))
			Expect(decoded).To(Equal(largeData))
		})

		It("should handle sequential bytes", func() {
			data := make([]byte, 256)
			for i := range data {
				data[i] = byte(i)
			}
			encoded := coder.Encode(data)
			decoded, err := coder.Decode(encoded)

			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(data))
		})
	})

	Describe("Reader Edge Cases", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should handle reader with immediate EOF", func() {
			reader := bytes.NewReader([]byte{})
			encReader := coder.EncodeReader(reader)

			buffer := make([]byte, 100)
			_, err := encReader.Read(buffer)
			Expect(err).To(Equal(io.EOF))
		})

		It("should handle reader errors in EncodeReader", func() {
			errReader := &errorReader{}
			encReader := coder.EncodeReader(errReader)

			buffer := make([]byte, 100)
			_, err := encReader.Read(buffer)
			Expect(err).To(HaveOccurred())
		})

		It("should handle reader errors in DecodeReader", func() {
			errReader := &errorReader{}
			decReader := coder.DecodeReader(errReader)

			buffer := make([]byte, 100)
			_, err := decReader.Read(buffer)
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in EncodeReader", func() {
			errReader := &errorReader{}
			encReader := coder.EncodeReader(errReader)

			err := encReader.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in DecodeReader", func() {
			errReader := &errorReader{}
			decReader := coder.DecodeReader(errReader)

			err := decReader.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle non-closeable readers", func() {
			type nonCloseableReader struct {
				*bytes.Reader
			}

			data := []byte("test")
			ncr := &nonCloseableReader{Reader: bytes.NewReader(data)}

			encReader := coder.EncodeReader(ncr)
			err := encReader.Close()
			Expect(err).ToNot(HaveOccurred()) // Should not error for non-closeable
		})
	})

	Describe("Writer Edge Cases", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should handle writer errors in EncodeWriter", func() {
			errWriter := &errorWriter{}
			encWriter := coder.EncodeWriter(errWriter)

			_, err := encWriter.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle writer errors in DecodeWriter", func() {
			hexEncoded := coder.Encode([]byte("test"))

			errWriter := &errorWriter{}
			decWriter := coder.DecodeWriter(errWriter)

			_, err := decWriter.Write(hexEncoded)
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in EncodeWriter", func() {
			errWriter := &errorWriter{}
			encWriter := coder.EncodeWriter(errWriter)

			err := encWriter.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in DecodeWriter", func() {
			errWriter := &errorWriter{}
			decWriter := coder.DecodeWriter(errWriter)

			err := decWriter.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle non-closeable writers", func() {
			type nonCloseableWriter struct {
				*bytes.Buffer
			}

			ncw := &nonCloseableWriter{Buffer: &bytes.Buffer{}}

			encWriter := coder.EncodeWriter(ncw)
			err := encWriter.Close()
			Expect(err).ToNot(HaveOccurred()) // Should not error for non-closeable
		})
	})

	Describe("Invalid Hex Sequences", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should detect invalid hex characters", func() {
			invalidSequences := []string{
				"4g",      // Invalid character 'g'
				"zz",      // Invalid characters 'z'
				"00@0",    // Special character
				"GH",      // Beyond 'F'
				"1234xyz", // Mix of valid and invalid
			}

			for _, seq := range invalidSequences {
				_, err := coder.Decode([]byte(seq))
				Expect(err).To(HaveOccurred(), "Should error for: "+seq)
			}
		})

		It("should detect odd-length hex strings", func() {
			oddLengths := []string{
				"1",
				"123",
				"12345",
				"1234567",
			}

			for _, seq := range oddLengths {
				_, err := coder.Decode([]byte(seq))
				Expect(err).To(HaveOccurred(), "Should error for odd length: "+seq)
			}
		})

		It("should handle whitespace in hex (should fail)", func() {
			hexWithSpace := []byte("48 65 6c 6c 6f")
			_, err := coder.Decode(hexWithSpace)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Concurrency Safety", func() {
		It("should handle concurrent encoding", func() {
			coder := enchex.New()

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func(id int) {
					defer GinkgoRecover()
					data := []byte(fmt.Sprintf("message %d", id))
					encoded := coder.Encode(data)
					Expect(encoded).ToNot(BeNil())
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent decoding", func() {
			coder := enchex.New()

			// Pre-encode messages
			var encoded [][]byte
			for i := 0; i < 10; i++ {
				data := []byte(fmt.Sprintf("message %d", i))
				encoded = append(encoded, coder.Encode(data))
			}

			done := make(chan bool, 10)
			for i, enc := range encoded {
				go func(id int, data []byte) {
					defer GinkgoRecover()
					decoded, err := coder.Decode(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(decoded).ToNot(BeNil())
					done <- true
				}(i, enc)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})

	Describe("Reset Behavior", func() {
		It("should handle multiple resets", func() {
			coder := enchex.New()

			coder.Reset()
			coder.Reset()
			coder.Reset()

			// After multiple resets, should still work
			result := coder.Encode([]byte("test"))
			Expect(len(result)).To(BeNumerically(">", 0))
		})

		It("should continue working after reset", func() {
			coder := enchex.New()

			plaintext := []byte("test before reset")
			encoded := coder.Encode(plaintext)
			Expect(len(encoded)).To(BeNumerically(">", 0))

			coder.Reset()

			// After reset, should still work
			result := coder.Encode([]byte("test after reset"))
			Expect(len(result)).To(BeNumerically(">", 0))

			// Original encoded data should still decode
			decoded, err := coder.Decode(encoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(plaintext))
		})
	})

	Describe("Hex Case Sensitivity", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should decode lowercase hex", func() {
			hexData := []byte("48656c6c6f")
			decoded, err := coder.Decode(hexData)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(decoded)).To(Equal("Hello"))
		})

		It("should decode uppercase hex", func() {
			hexData := []byte("48656C6C6F")
			decoded, err := coder.Decode(hexData)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(decoded)).To(Equal("Hello"))
		})

		It("should decode mixed case hex", func() {
			hexData := []byte("48656C6c6F")
			decoded, err := coder.Decode(hexData)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(decoded)).To(Equal("Hello"))
		})

		It("should produce lowercase hex when encoding", func() {
			plaintext := []byte("Hello")
			encoded := coder.Encode(plaintext)

			// Standard library produces lowercase
			Expect(string(encoded)).To(Equal("48656c6c6f"))
		})
	})
})
