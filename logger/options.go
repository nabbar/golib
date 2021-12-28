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
)

type FuncCustomConfig func(log Logger)

type NetworkType uint8

const (
	NetworkEmpty NetworkType = iota
	NetworkTCP
	NetworkUDP
)

func (n NetworkType) String() string {
	switch n {
	case NetworkTCP:
		return "tcp"
	case NetworkUDP:
		return "udp"
	default:
		return ""
	}
}

func MakeNetwork(net string) NetworkType {
	switch strings.ToLower(net) {
	case NetworkTCP.String():
		return NetworkTCP
	case NetworkUDP.String():
		return NetworkUDP
	default:
		return NetworkEmpty
	}
}

type OptionsFile struct {
	// LogLevel define the allowed level of log for this file.
	LogLevel []string `json:"logLevel,omitempty" yaml:"logLevel,omitempty" toml:"logLevel,omitempty" mapstructure:"logLevel,omitempty"`

	// Filepath define the file path for log to file.
	Filepath string `json:"filepath,omitempty" yaml:"filepath,omitempty" toml:"filepath,omitempty" mapstructure:"filepath,omitempty"`

	// Create define if the log file must exist or can create it.
	Create bool `json:"create,omitempty" yaml:"create,omitempty" toml:"create,omitempty" mapstructure:"create,omitempty"`

	// CreatePath define if the path of the log file must exist or can try to create it.
	CreatePath bool `json:"createPath,omitempty" yaml:"createPath,omitempty" toml:"createPath,omitempty" mapstructure:"createPath,omitempty"`

	// FileMode define mode to be used for the log file if the create it.
	FileMode os.FileMode `json:"fileMode,omitempty" yaml:"fileMode,omitempty" toml:"fileMode,omitempty" mapstructure:"fileMode,omitempty"`

	// PathMode define mode to be used for the path of the log file if create it.
	PathMode os.FileMode `json:"pathMode,omitempty" yaml:"pathMode,omitempty" toml:"pathMode,omitempty" mapstructure:"pathMode,omitempty"`

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool `json:"disableStack,omitempty" yaml:"disableStack,omitempty" toml:"disableStack,omitempty" mapstructure:"disableStack,omitempty"`

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool `json:"disableTimestamp,omitempty" yaml:"disableTimestamp,omitempty" toml:"disableTimestamp,omitempty" mapstructure:"disableTimestamp,omitempty"`

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool `json:"enableTrace,omitempty" yaml:"enableTrace,omitempty" toml:"enableTrace,omitempty" mapstructure:"enableTrace,omitempty"`
}

type OptionsSyslog struct {
	// LogLevel define the allowed level of log for this syslog.
	LogLevel []string `json:"logLevel,omitempty" yaml:"logLevel,omitempty" toml:"logLevel,omitempty" mapstructure:"logLevel,omitempty"`

	// Network define the network used to connect to this syslog (tcp, udp, or any other to a local connection).
	Network string `json:"network,omitempty" yaml:"network,omitempty" toml:"network,omitempty" mapstructure:"network,omitempty"`

	// Host define the remote syslog to use.
	// If Host and Network are empty, local syslog will be used.
	Host string `json:"host,omitempty" yaml:"host,omitempty" toml:"host,omitempty" mapstructure:"host,omitempty"`

	// Severity define the severity syslog to be used.
	Severity string `json:"severity,omitempty" yaml:"severity,omitempty" toml:"severity,omitempty" mapstructure:"severity,omitempty"`

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
}

type Options struct {
	// DisableStandard allow disabling to write log to standard output stdout/stderr.
	DisableStandard bool `json:"disableStandard,omitempty" yaml:"disableStandard,omitempty" toml:"disableStandard,omitempty" mapstructure:"disableStandard,omitempty"`

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool `json:"disableStack,omitempty" yaml:"disableStack,omitempty" toml:"disableStack,omitempty" mapstructure:"disableStack,omitempty"`

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool `json:"disableTimestamp,omitempty" yaml:"disableTimestamp,omitempty" toml:"disableTimestamp,omitempty" mapstructure:"disableTimestamp,omitempty"`

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool `json:"enableTrace,omitempty" yaml:"enableTrace,omitempty" toml:"enableTrace,omitempty" mapstructure:"enableTrace,omitempty"`

	// TraceFilter define the path to clean for trace.
	TraceFilter string `json:"traceFilter,omitempty" yaml:"traceFilter,omitempty" toml:"traceFilter,omitempty" mapstructure:"traceFilter,omitempty"`

	// DisableColor define if color could be use or not in messages format.
	// If the running process is not a tty, no color will be used.
	DisableColor bool `json:"disableColor,omitempty" yaml:"disableColor,omitempty" toml:"disableColor,omitempty" mapstructure:"disableColor,omitempty"`

	// LogFile define a list of log file configuration to allow log to files.
	LogFile []OptionsFile `json:"logFile,omitempty" yaml:"logFile,omitempty" toml:"logFile,omitempty" mapstructure:"logFile,omitempty"`

	// LogSyslog define a list of syslog configuration to allow log to syslog.
	LogSyslog []OptionsSyslog `json:"logSyslog,omitempty" yaml:"logSyslog,omitempty" toml:"logSyslog,omitempty" mapstructure:"logSyslog,omitempty"`

	// custom function handler.
	init   FuncCustomConfig
	change FuncCustomConfig
}

// RegisterFuncUpdateLogger allow to register a function called when init or update of logger.
// To clean function, just call RegisterFuncUpdateLogger with nil as param.
func (o Options) RegisterFuncUpdateLogger(fct FuncCustomConfig) {
	o.init = fct
}

// RegisterFuncUpdateLevel allow to register a function called when init or update level
// To clean function, just call RegisterFuncUpdateLevel with nil as param.
func (o Options) RegisterFuncUpdateLevel(fct FuncCustomConfig) {
	o.change = fct
}
