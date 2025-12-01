//go:build linux

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

// Package client provides a unified factory for creating socket clients
// across different network protocols on Linux platforms.
//
// This package serves as a convenience wrapper that creates appropriate
// client implementations based on the specified network protocol:
//   - TCP, TCP4, TCP6: Connection-oriented network sockets (see github.com/nabbar/golib/socket/client/tcp)
//   - UDP, UDP4, UDP6: Connectionless datagram network sockets (see github.com/nabbar/golib/socket/client/udp)
//   - Unix: Connection-oriented UNIX domain sockets (see github.com/nabbar/golib/socket/client/unix)
//   - UnixGram: Connectionless UNIX datagram sockets (see github.com/nabbar/golib/socket/client/unixgram)
//
// All created clients implement the github.com/nabbar/golib/socket.Client interface,
// providing a consistent API regardless of the underlying protocol.
//
// Example:
//
//	cfg := config.Client{
//	    Network: protocol.NetworkTCP,
//	    Address: "localhost:8080",
//	}
//	cli, err := client.New(cfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cli.Close()
package client

import (
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/tcp"
	sckclu "github.com/nabbar/golib/socket/client/udp"
	sckclx "github.com/nabbar/golib/socket/client/unix"
	sckgrm "github.com/nabbar/golib/socket/client/unixgram"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// New creates a new socket client based on the specified network protocol.
//
// This factory function instantiates the appropriate client implementation
// for the given protocol type. On Linux platforms, all protocol types are
// supported, including UNIX domain sockets and abstract socket namespace.
//
// Parameters:
//   - cfg: Client configuration from github.com/nabbar/golib/socket/config package.
//     Contains network type, address, and optional TLS configuration.
//     Supported network values:
//   - NetworkTCP, NetworkTCP4, NetworkTCP6: TCP clients
//   - NetworkUDP, NetworkUDP4, NetworkUDP6: UDP clients
//   - NetworkUnix: UNIX domain stream socket clients
//   - NetworkUnixGram: UNIX domain datagram socket clients
//   - def: Default TLS configuration (optional, can be nil).
//     Used as a base for TCP client TLS configuration if cfg.TLS.Enabled is true.
//
// Address format depends on the protocol:
//   - TCP/UDP: "host:port" format (e.g., "localhost:8080", "192.168.1.1:9000")
//   - UNIX: filesystem path (e.g., "/tmp/app.sock") or abstract socket (e.g., "@abstract")
//
// Returns:
//   - libsck.Client: A client instance implementing the socket.Client interface
//   - error: An error if:
//   - Configuration validation fails (invalid network or empty address)
//   - Protocol is not supported on this platform
//   - Underlying protocol implementation fails to create client
//   - TLS configuration is invalid (TCP only)
//
// The function uses panic recovery to catch and log unexpected errors during
// client creation. All panics are recovered and logged via RecoveryCaller.
//
// Example:
//
//	// Create TCP client
//	cfg := config.Client{
//	    Network: protocol.NetworkTCP,
//	    Address: "localhost:8080",
//	}
//	cli, err := New(cfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cli.Close()
//
//	// Create UNIX socket client
//	unixCfg := config.Client{
//	    Network: protocol.NetworkUnix,
//	    Address: "/tmp/app.sock",
//	}
//	unixCli, err := New(unixCfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer unixCli.Close()
//
//	// Create abstract UNIX socket client (Linux only)
//	abstractCfg := config.Client{
//	    Network: protocol.NetworkUnix,
//	    Address: "@myapp",
//	}
//	abstractCli, err := New(abstractCfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer abstractCli.Close()
func New(cfg sckcfg.Client, def libtls.TLSConfig) (libsck.Client, error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/client", r)
		}
	}()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	switch cfg.Network {
	case libptc.NetworkUnix:
		return sckclx.New(cfg.Address), nil
	case libptc.NetworkUnixGram:
		return sckgrm.New(cfg.Address), nil
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		c, err := sckclt.New(cfg.Address)
		if err != nil {
			return nil, err
		} else if err = c.SetTLS(cfg.TLS.Enabled, cfg.TLS.Config.NewFrom(def), cfg.TLS.ServerName); err != nil {
			return nil, err
		} else {
			return c, nil
		}
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		return sckclu.New(cfg.Address)
	default:
		return nil, sckcfg.ErrInvalidProtocol
	}
}
