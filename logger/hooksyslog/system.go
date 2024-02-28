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
	"os"
	"sync"
	"time"

	libsrv "github.com/nabbar/golib/server"
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

	for {
		select {
		case <-ctx.Done():
			return

		case <-o.Done():
			return

		case i := <-o.Data():
			w.Add(1)
			go o.writeWrapper(s, i, w.Done)
		}
	}
}

func (o *hks) writeWrapper(w Wrapper, d data, done func()) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			_, _ = fmt.Fprintf(os.Stderr, "recovering panic thread on writeWrapper function in golib/logger/hooksyslog/system\n%v\n", rec)
		}
		done()
	}()

	if w == nil {
		return
	} else if len(d.p) < 1 {
		return
	}

	switch d.s {
	case SyslogSeverityAlert:
		_, err = w.Panic(d.p)
	case SyslogSeverityCrit:
		_, err = w.Fatal(d.p)
	case SyslogSeverityErr:
		_, err = w.Error(d.p)
	case SyslogSeverityWarning:
		_, err = w.Warning(d.p)
	case SyslogSeverityInfo:
		_, err = w.Info(d.p)
	case SyslogSeverityDebug:
		_, err = w.Debug(d.p)
	default:
		_, err = w.Write(d.p)
	}

	if err != nil {
		fmt.Println(err.Error())
	}

}
