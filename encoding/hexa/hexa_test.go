/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package hexa_test

import (
	"bytes"
	"io"

	libenc "github.com/nabbar/golib/encoding"
	enchex "github.com/nabbar/golib/encoding/hexa"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("encoding/hexa", func() {
	Context("Simple encoding/decoding", func() {
		var (
			err error
			msg []byte
			res []byte
			sig []byte
			crp libenc.Coder
		)

		It("Create new instance must succeed", func() {
			crp = enchex.New()
			Expect(crp).ToNot(BeNil())
		})

		It("Encode must succeed", func() {
			msg = []byte("Hello World")
			sig = crp.Encode(msg)
			Expect(sig).ToNot(BeNil())
		})

		It("Decode must succeed", func() {
			res, err = crp.Decode(sig)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).ToNot(BeNil())
			Expect(res).To(BeEquivalentTo(msg)) // bytes.Equal(msg, []byte("Hello World"))(BeNil())
		})
	})

	Context("IO interface with encoding/decoding", func() {
		var (
			err error
			nbr int
			msg = []byte("Hello World")
			res = make([]byte, len(msg)*2)
			crp libenc.Coder
			buf = bytes.NewBuffer(make([]byte, 0, 32*1024))
			rdr io.Reader
			wrt io.Writer
		)

		It("Create new instance must succeed", func() {
			crp = enchex.New()
			Expect(crp).ToNot(BeNil())
		})

		It("Create and write an io.writer to encode must succeed", func() {
			wrt = crp.EncodeWriter(buf)
			Expect(wrt).ToNot(BeNil())

			nbr, err = wrt.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeEquivalentTo(len(msg)))
		})

		It("Create and reading an io.reader to decode must succeed", func() {
			rdr = crp.DecodeReader(buf)
			Expect(rdr).ToNot(BeNil())

			nbr, err = rdr.Read(res)
			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeEquivalentTo(11))
			Expect(res[:nbr]).To(BeEquivalentTo(msg[:nbr]))
		})

		It("Create an io.reader and read from it to encode string but use too small buffer must occur an error", func() {
			res = make([]byte, 1)

			rdr = crp.EncodeReader(buf)
			Expect(rdr).ToNot(BeNil())

			buf.Reset()
			buf.Write(msg)

			nbr, err = rdr.Read(res)
			Expect(err).To(HaveOccurred())
		})

		It("Create an io.reader and read from it to encode string must succeed", func() {
			res = make([]byte, cap(msg)*3)

			rdr = crp.EncodeReader(buf)
			Expect(rdr).ToNot(BeNil())

			buf.Reset()
			buf.Write(msg)

			nbr, err = rdr.Read(res)
			Expect(err).ToNot(HaveOccurred())
			res = res[:nbr]
		})

		It("Create an io.writer and write on it to decode must succeed", func() {
			wrt = crp.DecodeWriter(buf)
			Expect(wrt).ToNot(BeNil())

			buf.Reset()
			nbr, err = wrt.Write(res)

			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeNumerically(">", len(msg)))
			Expect(buf.Len()).To(BeEquivalentTo(len(msg)))
			Expect(buf.Bytes()).To(BeEquivalentTo(msg[:buf.Len()]))
		})
	})
})
