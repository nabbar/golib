package main

import (
	"fmt"
	"os"

	"github.com/nabbar/golib/ioutils"
)

func main() {
	println("test to print Max STDIO NOFILE capabilities !!")
	c, _, e := ioutils.SystemFileDescriptor(0)
	println(fmt.Sprintf("Actual limit is : %v | err : %v", c, e))

	if e != nil {
		os.Exit(1)
	}

	println("test to Change Max STDIO NOFILE capabilities !!")
	c, _, e = ioutils.SystemFileDescriptor(c + 512)
	println(fmt.Sprintf("New limit is : %v | err : %v", c, e))

	if e != nil {
		os.Exit(1)
	}
}
