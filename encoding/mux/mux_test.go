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

package mux_test

import (
	"bytes"
	"fmt"
	"io"

	encmux "github.com/nabbar/golib/encoding/mux"
	libsck "github.com/nabbar/golib/socket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("encoding/mux", func() {
	Context("complete mux/demux", func() {
		var (
			err error
			nbr int
			mux encmux.Multiplexer
			dmx encmux.DeMultiplexer

			buf = bytes.NewBuffer(make([]byte, 0, 32*1024)) // multiplexed buffer

			bsa io.Writer                                   // stream for 'a' source
			bra = bytes.NewBuffer(make([]byte, 0, 32*1024)) // buffer for 'a' result

			bsb io.Writer                                   // stream for 'b' source
			brb = bytes.NewBuffer(make([]byte, 0, 32*1024)) // buffer for 'b' result
		)

		It("Create new multiplexer must succeed", func() {
			mux = encmux.NewMultiplexer(buf, libsck.EOL)
			Expect(mux).ToNot(BeNil())
		})

		It("Create new de-multiplexer must succeed", func() {
			dmx = encmux.NewDeMultiplexer(buf, libsck.EOL, 0)
			Expect(dmx).ToNot(BeNil())
		})

		It("Create new channel must succeed", func() {
			dmx.NewChannel('a', bra)
			bsa = mux.NewChannel('a')
			Expect(bsa).NotTo(BeNil())

			dmx.NewChannel('b', brb)
			bsb = mux.NewChannel('b')
			Expect(bsb).NotTo(BeNil())
		})

		It("sending on an io.writer must succeed", func() {
			nbr, err = fmt.Fprintln(bsa, "Hello World")
			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeEquivalentTo(12))
		})

		It("sending on an io.writer must succeed", func() {
			nbr, err = fmt.Fprintln(bsb, "Hello World")
			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeEquivalentTo(12))
		})

		It("sending on an io.writer must succeed", func() {
			nbr, err = fmt.Fprintln(bsa, "Hello World #2\nHello World #3"+string(libsck.EOL)+"!!")
			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeEquivalentTo(33))
		})

		It("sending on an io.writer must succeed", func() {
			nbr, err = fmt.Fprintln(bsa, "Hello World #3")
			Expect(err).ToNot(HaveOccurred())
			Expect(nbr).To(BeEquivalentTo(15))
		})

		It("Reading on an io.reader must succeed", func() {
			err = dmx.Copy()
			Expect(err).ToNot(HaveOccurred())
			Expect(bra.Len()).To(BeEquivalentTo(60))
			Expect(brb.Len()).To(BeEquivalentTo(12))
		})
	})
})
