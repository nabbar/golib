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
	hscvrs "github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	artcli "github.com/nabbar/golib/artifact/client"
)

var _ = Describe("Helper aggregation", func() {
	var vers hscvrs.Collection

	BeforeEach(func() {
		// Versions out of order with various majors/minors
		txt := []string{"1.0.0", "1.2.0", "1.2.5", "2.0.0", "2.1.3", "2.1.9", "3.0.1"}
		vers = make(hscvrs.Collection, 0, len(txt))
		for _, t := range txt {
			v, _ := hscvrs.NewVersion(t)
			vers = append(vers, v)
		}
	})

	It("ListReleases* should group and sort versions", func() {
		h := &artcli.Helper{F: func() (hscvrs.Collection, error) { return vers, nil }}

		ord, err := h.ListReleasesOrder()
		Expect(err).To(BeNil())
		Expect(ord).To(HaveKey(1))
		Expect(ord).To(HaveKey(2))
		Expect(ord).To(HaveKey(3))
		Expect(ord[1]).To(HaveKey(0))
		Expect(ord[1]).To(HaveKey(2))

		maj, err := h.ListReleasesMajor(2)
		Expect(err).To(BeNil())
		Expect(maj).To(HaveLen(3))
		// Sorted ascending
		Expect(maj[0].String()).To(Equal("2.0.0"))
		Expect(maj[len(maj)-1].String()).To(Equal("2.1.9"))

		min, err := h.ListReleasesMinor(1, 2)
		Expect(err).To(BeNil())
		Expect(min).To(HaveLen(2))
		Expect(min[0].String()).To(Equal("1.2.0"))
		Expect(min[1].String()).To(Equal("1.2.5"))
	})

	It("GetLatest* should return highest versions", func() {
		h := &artcli.Helper{F: func() (hscvrs.Collection, error) { return vers, nil }}

		latest, err := h.GetLatest()
		Expect(err).To(BeNil())
		Expect(latest.String()).To(Equal("3.0.1"))

		latestMaj, err := h.GetLatestMajor(2)
		Expect(err).To(BeNil())
		Expect(latestMaj.String()).To(Equal("2.1.9"))

		latestMin, err := h.GetLatestMinor(1, 2)
		Expect(err).To(BeNil())
		Expect(latestMin.String()).To(Equal("1.2.5"))
	})
})
