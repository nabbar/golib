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
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/nabbar/golib/static"
)

//go:embed testdata
var exampleContent embed.FS

// Example_basic shows the simplest usage of the static package.
// This example serves files from an embedded filesystem with no security features.
func Example_basic() {
	// Create a static file handler
	_ = static.New(context.Background(), exampleContent, "testdata")

	// The handler can now be registered with Gin router
	// router := gin.Default()
	// handler.RegisterRouter("/static", router.GET)
	// router.Run(":8080")

	// Files will be accessible at http://localhost:8080/static/*
	fmt.Println("Basic static file handler created")
	// Output: Basic static file handler created
}

// Example_pathSecurity demonstrates path traversal protection.
func Example_pathSecurity() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Enable path security with default settings
	handler.SetPathSecurity(static.DefaultPathSecurityConfig())

	// Or customize the configuration
	handler.SetPathSecurity(static.PathSecurityConfig{
		Enabled:       true,
		AllowDotFiles: false, // Block .env, .git, etc.
		MaxPathDepth:  10,
		BlockedPatterns: []string{
			".git",
			".svn",
			"node_modules",
		},
	})

	// Check if a path is safe
	safe := handler.IsPathSafe("/static/test.txt")
	fmt.Printf("Path is safe: %v\n", safe)
	// Output: Path is safe: true
}

// Example_rateLimit demonstrates IP-based rate limiting.
func Example_rateLimit() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Configure rate limiting
	handler.SetRateLimit(static.RateLimitConfig{
		Enabled:         true,
		MaxRequests:     100,             // Max 100 unique files
		Window:          time.Minute,     // Per minute
		CleanupInterval: 5 * time.Minute, // Cleanup every 5 minutes
		WhitelistIPs: []string{
			"127.0.0.1", // Localhost
			"::1",       // IPv6 localhost
		},
	})

	// Check if an IP is rate limited
	limited := handler.IsRateLimited("192.168.1.100")
	fmt.Printf("IP is rate limited: %v\n", limited)
	// Output: IP is rate limited: false
}

// Example_httpCaching demonstrates HTTP caching with ETag and Cache-Control.
func Example_httpCaching() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Configure HTTP headers with default settings
	handler.SetHeaders(static.DefaultHeadersConfig())

	// Or customize
	handler.SetHeaders(static.HeadersConfig{
		EnableCacheControl: true,
		CacheMaxAge:        3600, // 1 hour
		CachePublic:        true, // Allow CDN caching
		EnableETag:         true, // Enable ETag validation
		EnableContentType:  true,
		CustomMimeTypes: map[string]string{
			".wasm": "application/wasm",
		},
	})

	fmt.Println("HTTP caching configured")
	// Output: HTTP caching configured
}

// Example_mimeTypeValidation demonstrates MIME type filtering.
func Example_mimeTypeValidation() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Block dangerous file types
	handler.SetHeaders(static.HeadersConfig{
		EnableContentType: true,
		DenyMimeTypes: []string{
			"application/x-executable",
			"application/x-msdownload",
			"application/x-sh",
		},
	})

	// Or whitelist only specific types
	handler.SetHeaders(static.HeadersConfig{
		EnableContentType: true,
		AllowedMimeTypes: []string{
			"text/html",
			"text/css",
			"application/javascript",
			"image/png",
			"image/jpeg",
		},
	})

	fmt.Println("MIME type validation configured")
	// Output: MIME type validation configured
}

// Example_suspiciousDetection demonstrates suspicious access pattern detection.
func Example_suspiciousDetection() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Enable suspicious access detection
	handler.SetSuspicious(static.DefaultSuspiciousConfig())

	// Or customize
	handler.SetSuspicious(static.SuspiciousConfig{
		Enabled:             true,
		LogSuccessfulAccess: true, // Log even successful suspicious requests
		SuspiciousPatterns: []string{
			".env",
			".git",
			"wp-admin",
			"phpmyadmin",
		},
		SuspiciousExtensions: []string{
			".php",
			".exe",
		},
	})

	fmt.Println("Suspicious access detection enabled")
	// Output: Suspicious access detection enabled
}

// Example_securityBackend demonstrates WAF/IDS/EDR integration.
func Example_securityBackend() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Configure security backend with webhook
	handler.SetSecurityBackend(static.SecurityConfig{
		Enabled:    true,
		WebhookURL: "https://waf.example.com/events",
		WebhookHeaders: map[string]string{
			"Authorization": "Bearer secret-token",
		},
		WebhookTimeout: 5 * time.Second,
		WebhookAsync:   true,     // Non-blocking
		MinSeverity:    "medium", // Only medium, high, critical
	})

	fmt.Println("Security backend configured")
	// Output: Security backend configured
}

// Example_securityBackendBatch demonstrates batch event processing.
func Example_securityBackendBatch() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Configure batch processing for efficiency
	handler.SetSecurityBackend(static.SecurityConfig{
		Enabled:      true,
		WebhookURL:   "https://siem.example.com/batch",
		BatchSize:    100,              // Send every 100 events
		BatchTimeout: 30 * time.Second, // Or every 30 seconds
		MinSeverity:  "low",            // All severity levels
	})

	fmt.Println("Batch security backend configured")
	// Output: Batch security backend configured
}

// Example_securityBackendCEF demonstrates CEF format for SIEM systems.
func Example_securityBackendCEF() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Configure CEF format for SIEM compatibility
	handler.SetSecurityBackend(static.SecurityConfig{
		Enabled:         true,
		WebhookURL:      "https://siem.example.com/cef",
		EnableCEFFormat: true, // Common Event Format
		MinSeverity:     "high",
	})

	fmt.Println("CEF format configured")
	// Output: CEF format configured
}

// Example_indexFiles demonstrates index file configuration.
func Example_indexFiles() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Set index file for root
	handler.SetIndex("", "/", "index.html")

	// Set index for specific routes
	handler.SetIndex("", "/docs", "docs/index.html")

	fmt.Println("Index files configured")
	// Output: Index files configured
}

// Example_downloadFiles demonstrates download configuration.
func Example_downloadFiles() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Mark files to be downloaded instead of displayed
	handler.SetDownload("/static/document.pdf", true)
	handler.SetDownload("/static/archive.zip", true)

	// Check if a file should be downloaded
	shouldDownload := handler.IsDownload("/static/document.pdf")
	fmt.Printf("Should download: %v\n", shouldDownload)
	// Output: Should download: false
}

// Example_redirects demonstrates URL redirection.
func Example_redirects() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Configure redirects (HTTP 301)
	handler.SetRedirect("", "/old-path", "", "/new-path")
	handler.SetRedirect("", "/legacy", "", "/modern")

	fmt.Println("Redirects configured")
	// Output: Redirects configured
}

// Example_production demonstrates a complete production-ready configuration.
func Example_production() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// 1. Path Security
	handler.SetPathSecurity(static.PathSecurityConfig{
		Enabled:       true,
		AllowDotFiles: false,
		MaxPathDepth:  10,
		BlockedPatterns: []string{
			".git", ".svn", ".env",
			"node_modules", "vendor",
		},
	})

	// 2. Rate Limiting
	handler.SetRateLimit(static.RateLimitConfig{
		Enabled:         true,
		MaxRequests:     1000,
		Window:          time.Minute,
		CleanupInterval: 5 * time.Minute,
		WhitelistIPs:    []string{"127.0.0.1"},
	})

	// 3. HTTP Caching
	handler.SetHeaders(static.HeadersConfig{
		EnableCacheControl: true,
		CacheMaxAge:        3600,
		CachePublic:        true,
		EnableETag:         true,
		EnableContentType:  true,
		DenyMimeTypes: []string{
			"application/x-executable",
		},
	})

	// 4. Suspicious Access Detection
	handler.SetSuspicious(static.SuspiciousConfig{
		Enabled:             true,
		LogSuccessfulAccess: true,
		SuspiciousPatterns:  []string{".env", ".git"},
	})

	// 5. Security Backend
	handler.SetSecurityBackend(static.SecurityConfig{
		Enabled:      true,
		WebhookURL:   "https://waf.example.com/events",
		WebhookAsync: true,
		MinSeverity:  "medium",
		BatchSize:    100,
		BatchTimeout: 30 * time.Second,
	})

	// 6. Index Files
	handler.SetIndex("", "/", "index.html")

	fmt.Println("Production configuration complete")
	// Output: Production configuration complete
}

// Example_development demonstrates a minimal development configuration.
func Example_development() {
	// Minimal setup for local development
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Only enable basic security
	handler.SetPathSecurity(static.PathSecurityConfig{
		Enabled:       true,
		AllowDotFiles: false,
	})

	fmt.Println("Development configuration complete")
	// Output: Development configuration complete
}

// Example_cdn demonstrates configuration optimized for CDN usage.
func Example_cdn() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Aggressive caching for CDN
	handler.SetHeaders(static.HeadersConfig{
		EnableCacheControl: true,
		CacheMaxAge:        31536000, // 1 year
		CachePublic:        true,     // Allow CDN caching
		EnableETag:         true,
	})

	// Relaxed rate limiting (CDN handles most requests)
	handler.SetRateLimit(static.RateLimitConfig{
		Enabled:     true,
		MaxRequests: 10000,
		Window:      time.Minute,
	})

	fmt.Println("CDN configuration complete")
	// Output: CDN configuration complete
}

// Example_apiAssets demonstrates serving assets for an API.
func Example_apiAssets() {
	handler := static.New(context.Background(), exampleContent, "testdata")

	// Strict security
	handler.SetPathSecurity(static.PathSecurityConfig{
		Enabled:       true,
		AllowDotFiles: false,
		MaxPathDepth:  5,
	})

	// Conservative rate limiting
	handler.SetRateLimit(static.RateLimitConfig{
		Enabled:     true,
		MaxRequests: 100,
		Window:      time.Minute,
	})

	// Long cache duration
	handler.SetHeaders(static.HeadersConfig{
		EnableCacheControl: true,
		CacheMaxAge:        86400, // 24 hours
		EnableETag:         true,
	})

	fmt.Println("API assets configuration complete")
	// Output: API assets configuration complete
}

// Example_fileOperations demonstrates file operations on embedded filesystem.
func Example_fileOperations() {
	_ = static.New(context.Background(), exampleContent, "testdata")

	// Check if file exists
	// exists := handler.Has("test.txt")

	// List all files
	// files, _ := handler.List("testdata")

	// Get file info
	// info, _ := handler.Info("test.txt")

	fmt.Println("File operations available")
	// Output: File operations available
}
