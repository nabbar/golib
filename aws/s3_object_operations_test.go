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
	"bytes"

	libsiz "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 Object - Operations", func() {
	BeforeEach(func() {
		// Ensure bucket exists (created in BeforeSuite)
		err := cli.Bucket().Check()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Object listing", func() {
		Context("Basic listing", func() {
			It("List() with empty token should succeed", func() {
				objects, token, count, err := cli.Object().List("")
				Expect(err).NotTo(HaveOccurred())
				// Note: objects may or may not be empty depending on test execution order
				_ = objects
				_ = token
				// Count is the MaxKeys parameter returned, not the number of objects
				Expect(count).To(BeNumerically(">=", 0))
			})
		})
	})

	Describe("Object Find", func() {
		Context("With pattern matching", func() {
			It("Find() should succeed with non-matching pattern", func() {
				objects, err := cli.Object().Find("non-existent-pattern-xyz123")
				Expect(err).NotTo(HaveOccurred())
				// Should find no objects with this pattern
				Expect(objects).To(HaveLen(0))
			})
		})

		Context("With objects in bucket", func() {
			It("Find() should locate uploaded object", func() {
				objectKey := "test-object.txt"
				content := []byte("test content")

				// Upload object
				err := cli.Object().Put(objectKey, bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())

				// Find object
				objects, err := cli.Object().Find(objectKey)
				Expect(err).NotTo(HaveOccurred())
				Expect(objects).To(HaveLen(1))
				Expect(objects[0]).To(Equal(objectKey))

				// Cleanup
				err = cli.Object().Delete(false, objectKey)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Object Put and Get", func() {
		Context("Basic operations", func() {
			It("Put() should upload an object successfully", func() {
				objectKey := "put-test.txt"
				content := []byte("Hello, World!")

				err := cli.Object().Put(objectKey, bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, objectKey)
				}()
			})

			It("Get() should retrieve uploaded object", func() {
				objectKey := "get-test.txt"
				content := []byte("Test content for Get")

				// Upload
				err := cli.Object().Put(objectKey, bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())

				// Get
				output, err := cli.Object().Get(objectKey)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())
				Expect(output.Body).NotTo(BeNil())

				defer output.Body.Close()

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, objectKey)
				}()
			})

			It("Head() should return object metadata", func() {
				objectKey := "head-test.txt"
				content := []byte("Metadata test")

				// Upload
				err := cli.Object().Put(objectKey, bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())

				// Head
				head, err := cli.Object().Head(objectKey)
				Expect(err).NotTo(HaveOccurred())
				Expect(head).NotTo(BeNil())
				Expect(head.ContentLength).NotTo(BeNil())
				Expect(*head.ContentLength).To(Equal(int64(len(content))))

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, objectKey)
				}()
			})
		})

		Context("Error cases", func() {
			It("Get() on non-existent object should fail", func() {
				output, err := cli.Object().Get("non-existent-object")
				Expect(err).To(HaveOccurred())
				if output != nil && output.Body != nil {
					_ = output.Body.Close()
				}
			})

			It("Head() on non-existent object should fail", func() {
				_, err := cli.Object().Head("non-existent-object")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Object Delete", func() {
		Context("With existing object", func() {
			It("Delete(check=false) should succeed", func() {
				objectKey := "delete-test.txt"

				// Upload object
				err := cli.Object().Put(objectKey, bytes.NewReader([]byte("to be deleted")))
				Expect(err).NotTo(HaveOccurred())

				// Delete without check
				err = cli.Object().Delete(false, objectKey)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Delete(check=true) should verify existence and delete", func() {
				objectKey := "delete-check-test.txt"

				// Upload object
				err := cli.Object().Put(objectKey, bytes.NewReader([]byte("to be deleted with check")))
				Expect(err).NotTo(HaveOccurred())

				// Delete with check
				err = cli.Object().Delete(true, objectKey)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("With non-existent object", func() {
			It("Delete(check=false) should succeed (no error for missing)", func() {
				err := cli.Object().Delete(false, "non-existent-object")
				Expect(err).NotTo(HaveOccurred())
			})

			It("Delete(check=true) should fail for non-existent object", func() {
				err := cli.Object().Delete(true, "non-existent-object")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Object Copy", func() {
		It("Copy() should duplicate an object", func() {
			sourceKey := "copy-source.txt"
			destKey := "copy-dest.txt"
			content := []byte("Content to copy")

			// Upload source
			err := cli.Object().Put(sourceKey, bytes.NewReader(content))
			Expect(err).NotTo(HaveOccurred())

			// Copy
			err = cli.Object().Copy(sourceKey, destKey)
			Expect(err).NotTo(HaveOccurred())

			// Verify destination exists
			head, err := cli.Object().Head(destKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(head).NotTo(BeNil())

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, sourceKey)
				_ = cli.Object().Delete(false, destKey)
			}()
		})
	})

	Describe("Multipart Put", func() {
		It("MultipartPut() should upload small object", func() {
			objectKey := "multipart-small.dat"

			err := cli.Object().MultipartPut(objectKey, randContent(500*libsiz.SizeKilo))
			Expect(err).NotTo(HaveOccurred())

			// Verify object exists
			objects, err := cli.Object().Find(objectKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(1))

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})

		It("MultipartPut() should upload large object", func() {
			objectKey := "multipart-large.dat"

			err := cli.Object().MultipartPut(objectKey, randContent(10*libsiz.SizeMega))
			Expect(err).NotTo(HaveOccurred())

			// Verify object exists
			objects, err := cli.Object().Find(objectKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(1))

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})
	})

	Describe("Object Size", func() {
		It("Size() should return correct object size", func() {
			objectKey := "size-test.txt"
			content := []byte("Size test content")

			// Upload
			err := cli.Object().Put(objectKey, bytes.NewReader(content))
			Expect(err).NotTo(HaveOccurred())

			// Get size
			size, err := cli.Object().Size(objectKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(len(content))))

			// Cleanup
			defer func() {
				_ = cli.Object().Delete(false, objectKey)
			}()
		})
	})
})
