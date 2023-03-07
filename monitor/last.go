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

package monitor

import (
	"fmt"
	"sync"
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
)

type lastRun struct {
	m sync.RWMutex

	status  monsts.Status
	runtime time.Time

	isRise bool
	isFall bool

	cntRise uint8
	cntFall uint8

	uptime   time.Duration
	downtime time.Duration
	riseTime time.Duration
	fallTime time.Duration
	latency  time.Duration

	err error
}

func newLastRun() *lastRun {
	return &lastRun{
		m:        sync.RWMutex{},
		status:   monsts.KO,
		runtime:  time.Now(),
		isRise:   false,
		isFall:   false,
		cntRise:  0,
		cntFall:  0,
		uptime:   0,
		downtime: 0,
		riseTime: 0,
		fallTime: 0,
		latency:  0,
		err:      fmt.Errorf("no healcheck still run"),
	}
}

func (o *lastRun) Latency() time.Duration {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.latency
}

func (o *lastRun) FallTime() time.Duration {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.fallTime
}

func (o *lastRun) RiseTime() time.Duration {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.riseTime
}

func (o *lastRun) UpTime() time.Duration {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.uptime
}

func (o *lastRun) DownTime() time.Duration {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.downtime
}

func (o *lastRun) Status() monsts.Status {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.status
}

func (o *lastRun) IsRise() bool {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.isRise
}

func (o *lastRun) IsFall() bool {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.isFall
}

func (o *lastRun) Error() error {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.err
}

func (o *lastRun) setStatus(err error, dur time.Duration, cfg *runCfg) {
	o.m.Lock()
	defer o.m.Unlock()

	o.latency = dur

	if err != nil {
		o.err = err
		o.setStatusFall(cfg)
	} else {
		o.err = nil
		o.setStatusRise(cfg)
	}
}

func (o *lastRun) setStatusFall(cfg *runCfg) {
	if cfg == nil {
		return
	}

	sts := o.status
	dur := time.Since(o.runtime)
	o.runtime = time.Now()

	o.cntRise = 0
	o.cntFall++

	switch sts {
	case monsts.OK:
		if o.cntFall >= cfg.riseCountWarn {
			o.cntFall = 0
			o.status = monsts.Warn
		} else {
			o.status = monsts.OK
		}
		o.isFall = true
		o.isRise = false
		o.fallTime += dur
		o.uptime += dur

	case monsts.Warn:
		if o.cntFall >= cfg.riseCountKO {
			o.isFall = false
			o.cntFall = 0
			o.status = monsts.KO
		} else {
			o.isFall = true
			o.status = monsts.Warn
		}
		o.isRise = false
		o.fallTime += dur
		o.downtime += dur

	default:
		o.cntFall = 0
		o.isFall = false
		o.isRise = false
		o.status = monsts.KO
		o.downtime += dur
	}
}

func (o *lastRun) setStatusRise(cfg *runCfg) {
	if cfg == nil {
		return
	}

	sts := o.status
	dur := time.Since(o.runtime)
	o.runtime = time.Now()

	o.cntFall = 0
	o.cntRise++

	switch sts {
	case monsts.KO:
		if o.cntRise >= cfg.riseCountKO {
			o.cntRise = 0
			o.status = monsts.Warn
		} else {
			o.status = monsts.KO
		}
		o.isFall = false
		o.isRise = true
		o.riseTime += dur
		o.downtime += dur

	case monsts.Warn:
		if o.cntRise >= cfg.riseCountWarn {
			o.cntRise = 0
			o.isRise = false
			o.status = monsts.OK
		} else {
			o.isRise = true
			o.status = monsts.Warn
		}
		o.isFall = false
		o.riseTime += dur
		o.downtime += dur

	default:
		o.cntRise = 0
		o.isFall = false
		o.isRise = false
		o.status = monsts.OK
		o.uptime += dur
	}
}
