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

package tls

import (
	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
)

type componentTls struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	tls libtls.TLSConfig
}

func (c *componentTls) Type() string {
	return ComponentType
}

func (c *componentTls) RegisterContext(fct libcfg.FuncContext) {
	c.ctx = fct
}

func (c *componentTls) RegisterGet(fct libcfg.FuncComponentGet) {
	c.get = fct
}

func (c *componentTls) RegisterFuncStartBefore(fct func() liberr.Error) {
	c.fsb = fct
}

func (c *componentTls) RegisterFuncStartAfter(fct func() liberr.Error) {
	c.fsa = fct
}

func (c *componentTls) RegisterFuncReloadBefore(fct func() liberr.Error) {
	c.frb = fct
}

func (c *componentTls) RegisterFuncReloadAfter(fct func() liberr.Error) {
	c.fra = fct
}

func (c *componentTls) _run(errCode liberr.CodeError, getCfg libcfg.FuncComponentConfigGet) (libtls.TLSConfig, liberr.Error) {
	cnf := libtls.Config{}

	if c == nil {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if err := getCfg(&cnf); err != nil {
		return nil, ErrorParamsInvalid.Error(err)
	}

	if err := cnf.Validate(); err != nil {
		return nil, ErrorConfigInvalid.Error(err)
	}

	if tls, err := cnf.New(); err != nil {
		return nil, errCode.Error(err)
	} else {
		return tls, nil
	}
}

func (c *componentTls) Start(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		err liberr.Error
		tls libtls.TLSConfig
	)

	if c.fsb != nil {
		if err = c.fsb(); err != nil {
			return err
		}
	}

	if tls, err = c._run(ErrorStartTLS, getCfg); err != nil {
		return err
	} else {
		c.tls = tls
	}

	if c.fsa != nil {
		if err = c.fsa(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentTls) Reload(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		err liberr.Error
		tls libtls.TLSConfig
	)

	if c.frb != nil {
		if err = c.frb(); err != nil {
			return err
		}
	}

	if tls, err = c._run(ErrorReloadTLS, getCfg); err != nil {
		return err
	} else {
		c.tls = tls
	}

	if c.fra != nil {
		if err = c.fra(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentTls) Stop() {}

func (c *componentTls) Dependencies() []string {
	return make([]string, 0)
}

func (c *componentTls) IsStarted() bool {
	return c.tls != nil
}

func (c *componentTls) IsRunning(atLeast bool) bool {
	return c.tls != nil
}

func (c *componentTls) GetTLS() libtls.TLSConfig {
	if c == nil || c.tls == nil {
		return nil
	}

	return c.tls
}

func (c *componentTls) SetTLS(tls libtls.TLSConfig) {
	if c == nil || tls == nil {
		return
	}

	c.tls = tls
}
