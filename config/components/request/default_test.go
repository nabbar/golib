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

package request_test

import (
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/request"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Default Configuration", func() {
	var (
		cpt ComponentRequest
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil)
	})

	Describe("DefaultConfig method", func() {
		It("should return non-empty configuration", func() {
			config := cpt.DefaultConfig("")
			Expect(config).NotTo(BeEmpty())
		})

		It("should return valid JSON", func() {
			config := cpt.DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).To(BeNil())
		})

		It("should support indentation", func() {
			config := cpt.DefaultConfig("  ")
			Expect(config).NotTo(BeEmpty())
		})

		It("should be consistent", func() {
			config1 := cpt.DefaultConfig("")
			config2 := cpt.DefaultConfig("")
			Expect(config1).To(Equal(config2))
		})
	})
})
