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

	tlscas "github.com/nabbar/golib/certificates/ca"
	. "github.com/nabbar/golib/config/components/httpcli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Model Methods", func() {
	var (
		cpt CptHTTPClient
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil, false, nil)
	})

	Describe("Config method", func() {
		It("should return empty config for unstarted component", func() {
			cfg := cpt.Config()
			Expect(cfg).NotTo(BeNil())
		})
	})

	Describe("SetFuncMessage method", func() {
		It("should set message function", func() {
			msg := func(m string) {}
			Expect(func() {
				cpt.SetFuncMessage(msg)
			}).NotTo(Panic())
		})

		It("should not panic with nil message", func() {
			Expect(func() {
				cpt.SetFuncMessage(nil)
			}).NotTo(Panic())
		})
	})

	Describe("SetAsDefaultHTTPClient method", func() {
		It("should set as default", func() {
			Expect(func() {
				cpt.SetAsDefaultHTTPClient(true)
			}).NotTo(Panic())
		})

		It("should unset as default", func() {
			Expect(func() {
				cpt.SetAsDefaultHTTPClient(false)
			}).NotTo(Panic())
		})
	})

	Describe("SetDefault method", func() {
		It("should not panic", func() {
			Expect(func() {
				cpt.SetDefault()
			}).NotTo(Panic())
		})
	})

	Describe("Creation with CA root", func() {
		It("should handle nil CA root", func() {
			c := New(ctx, nil, false, nil)
			Expect(c).NotTo(BeNil())
		})

		It("should handle custom CA root function", func() {
			defCARoot := func() tlscas.Cert {
				return nil
			}
			c := New(ctx, defCARoot, false, nil)
			Expect(c).NotTo(BeNil())
		})
	})
})
