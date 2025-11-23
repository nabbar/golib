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
 */

package static_test

import (
	"net/http"

	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Suspicious Access Detection", func() {
	Describe("Configuration", func() {
		Context("when setting suspicious detection config", func() {
			It("should store and retrieve configuration", func() {
				handler := newTestStatic().(static.Static)

				cfg := static.SuspiciousConfig{
					Enabled:              true,
					LogSuccessfulAccess:  true,
					SuspiciousPatterns:   []string{".env", "admin"},
					SuspiciousExtensions: []string{".php", ".exe"},
				}

				handler.SetSuspicious(cfg)

				retrieved := handler.GetSuspicious()
				Expect(retrieved.Enabled).To(BeTrue())
				Expect(retrieved.LogSuccessfulAccess).To(BeTrue())
				Expect(retrieved.SuspiciousPatterns).To(ContainElement(".env"))
				Expect(retrieved.SuspiciousExtensions).To(ContainElement(".php"))
			})

			It("should use default config", func() {
				cfg := static.DefaultSuspiciousConfig()

				Expect(cfg.Enabled).To(BeTrue())
				Expect(cfg.LogSuccessfulAccess).To(BeTrue())
				Expect(cfg.SuspiciousPatterns).To(ContainElement(".env"))
				Expect(cfg.SuspiciousPatterns).To(ContainElement(".git"))
				Expect(cfg.SuspiciousExtensions).To(ContainElement(".php"))
			})
		})
	})

	Describe("Suspicious Pattern Detection", func() {
		Context("when accessing suspicious files", func() {
			It("should log access to .env files (blocked)", func() {
				handler := newTestStatic().(static.Static)

				// Enable all security features
				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// This should be logged as suspicious AND blocked
				w := performRequest(engine, "GET", "/static/.env")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should log access to config.php files (404 but logged)", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// This should be logged as suspicious (404)
				w := performRequest(engine, "GET", "/static/config.php")
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})

			It("should log access to admin panels", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// Admin panel access attempts
				_ = performRequest(engine, "GET", "/static/admin/login.php")
				_ = performRequest(engine, "GET", "/static/wp-admin/")
				_ = performRequest(engine, "GET", "/static/phpmyadmin/")
			})

			It("should log backup file access attempts", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// Backup files
				_ = performRequest(engine, "GET", "/static/config.bak")
				_ = performRequest(engine, "GET", "/static/backup.old")
				_ = performRequest(engine, "GET", "/static/index.php.save")
			})
		})
	})

	Describe("Attack Pattern Detection", func() {
		Context("when detecting scanning patterns", func() {
			It("should detect directory traversal scanning", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Directory traversal - should log pattern
				w := performRequest(engine, "GET", "/static/../../../etc/passwd")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should detect path manipulation attempts", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// Double slashes
				_ = performRequest(engine, "GET", "/static//test.txt")
				_ = performRequest(engine, "GET", "/static\\\\test.txt")
			})

			It("should detect config file scanning", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// Config scanning
				_ = performRequest(engine, "GET", "/static/config.php")
				_ = performRequest(engine, "GET", "/static/configuration.inc")
			})
		})
	})

	Describe("Successful Suspicious Access", func() {
		Context("when suspicious file exists and is served", func() {
			It("should log even successful requests when enabled", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.SuspiciousConfig{
					Enabled:              true,
					LogSuccessfulAccess:  true,             // Log even successful access
					SuspiciousExtensions: []string{".txt"}, // Make .txt suspicious for this test
				})

				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// This file exists and will return 200, but should still be logged
				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should not log successful requests when disabled", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.SuspiciousConfig{
					Enabled:              true,
					LogSuccessfulAccess:  false, // Don't log successful access
					SuspiciousExtensions: []string{".txt"},
				})

				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// This won't be logged since LogSuccessfulAccess is false
				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Integration with SecurityBackend Features", func() {
		Context("when combining with path security", func() {
			It("should log and block dangerous combinations", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.DefaultPathSecurityConfig())
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 100,
					Window:      60000000000,
				})

				engine := setupTestRouter(handler, "/static")

				// Combination attack: path traversal + sensitive file
				w := performRequest(engine, "GET", "/static/../.env")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})
		})
	})

	Describe("Real-World Attack Scenarios", func() {
		Context("when simulating real attacks", func() {
			It("should detect WordPress scanning", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// Typical WordPress scanner
				_ = performRequest(engine, "GET", "/static/wp-admin/")
				_ = performRequest(engine, "GET", "/static/wp-login.php")
				_ = performRequest(engine, "GET", "/static/wp-config.php")
				_ = performRequest(engine, "GET", "/static/xmlrpc.php")
			})

			It("should detect database file access attempts", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// Database files
				_ = performRequest(engine, "GET", "/static/database.sql")
				_ = performRequest(engine, "GET", "/static/backup.db")
				_ = performRequest(engine, "GET", "/static/data.sqlite")
			})

			It("should detect source code exposure attempts", func() {
				handler := newTestStatic().(static.Static)

				handler.SetSuspicious(static.DefaultSuspiciousConfig())
				handler.SetPathSecurity(static.PathSecurityConfig{Enabled: false})

				engine := setupTestRouter(handler, "/static")

				// Source code
				_ = performRequest(engine, "GET", "/static/source.tar.gz")
				_ = performRequest(engine, "GET", "/static/backup.zip")
			})
		})
	})
})
