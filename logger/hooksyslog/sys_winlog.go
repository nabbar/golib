//go:build windows
// +build windows

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
	"time"

	libptc "github.com/nabbar/golib/network/protocol"

	"golang.org/x/sys/windows/svc/eventlog"
)

var _registred *atomic.Value

func init() {
	// register a function to clean event on stopping application to clean Windows Registry Database
	go func() {
		// clean all registered event
		defer func() {
			if _registred == nil {
				_registred = new(atomic.Value)
			}

			var (
				i  interface{}
				o  []string
				ok bool
			)

			if i = _registred.Load(); i == nil {
				i = make([]string, 0)
			}

			if o, ok = i.([]string); !ok {
				o = make([]string, 0)
			}

			_registred.Store(make([]string, 0))

			for _, s := range o {
				if err := eventlog.Remove(s); err != nil {
					println(err)
				}
			}
		}()

		for {
			// only wait the stopping process
			time.Sleep(200 * time.Millisecond)
		}
	}()
}

func windowsRegister(source string) error {
	if _registred == nil {
		_registred = new(atomic.Value)
	}

	var (
		i  interface{}
		o  []string
		ok bool
	)

	if i = _registred.Load(); i == nil {
		i = make([]string, 0)
	}

	if o, ok = i.([]string); !ok {
		o = make([]string, 0)
	}

	for _, s := range o {
		if s == source {
			return nil
		}
	}

	if err := eventlog.InstallAsEventCreate(source, eventlog.Error|eventlog.Warning|eventlog.Info); err != nil {
		return err
	}

	o = append(o, source)
	_registred.Store(o)

	return nil
}

const (
	_ErrorId uint32 = iota + 1
	_WarningId
	_InfoId
)

type _WinLog struct {
	r bool
	s string
	w *eventlog.Log
}

func newSyslog(net libptc.NetworkProtocol, host, tag string, facility SyslogFacility) (Wrapper, error) {
	var (
		sys *eventlog.Log
		err error
	)

	if net != libptc.NetworkEmpty {
		sys, err = eventlog.OpenRemote(host, tag)
	} else {
		if err = windowsRegister(tag); err != nil {
			println(err.Error())
		}

		sys, err = eventlog.Open(tag)
	}

	if err != nil {
		return nil, err
	}

	return &_WinLog{
		r: net != libptc.NetworkEmpty,
		s: tag,
		w: sys,
	}, nil
}

func (o *_WinLog) Close() error {
	var err error

	err = o.w.Close()
	o.w = nil

	if err != nil {
		return err
	}

	return nil
}

func (o *_WinLog) Write(p []byte) (n int, err error) {
	return o.Info(p)
}

func (o *_WinLog) Panic(p []byte) (n int, err error) {
	return o.Error(p)
}

func (o *_WinLog) Fatal(p []byte) (n int, err error) {
	return o.Error(p)
}

func (o *_WinLog) Error(p []byte) (n int, err error) {
	err = o.w.Error(_ErrorId*100, string(p))
	return len(p), err
}

func (o *_WinLog) Warning(p []byte) (n int, err error) {
	err = o.w.Warning(_WarningId*100, string(p))
	return len(p), err
}

func (o *_WinLog) Info(p []byte) (n int, err error) {
	err = o.w.Info(_InfoId*100, string(p))
	return len(p), err
}

func (o *_WinLog) Debug(p []byte) (n int, err error) {
	return o.Info(p)
}
