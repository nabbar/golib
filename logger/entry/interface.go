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
	// SetLogger sets the logger function of the entry. The logger function must return
	// a pointer to a logrus.Logger object. If the logger function is nil, the
	// entry will not log anything.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.SetLogger(func() *logrus.Logger { return logrus.New() })
	SetLogger(fct func() *logrus.Logger) Entry
	// SetLevel sets the level of the entry. The level determines when the entry is
	// logged. If the level is lower than the logger's level, the entry will not
	// be logged. If the level is loglvl.NilLevel, the entry will not be logged
	// regardless of the logger's level.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.SetLevel(loglvl.WarnLevel)
	SetLevel(lvl loglvl.Level) Entry
	// SetMessageOnly sets a flag to log only the message of the entry. If the
	// flag is true, the entry will only log the message. If the flag is
	// false, the entry will log all the fields and the message.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.SetMessageOnly(true)
	SetMessageOnly(flag bool) Entry
	// SetEntryContext sets the context fields of the entry. The context fields are
	// logged together with the message. The context fields are: time, stack,
	// caller, file, line, and message.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.SetEntryContext(time.Now(), 123, "caller", "file.go", 456, "message")
	SetEntryContext(etime time.Time, stack uint64, caller, file string, line uint64, msg string) Entry
	// SetGinContext sets the gin context pointer of the entry. The gin context pointer
	// is used to register the errors of the current entry into the gin context
	// error slice. If the gin context pointer is nil, the entry will not register any
	// errors into the gin context error slice.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.SetGinContext(ginContext)
	//
	// This function is useful when you want to log errors of your application and
	// register them into the gin context error slice.
	SetGinContext(ctx *ginsdk.Context) Entry

	// DataSet sets the data of the entry. The data is logged together with the
	// context fields and the message. The data can be any type that can be
	// marshaled into a JSON object. If the data is nil, no data will be logged.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.DataSet(map[string]string{"key1": "value1", "key2": "value2"})
	//
	// This function is useful when you want to log additional data of your application
	// and you want to log it in a structured way.
	DataSet(data interface{}) Entry
	// Check returns true if the level of the entry is greater than or equal to
	// lvlNoErr, and false otherwise.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   result := e.Check(loglvl.WarnLevel)
	//
	// This function is useful when you want to check if the entry should be logged
	// based on the level of the entry and the logger's level.
	Check(lvlNoErr loglvl.Level) bool
	// Log logs the entry into the logger. The entry is logged together with the
	// context fields and the message. If the logger is nil, the entry will not
	// be logged. If the logger's level is lower than the entry's level, the entry
	// will not be logged. If the logger's level is loglvl.NilLevel, the entry will
	// not be logged regardless of the entry's level.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.SetLogger(func() *logrus.Logger { return logrus.New() })
	//   e.Log()
	Log()

	// FieldAdd adds a new field to the entry. The field is a couple of a key
	// (string) and a value (interface{}). The value can be of any type that
	// can be marshaled into a JSON object. The key must be unique, if the
	// key already exists, it will be overwritten with the new value.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.FieldAdd("key", "value")
	//
	// This function is useful when you want to log additional data of your
	// application and you want to log it in a structured way.
	FieldAdd(key string, val interface{}) Entry
	// FieldMerge merges the fields of another logfld.Fields object into the
	// current entry. The fields are merged in a shallow way, meaning
	// that if the key already exists in the current entry, it will be
	// overwritten with the new value from the other logfld.Fields object.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   fields1 := logfld.New(nil)
	//   fields1.Add("key1", "value1")
	//   fields2 := logfld.New(nil)
	//   fields2.Add("key2", "value2")
	//   e = e.FieldMerge(fields1)
	//   e = e.FieldMerge(fields2)
	//
	// This function is useful when you want to log additional data of your
	// application and you want to log it in a structured way.
	FieldMerge(fields logfld.Fields) Entry
	// FieldSet sets the fields of the entry. The fields are a map of key-value
	// pairs where the key is a string and the value is an interface{}.
	// The fields are logged together with the context fields and the message.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   fields := logfld.New(nil)
	//   fields.Add("key1", "value1")
	//   fields.Add("key2", 123)
	//   e = e.FieldSet(fields)
	//
	// This function is useful when you want to log additional data of your
	// application and you want to log it in a structured way.
	FieldSet(fields logfld.Fields) Entry
	// FieldClean removes the fields specified by the keys from the entry. If
	// the key does not exist, it is ignored. If the keys argument is empty,
	// this function does nothing and returns the entry unchanged.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   fields := logfld.New(nil)
	//   fields.Add("key1", "value1")
	//   fields.Add("key2", "value2")
	//   e = e.FieldSet(fields)
	//   e = e.FieldClean("key1")
	//
	// This function is useful when you want to log additional data of your
	// application and you want to log it in a structured way.
	FieldClean(keys ...string) Entry

	// ErrorClean removes all errors from the entry. If the entry has no
	// errors, this function does nothing and returns the entry unchanged.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.ErrorAdd(errors.New("error1"), errors.New("error2"))
	//   e = e.ErrorClean()
	//
	// This function is useful when you want to clear all errors from the entry
	// and you want to log the entry in a structured way.
	ErrorClean() Entry
	// ErrorSet sets the errors of the entry. The errors are logged together with
	// the context fields and the message. If the errors argument is empty,
	// this function does nothing and returns the entry unchanged.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.ErrorSet([]error{errors.New("error1"), errors.New("error2")})
	//
	// This function is useful when you want to log errors of your application
	// and you want to log them in a structured way.
	ErrorSet(err []error) Entry
	// ErrorAdd adds the errors to the entry. The errors are logged together with
	// the context fields and the message. If the cleanNil argument is true, all
	// nil errors are removed from the entry. If the cleanNil argument is false,
	// all errors are added to the entry regardless of whether they are nil or not.
	//
	// Example:
	//   e := New(loglvl.InfoLevel)
	//   e = e.ErrorAdd(true, errors.New("error1"), nil)
	//
	// This function is useful when you want to log errors of your application
	// and you want to log them in a structured way.
	ErrorAdd(cleanNil bool, err ...error) Entry
}

// New returns a new Entry with the given level. The new Entry will be set with
// the current time and an empty error slice. The new Entry will also have
// a nil logger and gin context, and will not be cleaned. The new Entry will
// have no additional data and will have no additional fields.
//
// Example:
//
//	e := New(loglvl.InfoLevel)
//	e = e.ErrorAdd(errors.New("error1"), errors.New("error2"))
//
// This function is useful when you want to log messages of your application
// and you want to log them in a structured way.
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
