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

package static

import (
	"crypto/sha256"
	"fmt"
	"mime"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	loglvl "github.com/nabbar/golib/logger/level"
)

// SetHeaders configures HTTP headers behavior.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) SetHeaders(cfg HeadersConfig) {
	s.hdr.Store(&cfg)
}

// GetHeaders returns the current HTTP headers configuration.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) GetHeaders() HeadersConfig {
	if cfg := s.hdr.Load(); cfg != nil {
		return *cfg
	}
	return HeadersConfig{}
}

// getMimeType detects the MIME type of a file based on its extension.
// Custom MIME types are checked first, then falls back to standard detection.
func (s *staticHandler) getMimeType(filename string) string {
	cfg := s.GetHeaders()

	ext := strings.ToLower(path.Ext(filename))

	// Check custom MIME types first
	if cfg.CustomMimeTypes != nil {
		if customMime, ok := cfg.CustomMimeTypes[ext]; ok {
			return customMime
		}
	}

	// Use standard detection
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return mimeType
}

// isMimeTypeAllowed checks if a MIME type is allowed to be served.
// This method:
//  1. Checks deny list first
//  2. If allow list is empty, allows all (except denied)
//  3. Otherwise, checks allow list
func (s *staticHandler) isMimeTypeAllowed(mimeType string) bool {
	cfg := s.GetHeaders()

	if !cfg.EnableContentType {
		return true
	}

	// Extract base type (without charset, etc.)
	baseType := strings.Split(mimeType, ";")[0]
	baseType = strings.TrimSpace(baseType)

	// Check if in deny list
	if slices.Contains(cfg.DenyMimeTypes, baseType) {
		return false
	}

	// If allow list is empty, everything is allowed
	if len(cfg.AllowedMimeTypes) == 0 {
		return true
	}

	// Check if in allow list
	return slices.Contains(cfg.AllowedMimeTypes, baseType)
}

// generateETag generates an ETag for a file based on:
//   - Filename
//   - File size
//   - Modification time
//
// The ETag is a SHA-256 hash truncated to 16 bytes for efficiency.
func (s *staticHandler) generateETag(filename string, size int64, modTime time.Time) string {
	cfg := s.GetHeaders()

	if !cfg.EnableETag {
		return ""
	}

	// ETag based on: filename + size + modification date
	// Format: "hash-hex"
	data := fmt.Sprintf("%s-%d-%d", filename, size, modTime.Unix())
	hash := sha256.Sum256([]byte(data))
	etag := fmt.Sprintf(`"%x"`, hash[:16]) // Take first 16 bytes

	return etag
}

// checkETag verifies if the client's ETag matches the current file.
// This enables HTTP 304 Not Modified responses to save bandwidth.
func (s *staticHandler) checkETag(c *ginsdk.Context, etag string) bool {
	cfg := s.GetHeaders()

	if !cfg.EnableETag || etag == "" {
		return false
	}

	// Check If-None-Match header
	ifNoneMatch := c.GetHeader("If-None-Match")
	if ifNoneMatch == "" {
		return false
	}

	// Compare ETags
	return ifNoneMatch == etag
}

// setCacheHeaders sets HTTP caching headers (Cache-Control, Expires).
// These headers instruct browsers and CDNs how to cache the file.
func (s *staticHandler) setCacheHeaders(c *ginsdk.Context) {
	cfg := s.GetHeaders()

	if !cfg.EnableCacheControl {
		return
	}

	// Cache-Control
	var cacheControl string
	if cfg.CachePublic {
		cacheControl = fmt.Sprintf("public, max-age=%d", cfg.CacheMaxAge)
	} else {
		cacheControl = fmt.Sprintf("private, max-age=%d", cfg.CacheMaxAge)
	}

	c.Header("Cache-Control", cacheControl)

	// Expires (for HTTP/1.0 compatibility)
	expires := time.Now().Add(time.Duration(cfg.CacheMaxAge) * time.Second)
	c.Header("Expires", expires.UTC().Format(http.TimeFormat))
}

// setContentTypeHeader sets the Content-Type header and validates MIME type.
// Returns an error if the MIME type is not allowed.
// Notifies security backend if MIME type is denied.
func (s *staticHandler) setContentTypeHeader(c *ginsdk.Context, filename string) (string, error) {
	cfg := s.GetHeaders()
	mimeType := s.getMimeType(filename)

	// Validate MIME type only if EnableContentType is enabled
	if cfg.EnableContentType && !s.isMimeTypeAllowed(mimeType) {
		ent := s.getLogger().Entry(loglvl.WarnLevel, "mime type not allowed")
		ent.FieldAdd("filename", filename)
		ent.FieldAdd("mimeType", mimeType)
		ent.Log()

		// Notify WAF/IDS/EDR
		details := map[string]string{
			"filename":  filename,
			"mime_type": mimeType,
		}
		event := s.newSecuEvt(c, EventTypeMimeTypeDenied, "medium", true, details)
		s.notifySecuEvt(event)

		return "", ErrorMimeTypeDenied.Error(fmt.Errorf("mime type not allowed: %s", mimeType))
	}

	c.Header("Content-Type", mimeType)
	return mimeType, nil
}

// setETagHeader sets ETag and Last-Modified headers and checks cache validation.
// Returns true if the client has a valid cached version (cache hit).
// This allows the handler to return HTTP 304 Not Modified.
func (s *staticHandler) setETagHeader(c *ginsdk.Context, filename string, size int64, modTime time.Time) bool {
	etag := s.generateETag(filename, size, modTime)

	if etag == "" {
		return false
	}

	c.Header("ETag", etag)
	c.Header("Last-Modified", modTime.UTC().Format(http.TimeFormat))

	// Check if client already has cached version
	if s.checkETag(c, etag) {
		return true // Cache hit
	}

	return false // Cache miss
}
