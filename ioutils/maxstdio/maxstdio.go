// +build windows,cgo

package maxstdio

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "maxstdio.h"
import "C"

func GetMaxStdio() int {
	return int(C.CGetMaxSTDIO())
}

func SetMaxStdio(newMax int) int {
	return int(C.CSetMaxSTDIO(C.int(newMax)))
}
