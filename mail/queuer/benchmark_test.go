/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package queuer_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nabbar/golib/mail/queuer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Benchmarks", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Counter Performance", func() {
		It("should measure reset performance", func() {
			experiment := NewExperiment("Counter Reset")
			AddReportEntry(experiment.Name, experiment)

			cfg := &queuer.Config{
				Max:  100,
				Wait: 10 * time.Millisecond,
			}
			pooler := queuer.New(cfg, nil)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("reset", func() {
					_ = pooler.Reset()
				})
			}, SamplingConfig{N: 100, Duration: 5 * time.Second})
		})

		It("should measure clone performance", func() {
			experiment := NewExperiment("Counter Clone")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  100,
				Wait: 10 * time.Millisecond,
			}
			pooler := queuer.New(cfg, cli)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("clone", func() {
					_ = pooler.NewPooler()
				})
			}, SamplingConfig{N: 100, Duration: 5 * time.Second})
		})
	})

	Describe("Pooler Send Performance", func() {
		It("should measure send throughput without throttling", func() {
			experiment := NewExperiment("Send Without Throttle")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  0,
				Wait: 0,
			}
			pooler := queuer.New(cfg, cli)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("send", func() {
					_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test"))
				})
			}, SamplingConfig{N: 50, Duration: 10 * time.Second})
		})

		It("should measure send throughput with throttling", func() {
			experiment := NewExperiment("Send With Throttle")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}
			pooler := queuer.New(cfg, cli)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("send", func() {
					_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test"))
				})
			}, SamplingConfig{N: 20, Duration: 10 * time.Second})
		})

		It("should measure send with various message sizes", func() {
			experiment := NewExperiment("Send Message Sizes")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  0,
				Wait: 0,
			}
			pooler := queuer.New(cfg, cli)

			sizes := []int{100, 1024, 10240, 102400}

			for _, size := range sizes {
				content := make([]byte, size)
				for i := range content {
					content[i] = byte('A' + (i % 26))
				}
				message := newSimpleMessage(string(content))

				experiment.RecordValue(fmt.Sprintf("size-%d", size), float64(size))
				experiment.MeasureDuration(fmt.Sprintf("send-%d-bytes", size), func() {
					_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, message)
				})
			}
		})
	})

	Describe("Concurrent Performance", func() {
		It("should measure concurrent send throughput", func() {
			experiment := NewExperiment("Concurrent Send")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  50,
				Wait: 50 * time.Millisecond,
			}
			pooler := queuer.New(cfg, cli)

			numGoroutines := []int{1, 5, 10, 20}

			for _, n := range numGoroutines {
				messagesPerRoutine := 10

				experiment.MeasureDuration(fmt.Sprintf("concurrent-%d-goroutines", n), func() {
					var wg sync.WaitGroup
					wg.Add(n)

					for i := 0; i < n; i++ {
						go func(routineID int) {
							defer wg.Done()
							for j := 0; j < messagesPerRoutine; j++ {
								_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d-%d", routineID, j)))
							}
						}(i)
					}

					wg.Wait()
				})

				experiment.RecordValue(fmt.Sprintf("goroutines-%d", n), float64(n))
				experiment.RecordValue(fmt.Sprintf("total-messages-%d", n), float64(n*messagesPerRoutine))
			}
		})

		It("should measure clone performance under concurrent load", func() {
			experiment := NewExperiment("Concurrent Clone")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  50,
				Wait: 50 * time.Millisecond,
			}
			pooler := queuer.New(cfg, cli)

			numGoroutines := []int{1, 5, 10, 20}

			for _, n := range numGoroutines {
				experiment.MeasureDuration(fmt.Sprintf("clone-concurrent-%d", n), func() {
					var wg sync.WaitGroup
					wg.Add(n)

					for i := 0; i < n; i++ {
						go func() {
							defer wg.Done()
							_ = pooler.NewPooler()
						}()
					}

					wg.Wait()
				})

				experiment.RecordValue(fmt.Sprintf("clone-goroutines-%d", n), float64(n))
			}
		})

		It("should measure reset performance under concurrent load", func() {
			experiment := NewExperiment("Concurrent Reset")
			AddReportEntry(experiment.Name, experiment)

			cfg := &queuer.Config{
				Max:  50,
				Wait: 50 * time.Millisecond,
			}
			pooler := queuer.New(cfg, nil)

			numGoroutines := []int{1, 5, 10, 20}

			for _, n := range numGoroutines {
				experiment.MeasureDuration(fmt.Sprintf("reset-concurrent-%d", n), func() {
					var wg sync.WaitGroup
					wg.Add(n)

					for i := 0; i < n; i++ {
						go func() {
							defer wg.Done()
							_ = pooler.Reset()
						}()
					}

					wg.Wait()
				})

				experiment.RecordValue(fmt.Sprintf("reset-goroutines-%d", n), float64(n))
			}
		})
	})

	Describe("Throttle Efficiency", func() {
		It("should measure throttle overhead", func() {
			experiment := NewExperiment("Throttle Overhead")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)

			// Without throttle
			cfgNoThrottle := &queuer.Config{
				Max:  0,
				Wait: 0,
			}
			poolerNoThrottle := queuer.New(cfgNoThrottle, cli)

			var timeNoThrottle time.Duration
			experiment.MeasureDuration("no-throttle", func() {
				start := time.Now()
				for i := 0; i < 10; i++ {
					_ = poolerNoThrottle.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test"))
				}
				timeNoThrottle = time.Since(start)
			})

			// With throttle
			cfgWithThrottle := &queuer.Config{
				Max:  100,
				Wait: 100 * time.Millisecond,
			}
			poolerWithThrottle := queuer.New(cfgWithThrottle, cli)

			var timeWithThrottle time.Duration
			experiment.MeasureDuration("with-throttle", func() {
				start := time.Now()
				for i := 0; i < 10; i++ {
					_ = poolerWithThrottle.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test"))
				}
				timeWithThrottle = time.Since(start)
			})

			overhead := timeWithThrottle - timeNoThrottle
			experiment.RecordValue("overhead-ms", float64(overhead.Milliseconds()))
		})

		It("should measure throttle accuracy", func() {
			experiment := NewExperiment("Throttle Accuracy")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)

			testCases := []struct {
				max  int
				wait time.Duration
			}{
				{max: 5, wait: 100 * time.Millisecond},
				{max: 10, wait: 100 * time.Millisecond},
				{max: 5, wait: 200 * time.Millisecond},
			}

			for _, tc := range testCases {
				cfg := &queuer.Config{
					Max:  tc.max,
					Wait: tc.wait,
				}
				pooler := queuer.New(cfg, cli)

				numMessages := tc.max * 3

				experiment.MeasureDuration(fmt.Sprintf("throttle-max%d-wait%dms", tc.max, tc.wait.Milliseconds()), func() {
					for i := 0; i < numMessages; i++ {
						_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d", i)))
					}
				})

				experiment.RecordValue(fmt.Sprintf("max-%d", tc.max), float64(tc.max))
				experiment.RecordValue(fmt.Sprintf("wait-ms-%d", tc.max), float64(tc.wait.Milliseconds()))
			}
		})
	})

	Describe("Memory Allocation", func() {
		It("should measure clone allocation", func() {
			experiment := NewExperiment("Clone Memory")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  100,
				Wait: 10 * time.Millisecond,
			}
			pooler := queuer.New(cfg, cli)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("clone-operation", func() {
					clones := make([]queuer.Pooler, 100)
					for i := 0; i < 100; i++ {
						clones[i] = pooler.NewPooler()
					}
					// Keep clones in scope to prevent GC
					_ = clones
				})
			}, SamplingConfig{N: 10, Duration: 5 * time.Second})
		})

		It("should measure pooler instantiation", func() {
			experiment := NewExperiment("Pooler Instantiation")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("new-pooler", func() {
					cfg := &queuer.Config{
						Max:  100,
						Wait: 10 * time.Millisecond,
					}
					_ = queuer.New(cfg, cli)
				})
			}, SamplingConfig{N: 100, Duration: 5 * time.Second})
		})
	})

	Describe("Scalability", func() {
		It("should measure throughput scaling with goroutines", func() {
			experiment := NewExperiment("Throughput Scaling")
			AddReportEntry(experiment.Name, experiment)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  100,
				Wait: 50 * time.Millisecond,
			}
			pooler := queuer.New(cfg, cli)

			goroutineCounts := []int{1, 2, 4, 8, 16, 32}
			messagesPerRoutine := 10

			for _, numGoroutines := range goroutineCounts {
				var totalTime time.Duration

				experiment.MeasureDuration(fmt.Sprintf("scale-%d-goroutines", numGoroutines), func() {
					start := time.Now()

					var wg sync.WaitGroup
					wg.Add(numGoroutines)

					for i := 0; i < numGoroutines; i++ {
						go func(routineID int) {
							defer wg.Done()
							for j := 0; j < messagesPerRoutine; j++ {
								_ = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg-%d-%d", routineID, j)))
							}
						}(i)
					}

					wg.Wait()
					totalTime = time.Since(start)
				})

				totalMessages := numGoroutines * messagesPerRoutine
				throughput := float64(totalMessages) / totalTime.Seconds()

				experiment.RecordValue(fmt.Sprintf("throughput-%d-goroutines", numGoroutines), throughput)
				experiment.RecordValue(fmt.Sprintf("messages-%d-goroutines", numGoroutines), float64(totalMessages))
			}
		})
	})
})
