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

var _ = Describe("artifact/helpers", func() {
	Context("CheckRegex function", func() {
		It("should match valid patterns", func() {
			Expect(libart.CheckRegex("file-1.2.3.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeTrue())
			Expect(libart.CheckRegex("myapp-v2.0.1-linux-amd64.zip", `myapp-v\d+\.\d+\.\d+-linux-amd64\.zip`)).To(BeTrue())
			Expect(libart.CheckRegex("release-3.14.159.tar.gz", `release-\d+\.\d+\.\d+\.tar\.gz`)).To(BeTrue())
		})

		It("should not match invalid patterns", func() {
			Expect(libart.CheckRegex("file-alpha.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeFalse())
			Expect(libart.CheckRegex("no-version.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeFalse())
			Expect(libart.CheckRegex("file-1.2.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeFalse())
		})

		It("should handle special regex characters", func() {
			Expect(libart.CheckRegex("file.tar.gz", `file\.tar\.gz`)).To(BeTrue())
			Expect(libart.CheckRegex("file[1].tar.gz", `file\[\d+\]\.tar\.gz`)).To(BeTrue())
			Expect(libart.CheckRegex("file(test).tar.gz", `file\(test\)\.tar\.gz`)).To(BeTrue())
		})

		It("should handle empty strings", func() {
			Expect(libart.CheckRegex("", `.*`)).To(BeTrue())
			Expect(libart.CheckRegex("anything", ``)).To(BeTrue()) // empty regex matches empty string
			Expect(libart.CheckRegex("", `\d+`)).To(BeFalse())
		})

		It("should be case-sensitive by default", func() {
			Expect(libart.CheckRegex("File-1.2.3.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeFalse())
			Expect(libart.CheckRegex("FILE-1.2.3.tar.gz", `file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeFalse())
		})

		It("should support case-insensitive patterns", func() {
			Expect(libart.CheckRegex("File-1.2.3.tar.gz", `(?i)file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeTrue())
			Expect(libart.CheckRegex("FILE-1.2.3.TAR.GZ", `(?i)file-\d+\.\d+\.\d+\.tar\.gz`)).To(BeTrue())
		})
	})

	Context("ValidatePreRelease function", func() {
		It("should reject alpha versions", func() {
			v, _ := hscvrs.NewVersion("1.2.3-alpha")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-alpha.1")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-a")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("2.0.0-Alpha.5")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		})

		It("should reject beta versions", func() {
			v, _ := hscvrs.NewVersion("1.2.3-beta")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-beta.1")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-b")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("2.0.0-BETA.2")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		})

		It("should reject release candidate versions", func() {
			v, _ := hscvrs.NewVersion("1.2.3-rc")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-rc.1")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-RC.0")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		})

		It("should reject dev/test/draft versions", func() {
			v, _ := hscvrs.NewVersion("1.2.3-dev")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-test")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-draft")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		})

		It("should reject master/main branch tags", func() {
			v, _ := hscvrs.NewVersion("1.2.3-master")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-main")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		})

		It("should accept GA (general availability) versions", func() {
			v, _ := hscvrs.NewVersion("1.2.3")
			Expect(libart.ValidatePreRelease(v)).To(BeTrue())

			v, _ = hscvrs.NewVersion("2.0.0")
			Expect(libart.ValidatePreRelease(v)).To(BeTrue())

			v, _ = hscvrs.NewVersion("10.5.27")
			Expect(libart.ValidatePreRelease(v)).To(BeTrue())
		})

		It("should accept versions with valid custom prerelease tags", func() {
			// Tags that don't contain the blacklisted words
			v, _ := hscvrs.NewVersion("1.2.3-stable")
			Expect(libart.ValidatePreRelease(v)).To(BeTrue())

			v, _ = hscvrs.NewVersion("1.2.3-release")
			Expect(libart.ValidatePreRelease(v)).To(BeTrue())

			v, _ = hscvrs.NewVersion("1.2.3-final")
			Expect(libart.ValidatePreRelease(v)).To(BeTrue())

			v, _ = hscvrs.NewVersion("1.2.3-hotfix")
			Expect(libart.ValidatePreRelease(v)).To(BeTrue())
		})

		It("should be case-insensitive for prerelease validation", func() {
			v, _ := hscvrs.NewVersion("1.2.3-ALPHA")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-Beta")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())

			v, _ = hscvrs.NewVersion("1.2.3-RC")
			Expect(libart.ValidatePreRelease(v)).To(BeFalse())
		})
	})

	Context("DownloadRelease function", func() {
		It("should panic as not implemented", func() {
			Expect(func() {
				_, _ = libart.DownloadRelease("http://example.com/release.tar.gz")
			}).To(Panic())
		})

		It("should panic with empty link", func() {
			Expect(func() {
				_, _ = libart.DownloadRelease("")
			}).To(Panic())
		})
	})
})
