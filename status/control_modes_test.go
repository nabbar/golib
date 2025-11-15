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
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"
	stsctr "github.com/nabbar/golib/status/control"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status/ControlModes", func() {
	var (
		status libsts.Status
		pool   monpol.Pool
	)

	Describe("Must mode", func() {
		BeforeEach(func() {
			status = libsts.New(globalCtx)
			status.SetInfo("control-test", "v1.0.0", "abc123")

			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"critical-db"},
					},
				},
			}
			status.SetConfig(cfg)
		})

		Context("when component is healthy", func() {
			BeforeEach(func() {
				m := newHealthyMonitor("critical-db")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())

				// Wait for health check
				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report overall healthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})

			It("should report strictly healthy status", func() {
				strictlyHealthy := status.IsStrictlyHealthy()
				Expect(strictlyHealthy).To(BeTrue())
			})
		})

		Context("when component is unhealthy", func() {
			BeforeEach(func() {
				m := newUnhealthyMonitor("critical-db")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())

				// Wait for health check to fail
				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report overall unhealthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeFalse())
			})

			It("should report strictly unhealthy status", func() {
				strictlyHealthy := status.IsStrictlyHealthy()
				Expect(strictlyHealthy).To(BeFalse())
			})
		})

		Context("when component is missing", func() {
			It("should report unhealthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue()) // No monitors means OK by default
			})
		})
	})

	Describe("Should mode", func() {
		BeforeEach(func() {
			status = libsts.New(globalCtx)
			status.SetInfo("control-test", "v1.0.0", "abc123")

			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Should,
						Keys: []string{"optional-cache"},
					},
				},
			}
			status.SetConfig(cfg)
		})

		Context("when component is healthy", func() {
			BeforeEach(func() {
				m := newHealthyMonitor("optional-cache")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report healthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})
		})

		Context("when component is unhealthy", func() {
			BeforeEach(func() {
				m := newUnhealthyMonitor("optional-cache")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should still report as healthy (warn level)", func() {
				// Should mode downgrades to Warn, not KO
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue()) // >= Warn is considered healthy
			})

			It("should not be strictly healthy", func() {
				strictlyHealthy := status.IsStrictlyHealthy()
				Expect(strictlyHealthy).To(BeFalse())
			})
		})
	})

	Describe("AnyOf mode", func() {
		BeforeEach(func() {
			status = libsts.New(globalCtx)
			status.SetInfo("control-test", "v1.0.0", "abc123")

			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.AnyOf,
						Keys: []string{"db-primary", "db-secondary", "db-tertiary"},
					},
				},
			}
			status.SetConfig(cfg)
		})

		Context("when all components are healthy", func() {
			BeforeEach(func() {
				for _, name := range []string{"db-primary", "db-secondary", "db-tertiary"} {
					m := newHealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}
				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report healthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})
		})

		Context("when only one component is healthy", func() {
			BeforeEach(func() {
				// One healthy
				m1 := newHealthyMonitor("db-primary")
				err := pool.MonitorAdd(m1)
				Expect(err).ToNot(HaveOccurred())

				// Two unhealthy
				m2 := newUnhealthyMonitor("db-secondary")
				err = pool.MonitorAdd(m2)
				Expect(err).ToNot(HaveOccurred())

				m3 := newUnhealthyMonitor("db-tertiary")
				err = pool.MonitorAdd(m3)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report healthy status (at least one is OK)", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})
		})

		Context("when all components are unhealthy", func() {
			BeforeEach(func() {
				for _, name := range []string{"db-primary", "db-secondary", "db-tertiary"} {
					m := newUnhealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}
				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report unhealthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeFalse())
			})
		})

		Context("when no components are registered", func() {
			It("should report default status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue()) // No monitors = OK
			})
		})
	})

	Describe("Quorum mode", func() {
		BeforeEach(func() {
			status = libsts.New(globalCtx)
			status.SetInfo("control-test", "v1.0.0", "abc123")

			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Quorum,
						Keys: []string{"node-1", "node-2", "node-3", "node-4", "node-5"},
					},
				},
			}
			status.SetConfig(cfg)
		})

		Context("when majority (3/5) are healthy", func() {
			BeforeEach(func() {
				// 3 healthy
				for _, name := range []string{"node-1", "node-2", "node-3"} {
					m := newHealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}

				// 2 unhealthy
				for _, name := range []string{"node-4", "node-5"} {
					m := newUnhealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report healthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})
		})

		Context("when minority (2/5) are healthy", func() {
			BeforeEach(func() {
				// 2 healthy
				for _, name := range []string{"node-1", "node-2"} {
					m := newHealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}

				// 3 unhealthy
				for _, name := range []string{"node-3", "node-4", "node-5"} {
					m := newUnhealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report unhealthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeFalse())
			})
		})

		Context("when exactly half (2/4) are healthy", func() {
			BeforeEach(func() {
				cfg := libsts.Config{
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Quorum,
							Keys: []string{"node-1", "node-2", "node-3", "node-4"},
						},
					},
				}
				status.SetConfig(cfg)

				// 2 healthy
				for _, name := range []string{"node-1", "node-2"} {
					m := newHealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}

				// 2 unhealthy
				for _, name := range []string{"node-3", "node-4"} {
					m := newUnhealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should require >50% for quorum", func() {
				healthy := status.IsHealthy()
				// 50% is not enough, need >50%
				Expect(healthy).To(BeFalse())
			})
		})
	})

	Describe("Ignore mode", func() {
		BeforeEach(func() {
			status = libsts.New(globalCtx)
			status.SetInfo("control-test", "v1.0.0", "abc123")

			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Ignore,
						Keys: []string{"ignored-service"},
					},
				},
			}
			status.SetConfig(cfg)
		})

		Context("when component is unhealthy", func() {
			BeforeEach(func() {
				m := newUnhealthyMonitor("ignored-service")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should not affect overall health", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue()) // Ignored components don't affect status
			})
		})
	})

	Describe("Mixed modes", func() {
		BeforeEach(func() {
			status = libsts.New(globalCtx)
			status.SetInfo("control-test", "v1.0.0", "abc123")

			pool = newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"critical-db"},
					},
					{
						Mode: stsctr.Should,
						Keys: []string{"cache"},
					},
					{
						Mode: stsctr.AnyOf,
						Keys: []string{"queue-1", "queue-2"},
					},
				},
			}
			status.SetConfig(cfg)
		})

		Context("when all requirements are met", func() {
			BeforeEach(func() {
				// Must: healthy
				m1 := newHealthyMonitor("critical-db")
				err := pool.MonitorAdd(m1)
				Expect(err).ToNot(HaveOccurred())

				// Should: healthy
				m2 := newHealthyMonitor("cache")
				err = pool.MonitorAdd(m2)
				Expect(err).ToNot(HaveOccurred())

				// AnyOf: one healthy
				m3 := newHealthyMonitor("queue-1")
				err = pool.MonitorAdd(m3)
				Expect(err).ToNot(HaveOccurred())

				m4 := newUnhealthyMonitor("queue-2")
				err = pool.MonitorAdd(m4)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report healthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})
		})

		Context("when Must component fails", func() {
			BeforeEach(func() {
				// Must: unhealthy - this should fail everything
				m1 := newUnhealthyMonitor("critical-db")
				err := pool.MonitorAdd(m1)
				Expect(err).ToNot(HaveOccurred())

				// Should: healthy
				m2 := newHealthyMonitor("cache")
				err = pool.MonitorAdd(m2)
				Expect(err).ToNot(HaveOccurred())

				// AnyOf: healthy
				m3 := newHealthyMonitor("queue-1")
				err = pool.MonitorAdd(m3)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should report unhealthy status", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeFalse())
			})
		})

		Context("when only Should component fails", func() {
			BeforeEach(func() {
				// Must: healthy
				m1 := newHealthyMonitor("critical-db")
				err := pool.MonitorAdd(m1)
				Expect(err).ToNot(HaveOccurred())

				// Should: unhealthy - should only warn
				m2 := newUnhealthyMonitor("cache")
				err = pool.MonitorAdd(m2)
				Expect(err).ToNot(HaveOccurred())

				// AnyOf: healthy
				m3 := newHealthyMonitor("queue-1")
				err = pool.MonitorAdd(m3)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(testMonitorStabilizeDelay)
			})

			It("should still be healthy (warn level)", func() {
				healthy := status.IsHealthy()
				Expect(healthy).To(BeTrue())
			})

			It("should not be strictly healthy", func() {
				strictlyHealthy := status.IsStrictlyHealthy()
				Expect(strictlyHealthy).To(BeFalse())
			})
		})
	})
})
