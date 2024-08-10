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
	"io"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	awspsh "github.com/nabbar/golib/aws/pusher"
	libsiz "github.com/nabbar/golib/size"
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
		Context("With the object sent as mpu", func() {
			It("Must succeed", func() {
				var (
					err error
					rnd = randContent(10 * libsiz.SizeMega)
				)

				err = cli.Object().MultipartPut("object", rnd)
				Expect(err).ToNot(HaveOccurred())

				objects, err := cli.Object().Find("object")
				Expect(err).ToNot(HaveOccurred())
				Expect(objects).To(HaveLen(1))

				err = cli.Object().Delete(true, "object")
				Expect(err).ToNot(HaveOccurred())

				err = cli.Object().Delete(false, "object")
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("With a large object sent with pusher instance", func() {
			It("Must succeed", func() {
				var (
					err error
					nbr int64
					psh awspsh.Pusher
					pcf = &awspsh.Config{
						FuncGetClientS3: func() *sdksss.Client {
							return cli.GetClientS3()
						},
						ObjectS3Options: awspsh.ConfigObjectOptions{
							Bucket: sdkaws.String(cli.GetBucketName()),
							Key:    sdkaws.String("object"),
						},
					}
				)

				psh, err = awspsh.New(ctx, pcf)
				Expect(err).ToNot(HaveOccurred())

				nbr, err = io.Copy(psh, randContent(50*libsiz.SizeMega))
				Expect(err).ToNot(HaveOccurred())
				Expect(nbr).To(BeNumerically("==", 50*libsiz.SizeMega))

				err = psh.Complete()
				Expect(err).ToNot(HaveOccurred())

				err = psh.Close()
				Expect(err).To(HaveOccurred())

				objects, err := cli.Object().Find("object")
				Expect(err).ToNot(HaveOccurred())
				Expect(objects).To(HaveLen(1))

				err = cli.Object().Delete(true, "object")
				Expect(err).ToNot(HaveOccurred())

				err = cli.Object().Delete(false, "object")
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("With a small object sent with pusher instance", func() {
			It("Must succeed", func() {
				var (
					err error
					nbr int64
					psh awspsh.Pusher
					pcf = &awspsh.Config{
						FuncGetClientS3: func() *sdksss.Client {
							return cli.GetClientS3()
						},
						ObjectS3Options: awspsh.ConfigObjectOptions{
							Bucket: sdkaws.String(cli.GetBucketName()),
							Key:    sdkaws.String("object"),
						},
					}
				)

				psh, err = awspsh.New(ctx, pcf)
				Expect(err).ToNot(HaveOccurred())

				nbr, err = io.Copy(psh, randContent(500*libsiz.SizeKilo))
				Expect(err).ToNot(HaveOccurred())
				Expect(nbr).To(BeNumerically("==", 500*libsiz.SizeKilo))

				err = psh.Complete()
				Expect(err).ToNot(HaveOccurred())

				err = psh.Close()
				Expect(err).To(HaveOccurred())

				objects, err := cli.Object().Find("object")
				Expect(err).ToNot(HaveOccurred())
				Expect(objects).To(HaveLen(1))

				err = cli.Object().Delete(true, "object")
				Expect(err).ToNot(HaveOccurred())

				err = cli.Object().Delete(false, "object")
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("With a large object sent with pusher instance but abort", func() {
			It("Must fail", func() {
				var (
					err error
					nbr int64
					obj []string
					psh awspsh.Pusher
					pcf = &awspsh.Config{
						FuncGetClientS3: func() *sdksss.Client {
							return cli.GetClientS3()
						},
						ObjectS3Options: awspsh.ConfigObjectOptions{
							Bucket: sdkaws.String(cli.GetBucketName()),
							Key:    sdkaws.String("object"),
						},
					}
				)

				psh, err = awspsh.New(ctx, pcf)
				Expect(err).ToNot(HaveOccurred())

				nbr, err = io.Copy(psh, randContent(50*libsiz.SizeMega))
				Expect(err).ToNot(HaveOccurred())
				Expect(nbr).To(BeNumerically("==", 50*libsiz.SizeMega))

				err = psh.Close()
				Expect(err).ToNot(HaveOccurred())

				err = psh.Close()
				Expect(err).To(HaveOccurred())

				obj, err = cli.Object().Find("object")
				Expect(err).ToNot(HaveOccurred())
				Expect(obj).To(HaveLen(0))

				err = cli.Object().Delete(true, "object")
				Expect(err).To(HaveOccurred())

				err = cli.Object().Delete(false, "object")
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
