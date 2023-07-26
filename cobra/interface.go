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
	"time"

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

	AddFlagString(persistent bool, p *string, name, shorthand string, value string, usage string)
	AddFlagCount(persistent bool, p *int, name, shorthand string, usage string)
	AddFlagBool(persistent bool, p *bool, name, shorthand string, value bool, usage string)
	AddFlagDuration(persistent bool, p *time.Duration, name, shorthand string, value time.Duration, usage string)
	AddFlagFloat32(persistent bool, p *float32, name, shorthand string, value float32, usage string)
	AddFlagFloat64(persistent bool, p *float64, name, shorthand string, value float64, usage string)
	AddFlagInt(persistent bool, p *int, name, shorthand string, value int, usage string)
	AddFlagInt8(persistent bool, p *int8, name, shorthand string, value int8, usage string)
	AddFlagInt16(persistent bool, p *int16, name, shorthand string, value int16, usage string)
	AddFlagInt32(persistent bool, p *int32, name, shorthand string, value int32, usage string)
	AddFlagInt32Slice(persistent bool, p *[]int32, name, shorthand string, value []int32, usage string)
	AddFlagInt64(persistent bool, p *int64, name, shorthand string, value int64, usage string)
	AddFlagInt64Slice(persistent bool, p *[]int64, name, shorthand string, value []int64, usage string)
	AddFlagUint(persistent bool, p *uint, name, shorthand string, value uint, usage string)
	AddFlagUintSlice(persistent bool, p *[]uint, name, shorthand string, value []uint, usage string)
	AddFlagUint8(persistent bool, p *uint8, name, shorthand string, value uint8, usage string)
	AddFlagUint16(persistent bool, p *uint16, name, shorthand string, value uint16, usage string)
	AddFlagUint32(persistent bool, p *uint32, name, shorthand string, value uint32, usage string)
	AddFlagUint64(persistent bool, p *uint64, name, shorthand string, value uint64, usage string)
	AddFlagStringArray(persistent bool, p *[]string, name, shorthand string, value []string, usage string)
	AddFlagStringToInt(persistent bool, p *map[string]int, name, shorthand string, value map[string]int, usage string)
	AddFlagStringToInt64(persistent bool, p *map[string]int64, name, shorthand string, value map[string]int64, usage string)
	AddFlagStringToString(persistent bool, p *map[string]string, name, shorthand string, value map[string]string, usage string)

	NewCommand(cmd, short, long, useWithoutCmd, exampleWithoutCmd string) *spfcbr.Command
	AddCommand(subCmd ...*spfcbr.Command)

	AddCommandCompletion()
	AddCommandConfigure(basename string, defaultConfig func() io.Reader)
	AddCommandPrintErrorCode(fct FuncPrintErrorCode)

	Execute() error

	Cobra() *spfcbr.Command

	ConfigureCheckArgs(basename string, args []string) error
	ConfigureWriteConfig(basename string, defaultConfig func() io.Reader, printMsg func(pkg, file string)) error
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
