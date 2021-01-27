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
