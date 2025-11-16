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

// TextDirection represents the text reading direction for email content.
// This affects how the email content is laid out and displayed in email clients.
//
// Common use cases:
//   - LeftToRight: For languages like English, French, Spanish, etc.
//   - RightToLeft: For languages like Arabic, Hebrew, Persian, etc.
type TextDirection uint8

const (
	// LeftToRight indicates left-to-right text direction.
	// Used for most Western languages (English, French, Spanish, German, etc.).
	LeftToRight TextDirection = iota

	// RightToLeft indicates right-to-left text direction.
	// Used for RTL languages (Arabic, Hebrew, Persian, Urdu, etc.).
	RightToLeft
)

// getDirection converts the TextDirection enum to the corresponding hermes.TextDirection.
// This is an internal method used by the email generation process.
func (d TextDirection) getDirection() hermes.TextDirection {
	switch d {
	case LeftToRight:
		return hermes.TDLeftToRight
	case RightToLeft:
		return hermes.TDRightToLeft
	}

	return LeftToRight.getDirection()
}

// String returns the string representation of the text direction.
//
// Returns:
//   - "Left->Right" for LeftToRight
//   - "Right->Left" for RightToLeft
//
// Example:
//
//	dir := render.RightToLeft
//	fmt.Println(dir.String()) // Output: "Right->Left"
func (d TextDirection) String() string {
	switch d {
	case LeftToRight:
		return "Left->Right"
	case RightToLeft:
		return "Right->Left"
	}

	return LeftToRight.String()
}

// ParseTextDirection parses a text direction string and returns the corresponding TextDirection enum.
// The parsing is case-insensitive and supports multiple formats.
//
// Supported formats:
//   - "ltr", "LTR" -> LeftToRight
//   - "rtl", "RTL" -> RightToLeft
//   - "left", "left-to-right", "Left->Right" -> LeftToRight
//   - "right", "right-to-left", "Right->Left" -> RightToLeft
//
// If the direction string is not recognized or empty, LeftToRight is returned as the default.
//
// Example:
//
//	dir := render.ParseTextDirection("rtl")
//	dir = render.ParseTextDirection("right-to-left")
//	dir = render.ParseTextDirection("RTL")
//	dir = render.ParseTextDirection("unknown") // Returns LeftToRight
func ParseTextDirection(direction string) TextDirection {
	d := strings.ToLower(direction)

	// Check for common abbreviations first
	if strings.Contains(d, "rtl") {
		return RightToLeft
	}
	if strings.Contains(d, "ltr") {
		return LeftToRight
	}

	l := strings.Index(d, "left")
	r := strings.Index(d, "right")

	// If both "left" and "right" are found, check which comes first
	// "right->left" or "right-to-left" means RightToLeft
	if l >= 0 && r >= 0 && r < l {
		return RightToLeft
	} else if r >= 0 && l < 0 {
		// Only "right" found without "left" - assume RightToLeft
		return RightToLeft
	}

	return LeftToRight
}
