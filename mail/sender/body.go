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

import (
	"io"
)

// Body represents a content part of an email message with its associated content type.
// An email can contain multiple body parts, such as plain text and HTML alternatives.
//
// The body content is provided as an io.Reader, allowing for efficient streaming
// of large email content without loading everything into memory.
//
// Example usage:
//
//	plainText := sender.NewBody(sender.ContentPlainText, strings.NewReader("Hello"))
//	htmlBody := sender.NewBody(sender.ContentHTML, strings.NewReader("<p>Hello</p>"))
//
// See ContentType for available content type options.
type Body struct {
	contentType ContentType
	body        io.Reader
}

// NewBody creates a new Body instance with the specified content type and reader.
//
// Parameters:
//   - ct: The content type of the body (e.g., ContentPlainText, ContentHTML)
//   - body: An io.Reader providing the body content
//
// Returns a configured Body instance that can be added to a Mail object.
//
// Example:
//
//	body := sender.NewBody(sender.ContentPlainText, bytes.NewReader([]byte("Email content")))
//	mail.SetBody(sender.ContentPlainText, body)
func NewBody(ct ContentType, body io.Reader) Body {
	return Body{
		contentType: ct,
		body:        body,
	}
}
