package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/network"
)

func init() {
	errors.SetModeReturnError(errors.ErrorReturnCodeErrorFull)
}

func main() {
	ifs, err := network.GetAllInterfaces(context.Background(), true, true, 0, net.FlagUp)

	if err != nil {
		panic(err)
	}

	for _, i := range ifs {
		fmt.Printf("iface '%s' [%s] : \n", i.GetName(), i.GetHardwareAddr())
		fmt.Printf("\tFlags: %s\n", strings.Join(i.GetFlags(), " "))
		fmt.Printf("\tAddrs: %s\n", strings.Join(i.GetAddr(), " "))

		fmt.Print("\tIn: \n")
		l := i.GetStatIn()
		for _, k := range network.ListStatsSort() {
			s := network.Stats(k)
			if v, ok := l[s]; ok {
				fmt.Printf("\t\t%s\n", s.FormatLabelUnitPadded(v))
			} else {
				fmt.Printf("\t\t%s\n", s.FormatLabelUnitPadded(0))
			}
		}

		fmt.Print("\tOut: \n")
		l = i.GetStatOut()
		for _, k := range network.ListStatsSort() {
			s := network.Stats(k)
			if v, ok := l[s]; ok {
				fmt.Printf("\t\t%s\n", s.FormatLabelUnitPadded(v))
			} else {
				fmt.Printf("\t\t%s\n", s.FormatLabelUnitPadded(0))
			}
		}

		println("\n")
	}
}
