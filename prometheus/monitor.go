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

package prometheus

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	librtr "github.com/nabbar/golib/router"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GinPrometheus is an object that uses to set gin server monitor.
type monitor struct {
	m sync.Mutex

	exclude     []string
	slowTime    int32
	reqDuration []float64

	metrics map[string]*atomic.Value
}

// Expose adds metric path to a given router.
// The router can be different with the one passed to UseWithoutExposingEndpoint.
// This allows to expose metrics on different port.
func (m *monitor) Expose(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// MiddleWare as gin monitor middleware.
func (m *monitor) MiddleWare(c *gin.Context) {
	startTime := c.GetInt64(librtr.GinContextStartUnixNanoTime)

	if startTime == 0 {
		startTime = time.Now().UnixNano()
	}

	path := c.GetString(librtr.GinContextRequestPath)
	if path == "" {
		path = c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; len(raw) > 0 {
			path += "?" + raw
		}
	}

	if len(m.exclude) > 0 {
		r := c.Request.URL.Path
		if !strings.HasPrefix(r, "/") {
			r = "/" + r
		}
		for _, p := range m.exclude {
			if p != "" && strings.HasPrefix(r, p) {
				return
			}
		}
	}

	// execute normal process.
	c.Next()

	// after request
	m.CollectMetrics(c, time.Unix(0, startTime))
}

func (m *monitor) CollectMetrics(c *gin.Context, start time.Time) {
	for _, k := range m.ListMetric() {
		metric := m.GetMetric(k)
		metric.Collect(c, start)
		_ = m.SetMetric(metric)
	}
}

func (m *monitor) ExcludePath(startWith ...string) {
	m.m.Lock()
	defer m.m.Unlock()

	for _, p := range startWith {
		if p != "" {
			m.exclude = append(m.exclude, p)
		}
	}
}

// SetSlowTime set slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func (m *monitor) SetSlowTime(slowTime int32) {
	m.m.Lock()
	defer m.m.Unlock()

	m.slowTime = slowTime
}

// GetSlowTime retrieve the slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func (m *monitor) GetSlowTime() int32 {
	m.m.Lock()
	defer m.m.Unlock()

	return m.slowTime
}

// SetDuration set duration property. duration is used to ginRequestDuration
// metric buckets.
func (m *monitor) SetDuration(duration []float64) {
	m.m.Lock()
	defer m.m.Unlock()

	m.reqDuration = duration
}

// GetDuration retrieve the duration property. duration is used to ginRequestDuration
// metric buckets.
func (m *monitor) GetDuration() []float64 {
	m.m.Lock()
	defer m.m.Unlock()

	return m.reqDuration
}

// GetMetric used to get metric object by metric_name.
func (m *monitor) GetMetric(name string) *metrics {
	m.m.Lock()
	defer m.m.Unlock()

	if a, ok := m.metrics[name]; !ok || a == nil {
		return &metrics{}
	} else if i := a.Load(); i == nil {
		return &metrics{}
	} else if o, ok := i.(*metrics); !ok || o == nil {
		return &metrics{}
	} else {
		return o
	}
}

// SetMetric used to store an atomic value of the metric object by metric_name.
func (m *monitor) SetMetric(metric Metrics) error {
	m.m.Lock()
	defer m.m.Unlock()

	var (
		ok bool
		o  *metrics
	)

	if o, ok = metric.(*metrics); !ok {
		return errors.Errorf("metric is not a valid metric instance")
	}

	if o.n == "" {
		return errors.Errorf("metric name cannot be empty.")
	}

	if o.f == nil {
		return errors.Errorf("metric collect func cannot be empty.")
	}

	if _, ok = m.metrics[o.n]; !ok {
		if err := o.t.Register(o); err != nil {
			return err
		}

		prometheus.MustRegister(o.v)
	}

	if m.metrics[o.n] == nil {
		m.metrics[o.n] = new(atomic.Value)
	}

	m.metrics[o.n].Store(metric)

	return nil
}

// AddMetric add custom monitor metric.
func (m *monitor) AddMetric(metric Metrics) error {
	return m.SetMetric(metric)
}

// ListMetric retrieve a slice of metrics' name registered
func (m *monitor) ListMetric() []string {
	var res = make([]string, 0)

	m.m.Lock()
	defer m.m.Unlock()

	for k := range m.metrics {
		res = append(res, k)
	}

	return res
}
