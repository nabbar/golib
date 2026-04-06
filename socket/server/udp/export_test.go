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

package udp

import (
	"net"

	libsck "github.com/nabbar/golib/socket"
)

// # Internal Logic Exposure for Testing
//
// This file uses the "export_test.go" pattern to expose internal (private)
// package members to the external "udp_test" package.
//
// # Use Case: Unit Testing
//
// This allows testing the error reporting logic (fctError) and state flags
// without making them part of the public API of the package.
// It also enables testing "nil receiver" cases on internal types.

// SCtxPublic provides an exported alias for the internal 'sCtx' struct.
// It is used in tests to verify robust behavior on nil context pointers.
type SCtxPublic = sCtx

// TestFctError allows tests to manually trigger the error callback logic.
func (o *srv) TestFctError(e ...error) {
	o.fctError(e...)
}

// TestFctInfo allows tests to manually trigger the datagram info callback.
func (o *srv) TestFctInfo(local, remote net.Addr, state libsck.ConnState) {
	o.fctInfo(local, remote, state)
}

// TestFctInfoSrv allows tests to manually trigger the server lifecycle callback.
func (o *srv) TestFctInfoSrv(msg string, args ...interface{}) {
	o.fctInfoSrv(msg, args...)
}
