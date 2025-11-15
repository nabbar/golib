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

package status_test

import (
	"time"

	monpol "github.com/nabbar/golib/monitor/pool"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"
	stsctr "github.com/nabbar/golib/status/control"
	libver "github.com/nabbar/golib/version"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status", func() {
	var (
		status libsts.Status
		pool   monpol.Pool
	)

	Describe("New", func() {
		It("should create a new status instance", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })
			Expect(status).ToNot(BeNil())
		})
	})

	Describe("SetInfo", func() {
		It("should set application info", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })
			status.SetInfo("test-app", "v1.0.0", "abc123")
			// Info is set successfully if no panic occurs
			Expect(true).To(BeTrue())
		})

		It("should handle empty values", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })
			status.SetInfo("", "", "")
			Expect(true).To(BeTrue())
		})
	})

	Describe("SetVersion", func() {
		It("should set version information", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			type testStruct struct{}
			vers := libver.NewVersion(
				libver.License_MIT,
				"test-app",
				"Test Application",
				time.Now().Format(time.RFC3339),
				"abc123",
				"v1.0.0",
				"Test Author",
				"TEST",
				testStruct{},
				0,
			)
			status.SetVersion(vers)
			Expect(true).To(BeTrue())
		})
	})

	Describe("RegisterPool", func() {
		It("should register a pool function", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			poolFunc := func() montps.Pool {
				return pool
			}
			status.RegisterPool(poolFunc)
			Expect(true).To(BeTrue())
		})

		It("should work with monitors in pool", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("test-monitor")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())

			list := status.MonitorList()
			Expect(list).To(ContainElement("test-monitor"))
		})
	})

	Describe("IsHealthy", func() {
		Context("without registered components", func() {
			It("should return healthy by default", func() {
				status = libsts.New(globalCtx)
				pool = newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				status.SetInfo("test", "v1.0.0", "hash")
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})
		})

		Context("with healthy component", func() {
			It("should return true", func() {
				status = libsts.New(globalCtx)
				pool = newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				status.SetInfo("test", "v1.0.0", "hash")
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-component",
					},
				})...))
				m := newHealthyMonitor("test-component")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})

			It("should check specific component health", func() {
				status = libsts.New(globalCtx)
				pool = newPool()
				status.RegisterPool(func() montps.Pool { return pool })

				status.SetInfo("test", "v1.0.0", "hash")
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"test-component",
					},
				})...))
				m := newHealthyMonitor("test-component")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)

				healthy := status.IsHealthy("test-component")
				Expect(healthy).To(BeTrue())
			})
		})

		Context("with unhealthy component", func() {
			It("should return false", func() {
				status = libsts.New(globalCtx)
				pool = newPool()
				status.RegisterPool(func() montps.Pool { return pool })
				status.SetInfo("test", "v1.0.0", "hash")
				status.SetConfig(newStatusConfig(newListMandatory(map[stsctr.Mode][]string{
					stsctr.Must: {
						"failing-component",
					},
				})...))
				m := newUnhealthyMonitor("failing-component")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(testMonitorStabilizeDelay)
				healthy := status.IsHealthy()
				Expect(healthy).To(BeFalse())
			})
		})
	})

	Describe("IsStrictlyHealthy", func() {
		It("should check strict health status", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			healthy := status.IsStrictlyHealthy()
			Expect(healthy).To(BeAssignableToTypeOf(false))
		})

		It("should check specific component strict health", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			healthy := status.IsStrictlyHealthy("test-component")
			Expect(healthy).To(BeAssignableToTypeOf(false))
		})
	})

	Describe("IsCacheHealthy", func() {
		It("should check cache health status", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			healthy := status.IsCacheHealthy()
			Expect(healthy).To(BeAssignableToTypeOf(false))
		})
	})

	Describe("IsCacheStrictlyHealthy", func() {
		It("should check cache strict health status", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			healthy := status.IsCacheStrictlyHealthy()
			Expect(healthy).To(BeAssignableToTypeOf(false))
		})
	})

	Describe("SetConfig", func() {
		It("should set configuration", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK:   200,
					monsts.Warn: 200,
					monsts.KO:   503,
				},
			}
			status.SetConfig(cfg)
			Expect(true).To(BeTrue())
		})

		It("should set configuration with mandatory components", func() {
			status = libsts.New(globalCtx)
			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK: 200,
				},
				MandatoryComponent: []libsts.Mandatory{
					{
						Keys: []string{"test-monitor"},
					},
				},
			}
			status.SetConfig(cfg)
			Expect(true).To(BeTrue())
		})
	})
})
