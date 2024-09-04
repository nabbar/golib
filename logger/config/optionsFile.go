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
	libprm "github.com/nabbar/golib/file/perm"
	libsiz "github.com/nabbar/golib/size"
)

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
	FileMode libprm.Perm `json:"fileMode,omitempty" yaml:"fileMode,omitempty" toml:"fileMode,omitempty" mapstructure:"fileMode,omitempty"`

	// PathMode define mode to be used for the path of the log file if create it.
	PathMode libprm.Perm `json:"pathMode,omitempty" yaml:"pathMode,omitempty" toml:"pathMode,omitempty" mapstructure:"pathMode,omitempty"`

	// DisableStack allow to disable the goroutine id before each message.
	DisableStack bool `json:"disableStack,omitempty" yaml:"disableStack,omitempty" toml:"disableStack,omitempty" mapstructure:"disableStack,omitempty"`

	// DisableTimestamp allow to disable the timestamp before each message.
	DisableTimestamp bool `json:"disableTimestamp,omitempty" yaml:"disableTimestamp,omitempty" toml:"disableTimestamp,omitempty" mapstructure:"disableTimestamp,omitempty"`

	// EnableTrace allow to add the origin caller/file/line of each message.
	EnableTrace bool `json:"enableTrace,omitempty" yaml:"enableTrace,omitempty" toml:"enableTrace,omitempty" mapstructure:"enableTrace,omitempty"`

	// EnableAccessLog allow to add all message from api router for access log and error log.
	EnableAccessLog bool `json:"enableAccessLog,omitempty" yaml:"enableAccessLog,omitempty" toml:"enableAccessLog,omitempty" mapstructure:"enableAccessLog,omitempty"`

	// FileBufferSize define the size for buffer size (by default the buffer size is set to 32KB).
	FileBufferSize libsiz.Size `json:"file-buffer-size,omitempty" yaml:"file-buffer-size,omitempty" toml:"file-buffer-size,omitempty" mapstructure:"file-buffer-size,omitempty"`
}

type OptionsFiles []OptionsFile

func (o OptionsFile) Clone() OptionsFile {
	return OptionsFile{
		LogLevel:         o.LogLevel,
		Filepath:         o.Filepath,
		Create:           o.Create,
		CreatePath:       o.CreatePath,
		FileMode:         o.FileMode,
		PathMode:         o.PathMode,
		DisableStack:     o.DisableStack,
		DisableTimestamp: o.DisableTimestamp,
		EnableTrace:      o.EnableTrace,
		EnableAccessLog:  o.EnableAccessLog,
	}
}

func (o OptionsFiles) Clone() OptionsFiles {
	var c = make([]OptionsFile, 0)
	for _, i := range o {
		c = append(c, i.Clone())
	}
	return c
}
