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
	"net/url"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/nabbar/golib/aws/configCustom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWS Configuration Management", func() {
	Describe("Configuration Creation", func() {
		It("should create config successfully", func() {
			endpoint, _ := url.Parse("http://localhost:9000")
			config := awscfg.NewConfig("test-bucket", "access", "secret", endpoint, "us-east-1")
			Expect(config).ToNot(BeNil())
		})

		It("should validate configuration", func() {
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should check configuration", func() {
			err := cfg.Check(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Credentials Management", func() {
		It("should get access key", func() {
			accessKey := cfg.GetAccessKey()
			Expect(accessKey).ToNot(BeEmpty())
		})

		It("should get secret key", func() {
			secretKey := cfg.GetSecretKey()
			Expect(secretKey).ToNot(BeEmpty())
		})

		It("should set credentials", func() {
			originalAccess := cfg.GetAccessKey()
			originalSecret := cfg.GetSecretKey()

			cfg.SetCredentials("new-access", "new-secret")
			Expect(cfg.GetAccessKey()).To(Equal("new-access"))
			Expect(cfg.GetSecretKey()).To(Equal("new-secret"))

			// Restore
			cfg.SetCredentials(originalAccess, originalSecret)
		})
	})

	Describe("Region Management", func() {
		It("should get region", func() {
			region := cfg.GetRegion()
			Expect(region).ToNot(BeEmpty())
		})

		It("should set region", func() {
			originalRegion := cfg.GetRegion()

			cfg.SetRegion("eu-west-1")
			Expect(cfg.GetRegion()).To(Equal("eu-west-1"))

			// Restore
			cfg.SetRegion(originalRegion)
		})

		It("should get resolved region", func() {
			resolved := cfg.GetResolvedRegion()
			Expect(resolved).ToNot(BeEmpty())
		})
	})

	Describe("Endpoint Management", func() {
		It("should get endpoint", func() {
			endpoint := cfg.GetEndpoint()
			Expect(endpoint).ToNot(BeNil())
			Expect(endpoint.String()).ToNot(BeEmpty())
		})

		It("should set endpoint", func() {
			originalEndpoint := cfg.GetEndpoint()

			newEndpoint, _ := url.Parse("http://localhost:9001")
			cfg.SetEndpoint(newEndpoint)
			Expect(cfg.GetEndpoint().String()).To(Equal("http://localhost:9001"))

			// Restore
			cfg.SetEndpoint(originalEndpoint)
		})

		It("should check HTTPS status", func() {
			isHTTPS := cfg.IsHTTPs()
			Expect(isHTTPS).To(BeAssignableToTypeOf(false))
		})

		It("should get disable HTTPS setting", func() {
			disabled := cfg.GetDisableHTTPS()
			Expect(disabled).To(BeAssignableToTypeOf(false))
		})
	})

	Describe("Region and Endpoint Registration", func() {
		It("should register AWS region", func() {
			err := cfg.RegisterRegionAws(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should register custom region endpoint", func() {
			endpoint, _ := url.Parse("http://custom.endpoint.local:9000")
			err := cfg.RegisterRegionEndpoint("custom-region", endpoint)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reset region endpoints", func() {
			cfg.ResetRegionEndpoint()
			// Should not panic
		})
	})

	Describe("Bucket Management", func() {
		It("should get bucket name", func() {
			bucket := cfg.GetBucketName()
			Expect(bucket).ToNot(BeEmpty())
		})

		It("should set bucket name", func() {
			originalBucket := cfg.GetBucketName()

			cfg.SetBucketName("new-test-bucket")
			Expect(cfg.GetBucketName()).To(Equal("new-test-bucket"))

			// Restore
			cfg.SetBucketName(originalBucket)
		})
	})

	Describe("Checksum Validation", func() {
		It("should get checksum validation settings", func() {
			req, rsp := cfg.GetChecksumValidation()
			Expect(req).To(BeAssignableToTypeOf(sdkaws.RequestChecksumCalculationWhenSupported))
			Expect(rsp).To(BeAssignableToTypeOf(sdkaws.ResponseChecksumValidationWhenSupported))
		})

		It("should set checksum validation", func() {
			cfg.SetChecksumValidation(
				sdkaws.RequestChecksumCalculationWhenRequired,
				sdkaws.ResponseChecksumValidationWhenRequired,
			)

			req, rsp := cfg.GetChecksumValidation()
			Expect(req).To(Equal(sdkaws.RequestChecksumCalculationWhenRequired))
			Expect(rsp).To(Equal(sdkaws.ResponseChecksumValidationWhenRequired))
		})
	})

	Describe("Configuration Cloning", func() {
		It("should clone configuration", func() {
			cloned := cfg.Clone()
			Expect(cloned).ToNot(BeNil())
			Expect(cloned).ToNot(BeIdenticalTo(cfg))
		})

		It("cloned config should be independent", func() {
			cloned := cfg.Clone()

			originalBucket := cfg.GetBucketName()
			cloned.SetBucketName("cloned-bucket")

			Expect(cloned.GetBucketName()).To(Equal("cloned-bucket"))
			Expect(cfg.GetBucketName()).To(Equal(originalBucket))
		})
	})

	Describe("JSON Serialization", func() {
		It("should serialize to JSON", func() {
			jsonData, err := cfg.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(jsonData).ToNot(BeEmpty())
		})

		It("should deserialize from JSON", func() {
			jsonData, err := cfg.JSON()
			Expect(err).ToNot(HaveOccurred())

			newConfig, err := awscfg.NewConfigJsonUnmashal(jsonData)
			Expect(err).ToNot(HaveOccurred())
			Expect(newConfig).ToNot(BeNil())
			Expect(newConfig.GetRegion()).To(Equal(cfg.GetRegion()))
		})
	})

	Describe("AWS Config Generation", func() {
		It("should generate AWS SDK config", func() {
			awsConfig, err := cfg.GetConfig(ctx, cli.HTTPCli())
			Expect(err).ToNot(HaveOccurred())
			Expect(awsConfig).ToNot(BeNil())
		})

		It("AWS config should have credentials", func() {
			awsConfig, err := cfg.GetConfig(ctx, cli.HTTPCli())
			Expect(err).ToNot(HaveOccurred())

			creds, err := awsConfig.Credentials.Retrieve(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(creds.AccessKeyID).ToNot(BeEmpty())
			Expect(creds.SecretAccessKey).ToNot(BeEmpty())
		})

		It("AWS config should have region", func() {
			awsConfig, err := cfg.GetConfig(ctx, cli.HTTPCli())
			Expect(err).ToNot(HaveOccurred())
			Expect(awsConfig.Region).To(Equal(cfg.GetRegion()))
		})
	})

	Describe("Endpoint Resolution", func() {
		It("should resolve S3 endpoint", func() {
			endpoint, err := cfg.ResolveEndpoint("s3", cfg.GetRegion())
			Expect(err).ToNot(HaveOccurred())
			Expect(endpoint.URL).ToNot(BeEmpty())
		})

		It("should resolve IAM endpoint", func() {
			endpoint, err := cfg.ResolveEndpoint("iam", cfg.GetRegion())
			Expect(err).ToNot(HaveOccurred())
			Expect(endpoint.URL).ToNot(BeEmpty())
		})

		It("should resolve endpoint with options", func() {
			endpoint, err := cfg.ResolveEndpointWithOptions("s3", cfg.GetRegion())
			Expect(err).ToNot(HaveOccurred())
			Expect(endpoint.URL).ToNot(BeEmpty())
		})
	})
})
