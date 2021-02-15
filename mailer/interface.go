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

type Mailer interface {
	SetTheme(t Themes)
	GetTheme() Themes

	SetTextDirection(d TextDirection)
	GetTextDirection() TextDirection

	SetBody(b *hermes.Body)
	GetBody() *hermes.Body

	SetCSSInline(disable bool)

	SetName(name string)
	GetName() string

	SetCopyright(copy string)
	GetCopyright() string

	SetLink(link string)
	GetLink() string

	SetLogo(logoUrl string)
	GetLogo() string

	SetTroubleText(text string)
	GetTroubleText() string

	GenerateHTML() (*bytes.Buffer, liberr.Error)
	GeneratePlainText() (*bytes.Buffer, liberr.Error)
}

func New() Mailer {
	return &email{
		t: ThemeDefault,
		d: LeftToRight,
		p: hermes.Product{
			Name:        "",
			Link:        "",
			Logo:        "",
			Copyright:   "",
			TroubleText: "",
		},
		b: &hermes.Body{},
		c: false,
	}
}
