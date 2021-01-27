package mail

const (
	headerImportance     = "Importance"
	headerMSMailPriority = "X-MSMail-Priority"
	headerPriority       = "X-Priority"
)

type Priority uint8

const (
	// PriorityNormal sets the email priority to normal
	PriorityNormal Priority = iota
	// PriorityLow sets the email priority to Low
	PriorityLow
	// PriorityHigh sets the email priority to High
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityHigh:
		return "High"
	case PriorityNormal:
		return "Normal"
	}

	return PriorityNormal.String()
}

func (p Priority) headerPriority() string {
	switch p {
	case PriorityLow:
		return "5 (Lowest)"
	case PriorityHigh:
		return "1 (Highest)"
	case PriorityNormal:
		return ""
	}

	return PriorityNormal.headerPriority()
}

func (p Priority) headerImportance() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityHigh:
		return "High"
	case PriorityNormal:
		return ""
	}

	return PriorityNormal.headerImportance()
}

func (p Priority) headerMSMailPriority() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityHigh:
		return "High"
	case PriorityNormal:
		return ""
	}

	return PriorityNormal.headerMSMailPriority()
}

func (p Priority) getHeader(h func(key string, values ...string)) {
	for k, f := range map[string]func() string{
		headerPriority:       p.headerPriority,
		headerMSMailPriority: p.headerMSMailPriority,
		headerImportance:     p.headerImportance,
	} {
		if v := f(); k != "" && v != "" {
			h(k, v)
		}
	}
}
