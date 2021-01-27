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
