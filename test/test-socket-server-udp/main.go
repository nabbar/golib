/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	netptl "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

func config() sckcfg.ServerConfig {
	return sckcfg.ServerConfig{
		Network:  netptl.NetworkUDP,
		Address:  ":9001",
		PermFile: 0,
	}
}

func Handler(request libsck.Reader, response libsck.Writer) {
	_, e := io.Copy(os.Stdout, request)
	printError(e)
}

func printError(err ...error) {
	for _, e := range err {
		if e == nil {
			continue
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", e)
	}
}

func checkPanic(err ...error) {
	var found = false
	for _, e := range err {
		if e == nil {
			continue
		}

		found = true
		printError(err...)
		break
	}

	if found {
		panic(nil)
	}
}

func main() {
	srv, err := config().New(nil, Handler)
	checkPanic(err)

	srv.RegisterFuncError(func(e ...error) {
		printError(e...)
	})
	srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		_, _ = fmt.Fprintf(os.Stdout, "[%s %s]=>[%s %s] %s\n", remote.Network(), remote.String(), local.Network(), local.String(), state.String())
	})
	srv.RegisterFuncInfoServer(func(msg string) {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", msg)
	})

	checkPanic(srv.Listen(context.Background()))
}
