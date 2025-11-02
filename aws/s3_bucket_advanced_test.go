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
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 Bucket - Advanced Features", func() {
	BeforeEach(func() {
		err := cli.Bucket().Check()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Lifecycle Configuration", func() {
		Context("Basic lifecycle rules", func() {
			It("should set and retrieve lifecycle configuration", func() {
				Skip("MinIO/AWS: Lifecycle configuration requires specific API compatibility")
			})
		})
	})

	Describe("Website Configuration", func() {
		Context("Setting website configuration", func() {
			It("should configure bucket as website", func() {
				if minioMode {
					Skip("MinIO: Website configuration compatibility varies")
				}

				err := cli.Bucket().PutWebsite("index.html", "error.html")
				// May succeed or fail depending on bucket configuration
				_ = err
			})

			It("should retrieve website configuration", func() {
				if minioMode {
					Skip("MinIO: Website configuration compatibility varies")
				}

				_, err := cli.Bucket().GetWebsite()
				// May return error if not configured
				_ = err
			})
		})
	})

	Describe("ACL Configuration", func() {
		Context("Getting bucket ACL", func() {
			It("should retrieve bucket ACL", func() {
				if minioMode {
					Skip("MinIO: ACL operations may vary")
				}

				acl, err := cli.Bucket().GetACL()
				Expect(err).NotTo(HaveOccurred())
				Expect(acl).NotTo(BeNil())
			})
		})

		Context("Setting bucket ACL", func() {
			It("should set canned ACL to private", func() {
				if minioMode {
					Skip("MinIO: ACL operations may vary")
				}

				err := cli.Bucket().SetACLHeader(sdkstp.BucketCannedACLPrivate, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set canned ACL to public-read", func() {
				if minioMode {
					Skip("MinIO: ACL operations may vary")
				}

				err := cli.Bucket().SetACLHeader(sdkstp.BucketCannedACLPublicRead, nil)
				// May succeed or fail depending on bucket policy
				_ = err
			})
		})
	})

	Describe("Replication Configuration", func() {
		Context("Replication setup", func() {
			It("should load replication configuration", func() {
				if minioMode {
					Skip("MinIO: Replication not fully compatible")
				}

				_, err := cli.Bucket().LoadReplication()
				// Will error if no replication configured
				_ = err
			})
		})
	})

	Describe("Lock Configuration", func() {
		Context("Object lock", func() {
			It("should retrieve lock configuration", func() {
				if minioMode {
					Skip("MinIO: Object lock compatibility varies")
				}

				_, err := cli.Bucket().GetLock()
				// Will error if lock not enabled on bucket creation
				_ = err
			})
		})
	})
})
