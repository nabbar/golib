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

package logger

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	jww "github.com/spf13/jwalterweatherman"
)

func (o *logger) GetStdLogger(lvl Level, logFlags int) *log.Logger {
	o.SetIOWriterLevel(lvl)
	return log.New(o, "", logFlags)
}

func (o *logger) SetStdLogger(lvl Level, logFlags int) {
	o.SetIOWriterLevel(lvl)
	log.SetOutput(o)
	log.SetPrefix("")
	log.SetFlags(logFlags)
}

func (o *logger) SetSPF13Level(lvl Level, log *jww.Notepad) {
	var (
		fOutLog func(handle io.Writer)
		fLvl    func(threshold jww.Threshold)
	)

	if log == nil {
		jww.SetStdoutOutput(io.Discard)
		fOutLog = jww.SetLogOutput
		fLvl = jww.SetLogThreshold
	} else {
		fOutLog = log.SetLogOutput
		fLvl = log.SetLogThreshold
	}

	switch lvl {
	case NilLevel:
		fOutLog(io.Discard)
		fLvl(jww.LevelCritical)

	case DebugLevel:
		fOutLog(o)
		if opt := o.GetOptions(); opt.EnableTrace {
			fLvl(jww.LevelTrace)
		} else {
			fLvl(jww.LevelDebug)
		}

	case InfoLevel:
		fOutLog(o)
		fLvl(jww.LevelInfo)
	case WarnLevel:
		fOutLog(o)
		fLvl(jww.LevelWarn)
	case ErrorLevel:
		fOutLog(o)
		fLvl(jww.LevelError)
	case FatalLevel:
		fOutLog(o)
		fLvl(jww.LevelFatal)
	case PanicLevel:
		fOutLog(o)
		fLvl(jww.LevelCritical)
	}
}

func (o *logger) SetHashicorpHCLog() {
	hclog.SetDefault(&_hclog{
		l: o,
	})
}

func (o *logger) NewHashicorpHCLog() hclog.Logger {
	return &_hclog{
		l: o,
	}
}
