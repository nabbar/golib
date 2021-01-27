package mail

type recipientType uint8

const (
	RecipientTo recipientType = iota
	RecipientCC
	RecipientBCC
)

func (r recipientType) String() string {
	switch r {
	case RecipientTo:
		return "To"
	case RecipientCC:
		return "Cc"
	case RecipientBCC:
		return "Bcc"
	}

	return RecipientTo.String()
}
