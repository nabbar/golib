package mail

import (
	"io"
)

// contents represents the different content parts of an email body.
type Body struct {
	contentType ContentType
	body        io.Reader
}

func NewBody(ct ContentType, body io.Reader) Body {
	return Body{
		contentType: ct,
		body:        body,
	}
}
