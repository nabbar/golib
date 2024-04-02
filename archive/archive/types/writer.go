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

package types

import (
	"io"
	"io/fs"
)

type ReplaceName func(string) string

type Writer interface {
	io.Closer

	// Add will add the given file into the archive.
	//
	// Parameter(s):
	//   - fs.FileInfo: the file information for the given path (permission, size, etc...).
	//   - io.ReadCloser: the read/close stream to read the data and store it into the archive.
	//   - string: use to force a different pathname for the embedded file into the archive.
	//   - string: use to specify the target if the embedded file is a link.
	// Return type: error
	Add(fs.FileInfo, io.ReadCloser, string, string) error

	// FromPath will parse recursively the given path and add it into the archive
	//
	// Parameter(s):
	//   - string: the source path to parse recursively and add into the archive.
	//   - string: a filtering string to accept only certain files (empty do disable filtering).
	//   - ReplaceName: a function to replace the name of the embedded file, if needed.
	// Returns error if triggered
	FromPath(string, string, ReplaceName) error
}
