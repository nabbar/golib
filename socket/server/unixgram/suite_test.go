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

package unixgram_test

import (
	"context"
	"io"
	"net"
	"os"
	"testing"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var x context.Context

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestSocketServerUnixgram(t *testing.T) {
	x = context.Background()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Server Unixgram Suite")
}

func getTempSocketPath() string {
	f, _ := os.CreateTemp("", "test-*.sock")
	path := f.Name()

	_ = f.Close()
	_ = os.Remove(path)

	return path
}

var echoHandler = func(r libsck.Reader, w libsck.Writer) {
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	_, _ = io.Copy(w, r)
}

func createAndRegisterServer(path string, handler libsck.Handler) libsck.Server {
	srv := scksrv.New(nil, handler)
	Expect(srv).ToNot(BeNil())
	err := srv.RegisterSocket(path, 0600, -1)
	Expect(err).ToNot(HaveOccurred())
	return srv
}

func startServer(ctx context.Context, srv libsck.Server) {
	go func() {
		defer GinkgoRecover()
		_ = srv.Listen(ctx)
	}()
}

func waitForServerRunning(srv libsck.Server, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if srv.IsRunning() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	Fail("Server did not start")
}

func waitForServerStopped(srv libsck.Server, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if !srv.IsRunning() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	Fail("Server did not stop")
}

func sendDatagram(path string, data []byte) error {
	addr, _ := net.ResolveUnixAddr("unixgram", path)

	conn, err := net.DialUnix("unixgram", nil, addr)
	defer func() {
		_ = conn.Close()
	}()

	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}
