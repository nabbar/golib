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
	"strings"

	"github.com/go-hermes/hermes/v2"
)

// Themes represents the available email template themes.
// Each theme provides a different visual style for the generated emails.
//
// The themes are based on github.com/go-hermes/hermes/v2 theme implementations.
type Themes uint8

const (
	// ThemeDefault is the default theme provided by Hermes.
	// It offers a classic, clean email design with a centered layout.
	ThemeDefault Themes = iota

	// ThemeFlat is the flat theme provided by Hermes.
	// It offers a modern, minimalist email design with flat colors.
	ThemeFlat
)

// getHermes converts the Themes enum to the corresponding hermes.Theme implementation.
// This is an internal method used by the email generation process.
func (t Themes) getHermes() hermes.Theme {
	switch t {
	case ThemeDefault:
		return &hermes.Default{}
	case ThemeFlat:
		return &hermes.Flat{}
	}

	return ThemeDefault.getHermes()
}

// String returns the string representation of the theme.
// The string value matches the theme name from the Hermes library.
//
// Example:
//
//	theme := render.ThemeFlat
//	fmt.Println(theme.String()) // Output: "flat"
func (t Themes) String() string {
	return t.getHermes().Name()
}

// ParseTheme parses a theme name string and returns the corresponding Themes enum value.
// The parsing is case-insensitive.
//
// Supported theme names:
//   - "default" -> ThemeDefault
//   - "flat" -> ThemeFlat
//
// If the theme name is not recognized, ThemeDefault is returned.
//
// Example:
//
//	theme := render.ParseTheme("flat")
//	theme = render.ParseTheme("FLAT")     // Also works
//	theme = render.ParseTheme("unknown")  // Returns ThemeDefault
func ParseTheme(theme string) Themes {
	switch strings.ToLower(theme) {
	case strings.ToLower(ThemeDefault.String()):
		return ThemeDefault
	case strings.ToLower(ThemeFlat.String()):
		return ThemeFlat
	}

	return ThemeDefault
}
