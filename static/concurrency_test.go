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
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrency", func() {
	Describe("Concurrent File Access", func() {
		Context("when accessing files concurrently", func() {
			It("should handle concurrent Has calls", func() {
				handler := newTestStatic()
				h := handler.(interface{ Has(string) bool })

				var wg sync.WaitGroup
				errCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						if !h.Has("testdata/test.txt") {
							errCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(errCount.Load()).To(Equal(int32(0)))
			})

			It("should handle concurrent Find calls", func() {
				handler := newTestStatic()
				h := handler.(staticFind)

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						r, err := h.Find("testdata/test.txt")
						if err == nil && r != nil {
							_, _ = io.Copy(io.Discard, r)
							_ = r.Close()
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(50)))
			})

			It("should handle concurrent Info calls", func() {
				handler := newTestStatic()
				h := handler.(staticInfo)

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						info, err := h.Info("testdata/test.txt")
						if err == nil && info != nil {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(50)))
			})

			It("should handle concurrent List calls", func() {
				handler := newTestStatic()
				h := handler.(staticList)

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						files, err := h.List("testdata")
						if err == nil && len(files) > 0 {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(50)))
			})

			It("should handle concurrent Temp calls", func() {
				handler := newTestStatic()
				h := handler.(staticTemp)

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 20; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						tmp, err := h.Temp("testdata/test.txt")
						if err == nil && tmp != nil {
							_, _ = io.Copy(io.Discard, tmp)
							_ = tmp.Close()
							if namer, ok := tmp.(interface{ Name() string }); ok {
								_ = os.Remove(namer.Name())
							}
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(20)))
			})
		})
	})

	Describe("Concurrent Configuration Access", func() {
		Context("when modifying configurations concurrently", func() {
			It("should handle concurrent SetDownload and IsDownload", func() {
				handler := newTestStatic().(static.Static)

				var wg sync.WaitGroup
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						if idx%2 == 0 {
							handler.SetDownload("testdata/test.txt", true)
						} else {
							_ = handler.IsDownload("testdata/test.txt")
						}
					}(i)
				}

				wg.Wait()
				// Should not panic or race
			})

			It("should handle concurrent SetIndex and GetIndex", func() {
				handler := newTestStatic().(static.Static)

				var wg sync.WaitGroup
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						if idx%2 == 0 {
							handler.SetIndex("", "/", "testdata/index.html")
						} else {
							_ = handler.GetIndex("", "/")
						}
					}(i)
				}

				wg.Wait()
				// Should not panic or race
			})

			It("should handle concurrent SetRedirect and GetRedirect", func() {
				handler := newTestStatic().(static.Static)

				var wg sync.WaitGroup
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						if idx%2 == 0 {
							handler.SetRedirect("", "/old", "", "/new")
						} else {
							_ = handler.GetRedirect("", "/old")
						}
					}(i)
				}

				wg.Wait()
				// Should not panic or race
			})

			It("should handle concurrent SetSpecific and GetSpecific", func() {
				handler := newTestStatic().(static.Static)

				var wg sync.WaitGroup
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						if idx%2 == 0 {
							handler.SetSpecific("", "/custom", customMiddlewareOK("custom", nil))
						} else {
							_ = handler.GetSpecific("", "/custom")
						}
					}(i)
				}

				wg.Wait()
				// Should not panic or race
			})
		})

		Context("when mixing different configurations", func() {
			It("should handle mixed concurrent operations", func() {
				handler := newTestStatic().(static.Static)

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						switch idx % 4 {
						case 0:
							handler.SetDownload("testdata/test.txt", true)
							_ = handler.IsDownload("testdata/test.txt")
						case 1:
							handler.SetIndex("", "/", "testdata/index.html")
							_ = handler.GetIndex("", "/")
						case 2:
							handler.SetRedirect("", "/old", "", "/new")
							_ = handler.GetRedirect("", "/old")
						case 3:
							handler.SetSpecific("", "/custom", customMiddlewareOK("custom", nil))
							_ = handler.GetSpecific("", "/custom")
						}
					}(i)
				}

				wg.Wait()
				// Should not panic or race
			})
		})
	})

	Describe("Concurrent HTTP Requests", func() {
		Context("when serving files concurrently", func() {
			It("should handle concurrent HTTP requests", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						w := performRequest(engine, "GET", "/static/test.txt")
						if w.Code == http.StatusOK {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(100)))
			})

			It("should handle concurrent requests for different files", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				files := []string{
					"/static/test.txt",
					"/static/test.json",
					"/static/index.html",
					"/static/subdir/nested.txt",
					"/static/assets/style.css",
				}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						file := files[idx%len(files)]
						w := performRequest(engine, "GET", file)
						if w.Code == http.StatusOK {
							successCount.Add(1)
						}
					}(i)
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(50)))
			})

			It("should handle concurrent requests with middleware", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static", testMiddleware)

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						w := performRequest(engine, "GET", "/static/test.txt")
						if w.Code == http.StatusOK && w.Header().Get("X-Test-Middleware") == "true" {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(50)))
			})
		})

		Context("when serving with special configurations", func() {
			It("should handle concurrent requests with redirects", func() {
				handler := newTestStatic().(static.Static)
				handler.SetRedirect("", "/static/old", "", "/static/test.txt")
				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				redirectCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						w := performRequest(engine, "GET", "/static/old")
						if w.Code == http.StatusPermanentRedirect {
							redirectCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(redirectCount.Load()).To(Equal(int32(50)))
			})

			It("should handle concurrent requests with specific handlers", func() {
				handler := newTestStatic().(static.Static)

				callCount := atomic.Int32{}
				handler.SetSpecific("", "/static/custom", customMiddlewareOK("custom", func() {
					callCount.Add(1)
				}))

				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						w := performRequest(engine, "GET", "/static/custom")
						if w.Code == http.StatusOK {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(50)))
				Expect(callCount.Load()).To(Equal(int32(50)))
			})

			It("should handle concurrent requests with download flag", func() {
				handler := newTestStatic().(static.Static)
				handler.SetDownload("testdata/test.txt", true)
				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				downloadCount := atomic.Int32{}

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						w := performRequest(engine, "GET", "/static/test.txt")
						if w.Code == http.StatusOK && w.Header().Get("Content-Disposition") != "" {
							downloadCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(downloadCount.Load()).To(Equal(int32(50)))
			})
		})
	})

	Describe("Concurrent Map Operations", func() {
		Context("when mapping over files concurrently", func() {
			It("should handle concurrent Map calls", func() {
				handler := newTestStatic()
				h := handler.(staticMap)

				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				for i := 0; i < 20; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						err := h.Map(func(pathFile string, inf os.FileInfo) error {
							return nil
						})
						if err == nil {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(20)))
			})
		})
	})

	Describe("Stress Testing", func() {
		Context("when under heavy load", func() {
			It("should handle high concurrent file access", func() {
				handler := newTestStatic()
				h := handler.(staticFind)

				var wg sync.WaitGroup
				successCount := atomic.Int32{}
				iterations := 200

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						files := []string{
							"testdata/test.txt",
							"testdata/test.json",
							"testdata/index.html",
							"testdata/large.txt",
						}
						file := files[idx%len(files)]

						r, err := h.Find(file)
						if err == nil && r != nil {
							_, _ = io.Copy(io.Discard, r)
							_ = r.Close()
							successCount.Add(1)
						}
					}(i)
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(iterations)))
			})

			It("should handle high concurrent HTTP requests", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				successCount := atomic.Int32{}
				iterations := 200

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						files := []string{
							"/static/test.txt",
							"/static/test.json",
							"/static/index.html",
							"/static/subdir/nested.txt",
						}
						file := files[idx%len(files)]

						w := performRequest(engine, "GET", file)
						if w.Code == http.StatusOK {
							successCount.Add(1)
						}
					}(i)
				}

				wg.Wait()
				Expect(successCount.Load()).To(Equal(int32(iterations)))
			})

			It("should handle mixed concurrent operations under stress", func() {
				handler := newTestStatic().(static.Static)
				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				iterations := 200

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer GinkgoRecover()
						defer wg.Done()

						switch idx % 5 {
						case 0:
							handler.SetDownload("testdata/test.txt", idx%2 == 0)
						case 1:
							handler.SetIndex("", "/", "testdata/index.html")
						case 2:
							handler.SetRedirect("", "/old", "", "/new")
						case 3:
							_ = handler.Has("testdata/test.txt")
						case 4:
							w := performRequest(engine, "GET", "/static/test.txt")
							_ = w.Code
						}
					}(i)
				}

				wg.Wait()
				// Should complete without panic or race
			})
		})
	})

	Describe("Race Condition Prevention", func() {
		Context("when testing for race conditions", func() {
			It("should not race on configuration updates", func() {
				handler := newTestStatic().(static.Static)

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(2)

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						handler.SetDownload("testdata/test.txt", true)
					}()

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						_ = handler.IsDownload("testdata/test.txt")
					}()
				}

				wg.Wait()
				// Should complete without race detector warnings
			})

			It("should not race on file access", func() {
				handler := newTestStatic()
				h := handler.(staticFindHas)

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(2)

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						_ = h.Has("testdata/test.txt")
					}()

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						r, _ := h.Find("testdata/test.txt")
						if r != nil {
							_ = r.Close()
						}
					}()
				}

				wg.Wait()
				// Should complete without race detector warnings
			})

			It("should not race on HTTP serving", func() {
				handler := newTestStatic().(static.Static)
				engine := setupTestRouter(handler, "/static")

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(2)

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						handler.SetDownload("testdata/test.txt", true)
					}()

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						_ = performRequest(engine, "GET", "/static/test.txt")
					}()
				}

				wg.Wait()
				// Should complete without race detector warnings
			})
		})
	})
})
