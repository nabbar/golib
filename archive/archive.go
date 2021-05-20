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
	"path"

	libbz2 "github.com/nabbar/golib/archive/bz2"
	libgzp "github.com/nabbar/golib/archive/gzip"
	libtar "github.com/nabbar/golib/archive/tar"
	libzip "github.com/nabbar/golib/archive/zip"
	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
	liblog "github.com/nabbar/golib/logger"
)

type ArchiveType uint8

const (
	TypeTar = iota + 1
	TypeTarGzip
	TypeGzip
)

func ExtractFile(src, dst libiot.FileProgress, fileNameContain, fileNameRegex string) liberr.Error {
	var (
		tmp libiot.FileProgress
		err liberr.Error
	)

	if tmp, err = dst.NewFileTemp(); err != nil {
		return err
	}

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.ErrorParent(e)
		// #nosec
	}

	if err := libbz2.GetFile(src, tmp); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if err.IsCodeError(libbz2.ErrorIOCopy) {
		return err
	}

	if err := libgzp.GetFile(src, tmp); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCodeError(libgzp.ErrorGZReader) {
		return err
	}

	if err := libtar.GetFile(src, tmp, fileNameContain, fileNameRegex); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCodeError(libtar.ErrorTarNext) {
		return err
	}

	if err := libzip.GetFile(src, tmp, fileNameContain, fileNameRegex); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCodeError(libzip.ErrorZipOpen) {
		return err
	}

	_ = tmp.Close()

	if _, e := dst.ReadFrom(src); e != nil {
		//logger.ErrorLevel.LogErrorCtx(logger.DebugLevel, "reopening file", err)
		return ErrorIOCopy.ErrorParent(e)
	}

	return nil
}

func ExtractAll(src libiot.FileProgress, originalName, outputPath string, defaultDirPerm os.FileMode) liberr.Error {
	var (
		tmp libiot.FileProgress
		dst libiot.FileProgress
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
			_ = tmp.Close()
		}
	}()

	if tmp, err = src.NewFileTemp(); err != nil {
		return ErrorFileOpen.Error(err)
	}

	liblog.DebugLevel.Log("try BZ2...")
	if err = libbz2.GetFile(src, tmp); err == nil {
		liblog.DebugLevel.Log("try another archive...")
		return ExtractAll(tmp, originalName, outputPath, defaultDirPerm)
	} else if !err.IsCodeError(libbz2.ErrorIOCopy) {
		liblog.DebugLevel.Logf("error found on BZ2 : %v", err)
		return err
	} else {
		liblog.DebugLevel.Logf("not a BZ2 : %v", err)
	}

	liblog.DebugLevel.Log("try GZIP...")
	if err = libgzp.GetFile(src, tmp); err == nil {
		liblog.DebugLevel.Log("try another archive...")
		return ExtractAll(tmp, originalName, outputPath, defaultDirPerm)
	} else if !err.IsCodeError(libgzp.ErrorGZReader) {
		liblog.DebugLevel.Logf("error found on GZIP : %v", err)
		return err
	} else {
		liblog.DebugLevel.Logf("not a GZIP : %v", err)
	}

	if tmp != nil {
		_ = tmp.Close()
	}

	liblog.DebugLevel.Log("prepare output...")
	if i, e := os.Stat(outputPath); e != nil && os.IsNotExist(e) {
		//nolint #nosec
		/* #nosec */
		if e := os.MkdirAll(outputPath, 0775); e != nil {
			return ErrorDirCreate.ErrorParent(e)
		}
	} else if e != nil {
		return ErrorDirStat.ErrorParent(e)
	} else if !i.IsDir() {
		return ErrorDirNotDir.Error(nil)
	}

	liblog.DebugLevel.Log("try tar...")
	if err = libtar.GetAll(src, outputPath, defaultDirPerm); err == nil {
		liblog.DebugLevel.Log("extracting TAR finished...")
		return nil
	} else if !err.IsCodeError(libtar.ErrorTarNext) {
		liblog.DebugLevel.Logf("error found on TAR : %v", err)
		return err
	} else {
		liblog.DebugLevel.Logf("not a TAR : %v", err)
	}

	liblog.DebugLevel.Log("try zip...")
	if err = libzip.GetAll(src, outputPath, defaultDirPerm); err == nil {
		liblog.DebugLevel.Log("extracting ZIP finished...")
		return nil
	} else if !err.IsCodeError(libzip.ErrorZipOpen) {
		liblog.DebugLevel.Logf("error found on ZIP : %v", err)
		return err
	} else {
		liblog.DebugLevel.Logf("not a ZIP : %v", err)
	}

	liblog.DebugLevel.Log("writing original file...")
	if dst, err = src.NewFilePathWrite(path.Join(outputPath, originalName), true, true, 0664); err != nil {
		return ErrorFileOpen.Error(err)
	}

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.ErrorParent(e)
	} else if _, e := dst.ReadFrom(src); e != nil {
		return ErrorIOCopy.ErrorParent(e)
	}

	return nil
}

func CreateArchive(archiveType ArchiveType, archive libiot.FileProgress, stripPath string, pathContent ...string) (created bool, err liberr.Error) {
	if len(pathContent) < 1 {
		//nolint #goerr113
		return false, ErrorParamsEmpty.ErrorParent(fmt.Errorf("pathContent is empty"))
	}

	switch archiveType {
	case TypeGzip:
		return libgzp.Create(archive, stripPath, pathContent...)
	case TypeTar:
		return libtar.Create(archive, stripPath, pathContent...)
	case TypeTarGzip:
		return libtar.CreateGzip(archive, stripPath, pathContent...)

		//@TODO: add zip mode
	}

	return false, nil
}
