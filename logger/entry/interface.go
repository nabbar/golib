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

package entry

import (
	"time"

	ginsdk "github.com/gin-gonic/gin"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	"github.com/sirupsen/logrus"
)

type Entry interface {
	SetLogger(fct func() *logrus.Logger) Entry
	SetLevel(lvl loglvl.Level) Entry
	SetMessageOnly(flag bool) Entry
	SetEntryContext(etime time.Time, stack uint64, caller, file string, line uint64, msg string) Entry
	SetGinContext(ctx *ginsdk.Context) Entry

	DataSet(data interface{}) Entry
	Check(lvlNoErr loglvl.Level) bool
	Log()

	FieldAdd(key string, val interface{}) Entry
	FieldMerge(fields logfld.Fields) Entry
	FieldSet(fields logfld.Fields) Entry
	FieldClean(keys ...string) Entry

	ErrorClean() Entry
	ErrorSet(err []error) Entry
	ErrorAdd(cleanNil bool, err ...error) Entry
}

func New(lvl loglvl.Level) Entry {
	return &entry{
		log:    nil,
		gin:    nil,
		clean:  false,
		Level:  lvl,
		Time:   time.Now(),
		Error:  make([]error, 0),
		Data:   nil,
		Fields: nil,
	}
}
