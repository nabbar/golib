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
	"fmt"
	"os"
	"path"
	"strings"

	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	spfcbr "github.com/spf13/cobra"
)

type cobra struct {
	c *spfcbr.Command
	s libver.Version
	b bool
	d string

	v FuncViper
	i FuncInit
	l FuncLogger
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
		println(c.s.GetHeader())
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

	return liblog.GetDefault()
}

func (c *cobra) getPackageName() string {
	pkg := path.Base(os.Args[0])

	if pkg == "" {
		if f, e := os.Executable(); e == nil {
			pkg = path.Base(f)
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
