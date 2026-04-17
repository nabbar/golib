/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package unixgram

import (
	"context"
	"net"

	libprm "github.com/nabbar/golib/file/perm"
	libsck "github.com/nabbar/golib/socket"
)

func (o *srv) TestFctError(e ...error) {
	o.fctError(e...)
}

func (o *srv) TestFctInfo(local, remote net.Addr, state libsck.ConnState) {
	o.fctInfo(local, remote, state)
}

func (o *srv) TestFctInfoSrv(msg string, args ...interface{}) {
	o.fctInfoSrv(msg, args...)
}

func (o *srv) TestGetSocketFile() (string, error) {
	return o.getSocketFile()
}

func (o *srv) TestGetSocketPerm() libprm.Perm {
	return o.getSocketPerm()
}

func (o *srv) TestGetSocketGroup() int {
	return o.getSocketGroup()
}

func (o *srv) TestCheckFile(unixFile string) (string, error) {
	return o.checkFile(unixFile)
}

func (o *srv) TestGetGoneChan() <-chan struct{} {
	return o.getGoneChan()
}

func (o *srv) TestGetContext(ctx context.Context, cnl context.CancelFunc, con *net.UnixConn, loc string) libsck.Context {
	return o.getContext(ctx, cnl, con, loc)
}

func (o *srv) TestPutContext(c libsck.Context) {
	if sc, ok := c.(*sCtx); ok {
		o.putContext(sc)
	}
}

func NewTestEmptyContext() libsck.Context {
	return &sCtx{}
}

func NewTestNilServer() libsck.Server {
	var o *srv
	return o
}

func (o *sCtx) TestOnErrorClose(e error) error {
	return o.onErrorClose(e)
}
