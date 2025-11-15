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

// PathCheckCreate ensures a file or directory exists at the given path with the correct permissions.
//
// This function performs the following operations:
//   - Checks if the path exists
//   - Creates the path if it doesn't exist (file or directory based on isFile parameter)
//   - Creates parent directories if needed
//   - Validates that existing paths match the expected type (file vs directory)
//   - Updates permissions if they don't match the specified values
//
// Parameters:
//   - isFile: true to ensure path is a file, false for a directory
//   - path: the filesystem path to check/create
//   - permFile: permissions to apply to files (e.g., 0644)
//   - permDir: permissions to apply to directories (e.g., 0755)
//
// Returns:
//   - error: nil on success, or an error if:
//   - The path exists but is the wrong type (file when directory expected, or vice versa)
//   - Parent directory creation fails
//   - Permission updates fail
//   - File/directory creation fails
//
// Behavior:
//   - If path exists as expected type: validates and updates permissions if needed
//   - If path doesn't exist: creates it with specified permissions
//   - If parent directories don't exist: creates them with permDir permissions
//   - If path exists as wrong type: returns an error without modification
//
// Example:
//
//	// Ensure config directory exists with 0755 permissions
//	err := PathCheckCreate(false, "/etc/app/config", 0644, 0755)
//
//	// Ensure log file exists with 0644 permissions
//	err := PathCheckCreate(true, "/var/log/app.log", 0644, 0755)
//
// Thread Safety:
//
//	This function is safe to call concurrently for different paths, but
//	concurrent calls for the same path may result in race conditions.
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
	}

	rt, e := os.OpenRoot(filepath.Dir(path))

	defer func() {
		if rt != nil {
			_ = rt.Close()
		}
	}()

	if e != nil {
		return e
	}

	hf, e := rt.Create(filepath.Base(path))

	defer func() {
		if hf != nil {
			_ = hf.Close()
		}
	}()

	if e != nil {
		return e
	}

	_ = hf.Close()
	hf = nil

	_ = rt.Chmod(filepath.Base(path), permFile)

	return nil
}
