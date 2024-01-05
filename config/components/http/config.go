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

package http

import (
	"fmt"

	libvpr "github.com/nabbar/golib/viper"

	htpool "github.com/nabbar/golib/httpserver/pool"
	spfcbr "github.com/spf13/cobra"
)

func (o *componentHttp) RegisterFlag(Command *spfcbr.Command) error {
	return nil
}

func (o *componentHttp) _getConfig() (*htpool.Config, error) {
	var (
		key string
		cfg htpool.Config
		vpr libvpr.Viper
		err error
	)

	if vpr = o._getViper(); vpr == nil {
		return nil, ErrorComponentNotInitialized.Error(nil)
	} else if key = o._getKey(); len(key) < 1 {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if e := vpr.UnmarshalKey(key, &cfg); e != nil {
		return nil, ErrorParamInvalid.Error(e)
	}

	cfg.SetDefaultTLS(o._GetTLS)
	cfg.SetContext(o.x.GetContext)
	cfg.SetHandlerFunc(o._GetHandler)

	if err = cfg.Validate(); err != nil {
		return nil, ErrorConfigInvalid.Error(err)
	} else if o.h == nil {
		return nil, ErrorComponentNotInitialized.Error(fmt.Errorf("missing handler"))
	} else if len(o.h()) < 1 {
		return nil, ErrorComponentNotInitialized.Error(fmt.Errorf("missing handler"))
	}

	return &cfg, nil
}
