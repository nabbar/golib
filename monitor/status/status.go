/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package status

import (
	"strconv"
	"strings"
)

func (s Status) String() string {
	switch s {
	case OK:
		return "OK"
	case Warn:
		return "Warn"
	default:
		return "KO"
	}
}

func (s Status) Int() int64 {
	return int64(s)
}

func (s Status) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(s.String())+2)
	b = append(b, '"')
	b = append(b, []byte(s.String())...)
	b = append(b, '"')
	return b, nil
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var (
		e   error
		i   int64
		a   Status
		str string
	)

	str = string(data)

	if str == "null" {
		*s = KO
		return nil
	}

	if strings.HasPrefix(str, "\"") || strings.HasSuffix(str, "\"") {
		if str, e = strconv.Unquote(str); e != nil {
			return e
		}
	}

	if i, e = strconv.ParseInt(str, 10, 8); e != nil {
		*s = NewFromString(str)
		return nil
	} else if a = NewFromInt(i); a != KO {
		*s = a
		return nil
	} else {
		*s = NewFromString(str)
		return nil
	}
}
