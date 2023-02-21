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

func (o *logger) Debug(message string, data interface{}, args ...interface{}) {
	o.newEntry(DebugLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Info(message string, data interface{}, args ...interface{}) {
	o.newEntry(InfoLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Warning(message string, data interface{}, args ...interface{}) {
	o.newEntry(WarnLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Error(message string, data interface{}, args ...interface{}) {
	o.newEntry(ErrorLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Fatal(message string, data interface{}, args ...interface{}) {
	o.newEntry(FatalLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Panic(message string, data interface{}, args ...interface{}) {
	o.newEntry(PanicLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) LogDetails(lvl Level, message string, data interface{}, err []error, fields Fields, args ...interface{}) {
	o.newEntry(lvl, fmt.Sprintf(message, args...), err, fields, data).Log()
}

func (o *logger) CheckError(lvlKO, lvlOK Level, message string, err ...error) bool {
	ent := o.newEntry(lvlKO, message, err, nil, nil)
	return ent.Check(lvlOK)
}

func (o *logger) Entry(lvl Level, message string, args ...interface{}) *Entry {
	return o.newEntry(lvl, fmt.Sprintf(message, args...), nil, nil, nil)
}

func (o *logger) Access(remoteAddr, remoteUser string, localtime time.Time, latency time.Duration, method, request, proto string, status int, size int64) *Entry {
	return o.newEntryClean(fmt.Sprintf("%s - %s [%s] [%s] \"%s %s %s\" %d %d", remoteAddr, remoteUser, localtime.Format(time.RFC1123Z), latency.String(), method, request, proto, status, size))
}

func (o *logger) newEntry(lvl Level, message string, err []error, fields Fields, data interface{}) *Entry {
	opt := o.GetOptions()
	cLv := o.GetLevel()

	var ent = &Entry{
		clean:  false,
		Time:   time.Time{},
		Level:  lvl,
		Stack:  0,
		Caller: "",
		File:   "",
		Line:   0,
		Error:  err,
		Fields: o.GetFields().FieldsClone(nil),
	}

	if cLv == NilLevel || lvl > cLv {
		return ent
	} else {
		ent.log = o.getLogrus
		ent.Message = message
		ent.Data = data
	}

	ent.Fields.Merge(fields)

	if !opt.DisableTimestamp {
		ent.Time = time.Now()
	}

	if !opt.DisableStack {
		ent.Stack = o.getStack()
	}

	if opt.EnableTrace {
		frm := o.getCaller()

		if frm.Function != "" {
			ent.Caller = frm.Function
		}

		if frm.File != "" {
			ent.File = o.filterPath(frm.File)
		}

		if frm.Line > 0 {
			ent.Line = uint32(frm.Line)
		}
	}

	return ent
}

func (o *logger) newEntryClean(message string) *Entry {
	var ent = &Entry{
		log:     o.getLogrus,
		clean:   true,
		Message: message,
		Fields:  NewFields(o.x.GetContext),
	}

	return ent
}
