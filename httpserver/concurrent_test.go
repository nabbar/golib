/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package httpserver_test

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/nabbar/golib/httpserver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-CC] HTTPServer/Concurrent", func() {
	Describe("Concurrent configuration updates", func() {
		It("[TC-CC-001] should handle concurrent SetConfig calls", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "concurrent-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					newPort := GetFreePort()
					newCfg := httpserver.Config{
						Name:   fmt.Sprintf("concurrent-test-%d", index),
						Listen: fmt.Sprintf("127.0.0.1:%d", newPort),
						Expose: fmt.Sprintf("http://127.0.0.1:%d", newPort),
					}
					newCfg.RegisterHandlerFunc(func() map[string]http.Handler {
						return map[string]http.Handler{"": http.NotFoundHandler()}
					})

					_ = srv.SetConfig(newCfg, nil)
				}(i)
			}

			wg.Wait()
		})

		It("[TC-CC-002] should handle concurrent GetConfig calls", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "concurrent-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					c := srv.GetConfig()
					Expect(c).ToNot(BeNil())
				}()
			}

			wg.Wait()
		})
	})

	Describe("Concurrent info reads", func() {
		It("[TC-CC-003] should handle concurrent GetName calls", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "concurrent-info-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			results := make([]string, 100)

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()
					results[index] = srv.GetName()
				}(i)
			}

			wg.Wait()

			// All results should be consistent
			for _, result := range results {
				Expect(result).To(Equal(results[0]))
			}
		})

		It("[TC-CC-004] should handle concurrent info method calls", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "concurrent-info-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_ = srv.GetName()
					_ = srv.GetBindable()
					_ = srv.GetExpose()
					_ = srv.IsDisable()
					_ = srv.IsTLS()
				}()
			}

			wg.Wait()
		})
	})

	Describe("Concurrent handler operations", func() {
		It("[TC-CC-005] should handle concurrent Handler calls", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:       "handler-test",
				Listen:     fmt.Sprintf("127.0.0.1:%d", port),
				Expose:     fmt.Sprintf("http://127.0.0.1:%d", port),
				HandlerKey: "test",
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{
					"test": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					srv.Handler(func() map[string]http.Handler {
						return map[string]http.Handler{
							"test": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								w.WriteHeader(http.StatusOK)
							}),
						}
					})
				}(i)
			}

			wg.Wait()
		})
	})

	Describe("Concurrent IsRunning checks", func() {
		It("[TC-CC-006] should handle concurrent IsRunning calls", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "running-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					_ = srv.IsRunning()
				}()
			}

			wg.Wait()
		})
	})

	Describe("Concurrent Merge operations", func() {
		It("[TC-CC-007] should handle concurrent Merge calls", func() {
			port1 := GetFreePort()
			cfg1 := httpserver.Config{
				Name:   "merge-test-1",
				Listen: fmt.Sprintf("127.0.0.1:%d", port1),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port1),
			}

			cfg1.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv1, err := httpserver.New(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					port2 := GetFreePort()
					cfg2 := httpserver.Config{
						Name:   fmt.Sprintf("merge-test-2-%d", index),
						Listen: fmt.Sprintf("127.0.0.1:%d", port2),
						Expose: fmt.Sprintf("http://127.0.0.1:%d", port2),
					}
					cfg2.RegisterHandlerFunc(func() map[string]http.Handler {
						return map[string]http.Handler{"": http.NotFoundHandler()}
					})

					srv2, e := httpserver.New(cfg2, nil)
					if e == nil {
						_ = srv1.Merge(srv2, nil)
					}
				}(i)
			}

			wg.Wait()
		})
	})

	Describe("Concurrent MonitorName calls", func() {
		It("[TC-CC-008] should handle concurrent MonitorName calls", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "monitor-name-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			results := make([]string, 100)

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()
					results[index] = srv.MonitorName()
				}(i)
			}

			wg.Wait()

			// All results should be consistent
			for _, result := range results {
				Expect(result).To(Equal(results[0]))
			}
		})
	})
})
