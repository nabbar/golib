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

package tlsmode

import (
	"math"
	"strconv"
	"strings"
)

type TLSMode uint8

const (
	TLSNone TLSMode = iota
	TLSStartTLS
	TLSStrictTLS
)

func TLSModeFromString(str string) TLSMode {
	switch strings.ToLower(str) {
	case TLSStrictTLS.String():
		return TLSStrictTLS
	case TLSStartTLS.String():
		return TLSStartTLS
	}

	return TLSNone
}

func TLSModeFromInt(i int64) TLSMode {
	if i > math.MaxUint8 {
		return TLSNone
	} else if i < 0 {
		return TLSNone
	}

	switch TLSMode(i) {
	case TLSStrictTLS:
		return TLSStrictTLS
	case TLSStartTLS:
		return TLSStartTLS
	default:
		return TLSNone
	}
}

func (tlm TLSMode) String() string {
	switch tlm {
	case TLSStrictTLS:
		return "tls"
	case TLSStartTLS:
		return "starttls"
	case TLSNone:
		return "none"
	}

	return TLSNone.String()
}

func (tlm TLSMode) Int() int64 {
	return int64(tlm)
}

func (tlm TLSMode) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(tlm.String())+2)
	b = append(b, '"')
	b = append(b, []byte(tlm.String())...)
	b = append(b, '"')
	return b, nil
}

func (tlm *TLSMode) UnmarshalJSON(data []byte) error {
	var (
		e   error
		i   int64
		a   TLSMode
		str string
	)

	str = string(data)

	if str == "null" {
		*tlm = TLSNone
		return nil
	}

	if strings.HasPrefix(str, "\"") || strings.HasSuffix(str, "\"") {
		if str, e = strconv.Unquote(str); e != nil {
			return e
		}
	}

	if i, e = strconv.ParseInt(str, 10, 8); e != nil {
		*tlm = TLSModeFromString(str)
		return nil
	} else if a = TLSModeFromInt(i); a != TLSNone {
		*tlm = a
		return nil
	} else {
		*tlm = TLSModeFromString(str)
		return nil
	}
}
