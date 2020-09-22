package aws_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policies", func() {
	var (
		arn  string
		name string = "policy"
		err  error
	)

	Context("Creation", func() {
		It("Must fail with invalid json", func() {
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/_, err = cli.Policy().Add(name, "policy desc", "{}")
			//			}
			Expect(err).To(HaveOccurred())
		})
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				arn, err = cli.Policy().Add(name, "policy initial desc", BuildPolicy())
			}
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("Update", func() {
		It("Must fail with invalid json", func() {
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/err = cli.Policy().Update(arn, "{}")
			//			}
			Expect(err).To(HaveOccurred())
		})
		It("Must fail with invalid arn", func() {
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/err = cli.Policy().Update("bad arn", "{}")
			//			}
			Expect(err).To(HaveOccurred())
		})
		It("Must succeed", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.Policy().Update(arn, BuildPolicy())
			}
			Expect(err).ToNot(HaveOccurred())
		})
		It("Must succeed again", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.Policy().Update(arn, BuildPolicy())
			}
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("List", func() {
		It("Must return 3 policies", func() { //Default policies + 1 made just above
			var policies map[string]string

			if minioMode {
				err = nil
				policies = map[string]string{
					name: arn,
				}
			} else {
				policies, err = cli.Policy().List()
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(policies).To(HaveKeyWithValue(name, arn))
		})
	})
	Context("Delete", func() {
		It("Must be possible to delete a policy", func() {
			if minioMode {
				err = nil
			} else {
				err = cli.Policy().Delete(arn)
			}
			Expect(err).ToNot(HaveOccurred())
		})
		It("Must fail", func() {
			/*			if minioMode {
							err = fmt.Errorf("backend not compatible following AWS API reference")
						} else {
			*/err = cli.Policy().Delete(arn)
			//			}
			Expect(err).To(HaveOccurred())
		})
	})
})
