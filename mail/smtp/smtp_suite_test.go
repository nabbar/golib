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

package smtp_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	smtpcli "github.com/nabbar/golib/mail/smtp"
	smtpcfg "github.com/nabbar/golib/mail/smtp/config"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
	libptc "github.com/nabbar/golib/network/protocol"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	globalCtx    context.Context
	globalCancel context.CancelFunc
)

func TestSMTP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SMTP Client Suite")
}

var _ = BeforeSuite(func() {
	globalCtx, globalCancel = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	globalCancel()
})

// getFreePort returns a free TCP port
func getFreePort() int {
	addr, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	lstn, err := net.ListenTCP(libptc.NetworkTCP.Code(), addr)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		_ = lstn.Close()
	}()

	return lstn.Addr().(*net.TCPAddr).Port
}

// getUnusedPort returns a port that is guaranteed to NOT be in use (for error testing)
// It gets a free port, then immediately closes it, ensuring it's not allocated
func getUnusedPort() int {
	port := getFreePort()
	// Small delay to ensure the OS has released the port
	time.Sleep(10 * time.Millisecond)
	return port
}

// Helper functions for creating test configurations
func newTestConfig(host string, port int, tlsMode smtptp.TLSMode) smtpcfg.Config {
	dsn := fmt.Sprintf("tcp(%s:%d)/%s", host, port, tlsMode.String())
	model := smtpcfg.ConfigModel{DSN: dsn}
	cfg, err := model.Config()
	Expect(err).ToNot(HaveOccurred())
	return cfg
}

func newTestConfigWithAuth(host string, port int, tlsMode smtptp.TLSMode, user, pass string) smtpcfg.Config {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pass, host, port, tlsMode.String())
	model := smtpcfg.ConfigModel{DSN: dsn}
	cfg, err := model.Config()
	Expect(err).ToNot(HaveOccurred())
	return cfg
}

func newTestConfigInsecure(host string, port int, tlsMode smtptp.TLSMode) smtpcfg.Config {
	cfg := newTestConfig(host, port, tlsMode)
	cfg.ForceTLSSkipVerify(true)
	return cfg
}

// newTestSMTPClient creates a test SMTP client
func newTestSMTPClient(cfg smtpcfg.Config) smtpcli.SMTP {
	cli, err := smtpcli.New(cfg, cliTLS.TlsConfig(""))
	Expect(err).ToNot(HaveOccurred())
	Expect(cli).ToNot(BeNil())
	return cli
}

// testWriter implements io.WriterTo for test emails
type testWriter struct {
	data string
}

func (w *testWriter) WriteTo(writer io.Writer) (int64, error) {
	n, err := writer.Write([]byte(w.data))
	return int64(n), err
}

func newTestEmail(from, to, subject, body string) *testWriter {
	data := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body)
	return &testWriter{data: data}
}

// contextWithTimeout creates a context with timeout
func contextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
