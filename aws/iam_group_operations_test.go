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

var _ = Describe("IAM Group - Operations", func() {
	var (
		testGroupName  = "test-group-ops"
		testUserName   = "test-user-for-group"
		testPolicyName = "test-policy-for-group"
		testPolicyArn  string
	)

	Describe("Group creation and listing", func() {
		Context("Creating a new group", func() {
			It("Add() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM Group operations not fully compatible")
				}
				err := cli.Group().Add(testGroupName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Add() should fail with duplicate name", func() {
				if minioMode {
					Skip("MinIO: IAM Group operations not fully compatible")
				}
				err := cli.Group().Add(testGroupName)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Listing groups", func() {
			It("List() should return existing groups", func() {
				if minioMode {
					Skip("MinIO: IAM Group operations not fully compatible")
				}
				groups, err := cli.Group().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).NotTo(BeEmpty())
				Expect(groups).To(HaveKey(testGroupName))
			})
		})
	})

	Describe("Group user management", func() {
		BeforeEach(func() {
			if minioMode {
				Skip("MinIO: IAM Group user operations not fully compatible")
			}
			// Create test user
			err := cli.User().Create(testUserName)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if !minioMode {
				// Cleanup: Remove user from group if exists
				_ = cli.Group().UserRemove(testUserName, testGroupName)
				// Delete test user
				_ = cli.User().Delete(testUserName)
			}
		})

		Context("Adding users to group", func() {
			It("UserAdd() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				err := cli.Group().UserAdd(testUserName, testGroupName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("UserAdd() should handle duplicate gracefully", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				// Add user first time
				err := cli.Group().UserAdd(testUserName, testGroupName)
				Expect(err).NotTo(HaveOccurred())

				// Add same user again - may succeed or fail depending on AWS behavior
				_ = cli.Group().UserAdd(testUserName, testGroupName)
			})
		})

		Context("Checking user membership", func() {
			BeforeEach(func() {
				if !minioMode {
					err := cli.Group().UserAdd(testUserName, testGroupName)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("UserCheck() should return true for member", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				err, ok := cli.Group().UserCheck(testUserName, testGroupName)
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeTrue())
			})

			It("UserCheck() should fail with invalid parameters", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				err, ok := cli.Group().UserCheck("invalid-user", "invalid-group")
				Expect(err).To(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("Listing users in group", func() {
			BeforeEach(func() {
				if !minioMode {
					err := cli.Group().UserAdd(testUserName, testGroupName)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("UserList() should return group members", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				groups, err := cli.Group().UserList(testUserName)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(ContainElement(testGroupName))
			})

			It("UserList() should fail with invalid user", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				_, err := cli.Group().UserList("invalid-user")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Removing users from group", func() {
			BeforeEach(func() {
				if !minioMode {
					err := cli.Group().UserAdd(testUserName, testGroupName)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("UserRemove() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				err := cli.Group().UserRemove(testUserName, testGroupName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("UserRemove() should fail for non-member", func() {
				if minioMode {
					Skip("MinIO: IAM Group user operations not fully compatible")
				}
				// Remove once
				err := cli.Group().UserRemove(testUserName, testGroupName)
				Expect(err).NotTo(HaveOccurred())

				// Remove again should fail
				err = cli.Group().UserRemove(testUserName, testGroupName)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Group policy management", func() {
		BeforeEach(func() {
			if minioMode {
				Skip("MinIO: IAM policy operations not fully compatible")
			}
			// Create test policy
			var err error
			testPolicyArn, err = cli.Policy().Add(testPolicyName, "Test policy", BuildPolicy())
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if !minioMode {
				// Cleanup: Detach policy if exists
				_ = cli.Group().PolicyDetach(testGroupName, testPolicyArn)
				// Delete test policy
				_ = cli.Policy().Delete(testPolicyArn)
			}
		})

		Context("Attaching policies to group", func() {
			It("PolicyAttach() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Group().PolicyAttach(testGroupName, testPolicyArn)
				Expect(err).NotTo(HaveOccurred())
			})

			It("PolicyAttach() should handle duplicate gracefully", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				// Attach first time
				err := cli.Group().PolicyAttach(testGroupName, testPolicyArn)
				Expect(err).NotTo(HaveOccurred())

				// Attach again - may succeed or fail
				_ = cli.Group().PolicyAttach(testGroupName, testPolicyArn)
			})
		})

		Context("Listing policies in group", func() {
			BeforeEach(func() {
				if !minioMode {
					err := cli.Group().PolicyAttach(testGroupName, testPolicyArn)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("PolicyList() should return attached policies", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				policies, err := cli.Group().PolicyList(testGroupName)
				Expect(err).NotTo(HaveOccurred())
				Expect(policies).To(HaveKeyWithValue(testPolicyName, testPolicyArn))
			})

			It("PolicyList() should fail with invalid group", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				_, err := cli.Group().PolicyList("invalid-group")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Detaching policies from group", func() {
			BeforeEach(func() {
				if !minioMode {
					err := cli.Group().PolicyAttach(testGroupName, testPolicyArn)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("PolicyDetach() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.Group().PolicyDetach(testGroupName, testPolicyArn)
				Expect(err).NotTo(HaveOccurred())
			})

			It("PolicyDetach() should fail for already detached", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				// Detach once
				err := cli.Group().PolicyDetach(testGroupName, testPolicyArn)
				Expect(err).NotTo(HaveOccurred())

				// Detach again should fail
				err = cli.Group().PolicyDetach(testGroupName, testPolicyArn)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Group deletion", func() {
		It("Remove() should succeed", func() {
			if minioMode {
				Skip("MinIO: IAM Group operations not fully compatible")
			}
			err := cli.Group().Remove(testGroupName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Remove() should fail for already deleted group", func() {
			if minioMode {
				Skip("MinIO: IAM Group operations not fully compatible")
			}
			err := cli.Group().Remove(testGroupName)
			Expect(err).To(HaveOccurred())
		})
	})
})
