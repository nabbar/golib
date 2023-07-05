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

package status

import (
	"sync"
	"time"

	liberr "github.com/nabbar/golib/errors"

	montps "github.com/nabbar/golib/monitor/types"

	libctx "github.com/nabbar/golib/context"
	monsts "github.com/nabbar/golib/monitor/status"
	"golang.org/x/exp/slices"
)

type fctGetName func() string
type fctGetRelease func() string
type fctGetHash func() string
type fctGetDateBuild func() time.Time

type sts struct {
	m sync.RWMutex
	p montps.FuncPool
	r func() liberr.ReturnGin
	x libctx.Config[string]
	c ch

	fn fctGetName
	fr fctGetRelease
	fh fctGetHash
	fd fctGetDateBuild
}

func (o *sts) checkFunc() bool {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.fn != nil && o.fr != nil && o.fh != nil && o.fd != nil
}

func (o *sts) IsCacheHealthy() bool {
	return o.c.IsCache()
}

func (o *sts) IsHealthy(name ...string) bool {
	s, _ := o.getStatus(name...)
	return s == monsts.OK
}

func (o *sts) getStatus(keys ...string) (monsts.Status, string) {
	if len(keys) < 1 {
		keys = o.cfgGetMandatory()
	}

	o.m.RLock()
	defer o.m.RUnlock()

	s := monsts.OK
	m := ""

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if !slices.Contains(keys, name) {
			return true
		}

		if v := val.Status(); s > v {
			s = v
			m = val.Message()
		}

		return true
	})

	return s, m
}
