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
	"context"
	"sync"

	ginsdk "github.com/gin-gonic/gin"
	libctx "github.com/nabbar/golib/context"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmpol "github.com/nabbar/golib/prometheus/pool"
)

const (
	DefaultSlowTime = int32(5)
)

type FuncGetPrometheus func() Prometheus
type FuncCollectMetrics func(ctx context.Context, name ...string)

type MetricsCollection interface {
	// GetMetric is used to retrieve the metric instance from prometheus instance.
	GetMetric(name string) libmet.Metric

	// AddMetric is used to register the metric instance into prometheus instance.
	AddMetric(isAPI bool, metric libmet.Metric) error

	// DelMetric is used to unregister the metric instance into prometheus instance.
	DelMetric(name string)

	// ListMetric retrieve a slice of ginMet' name registered for all type API or not.
	ListMetric() []string
}

type GinRoute interface {
	Expose(ctx context.Context)
	ExposeGin(c *ginsdk.Context)

	MiddleWare(ctx context.Context)
	MiddleWareGin(c *ginsdk.Context)

	ExcludePath(startWith ...string)
}

type Collect interface {
	libmet.Collect
	CollectMetrics(ctx context.Context, name ...string)
}

type Prometheus interface {
	GinRoute
	MetricsCollection
	Collect

	SetSlowTime(slowTime int32)
	GetSlowTime() int32

	SetDuration(duration []float64)
	GetDuration() []float64
}

// New will return a new object that implement interface GinPrometheus.
func New(ctx libctx.FuncContext) Prometheus {
	return &prom{
		m:           sync.RWMutex{},
		exclude:     make([]string, 0),
		slowTime:    DefaultSlowTime,
		reqDuration: []float64{0.1, 0.3, 1.2, 5, 10},
		ginMet:      prmpol.New(ctx),
		othMet:      prmpol.New(ctx),
	}
}
