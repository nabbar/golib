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

package sha256_test

import (
	"bytes"
	"crypto/sha256"
	"io"

	libenc "github.com/nabbar/golib/encoding"
	encsha "github.com/nabbar/golib/encoding/sha256"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("encoding/sha256", func() {
	Context("Simple encoding/decoding", func() {
		var (
			err error
			msg = []byte("Hello World")
			sig []byte
			crp libenc.Coder
			chk = sha256.Sum256(msg)
		)

		It("Create new instance must succeed", func() {
			crp = encsha.New()
			Expect(crp).ToNot(BeNil())
		})

		It("must succeed to encode", func() {
			sig = crp.Encode(msg)
			Expect(sig).ToNot(BeNil())
			Expect(sig).To(BeEquivalentTo(chk[:])) // bytes.Equal(msg, []byte("Hello World"))(BeNil())
			crp.Reset()
		})

		It("must return error when decode is call", func() {
			crp = encsha.New()
			Expect(crp).ToNot(BeNil())

			_, err = crp.Decode(msg)
			Expect(err).To(HaveOccurred())
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
			rdr io.ReadCloser
			wrt io.WriteCloser
			chk = (sha256.Sum256(msg))
		)

		It("Create new instance must succeed", func() {
			crp = encsha.New()
			Expect(crp).ToNot(BeNil())
		})

		It("Create and write an io.writer to encode must succeed", func() {
			wrt = crp.EncodeWriter(buf)
			Expect(wrt).ToNot(BeNil())

			nbr, err = wrt.Write(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeEquivalentTo(len(msg)))
			Expect(buf.Bytes()).To(BeEquivalentTo(msg))
			Expect(crp.Encode(nil)).To(BeEquivalentTo(chk))
			Expect(wrt.Close()).ToNot(HaveOccurred())
		})

		It("Reset and Create new instance must succeed", func() {
			crp.Reset()
			crp = encsha.New()
			Expect(crp).ToNot(BeNil())
		})

		It("Create an io.reader and read from it to encode string must succeed", func() {
			buf.Reset()
			buf.Write(msg)

			rdr = crp.EncodeReader(buf)
			Expect(rdr).ToNot(BeNil())

			res, err = io.ReadAll(rdr)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(res)).To(BeEquivalentTo(len(msg)))
			Expect(res).To(BeEquivalentTo(msg))
			Expect(crp.Encode(nil)).To(BeEquivalentTo(chk))
			Expect(rdr.Close()).ToNot(HaveOccurred())
			crp.Reset()
		})

		It("must return error when reader/writer decoder is call", func() {
			crp = encsha.New()
			Expect(crp).ToNot(BeNil())

			rdr = crp.DecodeReader(bytes.NewReader(msg))
			Expect(rdr).To(BeNil())

			buf.Reset()
			wrt = crp.DecodeWriter(buf)
			Expect(rdr).To(BeNil())
		})
	})
})
