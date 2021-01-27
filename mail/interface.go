package mail

import (
	"io"
	"net/textproto"
	"time"

	liberr "github.com/nabbar/golib/errors"
)

type Mail interface {
	SetCharset(charset string)
	GetCharset() string

	SetPriority(p Priority)
	GetPriority() Priority

	SetSubject(subject string)
	GetSubject() string

	SetEncoding(enc Encoding)
	GetEncoding() Encoding

	SetDateTime(datetime time.Time)
	GetDateTime() time.Time
	SetDateString(layout, datetime string) liberr.Error
	GetDateString() string

	AddHeader(key string, values ...string)
	GetHeader(key string) []string
	GetHeaders() textproto.MIMEHeader

	SetBody(ct ContentType, body io.ReadCloser)
	AddBody(ct ContentType, body io.ReadCloser)
	GetBody() []Body

	SetAttachment(name string, mime string, data io.ReadCloser, inline bool)
	AddAttachment(name string, mime string, data io.ReadCloser, inline bool)
	AttachFile(filepath string, data io.ReadCloser, inline bool)
	GetAttachment(inline bool) []File

	Email() Email

	Sender() (Sender, liberr.Error)
}

func New() Mail {
	m := &mail{
		headers:  make(textproto.MIMEHeader),
		charset:  "UTF-8",
		encoding: EncodingNone,
		address: &email{
			from:       "",
			sender:     "",
			replyTo:    "",
			returnPath: "",
			to:         make([]string, 0),
			cc:         make([]string, 0),
			bcc:        make([]string, 0),
		},
		attach: make([]File, 0),
		inline: make([]File, 0),
		body:   make([]Body, 0),
	}

	m.headers.Set("MIME-Version", "1.0")

	return m
}

type Email interface {
	SetFrom(mail string)
	GetFrom() string

	SetSender(mail string)
	GetSender() string

	SetReplyTo(mail string)
	GetReplyTo() string

	SetReturnPath(mail string)
	GetReturnPath() string

	SetRecipients(rt recipientType, rcpt ...string)
	AddRecipients(rt recipientType, rcpt ...string)
	GetRecipients(rt recipientType) []string
}
