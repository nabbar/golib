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

package randRead_test

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"strconv"

	encrnd "github.com/nabbar/golib/encoding/randRead"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func simRequest() (io.ReadCloser, error) {
	var p = make([]byte, 1)
	if n, e := rand.Read(p); e != nil {
		return nil, e
	} else if n > 0 {
		return io.NopCloser(bytes.NewBuffer(p)), nil
	} else {
		return nil, errors.New("empty buffer")
	}
}

var _ = Describe("encoding/randRead", func() {
	Context("complete Random Reader", func() {
		var (
			err error
			nbr int64
			rnd io.ReadCloser
		)

		It("must succeed when create new random reader", func() {
			rnd = encrnd.New(simRequest)
			Expect(rnd).ToNot(BeNil())
		})

		for i := 0; i < 10; i++ {
			It("must succeed when using random reader on iteration #"+strconv.Itoa(i), func() {
				nbr = 0
				err = binary.Read(rnd, binary.BigEndian, &nbr)
				Expect(err).ToNot(HaveOccurred())
				Expect(nbr).ToNot(BeNumerically("==", 0))
			})
		}
	})
})
