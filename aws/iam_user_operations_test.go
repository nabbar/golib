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
	"github.com/nabbar/golib/password"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IAM User - Operations", func() {
	var (
		testUsername        = "test-user-ops"
		testPassword        = "TestP@ssw0rd123!"
		testAccessKey       string
		testGlobalAccessKey string
	)

	Describe("User creation and deletion", func() {
		Context("Creating a new user", func() {
			It("Create() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM User operations not fully compatible")
				}
				err := cli.User().Create(testUsername)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Create() should fail with duplicate username", func() {
				if minioMode {
					Skip("MinIO: IAM User operations not fully compatible")
				}
				err := cli.User().Create(testUsername)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Getting user information", func() {
			It("Get() should return user details", func() {
				if minioMode {
					Skip("MinIO: IAM User operations not fully compatible")
				}
				user, err := cli.User().Get(testUsername)
				Expect(err).NotTo(HaveOccurred())
				Expect(user).NotTo(BeNil())
				Expect(*user.UserName).To(Equal(testUsername))
			})

			It("Get() should fail for non-existent user", func() {
				if minioMode {
					Skip("MinIO: IAM User operations not fully compatible")
				}
				_, err := cli.User().Get("non-existent-user")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Listing users", func() {
			It("List() should return existing users", func() {
				if minioMode {
					Skip("MinIO: IAM User operations not fully compatible")
				}
				users, err := cli.User().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(users).NotTo(BeEmpty())
				Expect(users).To(HaveKey(testUsername))
			})
		})
	})

	Describe("User login credentials", func() {
		Context("Creating login profile", func() {
			It("LoginCreate() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM login operations not fully compatible")
				}
				err := cli.User().LoginCreate(testUsername, testPassword)
				Expect(err).NotTo(HaveOccurred())
			})

			It("LoginCreate() should fail with duplicate", func() {
				if minioMode {
					Skip("MinIO: IAM login operations not fully compatible")
				}
				err := cli.User().LoginCreate(testUsername, testPassword)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Checking login profile", func() {
			It("LoginCheck() should succeed for user with login", func() {
				if minioMode {
					Skip("MinIO: IAM login operations not fully compatible")
				}
				err := cli.User().LoginCheck(testUsername)
				Expect(err).NotTo(HaveOccurred())
			})

			It("LoginCheck() should fail for invalid username", func() {
				if minioMode {
					Skip("MinIO: IAM login operations not fully compatible")
				}
				err := cli.User().LoginCheck("non-existent-user")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Deleting login profile", func() {
			It("LoginDelete() should succeed", func() {
				if minioMode {
					Skip("MinIO: IAM login operations not fully compatible")
				}
				err := cli.User().LoginDelete(testUsername)
				Expect(err).NotTo(HaveOccurred())
			})

			It("LoginDelete() should fail for already deleted", func() {
				if minioMode {
					Skip("MinIO: IAM login operations not fully compatible")
				}
				err := cli.User().LoginDelete(testUsername)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("User access keys", func() {
		Context("Creating access keys", func() {
			It("AccessCreate() should succeed for specific user", func() {
				if minioMode {
					testAccessKey = password.Generate(20)
					return
				}
				var err error
				testAccessKey, _, err = cli.User().AccessCreate(testUsername)
				Expect(err).NotTo(HaveOccurred())
				Expect(testAccessKey).NotTo(BeEmpty())
			})

			It("AccessCreate() should succeed for default user", func() {
				if minioMode {
					testGlobalAccessKey = password.Generate(20)
					return
				}
				var err error
				testGlobalAccessKey, _, err = cli.User().AccessCreate("")
				Expect(err).NotTo(HaveOccurred())
				Expect(testGlobalAccessKey).NotTo(BeEmpty())
			})
		})

		Context("Listing access keys", func() {
			It("AccessList() should return user's access keys", func() {
				if minioMode {
					Skip("MinIO: IAM access key operations not fully compatible")
				}
				accessKeys, err := cli.User().AccessList(testUsername)
				Expect(err).NotTo(HaveOccurred())
				Expect(accessKeys).NotTo(BeEmpty())
				Expect(accessKeys).To(HaveKey(testAccessKey))
			})

			It("AccessList() should return global access keys", func() {
				if minioMode {
					Skip("MinIO: IAM access key operations not fully compatible")
				}
				accessKeys, err := cli.User().AccessList("")
				Expect(err).NotTo(HaveOccurred())
				Expect(accessKeys).To(HaveKey(testGlobalAccessKey))
			})

			It("AccessList() should fail for invalid username", func() {
				if minioMode {
					Skip("MinIO: IAM access key operations not fully compatible")
				}
				_, err := cli.User().AccessList("non-existent-user")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Deleting access keys", func() {
			It("AccessDelete() should fail with invalid username", func() {
				if minioMode {
					Skip("MinIO: IAM access key operations not fully compatible")
				}
				err := cli.User().AccessDelete("non-existent-user", testAccessKey)
				Expect(err).To(HaveOccurred())
			})

			It("AccessDelete() should succeed for user access key", func() {
				if minioMode {
					Skip("MinIO: IAM access key operations not fully compatible")
				}
				err := cli.User().AccessDelete(testUsername, testAccessKey)
				Expect(err).NotTo(HaveOccurred())
			})

			It("AccessDelete() should succeed for global access key", func() {
				if minioMode {
					Skip("MinIO: IAM access key operations not fully compatible")
				}
				err := cli.User().AccessDelete("", testGlobalAccessKey)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("User policies", func() {
		Context("Attaching policies", func() {
			It("PolicyPut() should fail with empty policy", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.User().PolicyPut("", "test-policy", testUsername)
				Expect(err).To(HaveOccurred())
			})

			It("PolicyAttach() should fail with invalid ARN", func() {
				if minioMode {
					Skip("MinIO: IAM policy operations not fully compatible")
				}
				err := cli.User().PolicyAttach("invalid-arn", testUsername)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("User deletion", func() {
		It("Delete() should succeed", func() {
			if minioMode {
				Skip("MinIO: IAM User operations not fully compatible")
			}
			err := cli.User().Delete(testUsername)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete() should fail for already deleted user", func() {
			if minioMode {
				Skip("MinIO: IAM User operations not fully compatible")
			}
			err := cli.User().Delete(testUsername)
			Expect(err).To(HaveOccurred())
		})
	})
})
