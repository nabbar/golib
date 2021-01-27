package mail

const (
	headerFrom       = "From"
	headerSender     = "Sender"
	headerReplyTo    = "Reply-To"
	headerReturnPath = "Return-Path"
	headerTo         = "To"
	headerCc         = "Cc"
	headerBcc        = "Bcc"
)

type email struct {
	from       string
	sender     string
	replyTo    string
	returnPath string

	to  []string
	cc  []string
	bcc []string
}

func (e *email) SetFrom(mail string) {
	e.from = mail
}

func (e *email) SetSender(mail string) {
	e.sender = mail
}

func (e *email) SetReplyTo(mail string) {
	e.replyTo = mail
}

func (e *email) SetReturnPath(mail string) {
	e.returnPath = mail
}

func (e *email) GetFrom() string {
	if e.from != "" {
		return e.from
	}

	return ""
}

func (e *email) GetSender() string {
	if e.sender != "" {
		return e.sender
	}

	if e.replyTo != "" {
		return e.replyTo
	}

	if e.returnPath != "" {
		return e.returnPath
	}

	return ""
}

func (e *email) GetReplyTo() string {
	if e.replyTo != "" {
		return e.replyTo
	}

	if e.sender != "" {
		return e.sender
	}

	if e.returnPath != "" {
		return e.returnPath
	}

	return ""
}

func (e *email) GetReturnPath() string {
	if e.returnPath != "" {
		return e.returnPath
	}

	if e.sender != "" {
		return e.sender
	}

	if e.replyTo != "" {
		return e.replyTo
	}

	return ""
}

func (e *email) GetRecipients(rt recipientType) []string {
	switch rt {
	case RecipientTo:
		return e.to
	case RecipientCC:
		return e.cc
	case RecipientBCC:
		return e.bcc
	}

	return make([]string, 0)
}

func (e *email) SetRecipients(rt recipientType, rcpt ...string) {
	switch rt {
	case RecipientTo:
		e.to = make([]string, 0)
	case RecipientCC:
		e.cc = make([]string, 0)
	case RecipientBCC:
		e.bcc = make([]string, 0)
	default:
		return
	}

	e.AddRecipients(rt, rcpt...)
}

func (e *email) AddRecipients(rt recipientType, rcpt ...string) {
	for _, s := range rcpt {
		switch rt {
		case RecipientTo:
			if !e.isInSlice(e.to, s) {
				e.to = append(e.to, s)
			}
		case RecipientCC:
			if !e.isInSlice(e.cc, s) {
				e.cc = append(e.cc, s)
			}
		case RecipientBCC:
			if !e.isInSlice(e.bcc, s) {
				e.bcc = append(e.bcc, s)
			}
		}
	}
}

func (e email) isInSlice(s []string, str string) bool {
	if str == "" {
		return true
	}

	for _, i := range s {
		if i == str {
			return true
		}
	}

	return false
}

func (e *email) getHeader(h func(key string, values ...string)) {
	h(headerFrom, e.GetFrom())
	h(headerSender, e.GetSender())
	h(headerReplyTo, e.GetReplyTo())
	h(headerReturnPath, e.GetReturnPath())
	h(headerTo, e.GetRecipients(RecipientTo)...)
	h(headerCc, e.GetRecipients(RecipientCC)...)
	h(headerBcc, e.GetRecipients(RecipientBCC)...)
}
