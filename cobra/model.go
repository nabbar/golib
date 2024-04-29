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
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"path/filepath"
	"strings"
	"time"

	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	spfcbr "github.com/spf13/cobra"
)

type cobra struct {
	c *spfcbr.Command
	s libver.Version
	b bool
	d string
	q []Question
	v FuncViper
	i FuncInit
	l FuncLogger
}

func (c *cobra) model() tea.Model {
	return &promptModel{questions: c.q, cursor: 0}
}

func (c *cobra) RunInteractiveUI() {
	if c.q == nil {
		return
	}
	p := tea.NewProgram(c.model())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
func (c *cobra) SetUIQuestions(questions []Question) {
	c.q = questions
}

func (c *cobra) Cobra() *spfcbr.Command {
	return c.c
}

func (c *cobra) Init() {

	c.c = &spfcbr.Command{
		TraverseChildren: true,
		Use:              c.getPackageName(),
		Version:          c.getPackageVersion(),
		Short:            c.getPackageDescShort(),
		Long:             c.getPackageDescLong(),
	}

	// launch cobra flag parsing
	spfcbr.OnInitialize(c.printHeader, c.i)
}

func (c *cobra) printHeader() {
	if !c.b {
		_, _ = fmt.Fprintln(os.Stdout, c.s.GetHeader())
	}
}

func (c *cobra) SetForceNoInfo(flag bool) {
	c.b = flag
}

func (c *cobra) Execute() error {
	return c.c.Execute()
}

func (c *cobra) SetVersion(vers libver.Version) {
	c.s = vers
}

func (c *cobra) SetFuncInit(fct FuncInit) {
	c.i = fct
}

func (c *cobra) SetViper(fct FuncViper) {
	c.v = fct
}

func (c *cobra) SetLogger(fct FuncLogger) {
	c.l = fct
}

func (c *cobra) SetFlagConfig(persistent bool, flagVar *string) error {
	if persistent {
		c.c.PersistentFlags().StringVarP(flagVar, "config", "c", "", "specify the config file to load (default is $HOME/."+strings.ToLower(c.getPackageName())+".[yaml|json|toml])")
		return c.c.MarkPersistentFlagFilename("config", "json", "toml", "yaml", "yml")
	} else {
		c.c.Flags().StringVarP(flagVar, "config", "c", "", "specify the config file to load (default is $HOME/."+strings.ToLower(c.getPackageName())+".[yaml|json|toml])")
		return c.c.MarkFlagFilename("config", "json", "toml", "yaml", "yml")
	}
}

func (c *cobra) SetFlagVerbose(persistent bool, flagVar *int) {
	if persistent {
		c.c.PersistentFlags().CountVarP(flagVar, "verbose", "v", "enable verbose mode (multi allowed v, vv, vvv)")
	} else {
		c.c.Flags().CountVarP(flagVar, "verbose", "v", "enable verbose mode (multi allowed v, vv, vvv)")
	}
}

func (c *cobra) NewCommand(cmd, short, long, useWithoutCmd, exampleWithoutCmd string) *spfcbr.Command {
	return &spfcbr.Command{
		Use:     fmt.Sprintf("%s %s", cmd, useWithoutCmd),
		Short:   short,
		Long:    long,
		Example: fmt.Sprintf("%s %s", cmd, exampleWithoutCmd),
	}
}

func (c *cobra) AddCommand(subCmd ...*spfcbr.Command) {
	c.c.AddCommand(subCmd...)
}

func (c *cobra) getLog() liblog.Logger {
	var l liblog.Logger

	if c.l != nil {
		l = c.l()
	}

	if l != nil {
		return l
	}

	return liblog.New(context.Background)
}

func (c *cobra) getPackageName() string {
	pkg := filepath.Base(os.Args[0])

	if pkg == "" {
		if f, e := os.Executable(); e == nil {
			pkg = filepath.Base(f)
		} else {
			pkg = c.s.GetPackage()
		}
	}

	return pkg
}

func (c *cobra) getPackageVersion() string {
	if c.s == nil {
		return "missing version"
	}

	return fmt.Sprintf("Version details: \n\tHash: %s\n\tVersion: %s\n\tRuntime: %s\n\tAuthor: %s\n\tDate: %s\n\tLicence: %s\n", c.s.GetBuild(), c.s.GetRelease(), c.s.GetAppId(), c.s.GetAuthor(), c.s.GetDate(), c.s.GetLicenseName())
}

func (c *cobra) getPackageGRootPath() string {
	return c.s.GetRootPackagePath()
}

func (c *cobra) getPackageDescShort() string {
	return c.d
}

func (c *cobra) getPackageDescLong() string {
	return c.s.GetDescription()
}

func (c *cobra) AddFlagString(persistent bool, p *string, name, shorthand string, value string, usage string) {
	if persistent {
		c.c.PersistentFlags().StringVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().StringVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagCount(persistent bool, p *int, name, shorthand string, usage string) {
	if persistent {
		c.c.PersistentFlags().CountVarP(p, name, shorthand, usage)
	} else {
		c.c.Flags().CountVarP(p, name, shorthand, usage)
	}
}

func (c *cobra) AddFlagBool(persistent bool, p *bool, name, shorthand string, value bool, usage string) {
	if persistent {
		c.c.PersistentFlags().BoolVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().BoolVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagDuration(persistent bool, p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	if persistent {
		c.c.PersistentFlags().DurationVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().DurationVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagFloat32(persistent bool, p *float32, name, shorthand string, value float32, usage string) {
	if persistent {
		c.c.PersistentFlags().Float32VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Float32VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagFloat64(persistent bool, p *float64, name, shorthand string, value float64, usage string) {
	if persistent {
		c.c.PersistentFlags().Float64VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Float64VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagInt(persistent bool, p *int, name, shorthand string, value int, usage string) {
	if persistent {
		c.c.PersistentFlags().IntVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().IntVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagInt8(persistent bool, p *int8, name, shorthand string, value int8, usage string) {
	if persistent {
		c.c.PersistentFlags().Int8VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Int8VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagInt16(persistent bool, p *int16, name, shorthand string, value int16, usage string) {
	if persistent {
		c.c.PersistentFlags().Int16VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Int16VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagInt32(persistent bool, p *int32, name, shorthand string, value int32, usage string) {
	if persistent {
		c.c.PersistentFlags().Int32VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Int32VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagInt32Slice(persistent bool, p *[]int32, name, shorthand string, value []int32, usage string) {
	if persistent {
		c.c.PersistentFlags().Int32SliceVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Int32SliceVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagInt64(persistent bool, p *int64, name, shorthand string, value int64, usage string) {
	if persistent {
		c.c.PersistentFlags().Int64VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Int64VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagInt64Slice(persistent bool, p *[]int64, name, shorthand string, value []int64, usage string) {
	if persistent {
		c.c.PersistentFlags().Int64SliceVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Int64SliceVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagUint(persistent bool, p *uint, name, shorthand string, value uint, usage string) {
	if persistent {
		c.c.PersistentFlags().UintVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().UintVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagUintSlice(persistent bool, p *[]uint, name, shorthand string, value []uint, usage string) {
	if persistent {
		c.c.PersistentFlags().UintSliceVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().UintSliceVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagUint8(persistent bool, p *uint8, name, shorthand string, value uint8, usage string) {
	if persistent {
		c.c.PersistentFlags().Uint8VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Uint8VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagUint16(persistent bool, p *uint16, name, shorthand string, value uint16, usage string) {
	if persistent {
		c.c.PersistentFlags().Uint16VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Uint16VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagUint32(persistent bool, p *uint32, name, shorthand string, value uint32, usage string) {
	if persistent {
		c.c.PersistentFlags().Uint32VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Uint32VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagUint64(persistent bool, p *uint64, name, shorthand string, value uint64, usage string) {
	if persistent {
		c.c.PersistentFlags().Uint64VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().Uint64VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagStringArray(persistent bool, p *[]string, name, shorthand string, value []string, usage string) {
	if persistent {
		c.c.PersistentFlags().StringArrayVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().StringArrayVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagStringToInt(persistent bool, p *map[string]int, name, shorthand string, value map[string]int, usage string) {
	if persistent {
		c.c.PersistentFlags().StringToIntVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().StringToIntVarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagStringToInt64(persistent bool, p *map[string]int64, name, shorthand string, value map[string]int64, usage string) {
	if persistent {
		c.c.PersistentFlags().StringToInt64VarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().StringToInt64VarP(p, name, shorthand, value, usage)
	}
}

func (c *cobra) AddFlagStringToString(persistent bool, p *map[string]string, name, shorthand string, value map[string]string, usage string) {
	if persistent {
		c.c.PersistentFlags().StringToStringVarP(p, name, shorthand, value, usage)
	} else {
		c.c.Flags().StringToStringVarP(p, name, shorthand, value, usage)
	}
}
