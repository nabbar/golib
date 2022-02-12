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

package log

import (
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

type DefaultModel struct {
	liblog.Options
}

type componentLog struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	d func() liblog.Logger
	l liblog.Logger
	v liblog.Level
}

func (c *componentLog) Type() string {
	return ComponentType
}

func (c *componentLog) RegisterContext(fct libcfg.FuncContext) {
	c.ctx = fct
}

func (c *componentLog) RegisterGet(fct libcfg.FuncComponentGet) {
	c.get = fct
}

func (c *componentLog) RegisterFuncStartBefore(fct func() liberr.Error) {
	c.fsb = fct
}

func (c *componentLog) RegisterFuncStartAfter(fct func() liberr.Error) {
	c.fsa = fct
}

func (c *componentLog) RegisterFuncReloadBefore(fct func() liberr.Error) {
	c.frb = fct
}

func (c *componentLog) RegisterFuncReloadAfter(fct func() liberr.Error) {
	c.fra = fct
}

func (c *componentLog) _run(errCode liberr.CodeError, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	if c.ctx == nil {
		return ErrorComponentNotInitialized.Error(nil)
	}

	if c.l == nil {
		c.l = liblog.New(c.ctx())
	}

	cnf := DefaultModel{}
	if err := getCfg(&cnf); err != nil {
		return ErrorParamsInvalid.Error(err)
	} else if err = cnf.Validate(); err != nil {
		return ErrorConfigInvalid.Error(err)
	} else if e := c.l.SetOptions(&cnf.Options); e != nil {
		return errCode.ErrorParent(e)
	}

	return nil
}

func (c *componentLog) Start(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var err liberr.Error

	if c.fsb != nil {
		if err = c.fsb(); err != nil {
			return err
		}
	}

	if err = c._run(ErrorStartLog, getCfg); err != nil {
		return err
	}

	if c.fsa != nil {
		if err = c.fsa(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentLog) Reload(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var err liberr.Error

	if c.frb != nil {
		if err = c.frb(); err != nil {
			return err
		}
	}

	if err = c._run(ErrorStartLog, getCfg); err != nil {
		return err
	}

	if c.fra != nil {
		if err = c.fra(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentLog) Stop() {

}

func (c *componentLog) IsStarted() bool {
	return c.l != nil
}

func (c *componentLog) IsRunning(atLeast bool) bool {
	return c.l != nil
}

func (c *componentLog) Dependencies() []string {
	return make([]string, 0)
}

func (c *componentLog) Log() liblog.Logger {
	if c.l != nil {
		_l, _ := c.l.Clone()
		return _l
	}

	return c.d()
}

func (c *componentLog) SetLevel(lvl liblog.Level) {
	if c.l != nil {
		return
	}

	c.l.SetLevel(lvl)
}

func (c *componentLog) SetField(fields liblog.Fields) {
	if c.l != nil {
		return
	}

	c.l.SetFields(fields)
}

func (c *componentLog) SetOptions(opt *liblog.Options) liberr.Error {
	if c.l != nil {
		return ErrorComponentNotInitialized.Error(nil)
	}

	if e := c.l.SetOptions(opt); e != nil {
		return ErrorConfigInvalid.ErrorParent(e)
	}

	return nil
}
