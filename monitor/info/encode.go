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
	String() string
	Bytes() []byte
}

type encodingModel struct {
	Name string
	Info map[string]interface{}
}

func (e *encodingModel) stringInfo() string {
	var (
		buf = bytes.NewBuffer(make([]byte, 0))
	)

	for n, i := range e.Info {
		buf.WriteString(fmt.Sprintf("%s: %v,", n, i))
	}

	return strings.Trim(strings.TrimSpace(buf.String()), ",")
}

func (e *encodingModel) stringClean(str string) string {
	str = strings.Replace(str, "\n", " ", -1)
	str = strings.Replace(str, "\r", "", -1)
	return str
}

func (e *encodingModel) String() string {
	if i := e.stringInfo(); len(i) > 0 {
		return e.stringClean(fmt.Sprintf("%s (%s)", e.Name, i))
	} else {
		return e.stringClean(e.Name)
	}
}

func (e *encodingModel) Bytes() []byte {
	return []byte(e.String())
}

func (o *inf) getEncodeModel() Encode {
	return &encodingModel{
		Name: o.Name(),
		Info: o.Info(),
	}
}

func (o *inf) MarshalText() (text []byte, err error) {
	return o.getEncodeModel().Bytes(), nil
}

func (o *inf) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.getEncodeModel())
}
