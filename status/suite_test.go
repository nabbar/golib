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
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	libdur "github.com/nabbar/golib/duration"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	monpol "github.com/nabbar/golib/monitor/pool"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libprm "github.com/nabbar/golib/prometheus"
	libsts "github.com/nabbar/golib/status"
	stsctr "github.com/nabbar/golib/status/control"
	stslmd "github.com/nabbar/golib/status/listmandatory"
	stsmdt "github.com/nabbar/golib/status/mandatory"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestStatus is the entry point for Ginkgo test suite.
// It integrates Ginkgo with Go's testing framework and runs all status package tests.
func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status Suite")
}

var (
	globalCtx context.Context
	globalCnl context.CancelFunc

	l liblog.Logger
	p libprm.Prometheus
	u monpol.Pool
)

var _ = BeforeSuite(func() {
	globalCtx, globalCnl = context.WithTimeout(context.Background(), 30*time.Second)

	l = liblog.New(globalCtx)
	e := l.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})
	Expect(e).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	globalCnl()
})

// newMonitor creates a new monitor with the given name and optional data.
// If data is nil, default data is used.
func newMonitor(name string, data map[string]interface{}, health func(ctx context.Context) error) libmon.Monitor {
	n, e := moninf.New(name)
	Expect(e).NotTo(HaveOccurred())

	if len(data) < 1 {
		data = map[string]interface{}{
			"version": "1.0.0",
			"env":     "test",
			"time":    time.Now(),
		}
	}

	n.RegisterInfo(func() (map[string]interface{}, error) {
		return data, nil
	})

	m, e := libmon.New(globalCtx, n)
	Expect(e).NotTo(HaveOccurred())

	e = m.SetConfig(globalCtx, newConfig(n))
	Expect(e).NotTo(HaveOccurred())

	// Register a health check that always succeeds
	m.SetHealthCheck(health)

	e = m.Start(globalCtx)
	Expect(e).NotTo(HaveOccurred())

	return m
}

// newHealthyMonitor creates a monitor that always reports OK status.
func newHealthyMonitor(name string) libmon.Monitor {
	return newMonitor(name, nil, func(ctx context.Context) error {
		return nil
	})
}

// newUnhealthyMonitor creates a monitor that always reports KO status.
func newUnhealthyMonitor(name string) libmon.Monitor {
	return newMonitor(name, nil, func(ctx context.Context) error {
		return fmt.Errorf("monitor %s is unhealthy", name)
	})
}

func newPool() monpol.Pool {
	if u != nil {
		Expect(u.Stop(globalCtx)).ToNot(HaveOccurred())
		Expect(len(u.UnregisterMetrics())).To(Equal(0))
		Expect(len(p.ClearMetric(true, true))).To(Equal(0))
		u = monpol.New(globalCtx)
		p = libprm.New(globalCtx)
	} else {
		u = monpol.New(globalCtx)
		p = libprm.New(globalCtx)
	}

	// Register metrics with shared Prometheus instance
	e := u.RegisterMetrics(func() libprm.Prometheus {
		return p
	}, func() liblog.Logger {
		return l
	})

	// if only to add breakpoint to debug
	if e != nil {
		Expect(e).NotTo(HaveOccurred())
	}

	// Start the pool so monitors begin health checks
	e = u.Start(globalCtx)
	Expect(e).NotTo(HaveOccurred())

	return u
}

// testMonitorStabilizeDelay is the time to wait after adding a monitor
// for it to stabilize and reach its expected state.
// With IntervalCheck=10ms and Count=1, we need time for:
// - Monitor to start (~10ms)
// - First health check to execute (~10ms)
// - Status to propagate (~10ms)
// Using 200ms to be safe across different system loads.
const testMonitorStabilizeDelay = 200 * time.Millisecond

// newConfig creates a fast, simple monitor configuration for tests.
// The goal is to test status behavior, not monitor behavior, so we use
// minimal delays and counts for quick, predictable test execution.
func newConfig(nf montps.Info) montps.Config {
	return montps.Config{
		Name:          nf.Name(),
		CheckTimeout:  libdur.ParseDuration(10 * time.Millisecond),
		IntervalCheck: libdur.ParseDuration(10 * time.Millisecond),
		IntervalFall:  libdur.ParseDuration(10 * time.Millisecond),
		IntervalRise:  libdur.ParseDuration(10 * time.Millisecond),
		FallCountKO:   1, // Immediate transition to KO on first failure
		FallCountWarn: 1,
		RiseCountKO:   1, // Immediate transition to OK on first success
		RiseCountWarn: 1,
		Logger: logcfg.Options{
			Stdout: &logcfg.OptionsStd{
				DisableStandard: true,
			},
		},
	}
}

func newListMandatory(l map[stsctr.Mode][]string) []stsmdt.Mandatory {
	r := stslmd.New()
	for c, i := range l {
		o := stsmdt.New()
		o.SetMode(c)
		o.KeyAdd(i...)
		r.Add(o)
	}
	return r.GetList()
}

func newStatusConfig(m ...stsmdt.Mandatory) libsts.Config {
	return libsts.Config{
		ReturnCode: map[monsts.Status]int{
			monsts.OK:   http.StatusOK,
			monsts.Warn: http.StatusAccepted,
			monsts.KO:   http.StatusServiceUnavailable,
		},
		MandatoryComponent: libsts.ParseList(m...),
	}
}
