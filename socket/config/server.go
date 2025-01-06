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

package config

import (
	"os"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server"
)

// ServerConfig define the server configuration
type ServerConfig struct {
	// network protocol
	Network libptc.NetworkProtocol ``
	// address to listen
	Address string
	// permission of owner for socket file
	PermFile os.FileMode
	// permission of group for socket file
	GroupPerm int32
}

// New returns a new server with the given handler and based on the ServerConfig
// handler libsck.Handler
// (libsck.Server, error)
func (o ServerConfig) New(updateCon libsck.UpdateConn, handler libsck.Handler) (libsck.Server, error) {
	return scksrv.New(updateCon, handler, o.Network, o.Address, o.PermFile, o.GroupPerm)
}
