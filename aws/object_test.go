package aws_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Object", func() {
	Context("List objects", func() {
		It("Must fail with invalid token", func() {
			_, _, _, err := cli.Object().List("token")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Put object", func() {
		It("Must fail as the bucket doesn't exists", func() {
			err := cli.Object().Put("object", bytes.NewReader([]byte("")))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Get object", func() {
		It("Must fail as the bucket doesn't exists", func() {
			_, c, err := cli.Object().Get("object")

			defer func() {
				for _, s := range c {
					if s != nil {
						_ = s.Close()
					}
				}
			}()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("Delete object", func() {
		It("Must fail as the object doesn't exists", func() {
			err := cli.Object().Delete("object")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Multipart Put object", func() {
		It("Must fail as the bucket doesn't exists", func() {
			err := cli.Object().MultipartPut("object", bytes.NewReader([]byte("")))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Delete object", func() {
		It("Must fail as the object doesn't exists", func() {
			err := cli.Object().Delete("object")
			Expect(err).To(HaveOccurred())
		})
	})

})
