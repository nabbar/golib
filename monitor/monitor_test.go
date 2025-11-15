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

package monitor_test

import (
	"context"
	"sync/atomic"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libmon "github.com/nabbar/golib/monitor"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor", func() {
	var (
		lc = lo.Clone()

		ctx context.Context
		cnl context.CancelFunc
		nfo montps.Info
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, time.Second)
		nfo = newInfo(nil)
	})

	AfterEach(func() {
		if cnl != nil {
			cnl()
		}
	})

	Describe("New", func() {
		Context("when info is nil", func() {
			It("should return an error", func() {
				mon, err := libmon.New(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("info cannot be nil"))
				Expect(mon).To(BeNil())
			})
		})

		Context("when info is valid", func() {
			It("should create a new monitor instance", func() {
				newMonitor(ctx, nfo)
			})
		})

		Context("when info is valid", func() {
			It("should create a new monitor instance can start", func() {
				xx, nn := context.WithTimeout(ctx, 10*time.Second)
				defer nn()

				mon := newMonitor(ctx, nfo)
				Expect(mon.Start(xx)).ToNot(HaveOccurred())

				time.Sleep(50 * time.Millisecond)
				Expect(mon.IsRunning()).To(BeTrue())

				time.Sleep(50 * time.Millisecond)
				Expect(mon.Stop(xx)).ToNot(HaveOccurred())

				time.Sleep(50 * time.Millisecond)
				Expect(mon.IsRunning()).To(BeFalse())
			})
		})
	})

	Describe("HealthCheck", func() {
		var mon montps.Monitor

		BeforeEach(func() {
			mon = newMonitor(ctx, nfo)
		})

		AfterEach(func() {
			Expect(mon.Stop(x)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeFalse())
		})

		It("should set and get health check function", func() {
			healthCheckCalled := new(atomic.Bool)
			healthCheckFunc := func(ctx context.Context) error {
				healthCheckCalled.Store(true)
				return nil
			}

			mon.SetHealthCheck(healthCheckFunc)
			Expect(mon.GetHealthCheck()).ToNot(BeNil())

			// Vérifier que la fonction peut être appelée
			lx, ln := context.WithTimeout(x, 10*time.Millisecond)
			defer ln()

			Expect(mon.GetHealthCheck()(lx)).ToNot(HaveOccurred())
			Expect(healthCheckCalled.Load()).To(BeTrue())
		})

		It("should set and get config function", func() {
			lx, ln := context.WithTimeout(x, 100*time.Millisecond)
			defer ln()

			Expect(mon.SetConfig(lx, montps.Config{
				Name:          nfo.Name(),
				CheckTimeout:  libdur.ParseDuration(50 * time.Millisecond),
				IntervalCheck: libdur.ParseDuration(50 * time.Millisecond),
				IntervalFall:  libdur.ParseDuration(50 * time.Millisecond),
				IntervalRise:  libdur.ParseDuration(50 * time.Millisecond),
				FallCountKO:   3,
				FallCountWarn: 2,
				RiseCountKO:   3,
				RiseCountWarn: 2,
				Logger:        lc,
			})).ToNot(HaveOccurred())
		})
	})
})
