/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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
 */

package queuer_test

import (
	"context"
	"time"

	"github.com/nabbar/golib/mail/queuer"
	libver "github.com/nabbar/golib/version"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Monitor Operations", func() {
		Context("with valid SMTP client", func() {
			It("should create monitor successfully", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				ver := libver.NewVersion(libver.License_MIT, "queuer", "Queuer", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)
				mon, err := pooler.Monitor(ctx, ver)
				Expect(err).ToNot(HaveOccurred())
				Expect(mon).ToNot(BeNil())
			})
		})

		Context("with nil SMTP client", func() {
			It("should return error", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)

				ver := libver.NewVersion(libver.License_MIT, "queuer", "Queuer", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)
				mon, err := pooler.Monitor(ctx, ver)
				Expect(err).To(HaveOccurred())
				Expect(mon).To(BeNil())
			})
		})
	})
})
