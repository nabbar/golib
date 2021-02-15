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

package mail

type Encoding uint8

const (
	// EncodingNone turns off encoding on the message body
	EncodingNone Encoding = iota

	// EncodingBinary is equal to EncodingNone, but string is set to binrary instead of none
	EncodingBinary

	// EncodingBase64 sets the message body encoding to base64
	EncodingBase64

	// EncodingQuotedPrintable sets the message body encoding to quoted-printable
	EncodingQuotedPrintable
)

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

func (e Encoding) getEncoding() string {
	switch e {
	case EncodingNone, EncodingBinary:
		return "binary"
	case EncodingBase64:
		return "base64"
	case EncodingQuotedPrintable:
		return "quoted-printable"
	}

	return EncodingNone.getEncoding()
}
