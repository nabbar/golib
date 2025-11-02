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

package artifact_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	hscvrs "github.com/hashicorp/go-version"
	libart "github.com/nabbar/golib/artifact"
)

var _ = Describe("artifact helpers", func() {
	It("CheckRegex should match names using Go regex syntax", func() {
		Expect(libart.CheckRegex("file-1.2.3.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeTrue())
		Expect(libart.CheckRegex("file-alpha.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeFalse())
	})

	It("ValidatePreRelease should reject common prerelease tags and accept GA", func() {
		v, _ := hscvrs.NewVersion("1.2.3-beta.1")
		Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		v, _ = hscvrs.NewVersion("1.2.3-rc.0")
		Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		v, _ = hscvrs.NewVersion("1.2.3")
		Expect(libart.ValidatePreRelease(v)).To(BeTrue())
	})

	It("DownloadRelease should panic as not implemented", func() {
		Expect(func() { _ = callDownloadReleasePanic() }).To(Panic())
	})
})

// helper to call the function causing panic
func callDownloadReleasePanic() any {
	f, err := libart.DownloadRelease("")
	_ = f
	return err
}
