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

	logent "github.com/nabbar/golib/logger/entry"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
)

func (o *logger) Debug(message string, data interface{}, args ...interface{}) {
	o.newEntry(loglvl.DebugLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Info(message string, data interface{}, args ...interface{}) {
	o.newEntry(loglvl.InfoLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Warning(message string, data interface{}, args ...interface{}) {
	o.newEntry(loglvl.WarnLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Error(message string, data interface{}, args ...interface{}) {
	o.newEntry(loglvl.ErrorLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Fatal(message string, data interface{}, args ...interface{}) {
	o.newEntry(loglvl.FatalLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) Panic(message string, data interface{}, args ...interface{}) {
	o.newEntry(loglvl.PanicLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

func (o *logger) LogDetails(lvl loglvl.Level, message string, data interface{}, err []error, fields logfld.Fields, args ...interface{}) {
	o.newEntry(lvl, fmt.Sprintf(message, args...), err, fields, data).Log()
}

func (o *logger) CheckError(lvlKO, lvlOK loglvl.Level, message string, err ...error) bool {
	ent := o.newEntry(lvlKO, message, err, nil, nil)
	return ent.Check(lvlOK)
}

func (o *logger) Entry(lvl loglvl.Level, message string, args ...interface{}) logent.Entry {
	return o.newEntry(lvl, fmt.Sprintf(message, args...), nil, nil, nil)
}

func (o *logger) Access(remoteAddr, remoteUser string, localtime time.Time, latency time.Duration, method, request, proto string, status int, size int64) logent.Entry {
	var msg = fmt.Sprintf("%s - %s [%s] [%s] \"%s %s %s\" %d %d", remoteAddr, remoteUser, localtime.Format(time.RFC1123Z), latency.String(), method, request, proto, status, size)
	return o.newEntryClean(msg)
}

func (o *logger) newEntry(lvl loglvl.Level, message string, err []error, fields logfld.Fields, data interface{}) logent.Entry {
	var (
		ent = logent.New(lvl)
		frm = o.getCaller()
	)

	ent.ErrorSet(err)
	ent.FieldSet(o.GetFields().FieldsClone(nil))
	ent.DataSet(data)
	ent.SetLogger(o.getLogrus)
	ent.SetEntryContext(time.Now(), o.getStack(), frm.Function, frm.File, uint64(frm.Line), message)
	ent.FieldMerge(fields)

	return ent
}

func (o *logger) newEntryClean(message string) logent.Entry {
	return o.newEntry(loglvl.InfoLevel, message, nil, nil, nil).SetMessageOnly(true)
}
