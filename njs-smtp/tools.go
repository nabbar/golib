/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package njs_smtp

import "strings"

func cleanMergeSlice(str []string, args ...string) []string {
	for _, s := range args {
		if s != "" {
			str = append(str, s)
		}
	}

	return str
}

func unicSliceString(str []string) []string {
	var new = make([]string, 0)

	for _, s := range str {
		if !existSliceString(new, s) {
			new = append(new, s)
		}
	}

	return new
}

func existSliceString(slc []string, str string) bool {
	for _, s := range slc {
		if s == str {
			return true
		}
	}

	return false
}

func cleanJoin(str []string, glue string) string {
	var new = make([]string, 0)

	for _, s := range str {
		if s != "" {
			new = append(new, s)
		}
	}

	return strings.Join(new, glue)
}
