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

var _ = Describe("IAM Role - Operations", func() {
	var (
		testRoleName   = "test-role-ops"
		testPolicyName = "test-policy-for-role"
		testRoleArn    string
		testPolicyArn  string
	)

	Describe("Role creation and validation", func() {
		Context("Creating a new role", func() {
			It("Add() should fail with invalid JSON", func() {
				if minioMode {
					Skip("MinIO: IAM Role operations not fully compatible")
				}
				_, err := cli.Role().Add(testRoleName, "{}")
				Expect(err).To(HaveOccurred())
			})

			It("Add() should fail with empty assume policy", func() {
				if minioMode {
					Skip("MinIO: IAM Role operations not fully compatible")
				}
				_, err := cli.Role().Add(testRoleName, "")
				Expect(err).To(HaveOccurred())
			})

			It("Add() should succeed with valid assume policy", func() {
				if minioMode {
					Skip("MinIO: IAM Role operations not fully compatible")
				}
				var err error
				testRoleArn, err = cli.Role().Add(testRoleName, BuildRole())
				Expect(err).NotTo(HaveOccurred())
				Expect(testRoleArn).NotTo(BeEmpty())
			})

			It("Add() should fail with duplicate role name", func() {
				if minioMode {
					Skip("MinIO: IAM Role operations not fully compatible")
				}
				_, err := cli.Role().Add(testRoleName, BuildRole())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Checking role existence", func() {
			It("Check() should return role ARN", func() {
				if minioMode {
					Skip("MinIO: IAM Role operations not fully compatible")
				}
				roleArn, err := cli.Role().Check(testRoleName)
				Expect(err).NotTo(HaveOccurred())
				Expect(roleArn).To(Equal(testRoleArn))
			})

			It("Check() should fail for non-existent role", func() {
				if minioMode {
					Skip("MinIO: IAM Role operations not fully compatible")
				}
				_, err := cli.Role().Check("non-existent-role")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Listing roles", func() {
			It("List() should return existing roles", func() {
				if minioMode {
					Skip("MinIO: IAM Role operations not fully compatible")
				}
				roles, err := cli.Role().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(roles).NotTo(BeEmpty())

				// Find our test role
				found := false
				for _, role := range roles {
					if role.RoleName != nil && *role.RoleName == testRoleName {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue())
			})
		})
	})

	Describe("Role policy management", func() {
		BeforeEach(func() {
			if minioMode {
				Skip("MinIO: IAM policy operations not fully compatible")
			}
			// Create test policy
			var err error
			testPolicyArn, err = cli.Policy().Add(testPolicyName, "Test policy for role", BuildPolicy())
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if !minioMode {
				// Cleanup: Detach policy if exists
				_ = cli.Role().PolicyDetach(testPolicyArn, testRoleName)
				// Delete test policy
				_ = cli.Policy().Delete(testPolicyArn)
			}
		})

		Context("Attaching policies to role", func() {
			It("PolicyAttach() should fail with invalid parameters", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Role().PolicyAttach("invalid-policy-arn", "invalid-role-name")
				Expect(err).To(HaveOccurred())
			})

			It("PolicyAttach() should succeed with valid parameters", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Role().PolicyAttach(testPolicyArn, testRoleName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("PolicyAttach() should handle duplicate attachment", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				// Attach first time
				err := cli.Role().PolicyAttach(testPolicyArn, testRoleName)
				Expect(err).NotTo(HaveOccurred())

				// Attach again - may succeed or fail depending on AWS behavior
				_ = cli.Role().PolicyAttach(testPolicyArn, testRoleName)
			})
		})

		Context("Listing attached policies", func() {
			BeforeEach(func() {
				if !minioMode {
					err := cli.Role().PolicyAttach(testPolicyArn, testRoleName)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("PolicyListAttached() should return attached policies", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				policies, err := cli.Role().PolicyListAttached(testRoleName)
				Expect(err).NotTo(HaveOccurred())
				Expect(policies).To(HaveLen(1))
				Expect(*policies[0].PolicyArn).To(Equal(testPolicyArn))
			})

			It("PolicyListAttached() should fail with invalid role name", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				_, err := cli.Role().PolicyListAttached("invalid-role-name")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Detaching policies from role", func() {
			BeforeEach(func() {
				if !minioMode {
					err := cli.Role().PolicyAttach(testPolicyArn, testRoleName)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("PolicyDetach() should fail with invalid parameters", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Role().PolicyDetach("invalid-policy-arn", "invalid-role-name")
				Expect(err).To(HaveOccurred())
			})

			It("PolicyDetach() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Role().PolicyDetach(testPolicyArn, testRoleName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("PolicyDetach() should fail for already detached policy", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				// Detach once
				err := cli.Role().PolicyDetach(testPolicyArn, testRoleName)
				Expect(err).NotTo(HaveOccurred())

				// Detach again should fail
				err = cli.Role().PolicyDetach(testPolicyArn, testRoleName)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Role deletion", func() {
		It("Delete() should succeed", func() {
			if minioMode {
				Skip("MinIO: IAM Role operations not fully compatible")
			}
			err := cli.Role().Delete(testRoleName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete() should fail for already deleted role", func() {
			if minioMode {
				Skip("MinIO: IAM Role operations not fully compatible")
			}
			err := cli.Role().Delete(testRoleName)
			Expect(err).To(HaveOccurred())
		})

		It("Delete() should fail for non-existent role", func() {
			if minioMode {
				Skip("MinIO: IAM Role operations not fully compatible")
			}
			err := cli.Role().Delete("non-existent-role")
			Expect(err).To(HaveOccurred())
		})
	})
})
