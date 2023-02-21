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

	liblog "github.com/nabbar/golib/logger"

	libctx "github.com/nabbar/golib/context"

	liberr "github.com/nabbar/golib/errors"
	monsts "github.com/nabbar/golib/monitor/status"
	libprm "github.com/nabbar/golib/prometheus"
	libsrv "github.com/nabbar/golib/server"
)

type HealthCheck func(ctx context.Context) error

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
	RegisterMetricsName(names ...string)
	RegisterMetricsAddName(names ...string)
	RegisterCollectMetrics(fct libprm.FuncCollectMetrics)

	CollectLatency() time.Duration
	CollectUpTime() time.Duration
	CollectDownTime() time.Duration
	CollectRiseTime() time.Duration
	CollectFallTime() time.Duration
	CollectStatus() (sts string, rise bool, fall bool)
}

type MonitorInfo interface {
	InfoGet() Info
	InfoUpd(inf Info)
	InfoName() string
	InfoMap() map[string]interface{}
}

type Monitor interface {
	MonitorInfo
	MonitorStatus
	MonitorMetrics
	libsrv.Server

	// SetConfig is used to set or update config of the monitor
	SetConfig(ctx libctx.FuncContext, cfg Config) liberr.Error

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
