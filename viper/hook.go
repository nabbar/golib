/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
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

package viper

import (
	"math"
	"reflect"

	libmap "github.com/mitchellh/mapstructure"
	spfvpr "github.com/spf13/viper"
)

func (v *viper) hookIdx() uint8 {
	i := v.i.Load()
	if i < math.MaxUint8 {
		return uint8(i)
	} else {
		return math.MaxUint8
	}
}

func (v *viper) hookIdxInc() uint8 {
	v.i.Add(1)
	return v.hookIdx()
}

func (v *viper) HookRegister(hook libmap.DecodeHookFunc) {
	v.h.Store(v.hookIdxInc(), hook)
}

func (v *viper) HookReset() {
	v.h.Clean()
}

func (v *viper) hookFuncViperDecoder() spfvpr.DecoderConfigOption {
	return func(c *libmap.DecoderConfig) {
		h := c.DecodeHook
		c.DecodeHook = libmap.ComposeDecodeHookFunc(
			v.hookFuncMapDecoder(),
			h,
		)
	}
}

func (v *viper) hookFuncMapDecoder() libmap.DecodeHookFunc {
	return func(f reflect.Value, t reflect.Value) (interface{}, error) {
		var err error
		data := f.Interface()
		newFrom := f

		v.h.Walk(func(key uint8, val interface{}) bool {
			if val == nil {
				return true
			}

			if data, err = libmap.DecodeHookExec(val, newFrom, t); err != nil {
				return false
			}

			newFrom = reflect.ValueOf(data)
			return true
		})

		if err != nil {
			return nil, err
		}

		return data, nil
	}
}
