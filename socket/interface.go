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

package socket

import (
	"context"
	"io"
	"net"
	"time"
)

const DefaultBufferSize = 32 * 1024

type ConnState uint8

const (
	ConnectionDial ConnState = iota
	ConnectionNew
	ConnectionRead
	ConnectionCloseRead
	ConnectionHandler
	ConnectionWrite
	ConnectionCloseWrite
	ConnectionClose
)

type FuncError func(e error)
type FuncInfo func(local, remote net.Addr, state ConnState)
type Handler func(request io.Reader, response io.Writer)
type Response func(r io.Reader)

type Server interface {
	RegisterFuncError(f FuncError)
	RegisterFuncInfo(f FuncInfo)

	SetReadTimeout(d time.Duration)
	SetWriteTimeout(d time.Duration)

	Listen(ctx context.Context) error
	Shutdown()
	Done() <-chan struct{}
}

type Client interface {
	RegisterFuncError(f FuncError)
	RegisterFuncInfo(f FuncInfo)

	Do(ctx context.Context, request io.Reader, fct Response) error
}
