//go:build linux || darwin

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

// helper_test.go provides shared test utilities and helper functions.
// Includes server configuration creation, connection helpers, test socket path
// management, and common handler implementations used across all test files.
package unix_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// getTestSocketPath returns a unique temporary socket file path for testing.
// The socket file is automatically created in the OS temp directory with a unique name.
// The caller is responsible for cleaning up the socket file after use.
func getTestSocketPath() string {
	tdr := os.TempDir()
	return filepath.Join(tdr, fmt.Sprintf("test-unix-%d.sock", time.Now().UnixNano()))
}

// echoHandler is a simple echo handler for testing.
// It reads data from the connection and writes it back unchanged.
// Used for basic connectivity and data transfer tests.
func echoHandler(c libsck.Context) {
	defer func() {
		_ = c.Close()
	}()

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

// counterHandler counts connections and echoes data back.
// Used for testing connection tracking and counting.
func counterHandler(cnt *atomic.Int32) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()
		cnt.Add(1)

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

// slowHandler adds a delay before processing connections.
// Used for testing timeout and cancellation behavior.
func slowHandler(dly time.Duration) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()

		time.Sleep(dly)
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

// closeHandler immediately closes connections without processing.
// Used for testing connection lifecycle and cleanup.
func closeHandler(c libsck.Context) {
	_ = c.Close()
}

// writeOnlyHandler writes a fixed message and closes the connection.
// Used for testing write-only scenarios and server-initiated data.
func writeOnlyHandler(msg string) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()
		_, _ = c.Write([]byte(msg))
	}
}

// readOnlyHandler reads data but doesn't write anything back.
// Used for testing read-only scenarios and connection draining.
func readOnlyHandler(c libsck.Context) {
	defer func() {
		_ = c.Close()
	}()

	buf := make([]byte, 1024)
	for {
		_, err := c.Read(buf)
		if err != nil {
			return
		}
	}
}

// createDefaultConfig creates a default Unix socket server configuration.
// The socket path is provided as parameter and permissions are set to 0600 (owner only).
func createDefaultConfig(socketPath string) sckcfg.Server {
	return sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(0600),
		GroupPerm: -1, // Use default group
	}
}

// createConfigWithPerms creates a Unix socket server configuration with custom permissions.
// Allows testing different permission scenarios.
func createConfigWithPerms(socketPath string, perm uint32, gid int32) sckcfg.Server {
	return sckcfg.Server{
		Network:   libptc.NetworkUnix,
		Address:   socketPath,
		PermFile:  libprm.Perm(perm),
		GroupPerm: gid,
	}
}

// createConfigWithIdleTimeout creates a configuration with idle connection timeout.
// Used for testing idle timeout behavior and automatic connection cleanup.
func createConfigWithIdleTimeout(socketPath string, timeout time.Duration) sckcfg.Server {
	return sckcfg.Server{
		Network:        libptc.NetworkUnix,
		Address:        socketPath,
		PermFile:       libprm.Perm(0600),
		GroupPerm:      -1,
		ConIdleTimeout: libdur.ParseDuration(timeout),
	}
}

// waitForServer waits for the server to be running within the given timeout.
// Uses Eventually from Gomega to poll the server state.
func waitForServer(srv scksru.ServerUnix, timeout time.Duration) {
	Eventually(func() bool {
		return srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// waitForServerStopped waits for the server to stop running within the given timeout.
func waitForServerStopped(srv scksru.ServerUnix, timeout time.Duration) {
	Eventually(func() bool {
		return !srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// waitForConnections waits for the server to have exactly the expected number of connections.
func waitForConnections(srv scksru.ServerUnix, exp int64, timeout time.Duration) {
	Eventually(func() int64 {
		return srv.OpenConnections()
	}, timeout, 10*time.Millisecond).Should(Equal(exp))
}

// waitForGone waits for the server to be fully gone (IsGone returns true).
func waitForGone(srv scksru.ServerUnix, timeout time.Duration) {
	Eventually(func() bool {
		return srv.IsGone()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// connectToServer establishes a Unix socket connection to the given socket path.
// Returns the connection or fails the test if connection cannot be established.
func connectToServer(socketPath string) net.Conn {
	con, err := net.DialTimeout(libptc.NetworkUnix.Code(), socketPath, 2*time.Second)
	Expect(err).ToNot(HaveOccurred())
	Expect(con).ToNot(BeNil())
	return con
}

// sendAndReceive sends data to the connection and receives the response.
// Expects to receive exactly the same amount of data as sent.
// Used for testing echo handlers and bidirectional communication.
func sendAndReceive(con net.Conn, data []byte) []byte {
	n, err := con.Write(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(n).To(Equal(len(data)))

	buf := make([]byte, len(data))
	n, err = io.ReadFull(con, buf)
	Expect(err).ToNot(HaveOccurred())
	Expect(n).To(Equal(len(data)))

	return buf
}

// startServerInBackground starts the server in a goroutine.
// The server will listen with the provided context.
func startServerInBackground(c context.Context, srv scksru.ServerUnix) {
	go func() {
		_ = srv.Listen(c)
	}()
}

// waitForServerAcceptingConnections waits for the server to start accepting connections.
// It polls by attempting to connect to the socket until successful or timeout.
func waitForServerAcceptingConnections(socketPath string, timeout time.Duration) {
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	tck := time.NewTicker(50 * time.Millisecond)
	defer tck.Stop()

	for {
		select {
		case <-tmr.C:
			Fail(fmt.Sprintf("Timeout waiting for server to accept connections at %s after %v", socketPath, timeout))
			return
		case <-tck.C:
			if c, e := net.DialTimeout(libptc.NetworkUnix.Code(), socketPath, 100*time.Millisecond); e == nil {
				_ = c.Close()
				return
			}
		}
	}
}

// cleanupSocketFile removes the socket file if it exists.
// Safe to call even if the file doesn't exist.
func cleanupSocketFile(socketPath string) {
	_ = os.Remove(socketPath)
}
