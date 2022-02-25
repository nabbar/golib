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
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultSlowTime = int32(5)
)

type Prometheus interface {
	Expose(c *gin.Context)
	MiddleWare(c *gin.Context)
	CollectMetrics(c *gin.Context, start time.Time)

	ExcludePath(startWith ...string)

	GetMetric(name string) *metrics
	SetMetric(metric Metrics) error
	AddMetric(metric Metrics) error
	ListMetric() []string

	SetSlowTime(slowTime int32)
	GetSlowTime() int32

	SetDuration(duration []float64)
	GetDuration() []float64
}

// New will return a new object that implement interface GinPrometheus.
func New() Prometheus {
	return &monitor{
		slowTime:    DefaultSlowTime,
		reqDuration: []float64{0.1, 0.3, 1.2, 5, 10},
		metrics:     make(map[string]*atomic.Value),
		exclude:     make([]string, 0),
	}
}
