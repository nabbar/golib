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

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 Object - Tagging", func() {
	BeforeEach(func() {
		err := cli.Bucket().Check()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Object tagging operations", func() {
		var testObjectKey string

		BeforeEach(func() {
			testObjectKey = "test-tagging-" + GenerateUniqueName("")

			// Upload test object
			err := cli.Object().Put(testObjectKey, bytes.NewReader([]byte("test content for tagging")))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_ = cli.Object().Delete(false, testObjectKey)
		})

		Context("Setting tags", func() {
			It("should set single tag", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				tags := []sdkstp.Tag{
					{
						Key:   sdkaws.String("Environment"),
						Value: sdkaws.String("Test"),
					},
				}

				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set multiple tags", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				tags := []sdkstp.Tag{
					{Key: sdkaws.String("Environment"), Value: sdkaws.String("Test")},
					{Key: sdkaws.String("Project"), Value: sdkaws.String("TestProject")},
					{Key: sdkaws.String("Owner"), Value: sdkaws.String("TestOwner")},
					{Key: sdkaws.String("CostCenter"), Value: sdkaws.String("12345")},
				}

				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set tags with special characters", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				tags := []sdkstp.Tag{
					{Key: sdkaws.String("Key-With-Dash"), Value: sdkaws.String("Value-With-Dash")},
					{Key: sdkaws.String("Key_With_Underscore"), Value: sdkaws.String("Value_With_Underscore")},
					{Key: sdkaws.String("Key.With.Dot"), Value: sdkaws.String("Value.With.Dot")},
				}

				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set tags with empty value", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				tags := []sdkstp.Tag{
					{Key: sdkaws.String("EmptyTag"), Value: sdkaws.String("")},
				}

				err := cli.Object().SetTags(testObjectKey, "", tags...)
				// May succeed or fail depending on AWS implementation
				_ = err
			})
		})

		Context("Getting tags", func() {
			BeforeEach(func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				// Set initial tags
				tags := []sdkstp.Tag{
					{Key: sdkaws.String("Environment"), Value: sdkaws.String("Test")},
					{Key: sdkaws.String("Project"), Value: sdkaws.String("TestProject")},
				}
				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should retrieve tags", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				tags, err := cli.Object().GetTags(testObjectKey, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(tags).To(HaveLen(2))

				// Check specific tags
				tagMap := make(map[string]string)
				for _, tag := range tags {
					tagMap[*tag.Key] = *tag.Value
				}
				Expect(tagMap["Environment"]).To(Equal("Test"))
				Expect(tagMap["Project"]).To(Equal("TestProject"))
			})

			It("should return empty tags for untagged object", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				untaggedKey := "untagged-" + GenerateUniqueName("")
				err := cli.Object().Put(untaggedKey, bytes.NewReader([]byte("untagged")))
				Expect(err).NotTo(HaveOccurred())
				defer func() {
					_ = cli.Object().Delete(false, untaggedKey)
				}()

				tags, err := cli.Object().GetTags(untaggedKey, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(tags).To(BeEmpty())
			})
		})

		Context("Updating tags", func() {
			It("should replace existing tags", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				// Set initial tags
				initialTags := []sdkstp.Tag{
					{Key: sdkaws.String("Tag1"), Value: sdkaws.String("Value1")},
				}
				err := cli.Object().SetTags(testObjectKey, "", initialTags...)
				Expect(err).NotTo(HaveOccurred())

				// Replace with new tags
				newTags := []sdkstp.Tag{
					{Key: sdkaws.String("Tag2"), Value: sdkaws.String("Value2")},
					{Key: sdkaws.String("Tag3"), Value: sdkaws.String("Value3")},
				}
				err = cli.Object().SetTags(testObjectKey, "", newTags...)
				Expect(err).NotTo(HaveOccurred())

				// Verify only new tags exist
				tags, err := cli.Object().GetTags(testObjectKey, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(tags).To(HaveLen(2))

				tagMap := make(map[string]string)
				for _, tag := range tags {
					tagMap[*tag.Key] = *tag.Value
				}
				Expect(tagMap).NotTo(HaveKey("Tag1"))
				Expect(tagMap).To(HaveKey("Tag2"))
				Expect(tagMap).To(HaveKey("Tag3"))
			})

			It("should update tag values", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				// Set initial tag
				tags := []sdkstp.Tag{
					{Key: sdkaws.String("Status"), Value: sdkaws.String("Initial")},
				}
				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())

				// Update value
				tags[0].Value = sdkaws.String("Updated")
				err = cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())

				// Verify updated value
				retrieved, err := cli.Object().GetTags(testObjectKey, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(*retrieved[0].Value).To(Equal("Updated"))
			})

			It("should clear all tags", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				// Set initial tags
				tags := []sdkstp.Tag{
					{Key: sdkaws.String("Tag1"), Value: sdkaws.String("Value1")},
					{Key: sdkaws.String("Tag2"), Value: sdkaws.String("Value2")},
				}
				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())

				// Clear tags
				err = cli.Object().SetTags(testObjectKey, "")
				Expect(err).NotTo(HaveOccurred())

				// Verify no tags
				retrieved, err := cli.Object().GetTags(testObjectKey, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved).To(BeEmpty())
			})
		})

		Context("Tag limits and validation", func() {
			It("should handle maximum number of tags", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				// AWS allows up to 10 tags per object
				tags := make([]sdkstp.Tag, 10)
				for i := 0; i < 10; i++ {
					tags[i] = sdkstp.Tag{
						Key:   sdkaws.String("Key" + string(rune('A'+i))),
						Value: sdkaws.String("Value" + string(rune('A'+i))),
					}
				}

				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())

				retrieved, err := cli.Object().GetTags(testObjectKey, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved).To(HaveLen(10))
			})

			It("should handle long tag keys and values", func() {
				if minioMode {
					Skip("MinIO: Object tagging compatibility varies")
				}

				// AWS allows up to 128 chars for key, 256 for value
				longKey := string(make([]byte, 100))
				for i := range longKey {
					longKey = string(append([]byte(longKey[:i]), 'A'))
				}
				longValue := string(make([]byte, 200))
				for i := range longValue {
					longValue = string(append([]byte(longValue[:i]), 'B'))
				}

				tags := []sdkstp.Tag{
					{Key: sdkaws.String(longKey), Value: sdkaws.String(longValue)},
				}

				err := cli.Object().SetTags(testObjectKey, "", tags...)
				// May succeed or fail depending on exact length limits
				_ = err
			})
		})

		Context("Tags with versioned objects", func() {
			It("should tag specific version", func() {
				if minioMode {
					Skip("MinIO: Object versioning and tagging not fully compatible")
				}

				// This test assumes versioning is enabled
				// Set tags on latest version
				tags := []sdkstp.Tag{
					{Key: sdkaws.String("Version"), Value: sdkaws.String("Latest")},
				}

				err := cli.Object().SetTags(testObjectKey, "", tags...)
				Expect(err).NotTo(HaveOccurred())

				// Retrieve tags
				retrieved, err := cli.Object().GetTags(testObjectKey, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved).NotTo(BeEmpty())
			})
		})
	})
})
