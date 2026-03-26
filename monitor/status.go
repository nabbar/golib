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

// Name returns the identifier of the monitor as configured during its initialization.
// If no specific name has been set, it returns the default monitor name ("not named").
func (o *mon) Name() string {
	return o.getName()
}

// InfoName retrieves the name stored within the monitor's metadata (Info).
// This method is thread-safe and is designed to return a descriptive name provided by the Info instance,
// or "invalid nil info" if the Info object is not initialized.
func (o *mon) InfoName() string {
	if i := o.i.Load(); i != nil {
		return i.Name()
	}
	return "invalid nil info"
}

// InfoMap returns a map representation of the metadata associated with the monitor.
// This is thread-safe and allows access to dynamic information stored within the Info object.
// It returns nil if no Info object is currently registered.
func (o *mon) InfoMap() map[string]interface{} {
	if i := o.i.Load(); i != nil {
		return i.Data()
	}
	return nil
}

// InfoGet returns the current Info instance (metadata implementation) associated with the monitor.
// This operation is thread-safe as it uses an atomic load to retrieve the reference.
func (o *mon) InfoGet() montps.Info {
	return o.i.Load()
}

// InfoUpd replaces the current Info instance (metadata) with a new one.
//
// Parameters:
//   - inf: The new metadata implementation to associate with the monitor.
//
// If the provided info instance is nil, the operation is ignored.
// This update is thread-safe and uses an atomic store to ensure consistency for concurrent readers.
func (o *mon) InfoUpd(inf montps.Info) {
	if inf == nil {
		return
	}

	o.i.Store(inf)
}

// Status returns the current health status of the monitored component (e.g., OK, Warn, KO).
//
// Implementation Detail:
// This method is optimized for high-frequency polling. It performs a lock-free atomic load
// from the internal state, ensuring near-zero CPU overhead even under extreme load.
func (o *mon) Status() monsts.Status {
	return o.l.Status()
}

// Message retrieves the error message associated with the most recent health check execution.
// If the last check was successful (status OK), it returns an empty string.
// Otherwise, it returns the descriptive error message captured from the last failed execution.
func (o *mon) Message() string {
	if err := o.l.Error(); err != nil {
		return err.Error()
	}

	return ""
}

// IsRise reports whether the monitored component is currently in a "rising" phase.
// A rising phase occurs when consecutive successful checks are being counted while the component
// is transitioning from a degraded state (KO or Warn) back toward an OK status.
func (o *mon) IsRise() bool {
	return o.l.IsRise()
}

// IsFall reports whether the monitored component is currently in a "falling" phase.
// A falling phase occurs when consecutive failed checks are being counted while the component
// is transitioning from a healthy state (OK or Warn) toward a degraded status (Warn or KO).
func (o *mon) IsFall() bool {
	return o.l.IsFall()
}

// Latency returns the execution duration of the very last health check that was performed.
// This metric represents the time taken by the monitored service or function to respond.
// It uses atomic loading for high-performance reading.
func (o *mon) Latency() time.Duration {
	return o.l.Latency()
}

// Uptime returns the total accumulated duration the component has spent in the OK status
// since the monitor was started or since its counters were last reset.
// It uses atomic loading for high-performance reading.
func (o *mon) Uptime() time.Duration {
	return o.l.UpTime()
}

// Downtime returns the total accumulated duration the component has spent in non-OK statuses (Warn or KO)
// since the monitor was started or since its counters were last reset.
// It uses atomic loading for high-performance reading.
func (o *mon) Downtime() time.Duration {
	return o.l.DownTime()
}

// mdlStatus acts as a middleware function for the health check execution pipeline.
//
// Responsibilities:
//  1. Records the start time of the check execution.
//  2. Triggers the next step in the middleware chain (eventually the actual HealthCheck).
//  3. Captures the result (success/failure) and calculates the elapsed time (latency).
//  4. Updates the internal metrics container (o.l) with the results and configuration.
//
// This middleware is essential for the monitor's state machine, as it handles the logic
// for status transitions (Rise/Fall) and metric accumulation.
func (o *mon) mdlStatus(m middleWare) error {
	ts := time.Now()
	err := m.Next()

	// Atomic update of the status and metrics within the metrics container.
	o.l.setStatus(err, time.Since(ts), m.Config())

	return err
}
