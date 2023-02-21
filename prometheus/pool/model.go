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
)

type pool struct {
	p libctx.Config[string]
}

func (p *pool) Get(name string) libmet.Metric {
	if i, l := p.p.Load(name); !l {
		return nil
	} else if v, k := i.(libmet.Metric); !k {
		return nil
	} else {
		return v
	}
}

func (p *pool) Add(metric libmet.Metric) error {
	if metric.GetName() == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	if metric.GetCollect() == nil {
		return fmt.Errorf("metric collect func cannot be empty")
	}

	if v, e := metric.GetType().Register(metric); e != nil {
		return e
	} else if e = metric.Register(v); e != nil {
		return e
	}

	p.p.Store(metric.GetName(), metric)
	return nil
}

func (p *pool) Set(key string, metric libmet.Metric) {
	p.p.Store(key, metric)
}

func (p *pool) Del(key string) {
	if i, l := p.p.LoadAndDelete(key); !l {
		return
	} else if m, k := i.(libmet.Metric); !k {
		return
	} else {
		m.UnRegister()
	}
}

func (p *pool) List() []string {
	var res = make([]string, 0)

	p.p.Walk(func(key string, val interface{}) bool {
		res = append(res, key)
		return true
	})

	return res
}

func (p *pool) Walk(fct FuncWalk, limit ...string) bool {
	f := func(key string, val interface{}) bool {
		if v, k := val.(libmet.Metric); k {
			return fct(p, key, v)
		}
		return true
	}

	return p.p.WalkLimit(f, limit...)
}
