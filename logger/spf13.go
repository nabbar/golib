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

	loglvl "github.com/nabbar/golib/logger/level"
	jww "github.com/spf13/jwalterweatherman"
)

func (o *logger) SetSPF13Level(lvl loglvl.Level, log *jww.Notepad) {
	if log == nil {
		jww.SetStdoutOutput(io.Discard)
	} else {
		jww.SetStdoutOutput(o)
	}

	switch lvl {
	case loglvl.NilLevel:
		jww.SetLogOutput(io.Discard)
		jww.SetLogThreshold(jww.LevelCritical)
	case loglvl.DebugLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelTrace)
	case loglvl.InfoLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelInfo)
	case loglvl.WarnLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelWarn)
	case loglvl.ErrorLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelError)
	case loglvl.FatalLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelFatal)
	case loglvl.PanicLevel:
		jww.SetLogOutput(o)
		jww.SetLogThreshold(jww.LevelCritical)
	}
}
