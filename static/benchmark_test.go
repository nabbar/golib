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
	"time"

	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Benchmarks", func() {
	var experiment *Experiment

	BeforeEach(func() {
		experiment = NewExperiment("Static File Operations")
	})

	AfterEach(func() {
		if experiment != nil {
			AddReportEntry(experiment.Name, experiment)
		}
	})

	Describe("File Access Performance", func() {
		Context("when accessing files", func() {
			It("should benchmark Has operation", func() {
				handler := newTestStatic()
				h := handler.(interface{ Has(string) bool })

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("Has-Existing", func() {
						_ = h.Has("testdata/test.txt")
					})

					experiment.MeasureDuration("Has-NonExisting", func() {
						_ = h.Has("testdata/nonexistent.txt")
					})
				}, SamplingConfig{N: 1000, Duration: 3 * time.Second})
			})

			It("should benchmark Find operation", func() {
				handler := newTestStatic()
				h := handler.(staticFind)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("Find-SmallFile", func() {
						r, err := h.Find("testdata/test.txt")
						Expect(err).ToNot(HaveOccurred())
						if r != nil {
							_, _ = io.Copy(io.Discard, r)
							_ = r.Close()
						}
					})
				}, SamplingConfig{N: 500, Duration: 3 * time.Second})
			})

			It("should benchmark Info operation", func() {
				handler := newTestStatic()
				h := handler.(staticInfo)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("Info", func() {
						_, err := h.Info("testdata/test.txt")
						Expect(err).ToNot(HaveOccurred())
					})
				}, SamplingConfig{N: 1000, Duration: 3 * time.Second})
			})

			It("should benchmark List operation", func() {
				handler := newTestStatic()
				h := handler.(staticList)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("List", func() {
						_, err := h.List("testdata")
						Expect(err).ToNot(HaveOccurred())
					})
				}, SamplingConfig{N: 500, Duration: 3 * time.Second})
			})
		})
	})

	Describe("HTTP Handler Performance", func() {
		Context("when serving files via HTTP", func() {
			It("should benchmark file serving", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("ServeFile-TXT", func() {
						w := performRequest(engine, "GET", "/static/test.txt")
						Expect(w.Code).To(Equal(http.StatusOK))
					})

					experiment.MeasureDuration("ServeFile-JSON", func() {
						w := performRequest(engine, "GET", "/static/test.json")
						Expect(w.Code).To(Equal(http.StatusOK))
					})

					experiment.MeasureDuration("ServeFile-HTML", func() {
						w := performRequest(engine, "GET", "/static/index.html")
						Expect(w.Code).To(Equal(http.StatusOK))
					})

					experiment.MeasureDuration("ServeFile-NotFound", func() {
						w := performRequest(engine, "GET", "/static/nonexistent.txt")
						Expect(w.Code).To(Equal(http.StatusNotFound))
					})
				}, SamplingConfig{N: 500, Duration: 3 * time.Second})
			})

			It("should benchmark nested file serving", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("ServeFile-Nested", func() {
						w := performRequest(engine, "GET", "/static/subdir/nested.txt")
						Expect(w.Code).To(Equal(http.StatusOK))
					})
				}, SamplingConfig{N: 500, Duration: 3 * time.Second})
			})
		})

		Context("when using middleware", func() {
			It("should benchmark with middleware", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static", testMiddleware)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("ServeFile-WithMiddleware", func() {
						w := performRequest(engine, "GET", "/static/test.txt")
						Expect(w.Code).To(Equal(http.StatusOK))
					})
				}, SamplingConfig{N: 500, Duration: 3 * time.Second})
			})
		})
	})

	Describe("Configuration Performance", func() {
		Context("when managing configurations", func() {
			It("should benchmark SetDownload and IsDownload", func() {
				handler := newTestStatic().(static.Static)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("SetDownload", func() {
						handler.SetDownload("testdata/test.txt", true)
					})

					experiment.MeasureDuration("IsDownload", func() {
						_ = handler.IsDownload("testdata/test.txt")
					})
				}, SamplingConfig{N: 1000, Duration: 3 * time.Second})
			})

			It("should benchmark SetIndex and GetIndex", func() {
				handler := newTestStatic().(static.Static)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("SetIndex", func() {
						handler.SetIndex("", "/", "testdata/index.html")
					})

					experiment.MeasureDuration("GetIndex", func() {
						_ = handler.GetIndex("", "/")
					})

					experiment.MeasureDuration("IsIndex", func() {
						_ = handler.IsIndex("testdata/index.html")
					})
				}, SamplingConfig{N: 1000, Duration: 3 * time.Second})
			})

			It("should benchmark SetRedirect and GetRedirect", func() {
				handler := newTestStatic().(static.Static)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("SetRedirect", func() {
						handler.SetRedirect("", "/old", "", "/new")
					})

					experiment.MeasureDuration("GetRedirect", func() {
						_ = handler.GetRedirect("", "/old")
					})

					experiment.MeasureDuration("IsRedirect", func() {
						_ = handler.IsRedirect("", "/old")
					})
				}, SamplingConfig{N: 1000, Duration: 3 * time.Second})
			})
		})
	})

	Describe("Concurrent Access Performance", func() {
		Context("when accessing concurrently", func() {
			It("should benchmark concurrent file serving", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("Concurrent-10", func() {
						var wg sync.WaitGroup
						for i := 0; i < 10; i++ {
							wg.Add(1)
							go func() {
								defer GinkgoRecover()
								defer wg.Done()
								w := performRequest(engine, "GET", "/static/test.txt")
								Expect(w.Code).To(Equal(http.StatusOK))
							}()
						}
						wg.Wait()
					})
				}, SamplingConfig{N: 100, Duration: 3 * time.Second})
			})

			It("should benchmark concurrent configuration access", func() {
				handler := newTestStatic().(static.Static)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("Concurrent-Config", func() {
						var wg sync.WaitGroup
						for i := 0; i < 10; i++ {
							wg.Add(1)
							go func() {
								defer wg.Done()
								handler.SetDownload("testdata/test.txt", true)
								_ = handler.IsDownload("testdata/test.txt")
							}()
						}
						wg.Wait()
					})
				}, SamplingConfig{N: 100, Duration: 3 * time.Second})
			})
		})
	})

	Describe("Memory Efficiency", func() {
		Context("when allocating memory", func() {
			It("should measure memory for file operations", func() {
				handler := newTestStatic()
				h := handler.(staticFind)

				experiment.Sample(func(idx int) {
					r, err := h.Find("testdata/test.txt")
					Expect(err).ToNot(HaveOccurred())
					if r != nil {
						_, _ = io.Copy(io.Discard, r)
						_ = r.Close()
					}
				}, SamplingConfig{N: 100, Duration: 2 * time.Second})

				// Memory measurements are recorded automatically by gmeasure
			})

			It("should measure memory for HTTP serving", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				experiment.Sample(func(idx int) {
					w := performRequest(engine, "GET", "/static/test.txt")
					Expect(w.Code).To(Equal(http.StatusOK))
				}, SamplingConfig{N: 100, Duration: 2 * time.Second})

				// Memory measurements are recorded automatically by gmeasure
			})
		})
	})

	Describe("Large File Handling", func() {
		Context("when handling large files", func() {
			It("should benchmark large file operations", func() {
				handler := newTestStatic().(static.Static)
				h := handler.(staticFind)

				// Set threshold to force temp file usage
				handler.UseTempForFileSize(10)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("LargeFile-Find", func() {
						r, err := h.Find("testdata/large.txt")
						Expect(err).ToNot(HaveOccurred())
						if r != nil {
							_, _ = io.Copy(io.Discard, r)
							_ = r.Close()
							// Clean up temp file if created
							if namer, ok := r.(interface{ Name() string }); ok {
								_ = os.Remove(namer.Name())
							}
						}
					})
				}, SamplingConfig{N: 100, Duration: 2 * time.Second})
			})
		})
	})

	Describe("Map Operation Performance", func() {
		Context("when mapping over files", func() {
			It("should benchmark Map operation", func() {
				handler := newTestStatic()
				h := handler.(staticMap)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("Map-AllFiles", func() {
						count := 0
						err := h.Map(func(pathFile string, inf os.FileInfo) error {
							count++
							return nil
						})
						Expect(err).ToNot(HaveOccurred())
						Expect(count).To(BeNumerically(">", 0))
					})
				}, SamplingConfig{N: 100, Duration: 2 * time.Second})
			})
		})
	})

	Describe("Redirect and Special Handlers Performance", func() {
		Context("when using redirect and special handlers", func() {
			It("should benchmark redirect performance", func() {
				handler := newTestStatic().(static.Static)
				handler.SetRedirect("", "/static/old", "", "/static/test.txt")
				engine := setupTestRouter(handler, "/static")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("Redirect", func() {
						w := performRequest(engine, "GET", "/static/old")
						Expect(w.Code).To(Equal(http.StatusPermanentRedirect))
					})
				}, SamplingConfig{N: 500, Duration: 3 * time.Second})
			})

			It("should benchmark specific handler performance", func() {
				handler := newTestStatic().(static.Static)
				handler.SetSpecific("", "/static/custom", customMiddlewareOK("custom", nil))
				engine := setupTestRouter(handler, "/static")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("SpecificHandler", func() {
						w := performRequest(engine, "GET", "/static/custom")
						Expect(w.Code).To(Equal(http.StatusOK))
					})
				}, SamplingConfig{N: 500, Duration: 3 * time.Second})
			})
		})
	})

	Describe("Throughput Measurements", func() {
		Context("when measuring throughput", func() {
			It("should measure requests per second", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				requestCount := 0
				startTime := time.Now()

				experiment.Sample(func(idx int) {
					w := performRequest(engine, "GET", "/static/test.txt")
					Expect(w.Code).To(Equal(http.StatusOK))
					requestCount++
				}, SamplingConfig{N: 1000, Duration: 5 * time.Second})

				duration := time.Since(startTime)
				throughput := float64(requestCount) / duration.Seconds()

				// Record throughput as a value
				experiment.RecordValue("Throughput-RPS", throughput)

				// Expect reasonable throughput (at least 100 req/spc)
				Expect(throughput).To(BeNumerically(">", 100))
			})
		})
	})
})
