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

package status

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	monpol "github.com/nabbar/golib/monitor/pool"

	ginsdk "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	liberr "github.com/nabbar/golib/errors"

	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
)

const (
	encTextSepStatus = ": "
	encTextSepPart   = " | "
	encTextSepTime   = " / "
)

type Encode interface {
	encoding.TextMarshaler

	String() string
	Bytes() []byte

	GinRender(c *ginsdk.Context, isText bool, isShort bool)
	GinCode() int
}

type encodeModel struct {
	Name      string
	Release   string
	Hash      string
	DateBuild time.Time
	Status    monsts.Status
	Message   string
	Component montps.Pool
	code      int
}

func (e *encodeModel) MarshalText() (text []byte, err error) {
	return e.Bytes(), nil
}

func (e *encodeModel) GinCode() int {
	return e.code
}

func (e *encodeModel) GinRender(c *ginsdk.Context, isText bool, isShort bool) {
	if isShort {
		e.Component = monpol.New(func() context.Context {
			return c
		})
	}

	if isText {
		c.Render(e.code, render.Data{
			ContentType: ginsdk.MIMEPlain,
			Data:        e.Bytes(),
		})
	} else {
		c.JSON(e.code, *e)
	}
}

func (e *encodeModel) cleanString(str string) string {
	str = strings.Replace(str, "\n", " ", -1)
	str = strings.Replace(str, "\r", "", -1)
	return str
}

func (e *encodeModel) stringName() string {
	var inf []string

	if len(e.Release) > 0 {
		inf = append(inf, e.Release)
	}

	if len(e.Hash) > 0 {
		inf = append(inf, e.Hash)
	}

	if !e.DateBuild.IsZero() {
		inf = append(inf, e.DateBuild.Format(time.RFC3339))
	}

	if len(inf) > 0 {
		return fmt.Sprintf("%s (%s)", e.Name, strings.Join(inf, " "))
	} else {
		return e.Name
	}
}

func (e *encodeModel) stringPart() string {
	item := make([]string, 0)
	item = append(item, e.stringName())

	if len(e.Message) > 0 {
		item = append(item, e.Message)
	}

	return strings.Join(item, encTextSepPart)
}

func (e *encodeModel) String() string {
	var buf = bytes.NewBuffer(make([]byte, 0))

	buf.WriteString(e.Status.String() + encTextSepStatus + e.stringPart())
	buf.WriteRune('\n')

	if p, err := e.Component.MarshalText(); err == nil {
		buf.Write(p)
	}

	return e.cleanString(buf.String())
}

func (e *encodeModel) Bytes() []byte {
	return []byte(e.String())
}

func (o *sts) getEncodeModel() Encode {
	o.m.RLock()
	defer o.m.RUnlock()

	var (
		m string
		s monsts.Status
	)

	s, m = o.getStatus()

	return &encodeModel{
		Name:      o.fn(),
		Release:   o.fr(),
		Hash:      o.fh(),
		DateBuild: o.fd(),
		Status:    s,
		Message:   m,
		Component: o._getPool(),
		code:      o.cfgGetReturnCode(s),
	}
}

func (o *sts) getMarshal() (Encode, liberr.Error) {
	if !o.checkFunc() {
		return nil, ErrorParamEmpty.ErrorParent(fmt.Errorf("missing status info for API"))
	}
	return o.getEncodeModel(), nil
}

func (o *sts) MarshalText() (text []byte, err error) {
	if enc, e := o.getMarshal(); e != nil {
		return nil, e
	} else {
		return enc.MarshalText()
	}
}

func (o *sts) MarshalJSON() ([]byte, error) {
	if enc, e := o.getMarshal(); e != nil {
		return nil, e
	} else {
		return json.Marshal(enc)
	}
}
