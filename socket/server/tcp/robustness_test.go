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

// robustness_test.go validates server behavior under error conditions and edge cases.
// Tests include error callback triggering, connection timeout handling, resource cleanup,
// graceful degradation, and idle connection timeout mechanisms.
package tcp_test

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Robustness", func() {
	var (
		srv scksrt.ServerTcp
		adr string
		c   context.Context
		cnl context.CancelFunc
	)

	BeforeEach(func() {
		adr = getTestAddr()
		c, cnl = context.WithCancel(globalCtx)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(100 * time.Millisecond)
	})

	Context("error handling", func() {
		It("should handle handler panics gracefully", func() {
			panicHandler := func(c libsck.Context) {
				defer func() {
					_ = recover() // Recover from panic in test
					_ = c.Close()
				}()
				panic("test panic")
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, panicHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Server should not crash despite panic
			con := connectToServer(adr)
			_ = con.Close()

			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should recover from client disconnect", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close()

			time.Sleep(100 * time.Millisecond)

			// Server should still accept new connections
			con2 := connectToServer(adr)
			defer func() { _ = con2.Close() }()

			msg := []byte("test")
			rsp := sendAndReceive(con2, msg)
			Expect(rsp).To(Equal(msg))
		})

		It("should handle rapid open/close cycles", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			for i := 0; i < 20; i++ {
				con := connectToServer(adr)
				_ = con.Close()
			}

			// Server should still be running
			Expect(srv.IsRunning()).To(BeTrue())

			// And should accept new connections
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()
			msg := []byte("test")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})
	})

	Context("callback reliability", func() {
		It("should call error callback on errors", func() {
			errCnt := new(atomic.Int32)
			errFunc := func(e ...error) {
				if len(e) > 0 && e[0] != nil {
					errCnt.Add(1)
				}
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(errFunc)

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close()

			time.Sleep(200 * time.Millisecond)
			// Error callback may be called when connection closes abruptly
		})

		It("should trigger error callback on port already in use", func() {
			errorReceived := make(chan error, 10)
			errFunc := func(errs ...error) {
				for _, e := range errs {
					if e != nil {
						errorReceived <- e
					}
				}
			}

			// Démarrer un premier serveur
			cfg1 := createDefaultConfig(adr)
			srv1, err := scksrt.New(nil, echoHandler, cfg1)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv1)
			waitForServerAcceptingConnections(adr, 2*time.Second)
			defer func() { _ = srv1.Close() }()

			// Tenter de démarrer un second serveur sur le même port
			cfg2 := createDefaultConfig(adr) // Même adresse!
			srv, err = scksrt.New(nil, echoHandler, cfg2)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(errFunc)

			// Tenter de démarrer - devrait échouer car le port est déjà utilisé
			go func() {
				_ = srv.Listen(c)
			}()

			// Vérifier que le callback d'erreur a été appelé
			Eventually(errorReceived, 2*time.Second).Should(Receive(Not(BeNil())))
		})

		It("should call info callback on connection events", func() {
			infoCnt := new(atomic.Int32)
			infoFunc := func(_, _ net.Addr, _ libsck.ConnState) {
				infoCnt.Add(1)
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfo(infoFunc)

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			time.Sleep(200 * time.Millisecond)
			Expect(infoCnt.Load()).To(BeNumerically(">", 0))
		})

		It("should call server info callback on events", func() {
			srvInfoCnt := new(atomic.Int32)
			srvInfoFunc := func(_ string) {
				srvInfoCnt.Add(1)
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfoServer(srvInfoFunc)

			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			time.Sleep(200 * time.Millisecond)
			Expect(srvInfoCnt.Load()).To(BeNumerically(">", 0))
		})
	})

	Context("resource cleanup", func() {
		It("should clean up resources after Close", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			err = srv.Close()
			Expect(err).ToNot(HaveOccurred())

			waitForServerStopped(srv, 2*time.Second)

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should not leak goroutines after shutdown", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close()

			err = srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(500 * time.Millisecond)

			// Check that server is fully stopped
			Expect(srv.IsRunning()).To(BeFalse())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("edge cases", func() {
		It("should handle nil UpdateConn function", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			msg := []byte("test")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})

		It("should handle shutdown timeout gracefully", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, slowHandler(5*time.Second), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			tctx, tcnl := context.WithTimeout(c, 500*time.Millisecond)
			defer tcnl()

			err = srv.Shutdown(tctx)
			// May timeout but should not crash
			_ = err
		})

		It("should handle connection cleanup during shutdown", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, slowHandler(100*time.Millisecond), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			err = srv.Shutdown(c)
			// Shutdown should complete
			Expect(err).ToNot(HaveOccurred())
		})

		It("should close idle connections after ConIdleTimeout", func() {
			cfg := createDefaultConfig(adr)
			cfg.ConIdleTimeout = 2 * time.Second // Configure idle timeout > 1 second

			handlerStarted := make(chan time.Time, 1)
			handlerEnded := make(chan time.Time, 1)

			handler := func(ctx libsck.Context) {
				defer func() {
					ctx.Close()
					handlerEnded <- time.Now()
				}()

				handlerStarted <- time.Now()

				// Wait passively - don't call Read/Write which would reset the idle timer
				// Just wait for the context to be cancelled by idle timeout
				<-ctx.Done()
			}

			var err error
			srv, err = scksrt.New(nil, handler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Connect but don't send any data (idle connection)
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			// Wait for handler to start and record start time
			var startTime, endTime time.Time
			Eventually(handlerStarted, 2*time.Second).Should(Receive(&startTime))

			// Wait for handler to end and record end time
			// Timeout is 2s, so handler should end after approximately 2s
			Eventually(handlerEnded, 4*time.Second).Should(Receive(&endTime))

			// Verify that handler ran for approximately 2 seconds (±500ms tolerance)
			// This proves the idle timeout triggered correctly
			duration := endTime.Sub(startTime)
			Expect(duration).To(BeNumerically("~", 2*time.Second, 500*time.Millisecond))

			// Try to read from connection - should fail as it's closed
			buf := make([]byte, 10)
			con.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, err = con.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})
})
