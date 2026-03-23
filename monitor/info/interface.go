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

	libatm "github.com/nabbar/golib/atomic"
	montps "github.com/nabbar/golib/monitor/types"
)

// Info defines the contract for a component that exposes identification
// and status information. It combines read access (via montps.Info)
// and write/update capabilities (via montps.InfoSet).
// Implementations must be thread-safe.
type Info interface {
	montps.InfoSet
}

// New initializes and returns a new Info instance with the specified default name.
// The default name serves as a fallback if no dynamic name function is registered
// or if the registered function returns an empty string or error.
//
// It returns an error if the provided defaultName is empty.
//
// Example usage:
//
//	inf, err := info.New("my-service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(inf.Name()) // Output: my-service
func New(defaultName string) (Info, error) {
	if len(defaultName) < 1 {
		return nil, fmt.Errorf("default name cannot be empty")
	}

	return &inf{
		n: defaultName,
		v: libatm.NewMapAny[string](),
	}, nil
}
