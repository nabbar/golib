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

package tar

import (
	"fmt"

	arcmod "github.com/nabbar/golib/archive/archive"
	liberr "github.com/nabbar/golib/errors"
)

const pkgName = "golib/archive/tar"

const (
	ErrorParamEmpty liberr.CodeError = iota + arcmod.MinPkgArchiveTar
	ErrorTarNext
	ErrorFileOpen
	ErrorFileSeek
	ErrorFileClose
	ErrorIOCopy
	ErrorDirCreate
	ErrorLinkCreate
	ErrorSymLinkCreate
	ErrorDestinationStat
	ErrorDestinationIsDir
	ErrorDestinationIsNotDir
	ErrorDestinationRemove
	ErrorTarCreate
	ErrorGzipCreate
	ErrorTarCreateAddFile
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision %s", pkgName))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case liberr.UnknownError:
		return liberr.NullMessage
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorTarNext:
		return "cannot get next tar file"
	case ErrorFileSeek:
		return "cannot seek into file"
	case ErrorFileOpen:
		return "cannot open file"
	case ErrorFileClose:
		return "closing file occurs error"
	case ErrorIOCopy:
		return "io copy occurs error"
	case ErrorDirCreate:
		return "make directory occurs error"
	case ErrorLinkCreate:
		return "creation of symlink occurs error"
	case ErrorSymLinkCreate:
		return "creation of symlink occurs error"
	case ErrorDestinationStat:
		return "cannot stat destination"
	case ErrorDestinationIsDir:
		return "cannot create destination not directory over an existing directory"
	case ErrorDestinationIsNotDir:
		return "cannot create destination directory over an existing non directory"
	case ErrorDestinationRemove:
		return "cannot remove destination "
	case ErrorTarCreate:
		return "cannot create Tar archive"
	case ErrorGzipCreate:
		return "cannot create Gzip compression for the Tar archive"
	case ErrorTarCreateAddFile:
		return "cannot add file content to tar archive"
	}

	return liberr.NullMessage
}
