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
