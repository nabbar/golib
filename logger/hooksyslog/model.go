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
	"strings"
	"sync/atomic"

	logtps "github.com/nabbar/golib/logger/types"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/sirupsen/logrus"
)

type ohks struct {
	format           logrus.Formatter
	levels           []logrus.Level
	disableStack     bool
	disableTimestamp bool
	enableTrace      bool
	enableAccessLog  bool

	network  libptc.NetworkProtocol
	endpoint string

	tag string
	fac SyslogFacility
	//	Sev SyslogSeverity
}

type hks struct {
	s *atomic.Value // channel stop struct{}
	d *atomic.Value // channel data []byte
	o ohks          // config data
	r *atomic.Bool  // running status
}

func (o *hks) Levels() []logrus.Level {
	return o.getLevel()
}

func (o *hks) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

func (o *hks) Fire(entry *logrus.Entry) error {
	ent := entry.Dup()
	ent.Level = entry.Level

	if o.getDisableStack() {
		ent.Data = o.filterKey(ent.Data, logtps.FieldStack)
	}

	if o.getDisableTimestamp() {
		ent.Data = o.filterKey(ent.Data, logtps.FieldTime)
	}

	if !o.getEnableTrace() {
		ent.Data = o.filterKey(ent.Data, logtps.FieldCaller)
		ent.Data = o.filterKey(ent.Data, logtps.FieldFile)
		ent.Data = o.filterKey(ent.Data, logtps.FieldLine)
	}

	var (
		p []byte
		e error
	)

	if o.getEnableAccessLog() {
		if len(entry.Message) > 0 {
			if !strings.HasSuffix(entry.Message, "\n") {
				entry.Message += "\n"
			}
			p = []byte(entry.Message)
		} else {
			return nil
		}
	} else {
		if len(ent.Data) < 1 {
			return nil
		}

		if f := o.getFormatter(); f != nil {
			p, e = f.Format(ent)
		} else {
			p, e = ent.Bytes()
		}

		if e != nil {
			return e
		}
	}

	switch ent.Level {
	case logrus.PanicLevel:
		_, e = o.WriteSev(SyslogSeverityAlert, p)
	case logrus.FatalLevel:
		_, e = o.WriteSev(SyslogSeverityCrit, p)
	case logrus.ErrorLevel:
		_, e = o.WriteSev(SyslogSeverityErr, p)
	case logrus.WarnLevel:
		_, e = o.WriteSev(SyslogSeverityWarning, p)
	case logrus.InfoLevel:
		_, e = o.WriteSev(SyslogSeverityInfo, p)
	case logrus.DebugLevel:
		_, e = o.WriteSev(SyslogSeverityDebug, p)
	default:
		_, e = o.Write(p)
	}
	if e != nil {
		return e
	}

	return nil
}

func (o *hks) filterKey(f logrus.Fields, key string) logrus.Fields {
	if len(f) < 1 {
		return f
	}

	if _, ok := f[key]; !ok {
		return f
	} else {
		delete(f, key)
		return f
	}
}
