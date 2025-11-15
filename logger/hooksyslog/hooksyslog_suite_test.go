/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package hooksyslog_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ctx, cnl = context.WithCancel(context.Background())

	lstMsgs []string
	msgMux  sync.Mutex

	sckSrv  libsck.Server
	sckAddr = getTempSocketPath()
)

func TestHookSyslog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger HookSyslog Suite")
}

var _ = BeforeSuite(func() {
	var err error

	sckSrv, err = scksrv.New(nil, hookHandler, libptc.NetworkUnixGram, sckAddr, 0600, -1)

	Expect(err).ToNot(HaveOccurred())
	Expect(sckSrv).ToNot(BeNil())

	// Start sckSrv
	go func() {
		defer GinkgoRecover()
		if err := sckSrv.Listen(ctx); err != nil && strings.Contains(err.Error(), "close") {
			return
		} else {
			_, _ = fmt.Fprintf(GinkgoWriter, "error listening on sokcet file '%s': %v\n", sckAddr, err)
		}
	}()

	time.Sleep(10 * time.Millisecond) // Give sckSrv time to start
	waitForServerRunning(sckAddr, 5*time.Second)
})

var _ = AfterSuite(func() {
	if sckSrv != nil {
		_ = sckSrv.Close()
		time.Sleep(50 * time.Millisecond) // Give time to clean up
	}

	if cnl != nil {
		cnl()
	}

	_ = os.Remove(sckAddr)
})

func getTempSocketPath() string {
	f, _ := os.CreateTemp("", "test-*.sock")
	path := f.Name()

	_ = f.Close()
	_ = os.Remove(path)

	return path
}

// waitForServerRunning waits for the server to be running by attempting to connect
func waitForServerRunning(address string, timeout time.Duration) {
	x, n := context.WithTimeout(ctx, timeout)
	defer n()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-x.Done():
			Fail(fmt.Sprintf("Timeout waiting for server to start at %s after %v", address, timeout))
			return
		case <-ticker.C:
			if c, e := net.DialTimeout(libptc.NetworkUnixGram.Code(), address, 100*time.Millisecond); e == nil {
				_ = c.Close()
				return
			}
		}
	}
}

// Helper function to get received messages safely
func getReceivedMessages() []string {
	msgMux.Lock()
	defer msgMux.Unlock()
	return append([]string{}, lstMsgs...)
}

// Helper function to clear received messages
func clearReceivedMessages() {
	msgMux.Lock()
	defer msgMux.Unlock()
	lstMsgs = []string{}
}

func addReceivedMessages(msg string) {
	msgMux.Lock()
	defer msgMux.Unlock()
	lstMsgs = append(lstMsgs, msg)
}

func hookHandler(rd libsck.Reader, wr libsck.Writer) {
	defer func() {
		if rd != nil {
			_ = rd.Close()
		}
		if wr != nil {
			_ = wr.Close()
		}
	}()

	buf := make([]byte, 10240) // 10KB as default log message in syslog

	for {
		if ctx.Err() != nil {
			return
		}

		n, err := rd.Read(buf)

		if n > 0 {
			addReceivedMessages(string(buf[:n]))
		}

		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			if !strings.Contains(err.Error(), "close") {
				_, _ = fmt.Fprintf(GinkgoWriter, "failed to read: %v\n", err)
			}
			return
		}
	}
}
