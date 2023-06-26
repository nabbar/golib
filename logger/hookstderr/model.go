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

package hookstderr

import (
	"context"
	"io"
	"strings"

	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

type hkerr struct {
	w io.Writer
	l []logrus.Level
	f logrus.Formatter
	s bool // Disable Stack
	d bool // Disable Timestamp
	t bool // Disable Trace
	c bool // Disable Color
	a bool // Enable AccessLog
}

func (o *hkerr) getFormatter() logrus.Formatter {
	return o.f
}

func (o *hkerr) Run(ctx context.Context) {
	return
}

func (o *hkerr) Levels() []logrus.Level {
	return o.l
}

func (o *hkerr) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

func (o *hkerr) Fire(entry *logrus.Entry) error {
	ent := entry.Dup()
	ent.Level = entry.Level

	if o.s {
		ent.Data = o.filterKey(ent.Data, logtps.FieldStack)
	}

	if o.d {
		ent.Data = o.filterKey(ent.Data, logtps.FieldTime)
	}

	if !o.t {
		ent.Data = o.filterKey(ent.Data, logtps.FieldCaller)
		ent.Data = o.filterKey(ent.Data, logtps.FieldFile)
		ent.Data = o.filterKey(ent.Data, logtps.FieldLine)
	}

	var (
		p []byte
		e error
	)

	if o.a {
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

	if _, e = o.Write(p); e != nil {
		return e
	}

	return nil
}

func (o *hkerr) filterKey(f logrus.Fields, key string) logrus.Fields {
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
