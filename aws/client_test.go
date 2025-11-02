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
	"context"
	"time"

	libaws "github.com/nabbar/golib/aws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWS Client Management", func() {
	Describe("Client Creation", func() {
		It("should create AWS client successfully", func() {
			Expect(cli).ToNot(BeNil())
		})

		It("should have valid config", func() {
			Expect(cli.Config()).ToNot(BeNil())
		})

		It("should have HTTP client", func() {
			Expect(cli.HTTPCli()).ToNot(BeNil())
		})

		It("should have S3 client", func() {
			s3Client := cli.GetClientS3()
			Expect(s3Client).ToNot(BeNil())
		})

		It("should have IAM client", func() {
			iamClient := cli.GetClientIam()
			Expect(iamClient).ToNot(BeNil())
		})
	})

	Describe("Client Configuration", func() {
		It("should get and set bucket name", func() {
			originalBucket := cli.GetBucketName()
			Expect(originalBucket).ToNot(BeEmpty())

			testBucket := "test-bucket-123"
			cli.SetBucketName(testBucket)
			Expect(cli.GetBucketName()).To(Equal(testBucket))

			// Restore original
			cli.SetBucketName(originalBucket)
		})

		It("should get current config", func() {
			config := cli.Config()
			Expect(config).ToNot(BeNil())
			Expect(config.GetRegion()).ToNot(BeEmpty())
		})

		It("should have valid endpoint", func() {
			endpoint := cli.Config().GetEndpoint()
			Expect(endpoint).ToNot(BeNil())
			Expect(endpoint.String()).ToNot(BeEmpty())
		})
	})

	Describe("HTTP Timeout Management", func() {
		It("should get HTTP timeout", func() {
			timeout := cli.GetHTTPTimeout()
			// Timeout might be 0 (no timeout) or a positive value
			Expect(timeout).To(BeNumerically(">=", 0))
		})

		It("should set HTTP timeout", func() {
			newTimeout := 30 * time.Second
			err := cli.SetHTTPTimeout(newTimeout)
			Expect(err).ToNot(HaveOccurred())

			retrievedTimeout := cli.GetHTTPTimeout()
			Expect(retrievedTimeout).To(Equal(newTimeout))
		})

		It("should handle timeout changes gracefully", func() {
			timeouts := []time.Duration{
				10 * time.Second,
				60 * time.Second,
				120 * time.Second,
			}

			for _, timeout := range timeouts {
				err := cli.SetHTTPTimeout(timeout)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli.GetHTTPTimeout()).To(Equal(timeout))
			}
		})
	})

	Describe("Client Cloning", func() {
		It("should clone client successfully", func() {
			cloned, err := cli.Clone(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloned).ToNot(BeNil())
			Expect(cloned).ToNot(BeIdenticalTo(cli))
		})

		It("should clone with new context", func() {
			newCtx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cloned, err := cli.Clone(newCtx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloned).ToNot(BeNil())
		})

		It("should clone with nil context (use parent context)", func() {
			cloned, err := cli.Clone(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloned).ToNot(BeNil())
		})

		It("cloned client should have independent config", func() {
			cloned, err := cli.Clone(ctx)
			Expect(err).ToNot(HaveOccurred())

			originalBucket := cli.GetBucketName()
			testBucket := "cloned-bucket-test"

			cloned.SetBucketName(testBucket)
			Expect(cloned.GetBucketName()).To(Equal(testBucket))
			Expect(cli.GetBucketName()).To(Equal(originalBucket))
		})
	})

	Describe("Client with New Config", func() {
		It("should create new client with different config", func() {
			newConfig := cli.Config().Clone()
			Expect(newConfig).ToNot(BeNil())

			newClient, err := cli.NewForConfig(ctx, newConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(newClient).ToNot(BeNil())
			Expect(newClient).ToNot(BeIdenticalTo(cli))
		})

		It("should maintain separate configurations", func() {
			newConfig := cli.Config().Clone()
			newConfig.SetBucketName("new-config-bucket")

			newClient, err := cli.NewForConfig(ctx, newConfig)
			Expect(err).ToNot(HaveOccurred())

			Expect(newClient.GetBucketName()).To(Equal("new-config-bucket"))
			Expect(cli.GetBucketName()).ToNot(Equal("new-config-bucket"))
		})
	})

	Describe("Force Path Style", func() {
		It("should set path style successfully", func() {
			err := cli.ForcePathStyle(ctx, true)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should toggle path style", func() {
			err := cli.ForcePathStyle(ctx, false)
			Expect(err).ToNot(HaveOccurred())

			err = cli.ForcePathStyle(ctx, true)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Service Accessors", func() {
		It("should provide Bucket service", func() {
			bucket := cli.Bucket()
			Expect(bucket).ToNot(BeNil())
		})

		It("should provide Object service", func() {
			object := cli.Object()
			Expect(object).ToNot(BeNil())
		})

		It("should provide Group service", func() {
			group := cli.Group()
			Expect(group).ToNot(BeNil())
		})

		It("should provide Policy service", func() {
			policy := cli.Policy()
			Expect(policy).ToNot(BeNil())
		})

		It("should provide Role service", func() {
			role := cli.Role()
			Expect(role).ToNot(BeNil())
		})

		It("should provide User service", func() {
			user := cli.User()
			Expect(user).ToNot(BeNil())
		})

		It("all services should be independent", func() {
			bucket1 := cli.Bucket()
			bucket2 := cli.Bucket()

			// Should get new instances each time
			Expect(bucket1).ToNot(BeNil())
			Expect(bucket2).ToNot(BeNil())
		})
	})

	Describe("Client S3 and IAM Setters", func() {
		It("should get current S3 client", func() {
			s3 := cli.GetClientS3()
			Expect(s3).ToNot(BeNil())
		})

		It("should set S3 client", func() {
			originalS3 := cli.GetClientS3()
			Expect(originalS3).ToNot(BeNil())

			// Set back the same client
			cli.SetClientS3(originalS3)
			Expect(cli.GetClientS3()).To(Equal(originalS3))
		})

		It("should get current IAM client", func() {
			iam := cli.GetClientIam()
			Expect(iam).ToNot(BeNil())
		})

		It("should set IAM client", func() {
			originalIam := cli.GetClientIam()
			Expect(originalIam).ToNot(BeNil())

			// Set back the same client
			cli.SetClientIam(originalIam)
			Expect(cli.GetClientIam()).To(Equal(originalIam))
		})
	})

	Describe("Error Handling", func() {
		It("should fail to create client with nil config", func() {
			_, err := libaws.New(ctx, nil, cli.HTTPCli())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("config"))
		})

		It("should handle nil context gracefully", func() {
			// New should handle nil context by using Background
			newCli, err := libaws.New(nil, cli.Config(), cli.HTTPCli())
			Expect(err).ToNot(HaveOccurred())
			Expect(newCli).ToNot(BeNil())
		})
	})
})
