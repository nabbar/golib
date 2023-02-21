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
	"net/http"
	"strings"
	"sync"

	prmpol "github.com/nabbar/golib/prometheus/pool"
)

// GinPrometheus is an object that uses to set gin server monitor.
type prom struct {
	m sync.RWMutex

	exclude     []string
	slowTime    int32
	reqDuration []float64

	ginMet prmpol.MetricPool
	othMet prmpol.MetricPool
	handle http.Handler
}

// SetSlowTime set slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func (m *prom) SetSlowTime(slowTime int32) {
	m.m.Lock()
	defer m.m.Unlock()

	m.slowTime = slowTime
}

// GetSlowTime retrieve the slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func (m *prom) GetSlowTime() int32 {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.slowTime
}

// SetDuration set duration property. duration is used to ginRequestDuration
// metric buckets.
func (m *prom) SetDuration(duration []float64) {
	m.m.Lock()
	defer m.m.Unlock()

	m.reqDuration = duration
}

// GetDuration retrieve the duration property. duration is used to ginRequestDuration
// metric buckets.
func (m *prom) GetDuration() []float64 {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.reqDuration
}

func (m *prom) ExcludePath(startWith ...string) {
	m.m.Lock()
	defer m.m.Unlock()

	for _, p := range startWith {
		if p != "" {
			m.exclude = append(m.exclude, p)
		}
	}
}

func (m *prom) isExclude(path string) bool {
	m.m.RLock()
	defer m.m.RUnlock()

	if len(m.exclude) < 1 {
		return false
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	for _, p := range m.exclude {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}
