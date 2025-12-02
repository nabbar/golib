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

package client_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server"
)

// basicHandler returns a simple handler that closes the connection immediately.
// This handler is used for basic server creation tests where we don't need
// actual I/O operations.
func basicHandler() libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}
}

// echoHandler returns a handler that echoes back received data.
// This is useful for integration tests where we need to verify data flow.
func echoHandler() libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				_, err = c.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}
	}
}

// getTestTCPAddress returns a test TCP address with a free port.
func getTestTCPAddress() string {
	return ":0" // Let OS choose port
}

// getTestUDPAddress returns a test UDP address with a free port.
func getTestUDPAddress() string {
	return ":0" // Let OS choose port
}

// getTestUnixPath generates a unique Unix socket path for testing.
func getTestUnixPath() string {
	tmpDir := os.TempDir()
	return filepath.Join(tmpDir, fmt.Sprintf("test-unix-%d.sock", time.Now().UnixNano()))
}

// getTestUnixGramPath generates a unique Unix datagram socket path for testing.
func getTestUnixGramPath() string {
	tmpDir := os.TempDir()
	return filepath.Join(tmpDir, fmt.Sprintf("test-unixgram-%d.sock", time.Now().UnixNano()))
}

// startTestServer starts a test server and returns the actual address it's listening on.
// The server is started in a goroutine and the caller is responsible for shutting it down.
func startTestServer(ctx context.Context, cfg sckcfg.Server, handler libsck.HandlerFunc) (libsck.Server, string, error) {
	srv, err := scksrv.New(nil, handler, cfg)
	if err != nil {
		return nil, "", err
	}

	// Start server in background
	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait for server to be ready
	time.Sleep(50 * time.Millisecond)

	// Get actual address (useful when using port 0)
	addr := cfg.Address
	if cfg.Network.IsTCP() || cfg.Network.IsUDP() {
		if _, adr, _ := srv.Listener(); len(adr) > 0 {
			addr = adr
		}
	}

	return srv, addr, nil
}

// startTestTCPServer starts a TCP server for testing.
func startTestTCPServer(ctx context.Context) (libsck.Server, string, error) {
	cfg := sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: getTestTCPAddress(),
	}
	return startTestServer(ctx, cfg, echoHandler())
}

// startTestUDPServer starts a UDP server for testing.
func startTestUDPServer(ctx context.Context) (libsck.Server, string, error) {
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: getTestUDPAddress(),
	}
	return startTestServer(ctx, cfg, echoHandler())
}

// getAvailablePort returns an available TCP port.
func getAvailablePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
