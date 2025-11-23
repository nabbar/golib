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

var _ = Describe("Path SecurityBackend", func() {
	Describe("Configuration", func() {
		Context("when setting path security config", func() {
			It("should store and retrieve configuration", func() {
				handler := newTestStatic().(static.Static)

				cfg := static.PathSecurityConfig{
					Enabled:         true,
					AllowDotFiles:   false,
					MaxPathDepth:    5,
					BlockedPatterns: []string{".git", "admin"},
				}

				handler.SetPathSecurity(cfg)

				retrieved := handler.GetPathSecurity()
				Expect(retrieved.Enabled).To(BeTrue())
				Expect(retrieved.AllowDotFiles).To(BeFalse())
				Expect(retrieved.MaxPathDepth).To(Equal(5))
				Expect(retrieved.BlockedPatterns).To(ContainElement(".git"))
				Expect(retrieved.BlockedPatterns).To(ContainElement("admin"))
			})

			It("should use default config", func() {
				cfg := static.DefaultPathSecurityConfig()

				Expect(cfg.Enabled).To(BeTrue())
				Expect(cfg.AllowDotFiles).To(BeFalse())
				Expect(cfg.MaxPathDepth).To(Equal(10))
				Expect(cfg.BlockedPatterns).To(ContainElement(".git"))
				Expect(cfg.BlockedPatterns).To(ContainElement(".env"))
			})
		})
	})

	Describe("Path Traversal Protection", func() {
		Context("when path security is disabled", func() {
			It("should allow all paths", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.PathSecurityConfig{
					Enabled: false,
				})

				// All paths should be considered safe when disabled
				Expect(handler.IsPathSafe("/static/../../../etc/passwd")).To(BeTrue())
				Expect(handler.IsPathSafe("/static/.git/config")).To(BeTrue())
			})
		})

		Context("when path security is enabled", func() {
			It("should block path traversal with ..", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())
				engine := setupTestRouter(handler, "/static")

				// Classic path traversal attempt
				w := performRequest(engine, "GET", "/static/../../../etc/passwd")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should block encoded path traversal", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())
				engine := setupTestRouter(handler, "/static")

				// URL encoded path traversal
				w := performRequest(engine, "GET", "/static/..%2F..%2Fetc/passwd")
				// Gin decodes this, so it should be blocked
				Expect(w.Code).To(Or(Equal(http.StatusForbidden), Equal(http.StatusNotFound)))
			})

			It("should block paths with null bytes", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				// Null byte injection attempt
				Expect(handler.IsPathSafe("/static/test.txt\x00.exe")).To(BeFalse())
			})

			It("should block dot files by default", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())
				engine := setupTestRouter(handler, "/static")

				// Dot file access
				w := performRequest(engine, "GET", "/static/.env")
				Expect(w.Code).To(Equal(http.StatusForbidden))

				w = performRequest(engine, "GET", "/static/.git/config")
				Expect(w.Code).To(Equal(http.StatusForbidden))

				w = performRequest(engine, "GET", "/static/subdir/.hidden")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should allow dot files when configured", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.PathSecurityConfig{
					Enabled:       true,
					AllowDotFiles: true,
					MaxPathDepth:  10,
				})

				// Dot files should be allowed (but may 404 if they don't exist)
				Expect(handler.IsPathSafe("/static/.env")).To(BeTrue())
				Expect(handler.IsPathSafe("/static/.git/config")).To(BeTrue())
			})

			It("should enforce max path depth", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.PathSecurityConfig{
					Enabled:      true,
					MaxPathDepth: 3,
				})

				// Shallow paths should pass
				Expect(handler.IsPathSafe("/a/b/c")).To(BeTrue())

				// Deep paths should be blocked
				Expect(handler.IsPathSafe("/a/b/c/d/e")).To(BeFalse())
			})

			It("should block configured patterns", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.PathSecurityConfig{
					Enabled:         true,
					BlockedPatterns: []string{"admin", "wp-admin", ".git"},
				})
				engine := setupTestRouter(handler, "/static")

				// Blocked patterns
				w := performRequest(engine, "GET", "/static/admin/config.php")
				Expect(w.Code).To(Equal(http.StatusForbidden))

				w = performRequest(engine, "GET", "/static/wp-admin/index.php")
				Expect(w.Code).To(Equal(http.StatusForbidden))

				w = performRequest(engine, "GET", "/static/.git/config")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should allow normal paths", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				// Normal paths should pass validation
				Expect(handler.IsPathSafe("/static/test.txt")).To(BeTrue())
				Expect(handler.IsPathSafe("/static/subdir/file.css")).To(BeTrue())
				Expect(handler.IsPathSafe("/static/assets/img/logo.png")).To(BeTrue())
			})
		})
	})

	Describe("Edge Cases", func() {
		Context("when handling special characters", func() {
			It("should handle double slashes", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				// Double slashes are cleaned by path.Clean()
				Expect(handler.IsPathSafe("/static//test.txt")).To(BeTrue())
				Expect(handler.IsPathSafe("/static///subdir//file.txt")).To(BeTrue())
			})

			It("should handle trailing slashes", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				// Trailing slashes should be OK
				Expect(handler.IsPathSafe("/static/test.txt/")).To(BeTrue())
				Expect(handler.IsPathSafe("/static/subdir/")).To(BeTrue())
			})

			It("should handle unicode characters", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				// Unicode should be allowed (if the file exists)
				Expect(handler.IsPathSafe("/static/regular-file.txt")).To(BeTrue())
				Expect(handler.IsPathSafe("/static/文件.txt")).To(BeTrue())
			})
		})
	})

	Describe("Real-World Attack Vectors", func() {
		Context("when testing known attack patterns", func() {
			It("should block Windows path traversal", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				// Windows-style path traversal
				Expect(handler.IsPathSafe("/static/..\\..\\windows\\system32")).To(BeFalse())
			})

			It("should block absolute path attempts", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())
				engine := setupTestRouter(handler, "/static")

				// Absolute path attempts (will be cleaned but may fail other checks)
				w := performRequest(engine, "GET", "/../../../etc/passwd")
				Expect(w.Code).NotTo(Equal(http.StatusOK))
			})

			It("should block common sensitive files", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.PathSecurityConfig{
					Enabled: true,
					BlockedPatterns: []string{
						".env", ".git", ".svn", ".htaccess",
						"config.php", "wp-config.php",
					},
				})
				engine := setupTestRouter(handler, "/static")

				sensitiveFiles := []string{
					"/static/.env",
					"/static/.git/HEAD",
					"/static/.svn/entries",
					"/static/.htaccess",
					"/static/config.php",
					"/static/wp-config.php",
				}

				for _, path := range sensitiveFiles {
					w := performRequest(engine, "GET", path)
					Expect(w.Code).To(Equal(http.StatusForbidden),
						"Path %s should be blocked", path)
				}
			})

			It("should handle mixed case pattern matching", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.PathSecurityConfig{
					Enabled:         true,
					AllowDotFiles:   true, // Allow dot files to test pattern matching only
					BlockedPatterns: []string{".git"},
				})

				// Patterns are case-sensitive
				Expect(handler.IsPathSafe("/static/.GIT/config")).To(BeTrue())
				Expect(handler.IsPathSafe("/static/.git/config")).To(BeFalse())
			})
		})
	})

	Describe("Integration with Route Handling", func() {
		Context("when combined with file serving", func() {
			It("should serve valid files and block invalid paths", func() {
				handler := newTestStatic().(static.Static)

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())
				engine := setupTestRouter(handler, "/static")

				// Valid file should work (200 or 404, but not 403)
				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).NotTo(Equal(http.StatusForbidden))

				// Path traversal should be blocked with 403
				w = performRequest(engine, "GET", "/static/../etc/passwd")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should work with rate limiting", func() {
				handler := newTestStatic().(static.Static)

				// Configure both rate limit and path security
				handler.SetRateLimit(static.RateLimitConfig{
					Enabled:     true,
					MaxRequests: 10,
					Window:      60000000000, // 1 minute in nanoseconds
				})

				handler.SetPathSecurity(static.DefaultPathSecurityConfig())

				engine := setupTestRouter(handler, "/static")

				// Normal request should pass both checks
				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).NotTo(Equal(http.StatusForbidden))
				Expect(w.Code).NotTo(Equal(http.StatusTooManyRequests))

				// Path traversal should be blocked before rate limit check
				w = performRequest(engine, "GET", "/static/../passwd")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})
		})
	})
})
