/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package config

import (
	liblog "github.com/nabbar/golib/logger"
	logent "github.com/nabbar/golib/logger/entry"
	loglvl "github.com/nabbar/golib/logger/level"
)

// RegisterDefaultLogger registers the default logger provider function.
// This logger is made available to all components during initialization.
// Components can use this logger for operational logging, error reporting, and debugging.
func (o *model) RegisterDefaultLogger(fct liblog.FuncLog) {
	o.fct.Store(fctLoggerDef, fct)
}

func (o *model) getDefaultLogger() liblog.Logger {
	if i, l := o.fct.Load(fctLoggerDef); !l {
		return nil
	} else if v, k := i.(liblog.FuncLog); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) logEntry(l loglvl.Level, pattern string, args ...interface{}) logent.Entry {
	if i := o.getDefaultLogger(); i != nil {
		return i.Entry(l, pattern, args...)
	} else {
		return logent.New(loglvl.NilLevel)
	}
}
