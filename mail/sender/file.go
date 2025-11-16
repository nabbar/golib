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

import "io"

// File represents an attachment or inline file that can be added to an email message.
//
// Files can be attached to emails in two ways:
//   - As regular attachments that appear in the attachment list
//   - As inline attachments embedded in the email body (e.g., images in HTML)
//
// The file data is provided as an io.ReadCloser, allowing for efficient streaming
// of large files and automatic cleanup after the email is sent.
//
// Example usage:
//
//	file, _ := os.Open("document.pdf")
//	attachment := sender.NewFile("document.pdf", "application/pdf", file)
//	mail.AddAttachment("document.pdf", "application/pdf", file, false)
//
// See Mail.AddAttachment for adding files to an email.
type File struct {
	name string        // Filename as it will appear in the email
	mime string        // MIME type of the file (e.g., "application/pdf", "image/png")
	data io.ReadCloser // File content as a readable stream
}

// NewFile creates a new File instance representing an email attachment.
//
// Parameters:
//   - name: The filename as it should appear in the email (e.g., "report.pdf")
//   - mime: The MIME type of the file (e.g., "application/pdf", "image/png", "text/plain")
//   - data: An io.ReadCloser providing the file content. The caller is responsible
//     for opening the file, but the email sender will handle closing it.
//
// Returns a configured File instance that can be added to a Mail object.
//
// Example:
//
//	file, err := os.Open("photo.jpg")
//	if err != nil {
//	    return err
//	}
//	attachment := sender.NewFile("photo.jpg", "image/jpeg", file)
//	mail.AddAttachment("photo.jpg", "image/jpeg", file, false)
//
// Common MIME types:
//   - "text/plain" - Plain text files
//   - "text/html" - HTML files
//   - "application/pdf" - PDF documents
//   - "image/jpeg", "image/png", "image/gif" - Images
//   - "application/zip" - ZIP archives
//   - "application/octet-stream" - Generic binary data
func NewFile(name string, mime string, data io.ReadCloser) File {
	return File{
		name: name,
		mime: mime,
		data: data,
	}
}
