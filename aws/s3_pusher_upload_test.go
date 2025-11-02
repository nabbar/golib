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

var _ = Describe("S3 Pusher - Upload Operations", func() {
	BeforeEach(func() {
		// Ensure bucket exists (created in BeforeSuite)
		err := cli.Bucket().Check()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Pusher configuration and lifecycle", func() {
		Context("Creating a pusher instance", func() {
			It("New() should create valid pusher", func() {
				cfg := &awspsh.Config{
					FuncGetClientS3: func() *sdksss.Client {
						return cli.GetClientS3()
					},
					ObjectS3Options: awspsh.ConfigObjectOptions{
						Bucket: sdkaws.String(cli.GetBucketName()),
						Key:    sdkaws.String("pusher-test-object"),
					},
				}

				psh, err := awspsh.New(ctx, cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(psh).NotTo(BeNil())

				// Cleanup
				err = psh.Abort()
				Expect(err).NotTo(HaveOccurred())
			})

			It("New() should fail with nil config", func() {
				psh, err := awspsh.New(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(psh).To(BeNil())
			})
		})
	})

	Describe("Small file upload (single part)", func() {
		It("should upload file smaller than part size", func() {
			objectKey := "pusher-small-file.dat"
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
						Key:    sdkaws.String(objectKey),
					},
				}
			)

			// Create pusher
			psh, err = awspsh.New(ctx, pcf)
			Expect(err).NotTo(HaveOccurred())
			Expect(psh).NotTo(BeNil())

			// Write data (500KB < default part size)
			nbr, err = io.Copy(psh, randContent(500*libsiz.SizeKilo))
			Expect(err).NotTo(HaveOccurred())
			Expect(nbr).To(Equal(int64(500 * libsiz.SizeKilo)))

			// Complete upload
			err = psh.Complete()
			Expect(err).NotTo(HaveOccurred())

			// Close should fail after Complete
			err = psh.Close()
			Expect(err).To(HaveOccurred())

			// Verify object exists
			objects, err := cli.Object().Find(objectKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(1))

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})
	})

	Describe("Large file upload (multipart)", func() {
		It("should upload file larger than part size", func() {
			objectKey := "pusher-large-file.dat"
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
						Key:    sdkaws.String(objectKey),
					},
					PartSize: 5 * 1024 * 1024, // 5MB parts
				}
			)

			// Create pusher
			psh, err = awspsh.New(ctx, pcf)
			Expect(err).NotTo(HaveOccurred())
			Expect(psh).NotTo(BeNil())

			// Write large data (50MB > part size)
			nbr, err = io.Copy(psh, randContent(50*libsiz.SizeMega))
			Expect(err).NotTo(HaveOccurred())
			Expect(nbr).To(Equal(int64(50 * libsiz.SizeMega)))

			// Complete upload
			err = psh.Complete()
			Expect(err).NotTo(HaveOccurred())

			// Close should fail after Complete
			err = psh.Close()
			Expect(err).To(HaveOccurred())

			// Verify object exists
			objects, err := cli.Object().Find(objectKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(1))

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})
	})

	Describe("Upload abortion", func() {
		It("should abort multipart upload successfully", func() {
			objectKey := "pusher-abort-test.dat"
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
						Key:    sdkaws.String(objectKey),
					},
				}
			)

			// Create pusher
			psh, err = awspsh.New(ctx, pcf)
			Expect(err).NotTo(HaveOccurred())

			// Write data
			nbr, err = io.Copy(psh, randContent(50*libsiz.SizeMega))
			Expect(err).NotTo(HaveOccurred())
			Expect(nbr).To(Equal(int64(50 * libsiz.SizeMega)))

			// Abort (Close without Complete)
			err = psh.Close()
			Expect(err).NotTo(HaveOccurred())

			// Close again should fail
			err = psh.Close()
			Expect(err).To(HaveOccurred())

			// Verify object doesn't exist
			objects, err := cli.Object().Find(objectKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(0))

			// Cleanup attempt (should fail gracefully)
			_ = cli.Object().Delete(false, objectKey)
		})
	})

	Describe("Pusher with custom options", func() {
		It("should upload with checksum enabled", func() {
			if minioMode {
				Skip("MinIO: Checksum validation not fully compatible with multipart")
			}

			objectKey := "pusher-checksum.dat"
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
						Key:    sdkaws.String(objectKey),
					},
					CheckSum:   true,
					PartSize:   10 * 1024 * 1024, // 10MB
					BufferSize: 4096,
				}
			)

			psh, err = awspsh.New(ctx, pcf)
			Expect(err).NotTo(HaveOccurred())

			nbr, err = io.Copy(psh, randContent(15*libsiz.SizeMega))
			Expect(err).NotTo(HaveOccurred())
			Expect(nbr).To(Equal(int64(15 * libsiz.SizeMega)))

			err = psh.Complete()
			Expect(err).NotTo(HaveOccurred())

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})

		It("should upload with custom metadata", func() {
			objectKey := "pusher-metadata.dat"
			var (
				err error
				psh awspsh.Pusher
				pcf = &awspsh.Config{
					FuncGetClientS3: func() *sdksss.Client {
						return cli.GetClientS3()
					},
					ObjectS3Options: awspsh.ConfigObjectOptions{
						Bucket: sdkaws.String(cli.GetBucketName()),
						Key:    sdkaws.String(objectKey),
						Metadata: map[string]string{
							"test-key": "test-value",
						},
						ContentType: sdkaws.String("application/octet-stream"),
					},
				}
			)

			psh, err = awspsh.New(ctx, pcf)
			Expect(err).NotTo(HaveOccurred())

			_, err = io.Copy(psh, randContent(1*libsiz.SizeMega))
			Expect(err).NotTo(HaveOccurred())

			err = psh.Complete()
			Expect(err).NotTo(HaveOccurred())

			// Verify metadata
			head, err := cli.Object().Head(objectKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(head.ContentType).NotTo(BeNil())
			Expect(*head.ContentType).To(Equal("application/octet-stream"))

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})
	})

	Describe("Error handling", func() {
		It("Complete() should fail if called twice", func() {
			objectKey := "pusher-double-complete.dat"
			var (
				err error
				psh awspsh.Pusher
				pcf = &awspsh.Config{
					FuncGetClientS3: func() *sdksss.Client {
						return cli.GetClientS3()
					},
					ObjectS3Options: awspsh.ConfigObjectOptions{
						Bucket: sdkaws.String(cli.GetBucketName()),
						Key:    sdkaws.String(objectKey),
					},
				}
			)

			psh, err = awspsh.New(ctx, pcf)
			Expect(err).NotTo(HaveOccurred())

			_, err = io.Copy(psh, randContent(1*libsiz.SizeMega))
			Expect(err).NotTo(HaveOccurred())

			err = psh.Complete()
			Expect(err).NotTo(HaveOccurred())

			// Second Complete should fail
			err = psh.Complete()
			Expect(err).To(HaveOccurred())

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})

		It("Abort() after Complete() should be handled", func() {
			objectKey := "pusher-abort-after-complete.dat"
			var (
				err error
				psh awspsh.Pusher
				pcf = &awspsh.Config{
					FuncGetClientS3: func() *sdksss.Client {
						return cli.GetClientS3()
					},
					ObjectS3Options: awspsh.ConfigObjectOptions{
						Bucket: sdkaws.String(cli.GetBucketName()),
						Key:    sdkaws.String(objectKey),
					},
				}
			)

			psh, err = awspsh.New(ctx, pcf)
			Expect(err).NotTo(HaveOccurred())

			_, err = io.Copy(psh, randContent(1*libsiz.SizeMega))
			Expect(err).NotTo(HaveOccurred())

			err = psh.Complete()
			Expect(err).NotTo(HaveOccurred())

			// Abort after Complete may error or succeed depending on implementation
			// Just verify it doesn't panic
			_ = psh.Abort()

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})
	})
})
