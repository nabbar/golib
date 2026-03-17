/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package status

import (
	"bytes"
	"encoding/json"

	montps "github.com/nabbar/golib/monitor/types"
)

// encComponent defines an interface for objects that can be marshaled into
// both text and JSON formats. This is used to handle different component
// encoding strategies (e.g., list vs. map).
type encComponent interface {
	MarshalText() (text []byte, err error)
	MarshalJSON() ([]byte, error)
}

// modPool is a helper struct that encapsulates the necessary information for
// encoding component statuses in "map mode". It holds the mandatory component
// list, the monitor pool, and a function to compute group status.
type modPool struct {
	cpt map[string]montps.MonitorStatus
}

// MarshalText implements the `encoding.TextMarshaler` interface. It generates a
// plain text representation of the component groups, their modes, and their statuses.
//
// Format:
//
//	STATUS - Mode:
//	    Component1: Status - Message
//	    Component2: Status - Message
func (o modPool) MarshalText() (text []byte, err error) {
	var buf = bytes.NewBuffer(make([]byte, 0))

	for _, v := range o.cpt {
		if p, e := v.MarshalText(); e != nil {
			err = e
			break
		} else {
			buf.Write(p)
			buf.WriteRune('\n')
		}
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MarshalJSON implements the `json.Marshaler` interface. It generates a JSON
// representation of the component groups.
func (o modPool) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.cpt)
}
