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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Context("Creation", func() {
		It("Must be possible to check if the bucket exists", func() {
			Expect(cli.Bucket().Check()).ToNot(Succeed())
		})
		It("Must be possible to create a bucket", func() {
			Expect(cli.Bucket().Create("")).To(Succeed())
			Expect(cli.Bucket().Check()).To(Succeed())
		})
	})
	Context("Find object", func() {
		Context("With no object in bucket", func() {
			It("Must succeed and return no object", func() {
				objects, err := cli.Object().Find("pattern")
				Expect(err).ToNot(HaveOccurred())
				Expect(objects).To(HaveLen(0))
			})
		})
		Context("With the object", func() {
			It("Must succeed", func() {
				var (
					err error
					rnd = randContent(64 * 1024)
				)

				err = cli.Object().MultipartPut("object", rnd)
				Expect(err).ToNot(HaveOccurred())

				objects, err := cli.Object().Find("object")
				Expect(err).ToNot(HaveOccurred())
				Expect(objects).To(HaveLen(1))

				err = cli.Object().Delete("object")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
	Context("List", func() {
		It("Must be possible to list buckets", func() {
			buckets, err := cli.Bucket().List()
			Expect(err).ToNot(HaveOccurred())
			Expect(buckets).To(HaveLen(1))
		})
	})

	Context("Versioning", func() {
		It("Must be possible to enable versioning", func() {
			var err error
			if !minioMode {
				err = cli.Bucket().SetVersioning(true)
			}
			Expect(err).To(Succeed())
		})
		It("Must be enabled", func() {
			var (
				err error
				sts string
			)
			if !minioMode {
				sts, err = cli.Bucket().GetVersioning()
			} else {
				sts = "Enabled"
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(sts).To(Equal("Enabled"))
		})
		It("Must be possible to suspend versioning", func() {
			var err error
			if !minioMode {
				err = cli.Bucket().SetVersioning(false)
			}
			Expect(err).To(Succeed())
		})
	})
	Context("Replication", func() {
		Context("Enable with invalid params", func() {
			It("Must fail", func() {
				Expect(cli.Bucket().EnableReplication("fake-src-role-arn", "fake-dst-role-arn", "fake-dst-bucket")).ToNot(Succeed())
			})
		})
		Context("Disable", func() {
			It("Must not return error", func() {
				var err error
				if !minioMode {
					err = cli.Bucket().DeleteReplication()
				}
				Expect(err).To(Succeed())
			})
		})
	})

	It("Must be possible to delete a bucket", func() {
		Expect(cli.Bucket().Delete()).To(Succeed())
	})
})
