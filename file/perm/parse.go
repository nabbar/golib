/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
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

package perm

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func parseString(s string) (Perm, error) {
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "'", "", -1)

	if v, e := strconv.ParseUint(s, 8, 32); e != nil {
		return 0, e
	} else if v > math.MaxUint32 {
		return Perm(0), fmt.Errorf("invalid permission")
	} else {
		return Perm(v), nil
	}
}

func (p *Perm) parseString(s string) error {
	if v, e := parseString(s); e != nil {
		return e
	} else {
		*p = v
		return nil
	}
}

func (p *Perm) unmarshall(val []byte) error {
	if tmp, err := ParseByte(val); err != nil {
		return err
	} else {
		*p = tmp
		return nil
	}
}
