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

package pool

import (
	"context"
	"errors"
	"time"

	liblog "github.com/nabbar/golib/logger"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libprm "github.com/nabbar/golib/prometheus"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	shlcmd "github.com/nabbar/golib/shell/command"
)

// Export internal types and methods for testing purposes (White-box testing via black-box package)

// --- Mocks ---

type MockMetric struct {
	libmet.Metric
}

func (m *MockMetric) Observe(labelValues []string, value float64) error {
	return errors.New("mock observe error")
}

func (m *MockMetric) SetGaugeValue(labelValues []string, value float64) error {
	return errors.New("mock gauge error")
}

type MockPrometheus struct {
	libprm.Prometheus
}

func (m *MockPrometheus) AddMetric(b bool, met libmet.Metric) error {
	return errors.New("mock add metric error")
}

type MockMonitor struct {
	NameVal string
}

// MonitorStatus
func (m *MockMonitor) Name() string                 { return m.NameVal }
func (m *MockMonitor) Status() monsts.Status        { return monsts.OK }
func (m *MockMonitor) Message() string              { return "mock message" }
func (m *MockMonitor) IsRise() bool                 { return false }
func (m *MockMonitor) IsFall() bool                 { return false }
func (m *MockMonitor) Latency() time.Duration       { return time.Second }
func (m *MockMonitor) Uptime() time.Duration        { return time.Second }
func (m *MockMonitor) Downtime() time.Duration      { return time.Second }
func (m *MockMonitor) MarshalText() ([]byte, error) { return []byte("mock"), nil }
func (m *MockMonitor) MarshalJSON() ([]byte, error) { return []byte(`"mock"`), nil }

// MonitorMetrics
func (m *MockMonitor) RegisterMetricsName(names ...string)                  {}
func (m *MockMonitor) RegisterMetricsAddName(names ...string)               {}
func (m *MockMonitor) RegisterCollectMetrics(fct libprm.FuncCollectMetrics) {}
func (m *MockMonitor) CollectLatency() time.Duration                        { return time.Second }
func (m *MockMonitor) CollectUpTime() time.Duration                         { return time.Second }
func (m *MockMonitor) CollectDownTime() time.Duration                       { return time.Second }
func (m *MockMonitor) CollectRiseTime() time.Duration                       { return time.Second }
func (m *MockMonitor) CollectFallTime() time.Duration                       { return time.Second }
func (m *MockMonitor) CollectStatus() (monsts.Status, bool, bool)           { return monsts.OK, true, false }

// MonitorInfo
func (m *MockMonitor) InfoGet() montps.Info            { return nil }
func (m *MockMonitor) InfoUpd(inf montps.Info)         {}
func (m *MockMonitor) InfoName() string                { return m.NameVal }
func (m *MockMonitor) InfoMap() map[string]interface{} { return nil }

// Runner
func (m *MockMonitor) Start(ctx context.Context) error   { return nil }
func (m *MockMonitor) Stop(ctx context.Context) error    { return nil }
func (m *MockMonitor) Restart(ctx context.Context) error { return nil }
func (m *MockMonitor) IsRunning() bool                   { return true }

// Monitor
func (m *MockMonitor) SetConfig(ctx context.Context, cfg montps.Config) error { return nil }
func (m *MockMonitor) GetConfig() montps.Config                               { return montps.Config{} }
func (m *MockMonitor) RegisterLoggerDefault(fct liblog.FuncLog)               {}
func (m *MockMonitor) SetHealthCheck(fct montps.HealthCheck)                  {}
func (m *MockMonitor) GetHealthCheck() montps.HealthCheck                     { return nil }
func (m *MockMonitor) Clone(ctx context.Context) (montps.Monitor, error)      { return m, nil }

// --- Accessors ---

func PoolCollectMetricLatency(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricLatency(ctx, m)
	}
}

func PoolCollectMetricUptime(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricUptime(ctx, m)
	}
}

func PoolCollectMetricDowntime(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricDowntime(ctx, m)
	}
}

func PoolCollectMetricRiseTime(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricRiseTime(ctx, m)
	}
}

func PoolCollectMetricFallTime(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricFallTime(ctx, m)
	}
}

func PoolCollectMetricStatus(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricStatus(ctx, m)
	}
}

func PoolCollectMetricRising(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricRising(ctx, m)
	}
}

func PoolCollectMetricFalling(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricFalling(ctx, m)
	}
}

func PoolCollectMetricSLis(p Pool, ctx context.Context, m libmet.Metric) {
	if v, ok := p.(*pool); ok {
		v.collectMetricSLis(ctx, m)
	}
}

func PoolCreateMetrics(p Pool) error {
	if v, ok := p.(*pool); ok {
		return v.createMetrics()
	}
	return errors.New("invalid pool type")
}

func PoolSetPrometheusFactory(p Pool, f libprm.FuncGetPrometheus) {
	if v, ok := p.(*pool); ok {
		v.m.Lock()
		defer v.m.Unlock()
		v.fp = f
	}
}

func PoolInjectStore(p Pool, key string, val interface{}) {
	if v, ok := p.(*pool); ok {
		v.p.Store(key, val)
	}
}

func PoolLoadStore(p Pool, key string) (interface{}, bool) {
	if v, ok := p.(*pool); ok {
		return v.p.Load(key)
	}
	return nil, false
}

func PoolShlCmdStart(p Pool, ctx context.Context) shlcmd.Command {
	if v, ok := p.(*pool); ok {
		return v.shlCmdStart(ctx)
	}
	return nil
}

func PoolShlCmdStop(p Pool, ctx context.Context) shlcmd.Command {
	if v, ok := p.(*pool); ok {
		return v.shlCmdStop(ctx)
	}
	return nil
}

func PoolShlCmdRestart(p Pool, ctx context.Context) shlcmd.Command {
	if v, ok := p.(*pool); ok {
		return v.shlCmdRestart(ctx)
	}
	return nil
}
