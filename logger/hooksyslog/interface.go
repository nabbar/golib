/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package hooksyslog

import (
	"sync/atomic"

	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/sirupsen/logrus"
)

type HookSyslog interface {
	logtps.Hook

	Done() <-chan struct{}
	WriteSev(s SyslogSeverity, p []byte) (n int, err error)
}

func New(opt logcfg.OptionsSyslog, format logrus.Formatter) (HookSyslog, error) {
	var (
		LVLs = make([]logrus.Level, 0)
	)

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, loglvl.Parse(ls).Logrus())
		}
	} else {
		LVLs = logrus.AllLevels
	}

	n := &hks{
		s: new(atomic.Value),
		d: new(atomic.Value),
		o: ohks{
			format:           format,
			levels:           LVLs,
			disableStack:     opt.DisableStack,
			disableTimestamp: opt.DisableTimestamp,
			enableTrace:      opt.EnableTrace,
			enableAccessLog:  opt.EnableAccessLog,
			network:          libptc.Parse(opt.Network),
			endpoint:         opt.Host,
			tag:              opt.Tag,
			fac:              MakeFacility(opt.Facility),
			//sev : MakeSeverity(opt.Severity),
		},
	}

	n.s.Store(make(chan struct{}))
	n.d.Store(make(chan data))

	if h, e := n.getSyslog(); e != nil {
		return nil, e
	} else {
		_ = h.Close()
	}

	return n, nil
}
