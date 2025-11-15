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

package gitlab_test

import (
	"net/http"

	. "github.com/nabbar/golib/artifact/gitlab"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"

	gl "gitlab.com/gitlab-org/api/client-go"
)

var _ = Describe("GetGitlabOptions", func() {
	It("should append /api/v4 to base URL and wire HTTP client when provided", func() {
		httpcli := &http.Client{}
		opts, err := GetGitlabOptions("https://gitlab.example.com", httpcli)
		Expect(err).To(BeNil())
		Expect(opts).ToNot(BeEmpty())

		// Build a client with these options and ensure BaseURL is correct
		c, err := gl.NewClient("token", opts...)
		Expect(err).To(BeNil())
		Expect(c.BaseURL().String()).To(Equal("https://gitlab.example.com/api/v4/"))
	})

	It("should not duplicate api/version when already present", func() {
		opts, err := GetGitlabOptions("https://gitlab.example.com/api/v4", nil)
		Expect(err).To(BeNil())
		c, err := gl.NewClient("token", opts...)
		Expect(err).To(BeNil())
		Expect(c.BaseURL().String()).To(Equal("https://gitlab.example.com/api/v4/"))
	})
})
