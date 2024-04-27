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
	. "github.com/onsi/ginkgo/v2"
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
