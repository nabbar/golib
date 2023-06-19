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

type OptionsStd struct {
	// DisableStandard allow disabling to write log to standard output stdout/stderr.
	DisableStandard bool `json:"disableStandard,omitempty" yaml:"disableStandard,omitempty" toml:"disableStandard,omitempty" mapstructure:"disableStandard,omitempty"`

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool `json:"disableStack,omitempty" yaml:"disableStack,omitempty" toml:"disableStack,omitempty" mapstructure:"disableStack,omitempty"`

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool `json:"disableTimestamp,omitempty" yaml:"disableTimestamp,omitempty" toml:"disableTimestamp,omitempty" mapstructure:"disableTimestamp,omitempty"`

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool `json:"enableTrace,omitempty" yaml:"enableTrace,omitempty" toml:"enableTrace,omitempty" mapstructure:"enableTrace,omitempty"`

	// DisableColor define if color could be use or not in messages format.
	// If the running process is not a tty, no color will be used.
	DisableColor bool `json:"disableColor,omitempty" yaml:"disableColor,omitempty" toml:"disableColor,omitempty" mapstructure:"disableColor,omitempty"`

	// EnableAccessLog allow to add all message from api router for access log and error log.
	EnableAccessLog bool `json:"enableAccessLog,omitempty" yaml:"enableAccessLog,omitempty" toml:"enableAccessLog,omitempty" mapstructure:"enableAccessLog,omitempty"`
}

func (o *OptionsStd) Clone() *OptionsStd {
	return &OptionsStd{
		DisableStandard:  o.DisableStandard,
		DisableStack:     o.DisableStack,
		DisableTimestamp: o.DisableTimestamp,
		EnableTrace:      o.EnableTrace,
		DisableColor:     o.DisableColor,
		EnableAccessLog:  o.EnableAccessLog,
	}
}
