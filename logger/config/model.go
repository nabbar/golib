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

package config

import (
	"fmt"

	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
)

type FuncOpt func() *Options

type Options struct {
	// InheritDefault define if the current options will override a default options
	InheritDefault bool `json:"inheritDefault" yaml:"inheritDefault" toml:"inheritDefault" mapstructure:"inheritDefault"`

	// TraceFilter define the path to clean for trace.
	TraceFilter string `json:"traceFilter,omitempty" yaml:"traceFilter,omitempty" toml:"traceFilter,omitempty" mapstructure:"traceFilter,omitempty"`

	// Stdout define the options for stdout/stderr log.
	Stdout *OptionsStd `json:"stdout,omitempty" yaml:"stdout,omitempty" toml:"stdout,omitempty" mapstructure:"stdout,omitempty"`

	// LogFileExtend define if the logFile given is in addition of default LogFile or a replacement.
	LogFileExtend bool `json:"logFileExtend,omitempty" yaml:"logFileExtend,omitempty" toml:"logFileExtend,omitempty" mapstructure:"logFileExtend,omitempty"`

	// LogFile define a list of log file configuration to allow log to files.
	LogFile OptionsFiles `json:"logFile,omitempty" yaml:"logFile,omitempty" toml:"logFile,omitempty" mapstructure:"logFile,omitempty"`

	// LogSyslogExtend define if the logFile given is in addition of default LogSyslog or a replacement.
	LogSyslogExtend bool `json:"logSyslogExtend,omitempty" yaml:"logSyslogExtend,omitempty" toml:"logSyslogExtend,omitempty" mapstructure:"logSyslogExtend,omitempty"`

	// LogSyslog define a list of syslog configuration to allow log to syslog.
	LogSyslog OptionsSyslogs `json:"logSyslog,omitempty" yaml:"logSyslog,omitempty" toml:"logSyslog,omitempty" mapstructure:"logSyslog,omitempty"`

	// default options
	opts FuncOpt
}

// RegisterDefaultFunc allow to register a function called to retrieve default options for inheritDefault.
// If not set, the previous options will be used as default options.
// To clean function, just call RegisterDefaultFunc with nil as param.
func (o *Options) RegisterDefaultFunc(fct FuncOpt) {
	o.opts = fct
}

// Validate allow checking if the options' struct is valid with the awaiting model
func (o *Options) Validate() liberr.Error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(o); err != nil {
		if er, ok := err.(*libval.InvalidValidationError); ok {
			e.Add(er)
		}

		for _, er := range err.(libval.ValidationErrors) {
			//nolint #goerr113
			e.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", er.Namespace(), er.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (o *Options) Clone() Options {
	var s *OptionsStd

	if o.Stdout != nil {
		s = o.Stdout.Clone()
	}

	return Options{
		InheritDefault: o.InheritDefault,
		TraceFilter:    o.TraceFilter,
		Stdout:         s,
		LogFile:        o.LogFile.Clone(),
		LogSyslog:      o.LogSyslog.Clone(),
	}
}

func (o *Options) Merge(opt *Options) {
	if opt == nil {
		return
	}

	if len(opt.TraceFilter) > 0 {
		o.TraceFilter = opt.TraceFilter
	}

	if opt.Stdout != nil {
		if o.Stdout == nil {
			o.Stdout = &OptionsStd{}
		}

		sdt := *opt.Stdout
		osd := *o.Stdout

		if sdt.DisableStandard {
			osd.DisableStandard = sdt.DisableStandard
		}
		if sdt.DisableColor {
			osd.DisableColor = sdt.DisableColor
		}
		if sdt.DisableTimestamp {
			osd.DisableTimestamp = sdt.DisableTimestamp
		}
		if sdt.DisableStack {
			osd.DisableStack = sdt.DisableStack
		}
		if sdt.EnableTrace {
			osd.EnableTrace = sdt.EnableTrace
		}
		if sdt.EnableAccessLog {
			osd.EnableAccessLog = sdt.EnableAccessLog
		}

		o.Stdout = &osd
	}

	if opt.LogFileExtend {
		o.LogFile = append(o.LogFile, opt.LogFile...)
	} else {
		o.LogFile = opt.LogFile
	}

	if opt.LogSyslogExtend {
		o.LogSyslog = append(o.LogSyslog, opt.LogSyslog...)
	} else {
		o.LogSyslog = opt.LogSyslog
	}

	if opt.opts != nil {
		o.opts = opt.opts
	}
}

func (o *Options) Options() *Options {
	var no Options

	if o.opts != nil && o.InheritDefault {
		no = *o.opts()
	}

	if len(o.TraceFilter) > 0 {
		no.TraceFilter = o.TraceFilter
	}

	if o.Stdout != nil {
		if no.Stdout == nil {
			no.Stdout = &OptionsStd{}
		}
		if o.Stdout.DisableStandard {
			no.Stdout.DisableStandard = o.Stdout.DisableStandard
		}
		if o.Stdout.DisableColor {
			no.Stdout.DisableColor = o.Stdout.DisableColor
		}
		if o.Stdout.DisableTimestamp {
			no.Stdout.DisableTimestamp = o.Stdout.DisableTimestamp
		}
		if o.Stdout.DisableStack {
			no.Stdout.DisableStack = o.Stdout.DisableStack
		}
		if o.Stdout.EnableTrace {
			no.Stdout.EnableTrace = o.Stdout.EnableTrace
		}
		if o.Stdout.EnableAccessLog {
			no.Stdout.EnableAccessLog = o.Stdout.EnableAccessLog
		}
	}

	if o.LogFileExtend {
		no.LogFile = append(no.LogFile, o.LogFile...)
	} else {
		no.LogFile = o.LogFile
	}

	if o.LogSyslogExtend {
		no.LogSyslog = append(no.LogSyslog, o.LogSyslog...)
	} else {
		no.LogSyslog = o.LogSyslog
	}

	return &no
}
