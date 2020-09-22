package aws_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Context("Creation", func() {
		It("Must be possible to check if the bucket exists", func() {
			Expect(cli.Bucket().Check()).ToNot(Succeed())
		})
		It("Must be possible to create a bucket", func() {
			Expect(cli.Bucket().Create()).To(Succeed())
			Expect(cli.Bucket().Check()).To(Succeed())
		})
	})
	Context("Find object", func() {
		Context("With no object in bucket", func() {
			It("Must succeed and return no object", func() {
				objects, err := cli.Object().Find("pattern")
				Expect(err).ToNot(HaveOccurred())
				Expect(objects).To(HaveLen(0))
			})
		})
		Context("With the object", func() {
			It("Must succeed", func() {
				var err error

				err = cli.Object().MultipartPut("object", bytes.NewReader([]byte("Hello")))
				Expect(err).ToNot(HaveOccurred())

				objects, err := cli.Object().Find("object")
				Expect(err).ToNot(HaveOccurred())
				Expect(objects).To(HaveLen(1))

				err = cli.Object().Delete("object")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
	Context("List", func() {
		It("Must be possible to list buckets", func() {
			buckets, err := cli.Bucket().List()
			Expect(err).ToNot(HaveOccurred())
			Expect(buckets).To(HaveLen(1))
		})
	})

	/*
		 * Not Implemented whit minio
		 *
		Context("Versioning", func() {
			It("Must be possible to enable versioning", func() {
				Expect(cli.Bucket().SetVersioning(true)).To(Succeed())
			})
			It("Must be enabled", func() {
				status, err := cli.Bucket().GetVersioning()
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal("Enabled"))
			})
			It("Must be possible to suspend versioning", func() {
				Expect(cli.Bucket().SetVersioning(false)).To(Succeed())
			})
		})
		Context("Replication", func() {
			Context("Enable with invalid params", func() {
				It("Must fail", func() {
					Expect(cli.Bucket().EnableReplication("fake-src-role-arn", "fake-dst-role-arn", "fake-dst-bucket")).ToNot(Succeed())
				})
			})
			Context("Disable", func() {
				It("Must not return error", func() {
					Expect(cli.Bucket().DeleteReplication()).To(Succeed())
				})
			})
		})
		 *
	*/

	It("Must be possible to delete a bucket", func() {
		Expect(cli.Bucket().Delete()).To(Succeed())
	})
})
