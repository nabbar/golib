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

package types

import (
	"encoding"
	"encoding/json"
)

// InfoData provides a method for retrieving dynamic information about a component.
// Implementations can return runtime-generated metadata as a key-value map.
type InfoData interface {
	// Info returns a map of string to interface that contains information about the
	// monitor. Common keys include "version", "build", "uptime", etc.
	Info() map[string]interface{}
}

// InfoName provides a method for retrieving the component name.
// This interface allows for static or dynamic name generation.
type InfoName interface {
	// Name returns the name of the component.
	// The name is used to identify the component in logs and metrics.
	Name() string
}

// Info is the main interface for component metadata management.
// It combines name and data retrieval with encoding capabilities.
type Info interface {
	encoding.TextMarshaler
	json.Marshaler

	InfoName
	InfoData
}
