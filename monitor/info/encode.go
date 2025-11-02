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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Encode interface {
	// String returns the string representation of the encoding model.
	// It concatenates all the key-value pairs of the Info map, separated by commas.
	// The key is separated from the value by a colon.
	String() string
	// Bytes returns the byte representation of the encoding model.
	// It is a JSON marshalled version of the encoding model.
	//
	Bytes() []byte
}

type encMod struct {
	Name string
	Info map[string]interface{}
}

func (e *encMod) String() string {
	if i := e.stringInfo(); len(i) > 0 {
		return e.stringClean(fmt.Sprintf("%s (%s)", e.Name, i))
	} else {
		return e.stringClean(e.Name)
	}
}

func (e *encMod) Bytes() []byte {
	return []byte(e.String())
}

func (e *encMod) stringInfo() string {
	var buf = bytes.NewBuffer(make([]byte, 0))
	for n, i := range e.Info {
		buf.WriteString(fmt.Sprintf("%s: %v,", n, i)) // nolint
	}
	return strings.Trim(strings.TrimSpace(buf.String()), ",")
}

func (e *encMod) stringClean(str string) string {
	str = strings.Replace(str, "\n", " ", -1) // nolint
	str = strings.Replace(str, "\r", "", -1)  // nolint
	return str
}

func (o *inf) getEncodeModel() Encode {
	return &encMod{
		Name: o.Name(),
		Info: o.Info(),
	}
}

// MarshalText returns the text representation of the encoding model.
// The returned text is a concatenation of the name and info fields, separated by a space.
// If the info field is empty, only the name is returned.
// The returned error is always nil.
func (o *inf) MarshalText() (text []byte, err error) {
	return o.getEncodeModel().Bytes(), nil
}

// MarshalJSON returns the JSON representation of the encoding model.
// The returned error is always nil.
func (o *inf) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.getEncodeModel())
}
