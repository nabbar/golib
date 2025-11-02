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
	"context"
	"fmt"
	"sync"
	"time"

	libsrv "github.com/nabbar/golib/runner"
)

func (o *hks) Run(ctx context.Context) {
	var (
		s Wrapper
		w = sync.WaitGroup{}
		e error
	)

	defer func() {
		libsrv.RecoveryCaller("golib/logger/hooksyslog/system", recover())
		if s != nil {
			w.Wait()
			_ = s.Close()
		}
		o.r.Store(false)
	}()

	for {
		if s, e = o.getSyslog(); e != nil {
			fmt.Println(e.Error())
		} else {
			break
		}
		time.Sleep(time.Second)
	}

	o.prepareChan()
	//fmt.Printf("starting hook for log syslog '%s'\n", o.getSyslogInfo())
	go func() {
		time.Sleep(100 * time.Millisecond)
		o.r.Store(true)
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case <-o.Done():
			return

		case i := <-o.Data():
			w.Add(1)
			go o.writeWrapper(s, w.Done, i...)
		}
	}
}

func (o *hks) IsRunning() bool {
	return o.r.Load()
}

func (o *hks) writeWrapper(w Wrapper, done func(), d ...data) {
	var err error

	defer func() {
		libsrv.RecoveryCaller("golib/logger/hooksyslog/system", recover())
		done()
	}()

	if w == nil {
		return
	} else if len(d) < 1 {
		return
	}

	for k := range d {
		if len(d[k].p) < 1 {
			continue
		}
		switch d[k].s {
		case SyslogSeverityAlert:
			_, err = w.Panic(d[k].p)
		case SyslogSeverityCrit:
			_, err = w.Fatal(d[k].p)
		case SyslogSeverityErr:
			_, err = w.Error(d[k].p)
		case SyslogSeverityWarning:
			_, err = w.Warning(d[k].p)
		case SyslogSeverityInfo:
			_, err = w.Info(d[k].p)
		case SyslogSeverityDebug:
			_, err = w.Debug(d[k].p)
		default:
			_, err = w.Write(d[k].p)
		}
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
