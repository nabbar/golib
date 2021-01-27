package smtp

import "strings"

type TLSMode uint8

const (
	TLS_NONE TLSMode = iota
	TLS_STARTTLS
	TLS_TLS
)

func parseTLSMode(str string) TLSMode {
	switch strings.ToLower(str) {
	case TLS_TLS.string():
		return TLS_TLS
	case TLS_STARTTLS.string():
		return TLS_STARTTLS
	}

	return TLS_NONE
}

func (tlm TLSMode) string() string {
	switch tlm {
	case TLS_TLS:
		return "tls"
	case TLS_STARTTLS:
		return "starttls"
	case TLS_NONE:
		return "none"
	}

	return TLS_NONE.string()
}
