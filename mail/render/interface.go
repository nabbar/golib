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

import (
	"bytes"

	"github.com/go-hermes/hermes/v2"
	liberr "github.com/nabbar/golib/errors"
)

// Mailer defines the interface for generating and managing email templates.
// It provides methods to configure email properties, manage template data,
// and generate both HTML and plain text versions of emails.
//
// The Mailer uses the github.com/go-hermes/hermes/v2 library for template
// rendering and supports features like:
//   - Theme selection (Default, Flat)
//   - Text direction (LTR, RTL)
//   - Template variable replacement
//   - Product information (name, logo, link, copyright)
//   - Rich body content (intros, outros, tables, actions)
//
// Thread Safety:
// Mailer instances are not thread-safe. Use Clone() to create independent
// copies for concurrent operations.
//
// Example:
//
//	mailer := render.New()
//	mailer.SetName("My Company")
//	mailer.SetTheme(render.ThemeFlat)
//	body := &hermes.Body{
//	    Name: "John Doe",
//	    Intros: []string{"Welcome to our service!"},
//	}
//	mailer.SetBody(body)
//	htmlBuf, err := mailer.GenerateHTML()
type Mailer interface {
	// Clone creates a deep copy of the Mailer instance.
	// All slices and nested structures are copied independently to avoid
	// shared references, making the clone safe for concurrent use.
	Clone() Mailer

	// SetTheme sets the visual theme for the email template.
	// Available themes: ThemeDefault, ThemeFlat
	SetTheme(t Themes)
	// GetTheme returns the currently configured theme.
	GetTheme() Themes

	// SetTextDirection sets the text direction for the email content.
	// Use LeftToRight for LTR languages, RightToLeft for RTL languages.
	SetTextDirection(d TextDirection)
	// GetTextDirection returns the currently configured text direction.
	GetTextDirection() TextDirection

	// SetBody sets the email body content.
	// See github.com/go-hermes/hermes/v2 Body for structure details.
	SetBody(b *hermes.Body)
	// GetBody returns the current email body content.
	GetBody() *hermes.Body

	// SetCSSInline controls CSS inlining in HTML output.
	// Set to true to disable CSS inlining (useful for some email clients).
	SetCSSInline(disable bool)

	// SetName sets the product/company name displayed in the email.
	SetName(name string)
	// GetName returns the configured product/company name.
	GetName() string

	// SetCopyright sets the copyright text displayed in the email footer.
	SetCopyright(copy string)
	// GetCopyright returns the configured copyright text.
	GetCopyright() string

	// SetLink sets the primary product/company link (usually a URL).
	SetLink(link string)
	// GetLink returns the configured product/company link.
	GetLink() string

	// SetLogo sets the URL of the logo image displayed in the email header.
	SetLogo(logoUrl string)
	// GetLogo returns the configured logo URL.
	GetLogo() string

	// SetTroubleText sets the help/support text displayed when actions fail.
	// Example: "Having trouble? Contact support@example.com"
	SetTroubleText(text string)
	// GetTroubleText returns the configured trouble text.
	GetTroubleText() string

	// ParseData replaces template variables in all email content.
	// Variables in the format "{{key}}" are replaced with corresponding values.
	// This affects all text fields including product info, body content, tables, and actions.
	//
	// Example:
	//   data := map[string]string{
	//       "{{user}}": "John Doe",
	//       "{{code}}": "123456",
	//   }
	//   mailer.ParseData(data)
	ParseData(data map[string]string)

	// GenerateHTML generates the HTML version of the email.
	// Returns a buffer containing the complete HTML document.
	GenerateHTML() (*bytes.Buffer, liberr.Error)
	// GeneratePlainText generates the plain text version of the email.
	// Returns a buffer containing formatted plain text content.
	GeneratePlainText() (*bytes.Buffer, liberr.Error)
}

// New creates a new Mailer instance with default configuration.
// The default configuration includes:
//   - Theme: ThemeDefault
//   - Text Direction: LeftToRight
//   - CSS Inlining: Enabled
//   - Empty product information and body content
//
// Example:
//
//	mailer := render.New()
//	mailer.SetName("My Company")
//	mailer.SetLink("https://example.com")
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

// Clone creates a deep copy of the email instance.
// This method performs a complete deep copy of all fields, including:
//   - All string slices (Intros, Outros)
//   - Dictionary entries
//   - Table data and column configurations
//   - Actions with buttons
//
// The cloned instance is completely independent and safe for concurrent use.
// Modifications to the clone will not affect the original instance.
//
// Example:
//
//	original := render.New()
//	original.SetName("Company")
//	clone := original.Clone()
//	clone.SetName("Other Company") // Does not affect original
func (e *email) Clone() Mailer {
	// Deep copy of slices to avoid shared references
	intros := make([]string, len(e.b.Intros))
	copy(intros, e.b.Intros)

	outros := make([]string, len(e.b.Outros))
	copy(outros, e.b.Outros)

	dictionary := make([]hermes.Entry, len(e.b.Dictionary))
	copy(dictionary, e.b.Dictionary)

	actions := make([]hermes.Action, len(e.b.Actions))
	copy(actions, e.b.Actions)

	// Deep copy tables
	var tables []hermes.Table
	if e.b.Tables != nil {
		tables = make([]hermes.Table, len(e.b.Tables))
		for i, table := range e.b.Tables {
			// Deep copy table data
			var tableData [][]hermes.Entry
			if table.Data != nil {
				tableData = make([][]hermes.Entry, len(table.Data))
				for j, row := range table.Data {
					tableData[j] = make([]hermes.Entry, len(row))
					copy(tableData[j], row)
				}
			}

			// Deep copy table columns
			var tableColumns hermes.Columns
			if table.Columns.CustomWidth != nil {
				tableColumns.CustomWidth = make(map[string]string, len(table.Columns.CustomWidth))
				for k, v := range table.Columns.CustomWidth {
					tableColumns.CustomWidth[k] = v
				}
			}
			if table.Columns.CustomAlignment != nil {
				tableColumns.CustomAlignment = make(map[string]string, len(table.Columns.CustomAlignment))
				for k, v := range table.Columns.CustomAlignment {
					tableColumns.CustomAlignment[k] = v
				}
			}

			tables[i] = hermes.Table{
				Title:        table.Title,
				Data:         tableData,
				Columns:      tableColumns,
				Class:        table.Class,
				TitleUnsafe:  table.TitleUnsafe,
				Footer:       table.Footer,
				FooterUnsafe: table.FooterUnsafe,
			}
		}
	}

	return &email{
		p: hermes.Product{
			Name:        e.p.Name,
			Link:        e.p.Link,
			Logo:        e.p.Logo,
			Copyright:   e.p.Copyright,
			TroubleText: e.p.TroubleText,
		},
		b: &hermes.Body{
			Name:         e.b.Name,
			Intros:       intros,
			Dictionary:   dictionary,
			Tables:       tables,
			Actions:      actions,
			Outros:       outros,
			Greeting:     e.b.Greeting,
			Signature:    e.b.Signature,
			Title:        e.b.Title,
			FreeMarkdown: e.b.FreeMarkdown,
		},
		t: e.t,
		d: e.d,
		c: e.c,
	}
}
