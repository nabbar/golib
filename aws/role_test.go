package aws_test

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	. "github.com/onsi/ginkgo"
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
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/_, err = cli.Role().Add(name, "{}")
			//			}
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
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/err = cli.Role().PolicyAttach("policyArn", "roleName")
			//			}
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
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/_, err = cli.Role().PolicyListAttached("invalidRoleName")
			//			}
			Expect(err).To(HaveOccurred())
		})
		It("Must return 1 policy", func() {
			var policies []iam.AttachedPolicy

			if minioMode {
				err = nil
				policies = []iam.AttachedPolicy{
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
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/err = cli.Role().PolicyDetach("policyArn", "roleName")
			//			}
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
			/*			if minioMode {
							err = nil
						} else {
			*/_, err = cli.Role().Check("invalid name")
			//			}

			Expect(err).To(HaveOccurred())
		})
	})
	Context("List", func() {
		It("Must return 1 role", func() {
			var roles []iam.Role

			if minioMode {
				err = nil
				roles = []iam.Role{
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
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/err = cli.Role().Delete(name)
			//			}
			Expect(err).To(HaveOccurred())
		})
	})
})
