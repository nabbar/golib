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
	"bytes"
	"encoding/json"
	"strings"
)

func (a Algorithm) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Algorithm) UnmarshalText(b []byte) error {
	*a = None

	s := strings.Trim(string(b), "\"")
	s = strings.Trim(s, "'")
	s = strings.TrimSpace(s)

	switch {
	case strings.EqualFold(s, Tar.String()):
		*a = Tar
	case strings.EqualFold(s, Zip.String()):
		*a = Zip
	default:
		*a = None
	}

	return nil
}

func (a Algorithm) MarshalJSON() ([]byte, error) {
	if a.IsNone() {
		return []byte("null"), nil
	}
	return append(append([]byte{'"'}, []byte(a.String())...), '"'), nil
}

func (a *Algorithm) UnmarshalJSON(b []byte) error {
	var s string

	if n := []byte("null"); bytes.Equal(b, n) {
		*a = None
		return nil
	} else if err := json.Unmarshal(b, &s); err != nil {
		return err
	} else {
		return a.UnmarshalText([]byte(s))
	}
}
