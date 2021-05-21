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
	"log/syslog"
	"os"
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

type OptionsFile struct {
	// LogLevel define the allowed level of log for this file.
	LogLevel []string

	// Filepath define the file path for log to file.
	Filepath string

	// Create define if the log file must exist or can create it.
	Create bool

	// CreatePath define if the path of the log file must exist or can try to create it.
	CreatePath bool

	// FileMode define mode to be used for the log file if the create it.
	FileMode os.FileMode

	// PathMode define mode to be used for the path of the log file if create it.
	PathMode os.FileMode

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool
}

type OptionsSyslog struct {
	// LogLevel define the allowed level of log for this syslog.
	LogLevel []string

	// Network define the network used to connect to this syslog.
	Network NetworkType

	// Host define the remote syslog to use.
	// If Host and Network are empty, local syslog will be used.
	Host string

	// Priority define the priority used for this syslog.
	Priority syslog.Priority

	// Tag define the syslog tag used for log message.
	Tag string

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool
}

type Options struct {
	// DisableStandard allow to disable writing log to standard output stdout/stderr.
	DisableStandard bool

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool

	// TraceFilter define the path to clean for trace.
	TraceFilter string

	// DisableColor define if color could be use or not in messages format.
	// If the running process is not a tty, no color will be used.
	DisableColor bool

	// LogFile define a list of log file configuration to allow log to files.
	LogFile []OptionsFile

	// LogSyslog define a list of syslog configuration to allow log to syslog.
	LogSyslog []OptionsSyslog

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
