/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package cobra

import (
	"io"

	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
)

type FuncInit func()
type FuncLogger func() liblog.Logger
type FuncViper func() libvpr.Viper
type FuncPrintErrorCode func(item, value string)

type Cobra interface {
	SetVersion(v libver.Version)

	SetFuncInit(fct FuncInit)
	SetViper(fct FuncViper)
	SetLogger(fct FuncLogger)
	SetForceNoInfo(flag bool)

	Init()

	SetFlagConfig(persistent bool, flagVar *string) error
	SetFlagVerbose(persistent bool, flagVar *int)

	NewCommand(cmd, short, long, useWithoutCmd, exampleWithoutCmd string) *spfcbr.Command
	AddCommand(subCmd ...*spfcbr.Command)

	AddCommandCompletion()
	AddCommandConfigure(basename string, defaultConfig func() io.Reader)
	AddCommandPrintErrorCode(fct FuncPrintErrorCode)

	Execute() error

	Cobra() *spfcbr.Command
}

func New() Cobra {
	return &cobra{
		c: nil,
		s: nil,
		v: nil,
		i: nil,
		l: nil,
	}
}
