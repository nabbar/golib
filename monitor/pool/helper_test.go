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
 */

package pool_test

import (
	"context"
	"time"

	libdur "github.com/nabbar/golib/duration"
	logcfg "github.com/nabbar/golib/logger/config"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"
	libprm "github.com/nabbar/golib/prometheus"

	. "github.com/onsi/gomega"
)

// Helper functions for test setup

// newPool creates a new pool instance for testing
func newPool(ctx context.Context) monpool.Pool {
	return monpool.New(ctx)
}

// newMonitor creates a new monitor instance for testing
func newMonitor(x context.Context, nf montps.Info) montps.Monitor {
	m, e := libmon.New(x, nf)
	Expect(e).ToNot(HaveOccurred())
	Expect(m).ToNot(BeNil())
	m.RegisterLoggerDefault(fl)
	return m
}

// newInfo creates a new info instance with default values
func newInfo(name string) montps.Info {
	return newInfoWithData(name, nil)
}

// newInfoWithData creates a new info instance with custom data
func newInfoWithData(name string, data moninf.FuncInfo) montps.Info {
	i, e := moninf.New(name)
	Expect(e).ToNot(HaveOccurred())

	if data != nil {
		i.RegisterInfo(data)
	} else {
		i.RegisterInfo(func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"version": "1.0.0",
				"env":     "test",
			}, nil
		})
	}

	return i
}

// newConfig creates a new config for testing with fast intervals
func newConfig(name string) montps.Config {
	return montps.Config{
		Name:          name,
		CheckTimeout:  libdur.ParseDuration(20 * time.Millisecond),
		IntervalCheck: libdur.ParseDuration(20 * time.Millisecond),
		IntervalFall:  libdur.ParseDuration(20 * time.Millisecond),
		IntervalRise:  libdur.ParseDuration(20 * time.Millisecond),
		FallCountKO:   2,
		FallCountWarn: 2,
		RiseCountKO:   2,
		RiseCountWarn: 2,
		Logger: logcfg.Options{
			Stdout: &logcfg.OptionsStd{
				DisableStandard: true,
			},
		},
	}
}

// configureMonitor sets up a monitor with a health check and config
func configureMonitor(mon montps.Monitor, name string, healthCheck montps.HealthCheck) {
	if healthCheck != nil {
		mon.SetHealthCheck(healthCheck)
	} else {
		// Default health check that always succeeds
		mon.SetHealthCheck(func(ctx context.Context) error {
			return nil
		})
	}

	cfg := newConfig(name)
	Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())
}

// createTestMonitor creates a fully configured test monitor
func createTestMonitor(name string, healthCheck montps.HealthCheck) montps.Monitor {
	info := newInfo(name)
	mon := newMonitor(x, info)
	configureMonitor(mon, name, healthCheck)
	return mon
}

// newProm creates a new Prometheus instance for testing
func newProm() libprm.Prometheus {
	// Create a Prometheus instance with a unique namespace for testing
	return p
}

// newPromFunc creates a function that returns a Prometheus instance
func newPromFunc() libprm.FuncGetPrometheus {
	return func() libprm.Prometheus {
		return p
	}
}
