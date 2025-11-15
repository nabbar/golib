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
	"sync"
	"time"

	libsiz "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 Stress Tests", func() {
	BeforeEach(func() {
		err := cli.Bucket().Check()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Concurrent operations", func() {
		Context("Concurrent uploads", func() {
			It("should handle 10 concurrent small uploads", func() {
				const numUploads = 10
				var wg sync.WaitGroup
				errors := make(chan error, numUploads)
				keys := make([]string, numUploads)

				wg.Add(numUploads)
				for i := 0; i < numUploads; i++ {
					keys[i] = "concurrent-small-" + GenerateUniqueName("") + ".dat"
					go func(idx int, key string) {
						defer wg.Done()
						defer GinkgoRecover()

						content := bytes.Repeat([]byte("test"), 1024) // 4KB
						err := cli.Object().Put(key, bytes.NewReader(content))
						if err != nil {
							errors <- err
						}
					}(i, keys[i])
				}

				wg.Wait()
				close(errors)

				// Check for errors
				for err := range errors {
					Expect(err).NotTo(HaveOccurred())
				}

				// Cleanup
				for _, key := range keys {
					_ = cli.Object().Delete(false, key)
				}
			})

			It("should handle 5 concurrent large uploads", func() {
				const numUploads = 5
				var wg sync.WaitGroup
				errors := make(chan error, numUploads)
				keys := make([]string, numUploads)

				wg.Add(numUploads)
				for i := 0; i < numUploads; i++ {
					keys[i] = "concurrent-large-" + GenerateUniqueName("") + ".dat"
					go func(idx int, key string) {
						defer wg.Done()
						defer GinkgoRecover()

						// 20MB each
						err := cli.Object().MultipartPut(key, randContent(20*libsiz.SizeMega))
						if err != nil {
							errors <- err
						}
					}(i, keys[i])
				}

				wg.Wait()
				close(errors)

				// Check for errors
				for err := range errors {
					Expect(err).NotTo(HaveOccurred())
				}

				// Cleanup
				for _, key := range keys {
					_ = cli.Object().Delete(false, key)
				}
			})
		})

		Context("Concurrent reads", func() {
			var testKey string

			BeforeEach(func() {
				testKey = "concurrent-read-" + GenerateUniqueName("") + ".dat"
				content := bytes.Repeat([]byte("test-content"), 1024) // ~12KB
				err := cli.Object().Put(testKey, bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				_ = cli.Object().Delete(false, testKey)
			})

			It("should handle 20 concurrent reads", func() {
				const numReads = 20
				var wg sync.WaitGroup
				errors := make(chan error, numReads)

				wg.Add(numReads)
				for i := 0; i < numReads; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						output, err := cli.Object().Get(testKey)
						if err != nil {
							errors <- err
							return
						}
						if output != nil && output.Body != nil {
							_ = output.Body.Close()
						}
					}()
				}

				wg.Wait()
				close(errors)

				// Check for errors
				for err := range errors {
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should handle concurrent Head requests", func() {
				const numRequests = 30
				var wg sync.WaitGroup
				errors := make(chan error, numRequests)

				wg.Add(numRequests)
				for i := 0; i < numRequests; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						_, err := cli.Object().Head(testKey)
						if err != nil {
							errors <- err
						}
					}()
				}

				wg.Wait()
				close(errors)

				// Check for errors
				for err := range errors {
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})

		Context("Mixed concurrent operations", func() {
			It("should handle mixed read/write operations", func() {
				const numOps = 15
				var wg sync.WaitGroup
				errors := make(chan error, numOps)
				keys := make([]string, numOps/3)

				// Upload some objects first
				for i := 0; i < numOps/3; i++ {
					keys[i] = "mixed-ops-" + GenerateUniqueName("") + ".dat"
					err := cli.Object().Put(keys[i], bytes.NewReader([]byte("initial content")))
					Expect(err).NotTo(HaveOccurred())
				}

				wg.Add(numOps)

				// Mix of uploads
				for i := 0; i < numOps/3; i++ {
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						key := "mixed-upload-" + GenerateUniqueName("") + ".dat"
						err := cli.Object().Put(key, bytes.NewReader([]byte("concurrent upload")))
						if err != nil {
							errors <- err
						}
						defer func() { _ = cli.Object().Delete(false, key) }()
					}(i)
				}

				// Mix of reads
				for i := 0; i < numOps/3; i++ {
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						if idx < len(keys) {
							output, err := cli.Object().Get(keys[idx])
							if err != nil {
								errors <- err
								return
							}
							if output != nil && output.Body != nil {
								_ = output.Body.Close()
							}
						}
					}(i)
				}

				// Mix of Head requests
				for i := 0; i < numOps/3; i++ {
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						if idx < len(keys) {
							_, err := cli.Object().Head(keys[idx])
							if err != nil {
								errors <- err
							}
						}
					}(i)
				}

				wg.Wait()
				close(errors)

				// Check for errors
				for err := range errors {
					Expect(err).NotTo(HaveOccurred())
				}

				// Cleanup
				for _, key := range keys {
					_ = cli.Object().Delete(false, key)
				}
			})
		})
	})

	Describe("Large file operations", func() {
		Context("Very large files", func() {
			It("should upload 100MB file", func() {
				if minioMode {
					Skip("Skipping large file test in MinIO to save time")
				}

				key := "large-100mb-" + GenerateUniqueName("") + ".dat"

				start := time.Now()
				err := cli.Object().MultipartPut(key, randContent(100*libsiz.SizeMega))
				elapsed := time.Since(start)

				Expect(err).NotTo(HaveOccurred())
				GinkgoWriter.Printf("100MB upload took: %v\n", elapsed)

				// Verify size
				size, err := cli.Object().Size(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(size).To(Equal(int64(100 * libsiz.SizeMega)))

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, key)
				}()
			})

			It("should upload 500MB file", func() {
				if minioMode {
					Skip("Skipping very large file test in MinIO to save time")
				}

				key := "large-500mb-" + GenerateUniqueName("") + ".dat"

				start := time.Now()
				err := cli.Object().MultipartPut(key, randContent(500*libsiz.SizeMega))
				elapsed := time.Since(start)

				Expect(err).NotTo(HaveOccurred())
				GinkgoWriter.Printf("500MB upload took: %v\n", elapsed)

				// Verify size
				size, err := cli.Object().Size(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(size).To(Equal(int64(500 * libsiz.SizeMega)))

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, key)
				}()
			})
		})

		Context("Custom part sizes", func() {
			It("should upload with small part size (5MB)", func() {
				key := "custom-part-5mb-" + GenerateUniqueName("") + ".dat"

				err := cli.Object().MultipartPutCustom(5*libsiz.SizeMega, key, randContent(25*libsiz.SizeMega))
				Expect(err).NotTo(HaveOccurred())

				size, err := cli.Object().Size(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(size).To(Equal(int64(25 * libsiz.SizeMega)))

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, key)
				}()
			})

			It("should upload with large part size (50MB)", func() {
				if minioMode {
					Skip("Skipping large part size test in MinIO to save time")
				}

				key := "custom-part-50mb-" + GenerateUniqueName("") + ".dat"

				err := cli.Object().MultipartPutCustom(50*libsiz.SizeMega, key, randContent(150*libsiz.SizeMega))
				Expect(err).NotTo(HaveOccurred())

				size, err := cli.Object().Size(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(size).To(Equal(int64(150 * libsiz.SizeMega)))

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, key)
				}()
			})
		})
	})

	Describe("Performance benchmarks", func() {
		Context("Upload performance", func() {
			It("should measure small file upload speed", func() {
				const numFiles = 10
				const fileSize = 1 * libsiz.SizeMega
				keys := make([]string, numFiles)

				start := time.Now()
				for i := 0; i < numFiles; i++ {
					keys[i] = "perf-small-" + GenerateUniqueName("") + ".dat"
					err := cli.Object().Put(keys[i], randContent(fileSize))
					Expect(err).NotTo(HaveOccurred())
				}
				elapsed := time.Since(start)

				totalMB := float64(numFiles*fileSize) / float64(libsiz.SizeMega)
				mbps := totalMB / elapsed.Seconds()

				GinkgoWriter.Printf("Uploaded %d x 1MB files in %v (%.2f MB/s)\n", numFiles, elapsed, mbps)

				// Cleanup
				for _, key := range keys {
					_ = cli.Object().Delete(false, key)
				}
			})

			It("should measure multipart upload speed", func() {
				const fileSize = 50 * libsiz.SizeMega
				key := "perf-multipart-" + GenerateUniqueName("") + ".dat"

				start := time.Now()
				err := cli.Object().MultipartPut(key, randContent(fileSize))
				elapsed := time.Since(start)

				Expect(err).NotTo(HaveOccurred())

				totalMB := float64(fileSize) / float64(libsiz.SizeMega)
				mbps := totalMB / elapsed.Seconds()

				GinkgoWriter.Printf("Uploaded 50MB file in %v (%.2f MB/s)\n", elapsed, mbps)

				// Cleanup
				defer func() {
					_ = cli.Object().Delete(false, key)
				}()
			})
		})

		Context("Download performance", func() {
			var testKey string

			BeforeEach(func() {
				testKey = "perf-download-" + GenerateUniqueName("") + ".dat"
				err := cli.Object().Put(testKey, randContent(10*libsiz.SizeMega))
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				_ = cli.Object().Delete(false, testKey)
			})

			It("should measure download speed", func() {
				start := time.Now()
				output, err := cli.Object().Get(testKey)
				Expect(err).NotTo(HaveOccurred())

				if output != nil && output.Body != nil {
					defer output.Body.Close()
					// Read all data
					buf := make([]byte, 4096)
					for {
						_, err := output.Body.Read(buf)
						if err != nil {
							break
						}
					}
				}
				elapsed := time.Since(start)

				totalMB := 10.0
				mbps := totalMB / elapsed.Seconds()

				GinkgoWriter.Printf("Downloaded 10MB file in %v (%.2f MB/s)\n", elapsed, mbps)
			})
		})
	})

	Describe("Batch operations", func() {
		It("should delete multiple objects at once", func() {
			const numObjects = 20
			keys := make([]string, numObjects)

			// Upload objects
			for i := 0; i < numObjects; i++ {
				keys[i] = "batch-delete-" + GenerateUniqueName("") + ".dat"
				err := cli.Object().Put(keys[i], bytes.NewReader([]byte("content")))
				Expect(err).NotTo(HaveOccurred())
			}

			// Delete all at once (if supported)
			for _, key := range keys {
				err := cli.Object().Delete(false, key)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should list many objects efficiently", func() {
			const numObjects = 50
			keys := make([]string, numObjects)
			prefix := "batch-list-" + GenerateUniqueName("")

			// Upload objects with same prefix
			for i := 0; i < numObjects; i++ {
				keys[i] = prefix + "-" + GenerateUniqueName("") + ".dat"
				err := cli.Object().Put(keys[i], bytes.NewReader([]byte("content")))
				Expect(err).NotTo(HaveOccurred())
			}

			// List with prefix
			objects, _, _, err := cli.Object().ListPrefix("", prefix)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(objects)).To(BeNumerically(">=", numObjects))

			// Cleanup
			for _, key := range keys {
				_ = cli.Object().Delete(false, key)
			}
		})
	})
})
