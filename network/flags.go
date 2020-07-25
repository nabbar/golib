package network

import (
	"net"
)

func FindFlagInList(list []string, flag net.Flags) bool {
	for _, f := range list {
		if flag.String() == f {
			return true
		}
	}

	return false
}

func FindAllFlagInList(list []string, flags []net.Flags) bool {
	for _, f := range flags {
		if !FindFlagInList(list, f) {
			return false
		}
	}

	return true
}
