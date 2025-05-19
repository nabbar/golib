package kvdriver

type CompareEqual[K comparable] func(ref, part K) bool
type CompareContains[K comparable] func(ref, part K) bool
type CompareEmpty[K comparable] func(part K) bool

type Compare[K comparable] interface {
	IsEqual(ref, part K) bool
	IsContains(ref, part K) bool
	IsEmpty(part K) bool
}

func NewCompare[K comparable](eq CompareEqual[K], cn CompareContains[K], em CompareEmpty[K]) Compare[K] {
	return &cmp[K]{
		feq: eq,
		fcn: cn,
		fem: em,
	}
}

type cmp[K comparable] struct {
	feq CompareEqual[K]
	fcn CompareContains[K]
	fem CompareEmpty[K]
}

func (o *cmp[K]) IsEqual(ref, part K) bool {
	if o == nil || o.feq == nil {
		return false
	}

	return o.feq(ref, part)
}

func (o *cmp[K]) IsContains(ref, part K) bool {
	if o == nil || o.fcn == nil {
		return false
	}

	return o.fcn(ref, part)
}

func (o *cmp[K]) IsEmpty(part K) bool {
	if o == nil || o.fem == nil {
		return false
	}

	return o.fem(part)
}
