//go:build amd64 || arm64 || arm64be || ppc64 || ppc64le || mips64 || mips64le || riscv64 || s390x || sparc64 || wasm
// +build amd64 arm64 arm64be ppc64 ppc64le mips64 mips64le riscv64 s390x sparc64 wasm

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

package cluster

import (
	"context"

	dgblog "github.com/lni/dragonboat/v3/logger"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
)

const LogLib = "DragonBoat"

func SetLoggerFactory(log liblog.FuncLog) {
	if log == nil {
		log = func() liblog.Logger {
			return liblog.New(context.Background)
		}
	}

	dgblog.SetLoggerFactory(func(pkgName string) dgblog.ILogger {
		return &logDragonBoat{
			pkg: pkgName,
			log: log,
		}
	})
}

type logDragonBoat struct {
	pkg string
	log liblog.FuncLog
}

func (l *logDragonBoat) SetLevel(level dgblog.LogLevel) {
	if l.log == nil {
		return
	}

	switch level {
	case dgblog.CRITICAL:
		l.log().SetLevel(loglvl.FatalLevel)
	case dgblog.ERROR:
		l.log().SetLevel(loglvl.ErrorLevel)
	case dgblog.WARNING:
		l.log().SetLevel(loglvl.WarnLevel)
	case dgblog.INFO:
		l.log().SetLevel(loglvl.InfoLevel)
	case dgblog.DEBUG:
		l.log().SetLevel(loglvl.DebugLevel)
	}
}

func (l *logDragonBoat) logMsg(lvl loglvl.Level, message string, args ...interface{}) {
	if l.log == nil {
		l.log = func() liblog.Logger {
			return liblog.New(context.Background)
		}
	}

	l.log().Entry(lvl, message, args...).FieldAdd("lib", LogLib).FieldAdd("pkg", l.pkg).Log()
}

func (l *logDragonBoat) Debugf(format string, args ...interface{}) {
	l.logMsg(loglvl.DebugLevel, format, args...)
}

func (l *logDragonBoat) Infof(format string, args ...interface{}) {
	l.logMsg(loglvl.InfoLevel, format, args...)
}

func (l *logDragonBoat) Warningf(format string, args ...interface{}) {
	l.logMsg(loglvl.WarnLevel, format, args...)
}

func (l *logDragonBoat) Errorf(format string, args ...interface{}) {
	l.logMsg(loglvl.ErrorLevel, format, args...)
}

func (l *logDragonBoat) Panicf(format string, args ...interface{}) {
	l.logMsg(loglvl.FatalLevel, format, args...)
}
