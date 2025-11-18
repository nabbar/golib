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

package udp_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/udp"
	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSocketClientUDP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Client UDP Suite")
}

var (
	// Global test context
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
func echoHandler(r libsck.Reader, w libsck.Writer) {
	defer r.Close()
	defer w.Close()
	_, _ = io.Copy(w, r)
}

// silentHandler accepts data but doesn't respond
func silentHandler(r libsck.Reader, w libsck.Writer) {
	defer r.Close()
	defer w.Close()
	buf := make([]byte, 8192)
	_, _ = r.Read(buf)
}

// closingHandler closes the connection immediately
func closingHandler(r libsck.Reader, w libsck.Writer) {
	defer r.Close()
	defer w.Close()
	// Just return to close
}

// countingHandler counts messages and stores in provided counter
func countingHandler(counter *atomic.Int32) libsck.HandlerFunc {
	return func(r libsck.Reader, w libsck.Writer) {
		defer r.Close()
		defer w.Close()
		buf := make([]byte, 8192)
		n, err := r.Read(buf)
		if err == nil && n > 0 {
			counter.Add(1)
			_, _ = w.Write(buf[:n])
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
func createServer(handler libsck.HandlerFunc) scksrv.ServerUdp {
	srv := scksrv.New(nil, handler)
	Expect(srv).ToNot(BeNil())
	return srv
}

// createAndRegisterServer creates a server with address and handler
func createAndRegisterServer(address string, handler libsck.HandlerFunc) scksrv.ServerUdp {
	srv := scksrv.New(nil, handler)
	Expect(srv).ToNot(BeNil())

	err := srv.RegisterServer(address)
	Expect(err).ToNot(HaveOccurred())

	return srv
}

// createSimpleTestServer creates and starts a simple echo server
func createSimpleTestServer(ctx context.Context, address string) scksrv.ServerUdp {
	srv := createAndRegisterServer(address, echoHandler)
	startServer(ctx, srv)
	waitForServerRunning(address, 5*time.Second)
	return srv
}
