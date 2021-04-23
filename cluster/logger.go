//+build amd64 arm64 arm64be ppc64 ppc64le mips64 mips64le riscv64 s390x sparc64 wasm

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
	dgblog "github.com/lni/dragonboat/v3/logger"
	liblog "github.com/nabbar/golib/logger"
)

func init() {
	dgblog.SetLoggerFactory(func(pkgName string) dgblog.ILogger {
		return &logDragonBoart{
			pkg: pkgName,
		}
	})

}

type logDragonBoart struct {
	pkg string
}

func (l *logDragonBoart) SetLevel(level dgblog.LogLevel) {
}

func (l *logDragonBoart) Debugf(format string, args ...interface{}) {
	var newArg = append(make([]interface{}, 0), l.pkg)
	liblog.DebugLevel.Logf("[DragonBoat: %s] "+format, append(newArg, args...)...)
}

func (l *logDragonBoart) Infof(format string, args ...interface{}) {
	var newArg = append(make([]interface{}, 0), l.pkg)
	liblog.InfoLevel.Logf("[DragonBoat: %s] "+format, append(newArg, args...)...)
}

func (l *logDragonBoart) Warningf(format string, args ...interface{}) {
	var newArg = append(make([]interface{}, 0), l.pkg)
	liblog.WarnLevel.Logf("[DragonBoat: %s] "+format, append(newArg, args...)...)
}

func (l *logDragonBoart) Errorf(format string, args ...interface{}) {
	var newArg = append(make([]interface{}, 0), l.pkg)
	liblog.ErrorLevel.Logf("[DragonBoat: %s] "+format, append(newArg, args...)...)
}

func (l *logDragonBoart) Panicf(format string, args ...interface{}) {
	var newArg = append(make([]interface{}, 0), l.pkg)
	liblog.FatalLevel.Logf("[DragonBoat: %s] "+format, append(newArg, args...)...)
}
