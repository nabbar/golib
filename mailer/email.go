package mailer

import "github.com/matcornic/hermes/v2"

type email struct {
	t Themes
	d TextDirection
	p hermes.Product
	b *hermes.Body
	c bool
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
