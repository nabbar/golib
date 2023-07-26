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

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	stsctr "github.com/nabbar/golib/status/control"
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
	stt := monsts.OK
	msg := ""
	ign := make([]string, 0)

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		if len(keys) > 0 && !slices.Contains(keys, name) {
			return true
		} else if len(ign) > 0 && slices.Contains(ign, name) {
			return true
		}

		v := val.Status()

		if c := o.cfgGetMode(name); c == stsctr.Ignore {
			return true
		} else if c == stsctr.Should && v < monsts.Warn {
			if stt > monsts.Warn {
				stt = monsts.Warn
				msg = val.Message()
			}
		} else if c == stsctr.One {
			lst := o.cfgGetOne(name)
			sta := monsts.KO
			res := ""

			o.MonitorWalk(func(nme string, val montps.Monitor) bool {
				if !slices.Contains(lst, nme) {
					return true
				}

				ign = append(ign, nme)

				w := val.Status()

				if w > sta {
					sta = w
				} else if w == sta && len(res) < 1 {
					res = val.Message()
				}

				return true
			})

			if stt > sta {
				stt = sta
				msg = res
			}
		} else if stt > v {
			stt = v
			msg = val.Message()
		}

		return true
	})

	return stt, msg
}
