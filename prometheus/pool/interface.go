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

package pool

import (
	libctx "github.com/nabbar/golib/context"
	libmet "github.com/nabbar/golib/prometheus/metrics"
)

type FuncGet func(name string) libmet.Metric
type FuncAdd func(metric libmet.Metric) error
type FuncSet func(key string, metric libmet.Metric)
type FuncWalk func(pool MetricPool, key string, val libmet.Metric) bool

type MetricPool interface {
	// Get is used to retrieve the metric instance from the pool.
	Get(name string) libmet.Metric

	// Add is used to register the metric instance into the pool.
	Add(metric libmet.Metric) error

	// Set is used to replace the metric instance into the pool.
	Set(key string, metric libmet.Metric)

	// Del is used to remove the metric instance into the pool.
	Del(key string)

	// List retrieve a slice of metrics' name registered into the pool.
	List() []string

	// Walk run the given function for each stored item into the pool.
	// If the function return false, the process is stopped.
	Walk(fct FuncWalk, limit ...string) bool
}

func New(ctx libctx.FuncContext) MetricPool {
	return &pool{
		p: libctx.NewConfig[string](ctx),
	}
}
