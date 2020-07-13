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
	"os"

	njs_archive "github.com/nabbar/golib/njs-archive"

	iou "github.com/nabbar/golib/njs-ioutils"
)

// git archive --format=tar --output=git.tar HEAD
//const fileName = "./git.tar"
const fileName = "./vendor.zip"

//const contain = "njs-version/license_mit.go"
const contain = "vendor/github.com/gin-gonic/gin/internal/json/json.go"

//const regex = ""
const regex = "vendor\\.tar(\\.(?:gz|bz))?"

func main() {
	var (
		src *os.File
		tmp *os.File
		rio *os.File
		err error
	)

	if src, err = os.Open(fileName); err != nil {
		panic(err)
	}

	defer func() {
		_ = src.Close()
	}()

	if tmp, err = iou.NewTempFile(); err != nil {
		panic(err)
	}

	defer func() {
		_ = iou.DelTempFile(tmp)
	}()

	if _, err = io.Copy(tmp, src); err != nil {
		panic(err)
	}

	if rio, err = njs_archive.ExtractFile(tmp, contain, regex); err != nil {
		panic(err)
	}

	defer func() {
		_ = rio.Close()
	}()

	if _, err = rio.Seek(0, 0); err != nil {
		panic(err)
	}

	if b, e := ioutil.ReadAll(rio); e != nil {
		panic(e)
	} else {
		println(string(b))
	}
}
