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

// Package static provides a secure, high-performance static file server for Gin framework
// with embedded filesystem support, rate limiting, path security, and WAF/IDS/EDR integration.
//
// This package is designed to serve static files from an embedded filesystem (embed.FS)
// with advanced security features including:
//
//   - Path traversal protection
//   - IP-based rate limiting
//   - Suspicious access detection
//   - MIME type validation
//   - HTTP caching (ETag, Cache-Control)
//   - Integration with WAF/IDS/EDR systems
//
// Thread Safety:
//
// All operations are thread-safe and use atomic operations without mutexes for
// maximum performance and scalability.
//
// Basic Usage:
//
//	package main
//
//	import (
//	    "context"
//	    "embed"
//	    "github.com/gin-gonic/gin"
//	    "github.com/nabbar/golib/static"
//	)
//
//	//go:embed assets/*
//	var content embed.FS
//
//	func main() {
//	    handler := static.New(context.Background(), content, "assets")
//	    router := gin.Default()
//	    handler.RegisterRouter("/static", router.GET)
//	    router.Run(":8080")
//	}
//
// For more information about related packages:
//   - github.com/nabbar/golib/logger - Logging interface
//   - github.com/nabbar/golib/router - Router registration helpers
//   - github.com/nabbar/golib/monitor - Health monitoring
package static

import (
	"context"
	"embed"
	"io"
	"os"
	"sync/atomic"

	ginsdk "github.com/gin-gonic/gin"
	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	libfpg "github.com/nabbar/golib/file/progress"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	librtr "github.com/nabbar/golib/router"
	libver "github.com/nabbar/golib/version"
)

// RateLimit interface provides IP-based rate limiting functionality to prevent
// scraping, enumeration attacks, and DoS attempts.
//
// The rate limiting tracks unique file paths per IP address and enforces
// configurable limits within time windows. All operations are thread-safe
// using atomic operations.
//
// See RateLimitConfig for configuration options.
type RateLimit interface {
	// SetRateLimit configures rate limiting parameters
	SetRateLimit(cfg RateLimitConfig)

	// GetRateLimit returns the current rate limit configuration
	GetRateLimit() RateLimitConfig

	// IsRateLimited checks if an IP address is currently rate limited
	IsRateLimited(ip string) bool

	// ResetRateLimit clears rate limit data for a specific IP address
	ResetRateLimit(ip string)
}

// PathSecurity interface provides protection against path traversal and other
// path-based security vulnerabilities.
//
// It validates requested paths against various security rules including:
//   - Path traversal attempts (../)
//   - Dot file access (.env, .git)
//   - Maximum path depth
//   - Blocked patterns
//   - Null byte injection
//
// See PathSecurityConfig for configuration options.
type PathSecurity interface {
	// SetPathSecurity configures path security validation rules
	SetPathSecurity(cfg PathSecurityConfig)

	// GetPathSecurity returns the current path security configuration
	GetPathSecurity() PathSecurityConfig

	// IsPathSafe validates if a requested path is safe to serve
	IsPathSafe(requestPath string) bool
}

// SuspiciousDetection interface provides detection and logging of suspicious
// file access patterns that may indicate security threats.
//
// It monitors access to files commonly targeted in attacks such as:
//   - Configuration files (.env, config.php)
//   - Backup files (.bak, .old)
//   - Admin panels (wp-admin, phpmyadmin)
//   - Database files (.sql, .db)
//
// See SuspiciousConfig for configuration options.
type SuspiciousDetection interface {
	// SetSuspicious configures suspicious access detection rules
	SetSuspicious(cfg SuspiciousConfig)

	// GetSuspicious returns the current suspicious detection configuration
	GetSuspicious() SuspiciousConfig
}

// HeadersControl interface provides HTTP caching and content-type validation.
//
// It manages:
//   - Cache-Control headers (public/private, max-age)
//   - ETag generation and validation (304 Not Modified)
//   - Content-Type detection and validation
//   - MIME type whitelisting/blacklisting
//
// See HeadersConfig for configuration options.
type HeadersControl interface {
	// SetHeaders configures HTTP headers and caching behavior
	SetHeaders(cfg HeadersConfig)

	// GetHeaders returns the current headers configuration
	GetHeaders() HeadersConfig
}

// SecurityBackend interface provides integration with WAF (Web Application Firewall),
// IDS (Intrusion Detection System), and EDR (Endpoint Detection and Response) systems.
//
// Security events are reported via:
//   - Webhooks (JSON or CEF format)
//   - Go callbacks for custom processing
//   - Batch processing for efficiency
//
// Supported event types include path traversal, rate limiting, suspicious access,
// and MIME type violations.
//
// See SecurityConfig for configuration options.
type SecurityBackend interface {
	// SetSecurityBackend configures integration with external security systems
	SetSecurityBackend(cfg SecurityConfig)

	// GetSecurityBackend returns the current security backend configuration
	GetSecurityBackend() SecurityConfig

	// AddSecurityCallback registers a Go callback function for security events
	AddSecurityCallback(callback SecuEvtCallback)
}

// StaticRegister interface provides methods for registering the static file handler
// with Gin routers and configuring logging.
//
// This interface integrates with github.com/nabbar/golib/router for flexible
// route registration in both root and group contexts.
type StaticRegister interface {
	// RegisterRouter registers the static handler on a route using the provided register function.
	// The route parameter specifies the URL path (e.g., "/static").
	// Additional middleware can be provided via router parameter.
	RegisterRouter(route string, register librtr.RegisterRouter, router ...ginsdk.HandlerFunc)

	// RegisterRouterInGroup registers the static handler in a router group.
	// This allows organizing routes under common prefixes or middleware.
	RegisterRouterInGroup(route, group string, register librtr.RegisterRouterInGroup, router ...ginsdk.HandlerFunc)

	// RegisterLogger sets the logger instance for the static handler.
	// See github.com/nabbar/golib/logger for logger implementation.
	RegisterLogger(log liblog.Logger)
}

// StaticIndex interface provides index file configuration for directory requests.
//
// Index files (e.g., index.html) are automatically served when a directory
// is requested, similar to Apache's DirectoryIndex or nginx's index directive.
type StaticIndex interface {
	// SetIndex configures an index file for a specific route and group.
	// When a directory is requested, this file will be served instead.
	SetIndex(group, route, pathFile string)

	// GetIndex returns the configured index file for a route and group.
	GetIndex(group, route string) string

	// IsIndex checks if a file is configured as an index file.
	IsIndex(pathFile string) bool

	// IsIndexForRoute checks if a file is the index for a specific route.
	IsIndexForRoute(pathFile, group, route string) bool
}

// StaticDownload interface configures files to be served as downloads
// (with Content-Disposition: attachment header).
//
// This forces the browser to download the file rather than displaying it inline.
type StaticDownload interface {
	// SetDownload marks a file to be served as an attachment download.
	SetDownload(pathFile string, flag bool)

	// IsDownload checks if a file is configured to be downloaded.
	IsDownload(pathFile string) bool
}

// StaticRedirect interface provides URL redirection configuration.
//
// This allows redirecting from one path to another, useful for maintaining
// backward compatibility or organizing file structure.
type StaticRedirect interface {
	// SetRedirect configures a redirect from source to destination route.
	// Returns HTTP 301 Permanent Redirect.
	SetRedirect(srcGroup, srcRoute, dstGroup, dstRoute string)

	// GetRedirect returns the destination for a source route.
	GetRedirect(srcGroup, srcRoute string) string

	// IsRedirect checks if a route is configured as a redirect.
	IsRedirect(group, route string) bool
}

// StaticSpecific interface allows overriding the default static file handler
// with custom handlers for specific routes.
//
// This is useful for adding special processing for certain paths while
// maintaining the default behavior for others.
type StaticSpecific interface {
	// SetSpecific registers a custom handler for a specific route.
	SetSpecific(group, route string, router ginsdk.HandlerFunc)

	// GetSpecific returns the custom handler for a route, if configured.
	GetSpecific(group, route string) ginsdk.HandlerFunc
}

// Static is the main interface for the static file handler.
//
// It combines all sub-interfaces and provides core file operations.
// All operations are thread-safe using atomic operations.
//
// The Static handler serves files from an embedded filesystem (embed.FS)
// with comprehensive security features and HTTP caching support.
//
// See the package documentation for usage examples.
type Static interface {
	// Has checks if a file exists in the embedded filesystem.
	Has(pathFile string) bool

	// List returns all files under a root path.
	List(rootPath string) ([]string, error)

	// Find opens a file and returns a ReadCloser.
	// The caller is responsible for closing the returned ReadCloser.
	Find(pathFile string) (io.ReadCloser, error)

	// Info returns file information (size, mod time, etc.).
	Info(pathFile string) (os.FileInfo, error)

	// Temp creates a temporary file copy with progress tracking.
	// Useful for large files. See github.com/nabbar/golib/file/progress.
	Temp(pathFile string) (libfpg.Progress, error)

	// Map iterates over all files in the embedded filesystem.
	// The provided function is called for each file.
	Map(func(pathFile string, inf os.FileInfo) error) error

	// UseTempForFileSize sets the size threshold for using temporary files.
	// Files larger than this size will be served via Temp() method.
	UseTempForFileSize(size int64)

	// Monitor returns health monitoring information.
	// See github.com/nabbar/golib/monitor/types for details.
	Monitor(ctx context.Context, cfg montps.Config, vrs libver.Version) (montps.Monitor, error)

	// Get is the main Gin handler function for serving files.
	// It handles all security checks, caching, and file serving.
	Get(c *ginsdk.Context)

	// SendFile sends a file to the client with appropriate headers.
	// This is typically called by Get() but can be used directly.
	SendFile(c *ginsdk.Context, filename string, size int64, isDownload bool, buf io.ReadCloser)

	// Embed router registration interface
	StaticRegister

	// Embed file serving configuration interfaces
	StaticIndex
	StaticDownload
	StaticRedirect
	StaticSpecific

	// Embed security interfaces
	RateLimit
	PathSecurity
	SuspiciousDetection
	HeadersControl
	SecurityBackend
}

// New creates a new Static file handler instance.
//
// Parameters:
//   - ctx: Context for lifecycle management
//   - content: Embedded filesystem containing static files
//   - embedRootDir: Optional root directory paths within the embed.FS
//
// The handler is initialized with:
//   - Default logger
//   - No security features enabled (must be configured)
//   - No rate limiting (must be configured)
//   - No index files (must be configured)
//
// Thread Safety:
//
// All internal data structures use atomic operations for thread-safe access
// without mutexes, ensuring high performance under concurrent load.
//
// Example:
//
//	//go:embed assets/*
//	var content embed.FS
//
//	handler := static.New(context.Background(), content, "assets")
//	handler.SetPathSecurity(static.DefaultPathSecurityConfig())
//	handler.SetRateLimit(static.RateLimitConfig{
//	    Enabled:     true,
//	    MaxRequests: 100,
//	    Window:      time.Minute,
//	})
func New(ctx context.Context, content embed.FS, embedRootDir ...string) Static {
	s := &staticHandler{
		log: libatm.NewValue[liblog.Logger](),
		rtr: libatm.NewValue[[]string](),

		efs: content,
		bph: libatm.NewValue[[]string](),
		siz: new(atomic.Int64),

		idx: libctx.New[string](ctx),
		dwn: libctx.New[string](ctx),
		flw: libctx.New[string](ctx),
		spc: libctx.New[string](ctx),

		rlc: libatm.NewValue[*RateLimitConfig](),
		rli: libatm.NewMapTyped[string, *ipTrack](),
		rlx: libatm.NewValue[context.CancelFunc](),

		psc: libatm.NewValue[*PathSecurityConfig](),
		sus: libatm.NewValue[*SuspiciousConfig](),
		hdr: libatm.NewValue[*HeadersConfig](),
		sec: libatm.NewValue[*SecurityConfig](),
		seb: libatm.NewValue[*evtBatch](),
	}

	s.setBase(embedRootDir...)
	s.setLogger(nil)

	return s
}
