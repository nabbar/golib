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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Groups", func() {
	var (
		groupName  = "myGroup"
		userName   = "myUsername"
		policyName = "myPolicy"
		err        error
	)

	Context("Create Group", func() {
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.Group().Add(groupName)
			}
			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail with already existing user", func() {
			if minioMode {
				//nolint #goerr113
				err = fmt.Errorf("backend not compatible following AWS API reference")
			} else {
				err = cli.Group().Add(groupName)
			}
			Expect(err).To(HaveOccurred())
		})
	})

	Context("List", func() {
		It("Must succeed", func() {
			var group map[string]string

			if minioMode {
				err = nil
				group = map[string]string{
					"skip1": "skip",
					"skip2": "skip",
					"skip3": "skip",
				}
			} else {
				group, err = cli.Group().List()
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(group).To(HaveLen(3))
		})
	})

	Context("User Operations", func() {
		Context("Add user to group", func() {
			It("Must succeed", func() {
				if minioMode {
					err = nil
				} else {
					err = cli.User().Create(userName)
				}
				Expect(err).ToNot(HaveOccurred())

				if minioMode {
					err = nil
				} else {
					err = cli.Group().UserAdd(userName, groupName)
				}
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("Check if user is in group", func() {
			It("Must succeed", func() {
				var ok bool
				if minioMode {
					err = nil
					ok = true
				} else {
					err, ok = cli.Group().UserCheck(userName, groupName)
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(Equal(true))
			})
			It("Must fail with invalid params", func() {
				var ok bool
				if minioMode {
					//nolint #goerr113
					err = fmt.Errorf("backend not compatible following AWS API reference")
					ok = false
				} else {
					err, ok = cli.Group().UserCheck("userName", "groupName")
				}
				Expect(err).To(HaveOccurred())
				Expect(ok).To(Equal(false))
			})
		})
		Context("List users in group", func() {
			It("Must succeed", func() {
				var group []string

				if minioMode {
					err = nil
					group = []string{groupName}
				} else {
					group, err = cli.Group().UserList(userName)
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(group).To(ContainElements(groupName))
			})
			It("Must fail with invalid groupName", func() {
				if minioMode {
					//nolint #goerr113
					err = fmt.Errorf("backend not compatible following AWS API reference")
				} else {
					_, err = cli.Group().UserList("groupName")
				}
				_, err := cli.Group().UserList("groupName")
				Expect(err).To(HaveOccurred())
			})
		})
		Context("Remove user from group", func() {
			It("Must succeed", func() {
				if minioMode {
					err = nil
				} else {
					err = cli.Group().UserRemove(userName, groupName)
				}
				Expect(err).ToNot(HaveOccurred())

				if minioMode {
					err = nil
				} else {
					err = cli.User().Delete(userName)
				}
				Expect(err).ToNot(HaveOccurred())
			})
			It("Must fail (already deleted)", func() {
				if minioMode {
					//nolint #goerr113
					err = fmt.Errorf("backend not compatible following AWS API reference")
				} else {
					err = cli.Group().UserRemove(userName, groupName)
				}
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Policy Operations", func() {

		var policyArn string

		Context("Attach policy to group", func() {
			It("Must succeed", func() {
				if minioMode {
					err = nil
				} else {
					policyArn, err = cli.Policy().Add(policyName, "description", BuildPolicy())
				}
				Expect(err).ToNot(HaveOccurred())

				if minioMode {
					err = nil
				} else {
					err = cli.Group().PolicyAttach(groupName, policyArn)
				}
				Expect(err).ToNot(HaveOccurred())

			})
		})
		Context("List policies in group", func() {
			It("Must succeed", func() {
				var policies map[string]string

				if minioMode {
					err = nil
					policies = map[string]string{
						policyName: policyArn,
					}
				} else {
					policies, err = cli.Group().PolicyList(groupName)
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(policies).To(HaveKeyWithValue(policyName, policyArn))
			})
			It("Must fail with invalid groupName", func() {
				if minioMode {
					//nolint #goerr113
					err = fmt.Errorf("backend not compatible following AWS API reference")
				} else {
					_, err = cli.Group().PolicyList("groupName")
				}
				Expect(err).To(HaveOccurred())
			})
		})
		Context("Remove policy from group", func() {
			It("Must succeed", func() {
				if minioMode {
					err = nil
				} else {
					err = cli.Group().PolicyDetach(groupName, policyArn)
				}
				Expect(err).ToNot(HaveOccurred())

				if minioMode {
					err = nil
				} else {
					err = cli.Policy().Delete(policyArn)
				}
				Expect(err).ToNot(HaveOccurred())
			})
			It("Must fail (already deleted)", func() {
				if minioMode {
					//nolint #goerr113
					err = fmt.Errorf("backend not compatible following AWS API reference")
				} else {
					err = cli.Group().PolicyDetach(groupName, policyArn)
				}
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Delete Group", func() {
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.Group().Remove(groupName)
			}
			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail (already deleted)", func() {
			if minioMode {
				//nolint #goerr113
				err = fmt.Errorf("backend not compatible following AWS API reference")
			} else {
				err = cli.Group().Remove(groupName)
			}
			Expect(err).To(HaveOccurred())
		})
	})
})
