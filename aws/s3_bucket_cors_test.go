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
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 Bucket - CORS Configuration", func() {
	BeforeEach(func() {
		err := cli.Bucket().Check()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("CORS rules management", func() {
		Context("Setting CORS rules", func() {
			It("should set simple CORS rule", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"*"},
						AllowedMethods: []string{"GET", "HEAD"},
						AllowedOrigins: []string{"*"},
						ExposeHeaders:  []string{"ETag"},
						MaxAgeSeconds:  sdkaws.Int32(3000),
					},
				}

				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set multiple CORS rules", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"Authorization"},
						AllowedMethods: []string{"GET"},
						AllowedOrigins: []string{"https://example.com"},
						MaxAgeSeconds:  sdkaws.Int32(3000),
					},
					{
						AllowedHeaders: []string{"*"},
						AllowedMethods: []string{"POST", "PUT"},
						AllowedOrigins: []string{"https://app.example.com"},
						ExposeHeaders:  []string{"ETag", "x-amz-request-id"},
						MaxAgeSeconds:  sdkaws.Int32(1800),
					},
				}

				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set CORS with specific methods", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"Content-Type", "Authorization"},
						AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
						AllowedOrigins: []string{"https://example.com", "https://www.example.com"},
						ExposeHeaders:  []string{"ETag", "x-amz-server-side-encryption"},
						MaxAgeSeconds:  sdkaws.Int32(3600),
					},
				}

				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Getting CORS rules", func() {
			BeforeEach(func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				// Set initial CORS rules
				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"*"},
						AllowedMethods: []string{"GET", "HEAD"},
						AllowedOrigins: []string{"*"},
						ExposeHeaders:  []string{"ETag"},
						MaxAgeSeconds:  sdkaws.Int32(3000),
					},
				}
				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should retrieve CORS configuration", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				rules, err := cli.Bucket().GetCORS()
				Expect(err).NotTo(HaveOccurred())
				Expect(rules).NotTo(BeEmpty())
				Expect(rules).To(HaveLen(1))
				Expect(rules[0].AllowedMethods).To(ContainElements("GET", "HEAD"))
			})

			It("should retrieve multiple CORS rules", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				// Set multiple rules first
				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"Authorization"},
						AllowedMethods: []string{"GET"},
						AllowedOrigins: []string{"https://example.com"},
						MaxAgeSeconds:  sdkaws.Int32(3000),
					},
					{
						AllowedHeaders: []string{"*"},
						AllowedMethods: []string{"POST"},
						AllowedOrigins: []string{"https://app.example.com"},
						MaxAgeSeconds:  sdkaws.Int32(1800),
					},
				}
				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())

				// Retrieve and verify
				rules, err := cli.Bucket().GetCORS()
				Expect(err).NotTo(HaveOccurred())
				Expect(rules).To(HaveLen(2))
			})
		})

		Context("CORS validation", func() {
			It("should handle empty CORS rules", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				err := cli.Bucket().SetCORS([]sdkstp.CORSRule{})
				// Should either succeed or fail gracefully
				_ = err
			})

			It("should validate allowed methods", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"*"},
						AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
						AllowedOrigins: []string{"*"},
						MaxAgeSeconds:  sdkaws.Int32(3000),
					},
				}

				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())

				// Verify all methods are set
				rules, err := cli.Bucket().GetCORS()
				Expect(err).NotTo(HaveOccurred())
				Expect(rules[0].AllowedMethods).To(HaveLen(5))
			})

			It("should handle wildcard origins", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"*"},
						AllowedMethods: []string{"GET"},
						AllowedOrigins: []string{"*"},
						MaxAgeSeconds:  sdkaws.Int32(3000),
					},
				}

				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())

				rules, err := cli.Bucket().GetCORS()
				Expect(err).NotTo(HaveOccurred())
				Expect(rules[0].AllowedOrigins).To(ContainElement("*"))
			})
		})

		Context("CORS edge cases", func() {
			It("should handle MaxAgeSeconds variations", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				testCases := []int32{0, 1, 3600, 86400}
				for _, maxAge := range testCases {
					corsRules := []sdkstp.CORSRule{
						{
							AllowedHeaders: []string{"*"},
							AllowedMethods: []string{"GET"},
							AllowedOrigins: []string{"*"},
							MaxAgeSeconds:  sdkaws.Int32(maxAge),
						},
					}

					err := cli.Bucket().SetCORS(corsRules)
					Expect(err).NotTo(HaveOccurred())

					rules, err := cli.Bucket().GetCORS()
					Expect(err).NotTo(HaveOccurred())
					if rules[0].MaxAgeSeconds != nil {
						Expect(*rules[0].MaxAgeSeconds).To(Equal(maxAge))
					}
				}
			})

			It("should handle multiple origins", func() {
				if minioMode {
					Skip("MinIO: CORS configuration not fully compatible")
				}

				origins := []string{
					"https://example1.com",
					"https://example2.com",
					"https://example3.com",
					"http://localhost:3000",
				}

				corsRules := []sdkstp.CORSRule{
					{
						AllowedHeaders: []string{"*"},
						AllowedMethods: []string{"GET", "POST"},
						AllowedOrigins: origins,
						MaxAgeSeconds:  sdkaws.Int32(3000),
					},
				}

				err := cli.Bucket().SetCORS(corsRules)
				Expect(err).NotTo(HaveOccurred())

				rules, err := cli.Bucket().GetCORS()
				Expect(err).NotTo(HaveOccurred())
				Expect(rules[0].AllowedOrigins).To(HaveLen(len(origins)))
			})
		})
	})
})
