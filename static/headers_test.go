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
	"net/http/httptest"

	ginsdk "github.com/gin-gonic/gin"
	"github.com/nabbar/golib/static"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP Headers", func() {
	Describe("Configuration", func() {
		Context("when setting headers config", func() {
			It("should store and retrieve configuration", func() {
				handler := newTestStatic().(static.Static)

				cfg := static.HeadersConfig{
					EnableCacheControl: true,
					CacheMaxAge:        7200,
					CachePublic:        false,
					EnableETag:         true,
					EnableContentType:  true,
					AllowedMimeTypes:   []string{"text/plain", "text/html"},
					DenyMimeTypes:      []string{"application/x-executable"},
					CustomMimeTypes: map[string]string{
						".custom": "application/x-custom",
					},
				}

				handler.SetHeaders(cfg)

				retrieved := handler.GetHeaders()
				Expect(retrieved.EnableCacheControl).To(BeTrue())
				Expect(retrieved.CacheMaxAge).To(Equal(7200))
				Expect(retrieved.CachePublic).To(BeFalse())
				Expect(retrieved.EnableETag).To(BeTrue())
				Expect(retrieved.AllowedMimeTypes).To(ContainElement("text/plain"))
			})

			It("should use default config", func() {
				cfg := static.DefaultHeadersConfig()

				Expect(cfg.EnableCacheControl).To(BeTrue())
				Expect(cfg.CacheMaxAge).To(Equal(3600))
				Expect(cfg.CachePublic).To(BeTrue())
				Expect(cfg.EnableETag).To(BeTrue())
				Expect(cfg.DenyMimeTypes).To(ContainElement("application/x-executable"))
			})
		})
	})

	Describe("Cache-Control Headers", func() {
		Context("when cache control is enabled", func() {
			It("should add Cache-Control header for public cache", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableCacheControl: true,
					CacheMaxAge:        3600,
					CachePublic:        true,
					EnableContentType:  false,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				cacheControl := w.Header().Get("Cache-Control")
				Expect(cacheControl).To(ContainSubstring("public"))
				Expect(cacheControl).To(ContainSubstring("max-age=3600"))
			})

			It("should add Cache-Control header for private cache", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableCacheControl: true,
					CacheMaxAge:        1800,
					CachePublic:        false,
					EnableContentType:  false,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				cacheControl := w.Header().Get("Cache-Control")
				Expect(cacheControl).To(ContainSubstring("private"))
				Expect(cacheControl).To(ContainSubstring("max-age=1800"))
			})

			It("should add Expires header", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableCacheControl: true,
					CacheMaxAge:        3600,
					EnableContentType:  false,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				expires := w.Header().Get("Expires")
				Expect(expires).NotTo(BeEmpty())
			})

			It("should not add cache headers when disabled", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableCacheControl: false,
					EnableContentType:  false,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				cacheControl := w.Header().Get("Cache-Control")
				Expect(cacheControl).To(BeEmpty())
			})
		})
	})

	Describe("ETag Support", func() {
		Context("when ETag is enabled", func() {
			It("should add ETag and Last-Modified headers", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableETag:        true,
					EnableContentType: false,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				etag := w.Header().Get("ETag")
				Expect(etag).NotTo(BeEmpty())
				Expect(etag).To(ContainSubstring(`"`)) // ETags are quoted

				lastModified := w.Header().Get("Last-Modified")
				Expect(lastModified).NotTo(BeEmpty())
			})

			It("should return 304 Not Modified when ETag matches", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableETag:        true,
					EnableContentType: false,
				})

				engine := setupTestRouter(handler, "/static")

				// First call to get the ETag
				w1 := performRequest(engine, "GET", "/static/test.txt")
				Expect(w1.Code).To(Equal(http.StatusOK))

				etag := w1.Header().Get("ETag")
				Expect(etag).NotTo(BeEmpty())

				// Second call with If-None-Match
				req := performRequestWithHeaders(engine, "GET", "/static/test.txt", map[string]string{
					"If-None-Match": etag,
				})

				Expect(req.Code).To(Equal(http.StatusNotModified))
				Expect(req.Body.Len()).To(Equal(0)) // Body should be empty for 304
			})

			It("should return 200 when ETag does not match", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableETag:        true,
					EnableContentType: false,
				})

				engine := setupTestRouter(handler, "/static")

				// Appel avec un ETag qui ne correspond pas
				w := performRequestWithHeaders(engine, "GET", "/static/test.txt", map[string]string{
					"If-None-Match": `"wrong-etag"`,
				})

				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should not add ETag when disabled", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableETag:        false,
					EnableContentType: false,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				etag := w.Header().Get("ETag")
				Expect(etag).To(BeEmpty())
			})
		})
	})

	Describe("Content-Type Validation", func() {
		Context("when content type validation is enabled", func() {
			It("should set correct Content-Type header", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableContentType: true,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(ContainSubstring("text/plain"))
			})

			It("should use custom MIME types", func() {
				handler := newTestStaticWithRoot("testdata").(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableContentType: true,
					CustomMimeTypes: map[string]string{
						".txt": "text/x-custom",
					},
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(Equal("text/x-custom"))
			})

			It("should block denied MIME types", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableContentType: true,
					DenyMimeTypes:     []string{"text/plain"},
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should allow only whitelisted MIME types", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableContentType: true,
					AllowedMimeTypes:  []string{"image/png", "image/jpeg"},
				})

				engine := setupTestRouter(handler, "/static")

				// text/plain n'est pas dans la whitelist
				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			It("should allow all MIME types when allow list is empty", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableContentType: true,
					AllowedMimeTypes:  []string{}, // Empty = all allowed
					DenyMimeTypes:     []string{},
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Combined Features", func() {
		Context("when using all features together", func() {
			It("should apply cache, ETag, and content-type together", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableCacheControl: true,
					CacheMaxAge:        3600,
					CachePublic:        true,
					EnableETag:         true,
					EnableContentType:  true,
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusOK))

				// Check all headers
				Expect(w.Header().Get("Cache-Control")).To(ContainSubstring("public"))
				Expect(w.Header().Get("ETag")).NotTo(BeEmpty())
				Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/plain"))
				Expect(w.Header().Get("Expires")).NotTo(BeEmpty())
				Expect(w.Header().Get("Last-Modified")).NotTo(BeEmpty())
			})

			It("should respect MIME type restrictions with caching", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.HeadersConfig{
					EnableCacheControl: true,
					EnableContentType:  true,
					DenyMimeTypes:      []string{"text/plain"},
				})

				engine := setupTestRouter(handler, "/static")

				w := performRequest(engine, "GET", "/static/test.txt")
				Expect(w.Code).To(Equal(http.StatusForbidden))

				// No cache headers for blocked requests
				Expect(w.Header().Get("Cache-Control")).To(BeEmpty())
			})
		})
	})

	Describe("Performance", func() {
		Context("when benchmarking header operations", func() {
			It("should handle ETag generation efficiently", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.DefaultHeadersConfig())

				engine := setupTestRouter(handler, "/static")

				// Multiple requests should get consistent ETags
				etags := make([]string, 10)
				for i := 0; i < 10; i++ {
					w := performRequest(engine, "GET", "/static/test.txt")
					Expect(w.Code).To(Equal(http.StatusOK))
					etags[i] = w.Header().Get("ETag")
				}

				// All ETags should be identical for same file
				for i := 1; i < len(etags); i++ {
					Expect(etags[i]).To(Equal(etags[0]))
				}
			})

			It("should minimize bandwidth with 304 responses", func() {
				handler := newTestStatic().(static.Static)

				handler.SetHeaders(static.DefaultHeadersConfig())

				engine := setupTestRouter(handler, "/static")

				// First call - full response
				w1 := performRequest(engine, "GET", "/static/test.txt")
				Expect(w1.Code).To(Equal(http.StatusOK))
				fullSize := w1.Body.Len()

				etag := w1.Header().Get("ETag")

				// Second call with ETag - 304 response
				w2 := performRequestWithHeaders(engine, "GET", "/static/test.txt", map[string]string{
					"If-None-Match": etag,
				})
				Expect(w2.Code).To(Equal(http.StatusNotModified))
				cachedSize := w2.Body.Len()

				// 304 response should have no body
				Expect(cachedSize).To(Equal(0))
				Expect(fullSize).To(BeNumerically(">", 0))
			})
		})
	})
})

// Helper function for requests with custom headers
func performRequestWithHeaders(engine interface{}, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	engine.(*ginsdk.Engine).ServeHTTP(w, req)
	return w
}
