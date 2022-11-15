//+build examples
//go:build examples
// +build examples

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
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/nabbar/golib/archive"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils"
	"github.com/nabbar/golib/logger"
)

// git archive --format=tar --output=git.tar HEAD && gzip git.tar
const fileName = "./git.tar.gz"

func main() {
	var (
		src ioutils.FileProgress
		tmp ioutils.FileProgress
		out string
		err error
	)

	defer func() {
		if src != nil {
			_ = src.Close()
		}
		if tmp != nil {
			_ = tmp.Close()
		}
	}()

	logger.SetLevel(logger.DebugLevel)
	logger.EnableColor()
	logger.FileTrace(true)
	errors.SetModeReturnError(errors.ErrorReturnCodeErrorTraceFull)

	if src, err = ioutils.NewFileProgressPathOpen(fileName); err != nil {
		panic(err)
	}

	if tmp, err = ioutils.NewFileProgressTemp(); err != nil {
		panic(err)
	} else {
		out = tmp.FilePath()
		_ = tmp.Close()
	}

	if err = archive.ExtractAll(src, path.Base(src.FilePath()), out, 0775); err != nil {
		panic(err)
	}

	if list, e := ioutil.ReadDir(out); e != nil {
		panic(e)
	} else {
		for _, f := range list {
			var (
				isDir  bool
				isLink bool
				isFile bool
			)
			isDir = f.IsDir()
			isLink = f.Mode()&os.ModeSymlink == os.ModeSymlink
			isFile = !isLink && !isDir
			println(fmt.Sprintf("Item '%s' is Dir '%v', is Link '%v', is File '%v'", f.Name(), isDir, isLink, isFile))
		}
	}
}
