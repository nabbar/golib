/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package httpserver

import (
	"context"
	"net"
	"strings"

	liberr "github.com/nabbar/golib/errors"
	srvtps "github.com/nabbar/golib/httpserver/types"
)

// PortNotUse checks if the specified port is available (not in use).
// Returns nil if the port is available, or an error if the port is in use or invalid.
// The listen parameter should be in "host:port" format.
func PortNotUse(ctx context.Context, listen string) error {
	var (
		err error
		con net.Conn
		dia = net.Dialer{}
	)

	defer func() {
		if con != nil {
			_ = con.Close()
		}
	}()

	if strings.Contains(listen, ":") {
		part := strings.Split(listen, ":")

		if len(part) < 2 {
			return ErrorInvalidAddress.Error()
		}

		port := part[len(part)-1]
		addr := strings.Join(part[:len(part)-1], ":")

		if strings.HasPrefix(addr, "0") || strings.HasPrefix(addr, "::") {
			listen = "127.0.0.1:" + port
		}
	}

	if _, ok := ctx.Deadline(); !ok {
		var cnl context.CancelFunc
		ctx, cnl = context.WithTimeout(ctx, srvtps.TimeoutWaitingPortFreeing)
		defer cnl()
	}

	con, err = dia.DialContext(ctx, "tcp", listen)
	return err
}

// PortInUse checks if the specified port is currently in use.
// Returns nil if the port is in use, or ErrorPortUse if the port is available.
// The listen parameter should be in "host:port" format.
func PortInUse(ctx context.Context, listen string) liberr.Error {
	var (
		dia = net.Dialer{}
		con net.Conn
		err error
	)

	defer func() {
		if con != nil {
			_ = con.Close()
		}
	}()

	if strings.Contains(listen, ":") {
		part := strings.Split(listen, ":")

		if len(part) < 2 {
			return ErrorInvalidAddress.Error()
		}

		port := part[len(part)-1]
		addr := strings.Join(part[:len(part)-1], ":")

		if strings.HasPrefix(addr, "0") || strings.HasPrefix(addr, "::") {
			listen = "127.0.0.1:" + port
		}
	}

	if _, ok := ctx.Deadline(); !ok {
		var cnl context.CancelFunc
		ctx, cnl = context.WithTimeout(ctx, srvtps.TimeoutWaitingPortFreeing)
		defer cnl()
	}

	con, err = dia.DialContext(ctx, "tcp", listen)
	if err != nil {
		return nil
	}

	return ErrorPortUse.Error(nil)
}
