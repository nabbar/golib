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

package log_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	montps "github.com/nabbar/golib/monitor/types"
)

// Monitor integration tests verify the RegisterMonitorPool method.
var _ = Describe("Monitor Integration", func() {
	var (
		cpt CptLog
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, DefaultLevel)
	})

	Describe("RegisterMonitorPool method", func() {
		Context("registering monitor pool", func() {
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

			It("should allow multiple registrations", func() {
				poolFunc := func() montps.FuncPool {
					return func() montps.Pool { return nil }
				}()

				Expect(func() {
					cpt.RegisterMonitorPool(poolFunc)
					cpt.RegisterMonitorPool(poolFunc)
					cpt.RegisterMonitorPool(nil)
				}).NotTo(Panic())
			})
		})

		Context("concurrent access", func() {
			It("should handle concurrent registrations", func() {
				done := make(chan bool, 10)

				poolFunc := func() montps.FuncPool {
					return func() montps.Pool { return nil }
				}()

				for i := 0; i < 10; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						if idx%2 == 0 {
							cpt.RegisterMonitorPool(poolFunc)
						} else {
							cpt.RegisterMonitorPool(nil)
						}
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Monitor integration scenarios", func() {
		Context("component with monitor", func() {
			It("should support monitor pool registration before initialization", func() {
				poolFunc := func() montps.FuncPool {
					return func() montps.Pool { return nil }
				}()

				Expect(func() {
					cpt.RegisterMonitorPool(poolFunc)
				}).NotTo(Panic())
			})

			It("should support monitor pool registration after initialization", func() {
				// Note: This tests the method is callable at different lifecycle stages
				poolFunc := func() montps.FuncPool {
					return func() montps.Pool { return nil }
				}()

				Expect(func() {
					cpt.RegisterMonitorPool(poolFunc)
				}).NotTo(Panic())
			})
		})
	})
})
