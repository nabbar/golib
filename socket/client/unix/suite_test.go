//go:build linux || darwin

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

package unix_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/unix"
	scksrv "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSocketClientUnix(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Client UNIX Suite")
}

var (
	globalCtx context.Context
	globalCnl context.CancelFunc
)

var _ = BeforeSuite(func() {
	globalCtx, globalCnl = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	if globalCnl != nil {
		globalCnl()
	}
})

// getTestSocketPath returns a unique temp file path for UNIX socket
func getTestSocketPath() string {
	tmpDir := os.TempDir()
	socketName := fmt.Sprintf("test-unix-%d-%d.sock", time.Now().UnixNano(), os.Getpid())
	return filepath.Join(tmpDir, socketName)
}

// cleanupSocket removes the socket file if it exists
func cleanupSocket(socketPath string) {
	_ = os.Remove(socketPath)
}

// echoHandler echoes back the received data
func echoHandler(r libsck.Reader, w libsck.Writer) {
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	_, _ = io.Copy(w, r)
}

// silentHandler accepts data but doesn't respond
func silentHandler(r libsck.Reader, w libsck.Writer) {
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	buf := make([]byte, 8192)
	_, _ = r.Read(buf)
}

// closingHandler closes the connection immediately
func closingHandler(r libsck.Reader, w libsck.Writer) {
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	// Just return to close
}

// countingHandler counts messages and stores in provided counter
func countingHandler(counter *atomic.Int32) libsck.HandlerFunc {
	return func(r libsck.Reader, w libsck.Writer) {
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()

		buf := make([]byte, 8192)
		n, err := r.Read(buf)
		if err == nil && n > 0 {
			counter.Add(1)
			_, _ = w.Write(buf[:n])
		}
	}
}

// startServer starts a UNIX server in a goroutine
func startServer(ctx context.Context, srv scksrv.ServerUnix) {
	go func() {
		_ = srv.Listen(ctx)
	}()
}

// waitForServerRunning waits for server to be running
func waitForServerRunning(socketPath string, timeout time.Duration) {
	Eventually(func() bool {
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// waitForServerStopped waits for server to stop
func waitForServerStopped(srv scksrv.ServerUnix, timeout time.Duration) {
	Eventually(func() bool {
		return !srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// createClient creates a UNIX client
func createClient(socketPath string) sckclt.ClientUnix {
	cli := sckclt.New(socketPath)
	Expect(cli).ToNot(BeNil())
	return cli
}

// connectClient connects a client
func connectClient(ctx context.Context, cli sckclt.ClientUnix) {
	err := cli.Connect(ctx)
	Expect(err).ToNot(HaveOccurred())
}

// waitForClientConnected waits for the client to be connected
func waitForClientConnected(cli sckclt.ClientUnix, timeout time.Duration) {
	Eventually(func() bool {
		return cli.IsConnected()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// createServer creates a UNIX server with handler
func createServer(handler libsck.HandlerFunc) scksrv.ServerUnix {
	srv := scksrv.New(nil, handler)
	Expect(srv).ToNot(BeNil())
	return srv
}

// createAndRegisterServer creates a server with socket path and handler
func createAndRegisterServer(socketPath string, handler libsck.HandlerFunc) scksrv.ServerUnix {
	srv := scksrv.New(nil, handler)
	Expect(srv).ToNot(BeNil())

	err := srv.RegisterSocket(socketPath, 0600, -1)
	Expect(err).ToNot(HaveOccurred())

	return srv
}

// createSimpleTestServer creates and starts a simple echo server
func createSimpleTestServer(ctx context.Context, socketPath string) scksrv.ServerUnix {
	cleanupSocket(socketPath)
	srv := createAndRegisterServer(socketPath, echoHandler)
	startServer(ctx, srv)
	waitForServerRunning(socketPath, 5*time.Second)
	return srv
}
