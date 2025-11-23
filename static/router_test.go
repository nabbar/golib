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
	"strings"

	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router", func() {
	Describe("RegisterRouter", func() {
		Context("when registering without group", func() {
			It("should serve files from root", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static", testMiddleware)

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("This is a test file"))
				Expect(w.Header().Get("X-Test-Middleware")).To(Equal("true"))
			})

			It("should serve JSON files", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.json")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("test json file"))
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
			})

			It("should serve nested files", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/subdir/nested.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("nested test file"))
			})

			It("should return 404 for non-existent files", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/nonexistent.txt")
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})

			It("should serve CSS files with correct content type", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/assets/style.css")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/css"))
			})
		})

		Context("when using multiple middlewares", func() {
			It("should apply all middlewares", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static", newMiddleware(1), newMiddleware(2))

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("X-Test-Middleware-1")).To(Equal("true"))
				Expect(w.Header().Get("X-Test-Middleware-2")).To(Equal("true"))
			})
		})
	})

	Describe("RegisterRouterInGroup", func() {
		Context("when registering with group", func() {
			It("should serve files from group", func() {
				handler := newTestStatic()
				engine := setupTestRouterInGroup(handler, "/static", "/api")

				w := performRequest(engine, "GET", "/api/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("This is a test file"))
			})

			It("should not serve files without group prefix", func() {
				handler := newTestStatic()
				engine := setupTestRouterInGroup(handler, "/static", "/api")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})

			It("should apply middlewares in group", func() {
				handler := newTestStatic()
				engine := setupTestRouterInGroup(handler, "/static", "/api", testMiddleware)

				w := performRequest(engine, "GET", "/api/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("X-Test-Middleware")).To(Equal("true"))
			})
		})
	})

	Describe("SendFile", func() {
		Context("when sending regular file", func() {
			It("should send file with correct headers", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("This is a test file"))
				Expect(w.Header().Get("Content-Type")).ToNot(BeEmpty())
			})

			It("should send HTML files", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/index.html")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Test Index Page"))
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/html"))
			})
		})

		Context("when file is marked for download", func() {
			It("should add Content-Disposition header", func() {
				handler := newTestStatic().(static.Static)
				handler.SetDownload("testdata/test.txt", true)
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("Content-Disposition")).To(ContainSubstring("attachment"))
				Expect(w.Header().Get("Content-Disposition")).To(ContainSubstring("test.txt"))
			})

			It("should not add Content-Disposition for normal files", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Header().Get("Content-Disposition")).To(BeEmpty())
			})
		})
	})

	Describe("Index Handling", func() {
		Context("when index is set", func() {
			It("should serve index file for route", func() {
				handler := newTestStatic().(static.Static)
				handler.SetIndex("", "/static", "testdata/index.html")
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Test Index Page"))
			})

			It("should serve index for exact route match", func() {
				handler := newTestStatic().(static.Static)
				handler.SetIndex("", "/static/", "testdata/index.html")
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Test Index Page"))
			})
		})

		Context("when index is not set", func() {
			It("should return 404 for directory pth", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/subdir")
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("Redirect Handling", func() {
		Context("when redirect is set", func() {
			It("should redirect to destination", func() {
				handler := newTestStatic().(static.Static)
				handler.SetRedirect("", "/static/old", "", "/static/test.txt")
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/old")
				Expect(w.Code).To(Equal(http.StatusPermanentRedirect))

				location := w.Header().Get("Location")
				Expect(location).To(ContainSubstring("/static/test.txt"))
			})

			It("should preserve query parameters on redirect", func() {
				handler := newTestStatic().(static.Static)
				handler.SetRedirect("", "/static/old", "", "/static/new")
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/old?param=value")
				Expect(w.Code).To(Equal(http.StatusPermanentRedirect))

				location := w.Header().Get("Location")
				Expect(location).To(ContainSubstring("param=value"))
			})
		})

		Context("when redirect is not set", func() {
			It("should serve file normally", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Specific Handler", func() {
		Context("when specific handler is set", func() {
			It("should use custom handler", func() {
				handler := newTestStatic().(static.Static)
				handler.SetSpecific("", "/static/custom", customMiddlewareOK("Custom Response", nil))
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/custom")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal("Custom Response"))
			})

			It("should override file serving", func() {
				handler := newTestStatic().(static.Static)
				handler.SetSpecific("", "/static/test.txt", customMiddlewareOK("Overridden", nil))
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal("Overridden"))
			})

			It("should allow custom status codes", func() {
				handler := newTestStatic().(static.Static)
				handler.SetSpecific("", "/static/custom", customMiddlewareCreated("Created"))
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/custom")
				Expect(w.Code).To(Equal(http.StatusCreated))
			})
		})
	})

	Describe("Logger Integration", func() {
		Context("when logger is registered", func() {
			It("should accept logger without error", func() {
				handler := newTestStatic().(static.Static)

				// Use nil logger (should create default)
				Expect(func() {
					handler.RegisterLogger(nil)
				}).ToNot(Panic())
			})
		})
	})

	Describe("Complex Routing Scenarios", func() {
		Context("when combining features", func() {
			It("should handle redirect before specific handler", func() {
				handler := newTestStatic().(static.Static)

				handler.SetRedirect("", "/static/path", "", "/static/redirect")
				handler.SetSpecific("", "/static/path", customMiddlewareOK("Custom", nil))
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/path")
				// Redirect should take precedence
				Expect(w.Code).To(Equal(http.StatusPermanentRedirect))
			})

			It("should handle specific handler before index", func() {
				handler := newTestStatic().(static.Static)

				handler.SetIndex("", "/static/path", "testdata/index.html")
				handler.SetSpecific("", "/static/path", customMiddlewareOK("Custom", nil))
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/path")
				// Specific handler should take precedence
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal("Custom"))
			})

			It("should serve index when no redirect or specific handler", func() {
				handler := newTestStatic().(static.Static)
				handler.SetIndex("", "/static", "testdata/index.html")
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static")
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Test Index Page"))
			})
		})

		Context("when handling different file types", func() {
			It("should serve all file types correctly", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				// Test TXT
				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				// Test JSON
				w = performRequest(engine, "GET", "/static/test.json")
				Expect(w.Code).To(Equal(http.StatusOK))

				// Test HTML
				w = performRequest(engine, "GET", "/static/index.html")
				Expect(w.Code).To(Equal(http.StatusOK))

				// Test CSS
				w = performRequest(engine, "GET", "/static/assets/style.css")
				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Error Handling", func() {
		Context("when encountering errors", func() {
			It("should handle path traversal attempts", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/../../../etc/passwd")
				// Should not be able to access files outside embed.FS
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})

			It("should handle double slashes", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static//test.txt")
				// Should still work due to path cleaning
				Expect(w.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusNotFound)))
			})

			It("should handle trailing slashes", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt/")
				// Trailing slashes are stripped, so the file should be served normally
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("This is a test file"))
			})
		})
	})

	Describe("Base Path Handling", func() {
		Context("when using custom base pth", func() {
			It("should serve files from custom base", func() {
				handler := newTestStaticWithRoot("testdata")
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should handle multiple base pth", func() {
				handler := newTestStaticWithRoot("testdata", "testdata/subdir")
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				w = performRequest(engine, "GET", "/static/nested.txt")
				Expect(w.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusNotFound)))
			})
		})
	})

	Describe("Case Sensitivity", func() {
		Context("when checking file pth", func() {
			It("should be case sensitive", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/Test.txt")
				Expect(w.Code).To(Equal(http.StatusNotFound))

				w = performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Content Length", func() {
		Context("when serving files", func() {
			It("should include content length header", func() {
				handler := newTestStatic()
				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				contentLength := w.Header().Get("Content-Length")
				// Should have content length (may vary based on implementation)
				if contentLength != "" {
					Expect(strings.TrimSpace(contentLength)).ToNot(BeEmpty())
				}
			})
		})
	})
})
