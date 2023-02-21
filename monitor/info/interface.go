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

package info

import (
	"fmt"
	"sync"

	montps "github.com/nabbar/golib/monitor/types"
)

type FuncName func() (string, error)
type FuncInfo func() (map[string]interface{}, error)

type Info interface {
	montps.Info

	RegisterName(fct FuncName)
	RegisterInfo(fct FuncInfo)
}

func New(defaultName string) (Info, error) {
	i := &inf{
		m:  sync.RWMutex{},
		fi: nil,
		fn: nil,
		v:  sync.Map{},
	}

	if len(defaultName) < 1 {
		return nil, fmt.Errorf("default name cannot be empty")
	}

	i.v.Store(keyDefName, defaultName)
	return i, nil
}
