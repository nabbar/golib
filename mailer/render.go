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
	"strings"

	"github.com/matcornic/hermes/v2"
	liberr "github.com/nabbar/golib/errors"
)

func (e *email) ParseData(data map[string]string) {
	for k, v := range data {
		e.p.Copyright = strings.ReplaceAll(e.p.Copyright, k, v)
		e.p.Link = strings.ReplaceAll(e.p.Link, k, v)
		e.p.Logo = strings.ReplaceAll(e.p.Logo, k, v)
		e.p.Name = strings.ReplaceAll(e.p.Name, k, v)
		e.p.TroubleText = strings.ReplaceAll(e.p.TroubleText, k, v)

		e.b.Greeting = strings.ReplaceAll(e.b.Greeting, k, v)
		e.b.Name = strings.ReplaceAll(e.b.Name, k, v)
		e.b.Signature = strings.ReplaceAll(e.b.Signature, k, v)
		e.b.Title = strings.ReplaceAll(e.b.Title, k, v)

		e.b.FreeMarkdown = hermes.Markdown(strings.ReplaceAll(string(e.b.FreeMarkdown), k, v))

		for i := range e.b.Intros {
			e.b.Intros[i] = strings.ReplaceAll(e.b.Intros[i], k, v)
		}

		for i := range e.b.Outros {
			e.b.Outros[i] = strings.ReplaceAll(e.b.Outros[i], k, v)
		}

		for i := range e.b.Dictionary {
			e.b.Dictionary[i].Key = strings.ReplaceAll(e.b.Dictionary[i].Key, k, v)
			e.b.Dictionary[i].Value = strings.ReplaceAll(e.b.Dictionary[i].Value, k, v)
		}

		for i := range e.b.Table.Data {
			for j := range e.b.Table.Data[i] {
				e.b.Table.Data[i][j].Key = strings.ReplaceAll(e.b.Table.Data[i][j].Key, k, v)
				e.b.Table.Data[i][j].Value = strings.ReplaceAll(e.b.Table.Data[i][j].Value, k, v)
			}
		}

		for i := range e.b.Actions {
			e.b.Actions[i].Instructions = strings.ReplaceAll(e.b.Actions[i].Instructions, k, v)
			e.b.Actions[i].InviteCode = strings.ReplaceAll(e.b.Actions[i].InviteCode, k, v)
			e.b.Actions[i].Button.Link = strings.ReplaceAll(e.b.Actions[i].Button.Link, k, v)
			e.b.Actions[i].Button.Text = strings.ReplaceAll(e.b.Actions[i].Button.Text, k, v)
			e.b.Actions[i].Button.Color = strings.ReplaceAll(e.b.Actions[i].Button.Color, k, v)
			e.b.Actions[i].Button.TextColor = strings.ReplaceAll(e.b.Actions[i].Button.TextColor, k, v)
		}
	}
}

func (e *email) GenerateHTML() (*bytes.Buffer, liberr.Error) {
	return e.generated(e.genHtml)
}

func (e *email) GeneratePlainText() (*bytes.Buffer, liberr.Error) {
	return e.generated(e.genText)
}

func (e email) generated(f func(h hermes.Hermes, m hermes.Email) (string, error)) (*bytes.Buffer, liberr.Error) {
	var buf = bytes.NewBuffer(make([]byte, 0))

	h := hermes.Hermes{
		Theme:              e.t.getHermes(),
		TextDirection:      e.d.getDirection(),
		Product:            e.p,
		DisableCSSInlining: e.c,
	}

	if p, e := f(h, hermes.Email{
		Body: *e.b,
	}); e != nil {
		return nil, ErrorMailerText.Error(e)
	} else {
		buf.WriteString(p)
	}

	return buf, nil
}

func (e *email) genText(h hermes.Hermes, m hermes.Email) (string, error) {
	return h.GeneratePlainText(m)
}

func (e *email) genHtml(h hermes.Hermes, m hermes.Email) (string, error) {
	return h.GenerateHTML(m)
}
