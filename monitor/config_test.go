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
	"time"

	libdur "github.com/nabbar/golib/duration"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor Configuration", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		nfo montps.Info
		mon montps.Monitor
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 5*time.Second)
		nfo = newInfo(nil)
		mon = newMonitor(x, nfo)
	})

	AfterEach(func() {
		if mon != nil && mon.IsRunning() {
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("SetConfig and GetConfig", func() {
		It("should set and retrieve configuration", func() {
			cfg := montps.Config{
				Name:          "test-monitor",
				CheckTimeout:  libdur.ParseDuration(10 * time.Second),
				IntervalCheck: libdur.ParseDuration(2 * time.Second),
				IntervalFall:  libdur.ParseDuration(1 * time.Second),
				IntervalRise:  libdur.ParseDuration(1 * time.Second),
				FallCountKO:   5,
				FallCountWarn: 3,
				RiseCountKO:   5,
				RiseCountWarn: 3,
				Logger:        lo.Clone(),
			}

			Expect(mon.SetConfig(ctx, cfg)).ToNot(HaveOccurred())

			retrieved := mon.GetConfig()
			Expect(retrieved.Name).To(Equal("test-monitor"))
			Expect(retrieved.CheckTimeout.Time()).To(Equal(10 * time.Second))
			Expect(retrieved.IntervalCheck.Time()).To(Equal(2 * time.Second))
			Expect(retrieved.IntervalFall.Time()).To(Equal(1 * time.Second))
			Expect(retrieved.IntervalRise.Time()).To(Equal(1 * time.Second))
			Expect(retrieved.FallCountKO).To(Equal(uint8(5)))
			Expect(retrieved.FallCountWarn).To(Equal(uint8(3)))
			Expect(retrieved.RiseCountKO).To(Equal(uint8(5)))
			Expect(retrieved.RiseCountWarn).To(Equal(uint8(3)))
		})

		It("should normalize values below minimums", func() {
			cfg := montps.Config{
				Name:          "test-monitor",
				CheckTimeout:  libdur.ParseDuration(1 * time.Second),        // Below 5s minimum
				IntervalCheck: libdur.ParseDuration(100 * time.Millisecond), // Below 1s minimum
				IntervalFall:  libdur.ParseDuration(10 * time.Nanosecond),
				IntervalRise:  libdur.ParseDuration(100 * time.Millisecond),
				FallCountKO:   0, // Below 1 minimum
				FallCountWarn: 0,
				RiseCountKO:   0,
				RiseCountWarn: 0,
				Logger:        lo.Clone(),
			}

			Expect(mon.SetConfig(ctx, cfg)).ToNot(HaveOccurred())

			retrieved := mon.GetConfig()
			Expect(retrieved.CheckTimeout.Time()).To(Equal(1 * time.Second))
			Expect(retrieved.IntervalCheck.Time()).To(Equal(100 * time.Millisecond))
			Expect(retrieved.IntervalFall.Time()).To(Equal(100 * time.Millisecond))
			Expect(retrieved.IntervalRise.Time()).To(Equal(100 * time.Millisecond))
			Expect(retrieved.FallCountKO).To(Equal(uint8(1)))
			Expect(retrieved.FallCountWarn).To(Equal(uint8(1)))
			Expect(retrieved.RiseCountKO).To(Equal(uint8(1)))
			Expect(retrieved.RiseCountWarn).To(Equal(uint8(1)))
		})

		It("should use default name when empty", func() {
			cfg := montps.Config{
				Name:          "",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(1 * time.Second),
				Logger:        lo.Clone(),
			}

			Expect(mon.SetConfig(ctx, cfg)).ToNot(HaveOccurred())

			retrieved := mon.GetConfig()
			Expect(retrieved.Name).To(Equal("not named"))
		})

		It("should default IntervalFall to IntervalCheck when too low", func() {
			cfg := montps.Config{
				Name:          "test",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(3 * time.Second),
				IntervalFall:  libdur.ParseDuration(500 * time.Nanosecond),
				Logger:        lo.Clone(),
			}

			Expect(mon.SetConfig(ctx, cfg)).ToNot(HaveOccurred())

			retrieved := mon.GetConfig()
			Expect(retrieved.IntervalFall.Time()).To(Equal(3 * time.Second))
		})

		It("should default IntervalRise to IntervalCheck when too low", func() {
			cfg := montps.Config{
				Name:          "test",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(3 * time.Second),
				IntervalRise:  libdur.ParseDuration(500 * time.Nanosecond),
				Logger:        lo.Clone(),
			}

			Expect(mon.SetConfig(ctx, cfg)).ToNot(HaveOccurred())

			retrieved := mon.GetConfig()
			Expect(retrieved.IntervalRise.Time()).To(Equal(3 * time.Second))
		})
	})

	Describe("Default Configuration", func() {
		It("should have valid default values", func() {
			cfg := mon.GetConfig()

			Expect(cfg.CheckTimeout.Time()).To(BeNumerically(">=", 5*time.Second))
			Expect(cfg.IntervalCheck.Time()).To(BeNumerically(">=", 1*time.Second))
			Expect(cfg.IntervalFall.Time()).To(BeNumerically(">=", 1*time.Second))
			Expect(cfg.IntervalRise.Time()).To(BeNumerically(">=", 1*time.Second))
			Expect(cfg.FallCountKO).To(BeNumerically(">=", 1))
			Expect(cfg.FallCountWarn).To(BeNumerically(">=", 1))
			Expect(cfg.RiseCountKO).To(BeNumerically(">=", 1))
			Expect(cfg.RiseCountWarn).To(BeNumerically(">=", 1))
		})
	})

	Describe("Name Operations", func() {
		It("should return configured name", func() {
			cfg := montps.Config{
				Name:          "my-custom-name",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(1 * time.Second),
				Logger:        lo.Clone(),
			}

			Expect(mon.SetConfig(ctx, cfg)).ToNot(HaveOccurred())
			Expect(mon.Name()).To(Equal("my-custom-name"))
		})

		It("should return default name when not configured", func() {
			name := mon.Name()
			Expect(name).To(Equal("not named"))
		})

		It("should return info name separately", func() {
			Expect(mon.InfoName()).To(Equal(key))
		})
	})

	Describe("Logger Configuration", func() {
		It("should configure logger with provided options", func() {
			Expect(mon.SetConfig(ctx, newConfig(nfo))).ToNot(HaveOccurred())
		})
	})

	Describe("Context Provider", func() {
		It("should use provided context function", func() {
			lctx, lcnl := context.WithTimeout(context.Background(), 5*time.Second)
			defer lcnl()

			cfg := newConfig(nfo)
			cfg.CheckTimeout = libdur.ParseDuration(5 * time.Second)
			cfg.IntervalCheck = libdur.ParseDuration(1 * time.Second)

			Expect(mon.SetConfig(lctx, cfg)).ToNot(HaveOccurred())
		})
	})
})
