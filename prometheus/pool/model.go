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
	"fmt"

	libctx "github.com/nabbar/golib/context"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// pool is the internal implementation of the MetricPool interface.
// It provides thread-safe metric storage using github.com/nabbar/golib/context.
type pool struct {
	p libctx.Config[string] // Thread-safe storage for metrics
	r prmsdk.Registerer     // Prometheus registerer for metric registration and unregistration
}

// Get retrieves a metric from the pool by name.
// Returns nil if the metric doesn't exist or is not a valid Metric type.
func (p *pool) Get(name string) libmet.Metric {
	if i, l := p.p.Load(name); !l {
		return nil
	} else if v, k := i.(libmet.Metric); !k {
		return nil
	} else {
		return v
	}
}

// Add validates and registers a metric in the pool.
//
// This method performs the following steps:
//  1. Validates that the metric has a non-empty name
//  2. Validates that the metric has a collection function
//  3. Creates the appropriate Prometheus collector based on metric type
//  4. Registers the collector with Prometheus
//  5. Stores the metric in the pool
//
// Returns an error if any validation or registration step fails.
func (p *pool) Add(metric libmet.Metric) error {
	if metric.GetName() == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	if metric.GetCollect() == nil {
		return fmt.Errorf("metric collect func cannot be empty")
	}

	if v, e := metric.GetType().Register(metric); e != nil {
		return e
	} else if e = metric.Register(p.r, v); e != nil {
		return e
	}

	p.p.Store(metric.GetName(), metric)
	return nil
}

// Set stores a metric in the pool without validation or registration.
//
// This is a low-level method that bypasses all validation and registration.
// Use Add for normal metric registration.
func (p *pool) Set(key string, metric libmet.Metric) {
	p.p.Store(key, metric)
}

// Del removes a metric from the pool and unregisters it from Prometheus.
//
// If the metric exists and is valid, it will be unregistered from Prometheus
// before being removed from the pool.
func (p *pool) Del(key string) error {
	if i, l := p.p.LoadAndDelete(key); !l {
		return nil
	} else if m, k := i.(libmet.Metric); !k {
		return nil
	} else {
		return m.UnRegister(p.r)
	}
}

// List returns the names of all metrics currently in the pool.
//
// The returned slice is a snapshot and will not reflect subsequent changes.
func (p *pool) List() []string {
	var res = make([]string, 0)

	p.p.Walk(func(key string, val interface{}) bool {
		res = append(res, key)
		return true
	})

	return res
}

// Walk iterates over metrics in the pool, calling the provided function for each.
//
// If a stored value is not a valid Metric, it is skipped and iteration continues.
// The function can return false to stop iteration early.
//
// If limit keys are provided, only those specific metrics will be visited.
func (p *pool) Walk(fct FuncWalk, limit ...string) {
	f := func(key string, val interface{}) bool {
		if v, k := val.(libmet.Metric); k {
			return fct(p, key, v)
		}
		return true
	}

	p.p.WalkLimit(f, limit...)
}

// Clear removes all metrics from the pool and unregisters them from Prometheus.
//
// This method iterates through all metrics, unregisters each from Prometheus,
// and removes it from the pool storage. Any errors during unregistration are
// collected and returned.
//
// Thread-safe: uses WalkLimit for safe concurrent iteration.
func (p *pool) Clear() []error {
	var err = make([]error, 0)

	p.p.WalkLimit(func(key string, val interface{}) bool {
		if v, k := val.(libmet.Metric); v != nil && k {
			if e := v.UnRegister(p.r); e != nil {
				err = append(err, e)
			}
		}
		p.p.Delete(key)
		return true
	})

	return err
}
