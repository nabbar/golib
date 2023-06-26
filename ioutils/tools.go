/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package ioutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func PathCheckCreate(isFile bool, path string, permFile os.FileMode, permDir os.FileMode) error {
	if inf, err := os.Stat(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	} else if err == nil && inf.IsDir() {
		if isFile {
			return fmt.Errorf("path '%s' still exising but is a directory", path)
		}
		if inf.Mode() != permDir {
			_ = os.Chmod(path, permDir)
		}
		return nil
	} else if err == nil && !inf.IsDir() {
		if !isFile {
			return fmt.Errorf("path '%s' still exising but is not a directory", path)
		}
		if inf.Mode() != permFile {
			_ = os.Chmod(path, permFile)
		}
		return nil
	} else if !isFile {
		return os.MkdirAll(path, permDir)
	} else if err = PathCheckCreate(false, filepath.Dir(path), permFile, permDir); err != nil {
		return err
	} else if hf, e := os.Create(path); e != nil {
		return e
	} else {
		_ = hf.Close()
	}

	_ = os.Chmod(path, permFile)

	return nil
}
