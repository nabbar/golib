package network

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	_MaxSizeOfPad_  = 4
	_PadIntPattern_ = "%4d"
	_PowerYotta_    = 24
	_PowerZetta_    = 21
	_PowerExa_      = 18
	_PowerPeta_     = 15
	_PowerTera_     = 12
	_PowerGiga_     = 9
	_PowerMega_     = 6
	_PowerKilo_     = 3
	_PowerUnit_     = 0
)

type Number uint64

func (n Number) String() string {
	return fmt.Sprintf("%d", n)
}

func (n Number) FormatUnitFloat(precision int) string {
	if precision < 1 {
		return n.FormatUnitInt()
	}

	m := float64(n)

	f := fmt.Sprintf("%%.%df", precision)

	for _, p := range powerList() {
		if m < math.Pow10(p+1) {
			continue
		}

		r := m / math.Pow10(p)
		q := strings.SplitN(fmt.Sprintf(f, r), ".", 2)

		if len(q) > 0 {
			if len(q[0]) < _MaxSizeOfPad_ {
				return strings.Repeat(" ", _MaxSizeOfPad_-len(q[0])) + fmt.Sprintf(f+" %s", r, power2Unit(p))
			}

			return fmt.Sprintf(f+" %s", r, power2Unit(p))
		}
	}

	return strings.Repeat(" ", _MaxSizeOfPad_) + fmt.Sprintf(f+" %s", m, " ")
}

func (n Number) FormatUnitInt() string {
	m := float64(n)

	for _, p := range powerList() {
		if m < math.Pow10(p+1) {
			continue
		}

		r := int(math.Round(m / math.Pow10(p)))
		return fmt.Sprintf(_PadIntPattern_+" %s", r, power2Unit(p))
	}

	return fmt.Sprintf(_PadIntPattern_+" %s", n, " ")
}

func (n Number) AsBytes() Bytes {
	return Bytes(n)
}

func (n Number) AsUint64() uint64 {
	return uint64(n)
}

func (n Number) AsFloat64() float64 {
	return float64(n)
}

func power2Unit(power int) string {
	switch {
	case power >= _PowerYotta_:
		return "Y"
	case power >= _PowerZetta_:
		return "Z"
	case power >= _PowerExa_:
		return "E"
	case power >= _PowerPeta_:
		return "P"
	case power >= _PowerTera_:
		return "T"
	case power >= _PowerGiga_:
		return "G"
	case power >= _PowerMega_:
		return "M"
	case power >= _PowerKilo_:
		return "K"
	case power >= _PowerUnit_:
		return ""
	}

	return ""
}

func powerList() []int {
	var p = []int{_PowerYotta_, _PowerZetta_, _PowerExa_, _PowerPeta_, _PowerTera_, _PowerGiga_, _PowerMega_, _PowerKilo_, _PowerUnit_}
	sort.Sort(sort.Reverse(sort.IntSlice(p)))
	return p
}
