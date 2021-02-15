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

package mailer

import (
	"strings"

	"github.com/matcornic/hermes/v2"
)

type TextDirection uint8

const (
	LeftToRight TextDirection = iota
	RightToLeft
)

func (d TextDirection) getDirection() hermes.TextDirection {
	switch d {
	case LeftToRight:
		return hermes.TDLeftToRight
	case RightToLeft:
		return hermes.TDRightToLeft
	}

	return LeftToRight.getDirection()
}

func (d TextDirection) String() string {
	switch d {
	case LeftToRight:
		return "Left->Right"
	case RightToLeft:
		return "Right->Left"
	}

	return LeftToRight.String()
}

func ParseTextDirection(direction string) TextDirection {
	d := strings.ToLower(direction)

	l := strings.Index(d, "left")
	r := strings.Index(d, "right")

	if l > 0 && r > 0 && l > r {
		return RightToLeft
	} else {
		return LeftToRight
	}
}
