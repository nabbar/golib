/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

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
