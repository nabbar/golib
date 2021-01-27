package mail

type Encoding uint8

const (
	// EncodingNone turns off encoding on the message body
	EncodingNone Encoding = iota

	// EncodingBinary is equal to EncodingNone, but string is set to binrary instead of none
	EncodingBinary

	// EncodingBase64 sets the message body encoding to base64
	EncodingBase64

	// EncodingQuotedPrintable sets the message body encoding to quoted-printable
	EncodingQuotedPrintable
)

func (e Encoding) String() string {
	switch e {
	case EncodingBinary:
		return "Binary"
	case EncodingBase64:
		return "Base 64"
	case EncodingQuotedPrintable:
		return "Quoted Printable"
	case EncodingNone:
		return "None"
	}
	return EncodingNone.String()
}

func (e Encoding) getEncoding() string {
	switch e {
	case EncodingNone, EncodingBinary:
		return "binary"
	case EncodingBase64:
		return "base64"
	case EncodingQuotedPrintable:
		return "quoted-printable"
	}

	return EncodingNone.getEncoding()
}
