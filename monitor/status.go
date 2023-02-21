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
	if i, l := o.x.Load(keyStatus); !l {
		return monsts.KO
	} else if v, k := i.(monsts.Status); !k {
		return monsts.KO
	} else {
		return v
	}
}

func (o *mon) Message() string {
	if i, l := o.x.Load(keyMessage); !l {
		return ""
	} else if v, k := i.(error); !k {
		return ""
	} else if v == nil {
		return ""
	} else {
		return v.Error()
	}
}

func (o *mon) IsRise() bool {
	if sts := o.Status(); sts == monsts.OK {
		return false
	} else {
		return o.riseGet() > 0
	}
}

func (o *mon) IsFall() bool {
	if sts := o.Status(); sts == monsts.KO {
		return false
	} else {
		return o.fallGet() > 0
	}
}

func (o *mon) Latency() time.Duration {
	if i, l := o.x.Load(keyMetricLatency); !l {
		return 0
	} else if v, k := i.(time.Duration); !k {
		return 0
	} else {
		return v
	}
}

func (o *mon) Uptime() time.Duration {
	if i, l := o.x.Load(keyMetricUpTime); !l {
		return 0
	} else if v, k := i.(time.Duration); !k {
		return 0
	} else {
		return v
	}
}

func (o *mon) Downtime() time.Duration {
	if i, l := o.x.Load(keyMetricDownTime); !l {
		return 0
	} else if v, k := i.(time.Duration); !k {
		return 0
	} else {
		return v
	}
}

func (o *mon) riseInc() uint8 {
	i := o.riseGet() + 1
	o.x.Store(keyRise, i)
	return i
}

func (o *mon) riseReset() {
	o.x.Delete(keyRise)
}

func (o *mon) riseGet() uint8 {
	if i, l := o.x.Load(keyRise); !l {
		return 0
	} else if v, k := i.(uint8); !k {
		return 0
	} else {
		return v
	}
}

func (o *mon) fallInc() uint8 {
	i := o.fallGet() + 1
	o.x.Store(keyFall, i)
	return i
}

func (o *mon) fallReset() {
	o.x.Delete(keyFall)
}

func (o *mon) fallGet() uint8 {
	if i, l := o.x.Load(keyFall); !l {
		return 0
	} else if v, k := i.(uint8); !k {
		return 0
	} else {
		return v
	}
}

func (o *mon) setStatus(err error, cfg *runCfg) error {
	if err != nil {
		o.x.Store(keyMessage, err)
		o.setStatusFall(cfg)
	} else {
		o.x.Delete(keyMessage)
		o.setStatusRise(cfg)
	}

	return err
}

func (o *mon) mdlStatus(m middleWare) error {
	return o.setStatus(m.Next(), m.Config())
}

func (o *mon) setStatusFall(cfg *runCfg) {
	sts := o.Status()

	if sts == monsts.KO || cfg == nil {
		return
	}

	o.riseReset()
	i := o.fallInc()

	if i > cfg.fallCountKO {
		o.x.Store(keyStatus, monsts.KO)
	} else if i > cfg.fallCountWarn {
		o.x.Store(keyStatus, monsts.Warn)
	} else {
		o.x.Store(keyStatus, monsts.OK)
	}
}

func (o *mon) setStatusRise(cfg *runCfg) {
	sts := o.Status()

	if sts == monsts.OK || cfg == nil {
		return
	}

	o.fallReset()
	i := o.riseInc()

	if i >= cfg.riseCountWarn {
		o.x.Store(keyStatus, monsts.OK)
	} else if i >= cfg.riseCountKO {
		o.x.Store(keyStatus, monsts.Warn)
	} else {
		o.x.Store(keyStatus, monsts.KO)
	}
}
