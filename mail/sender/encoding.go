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

import "strings"

// Encoding defines the transfer encoding method for email body content.
//
// Transfer encoding is used to represent 8-bit data in a 7-bit environment,
// ensuring that email content can be safely transmitted through all SMTP servers.
// Different encoding methods offer different trade-offs between size and
// readability.
//
// Example usage:
//
//	mail.SetEncoding(sender.EncodingBase64)
//
// See RFC 2045 for more details on MIME content transfer encodings.
type Encoding uint8

const (
	// EncodingNone indicates no transfer encoding is applied to the message body.
	// The content is sent as-is, which is suitable for 7-bit ASCII text.
	//
	// Use this for simple ASCII-only emails without special characters.
	// Not recommended for binary data or text with extended characters.
	//
	// Example:
	//	mail.SetEncoding(sender.EncodingNone)
	EncodingNone Encoding = iota

	// EncodingBinary is functionally equivalent to EncodingNone but explicitly
	// declares the content as binary data. This encoding is rarely used in practice
	// as many mail servers don't support true 8-bit binary transmission.
	//
	// Use this only if you specifically need to indicate binary content without
	// actually encoding it.
	EncodingBinary

	// EncodingBase64 encodes the message body using Base64 encoding (RFC 2045).
	// This encoding converts binary data into ASCII text using a 64-character alphabet.
	//
	// Characteristics:
	//   - Safe for all email servers and intermediate systems
	//   - Increases size by approximately 33%
	//   - Ideal for binary attachments and non-ASCII text
	//   - Most commonly used encoding for email attachments
	//
	// Use this for:
	//   - Email attachments (images, documents, etc.)
	//   - HTML content with non-ASCII characters
	//   - When maximum compatibility is required
	//
	// Example:
	//	mail.SetEncoding(sender.EncodingBase64)
	EncodingBase64

	// EncodingQuotedPrintable encodes the message body using Quoted-Printable encoding (RFC 2045).
	// This encoding represents special characters using "=" followed by their hex value.
	//
	// Characteristics:
	//   - Preserves readability for ASCII text
	//   - Minimal size increase for mostly ASCII content
	//   - Efficient for text with occasional special characters
	//   - Lines longer than 76 characters are soft-wrapped
	//
	// Use this for:
	//   - Text emails with occasional non-ASCII characters
	//   - HTML emails with mostly ASCII content
	//   - When human readability of encoded content is important
	//
	// Example:
	//	mail.SetEncoding(sender.EncodingQuotedPrintable)
	EncodingQuotedPrintable
)

// String returns a human-readable string representation of the Encoding.
//
// Returns:
//   - "None" for EncodingNone
//   - "Binary" for EncodingBinary
//   - "Base 64" for EncodingBase64
//   - "Quoted Printable" for EncodingQuotedPrintable
//   - Defaults to "None" for unknown values
//
// This method is useful for logging, configuration display, and debugging.
func (e Encoding) String() string {
	switch e {
	case EncodingBinary:
		return "Binary"
	case EncodingBase64:
		return "Base 64"
	case EncodingQuotedPrintable:
		return "Quoted Printable"
	case EncodingNone:
		return "None"
	}
	return EncodingNone.String()
}

// ParseEncoding converts a string representation into an Encoding value.
// The comparison is case-insensitive for flexibility.
//
// Parameters:
//   - s: String representation of the encoding. Valid values are:
//     "None", "Binary", "Base 64", "Quoted Printable" (case-insensitive)
//
// Returns:
//   - The corresponding Encoding value
//   - EncodingNone if the string doesn't match any known encoding
//
// Example:
//
//	encoding := sender.ParseEncoding("Base 64")      // Returns EncodingBase64
//	encoding := sender.ParseEncoding("base 64")      // Also returns EncodingBase64
//	encoding := sender.ParseEncoding("unknown")      // Returns EncodingNone
func ParseEncoding(s string) Encoding {
	switch strings.ToUpper(s) {
	case strings.ToUpper(EncodingBinary.String()):
		return EncodingBinary
	case strings.ToUpper(EncodingBase64.String()):
		return EncodingBase64
	case strings.ToUpper(EncodingQuotedPrintable.String()):
		return EncodingQuotedPrintable
	default:
		return EncodingNone
	}
}
