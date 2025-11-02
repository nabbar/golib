/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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
	"os"
	"sync/atomic"
	"testing"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	x context.Context
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestSocketServerUDP(t *testing.T) {
	x = context.Background()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Server UDP Suite")
}

// Helper functions

// getTestAddress returns a free UDP address for testing
func getTestAddress() string {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	Expect(err).ToNot(HaveOccurred())

	conn, err := net.ListenUDP("udp", addr)
	Expect(err).ToNot(HaveOccurred())

	port := conn.LocalAddr().(*net.UDPAddr).Port
	_ = conn.Close()

	return fmt.Sprintf("127.0.0.1:%d", port)
}

// echoHandler is a simple echo handler for testing
var echoHandler = func(request libsck.Reader, response libsck.Writer) {
	defer func() {
		_ = request.Close()
		_ = response.Close()
	}()
	_, _ = io.Copy(response, request)
}

// createAndRegisterServer creates a UDP server with the given handler
func createAndRegisterServer(address string, handler libsck.Handler, updateConn libsck.UpdateConn) libsck.Server {
	srv := scksrv.New(updateConn, handler)
	Expect(srv).ToNot(BeNil())

	err := srv.RegisterServer(address)
	Expect(err).ToNot(HaveOccurred())

	return srv
}

// startServer starts the server in a goroutine
func startServer(ctx context.Context, srv libsck.Server) {
	go func() {
		defer GinkgoRecover()
		err := srv.Listen(ctx)
		if err != nil && err.Error() != "context closed" {
			GinkgoWriter.Printf("Server listen error: %v\n", err)
		}
	}()
}

// waitForServerRunning waits for the server to be running
func waitForServerRunning(srv libsck.Server, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if srv.IsRunning() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	Fail("Server did not start within timeout")
}

// waitForServerStopped waits for the server to be stopped
func waitForServerStopped(srv libsck.Server, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if !srv.IsRunning() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	Fail("Server did not stop within timeout")
}

// sendDatagram sends a UDP datagram to the specified address
func sendDatagram(address string, data []byte) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Write(data)
	return err
}

// receiveDatagram receives a UDP datagram
func receiveDatagram(conn *net.UDPConn, timeout time.Duration) ([]byte, error) {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 65507) // Maximum UDP packet size
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Counter helpers using atomic operations
type atomicCounter struct {
	value atomic.Int64
}

func (c *atomicCounter) Increment() {
	c.value.Add(1)
}

func (c *atomicCounter) Get() int64 {
	return c.value.Load()
}

func (c *atomicCounter) Reset() {
	c.value.Store(0)
}
