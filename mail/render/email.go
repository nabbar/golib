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

package render

import "github.com/go-hermes/hermes/v2"

// email is the internal implementation of the Mailer interface.
// It encapsulates all configuration and content needed to generate emails.
//
// Fields:
//   - p: Product information (name, logo, link, copyright, trouble text)
//   - b: Email body content (intros, outros, tables, actions, etc.)
//   - t: Visual theme for the email template
//   - d: Text direction (LTR or RTL)
//   - c: CSS inlining control (true = disable inlining)
type email struct {
	p hermes.Product // Product/company information
	b *hermes.Body   // Email body content and structure
	t Themes         // Visual theme selection
	d TextDirection  // Text direction (LTR/RTL)
	c bool           // CSS inlining disabled flag
}

func (e *email) SetTheme(t Themes) {
	e.t = t
}

func (e *email) GetTheme() Themes {
	return e.t
}

func (e *email) SetTextDirection(d TextDirection) {
	e.d = d
}

func (e *email) GetTextDirection() TextDirection {
	return e.d
}

func (e *email) SetBody(b *hermes.Body) {
	e.b = b
}

func (e *email) GetBody() *hermes.Body {
	return e.b
}

func (e *email) SetCSSInline(disable bool) {
	e.c = disable
}

func (e *email) SetName(name string) {
	e.p.Name = name
}

func (e *email) GetName() string {
	return e.p.Name
}

func (e *email) SetCopyright(copy string) {
	e.p.Copyright = copy
}

func (e *email) GetCopyright() string {
	return e.p.Copyright
}

func (e *email) SetLink(link string) {
	e.p.Link = link
}

func (e *email) GetLink() string {
	return e.p.Link
}

func (e *email) SetLogo(logoUrl string) {
	e.p.Logo = logoUrl
}

func (e *email) GetLogo() string {
	return e.p.Logo
}

func (e *email) SetTroubleText(text string) {
	e.p.TroubleText = text
}

func (e *email) GetTroubleText() string {
	return e.p.TroubleText
}
