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

package sender

// ContentType defines the type of content in an email body part.
//
// Email bodies can contain multiple parts with different content types,
// allowing clients to display the most appropriate version. For example,
// an email can include both plain text and HTML versions, with the email
// client choosing which to display based on its capabilities.
//
// Example usage:
//
//	// Add plain text body
//	mail.SetBody(sender.ContentPlainText, strings.NewReader("Hello World"))
//
//	// Add HTML alternative
//	mail.AddBody(sender.ContentHTML, strings.NewReader("<p>Hello World</p>"))
//
// Best practice is to always include a ContentPlainText version for
// maximum compatibility with all email clients.
type ContentType uint8

const (
	// ContentPlainText represents plain text content (MIME type: text/plain).
	// This is the most basic and universally supported content type.
	// Plain text emails are displayed exactly as written, without any formatting.
	//
	// Use this for simple text-only emails or as a fallback when also
	// providing an HTML version.
	//
	// Example:
	//	mail.SetBody(sender.ContentPlainText, strings.NewReader("Hello"))
	ContentPlainText ContentType = iota

	// ContentHTML represents HTML content (MIME type: text/html).
	// HTML emails can include rich formatting, colors, images, and links.
	// Most modern email clients support HTML rendering.
	//
	// When using HTML content, it's recommended to also provide a plain text
	// alternative using ContentPlainText for email clients that don't support
	// HTML or for users who prefer plain text.
	//
	// Example:
	//	mail.AddBody(sender.ContentHTML, strings.NewReader("<p><b>Hello</b></p>"))
	ContentHTML
)

// String returns a human-readable string representation of the ContentType.
//
// Returns:
//   - "Plain Text" for ContentPlainText
//   - "HTML" for ContentHTML
//   - Defaults to "Plain Text" for unknown values
//
// This method is useful for logging and debugging purposes.
func (c ContentType) String() string {
	switch c {
	case ContentPlainText:
		return "Plain Text"
	case ContentHTML:
		return "HTML"
	}

	return ContentPlainText.String()
}
