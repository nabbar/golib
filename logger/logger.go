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

func (l *logger) GetStdLogger(lvl Level, logFlags int) *log.Logger {
	l.SetIOWriterLevel(lvl)
	return log.New(l, "", logFlags)
}

func (l *logger) SetStdLogger(lvl Level, logFlags int) {
	l.SetIOWriterLevel(lvl)
	log.SetOutput(l)
	log.SetPrefix("")
	log.SetFlags(logFlags)
}

func (l *logger) SetSPF13Level(lvl Level, log *jww.Notepad) {
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
		fOutLog(l)
		if opt := l.GetOptions(); opt.EnableTrace {
			fLvl(jww.LevelTrace)
		} else {
			fLvl(jww.LevelDebug)
		}

	case InfoLevel:
		fOutLog(l)
		fLvl(jww.LevelInfo)
	case WarnLevel:
		fOutLog(l)
		fLvl(jww.LevelWarn)
	case ErrorLevel:
		fOutLog(l)
		fLvl(jww.LevelError)
	case FatalLevel:
		fOutLog(l)
		fLvl(jww.LevelFatal)
	case PanicLevel:
		fOutLog(l)
		fLvl(jww.LevelCritical)
	}
}

func (l *logger) SetHashicorpHCLog() {
	hclog.SetDefault(&_hclog{
		l: l,
	})
}

func (l *logger) NewHashicorpHCLog() hclog.Logger {
	return &_hclog{
		l: l,
	}
}
