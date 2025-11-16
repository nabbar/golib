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

package render

import (
	"fmt"

	libhms "github.com/go-hermes/hermes/v2"
	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
)

// Config represents the configuration structure for creating a Mailer instance.
// It supports multiple serialization formats (JSON, YAML, TOML) and includes
// validation tags for ensuring configuration correctness.
//
// All fields except DisableCSSInline are required and must pass validation.
// The Link and Logo fields must be valid URLs.
//
// This struct is designed to be used with configuration files or environment
// variables via libraries like Viper or similar configuration management tools.
//
// Example JSON configuration:
//
//	{
//	  "theme": "flat",
//	  "direction": "ltr",
//	  "name": "My Company",
//	  "link": "https://example.com",
//	  "logo": "https://example.com/logo.png",
//	  "copyright": "© 2024 My Company",
//	  "troubleText": "Need help? Contact support@example.com",
//	  "disableCSSInline": false,
//	  "body": {
//	    "name": "John Doe",
//	    "intros": ["Welcome to our service!"]
//	  }
//	}
type Config struct {
	// Theme specifies the email template theme ("default" or "flat")
	Theme string `json:"theme,omitempty" yaml:"theme,omitempty" toml:"theme,omitempty" mapstructure:"theme,omitempty" validate:"required"`
	// Direction specifies the text direction ("ltr" or "rtl")
	Direction string `json:"direction,omitempty" yaml:"direction,omitempty" toml:"direction,omitempty" mapstructure:"direction,omitempty" validate:"required"`
	// Name is the product/company name displayed in the email
	Name string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty" mapstructure:"name,omitempty" validate:"required"`
	// Link is the primary product/company URL (must be a valid URL)
	Link string `json:"link,omitempty" yaml:"link,omitempty" toml:"link,omitempty" mapstructure:"link,omitempty" validate:"required,url"`
	// Logo is the URL of the logo image (must be a valid URL)
	Logo string `json:"logo,omitempty" yaml:"logo,omitempty" toml:"logo,omitempty" mapstructure:"logo,omitempty" validate:"required,url"`
	// Copyright is the copyright text displayed in the footer
	Copyright string `json:"copyright,omitempty" yaml:"copyright,omitempty" toml:"copyright,omitempty" mapstructure:"copyright,omitempty" validate:"required"`
	// TroubleText is the help/support text displayed for troubleshooting
	TroubleText string `json:"troubleText,omitempty" yaml:"troubleText,omitempty" toml:"troubleText,omitempty" mapstructure:"troubleText,omitempty" validate:"required"`
	// DisableCSSInline controls whether to disable CSS inlining in HTML output
	DisableCSSInline bool `json:"disableCSSInline,omitempty" yaml:"disableCSSInline,omitempty" toml:"disableCSSInline,omitempty" mapstructure:"disableCSSInline,omitempty"`
	// Body contains the email body content structure
	// See github.com/go-hermes/hermes/v2 Body for field details
	Body libhms.Body `json:"body" yaml:"body" toml:"body" mapstructure:"body" validate:"required"`
}

// Validate validates the Config struct using validator tags.
// It checks all required fields and validates URL formats for Link and Logo.
//
// Returns:
//   - nil if validation succeeds
//   - liberr.Error containing all validation errors if validation fails
//
// Validation rules:
//   - All fields except DisableCSSInline are required
//   - Link must be a valid URL
//   - Logo must be a valid URL
//
// Example:
//
//	config := render.Config{
//	    Theme: "flat",
//	    Direction: "ltr",
//	    Name: "Company",
//	    Link: "https://example.com",
//	    Logo: "https://example.com/logo.png",
//	    Copyright: "© 2024",
//	    TroubleText: "Help",
//	    Body: hermes.Body{},
//	}
//	if err := config.Validate(); err != nil {
//	    // Handle validation errors
//	}
func (c Config) Validate() liberr.Error {
	err := ErrorMailerConfigInvalid.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

// NewMailer creates a new Mailer instance from the configuration.
// It parses the theme and direction strings and initializes all fields.
//
// Note: This method does not validate the configuration. Call Validate()
// before NewMailer() to ensure the configuration is valid.
//
// Returns:
//   - A configured Mailer instance ready to use
//
// Example:
//
//	config := render.Config{
//	    Theme: "flat",
//	    Direction: "ltr",
//	    Name: "My Company",
//	    Link: "https://example.com",
//	    Logo: "https://example.com/logo.png",
//	    Copyright: "© 2024",
//	    TroubleText: "Need help?",
//	    Body: hermes.Body{Name: "User"},
//	}
//	if err := config.Validate(); err == nil {
//	    mailer := config.NewMailer()
//	    htmlBuf, _ := mailer.GenerateHTML()
//	}
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
