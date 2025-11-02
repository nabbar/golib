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

// FuncName is a function type that returns a dynamic name and an optional error.
// The function is called once and the result is cached unless an error occurs.
type FuncName func() (string, error)

// FuncInfo is a function type that returns dynamic info data and an optional error.
// The function is called once and the results are cached unless an error occurs.
type FuncInfo func() (map[string]interface{}, error)

type Info interface {
	montps.Info

	// RegisterName registers a function that returns a default name.
	// The function must return a string and an error.
	// If the function returns an error, the default name is not registered.
	//
	RegisterName(fct FuncName)
	// RegisterInfo registers a function that returns a default info.
	// The function must return a map of string to interface{} and an error.
	// If the function returns an error, the default info is not registered.
	RegisterInfo(fct FuncInfo)
}

// New creates a new Info instance with the given default name.
// The default name is used when no dynamic name function is registered or when it fails.
// Returns an error if the default name is empty.
//
// Example:
//
//	info, err := info.New("my-service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(info.Name()) // Output: my-service
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
