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
	"testing"
	"time"

	libdur "github.com/nabbar/golib/duration"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	n context.CancelFunc
	x context.Context

	l  liblog.Logger
	fl = func() liblog.Logger {
		return l
	}
	lo = logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	}

	key = "test-monitor"
)

func TestMonitor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Monitor Suite")
}

var _ = BeforeSuite(func() {
	x, n = context.WithTimeout(context.Background(), 30*time.Second)

	l = liblog.New(x)
	Expect(l.SetOptions(&lo)).ToNot(HaveOccurred())

})

var _ = AfterSuite(func() {
	if n != nil {
		n()
	}
})

func newMonitor(x context.Context, nf montps.Info) montps.Monitor {
	m, e := libmon.New(x, nf)
	Expect(e).ToNot(HaveOccurred())
	Expect(m).ToNot(BeNil())
	m.RegisterLoggerDefault(fl)
	return m
}

func newInfo(d moninf.FuncInfo) montps.Info {
	return newInfoWithName(key, d)
}

func newInfoWithName(name string, d moninf.FuncInfo) montps.Info {
	i, e := moninf.New(name)
	Expect(e).ToNot(HaveOccurred())

	if d != nil {
		i.RegisterInfo(d)
	} else {
		i.RegisterInfo(func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"version": "1.0.0",
				"check":   "beforeEach",
			}, nil
		})
	}

	return i
}

func newConfig(nf montps.Info) montps.Config {
	return montps.Config{
		Name:          nf.Name(),
		CheckTimeout:  libdur.ParseDuration(20 * time.Millisecond),
		IntervalCheck: libdur.ParseDuration(20 * time.Millisecond),
		IntervalFall:  libdur.ParseDuration(20 * time.Millisecond),
		IntervalRise:  libdur.ParseDuration(20 * time.Millisecond),
		FallCountKO:   2,
		FallCountWarn: 2,
		RiseCountKO:   2,
		RiseCountWarn: 2,
		Logger:        lo,
	}
}
