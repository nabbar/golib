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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	encmux "github.com/nabbar/golib/encoding/mux"
)

// mockWriter for testing error conditions
type mockWriter struct {
	buffer bytes.Buffer
	err    error
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.buffer.Write(p)
}

var _ = Describe("Multiplexer Operations", func() {
	Describe("NewMultiplexer", func() {
		It("should create a new multiplexer instance", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\n')
			Expect(mux).ToNot(BeNil())
		})

		It("should create multiplexer with custom delimiter", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '|')
			Expect(mux).ToNot(BeNil())
		})

		It("should create multiplexer with null byte delimiter", func() {
			buf := &bytes.Buffer{}
			mux := encmux.NewMultiplexer(buf, '\x00')
			Expect(mux).ToNot(BeNil())
		})
	})

	Describe("NewChannel", func() {
		var (
			buf *bytes.Buffer
			mux encmux.Multiplexer
		)

		BeforeEach(func() {
			buf = &bytes.Buffer{}
			mux = encmux.NewMultiplexer(buf, '\n')
		})

		It("should create a new channel with single byte key", func() {
			channel := mux.NewChannel('a')
			Expect(channel).ToNot(BeNil())
		})

		It("should create a new channel with numeric key", func() {
			channel := mux.NewChannel('1')
			Expect(channel).ToNot(BeNil())
		})

		It("should create a new channel with unicode key", func() {
			channel := mux.NewChannel('世')
			Expect(channel).ToNot(BeNil())
		})

		It("should create multiple independent channels", func() {
			ch1 := mux.NewChannel('a')
			ch2 := mux.NewChannel('b')
			ch3 := mux.NewChannel('c')

			Expect(ch1).ToNot(BeNil())
			Expect(ch2).ToNot(BeNil())
			Expect(ch3).ToNot(BeNil())
		})

		It("should allow reusing same key for multiple channels", func() {
			ch1 := mux.NewChannel('a')
			ch2 := mux.NewChannel('a')

			Expect(ch1).ToNot(BeNil())
			Expect(ch2).ToNot(BeNil())
		})
	})

	Describe("Channel Write Operations", func() {
		var (
			buf *bytes.Buffer
			mux encmux.Multiplexer
		)

		BeforeEach(func() {
			buf = &bytes.Buffer{}
			mux = encmux.NewMultiplexer(buf, '\n')
		})

		It("should write simple message to channel", func() {
			channel := mux.NewChannel('a')
			msg := []byte("Hello, World!")

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("should write empty message to channel", func() {
			channel := mux.NewChannel('a')
			msg := []byte{}

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should write nil message to channel", func() {
			channel := mux.NewChannel('a')

			n, err := channel.Write(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should write binary data to channel", func() {
			channel := mux.NewChannel('a')
			binary := []byte{0x00, 0xFF, 0x7F, 0x80, 0xDE, 0xAD, 0xBE, 0xEF}

			n, err := channel.Write(binary)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(binary)))
		})

		It("should write large message to channel", func() {
			channel := mux.NewChannel('a')
			largeMsg := make([]byte, 10*1024) // 10KB
			for i := range largeMsg {
				largeMsg[i] = byte(i % 256)
			}

			n, err := channel.Write(largeMsg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeMsg)))
		})

		It("should write UTF-8 text to channel", func() {
			channel := mux.NewChannel('a')
			msg := []byte("Hello 世界 🔒")

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))
		})

		It("should include delimiter in multiplexed output", func() {
			channel := mux.NewChannel('a')
			msg := []byte("test")

			channel.Write(msg)

			output := buf.Bytes()
			Expect(output[len(output)-1]).To(Equal(byte('\n')))
		})

		It("should encode message with CBOR and hex", func() {
			channel := mux.NewChannel('a')
			msg := []byte("test")

			channel.Write(msg)

			// Output should be CBOR-encoded structure with hex-encoded data
			output := buf.Bytes()
			Expect(len(output)).To(BeNumerically(">", len(msg)))
		})
	})

	Describe("Multiple Channel Operations", func() {
		var (
			buf *bytes.Buffer
			mux encmux.Multiplexer
		)

		BeforeEach(func() {
			buf = &bytes.Buffer{}
			mux = encmux.NewMultiplexer(buf, '\n')
		})

		It("should write to multiple channels", func() {
			ch1 := mux.NewChannel('a')
			ch2 := mux.NewChannel('b')

			n1, err1 := ch1.Write([]byte("Message A"))
			n2, err2 := ch2.Write([]byte("Message B"))

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(n1).To(Equal(9))
			Expect(n2).To(Equal(9))
		})

		It("should interleave messages from different channels", func() {
			ch1 := mux.NewChannel('a')
			ch2 := mux.NewChannel('b')
			ch3 := mux.NewChannel('c')

			ch1.Write([]byte("A1"))
			ch2.Write([]byte("B1"))
			ch3.Write([]byte("C1"))
			ch1.Write([]byte("A2"))
			ch2.Write([]byte("B2"))

			// All messages should be in buffer
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("should handle concurrent writes to same channel", func() {
			channel := mux.NewChannel('a')

			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func(id int) {
					defer GinkgoRecover()
					msg := []byte(fmt.Sprintf("Message %d", id))
					_, err := channel.Write(msg)
					Expect(err).ToNot(HaveOccurred())
					done <- true
				}(i)
			}

			for i := 0; i < 5; i++ {
				<-done
			}

			Expect(buf.Len()).To(BeNumerically(">", 0))
		})
	})

	Describe("Error Handling", func() {
		It("should handle write errors from underlying writer", func() {
			expectedErr := errors.New("write error")
			mockW := &mockWriter{err: expectedErr}
			mux := encmux.NewMultiplexer(mockW, '\n')

			channel := mux.NewChannel('a')
			_, err := channel.Write([]byte("test"))

			Expect(err).To(Equal(expectedErr))
		})

		It("should export ErrInvalidInstance error", func() {
			Expect(encmux.ErrInvalidInstance).ToNot(BeNil())
			Expect(encmux.ErrInvalidInstance.Error()).To(ContainSubstring("invalid"))
		})

		It("should export ErrInvalidChannel error", func() {
			Expect(encmux.ErrInvalidChannel).ToNot(BeNil())
			Expect(encmux.ErrInvalidChannel.Error()).To(ContainSubstring("channel"))
		})
	})

	Describe("Special Characters and Edge Cases", func() {
		var (
			buf *bytes.Buffer
			mux encmux.Multiplexer
		)

		BeforeEach(func() {
			buf = &bytes.Buffer{}
			mux = encmux.NewMultiplexer(buf, '\n')
		})

		It("should handle messages containing delimiter", func() {
			channel := mux.NewChannel('a')
			msg := []byte("Line 1\nLine 2\nLine 3")

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))
		})

		It("should handle messages with null bytes", func() {
			channel := mux.NewChannel('a')
			msg := []byte("Before\x00After")

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))
		})

		It("should handle all zero bytes", func() {
			channel := mux.NewChannel('a')
			msg := make([]byte, 100)

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))
		})

		It("should handle all 0xFF bytes", func() {
			channel := mux.NewChannel('a')
			msg := make([]byte, 100)
			for i := range msg {
				msg[i] = 0xFF
			}

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))
		})

		It("should handle sequential byte patterns", func() {
			channel := mux.NewChannel('a')
			msg := make([]byte, 256)
			for i := range msg {
				msg[i] = byte(i)
			}

			n, err := channel.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(msg)))
		})
	})
})
