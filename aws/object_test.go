/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package aws_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Object", func() {
	Context("List objects", func() {
		It("Must fail with invalid token -1 ", func() {
			_, _, _, err := cli.Object().List("token")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Put object", func() {
		It("Must fail as the bucket doesn't exists - 2", func() {
			err := cli.Object().Put("object", bytes.NewReader([]byte("")))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Get object", func() {
		It("Must fail as the bucket doesn't exists - 3", func() {
			o, err := cli.Object().Get("object")

			defer func() {
				if o != nil && o.Body != nil {
					_ = o.Body.Close()
				}
			}()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("Delete object", func() {
		It("Must fail as the object doesn't exists - 4", func() {
			err := cli.Object().Delete("object")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Multipart Put object", func() {
		It("Must fail as the bucket doesn't exists - 5", func() {
			err := cli.Object().MultipartPut("object", randContent(4*1024))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Delete object", func() {
		It("Must fail as the object doesn't exists - 6", func() {
			err := cli.Object().Delete("object")
			Expect(err).To(HaveOccurred())
		})
	})

})
