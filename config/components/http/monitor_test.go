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

package http_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	montps "github.com/nabbar/golib/monitor/types"
	libcmd "github.com/nabbar/golib/shell/command"
)

// Monitor tests verify monitoring integration
var _ = Describe("Monitoring Integration", func() {
	var (
		cpt CptHttp
		ctx context.Context
		cnl context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cnl = context.WithCancel(x)
		cpt = New(ctx, DefaultTlsKey, nil)
	})

	AfterEach(func() {
		if cnl != nil {
			cnl()
		}
	})

	Describe("RegisterMonitorPool method", func() {
		Context("registering monitor pool", func() {
			It("should accept valid monitor pool function", func() {
				poolFunc := func() montps.Pool {
					return nil
				}

				Expect(func() {
					cpt.RegisterMonitorPool(poolFunc)
				}).NotTo(Panic())
			})

			It("should accept nil monitor pool function", func() {
				Expect(func() {
					cpt.RegisterMonitorPool(nil)
				}).NotTo(Panic())
			})

			It("should allow registering multiple times", func() {
				poolFunc1 := func() montps.Pool { return nil }
				poolFunc2 := func() montps.Pool { return nil }

				cpt.RegisterMonitorPool(poolFunc1)
				cpt.RegisterMonitorPool(poolFunc2)
				// Should not panic
			})

			It("should handle function returning nil pool", func() {
				poolFunc := func() montps.Pool {
					return nil
				}

				cpt.RegisterMonitorPool(poolFunc)
				// Should not panic
			})
		})

		Context("with real monitor pool", func() {
			It("should register pool successfully", func() {
				Expect(func() {
					cpt.RegisterMonitorPool(fp)
				}).NotTo(Panic())
			})
		})

		Context("with mock monitor pool", func() {
			It("should register mock pool successfully", func() {
				mockPool := &mockMonitorPool{}
				poolFunc := func() montps.Pool {
					return mockPool
				}

				Expect(func() {
					cpt.RegisterMonitorPool(poolFunc)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Monitor pool integration", func() {
		Context("without registered pool", func() {
			It("should handle missing monitor pool gracefully", func() {
				// Without registered pool, component should still work
				// (monitor is optional)
				Expect(cpt).NotTo(BeNil())
			})
		})

		Context("with registered pool", func() {
			It("should work with registered pool", func() {
				mockPool := &mockMonitorPool{}
				poolFunc := func() montps.Pool {
					return mockPool
				}

				cpt.RegisterMonitorPool(poolFunc)

				// Component should work normally
				Expect(cpt.Type()).To(Equal(ComponentType))
			})
		})
	})

	Describe("Concurrent operations", func() {
		Context("concurrent monitor registration", func() {
			It("should handle concurrent RegisterMonitorPool calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						poolFunc := func() montps.Pool {
							return nil
						}
						cpt.RegisterMonitorPool(poolFunc)
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Edge cases", func() {
		Context("nil component", func() {
			It("should panic on RegisterMonitorPool with nil component", func() {
				var nilCpt CptHttp
				poolFunc := func() montps.Pool { return nil }

				Expect(func() {
					nilCpt.RegisterMonitorPool(poolFunc)
				}).To(Panic())
			})
		})

		Context("monitor pool returning different values", func() {
			It("should handle dynamic pool function", func() {
				counter := 0
				poolFunc := func() montps.Pool {
					counter++
					if counter%2 == 0 {
						return &mockMonitorPool{}
					}
					return nil
				}

				cpt.RegisterMonitorPool(poolFunc)
				// Should not panic
			})
		})
	})
})

// mockMonitorPool is a mock implementation of montps.Pool for testing
type mockMonitorPool struct{}

func (m *mockMonitorPool) MonitorSet(mon montps.Monitor) error {
	return nil
}

func (m *mockMonitorPool) MonitorGet(key string) montps.Monitor {
	return nil
}

func (m *mockMonitorPool) MonitorList() []string {
	return []string{}
}

func (m *mockMonitorPool) MonitorWalk(fct func(key string, mon montps.Monitor) bool, exclude ...string) {
}

func (m *mockMonitorPool) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockMonitorPool) SetRouteHealth(route string) {
}

func (m *mockMonitorPool) RegisterLoggerDefault(fct interface{}) {
}

func (m *mockMonitorPool) GetShellCommand(ctx context.Context) []libcmd.Command {
	return nil
}

func (m *mockMonitorPool) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}

func (m *mockMonitorPool) MarshalText() ([]byte, error) {
	return []byte("mockPool"), nil
}

func (m *mockMonitorPool) MonitorAdd(mon montps.Monitor) error {
	return nil
}

func (m *mockMonitorPool) MonitorDel(key string) {
}
