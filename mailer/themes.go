package mailer

import (
	"strings"

	"github.com/matcornic/hermes/v2"
)

type Themes uint8

const (
	ThemeDefault Themes = iota
	ThemeFlat
)

func (t Themes) getHermes() hermes.Theme {
	switch t {
	case ThemeDefault:
		return &hermes.Default{}
	case ThemeFlat:
		return &hermes.Flat{}
	}

	return ThemeDefault.getHermes()
}

func (t Themes) String() string {
	return t.getHermes().Name()
}

func ParseTheme(theme string) Themes {
	switch strings.ToLower(theme) {
	case strings.ToLower(ThemeDefault.String()):
		return ThemeDefault
	case strings.ToLower(ThemeFlat.String()):
		return ThemeFlat
	}

	return ThemeDefault
}
