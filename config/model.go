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

package config

import (
	"sync"

	libctx "github.com/nabbar/golib/context"
)

const (
	fctViper uint8 = iota + 1
	fctStartBefore
	fctStartAfter
	fctReloadBefore
	fctReloadAfter
	fctStopBefore
	fctStopAfter
	fctVersion
	fctLoggerDef
	fctMonitorPool
)

type configModel struct {
	m sync.RWMutex

	ctx libctx.Config[string]
	cpt libctx.Config[string]
	fct libctx.Config[uint8]

	fcnl []func()
}

func (c *configModel) _ComponentGetConfig(key string, model interface{}) error {
	if vpr := c.getViper(); vpr == nil {
		return ErrorConfigMissingViper.Error(nil)
	} else if vip := vpr.Viper(); vip == nil {
		return ErrorConfigMissingViper.Error(nil)
	} else if err := vpr.Viper().UnmarshalKey(key, model); err != nil {
		return ErrorComponentConfigError.Error(err)
	}

	return nil
}
