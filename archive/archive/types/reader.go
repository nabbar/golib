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

type FuncExtract func(fs.FileInfo, io.ReadCloser, string, string) bool

type Reader interface {
	io.Closer

	// List try to return a slice of all path into the archive or an error.
	List() ([]string, error)
	// Info returns the file information as fs.FileInfo for the given path or an error.
	//
	// Parameters:
	// - string: the path of the embedded file into the archive to get information for.
	//
	// Returns:
	// - fs.FileInfo: the file information for the given path.
	// - error: an error if the file information could not be retrieved.
	Info(string) (fs.FileInfo, error)
	// Get retrieves an io.ReadCloser or an error based on the provided file path.
	//
	// Parameters:
	// - string: the path of the embedded file into the archive to get information for.
	//
	// Returns:
	// - io.ReadCloser: the read/close stream for the given path.
	// - error: an error if the file information could not be retrieved.
	Get(string) (io.ReadCloser, error)
	// Has will check if the archive contains the given path.
	//
	// Parameters:
	// - string: the path of the embedded file into the archive to get information for.
	//
	// Returns:
	// - bool: true if the archive contains the given path.
	Has(string) bool
	// Walk applies the given function to each element in the archive.
	//
	// Parameters:
	// - FuncExtract: the function will be call on each item in the archive.
	// The function can return false to stop or true to continue the walk.
	// The function will accept the following parameters:
	// - fs.FileInfo: the file information for the given path.
	// - io.ReadCloser: the read/close stream for the given path.
	// - string: the path of the embedded file into the archive.
	// - string: the link target of the embedded file if it is a link or a symlink.
	Walk(FuncExtract)
}
