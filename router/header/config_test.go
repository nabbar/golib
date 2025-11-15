/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package header_test

import (
	rtrhdr "github.com/nabbar/golib/router/header"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Header/Config", func() {
	Describe("HeadersConfig", func() {
		It("should create Headers from empty config", func() {
			config := rtrhdr.HeadersConfig{}
			headers := config.New()

			Expect(headers).ToNot(BeNil())
			Expect(headers.Header()).To(BeEmpty())
		})

		It("should create Headers from config with values", func() {
			config := rtrhdr.HeadersConfig{
				"X-API-Version": "v1",
				"X-Request-ID":  "12345",
				"Cache-Control": "no-cache",
			}

			headers := config.New()
			Expect(headers).ToNot(BeNil())
			Expect(headers.Get("X-API-Version")).To(Equal("v1"))
			Expect(headers.Get("X-Request-ID")).To(Equal("12345"))
			Expect(headers.Get("Cache-Control")).To(Equal("no-cache"))
		})

		It("should validate config", func() {
			config := rtrhdr.HeadersConfig{
				"X-Custom": "value",
			}

			err := config.Validate()
			Expect(err).To(BeNil())
		})
	})
})
