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

	. "github.com/onsi/gomega"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	"github.com/nabbar/golib/socket/server/udp"
)

// testHandler is a simple handler that stores received data
type testHandler struct {
	mu       sync.Mutex
	data     [][]byte
	count    atomic.Int64
	lastErr  error
	ctx      context.Context
	cancel   context.CancelFunc
	readOnce bool
}

// newTestHandler creates a new test handler
func newTestHandler(readOnce bool) *testHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &testHandler{
		data:     make([][]byte, 0),
		ctx:      ctx,
		cancel:   cancel,
		readOnce: readOnce,
	}
}

// handler is the HandlerFunc implementation
func (h *testHandler) handler(ctx libsck.Context) {
	defer ctx.Close()

	buf := make([]byte, 65507) // Max UDP datagram size

	if h.readOnce {
		// Read once and return
		n, err := ctx.Read(buf)
		if err != nil && err != io.EOF {
			h.setError(err)
			return
		}

		if n > 0 {
			h.addData(buf[:n])
		}
		return
	}

	// Continuous reading
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
		}

		n, err := ctx.Read(buf)
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe {
				return
			}
			h.setError(err)
			return
		}

		if n > 0 {
			h.addData(buf[:n])
		}
	}
}

// addData adds received data to the slice
func (h *testHandler) addData(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	copied := make([]byte, len(data))
	copy(copied, data)
	h.data = append(h.data, copied)
	h.count.Add(1)
}

// getData returns received data
func (h *testHandler) getData() [][]byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.data
}

// getCount returns number of received datagrams
func (h *testHandler) getCount() int64 {
	return h.count.Load()
}

// setError stores error
func (h *testHandler) setError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastErr = err
}

// getError returns last error
func (h *testHandler) getError() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lastErr
}

// stop stops the handler
func (h *testHandler) stop() {
	h.cancel()
}

// errorCollector collects errors from callbacks
type errorCollector struct {
	mu     sync.Mutex
	errors []error
}

// newErrorCollector creates a new error collector
func newErrorCollector() *errorCollector {
	return &errorCollector{
		errors: make([]error, 0),
	}
}

// callback is the FuncError callback
func (e *errorCollector) callback(errs ...error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, err := range errs {
		if err != nil {
			e.errors = append(e.errors, err)
		}
	}
}

// getErrors returns collected errors
func (e *errorCollector) getErrors() []error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.errors
}

// hasErrors returns true if errors were collected
func (e *errorCollector) hasErrors() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.errors) > 0
}

// clear clears collected errors
func (e *errorCollector) clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errors = make([]error, 0)
}

// infoCollector collects connection info events
type infoCollector struct {
	mu     sync.Mutex
	events []connEvent
}

type connEvent struct {
	local  string
	remote string
	state  libsck.ConnState
}

// newInfoCollector creates a new info collector
func newInfoCollector() *infoCollector {
	return &infoCollector{
		events: make([]connEvent, 0),
	}
}

// callback is the FuncInfo callback
func (i *infoCollector) callback(local, remote net.Addr, state libsck.ConnState) {
	i.mu.Lock()
	defer i.mu.Unlock()

	evt := connEvent{
		state: state,
	}

	if local != nil {
		evt.local = local.String()
	}

	if remote != nil {
		evt.remote = remote.String()
	}

	i.events = append(i.events, evt)
}

// getEvents returns collected events
func (i *infoCollector) getEvents() []connEvent {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.events
}

// hasState returns true if state was observed
func (i *infoCollector) hasState(state libsck.ConnState) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, evt := range i.events {
		if evt.state == state {
			return true
		}
	}
	return false
}

// clear clears collected events
func (i *infoCollector) clear() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.events = make([]connEvent, 0)
}

// serverInfoCollector collects server info messages
type serverInfoCollector struct {
	mu       sync.Mutex
	messages []string
}

// newServerInfoCollector creates a new server info collector
func newServerInfoCollector() *serverInfoCollector {
	return &serverInfoCollector{
		messages: make([]string, 0),
	}
}

// callback is the FuncInfoSrv callback
func (s *serverInfoCollector) callback(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

// getMessages returns collected messages
func (s *serverInfoCollector) getMessages() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messages
}

// hasMessage returns true if message contains substring
func (s *serverInfoCollector) hasMessage(substr string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, msg := range s.messages {
		if contains(msg, substr) {
			return true
		}
	}
	return false
}

// clear clears collected messages
func (s *serverInfoCollector) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make([]string, 0)
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createBasicConfig creates a basic test configuration
func createBasicConfig() sckcfg.Server {
	return sckcfg.Server{
		Network: libptc.NetworkUDP,
		Address: "127.0.0.1:0", // Use loopback, let OS choose port
	}
}

// createServerWithHandler creates server with handler
func createServerWithHandler(handler libsck.HandlerFunc) (udp.ServerUdp, error) {
	cfg := createBasicConfig()
	return udp.New(nil, handler, cfg)
}

// startServer starts server and waits for it to be running
func startServer(srv udp.ServerUdp, ctx context.Context) {
	go func() {
		_ = srv.Listen(ctx)
	}()

	// Wait for server to start
	Eventually(func() bool {
		return srv.IsRunning()
	}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
}

// stopServer stops server and waits for it to stop
func stopServer(srv udp.ServerUdp, cancel context.CancelFunc) {
	cancel()

	Eventually(func() bool {
		return !srv.IsRunning()
	}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
}

// sendUDPDatagram sends a UDP datagram to address
func sendUDPDatagram(address string, data []byte) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(data)
	return err
}

// getServerAddress extracts server address from info messages
func getServerAddress(collector *serverInfoCollector) string {
	messages := collector.getMessages()
	for _, msg := range messages {
		// Extract address from "starting listening socket 'udp 127.0.0.1:xxxxx'"
		if contains(msg, "starting listening socket") {
			// Simple extraction
			for i := 0; i < len(msg); i++ {
				if msg[i] == '1' && i+1 < len(msg) && msg[i+1] == '2' {
					// Found "12", likely start of 127.0.0.1
					for j := i; j < len(msg); j++ {
						if msg[j] == '\'' {
							return msg[i:j]
						}
					}
				}
			}
		}
	}
	return ""
}

// waitForCondition waits for condition with timeout
func waitForCondition(condition func() bool, timeout time.Duration, message string) {
	Eventually(condition, timeout, 10*time.Millisecond).Should(BeTrue(), message)
}

// customUpdateConn is a test UpdateConn callback
type customUpdateConn struct {
	mu     sync.Mutex
	called bool
	conn   net.Conn
}

// newCustomUpdateConn creates new custom update conn
func newCustomUpdateConn() *customUpdateConn {
	return &customUpdateConn{}
}

// callback is the UpdateConn callback
func (c *customUpdateConn) callback(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.called = true
	c.conn = conn

	// Configure socket
	if udpConn, ok := conn.(*net.UDPConn); ok {
		_ = udpConn.SetReadBuffer(1024 * 1024)
		_ = udpConn.SetWriteBuffer(1024 * 1024)
	}
}

// wasCalled returns true if callback was called
func (c *customUpdateConn) wasCalled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.called
}

// getConn returns stored connection
func (c *customUpdateConn) getConn() net.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}

// assertServerState checks server state
func assertServerState(srv udp.ServerUdp, expectedRunning, expectedGone bool, expectedConns int64) {
	Expect(srv.IsRunning()).To(Equal(expectedRunning),
		fmt.Sprintf("Expected IsRunning=%v", expectedRunning))
	Expect(srv.IsGone()).To(Equal(expectedGone),
		fmt.Sprintf("Expected IsGone=%v", expectedGone))
	Expect(srv.OpenConnections()).To(Equal(expectedConns),
		fmt.Sprintf("Expected OpenConnections=%d", expectedConns))
}
