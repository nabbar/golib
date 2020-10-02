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
	"io"
	"os"
	"path"

	"github.com/nabbar/golib/archive/bz2"
	"github.com/nabbar/golib/archive/gzip"
	"github.com/nabbar/golib/archive/tar"
	"github.com/nabbar/golib/archive/zip"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/ioutils"
	"github.com/nabbar/golib/logger"
)

type ArchiveType uint8

func ExtractFile(src, dst ioutils.FileProgress, fileNameContain, fileNameRegex string) errors.Error {
	var (
		tmp ioutils.FileProgress
		err errors.Error
	)

	if tmp, err = dst.NewFileTemp(); err != nil {
		return err
	}

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.ErrorParent(e)
		// #nosec
	}

	if err := bz2.GetFile(src, tmp); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if err.IsCodeError(bz2.ErrorIOCopy) {
		return err
	}

	if err := gzip.GetFile(src, tmp); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCodeError(gzip.ErrorGZReader) {
		return err
	}

	if err := tar.GetFile(src, tmp, fileNameContain, fileNameRegex); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCodeError(tar.ErrorTarNext) {
		return err
	}

	if err := zip.GetFile(src, tmp, fileNameContain, fileNameRegex); err == nil {
		//logger.DebugLevel.Log("try another archive...")
		return ExtractFile(tmp, dst, fileNameContain, fileNameRegex)
	} else if !err.IsCodeError(zip.ErrorZipOpen) {
		return err
	}

	_ = tmp.Close()

	if _, e := dst.ReadFrom(src); e != nil {
		//logger.ErrorLevel.LogErrorCtx(logger.DebugLevel, "reopening file", err)
		return ErrorIOCopy.ErrorParent(e)
	}

	return nil
}

func ExtractAll(src ioutils.FileProgress, originalName, outputPath string, defaultDirPerm os.FileMode) errors.Error {
	var (
		tmp ioutils.FileProgress
		dst ioutils.FileProgress
		err errors.Error
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

	logger.DebugLevel.Log("try BZ2...")
	if err = bz2.GetFile(src, tmp); err == nil {
		logger.DebugLevel.Log("try another archive...")
		return ExtractAll(tmp, originalName, outputPath, defaultDirPerm)
	} else if !err.IsCodeError(bz2.ErrorIOCopy) {
		logger.DebugLevel.Logf("error found on BZ2 : %v", err)
		return err
	} else {
		logger.DebugLevel.Logf("not a BZ2 : %v", err)
	}

	logger.DebugLevel.Log("try GZIP...")
	if err = gzip.GetFile(src, tmp); err == nil {
		logger.DebugLevel.Log("try another archive...")
		return ExtractAll(tmp, originalName, outputPath, defaultDirPerm)
	} else if !err.IsCodeError(gzip.ErrorGZReader) {
		logger.DebugLevel.Logf("error found on GZIP : %v", err)
		return err
	} else {
		logger.DebugLevel.Logf("not a GZIP : %v", err)
	}

	if tmp != nil {
		_ = tmp.Close()
	}

	logger.DebugLevel.Log("prepare output...")
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

	logger.DebugLevel.Log("try tar...")
	if err = tar.GetAll(src, outputPath, defaultDirPerm); err == nil {
		logger.DebugLevel.Log("extracting TAR finished...")
		return nil
	} else if !err.IsCodeError(tar.ErrorTarNext) {
		logger.DebugLevel.Logf("error found on TAR : %v", err)
		return err
	} else {
		logger.DebugLevel.Logf("not a TAR : %v", err)
	}

	logger.DebugLevel.Log("try zip...")
	if err = zip.GetAll(src, outputPath, defaultDirPerm); err == nil {
		logger.DebugLevel.Log("extracting ZIP finished...")
		return nil
	} else if !err.IsCodeError(zip.ErrorZipOpen) {
		logger.DebugLevel.Logf("error found on ZIP : %v", err)
		return err
	} else {
		logger.DebugLevel.Logf("not a ZIP : %v", err)
	}

	logger.DebugLevel.Log("writing original file...")
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
