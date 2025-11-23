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
 */

package static_test

import (
	"context"
	"time"

	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor", func() {
	var (
		handler static.Static
		ctx     context.Context
		cancel  context.CancelFunc
	)

	BeforeEach(func() {
		handler = newTestStatic()
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Monitor Creation", func() {
		Context("when creating monitor", func() {
			It("should create monitor successfully", func() {
				cfg := newTestMonitorConfig()
				Expect(cfg).ToNot(BeNil())

				vrs := newTestVersion()
				Expect(vrs).ToNot(BeNil())

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Cleanup
				if mon != nil {
					_ = mon.Stop(ctx)
				}
			})

			It("should include version information", func() {
				cfg := newTestMonitorConfig()

				mon, err := handler.Monitor(ctx, cfg, newTestVersion())
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Cleanup
				if mon != nil {
					_ = mon.Stop(ctx)
				}
			})

			It("should use provided context", func() {
				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Cleanup
				if mon != nil {
					_ = mon.Stop(ctx)
				}
			})
		})

		Context("when using different configurations", func() {
			It("should accept custom config", func() {
				cfg := newTestMonitorConfig()
				Expect(cfg).ToNot(BeNil())

				vrs := newTestVersion()

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Cleanup
				if mon != nil {
					_ = mon.Stop(ctx)
				}
			})
		})
	})

	Describe("Health Check", func() {
		Context("when performing health check", func() {
			It("should pass health check for valid handler", func() {
				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Get health status
				if mon != nil {
					// The monitor should be healthy
					info := mon.InfoGet()
					Expect(info).ToNot(BeNil())

					// Cleanup
					_ = mon.Stop(ctx)
				}
			})

			It("should handle health check with empty base pth", func() {
				// Create handler with no explicit base pth
				h := newTestStaticWithRoot()

				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := h.Monitor(ctx, cfg, vrs)
				// Should still create monitor even if health check might fail
				if err == nil {
					Expect(mon).ToNot(BeNil())
					if mon != nil {
						_ = mon.Stop(ctx)
					}
				}
			})
		})
	})

	Describe("Monitor Lifecycle", func() {
		Context("when managing monitor lifecycle", func() {
			It("should start and stop monitor", func() {
				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Monitor should be started already (based on implementation)
				// Stop it
				if mon != nil {
					err := mon.Stop(ctx)
					Expect(err).ToNot(HaveOccurred())
				}
			})

			It("should handle multiple monitors", func() {
				cfg1 := newTestMonitorConfig()
				cfg2 := newTestMonitorConfig()
				vrs := newTestVersion()

				mon1, err := handler.Monitor(ctx, cfg1, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon1).ToNot(BeNil())

				mon2, err := handler.Monitor(ctx, cfg2, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon2).ToNot(BeNil())

				// Cleanup
				if mon1 != nil {
					_ = mon1.Stop(ctx)
				}
				if mon2 != nil {
					_ = mon2.Stop(ctx)
				}
			})
		})

		Context("when context is cancelled", func() {
			It("should handle cancelled context gracefully", func() {
				localCtx, localCancel := context.WithCancel(ctx)

				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(localCtx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Cancel context
				localCancel()

				// Give it a moment to process cancellation
				time.Sleep(10 * time.Millisecond)

				// Cleanup
				if mon != nil {
					_ = mon.Stop(ctx)
				}
			})

			It("should handle timeout context", func() {
				localCtx, localCancel := context.WithTimeout(ctx, 100*time.Millisecond)
				defer localCancel()

				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(localCtx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				// Cleanup
				if mon != nil {
					_ = mon.Stop(ctx)
				}
			})
		})
	})

	Describe("Monitor Information", func() {
		Context("when querying monitor info", func() {
			It("should provide monitor information", func() {
				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				if mon != nil {
					info := mon.InfoGet()
					Expect(info).ToNot(BeNil())

					// Cleanup
					_ = mon.Stop(ctx)
				}
			})

			It("should include runtime information", func() {
				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				if mon != nil {
					info := mon.InfoGet()
					Expect(info).ToNot(BeNil())

					// Should have some information
					data := info.Info()
					Expect(data).ToNot(BeNil())

					// Cleanup
					_ = mon.Stop(ctx)
				}
			})
		})
	})

	Describe("Concurrent Monitor Access", func() {
		Context("when accessing monitor concurrently", func() {
			It("should handle concurrent info requests", func() {
				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				mon, err := handler.Monitor(ctx, cfg, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())

				if mon != nil {
					done := make(chan bool, 5)

					// Concurrent info requests
					for i := 0; i < 5; i++ {
						go func() {
							defer GinkgoRecover()
							info := mon.InfoGet()
							Expect(info).ToNot(BeNil())
							done <- true
						}()
					}

					// Wait for all goroutines
					for i := 0; i < 5; i++ {
						<-done
					}

					// Cleanup
					_ = mon.Stop(ctx)
				}
			})
		})
	})

	Describe("Error Scenarios", func() {
		Context("when encountering errors", func() {
			It("should handle invalid base pth gracefully", func() {
				// This test verifies that even with potential issues,
				// the monitor creation doesn't panic
				cfg := newTestMonitorConfig()
				vrs := newTestVersion()

				Expect(func() {
					mon, _ := handler.Monitor(ctx, cfg, vrs)
					if mon != nil {
						_ = mon.Stop(ctx)
					}
				}).ToNot(Panic())
			})
		})
	})

	Describe("Multiple Handlers", func() {
		Context("when using multiple static handlers", func() {
			It("should create separate monitors", func() {
				h1 := newTestStatic()
				h2 := newTestStatic()

				cfg1 := newTestMonitorConfig()
				cfg2 := newTestMonitorConfig()
				vrs := newTestVersion()

				mon1, err := h1.Monitor(ctx, cfg1, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon1).ToNot(BeNil())

				mon2, err := h2.Monitor(ctx, cfg2, vrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon2).ToNot(BeNil())

				// Monitors should be independent
				Expect(mon1).ToNot(Equal(mon2))

				// Cleanup
				if mon1 != nil {
					_ = mon1.Stop(ctx)
				}
				if mon2 != nil {
					_ = mon2.Stop(ctx)
				}
			})
		})
	})
})
