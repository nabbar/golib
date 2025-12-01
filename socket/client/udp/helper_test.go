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

package udp_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/udp"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// getFreePort returns a free TCP port
func getFreePort() int {
	addr, err := net.ResolveUDPAddr(libptc.NetworkUDP.Code(), "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	lstn, err := net.ListenUDP(libptc.NetworkUDP.Code(), addr)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		_ = lstn.Close()
	}()

	return lstn.LocalAddr().(*net.UDPAddr).Port
}

// getTestAddress returns a unique address for each test
func getTestAddress() string {
	return fmt.Sprintf("localhost:%d", getFreePort())
}

// echoHandler echoes back the received data
func echoHandler(ctx libsck.Context) {
	defer func() {
		_ = ctx.Close()
	}()
	_, _ = io.Copy(ctx, ctx)
}

// silentHandler accepts data but doesn't respond
func silentHandler(ctx libsck.Context) {
	defer func() {
		_ = ctx.Close()
	}()
	buf := make([]byte, 8192)
	_, _ = ctx.Read(buf)
}

// closingHandler closes the connection immediately
func closingHandler(ctx libsck.Context) {
	defer func() {
		_ = ctx.Close()
	}()
	// Just return to close
}

// countingHandler counts messages and stores in provided counter
func countingHandler(counter *atomic.Int32) libsck.HandlerFunc {
	return func(ctx libsck.Context) {
		defer func() {
			_ = ctx.Close()
		}()
		buf := make([]byte, 8192)
		n, err := ctx.Read(buf)
		if err == nil && n > 0 {
			counter.Add(1)
			_, _ = ctx.Write(buf[:n])
		}
	}
}

// startServer starts a UDP server in a goroutine
func startServer(ctx context.Context, srv scksrv.ServerUdp) {
	go func() {
		_ = srv.Listen(ctx)
	}()
}

// waitForServerRunning waits for the server to be running by attempting to connect
func waitForServerRunning(address string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(globalCtx, timeout)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			Fail(fmt.Sprintf("Timeout waiting for server to start at %s after %v", address, timeout))
			return
		case <-ticker.C:
			// Try to dial UDP to check if server is accepting
			if c, e := net.DialTimeout("udp", address, 100*time.Millisecond); e == nil {
				_ = c.Close()
				// Give server a bit more time to fully initialize
				time.Sleep(50 * time.Millisecond)
				return
			}
		}
	}
}

// waitForServerStopped waits for the server to stop
func waitForServerStopped(srv scksrv.ServerUdp, timeout time.Duration) {
	Eventually(func() bool {
		return !srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// createClient creates a new UDP client
func createClient(address string) sckclt.ClientUDP {
	cli, err := sckclt.New(address)
	Expect(err).ToNot(HaveOccurred())
	Expect(cli).ToNot(BeNil())
	return cli
}

// connectClient connects a client to the server
func connectClient(ctx context.Context, cli sckclt.ClientUDP) {
	err := cli.Connect(ctx)
	Expect(err).ToNot(HaveOccurred())
}

// waitForClientConnected waits for the client to be connected
func waitForClientConnected(cli sckclt.ClientUDP, timeout time.Duration) {
	Eventually(func() bool {
		return cli.IsConnected()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// createServer creates a UDP server with handler
func createServer(hdl libsck.HandlerFunc, adr string) scksrv.ServerUdp {
	cfg := sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: adr,
	}

	srv, err := scksrv.New(nil, hdl, cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(srv).ToNot(BeNil())

	return srv
}

// createAndRegisterServer creates a server with address and handler
func createAndRegisterServer(adr string, hdl libsck.HandlerFunc) scksrv.ServerUdp {
	srv := createServer(hdl, adr)
	return srv
}

// createSimpleTestServer creates and starts a simple echo server
func createSimpleTestServer(ctx context.Context, adr string) scksrv.ServerUdp {
	srv := createServer(echoHandler, adr)
	startServer(ctx, srv)
	waitForServerRunning(adr, 5*time.Second)
	return srv
}

// testError tracks error state for testing
type testError struct {
	mu     sync.Mutex
	errors []error
}

func newTestError() *testError {
	return &testError{
		errors: make([]error, 0),
	}
}

func (te *testError) add(errs ...error) {
	te.mu.Lock()
	defer te.mu.Unlock()
	for _, err := range errs {
		if err != nil {
			te.errors = append(te.errors, err)
		}
	}
}

func (te *testError) count() int {
	te.mu.Lock()
	defer te.mu.Unlock()
	return len(te.errors)
}

func (te *testError) last() error {
	te.mu.Lock()
	defer te.mu.Unlock()
	if len(te.errors) == 0 {
		return nil
	}
	return te.errors[len(te.errors)-1]
}

func (te *testError) clear() {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.errors = make([]error, 0)
}

// testDatagramCounter tracks received datagrams
type testDatagramCounter struct {
	mu    sync.Mutex
	count int
	data  [][]byte
}

func newDatagramCounter() *testDatagramCounter {
	return &testDatagramCounter{
		data: make([][]byte, 0),
	}
}

func (dc *testDatagramCounter) add(data []byte) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	buf := make([]byte, len(data))
	copy(buf, data)
	dc.data = append(dc.data, buf)
	dc.count++
}

func (dc *testDatagramCounter) getCount() int {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	return dc.count
}

func (dc *testDatagramCounter) getData() [][]byte {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	result := make([][]byte, len(dc.data))
	copy(result, dc.data)
	return result
}

// cleanupServer shuts down a server if it's running
func cleanupServer(srv scksrv.ServerUdp, ctx context.Context) {
	if srv != nil && srv.IsRunning() {
		_ = srv.Shutdown(ctx)
	}
}

// cleanupClient closes a client if it's connected
func cleanupClient(cli sckclt.ClientUDP) {
	if cli != nil && cli.IsConnected() {
		_ = cli.Close()
	}
}

// waitForCondition waits for a condition to be true with timeout
func waitForCondition(timeout time.Duration, check func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// simpleEchoHandler creates an echo handler that responds to datagrams
func simpleEchoHandler() libsck.HandlerFunc {
	return func(ctx libsck.Context) {
		defer func() {
			_ = ctx.Close()
		}()
		_, _ = io.Copy(ctx, ctx)
	}
}

// countingEchoHandler creates a handler that counts and echoes datagrams
func countingEchoHandler(counter *testDatagramCounter) libsck.HandlerFunc {
	return func(ctx libsck.Context) {
		defer func() {
			_ = ctx.Close()
		}()
		buf := make([]byte, 8192)
		for {
			n, err := ctx.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				counter.add(buf[:n])
				_, _ = ctx.Write(buf[:n])
			}
		}
	}
}

// createTestServerAndClient creates a server and client pair for testing
func createTestServerAndClient(handler libsck.HandlerFunc) (scksrv.ServerUdp, sckclt.ClientUDP, string, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(globalCtx, 30*time.Second)
	address := getTestAddress()

	srv := createServer(handler, address)
	startServer(ctx, srv)
	waitForServerRunning(address, 5*time.Second)

	cli := createClient(address)

	return srv, cli, address, ctx, cancel
}
