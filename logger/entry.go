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
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	liberr "github.com/nabbar/golib/errors"
	"github.com/sirupsen/logrus"
)

const (
	FieldTime    = "time"
	FieldLevel   = "level"
	FieldStack   = "stack"
	FieldCaller  = "caller"
	FieldFile    = "file"
	FieldLine    = "line"
	FieldMessage = "message"
	FieldError   = "error"
	FieldData    = "data"
)

type Entry struct {
	log   func() *logrus.Logger
	gin   *gin.Context
	clean bool

	//Time is the time of the event (can be empty time if disabled timestamp)
	Time time.Time `json:"time"`

	//Level define the level of the entry (cannot be empty or nil)
	Level Level `json:"level"`

	//Stack define the process goroutine number (can be 0 if disabled)
	Stack uint64 `json:"stack"`

	//Caller define the function caller of the entry (can be empty if trace disabled, not found or anonymous function)
	Caller string `json:"caller"`

	//File define the file function caller of the entry (can be empty if trace disabled, not found or anonymous function)
	File string `json:"file"`

	//Caller define the line in file caller of the entry (can be 0 if trace disabled, not found or anonymous function)
	Line uint32 `json:"line"`

	//Message define the main message of the entry (can be empty)
	Message string `json:"message"`

	//Error define a slice of error interface (can be nil, or a silce with one or more nil values)
	Error []error `json:"error"`

	//Data is a unknown type data to add to logger (can be nil)
	Data interface{} `json:"data"`

	//Fields are a list of custom information to add to log entry (can be nil or can overwrite Entry values)
	Fields Fields `json:"fields"`
}

// SetGinContext allow to register a gin context pointer to register the errors of the current entry intro gin Context Error Slice.
func (e *Entry) SetGinContext(ctx *gin.Context) *Entry {
	e.gin = ctx
	return e
}

// FieldAdd allow to add one couple key/val as type string/interface into the custom field of the entry.
func (e *Entry) FieldAdd(key string, val interface{}) *Entry {
	e.Fields = e.Fields.Add(key, val)
	return e
}

// FieldMerge allow to merge a Field pointer into the custom field of the entry.
func (e *Entry) FieldMerge(fields Fields) *Entry {
	e.Fields = e.Fields.Merge(fields)
	return e
}

// FieldSet allow to change the custom field of the entry with the given Fields in parameter.
func (e *Entry) FieldSet(fields Fields) *Entry {
	e.Fields = fields
	return e
}

func (e *Entry) FieldClean(keys ...string) *Entry {
	e.Fields = e.Fields.Clean(keys...)
	return e
}

func (e *Entry) DataSet(data interface{}) *Entry {
	e.Data = data
	return e
}

func (e *Entry) ErrorClean() *Entry {
	e.Error = make([]error, 0)
	return e
}

func (e *Entry) ErrorSet(err []error) *Entry {
	e.Error = err
	return e
}

func (e *Entry) ErrorAdd(cleanNil bool, err ...error) *Entry {
	for _, er := range err {
		if cleanNil && er == nil {
			continue
		}
		e.Error = append(e.Error, er)
	}
	return e
}

func (e *Entry) ErrorAddLib(cleanNil bool, err ...liberr.Error) *Entry {
	for _, er := range err {
		e.ErrorAdd(cleanNil, er.GetErrorSlice()...)
	}
	return e
}

func (e *Entry) Check(lvlNoErr Level) bool {
	var found = false
	if len(e.Error) > 0 {
		for _, er := range e.Error {
			if er == nil {
				continue
			}

			found = true
			break
		}
	}

	if !found {
		e.Level = lvlNoErr
	}

	e.Log()
	return found
}

func (e *Entry) _logClean() {
	var (
		log *logrus.Logger
	)

	if e.log == nil {
		return
	} else if log = e.log(); log == nil {
		return
	} else {
		log.Info(e.Message)
	}
}

func (e *Entry) Log() {
	if e.clean {
		e._logClean()
		return
	}

	var (
		ent *logrus.Entry
		tag = NewFields().Add(FieldLevel, e.Level.String())
		log *logrus.Logger
	)

	if !e.Time.IsZero() {
		tag = tag.Add(FieldTime, e.Time.Format(time.RFC3339Nano))
	}

	if e.Stack > 0 {
		tag = tag.Add(FieldStack, e.Stack)
	}

	if e.Caller != "" {
		tag = tag.Add(FieldCaller, e.Caller)
	} else if e.File != "" {
		tag = tag.Add(FieldFile, e.File)
	}

	if e.Line > 0 {
		tag = tag.Add(FieldLine, e.Line)
	}

	if e.Message != "" {
		tag = tag.Add(FieldMessage, e.Message)
	}

	if len(e.Error) > 0 {
		var msg = make([]string, 0)

		for _, er := range e.Error {
			if er == nil {
				continue
			}
			msg = append(msg, er.Error())
		}

		tag = tag.Add(FieldError, strings.Join(msg, ", "))
	}

	if e.Data != nil {
		tag = tag.Add(FieldData, e.Data)
	}

	if len(e.Fields) > 0 {
		tag = tag.Merge(e.Fields)
	}

	if e.log == nil {
		return
	} else if log = e.log(); log == nil {
		return
	} else {
		ent = log.WithFields(tag.Logrus())
	}

	if e.gin != nil && len(e.Error) > 0 {
		for _, err := range e.Error {
			_ = e.gin.Error(err)
		}
	}

	if e.Level == NilLevel {
		return
	}

	ent.Log(e.Level.Logrus())

	if e.Level <= FatalLevel {
		os.Exit(1)
	}
}
