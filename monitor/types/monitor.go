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

package types

import (
	"context"
	"encoding"
	"encoding/json"
	"time"

	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	monsts "github.com/nabbar/golib/monitor/status"
	libprm "github.com/nabbar/golib/prometheus"
	libsrv "github.com/nabbar/golib/runner"
)

// HealthCheck is a function type that performs a health check operation.
// The function should return nil if the component is healthy, or an error describing
// the issue if it's not. The context should be respected for timeout and cancellation.
//
// Example:
//
//	healthCheck := func(ctx context.Context) error {
//	    return db.PingContext(ctx)
//	}
type HealthCheck func(ctx context.Context) error

// MonitorStatus provides methods for querying the current status and state of a monitor.
// It includes encoding capabilities for serializing status information.
type MonitorStatus interface {
	encoding.TextMarshaler
	json.Marshaler

	// Name return the name of the monitor.
	Name() string

	// Status return the last status (OK / Warn / KO).
	Status() monsts.Status

	// Message return the last error, warning, message of the last status
	Message() string

	// IsRise return true if rising status from KO or Warn
	IsRise() bool

	// IsFall return true if falling status to KO or Warn
	IsFall() bool

	// Latency return the last check's latency
	Latency() time.Duration

	// Uptime return the total duration of uptime (OK status)
	Uptime() time.Duration

	// Downtime return the total duration of downtime (KO status)
	Downtime() time.Duration
}

type MonitorMetrics interface {
	// RegisterMetricsName registers names of metrics for this monitor.
	// The names are stored in memory and used to register metrics
	// when calling RegisterCollectMetrics.
	//
	// The metric names should be in the format:
	//   <monitor_name>_<metric_name>
	//
	// For example:
	//   monitor_latency
	//   monitor_uptime
	//   monitor_downtime
	//
	// The names are case-sensitive.
	//
	// Parameters:
	//   names - a list of metric names to register.
	//
	// Returns:
	//   None.
	RegisterMetricsName(names ...string)
	// RegisterMetricsAddName registers additional names of metrics for this monitor.
	// The names are added to the list of names stored in memory and used to register metrics
	// when calling RegisterCollectMetrics.
	//
	// The metric names should be in the format:
	//   <monitor_name>_<metric_name>
	//
	// For example:
	//   monitor_latency
	//   monitor_uptime
	//   monitor_downtime
	//
	// The names are case-sensitive.
	//
	// Parameters:
	//   names - a list of metric names to register.
	//
	// Returns:
	//   None.
	RegisterMetricsAddName(names ...string)
	// RegisterCollectMetrics registers a function to collect metrics for this monitor.
	//
	// The function will be called by the monitor with the correct context and metric names.
	// The function should update the metrics with the correct values.
	//
	// Parameters:
	//   fct - the function to register.
	//
	// Returns:
	//   None.
	RegisterCollectMetrics(fct libprm.FuncCollectMetrics)

	// CollectLatency returns the last check's latency.
	//
	// It returns the time spent between the start of the last check and the moment
	// the last check returned a status (OK, Warn, KO).
	//
	// It returns an error if the last check didn't return a status.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   time.Duration - the last check's latency.
	//   liberr.Error - an error if the last check didn't return a status.
	CollectLatency() time.Duration
	// CollectUpTime returns the total duration of uptime (OK status) since the
	// monitor was created.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   time.Duration - the total duration of uptime (OK status).
	CollectUpTime() time.Duration
	// CollectDownTime returns the total duration of downtime (KO status) since the
	// monitor was created.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   time.Duration - the total duration of downtime (KO status).
	CollectDownTime() time.Duration
	// CollectRiseTime returns the total duration of rising status since the monitor was created.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   time.Duration - the total duration of rising status.
	CollectRiseTime() time.Duration
	// CollectFallTime returns the total duration of falling status since the
	// monitor was created.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   time.Duration - the total duration of falling status.
	CollectFallTime() time.Duration
	// CollectStatus returns the current status of the monitor and whether
	// the status is rising or falling.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   sts monsts.Status - the current status of the monitor.
	//   rise bool - whether the status is rising.
	//   fall bool - whether the status is falling.
	CollectStatus() (sts monsts.Status, rise bool, fall bool)
}

type MonitorInfo interface {
	// InfoGet returns the current information of the monitor.
	//
	// It returns the monitor information as a struct of type Info.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   Info - the current information of the monitor.
	InfoGet() Info
	// InfoUpd updates the information of the monitor.
	//
	// Parameters:
	// - inf Info: the new information of the monitor.
	//
	// Returns:
	// - None.
	InfoUpd(inf Info)
	// InfoName returns the name of the monitor information.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   string - the name of the monitor information.
	InfoName() string
	// InfoMap returns the current information of the monitor as a map of strings to
	// interface values.
	//
	// It is a convenience function that allows to access the information
	// without having to know the exact structure of the Info type.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   map[string]interface{} - the current information of the monitor as a map.
	//
	InfoMap() map[string]interface{}
}

type Monitor interface {
	MonitorInfo
	MonitorStatus
	MonitorMetrics
	libsrv.Runner

	// SetConfig is used to set or update config of the monitor
	SetConfig(ctx context.Context, cfg Config) liberr.Error

	// RegisterLoggerDefault is used to define the default logger.
	// Default logger can be used to extend options logger from it
	RegisterLoggerDefault(fct liblog.FuncLog)

	// GetConfig is used to retrieve config of the monitor
	GetConfig() Config

	// SetHealthCheck is used to set or update the healthcheck func
	SetHealthCheck(fct HealthCheck)

	// GetHealthCheck is used to retrieve the healthcheck func
	GetHealthCheck() HealthCheck

	// Clone is used to clone monitor to another standalone instance
	Clone(ctx context.Context) (Monitor, liberr.Error)
}
