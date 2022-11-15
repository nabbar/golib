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
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	pathSeparatorOs     = string(os.PathSeparator)
	pathSeparatorCommon = "/"
	pathParentCommon    = ".."
)

type File struct {
	Name string
	Path string
}

func (a File) Matching(containt string) bool {
	if len(containt) < 1 {
		return false
	}
	return strings.Contains(strings.ToLower(a.Name), strings.ToLower(containt))
}

func (a File) Regex(regex string) bool {
	if len(regex) < 1 {
		return false
	}

	b, _ := regexp.MatchString(regex, a.Name)

	return b
}

func (a File) MatchingFullPath(containt string) bool {
	if len(containt) < 1 {
		return false
	}
	return strings.Contains(strings.ToLower(a.GetKeyMap()), strings.ToLower(containt))
}

func (a File) RegexFullPath(regex string) bool {
	if len(regex) < 1 {
		return false
	}

	b, _ := regexp.MatchString(regex, a.GetKeyMap())

	return b
}

func (a File) GetKeyMap() string {
	return filepath.Join(a.Path, a.Name)
}

func (a File) GetDestFileOnly(baseDestination string) string {
	return filepath.Join(baseDestination, a.Name)
}

func (a File) GetDestWithPath(baseDestination string) string {
	return filepath.Join(baseDestination, a.Path, a.Name)
}

func NewFile(name, path string) File {
	return File{
		Name: name,
		Path: path,
	}
}

func NewFileFullPath(fullpath string) File {
	return NewFile(filepath.Base(fullpath), filepath.Dir(fullpath))
}

func CleanPath(p string) string {
	for {
		if strings.HasPrefix(p, pathParentCommon) {
			p = strings.TrimPrefix(p, pathParentCommon)
		} else if strings.HasPrefix(p, pathSeparatorCommon+pathParentCommon) {
			p = strings.TrimPrefix(p, pathSeparatorCommon+pathParentCommon)
		} else if strings.HasPrefix(p, pathSeparatorOs+pathParentCommon) {
			p = strings.TrimPrefix(p, pathSeparatorOs+pathParentCommon)
		} else {
			return filepath.Clean(p)
		}
	}
}
