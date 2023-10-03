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

package archive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	libbz2 "github.com/nabbar/golib/archive/bz2"
	libgzp "github.com/nabbar/golib/archive/gzip"
	libtar "github.com/nabbar/golib/archive/tar"
	libzip "github.com/nabbar/golib/archive/zip"
	liberr "github.com/nabbar/golib/errors"
	libfpg "github.com/nabbar/golib/file/progress"
)

type ArchiveType uint8

const (
	permDir  = os.FileMode(0775)
	permFile = os.FileMode(0664)
)

const (
	TypeTar = iota + 1
	TypeTarGzip
	TypeGzip
	TypeZip
)

func ExtractFile(src, dst libfpg.Progress, fileNameContain, fileNameRegex string) liberr.Error {
	var (
		e error

		tmp libfpg.Progress
		err liberr.Error
	)

	defer func() {
		if tmp != nil {
			_ = tmp.CloseDelete()
		}
	}()

	if tmp, e = libfpg.Temp(""); e != nil {
		return ErrorFileOpen.Error(e)
	} else {
		dst.SetRegisterProgress(tmp)
	}

	if _, e = src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
		// #nosec
	}

	if err = libbz2.GetFile(src, tmp); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if err.IsCode(libbz2.ErrorIOCopy) {
		return err
	}

	if err = libgzp.GetFile(src, tmp); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCode(libgzp.ErrorGZReader) {
		return err
	}

	if err = libtar.GetFile(src, tmp, fileNameContain, fileNameRegex); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCode(libtar.ErrorTarNext) {
		return err
	}

	if err = libzip.GetFile(src, tmp, fileNameContain, fileNameRegex); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCode(libzip.ErrorZipOpen) {
		return err
	}

	if _, e = dst.ReadFrom(src); e != nil {
		//logger.ErrorLevel.LogErrorCtx(logger.DebugLevel, "reopening file", err)
		return ErrorIOCopy.Error(e)
	}

	return nil
}

func ExtractAll(src libfpg.Progress, originalName, outputPath string, defaultDirPerm os.FileMode) liberr.Error {
	var (
		e error
		i os.FileInfo

		tmp libfpg.Progress
		dst libfpg.Progress
		err liberr.Error
	)

	defer func() {
		if src != nil {
			_ = src.Close()
		}
		if dst != nil {
			_ = dst.Close()
		}
		if tmp != nil {
			_ = tmp.CloseDelete()
		}
	}()

	if tmp, e = libfpg.Temp(""); e != nil {
		return ErrorFileOpen.Error(e)
	} else {
		src.SetRegisterProgress(tmp)
	}

	if err = libbz2.GetFile(src, tmp); err == nil {
		if inf, er := tmp.Stat(); er == nil {
			tmp.Reset(inf.Size())
		}
		return ExtractAll(tmp, originalName, outputPath, defaultDirPerm)
	} else if !err.IsCode(libbz2.ErrorIOCopy) {
		return err
	}

	if err = libgzp.GetFile(src, tmp); err == nil {
		if inf, er := tmp.Stat(); er == nil {
			tmp.Reset(inf.Size())
		}
		return ExtractAll(tmp, originalName, outputPath, defaultDirPerm)
	} else if !err.IsCode(libgzp.ErrorGZReader) {
		return err
	}

	if tmp != nil {
		_ = tmp.CloseDelete()
	}

	if i, e = os.Stat(outputPath); e != nil && os.IsNotExist(e) {
		//nolint #nosec
		/* #nosec */
		if e = os.MkdirAll(outputPath, permDir); e != nil {
			return ErrorDirCreate.Error(e)
		}
	} else if e != nil {
		return ErrorDirStat.Error(e)
	} else if !i.IsDir() {
		return ErrorDirNotDir.Error(nil)
	}

	if err = libtar.GetAll(src, outputPath, defaultDirPerm); err == nil {
		return nil
	} else if !err.IsCode(libtar.ErrorTarNext) {
		return err
	}

	if err = libzip.GetAll(src, outputPath, defaultDirPerm); err == nil {
		return nil
	} else if !err.IsCode(libzip.ErrorZipOpen) {
		return err
	}

	if dst, e = libfpg.New(filepath.Join(outputPath, originalName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, permFile); e != nil {
		return ErrorFileOpen.Error(e)
	} else {
		src.SetRegisterProgress(dst)
	}

	if _, e = src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	} else if _, e = dst.ReadFrom(src); e != nil {
		return ErrorIOCopy.Error(e)
	}

	return nil
}

func CreateArchive(archiveType ArchiveType, archive libfpg.Progress, stripPath string, comment string, pathContent ...string) (created bool, err liberr.Error) {
	if len(pathContent) < 1 {
		//nolint #goerr113
		return false, ErrorParamEmpty.Error(fmt.Errorf("pathContent is empty"))
	}

	switch archiveType {
	case TypeGzip:
		return libgzp.Create(archive, stripPath, comment, pathContent...)
	case TypeTar:
		return libtar.Create(archive, stripPath, comment, pathContent...)
	case TypeTarGzip:
		return libtar.CreateGzip(archive, stripPath, comment, pathContent...)
	case TypeZip:
		return libzip.Create(archive, stripPath, comment, pathContent...)

		//@TODO: add zip mode
	}

	return false, nil
}
