/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package progress_test

import (
	"os"

	"github.com/nabbar/golib/file/progress"
)

// createTestFile creates a test file with the given content and returns the path.
func createTestFile(content []byte) (string, error) {
	tmp, err := os.CreateTemp("", "progress-test-*.txt")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := tmp.Write(content); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

// cleanup removes the test file.
func cleanup(path string) {
	if path != "" {
		os.Remove(path)
	}
}

// createProgressFile creates a Progress instance with test data.
func createProgressFile(content []byte) (progress.Progress, string, error) {
	path, err := createTestFile(content)
	if err != nil {
		return nil, "", err
	}

	p, err := progress.Open(path)
	if err != nil {
		cleanup(path)
		return nil, "", err
	}

	return p, path, nil
}

// createProgressFileRW creates a Progress instance for read/write.
func createProgressFileRW(content []byte) (progress.Progress, string, error) {
	path, err := createTestFile(content)
	if err != nil {
		return nil, "", err
	}

	p, err := progress.New(path, os.O_RDWR, 0644)
	if err != nil {
		cleanup(path)
		return nil, "", err
	}

	return p, path, nil
}
