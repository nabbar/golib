package smtp

import "strings"

type NETMode uint8

const (
	NET_TCP NETMode = iota
	NET_TCP_4
	NET_TCP_6
	NET_UNIX
)

func parseNETMode(str string) NETMode {
	switch strings.ToLower(str) {
	case NET_TCP_4.string():
		return NET_TCP_4
	case NET_TCP_6.string():
		return NET_TCP_6
	case NET_UNIX.string():
		return NET_UNIX
	}

	return NET_TCP
}

func (n NETMode) string() string {
	switch n {
	case NET_TCP_4:
		return "tcp4"
	case NET_TCP_6:
		return "tcp6"
	case NET_UNIX:
		return "unix"
	case NET_TCP:
		return "tcp"
	}

	return NET_TCP.string()
}
