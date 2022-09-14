/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package zip

import (
	"fmt"

	arcmod "github.com/nabbar/golib/archive/archive"
	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorParamEmpty liberr.CodeError = iota + arcmod.MinPkgArchiveZip
	ErrorFileOpen
	ErrorFileClose
	ErrorFileSeek
	ErrorFileStat
	ErrorIOCopy
	ErrorZipOpen
	ErrorZipCreate
	ErrorZipComment
	ErrorZipAddFile
	ErrorZipFileOpen
	ErrorZipFileClose
	ErrorDirCreate
	ErrorDestinationStat
	ErrorDestinationIsDir
	ErrorDestinationIsNotDir
	ErrorDestinationRemove
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision golib/archive/zip"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorFileOpen:
		return "cannot open zipped file"
	case ErrorFileClose:
		return "closing file occurs error"
	case ErrorFileStat:
		return "getting file stat occurs error"
	case ErrorFileSeek:
		return "cannot seek into file"
	case ErrorIOCopy:
		return "io copy occurs error"
	case ErrorZipOpen:
		return "cannot open zip file"
	case ErrorZipCreate:
		return "cannot create zip file"
	case ErrorZipComment:
		return "cannot set comment to zip file"
	case ErrorZipAddFile:
		return "cannot add file to zip file"
	case ErrorZipFileOpen:
		return "cannot open file into zip file"
	case ErrorZipFileClose:
		return "cannot flose file into zip file"
	case ErrorDirCreate:
		return "make directory occurs error"
	case ErrorDestinationStat:
		return "cannot stat destination"
	case ErrorDestinationIsDir:
		return "cannot create destination not directory over an existing directory"
	case ErrorDestinationIsNotDir:
		return "cannot create destination directory over an existing non directory"
	case ErrorDestinationRemove:
		return "cannot remove destination "
	}

	return liberr.NullMessage
}
