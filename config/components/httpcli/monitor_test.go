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
	montps "github.com/nabbar/golib/monitor/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor Integration", func() {
	var (
		cpt CptHTTPClient
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil, false, nil)
	})

	Describe("RegisterMonitorPool method", func() {
		It("should not panic when registering nil pool", func() {
			Expect(func() {
				cpt.RegisterMonitorPool(nil)
			}).NotTo(Panic())
		})

		It("should not panic when registering valid pool function", func() {
			poolFunc := func() montps.FuncPool {
				return func() montps.Pool { return nil }
			}()

			Expect(func() {
				cpt.RegisterMonitorPool(poolFunc)
			}).NotTo(Panic())
		})
	})
})
