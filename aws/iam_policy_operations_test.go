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

var _ = Describe("IAM Policy - Operations", func() {
	var (
		testPolicyName = "test-policy-ops"
		testPolicyArn  string
	)

	Describe("Policy creation and validation", func() {
		Context("Creating a new policy", func() {
			It("Add() should fail with invalid JSON", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				_, err := cli.Policy().Add(testPolicyName, "Test policy description", "{}")
				Expect(err).To(HaveOccurred())
			})

			It("Add() should fail with empty policy document", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				_, err := cli.Policy().Add(testPolicyName, "Test policy description", "")
				Expect(err).To(HaveOccurred())
			})

			It("Add() should succeed with valid policy document", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				var err error
				testPolicyArn, err = cli.Policy().Add(testPolicyName, "Test policy initial description", BuildPolicy())
				Expect(err).NotTo(HaveOccurred())
				Expect(testPolicyArn).NotTo(BeEmpty())
			})

			It("Add() should fail with duplicate policy name", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				_, err := cli.Policy().Add(testPolicyName, "Duplicate description", BuildPolicy())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Listing policies", func() {
			It("List() should return existing policies", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				policies, err := cli.Policy().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(policies).NotTo(BeEmpty())
				Expect(policies).To(HaveKeyWithValue(testPolicyName, testPolicyArn))
			})

			It("List() should return at least the test policy", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				policies, err := cli.Policy().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(policies).To(HaveKey(testPolicyName))
			})
		})
	})

	Describe("Policy updates", func() {
		Context("Updating policy document", func() {
			It("Update() should fail with invalid JSON", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Update(testPolicyArn, "{}")
				Expect(err).To(HaveOccurred())
			})

			It("Update() should fail with invalid ARN", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Update("invalid-arn", BuildPolicy())
				Expect(err).To(HaveOccurred())
			})

			It("Update() should succeed with valid policy", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Update(testPolicyArn, BuildPolicy())
				Expect(err).NotTo(HaveOccurred())
			})

			It("Update() should succeed multiple times", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				// First update
				err := cli.Policy().Update(testPolicyArn, BuildPolicy())
				Expect(err).NotTo(HaveOccurred())

				// Second update
				err = cli.Policy().Update(testPolicyArn, BuildPolicy())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Policy versioning", func() {
			It("Update() should create new version", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				// AWS maintains policy versions, update should succeed
				err := cli.Policy().Update(testPolicyArn, BuildPolicy())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Policy document validation", func() {
		Context("Valid policy documents", func() {
			It("should accept policy with multiple statements", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				multiStatement := `{
					"Version": "2012-10-17",
					"Statement": [
						{
							"Effect": "Allow",
							"Action": ["s3:Get*"],
							"Resource": ["arn:aws:s3:::*/*"]
						},
						{
							"Effect": "Allow",
							"Action": ["s3:List*"],
							"Resource": ["arn:aws:s3:::*"]
						}
					]
				}`
				err := cli.Policy().Update(testPolicyArn, multiStatement)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept policy with conditions", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				conditionalPolicy := `{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Action": ["s3:Get*"],
						"Resource": ["arn:aws:s3:::*/*"],
						"Condition": {
							"StringEquals": {
								"s3:prefix": ["home/"]
							}
						}
					}]
				}`
				err := cli.Policy().Update(testPolicyArn, conditionalPolicy)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Invalid policy documents", func() {
			It("should reject malformed JSON", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Update(testPolicyArn, "{invalid json")
				Expect(err).To(HaveOccurred())
			})

			It("should reject policy without Version", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				noVersion := `{"Statement":[{"Effect":"Allow","Action":["s3:Get*"],"Resource":["*"]}]}`
				err := cli.Policy().Update(testPolicyArn, noVersion)
				Expect(err).To(HaveOccurred())
			})

			It("should reject policy without Statement", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				noStatement := `{"Version":"2012-10-17"}`
				err := cli.Policy().Update(testPolicyArn, noStatement)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Policy deletion", func() {
		Context("Deleting existing policy", func() {
			It("Delete() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Delete(testPolicyArn)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Delete() should fail for already deleted policy", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Delete(testPolicyArn)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Deleting with invalid parameters", func() {
			It("Delete() should fail with invalid ARN", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Delete("invalid-arn")
				Expect(err).To(HaveOccurred())
			})

			It("Delete() should fail with empty ARN", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Policy().Delete("")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Policy attachment lifecycle", func() {
		var (
			tempPolicyArn string
			tempRoleName  = "temp-role-for-policy-test"
		)

		BeforeEach(func() {
			if minioMode {
				Skip("MinIO: IAM policy operations not fully compatible")
			}
			var err error
			// Create temporary policy
			tempPolicyArn, err = cli.Policy().Add("temp-policy", "Temporary policy", BuildPolicy())
			Expect(err).NotTo(HaveOccurred())

			// Create temporary role
			_, err = cli.Role().Add(tempRoleName, BuildRole())
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if !minioMode {
				// Cleanup
				_ = cli.Role().PolicyDetach(tempPolicyArn, tempRoleName)
				_ = cli.Role().Delete(tempRoleName)
				_ = cli.Policy().Delete(tempPolicyArn)
			}
		})

		It("should be attachable to roles", func() {
			if minioMode {
				Skip("MinIO: IAM policy operations not fully compatible")
			}
			err := cli.Role().PolicyAttach(tempPolicyArn, tempRoleName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should prevent deletion when attached", func() {
			if minioMode {
				Skip("MinIO: IAM policy operations not fully compatible")
			}
			// Attach policy
			err := cli.Role().PolicyAttach(tempPolicyArn, tempRoleName)
			Expect(err).NotTo(HaveOccurred())

			// Try to delete while attached - should fail
			err = cli.Policy().Delete(tempPolicyArn)
			Expect(err).To(HaveOccurred())

			// Detach and then delete should work
			err = cli.Role().PolicyDetach(tempPolicyArn, tempRoleName)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
