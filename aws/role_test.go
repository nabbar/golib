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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Role", func() {
	var (
		arn       string
		policyArn string
		name      string = "role"
		err       error
	)

	Context("Creation", func() {
		It("Must fail with invalid json", func() {
			_, err = cli.Role().Add(name, "{}")
			Expect(err).To(HaveOccurred())
		})
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				arn, err = cli.Role().Add(name, BuildRole())
			}
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("Attach", func() {
		It("Must fail with invalid params", func() {
			err = cli.Role().PolicyAttach("policyArn", "roleName")
			Expect(err).To(HaveOccurred())
		})
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				policyArn, err = cli.Policy().Add("tmp", "tmp", BuildPolicy())
			}
			Expect(err).ToNot(HaveOccurred())

			if minioMode {
				err = nil
			} else {
				err = cli.Role().PolicyAttach(policyArn, name)
			}
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("List attached policies to role", func() {
		It("Must fail with invalid role name", func() {
			_, err = cli.Role().PolicyListAttached("invalidRoleName")
			Expect(err).To(HaveOccurred())
		})
		It("Must return 1 policy", func() {
			var policies []types.AttachedPolicy

			if minioMode {
				err = nil
				policies = []types.AttachedPolicy{
					{
						PolicyArn:  aws.String(policyArn),
						PolicyName: aws.String(name),
					},
				}
			} else {
				policies, err = cli.Role().PolicyListAttached(name)
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(policies).To(HaveLen(1))
		})
	})
	Context("Detach", func() {
		It("Must fail with invalid params", func() {
			err = cli.Role().PolicyDetach("policyArn", "roleName")
			Expect(err).To(HaveOccurred())
		})
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.Role().PolicyDetach(policyArn, name)
			}
			Expect(err).ToNot(HaveOccurred())

			if minioMode {
				err = nil
			} else {
				err = cli.Policy().Delete(policyArn)
			}
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("Check", func() {
		It("Must return role arn", func() {
			var roleArn string

			if minioMode {
				err = nil
				roleArn = arn
			} else {
				roleArn, err = cli.Role().Check(name)
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(roleArn).To(Equal(arn))
		})
		It("Must fail with invalid name", func() {
			_, err = cli.Role().Check("invalid name")
			Expect(err).To(HaveOccurred())
		})
	})
	Context("List", func() {
		It("Must return 1 role", func() {
			var roles []types.Role

			if minioMode {
				err = nil
				roles = []types.Role{
					{
						Arn:      aws.String(arn),
						RoleName: aws.String(name),
					},
				}
			} else {
				roles, err = cli.Role().List()
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(roles).To(HaveLen(1))
		})
	})
	Context("Delete", func() {
		It("Must be possible to delete a role", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.Role().Delete(name)
			}
			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail", func() {
			err = cli.Role().Delete(name)
			Expect(err).To(HaveOccurred())
		})
	})
})
