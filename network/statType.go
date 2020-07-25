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
