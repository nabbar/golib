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

package multi_test

import (
	"bytes"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/ioutils/multi"
)

var _ = Describe("[TC-MD] Multi Mode Operations", func() {
	var (
		m   multi.Multi
		cfg multi.Config
	)

	BeforeEach(func() {
		cfg = multi.DefaultConfig()
		// Lower thresholds for faster testing
		cfg.SampleWrite = 2
		cfg.ThresholdLatency = 100 * 1000 // 100µs
		cfg.MinimalWriter = 2
		cfg.MinimalSize = 10
	})

	Describe("Initialization Modes", func() {
		Context("Forced Sequential", func() {
			It("[TC-MD-001] should initialize in sequential mode", func() {
				m = multi.New(false, false, cfg)
				Expect(m.IsAdaptive()).To(BeFalse())
				Expect(m.IsParallel()).To(BeFalse())
				Expect(m.IsSequential()).To(BeTrue())
			})
		})

		Context("Forced Parallel", func() {
			It("[TC-MD-003] should initialize in parallel mode", func() {
				m = multi.New(false, true, cfg)
				Expect(m.IsAdaptive()).To(BeFalse())
				Expect(m.IsParallel()).To(BeTrue())
				Expect(m.IsSequential()).To(BeFalse())
			})
		})

		Context("Adaptive Sequential", func() {
			It("[TC-MD-001] should initialize in adaptive mode starting sequential", func() {
				m = multi.New(true, false, cfg)
				Expect(m.IsAdaptive()).To(BeTrue())
				Expect(m.IsParallel()).To(BeFalse())
				Expect(m.IsSequential()).To(BeTrue())
			})
		})

		Context("Adaptive Parallel", func() {
			It("[TC-MD-001] should initialize in adaptive mode starting parallel", func() {
				m = multi.New(true, true, cfg)
				Expect(m.IsAdaptive()).To(BeTrue())
				Expect(m.IsParallel()).To(BeTrue())
				Expect(m.IsSequential()).To(BeFalse())
			})
		})
	})

	Describe("Adaptive Behavior", func() {
		Context("Switching from Sequential to Parallel", func() {
			It("[TC-MD-004] should switch to parallel when latency is high and writers are enough", func() {
				m = multi.New(true, false, cfg)

				// Add slow writers
				w1 := &slowWriter{delay: 500 * time.Microsecond}
				w2 := &slowWriter{delay: 500 * time.Microsecond}
				m.AddWriter(w1, w2)

				// Initial state
				Expect(m.IsParallel()).To(BeFalse())

				// Perform writes to trigger sampling
				// SampleWrite is 2, so we need at least 4 writes (2 * SampleWrite) to trigger check?
				// Actually logic is: if cnt >= smp*2 then reset and call check.
				// Wait, the code says: if w.cnt.Load() >= w.smp*2

				data := make([]byte, 20) // > MinimalSize (10)

				for i := 0; i < 10; i++ {
					m.Write(data)
				}

				// Should have switched
				Expect(m.IsParallel()).To(BeTrue())
			})

			It("[TC-MD-002] should not switch if writer count is low", func() {
				m = multi.New(true, false, cfg)

				// Add only 1 slow writer (MinimalWriter is 2)
				w1 := &slowWriter{delay: 500 * time.Microsecond}
				m.AddWriter(w1)

				data := make([]byte, 20)
				for i := 0; i < 10; i++ {
					m.Write(data)
				}

				Expect(m.IsParallel()).To(BeFalse())
			})
		})

		Context("Switching from Parallel to Sequential", func() {
			It("[TC-MD-005] should switch to sequential when latency is low", func() {
				m = multi.New(true, true, cfg) // Start parallel

				var buf1, buf2 bytes.Buffer
				m.AddWriter(&buf1, &buf2)

				// Fast writes
				data := make([]byte, 20)
				for i := 0; i < 10; i++ {
					m.Write(data)
				}

				// Should have switched because buffer writes are fast (< 10µs)
				Expect(m.IsParallel()).To(BeFalse())
			})
		})
	})
})

type slowWriter struct {
	delay time.Duration
}

func (s *slowWriter) Write(p []byte) (n int, err error) {
	time.Sleep(s.delay)
	return len(p), nil
}
