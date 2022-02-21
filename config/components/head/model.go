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

package head

import (
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libhts "github.com/nabbar/golib/httpserver"
	librtr "github.com/nabbar/golib/router"
)

type DefaultModel struct {
	Head librtr.HeadersConfig
	Http libhts.PoolServerConfig
}

type componentHead struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	head librtr.Headers
}

func (c *componentHead) Type() string {
	return ComponentType
}

func (c *componentHead) RegisterContext(fct libcfg.FuncContext) {
	c.ctx = fct
}

func (c *componentHead) RegisterGet(fct libcfg.FuncComponentGet) {
	c.get = fct
}

func (c *componentHead) RegisterFuncStartBefore(fct func() liberr.Error) {
	c.fsb = fct
}

func (c *componentHead) RegisterFuncStartAfter(fct func() liberr.Error) {
	c.fsa = fct
}

func (c *componentHead) RegisterFuncReloadBefore(fct func() liberr.Error) {
	c.frb = fct
}

func (c *componentHead) RegisterFuncReloadAfter(fct func() liberr.Error) {
	c.fra = fct
}

func (c *componentHead) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	if c == nil {
		return ErrorComponentNotInitialized.Error(nil)
	}

	cnf := DefaultModel{}
	if err := getCfg(&cnf); err != nil {
		return ErrorParamsInvalid.Error(err)
	}

	c.head = cnf.Head.New()

	return nil
}

func (c *componentHead) Start(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var err liberr.Error

	if c.fsb != nil {
		if err = c.fsb(); err != nil {
			return err
		}
	}

	if err = c._run(getCfg); err != nil {
		return err
	}

	if c.fsa != nil {
		if err = c.fsa(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentHead) Reload(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		err liberr.Error
	)

	if c.frb != nil {
		if err = c.frb(); err != nil {
			return err
		}
	}

	if err = c._run(getCfg); err != nil {
		return err
	}

	if c.fra != nil {
		if err = c.fra(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentHead) Stop() {

}

func (c *componentHead) IsStarted() bool {
	return c.head == nil
}

func (c *componentHead) IsRunning(atLeast bool) bool {
	return c.head == nil
}

func (c *componentHead) Dependencies() []string {
	return []string{}
}

func (c *componentHead) GetHeaders() librtr.Headers {
	if c == nil || c.head == nil {
		return nil
	}

	return c.head
}

func (c *componentHead) SetHeaders(head librtr.Headers) {
	if c == nil {
		return
	}

	c.head = head
}
