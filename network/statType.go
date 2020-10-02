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

package network

import (
	"fmt"
	"sort"
	"strings"
)

const (
	_MaxSizePadStatNamed_ = 7
	_DefaultPrecision_    = 2
)

type Stats uint8

const (
	StatBytes Stats = iota + 1
	StatPackets
	StatFifo
	StatDrop
	StatErr
)

func (s Stats) String() string {
	switch s {
	case StatBytes:
		return "Traffic"
	case StatPackets:
		return "Packets"
	case StatFifo:
		return "Fifo"
	case StatDrop:
		return "Drop"
	case StatErr:
		return "Error"
	}

	return ""
}

func (s Stats) FormatUnitInt(n Number) string {
	switch s {
	case StatBytes:
		return n.AsBytes().FormatUnitInt()
	case StatPackets, StatFifo, StatDrop, StatErr:
		return n.FormatUnitInt()
	}

	return ""
}

func (s Stats) FormatUnitFloat(n Number, precision int) string {
	switch s {
	case StatBytes:
		return n.AsBytes().FormatUnitFloat(precision)
	case StatPackets, StatFifo, StatDrop, StatErr:
		return n.FormatUnitFloat(precision)
	}

	return ""
}

func (s Stats) FormatUnit(n Number) string {
	switch s {
	case StatBytes:
		return n.AsBytes().FormatUnitFloat(_DefaultPrecision_)
	case StatPackets, StatFifo, StatDrop, StatErr:
		return n.FormatUnitInt()
	}

	return ""
}

func (s Stats) FormatLabelUnitPadded(n Number) string {
	return fmt.Sprintf("%s: %s%s", s.String(), strings.Repeat(" ", _MaxSizePadStatNamed_-len(s.String())), s.FormatUnit(n))
}

func (s Stats) FormatLabelUnit(n Number) string {
	return fmt.Sprintf("%s: %s", s.String(), s.FormatUnit(n))
}

func ListStatsSort() []int {
	l := []int{
		int(StatBytes),
		int(StatPackets),
		int(StatFifo),
		int(StatDrop),
		int(StatErr),
	}
	sort.Ints(l)
	return l
}
