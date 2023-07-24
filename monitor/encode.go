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

package monitor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
	moninf "github.com/nabbar/golib/monitor/types"
)

const (
	encTextSepStatus = ": "
	encTextSepPart   = " | "
	encTextSepTime   = " / "
)

type Encode interface {
	String() string
	Bytes() []byte
}

type encodeModel struct {
	Status monsts.Status
	Name   string
	Info   moninf.Info

	Latency  string
	Uptime   string
	Downtime string

	Message string
}

func (e *encodeModel) Bytes() []byte {
	return []byte(e.String())
}

func (e *encodeModel) String() string {
	return e.Status.String() + encTextSepStatus + e.stringPart()
}

func (e *encodeModel) stringDuration() string {
	part := append(make([]string, 0), e.Latency, e.Uptime, e.Downtime)
	return strings.Join(part, encTextSepTime)
}

func (e *encodeModel) stringName() string {
	var inf string

	if e.Info != nil {
		i, _ := e.Info.MarshalText()
		inf = string(i)
	}

	if len(inf) > 0 {
		return fmt.Sprintf("%s (%s)", e.Name, inf)
	} else {
		return e.Name
	}
}

func (e *encodeModel) stringPart() string {
	item := make([]string, 0)
	item = append(item, e.stringName())
	item = append(item, e.stringDuration())

	if len(e.Message) > 0 {
		item = append(item, e.Message)
	}

	return strings.Join(item, encTextSepPart)
}

func (o *mon) getEncodeModel() Encode {
	return &encodeModel{
		Status:   o.Status(),
		Name:     o.Name(),
		Info:     o.InfoGet(),
		Latency:  o.Latency().Truncate(time.Millisecond).String(),
		Uptime:   o.Uptime().Truncate(time.Second).String(),
		Downtime: o.Downtime().Truncate(time.Second).String(),
		Message:  o.Message(),
	}
}

func (o *mon) MarshalText() (text []byte, err error) {
	return o.getEncodeModel().Bytes(), nil
}

func (o *mon) MarshalJSON() (text []byte, err error) {
	return json.Marshal(o.getEncodeModel())
}
