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

package monitor

import (
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
)

func (o *mon) Name() string {
	return o.getName()
}

func (o *mon) InfoName() string {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.i.Name()
}

func (o *mon) InfoMap() map[string]interface{} {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.i.Info()
}

func (o *mon) InfoGet() montps.Info {
	o.m.Lock()
	defer o.m.Unlock()

	return o.i
}

func (o *mon) InfoUpd(inf montps.Info) {
	o.m.Lock()
	defer o.m.Unlock()

	o.i = inf
}

func (o *mon) Status() monsts.Status {
	return o.getLastCheck().Status()
}

func (o *mon) Message() string {
	if err := o.getLastCheck().Error(); err != nil {
		return err.Error()
	}

	return ""
}

func (o *mon) IsRise() bool {
	return o.getLastCheck().IsRise()
}

func (o *mon) IsFall() bool {
	return o.getLastCheck().IsFall()
}

func (o *mon) Latency() time.Duration {
	return o.getLastCheck().Latency()
}

func (o *mon) Uptime() time.Duration {
	return o.getLastCheck().UpTime()
}

func (o *mon) Downtime() time.Duration {
	return o.getLastCheck().DownTime()
}

func (o *mon) mdlStatus(m middleWare) error {
	ts := time.Now()
	err := m.Next()

	lst := o.getLastCheck()
	lst.setStatus(err, time.Since(ts), m.Config())
	o.setLastCheck(lst)

	return err
}
