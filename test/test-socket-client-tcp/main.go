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
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	netptl "github.com/nabbar/golib/network/protocol"
	libsiz "github.com/nabbar/golib/size"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

func config() sckcfg.ClientConfig {
	return sckcfg.ClientConfig{
		Network:      netptl.NetworkTCP,
		Address:      ":9000",
		ReadBuffSize: 32 * libsiz.SizeKilo,
	}
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

func request() io.Reader {
	var buf = bytes.NewBuffer(make([]byte, 0))
	buf.WriteString(fmt.Sprintf("<14>%s myapp: This is a sample syslog message #%d\n", time.Now().Format(time.RFC3339), buf.Len()))
	buf.WriteString(fmt.Sprintf("<14>%s myapp: This is a sample syslog message #%d\n", time.Now().Format(time.RFC3339), buf.Len()))
	buf.WriteString(fmt.Sprintf("<14>%s myapp: This is a sample syslog message #%d\n", time.Now().Format(time.RFC3339), buf.Len()))
	buf.WriteString(fmt.Sprintf("<14>%s myapp: This is a sample syslog message #%d\n", time.Now().Format(time.RFC3339), buf.Len()))
	return buf
}

func main() {
	cli, err := config().New()
	checkPanic(err)

	cli.RegisterFuncError(func(e error) {
		printError(e)
	})
	cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
		_, _ = fmt.Fprintf(os.Stdout, "[%s %s]=>[%s %s] %s\n", remote.Network(), remote.String(), local.Network(), local.String(), state.String())
	})

	checkPanic(cli.Do(context.Background(), request(), func(r io.Reader) {
		_, e := io.Copy(os.Stdout, r)
		printError(e)
	}))
}
