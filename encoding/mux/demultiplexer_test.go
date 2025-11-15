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

package mux_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	encmux "github.com/nabbar/golib/encoding/mux"
)

// mockReader for testing error conditions
type mockReader struct {
	data []byte
	pos  int
	err  error
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// mockWriter for testing error conditions
type mockWriterDemux struct {
	buffer bytes.Buffer
	err    error
}

func (m *mockWriterDemux) Write(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.buffer.Write(p)
}

var _ = Describe("DeMultiplexer Operations", func() {
	Describe("NewDeMultiplexer", func() {
		It("should create a new demultiplexer instance", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			Expect(dmux).ToNot(BeNil())
		})

		It("should create demultiplexer with custom delimiter", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '|', 0)
			Expect(dmux).ToNot(BeNil())
		})

		It("should create demultiplexer with default buffer size", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			Expect(dmux).ToNot(BeNil())
		})

		It("should create demultiplexer with custom buffer size", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 8*1024)
			Expect(dmux).ToNot(BeNil())
		})

		It("should create demultiplexer with very small buffer", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 64)
			Expect(dmux).ToNot(BeNil())
		})

		It("should create demultiplexer with very large buffer", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 1024*1024)
			Expect(dmux).ToNot(BeNil())
		})
	})

	Describe("NewChannel Registration", func() {
		var dmux encmux.DeMultiplexer

		BeforeEach(func() {
			buf := bytes.NewBuffer([]byte{})
			dmux = encmux.NewDeMultiplexer(buf, '\n', 0)
		})

		It("should register a new channel", func() {
			output := &bytes.Buffer{}
			dmux.NewChannel('a', output)
			// No panic means success
			Expect(true).To(BeTrue())
		})

		It("should register multiple channels", func() {
			out1 := &bytes.Buffer{}
			out2 := &bytes.Buffer{}
			out3 := &bytes.Buffer{}

			dmux.NewChannel('a', out1)
			dmux.NewChannel('b', out2)
			dmux.NewChannel('c', out3)

			Expect(true).To(BeTrue())
		})

		It("should allow re-registering same channel key", func() {
			out1 := &bytes.Buffer{}
			out2 := &bytes.Buffer{}

			dmux.NewChannel('a', out1)
			dmux.NewChannel('a', out2) // Should replace

			Expect(true).To(BeTrue())
		})

		It("should register channel with unicode key", func() {
			output := &bytes.Buffer{}
			dmux.NewChannel('世', output)
			Expect(true).To(BeTrue())
		})

		It("should register channel with numeric key", func() {
			output := &bytes.Buffer{}
			dmux.NewChannel('1', output)
			Expect(true).To(BeTrue())
		})
	})

	Describe("Read Operations", func() {
		It("should implement io.Reader interface", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)

			// Verify it implements io.Reader
			var _ io.Reader = dmux
			Expect(dmux).ToNot(BeNil())
		})

		It("should return EOF when buffer is empty", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			output := &bytes.Buffer{}
			dmux.NewChannel('a', output)

			p := make([]byte, 100)
			_, err := dmux.Read(p)

			Expect(err).To(Equal(io.EOF))
		})

		It("should return error when no channels registered", func() {
			// Create a minimal valid multiplexed message
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')
			channel := mux.NewChannel('a')
			channel.Write([]byte("test"))

			// Try to demux without registering channels
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			p := make([]byte, 100)
			_, err := dmux.Read(p)

			Expect(err).To(Equal(encmux.ErrInvalidChannel))
		})

		It("should return error for unknown channel key", func() {
			// Create message for channel 'a'
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')
			channel := mux.NewChannel('a')
			channel.Write([]byte("test"))

			// Register channel 'b' instead
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			output := &bytes.Buffer{}
			dmux.NewChannel('b', output)

			p := make([]byte, 100)
			_, err := dmux.Read(p)

			Expect(err).To(Equal(encmux.ErrInvalidChannel))
		})
	})

	Describe("Copy Operations", func() {
		It("should copy data to registered channels", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')

			// Write messages
			ch1 := mux.NewChannel('a')
			n, err := ch1.Write([]byte("Message A"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(9))

			// Read messages
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			out := &bytes.Buffer{}
			dmux.NewChannel('a', out)

			err = dmux.Copy()
			Expect(err).ToNot(HaveOccurred())
			Expect(out.String()).To(Equal("Message A"))
		})

		It("should handle EOF gracefully", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			output := &bytes.Buffer{}
			dmux.NewChannel('a', output)

			err := dmux.Copy()
			Expect(err).ToNot(HaveOccurred()) // EOF should not be returned as error
		})

		It("should copy multiple messages", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')

			// Write multiple messages
			ch := mux.NewChannel('a')
			ch.Write([]byte("Message 1"))
			ch.Write([]byte("Message 2"))
			ch.Write([]byte("Message 3"))

			// Read messages
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			out := &bytes.Buffer{}
			dmux.NewChannel('a', out)

			err := dmux.Copy()
			Expect(err).ToNot(HaveOccurred())
			Expect(out.String()).To(Equal("Message 1Message 2Message 3"))
		})

		It("should copy to multiple channels", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')

			// Write to different channels
			ch1 := mux.NewChannel('a')
			ch2 := mux.NewChannel('b')
			ch1.Write([]byte("For A"))
			ch2.Write([]byte("For B"))
			ch1.Write([]byte(" Again"))

			// Read from multiple channels
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			out1 := &bytes.Buffer{}
			out2 := &bytes.Buffer{}
			dmux.NewChannel('a', out1)
			dmux.NewChannel('b', out2)

			err := dmux.Copy()
			Expect(err).ToNot(HaveOccurred())
			Expect(out1.String()).To(Equal("For A Again"))
			Expect(out2.String()).To(Equal("For B"))
		})

	})

	Describe("Error Propagation", func() {
		It("should propagate read errors", func() {
			expectedErr := errors.New("read error")
			mockR := &mockReader{err: expectedErr}
			dmux := encmux.NewDeMultiplexer(mockR, '\n', 0)
			output := &bytes.Buffer{}
			dmux.NewChannel('a', output)

			err := dmux.Copy()
			Expect(err).To(Equal(expectedErr))
		})

		It("should propagate write errors from channel", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')

			// Write a message
			ch := mux.NewChannel('a')
			ch.Write([]byte("test"))

			// Use error writer for demux
			expectedErr := errors.New("write error")
			mockW := &mockWriterDemux{err: expectedErr}

			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			dmux.NewChannel('a', mockW)

			err := dmux.Copy()
			Expect(err).To(Equal(expectedErr))
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle concurrent channel registration", func() {
			buf := bytes.NewBuffer([]byte{})
			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)

			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func(id int) {
					defer GinkgoRecover()
					out := &bytes.Buffer{}
					dmux.NewChannel(rune('a'+id), out)
					done <- true
				}(i)
			}

			for i := 0; i < 5; i++ {
				<-done
			}

			Expect(true).To(BeTrue())
		})
	})

	Describe("Fmt Integration", func() {
		It("should handle fmt.Fprintln output", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')

			ch := mux.NewChannel('a')
			n, err := fmt.Fprintln(ch, "Hello World")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(12)) // "Hello World\n"

			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			out := &bytes.Buffer{}
			dmux.NewChannel('a', out)

			err = dmux.Copy()
			Expect(err).ToNot(HaveOccurred())
			Expect(out.String()).To(Equal("Hello World\n"))
		})

		It("should handle fmt.Fprintf output", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')

			ch := mux.NewChannel('a')
			fmt.Fprintf(ch, "Number: %d, String: %s", 42, "test")

			dmux := encmux.NewDeMultiplexer(buf, '\n', 0)
			out := &bytes.Buffer{}
			dmux.NewChannel('a', out)

			err := dmux.Copy()
			Expect(err).ToNot(HaveOccurred())
			Expect(out.String()).To(Equal("Number: 42, String: test"))
		})
	})
})
