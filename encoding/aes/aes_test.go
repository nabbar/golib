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

package aes_test

import (
	"bytes"
	"encoding/hex"
	"io"

	libenc "github.com/nabbar/golib/encoding"
	encaes "github.com/nabbar/golib/encoding/aes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("encoding/aes", func() {
	Context("Validating Key/Nonce creation", func() {
		var (
			err error
			crp libenc.Coder
			key [32]byte
			non [12]byte
			hks string
			hns string
			hkb [32]byte
			hnb [12]byte
		)

		It("Create key must succeed", func() {
			key, err = encaes.GenKey()
			Expect(key).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Create nonce must succeed", func() {
			non, err = encaes.GenNonce()
			Expect(non).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Create new instance must succeed", func() {
			crp, err = encaes.New(key, non)
			Expect(crp).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Get key from Hex must succeed", func() {
			hks = hex.EncodeToString(key[:])
			Expect(hks).ToNot(BeNil())
			Expect(len(hks)).To(BeEquivalentTo(64))

			hkb, err = encaes.GetHexKey(hks)
			Expect(err).ToNot(HaveOccurred())
			Expect(hkb).ToNot(BeNil())
			Expect(hkb).To(BeEquivalentTo(key))
		})

		It("Get nonce from Hex must succeed", func() {
			hns = hex.EncodeToString(non[:])
			Expect(hns).ToNot(BeNil())
			Expect(len(hns)).To(BeEquivalentTo(24))

			hnb, err = encaes.GetHexNonce(hns)
			Expect(err).ToNot(HaveOccurred())
			Expect(hnb).ToNot(BeNil())
			Expect(hnb).To(BeEquivalentTo(non))
		})

		It("Create new instance must succeed", func() {
			crp, err = encaes.New(hkb, hnb)
			Expect(crp).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Simple encoding/decoding", func() {
		var (
			err error
			msg []byte
			sig []byte
			crp libenc.Coder
			key [32]byte
			non [12]byte
		)

		It("Create key must succeed", func() {
			key, err = encaes.GenKey()
			Expect(key).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Create nonce must succeed", func() {
			non, err = encaes.GenNonce()
			Expect(non).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Create new instance must succeed", func() {
			crp, err = encaes.New(key, non)
			Expect(crp).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Encode must succeed", func() {
			msg = []byte("Hello World")
			sig = crp.Encode(msg)
			Expect(sig).ToNot(BeNil())
		})

		It("Decode must succeed", func() {
			msg, err = crp.Decode(sig)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg).ToNot(BeNil())
			Expect(msg).To(BeEquivalentTo([]byte("Hello World"))) // bytes.Equal(msg, []byte("Hello World"))(BeNil())
		})
	})

	Context("IO interface with encoding/decoding", func() {
		var (
			err error
			nbr int
			msg = []byte("Hello World")
			res = make([]byte, len(msg)*2)
			crp libenc.Coder
			key [32]byte
			non [12]byte
			buf = bytes.NewBuffer(make([]byte, 0, 32*1024))
			rdr io.Reader
			wrt io.Writer
		)

		It("Create key must succeed", func() {
			key, err = encaes.GenKey()
			Expect(key).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Create nonce must succeed", func() {
			non, err = encaes.GenNonce()
			Expect(non).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Create new instance must succeed", func() {
			crp, err = encaes.New(key, non)
			Expect(crp).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
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

		It("Create an io.reader and read from it to encode string but with small buffer occurs error", func() {
			res = make([]byte, 5)

			rdr = crp.EncodeReader(buf)
			Expect(rdr).ToNot(BeNil())

			buf.Reset()
			buf.Write(msg)

			nbr, err = rdr.Read(res)
			Expect(err).To(HaveOccurred())
		})

		It("Create an io.reader and read from it to encode string must succeed", func() {
			res = make([]byte, 50)

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
