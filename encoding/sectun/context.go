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

package sectun

import (
	"time"

	libsck "github.com/nabbar/golib/socket"
)

type ctx struct {
	x libsck.Context
	t SecureTunnel
}

func (o *ctx) Deadline() (deadline time.Time, ok bool) {
	return o.x.Deadline()
}

func (o *ctx) Done() <-chan struct{} {
	return o.x.Done()
}

func (o *ctx) Err() error {
	return o.x.Err()
}

func (o *ctx) Value(key any) any {
	return o.x.Value(key)
}

func (o *ctx) Read(p []byte) (n int, err error) {
	return o.t.Read(p)
}

func (o *ctx) Write(p []byte) (n int, err error) {
	return o.t.Write(p)
}

func (o *ctx) Close() error {
	err := o.x.Close()
	_ = o.t.Close()
	return err
}

func (o *ctx) IsConnected() bool {
	return o.x.IsConnected()
}

func (o *ctx) RemoteHost() string {
	return o.x.RemoteHost()
}

func (o *ctx) LocalHost() string {
	return o.x.LocalHost()
}
