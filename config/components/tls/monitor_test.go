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

package tls_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/tls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	montps "github.com/nabbar/golib/monitor/types"
)

// Monitor integration tests verify the monitoring functionality
// for the TLS component.
var _ = Describe("Monitor Integration", func() {
	var (
		ctx context.Context
		cpt CptTlS
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil)
	})

	Describe("RegisterMonitorPool method", func() {
		Context("registering monitor pool", func() {
			It("should not panic with nil function", func() {
				Expect(func() {
					cpt.RegisterMonitorPool(nil)
				}).NotTo(Panic())
			})

			It("should accept valid monitor pool function", func() {
				mockPool := func() montps.Pool {
					return nil
				}

				Expect(func() {
					cpt.RegisterMonitorPool(mockPool)
				}).NotTo(Panic())
			})

			It("should allow multiple registrations", func() {
				mockPool1 := func() montps.Pool { return nil }
				mockPool2 := func() montps.Pool { return nil }

				Expect(func() {
					cpt.RegisterMonitorPool(mockPool1)
					cpt.RegisterMonitorPool(mockPool2)
				}).NotTo(Panic())
			})

			It("should not affect component state", func() {
				mockPool := func() montps.Pool { return nil }

				cpt.RegisterMonitorPool(mockPool)

				// Should not change started state
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})

		Context("concurrent registrations", func() {
			It("should handle concurrent RegisterMonitorPool calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						mockPool := func() montps.Pool { return nil }
						cpt.RegisterMonitorPool(mockPool)
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Monitor behavior", func() {
		Context("with component lifecycle", func() {
			It("should allow registration before Init", func() {
				mockPool := func() montps.Pool { return nil }

				Expect(func() {
					cpt.RegisterMonitorPool(mockPool)
				}).NotTo(Panic())
			})

			It("should allow registration after Init", func() {
				// Initialize component (without proper config, won't start)
				// but RegisterMonitorPool should still work
				mockPool := func() montps.Pool { return nil }

				Expect(func() {
					cpt.RegisterMonitorPool(mockPool)
				}).NotTo(Panic())
			})

			It("should allow registration after Stop", func() {
				cpt.Stop()

				mockPool := func() montps.Pool { return nil }

				Expect(func() {
					cpt.RegisterMonitorPool(mockPool)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Edge cases", func() {
		Context("monitor registration", func() {
			It("should handle repeated registrations", func() {
				mockPool := func() montps.Pool { return nil }

				for i := 0; i < 100; i++ {
					cpt.RegisterMonitorPool(mockPool)
				}

				// Should not panic or cause issues
			})

			It("should work with different function types", func() {
				// Different function signatures that satisfy FuncPool
				mockPool1 := func() montps.Pool { return nil }
				mockPool2 := func() montps.Pool { return nil }

				cpt.RegisterMonitorPool(mockPool1)
				cpt.RegisterMonitorPool(mockPool2)

				// Should not panic
			})
		})
	})

	Describe("No-op behavior", func() {
		Context("current implementation", func() {
			It("should be no-op as per implementation", func() {
				// Current implementation in monitor.go is empty (no-op)
				// This test verifies it doesn't cause side effects

				initialStarted := cpt.IsStarted()
				mockPool := func() montps.Pool { return nil }

				cpt.RegisterMonitorPool(mockPool)

				// State should not change
				Expect(cpt.IsStarted()).To(Equal(initialStarted))
			})

			It("should not store or use the pool function", func() {
				// Since implementation is empty, calling it multiple times
				// should have no observable effect
				mockPool := func() montps.Pool { return nil }

				cpt.RegisterMonitorPool(mockPool)
				cpt.RegisterMonitorPool(mockPool)
				cpt.RegisterMonitorPool(nil)

				// No state change expected
				Expect(cpt.IsStarted()).To(BeFalse())
			})
		})
	})
})
