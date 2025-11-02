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

package jfrog_test

import (
	"context"
	"net/http"

	. "github.com/nabbar/golib/artifact/jfrog"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

var _ = Describe("NewArtifactory", func() {
	It("returns error on invalid URL", func() {
		cli, err := NewArtifactory(context.Background(), func(req *http.Request) (*http.Response, error) { return nil, nil }, "::://bad url", `v(\d+\.\d+\.\d+)`, 1, "repo")
		Expect(cli).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	It("initializes model with parsed URL and fields without network", func() {
		// dummy Do func never called here
		cli, err := NewArtifactory(context.Background(), func(req *http.Request) (*http.Response, error) { return nil, nil }, "https://artifactory.example.com", `v(\d+\.\d+\.\d+)`, 1, "repo", "path")
		Expect(err).To(BeNil())
		Expect(cli).ToNot(BeNil())
		// We can't access unexported fields, but successful construction is enough here
	})
})
