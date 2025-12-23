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
 *
 */

package httpserver_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/nabbar/golib/httpserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-MON] Server Monitoring", func() {
	var (
		srv      Server
		err      error
		testPort string
	)

	BeforeEach(func() {
		testPort = fmt.Sprintf("127.0.0.1:%d", GetFreePort())
	})

	AfterEach(func() {
		if srv != nil {
			srv.Stop(context.Background())
		}
	})

	Describe("Server State", func() {
		It("[TC-MON-014] should not be running before start", func() {
			cfg := Config{
				Name:   "state-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NewServeMux()}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			// Server should not be running before start
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("[TC-MON-015] should be running after start", func() {
			cfg := Config{
				Name:   "running-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NewServeMux()}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			err = srv.Start(context.Background())
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			// Server should be running after start
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Monitor Name", func() {
		It("[TC-MON-016] should return a valid monitor name", func() {
			cfg := Config{
				Name:   "monitor-name-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NewServeMux()}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			name := srv.MonitorName()
			Expect(name).To(ContainSubstring("HTTP Server"))
			Expect(name).To(ContainSubstring(testPort))
		})
	})

	Describe("Server Uptime", func() {
		It("[TC-MON-017] should track uptime correctly", func() {
			cfg := Config{
				Name:   "uptime-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
				return map[string]http.Handler{"": mux}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			err = srv.Start(context.Background())
			Expect(err).NotTo(HaveOccurred())

			// Wait a bit and check uptime increases
			time.Sleep(100 * time.Millisecond)
			uptime1 := srv.Uptime()
			Expect(uptime1).To(BeNumerically(">", 0))

			time.Sleep(50 * time.Millisecond)
			uptime2 := srv.Uptime()
			Expect(uptime2).To(BeNumerically(">", uptime1))
		})
	})

	Describe("Server Configuration", func() {
		It("[TC-MON-018] should allow configuration updates when stopped", func() {
			cfg := Config{
				Name:   "config-update-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NewServeMux()}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			// Verify initial config
			initialCfg := srv.GetConfig()
			Expect(initialCfg).NotTo(BeNil())
			Expect(initialCfg.Name).To(Equal("config-update-test"))
		})
	})
})
