/*
 * MIT License
 *
 * Copyright (c) 2021 Nicolas JUHEL
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

package mailer

import (
	"fmt"

	libval "github.com/go-playground/validator/v10"
	libhms "github.com/matcornic/hermes/v2"
	liberr "github.com/nabbar/golib/errors"
)

type Config struct {
	Theme            string      `json:"theme,omitempty" yaml:"theme,omitempty" toml:"theme,omitempty" mapstructure:"theme,omitempty" validate:"required"`
	Direction        string      `json:"direction,omitempty" yaml:"direction,omitempty" toml:"direction,omitempty" mapstructure:"direction,omitempty" validate:"required"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty" mapstructure:"name,omitempty" validate:"required"`
	Link             string      `json:"link,omitempty" yaml:"link,omitempty" toml:"link,omitempty" mapstructure:"link,omitempty" validate:"required,url"`
	Logo             string      `json:"logo,omitempty" yaml:"logo,omitempty" toml:"logo,omitempty" mapstructure:"logo,omitempty" validate:"required,url"`
	Copyright        string      `json:"copyright,omitempty" yaml:"copyright,omitempty" toml:"copyright,omitempty" mapstructure:"copyright,omitempty" validate:"required"`
	TroubleText      string      `json:"troubleText,omitempty" yaml:"troubleText,omitempty" toml:"troubleText,omitempty" mapstructure:"troubleText,omitempty" validate:"required"`
	DisableCSSInline bool        `json:"disableCSSInline,omitempty" yaml:"disableCSSInline,omitempty" toml:"disableCSSInline,omitempty" mapstructure:"disableCSSInline,omitempty"`
	Body             libhms.Body `json:"body" yaml:"body" toml:"body" mapstructure:"body" validate:"required"`
}

func (c Config) Validate() liberr.Error {
	err := ErrorMailerConfigInvalid.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.AddParent(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c Config) NewMailer() Mailer {
	return &email{
		t: ParseTheme(c.Theme),
		d: ParseTextDirection(c.Direction),
		p: libhms.Product{
			Name:        c.Name,
			Link:        c.Link,
			Logo:        c.Logo,
			Copyright:   c.Copyright,
			TroubleText: c.TroubleText,
		},
		b: &c.Body,
		c: c.DisableCSSInline,
	}
}
