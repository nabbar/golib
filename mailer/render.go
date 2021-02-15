/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package mailer

import (
	"bytes"

	"github.com/matcornic/hermes/v2"
	liberr "github.com/nabbar/golib/errors"
)

func (e *email) GenerateHTML() (*bytes.Buffer, liberr.Error) {
	var buf = bytes.NewBuffer(make([]byte, 0))

	h := hermes.Hermes{
		Theme:              e.t.getHermes(),
		TextDirection:      e.d.getDirection(),
		Product:            e.p,
		DisableCSSInlining: e.c,
	}

	if p, e := h.GenerateHTML(hermes.Email{
		Body: *e.b,
	}); e != nil {
		return nil, ErrorMailerHtml.ErrorParent(e)
	} else {
		buf.WriteString(p)
	}

	return buf, nil
}

func (e *email) GeneratePlainText() (*bytes.Buffer, liberr.Error) {
	var buf = bytes.NewBuffer(make([]byte, 0))

	h := hermes.Hermes{
		Theme:              e.t.getHermes(),
		TextDirection:      e.d.getDirection(),
		Product:            e.p,
		DisableCSSInlining: e.c,
	}

	if p, e := h.GeneratePlainText(hermes.Email{
		Body: *e.b,
	}); e != nil {
		return nil, ErrorMailerText.ErrorParent(e)
	} else {
		buf.WriteString(p)
	}

	return buf, nil
}
