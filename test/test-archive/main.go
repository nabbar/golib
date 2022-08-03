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

package main

import (
	"io"
	"io/ioutil"

	"github.com/nabbar/golib/errors"

	"github.com/nabbar/golib/archive"
	"github.com/nabbar/golib/ioutils"
)

// git archive --format=tar --output=git.tar HEAD
// const fileName = "./git.tar"
const fileName = "./vendor.zip"

// const contain = "version/license_mit.go"
const contain = "vendor/github.com/gin-gonic/gin/internal/json/json.go"

// const regex = ""
const regex = "vendor\\.tar(\\.(?:gz|bz))?"

func main() {
	var (
		src ioutils.FileProgress
		tmp ioutils.FileProgress
		rio ioutils.FileProgress
		err errors.Error
	)

	defer func() {
		if src != nil {
			_ = src.Close()
		}
		if tmp != nil {
			_ = tmp.Close()
		}
		if rio != nil {
			_ = rio.Close()
		}
	}()

	if src, err = ioutils.NewFileProgressPathOpen(fileName); err != nil {
		panic(err)
	}

	if tmp, err = ioutils.NewFileProgressTemp(); err != nil {
		panic(err)
	}

	if rio, err = ioutils.NewFileProgressTemp(); err != nil {
		panic(err)
	}

	if _, e := tmp.ReadFrom(src); e != nil {
		panic(e)
	}

	if err = archive.ExtractFile(tmp, rio, contain, regex); err != nil {
		panic(err)
	}

	if _, e := rio.Seek(0, io.SeekStart); e != nil {
		panic(e)
	}

	if b, e := ioutil.ReadAll(rio); e != nil {
		panic(e)
	} else {
		println(string(b))
	}
}
