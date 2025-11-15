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
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
)

// Name returns the configured name of the monitor.
// If no name is configured, returns the default monitor name.
func (o *mon) Name() string {
	return o.getName()
}

// InfoName returns the name from the Info metadata.
// This is thread-safe and returns the dynamic name if registered.
func (o *mon) InfoName() string {
	if i := o.i.Load(); i != nil {
		return i.Name()
	}
	return "invalid nil info"
}

// InfoMap returns the info data map from the Info metadata.
// This is thread-safe and returns the dynamic info if registered.
func (o *mon) InfoMap() map[string]interface{} {
	if i := o.i.Load(); i != nil {
		return i.Info()
	}
	return nil
}

// InfoGet returns the Info instance used by the monitor.
// This is thread-safe and acquires a write lock.
func (o *mon) InfoGet() montps.Info {
	return o.i.Load()
}

// InfoUpd updates the Info instance used by the monitor.
// This is thread-safe and acquires a write lock.
func (o *mon) InfoUpd(inf montps.Info) {
	if inf == nil {
		return
	}

	o.i.Store(inf)
}

// Status returns the current health status of the monitored component.
// Returns KO, Warn, or OK based on the last health check results.
func (o *mon) Status() monsts.Status {
	return o.getLastCheck().Status()
}

// Message returns the error message from the last health check.
// Returns an empty string if the last check was successful.
func (o *mon) Message() string {
	if err := o.getLastCheck().Error(); err != nil {
		return err.Error()
	}

	return ""
}

// IsRise returns true if the monitor is currently transitioning from a lower to higher health status.
// This indicates the component is recovering.
func (o *mon) IsRise() bool {
	return o.getLastCheck().IsRise()
}

// IsFall returns true if the monitor is currently transitioning from a higher to lower health status.
// This indicates the component is degrading.
func (o *mon) IsFall() bool {
	return o.getLastCheck().IsFall()
}

// Latency returns the duration of the last health check execution.
func (o *mon) Latency() time.Duration {
	return o.getLastCheck().Latency()
}

// Uptime returns the total time the component has been in OK status.
func (o *mon) Uptime() time.Duration {
	return o.getLastCheck().UpTime()
}

// Downtime returns the total time the component has been in KO or Warn status.
func (o *mon) Downtime() time.Duration {
	return o.getLastCheck().DownTime()
}

// mdlStatus is a middleware function that wraps the health check execution.
// It measures latency and updates the last check status.
func (o *mon) mdlStatus(m middleWare) error {
	ts := time.Now()
	err := m.Next()

	lst := o.getLastCheck()
	lst.setStatus(err, time.Since(ts), m.Config())
	o.setLastCheck(lst)

	return err
}
