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

package client_test

import (
	"errors"

	hscvrs "github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	artcli "github.com/nabbar/golib/artifact/client"
)

var _ = Describe("client/Helper advanced scenarios", func() {
	var baseVersions hscvrs.Collection

	BeforeEach(func() {
		// Create a comprehensive set of versions for testing
		versionStrings := []string{
			"0.1.0", "0.1.1", "0.2.0",
			"1.0.0", "1.0.1", "1.0.2",
			"1.1.0", "1.1.1",
			"1.2.0", "1.2.1", "1.2.5", "1.2.10",
			"2.0.0", "2.0.1",
			"2.1.0", "2.1.3", "2.1.9",
			"3.0.0", "3.0.1",
			"3.1.0",
			"10.0.0", "10.1.5",
		}
		baseVersions = make(hscvrs.Collection, 0, len(versionStrings))
		for _, vs := range versionStrings {
			v, _ := hscvrs.NewVersion(vs)
			baseVersions = append(baseVersions, v)
		}
	})

	Context("ListReleasesOrder functionality", func() {
		It("should correctly organize versions by major and minor", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			ordered, err := h.ListReleasesOrder()
			Expect(err).ToNot(HaveOccurred())
			Expect(ordered).ToNot(BeNil())

			// Verify major version keys
			Expect(ordered).To(HaveKey(0))
			Expect(ordered).To(HaveKey(1))
			Expect(ordered).To(HaveKey(2))
			Expect(ordered).To(HaveKey(3))
			Expect(ordered).To(HaveKey(10))

			// Verify minor version keys for major 1
			Expect(ordered[1]).To(HaveKey(0))
			Expect(ordered[1]).To(HaveKey(1))
			Expect(ordered[1]).To(HaveKey(2))

			// Verify versions are sorted
			v12 := ordered[1][2]
			Expect(v12[0].String()).To(Equal("1.2.0"))
			Expect(v12[len(v12)-1].String()).To(Equal("1.2.10"))
		})

		It("should handle empty version collection", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return hscvrs.Collection{}, nil },
			}

			ordered, err := h.ListReleasesOrder()
			Expect(err).ToNot(HaveOccurred())
			Expect(ordered).To(BeEmpty())
		})

		It("should handle single version", func() {
			v, _ := hscvrs.NewVersion("1.0.0")
			singleVersion := hscvrs.Collection{v}

			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return singleVersion, nil },
			}

			ordered, err := h.ListReleasesOrder()
			Expect(err).ToNot(HaveOccurred())
			Expect(ordered).To(HaveKey(1))
			Expect(ordered[1]).To(HaveKey(0))
			Expect(ordered[1][0]).To(HaveLen(1))
		})

		It("should propagate errors from version fetcher", func() {
			testErr := errors.New("fetch error")
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return nil, testErr },
			}

			_, err := h.ListReleasesOrder()
			Expect(err).To(Equal(testErr))
		})
	})

	Context("ListReleasesMajor functionality", func() {
		It("should return all versions for a specific major version", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			major1, err := h.ListReleasesMajor(1)
			Expect(err).ToNot(HaveOccurred())
			Expect(major1).ToNot(BeNil())
			Expect(major1).To(HaveLen(9)) // 1.0.0, 1.0.1, 1.0.2, 1.1.0, 1.1.1, 1.2.0, 1.2.1, 1.2.5, 1.2.10

			// Verify all are major version 1
			for _, v := range major1 {
				Expect(v.Segments()[0]).To(Equal(1))
			}

			// Verify sorting (ascending)
			Expect(major1[0].String()).To(Equal("1.0.0"))
			Expect(major1[len(major1)-1].String()).To(Equal("1.2.10"))
		})

		It("should return empty collection for non-existent major version", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			result, err := h.ListReleasesMajor(99)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil()) // Returns nil, not an error
		})

		It("should handle major version 0", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			major0, err := h.ListReleasesMajor(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(major0).To(HaveLen(3)) // 0.1.0, 0.1.1, 0.2.0
		})

		It("should handle double-digit major versions", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			major10, err := h.ListReleasesMajor(10)
			Expect(err).ToNot(HaveOccurred())
			Expect(major10).To(HaveLen(2)) // 10.0.0, 10.1.5
		})
	})

	Context("ListReleasesMinor functionality", func() {
		It("should return all versions for a specific major.minor", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			minor12, err := h.ListReleasesMinor(1, 2)
			Expect(err).ToNot(HaveOccurred())
			Expect(minor12).ToNot(BeNil())
			Expect(minor12).To(HaveLen(4)) // 1.2.0, 1.2.1, 1.2.5, 1.2.10

			// Verify all are version 1.2.x
			for _, v := range minor12 {
				Expect(v.Segments()[0]).To(Equal(1))
				Expect(v.Segments()[1]).To(Equal(2))
			}

			// Verify sorting
			Expect(minor12[0].String()).To(Equal("1.2.0"))
			Expect(minor12[len(minor12)-1].String()).To(Equal("1.2.10"))
		})

		It("should return empty collection for non-existent major.minor", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			result, err := h.ListReleasesMinor(1, 99)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil()) // Returns nil, not an error

			result, err = h.ListReleasesMinor(99, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil()) // Returns nil, not an error
		})

		It("should handle minor version 0", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			minor10, err := h.ListReleasesMinor(1, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(minor10).To(HaveLen(3)) // 1.0.0, 1.0.1, 1.0.2
		})
	})

	Context("GetLatest functionality", func() {
		It("should return the highest version overall", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			latest, err := h.GetLatest()
			Expect(err).ToNot(HaveOccurred())
			Expect(latest).ToNot(BeNil())
			Expect(latest.String()).To(Equal("10.1.5"))
		})

		It("should return nil for empty version collection", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return hscvrs.Collection{}, nil },
			}

			result, err := h.GetLatest()
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil()) // Returns nil, not an error
		})

		It("should work with single version", func() {
			v, _ := hscvrs.NewVersion("5.0.0")
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return hscvrs.Collection{v}, nil },
			}

			latest, err := h.GetLatest()
			Expect(err).ToNot(HaveOccurred())
			Expect(latest.String()).To(Equal("5.0.0"))
		})
	})

	Context("GetLatestMajor functionality", func() {
		It("should return the highest version for a specific major", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			latestMajor1, err := h.GetLatestMajor(1)
			Expect(err).ToNot(HaveOccurred())
			Expect(latestMajor1).ToNot(BeNil())
			Expect(latestMajor1.String()).To(Equal("1.2.10"))

			latestMajor2, err := h.GetLatestMajor(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(latestMajor2.String()).To(Equal("2.1.9"))
		})

		It("should return nil for non-existent major version", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			result, err := h.GetLatestMajor(99)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil()) // Returns nil, not an error
		})
	})

	Context("GetLatestMinor functionality", func() {
		It("should return the highest version for a specific major.minor", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			latestMinor12, err := h.GetLatestMinor(1, 2)
			Expect(err).ToNot(HaveOccurred())
			Expect(latestMinor12).ToNot(BeNil())
			Expect(latestMinor12.String()).To(Equal("1.2.10"))

			latestMinor21, err := h.GetLatestMinor(2, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(latestMinor21.String()).To(Equal("2.1.9"))
		})

		It("should return nil for non-existent major.minor", func() {
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return baseVersions, nil },
			}

			result, err := h.GetLatestMinor(1, 99)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil()) // Returns nil, not an error

			result, err = h.GetLatestMinor(99, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil()) // Returns nil, not an error
		})
	})

	Context("Error propagation", func() {
		It("should propagate errors from all methods", func() {
			testErr := errors.New("test error")
			h := &artcli.Helper{
				F: func() (hscvrs.Collection, error) { return nil, testErr },
			}

			_, err := h.ListReleasesOrder()
			Expect(err).To(Equal(testErr))

			_, err = h.ListReleasesMajor(1)
			Expect(err).To(Equal(testErr))

			_, err = h.ListReleasesMinor(1, 0)
			Expect(err).To(Equal(testErr))

			_, err = h.GetLatest()
			Expect(err).To(Equal(testErr))

			_, err = h.GetLatestMajor(1)
			Expect(err).To(Equal(testErr))

			_, err = h.GetLatestMinor(1, 0)
			Expect(err).To(Equal(testErr))
		})
	})
})
