/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package httpcli_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/httpcli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS Operations", func() {
	var (
		cpt CptHTTPClient
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil, false, nil)
	})

	Describe("DNS mapping operations", func() {
		It("should handle Add without panic", func() {
			Expect(func() {
				cpt.Add("example.com", "127.0.0.1")
			}).NotTo(Panic())
		})

		It("should handle Get on unstarted component", func() {
			result := cpt.Get("example.com")
			Expect(result).To(Equal(""))
		})

		It("should handle Del without panic", func() {
			Expect(func() {
				cpt.Del("example.com")
			}).NotTo(Panic())
		})

		It("should handle Len on unstarted component", func() {
			length := cpt.Len()
			Expect(length).To(Equal(0))
		})

		It("should handle Walk without panic", func() {
			Expect(func() {
				cpt.Walk(func(from, to string) bool {
					return true
				})
			}).NotTo(Panic())
		})
	})

	Describe("DNS client operations", func() {
		It("should return error for DialContext on unstarted component", func() {
			_, err := cpt.DialContext(context.Background(), "tcp", "example.com:80")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for Clean on unstarted component", func() {
			_, _, err := cpt.Clean("example.com:80")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for Search on unstarted component", func() {
			_, err := cpt.Search("example.com")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for SearchWithCache on unstarted component", func() {
			_, err := cpt.SearchWithCache("example.com")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("HTTP client operations", func() {
		It("should return nil Transport on unstarted component", func() {
			transport := cpt.DefaultTransport()
			Expect(transport).To(BeNil())
		})

		It("should return nil Client on unstarted component", func() {
			client := cpt.DefaultClient()
			Expect(client).To(BeNil())
		})

		It("should handle RegisterTransport without panic", func() {
			Expect(func() {
				cpt.RegisterTransport(nil)
			}).NotTo(Panic())
		})

		It("should handle Close on unstarted component", func() {
			err := cpt.Close()
			Expect(err).To(BeNil())
		})
	})

	Describe("Configuration retrieval", func() {
		It("should return empty config from GetConfig on unstarted component", func() {
			cfg := cpt.GetConfig()
			Expect(cfg).NotTo(BeNil())
		})
	})
})
