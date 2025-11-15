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
	"os"
	"strings"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

type entry struct {
	log   func() *logrus.Logger
	gin   *ginsdk.Context
	clean bool

	//Time is the time of the event (can be empty time if disabled timestamp)
	Time time.Time `json:"time"`

	//Level define the level of the entry (cannot be empty or nil)
	Level loglvl.Level `json:"level"`

	//Stack define the process goroutine number (can be 0 if disabled)
	Stack uint64 `json:"stack"`

	//Caller define the function caller of the entry (can be empty if trace disabled, not found or anonymous function)
	Caller string `json:"caller"`

	//File define the file function caller of the entry (can be empty if trace disabled, not found or anonymous function)
	File string `json:"file"`

	//Caller define the line in file caller of the entry (can be 0 if trace disabled, not found or anonymous function)
	Line uint64 `json:"line"`

	//Message define the main message of the entry (can be empty)
	Message string `json:"message"`

	//Error define a slice of error interface (can be nil, or a silce with one or more nil values)
	Error []error `json:"error"`

	//Data is a unknown type data to add to logger (can be nil)
	Data interface{} `json:"data"`

	//Fields are a list of custom information to add to log entry (can be nil or can overwrite Entry values)
	Fields logfld.Fields `json:"fields"`
}

func (e *entry) SetEntryContext(etime time.Time, stack uint64, caller, file string, line uint64, msg string) Entry {
	if e == nil {
		return nil
	}

	e.Time = etime
	e.Stack = stack
	e.Caller = caller
	e.File = file
	e.Line = line
	e.Message = msg

	return e
}

func (e *entry) SetMessageOnly(flag bool) Entry {
	if e == nil {
		return nil
	}

	e.clean = flag
	return e
}

func (e *entry) SetLevel(lvl loglvl.Level) Entry {
	if e == nil {
		return nil
	}

	e.Level = lvl

	return e
}

func (e *entry) SetLogger(fct func() *logrus.Logger) Entry {
	if e == nil {
		return nil
	}

	e.log = fct
	return e
}

// SetGinContext allow to register a gin context pointer to register the errors of the current entry intro gin Context Error Slice.
func (e *entry) SetGinContext(ctx *ginsdk.Context) Entry {
	if e == nil {
		return nil
	}

	e.gin = ctx
	return e
}

func (e *entry) DataSet(data interface{}) Entry {
	if e == nil {
		return nil
	}

	e.Data = data
	return e
}

func (e *entry) Check(lvlNoErr loglvl.Level) bool {
	if e == nil {
		return false
	}

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

func (e *entry) Log() {
	if e == nil {
		return
	} else if e.log == nil {
		return
	} else if e.Fields == nil {
		return
	} else if e.Fields.Err() != nil {
		return
	} else if e.clean {
		e._logClean()
		return
	} else if e.gin != nil && len(e.Error) > 0 {
		for _, err := range e.Error {
			_ = e.gin.Error(err)
		}
	}

	if e.Level == loglvl.NilLevel {
		return
	}

	var (
		ent *logrus.Entry
		tag = logfld.New(e.Fields).Add(logtps.FieldLevel, e.Level.String())
		log *logrus.Logger
	)

	if !e.Time.IsZero() {
		tag = tag.Add(logtps.FieldTime, e.Time.Format(time.RFC3339Nano))
	}

	if e.Stack > 0 {
		tag = tag.Add(logtps.FieldStack, e.Stack)
	}

	if e.Caller != "" {
		tag = tag.Add(logtps.FieldCaller, e.Caller)
	} else if e.File != "" {
		tag = tag.Add(logtps.FieldFile, e.File)
	}

	if e.Line > 0 {
		tag = tag.Add(logtps.FieldLine, e.Line)
	}

	if e.Message != "" {
		tag = tag.Add(logtps.FieldMessage, e.Message)
	}

	if len(e.Error) > 0 {
		var msg = make([]string, 0)

		for _, er := range e.Error {
			if er == nil {
				continue
			}
			msg = append(msg, er.Error())
		}

		tag = tag.Add(logtps.FieldError, strings.Join(msg, ", "))
	}

	if e.Data != nil {
		tag = tag.Add(logtps.FieldData, e.Data)
	}

	tag.Merge(e.Fields)

	if e.log == nil {
		return
	} else if log = e.log(); log == nil {
		return
	} else {
		ent = log.WithFields(tag.Logrus())
	}

	ent.Log(e.Level.Logrus())

	if e.Level <= loglvl.FatalLevel {
		os.Exit(1)
	}
}

func (e *entry) _logClean() {
	var (
		log *logrus.Logger
	)

	if e.log == nil {
		return
	} else if log = e.log(); log == nil {
		return
	} else {
		//log.SetLevel(logrus.InfoLevel)
		log.Info(e.Message)
	}
}
