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

type OptionsSyslog struct {
	// LogLevel define the allowed level of log for this syslog.
	LogLevel []string `json:"logLevel,omitempty" yaml:"logLevel,omitempty" toml:"logLevel,omitempty" mapstructure:"logLevel,omitempty"`

	// Network define the network used to connect to this syslog (tcp, udp, or any other to a local connection).
	Network string `json:"network,omitempty" yaml:"network,omitempty" toml:"network,omitempty" mapstructure:"network,omitempty"`

	// Host define the remote syslog to use.
	// If Host and Network are empty, local syslog will be used.
	Host string `json:"host,omitempty" yaml:"host,omitempty" toml:"host,omitempty" mapstructure:"host,omitempty"`
	/*
		// Severity define the severity syslog to be used.
		Severity string `json:"severity,omitempty" yaml:"severity,omitempty" toml:"severity,omitempty" mapstructure:"severity,omitempty"`
	*/
	// Facility define the facility syslog to be used.
	Facility string `json:"facility,omitempty" yaml:"facility,omitempty" toml:"facility,omitempty" mapstructure:"facility,omitempty"`

	// Tag define the syslog tag used in linux syslog system or name of logger for windows event logger.
	// For window, this value must be unic for each syslog config
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty" toml:"tag,omitempty" mapstructure:"tag,omitempty"`

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool `json:"disableStack,omitempty" yaml:"disableStack,omitempty" toml:"disableStack,omitempty" mapstructure:"disableStack,omitempty"`

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool `json:"disableTimestamp,omitempty" yaml:"disableTimestamp,omitempty" toml:"disableTimestamp,omitempty" mapstructure:"disableTimestamp,omitempty"`

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool `json:"enableTrace,omitempty" yaml:"enableTrace,omitempty" toml:"enableTrace,omitempty" mapstructure:"enableTrace,omitempty"`

	// EnableAccessLog allow to add all message from api router for access log and error log.
	EnableAccessLog bool `json:"enableAccessLog,omitempty" yaml:"enableAccessLog,omitempty" toml:"enableAccessLog,omitempty" mapstructure:"enableAccessLog,omitempty"`
}

type OptionsSyslogs []OptionsSyslog

func (o OptionsSyslog) Clone() OptionsSyslog {
	// Deep copy LogLevel slice
	var logLevel []string
	if o.LogLevel != nil {
		logLevel = make([]string, len(o.LogLevel))
		copy(logLevel, o.LogLevel)
	}

	return OptionsSyslog{
		LogLevel:         logLevel,
		Network:          o.Network,
		Host:             o.Host,
		Facility:         o.Facility,
		Tag:              o.Tag,
		DisableStack:     o.DisableStack,
		DisableTimestamp: o.DisableTimestamp,
		EnableTrace:      o.EnableTrace,
		EnableAccessLog:  o.EnableAccessLog,
	}
}

func (o OptionsSyslogs) Clone() OptionsSyslogs {
	var c = make([]OptionsSyslog, 0)
	for _, i := range o {
		c = append(c, i.Clone())
	}
	return c
}
