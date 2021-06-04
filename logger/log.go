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
	"fmt"
	"time"
)

func (l *logger) Debug(message string, data interface{}, args ...interface{}) {
	l.newEntry(DebugLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (l *logger) Info(message string, data interface{}, args ...interface{}) {
	l.newEntry(InfoLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (l *logger) Warning(message string, data interface{}, args ...interface{}) {
	l.newEntry(WarnLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (l *logger) Error(message string, data interface{}, args ...interface{}) {
	l.newEntry(ErrorLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (l *logger) Fatal(message string, data interface{}, args ...interface{}) {
	l.newEntry(FatalLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (l *logger) Panic(message string, data interface{}, args ...interface{}) {
	l.newEntry(PanicLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (l *logger) LogDetails(lvl Level, message string, data interface{}, err []error, fields Fields, args ...interface{}) {
	l.newEntry(lvl, fmt.Sprintf(message, args...), err, fields, data).Log()
}

func (l *logger) CheckError(lvlKO, lvlOK Level, message string, err ...error) bool {
	ent := l.newEntry(lvlKO, message, err, nil, nil)
	return ent.Check(lvlOK)
}

func (l *logger) Entry(lvl Level, message string, args ...interface{}) *Entry {
	return l.newEntry(lvl, fmt.Sprintf(message, args...), nil, nil, nil)
}

func (l *logger) newEntry(lvl Level, message string, err []error, fields Fields, data interface{}) *Entry {
	opt := l.GetOptions()
	cLv := l.GetLevel()

	if cLv == NilLevel || lvl > cLv {
		return &Entry{}
	}

	var ent = &Entry{
		log:     l.getLog,
		Time:    time.Time{},
		Level:   lvl,
		Stack:   0,
		Caller:  "",
		File:    "",
		Line:    0,
		Message: message,
		Error:   err,
		Data:    data,
		Fields:  NewFields().Merge(l.GetFields()).Merge(fields),
	}

	if !opt.DisableTimestamp {
		ent.Time = time.Now()
	}

	if !opt.DisableStack {
		ent.Stack = l.getStack()
	}

	if opt.EnableTrace {
		frm := l.getCaller()

		if frm.Function != "" {
			ent.Caller = frm.Function
		}

		if frm.File != "" {
			ent.File = l.filterPath(frm.File)
		}

		if frm.Line > 0 {
			ent.Line = uint32(frm.Line)
		}
	}

	return ent
}
