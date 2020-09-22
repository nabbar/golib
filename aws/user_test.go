package aws_test

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/password"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User", func() {
	var (
		username        string = "myUsername"
		userpass        string = "myPassword"
		accessKey       string
		globalAccessKey string
		policyName      string = "myPolicy"
		err             error
	)

	Context("Create User", func() {
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.User().Create(username)
			}

			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail with already existing user", func() {
			Expect(cli.User().Create(username)).To(HaveOccurred())
		})
	})
	Context("Get", func() {
		It("Must succeed", func() {
			var user *iam.User

			if minioMode {
				err = nil
				user = &iam.User{
					UserName: aws.String(username),
				}
			} else {
				user, err = cli.User().Get(username)
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(*user.UserName).To(Equal(username))
		})
		It("Must fail with invalid username", func() {
			_, err := cli.User().Get("username")
			Expect(err).To(HaveOccurred())
		})
	})
	Context("List", func() {
		It("Must succeed", func() {
			var users map[string]string

			if minioMode {
				err = nil
				users = map[string]string{
					username: username,
				}
			} else {
				users, err = cli.User().List()
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(users).To(HaveLen(1))
		})
	})

	Context("Create Login", func() {
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.User().LoginCreate(username, userpass)
			}

			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail with already existing user", func() {
			Expect(cli.User().LoginCreate(username, userpass)).To(HaveOccurred())
		})
	})
	Context("Check Login", func() {
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.User().LoginCheck(username)
			}

			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail with invalid username", func() {
			Expect(cli.User().LoginCheck("username")).To(HaveOccurred())
		})
	})
	Context("Delete Login", func() {
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.User().LoginDelete(username)
			}

			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail (already deleted)", func() {
			Expect(cli.User().LoginDelete(username)).To(HaveOccurred())
		})
	})

	Context("Create Access", func() {
		It("Must succeed with username", func() {
			if minioMode {
				err = nil
				accessKey = password.Generate(20)
			} else {
				accessKey, _, err = cli.User().AccessCreate(username)
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(accessKey).ToNot(Equal(""))
		})
		It("Must succeed without username", func() {
			if minioMode {
				err = nil
				globalAccessKey = password.Generate(20)
			} else {
				globalAccessKey, _, err = cli.User().AccessCreate("")
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(globalAccessKey).ToNot(Equal(""))
		})
	})
	Context("List Access", func() {
		It("With username must return the accessKey", func() {
			var access map[string]bool

			if minioMode {
				err = nil
				access = map[string]bool{
					accessKey: true,
				}
			} else {
				access, err = cli.User().AccessList(username)
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(access).To(HaveLen(1))
		})
		It("Must return global account's access keys", func() {
			var access map[string]bool

			if minioMode {
				err = nil
				access = map[string]bool{
					globalAccessKey: true,
				}
			} else {
				access, err = cli.User().AccessList("")
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(access).To(HaveKeyWithValue(globalAccessKey, true))
		})
		It("Must fail with invalid username", func() {
			_, err = cli.User().AccessList("username")
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Delete Access", func() {
		It("Must fail with invalid username", func() {
			Expect(cli.User().AccessDelete("username", accessKey)).To(HaveOccurred())
		})
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.User().AccessDelete(username, accessKey)
			}

			Expect(err).ToNot(HaveOccurred())
		})
		It("Must succeed for the global accessKey", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.User().AccessDelete("", globalAccessKey)
			}

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Put policy", func() {
		It("Must fail with empty policy", func() {
			Expect(cli.User().PolicyPut("", policyName, username)).To(HaveOccurred())
		})
	})
	Context("Attach policy", func() {
		It("Must fail with fake policy ARN", func() {
			Expect(cli.User().PolicyAttach("fake arn", username)).To(HaveOccurred())
		})
	})

	Context("Delete User", func() {
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.User().Delete(username)
			}

			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail (already deleted)", func() {
			Expect(cli.User().Delete(username)).To(HaveOccurred())
		})
	})
})
