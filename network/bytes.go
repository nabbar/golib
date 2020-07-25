package network

import (
	"fmt"
	"math"
	"strings"
)

const (
	_PowBytesNumber_ float64 = 2
	_PowBytesPower_  float64 = 10
	_PowBytesFactor_ int     = 3
)

type Bytes uint64

func (n Bytes) String() string {
	return fmt.Sprintf("%d", n)
}

func (n Bytes) FormatUnitFloat(precision int) string {
	if precision < 1 {
		return n.FormatUnitInt()
	}

	m := float64(n)

	f := fmt.Sprintf("%%.%df", precision)

	for _, p := range powerList() {
		if m < math.Pow10(p+1) {
			continue
		}

		r := m / math.Pow(math.Pow(_PowBytesNumber_, _PowBytesPower_), float64(p/_PowBytesFactor_))
		q := strings.SplitN(fmt.Sprintf(f, r), ".", 2)

		if len(q) > 0 {
			if len(q[0]) < _MaxSizeOfPad_ {
				return strings.Repeat(" ", _MaxSizeOfPad_-len(q[0])) + fmt.Sprintf(f+" %s", r, power2Unit(p)+"B")
			}
			return fmt.Sprintf(f+" %s", r, power2Unit(p)+"B")
		}
	}

	return strings.Repeat(" ", _MaxSizeOfPad_) + fmt.Sprintf(f+" %s", m, " ")
}

func (n Bytes) FormatUnitInt() string {
	m := float64(n)

	for _, p := range powerList() {
		if m < math.Pow10(p+1) {
			continue
		}

		return fmt.Sprintf(_PadIntPattern_+" %s", int(math.Round(m/math.Pow(math.Pow(_PowBytesNumber_, _PowBytesPower_), float64(p/_PowBytesFactor_)))), power2Unit(p)+"B")
	}

	return fmt.Sprintf(_PadIntPattern_+" %s", n, " ")
}

func (n Bytes) AsNumber() Number {
	return Number(n)
}

func (n Bytes) AsUint64() uint64 {
	return uint64(n)
}

func (n Bytes) AsFloat64() float64 {
	return float64(n)
}
