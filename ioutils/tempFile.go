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

package ioutils

import (
	"os"
	"path/filepath"

	. "github.com/nabbar/golib/errors"
)

func NewTempFile() (*os.File, Error) {
	f, e := os.CreateTemp(os.TempDir(), "")
	return f, ErrorIOFileTempNew.Iferror(e)
}

func GetTempFilePath(f *os.File) string {
	if f == nil {
		return ""
	}

	return filepath.Join(os.TempDir(), filepath.Base(f.Name()))
}

func DelTempFile(f *os.File) Error {
	if f == nil {
		return nil
	}

	n := GetTempFilePath(f)

	a := f.Close()
	e1 := ErrorIOFileTempClose.Iferror(a)

	b := os.Remove(n)
	e2 := ErrorIOFileTempRemove.Iferror(b)

	return MakeErrorIfError(e2, e1)
}
