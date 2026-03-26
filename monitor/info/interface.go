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

// Info is an interface that describes the behavior of a metadata container for monitored components.
// It extends the montps.InfoSet interface, which already incorporates read-only access (montps.Info)
// and write/update methods (montps.InfoSet).
// This structure is typically used to store identifying information (like a name) and dynamic
// status data (key-value pairs) about a monitored service.
// Implementations of this interface MUST ensure thread-safe operations as they are often accessed from
// concurrent health check routines and status reporting endpoints.
type Info interface {
	montps.InfoSet
}

// New initializes and returns a new instance of the Info interface, configured with a mandatory default name.
// This default name serves as a permanent fallback identifier for the monitored component.
//
// Arguments:
//   - defaultName: A string used as the initial name and fallback for the component. It cannot be empty.
//
// Returns:
//   - Info: A pointer to a thread-safe implementation of the Info interface.
//   - error: Returns an error if the defaultName provided is empty, indicating an invalid initialization attempt.
//
// Example usage:
//
//	// Create a new Info instance for a database service.
//	inf, err := info.New("database-primary")
//	if err != nil {
//	    log.Fatalf("Failed to initialize monitor info: %v", err)
//	}
//
//	// The default name is immediately available.
//	fmt.Println(inf.Name()) // Output: database-primary
func New(defaultName string) (Info, error) {
	if len(defaultName) < 1 {
		return nil, fmt.Errorf("default name cannot be empty")
	}

	// Returns a new 'inf' struct, which is the private implementation of the Info interface.
	// It uses an atomic map to store dynamic data, ensuring thread-safety.
	return &inf{
		n: defaultName,
		v: libatm.NewMapAny[string](),
	}, nil
}
