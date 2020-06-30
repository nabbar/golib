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

package njs_ioutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func NewTempFile() (*os.File, error) {
	if f, e := ioutil.TempFile(os.TempDir(), ""); e != nil {
		return nil, e
	} else {
		return f, nil
	}
}

func GetTempFilePath(f *os.File) string {
	if f == nil {
		return ""
	}

	return path.Join(os.TempDir(), path.Base(f.Name()))
}

func DelTempFile(f *os.File, ignoreErrClose bool) error {
	if f == nil {
		return nil
	}

	n := GetTempFilePath(f)
	a := f.Close()
	b := os.Remove(n)

	if ignoreErrClose {
		return b
	}

	if a != nil && b != nil {
		return fmt.Errorf("%v, %v", a, b)
	} else if a != nil {
		return a
	}

	return b
}
