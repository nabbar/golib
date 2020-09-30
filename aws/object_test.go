package aws_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Object", func() {
	Context("List objects", func() {
		It("Must fail with invalid token -1 ", func() {
			_, _, _, err := cli.Object().List("token")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Put object", func() {
		It("Must fail as the bucket doesn't exists - 2", func() {
			err := cli.Object().Put("object", bytes.NewReader([]byte("")))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Get object", func() {
		It("Must fail as the bucket doesn't exists - 3", func() {
			o, err := cli.Object().Get("object")

			defer func() {
				if o != nil && o.Body != nil {
					_ = o.Body.Close()
				}
			}()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("Delete object", func() {
		It("Must fail as the object doesn't exists - 4", func() {
			err := cli.Object().Delete("object")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Multipart Put object", func() {
		It("Must fail as the bucket doesn't exists - 5", func() {
			err := cli.Object().MultipartPut("object", randContent(4*1024))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Delete object", func() {
		It("Must fail as the object doesn't exists - 6", func() {
			err := cli.Object().Delete("object")
			Expect(err).To(HaveOccurred())
		})
	})

})
