package mail

type ContentType uint8

const (
	// TextPlain sets body type to text/plain in message body
	ContentPlainText ContentType = iota
	// TextHTML sets body type to text/html in message body
	ContentHTML
)

func (c ContentType) String() string {
	switch c {
	case ContentPlainText:
		return "Plain Text"
	case ContentHTML:
		return "HTML"
	}

	return ContentPlainText.String()
}

func (c ContentType) getContentType() string {
	switch c {
	case ContentPlainText:
		return "text/plain"
	case ContentHTML:
		return "text/html"
	}

	return ContentPlainText.getContentType()
}
