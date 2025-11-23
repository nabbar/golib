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

import "time"

// HeadersConfig configures HTTP headers for caching and content-type validation.
//
// This configuration allows fine-grained control over:
//   - HTTP caching (Cache-Control, Expires, ETag)
//   - Content-Type detection and validation
//   - MIME type whitelisting/blacklisting
type HeadersConfig struct {
	// EnableCacheControl activates HTTP cache control headers
	EnableCacheControl bool

	// CacheMaxAge is the cache duration in seconds (e.g., 3600 = 1 hour)
	CacheMaxAge int

	// CachePublic when true, cache is public (CDN), otherwise private (browser only)
	CachePublic bool

	// EnableETag activates ETag generation for cache validation
	EnableETag bool

	// EnableContentType activates Content-Type detection and validation
	EnableContentType bool

	// AllowedMimeTypes is a list of allowed MIME types (empty = all allowed)
	AllowedMimeTypes []string

	// DenyMimeTypes is a list of forbidden MIME types
	DenyMimeTypes []string

	// CustomMimeTypes overrides MIME detection (extension -> mime-type mapping)
	CustomMimeTypes map[string]string
}

// SecurityEventType represents the type of security event that occurred.
// It is used to categorize security incidents for monitoring and analysis.
type SecurityEventType string

const (
	// EventTypePathTraversal indicates an attempt to access files outside the allowed directory
	EventTypePathTraversal SecurityEventType = "path_traversal"

	// EventTypeRateLimit indicates that an IP exceeded the allowed request rate
	EventTypeRateLimit SecurityEventType = "rate_limit_exceeded"

	// EventTypeSuspiciousAccess indicates suspicious file access patterns
	EventTypeSuspiciousAccess SecurityEventType = "suspicious_access"

	// EventTypeMimeTypeDenied indicates an attempt to access a file with a blocked MIME type
	EventTypeMimeTypeDenied SecurityEventType = "mime_type_denied"

	// EventTypeDotFileAccess indicates an attempt to access hidden files (starting with .)
	EventTypeDotFileAccess SecurityEventType = "dot_file_access"

	// EventTypePatternBlocked indicates an attempt to access a path matching a blocked pattern
	EventTypePatternBlocked SecurityEventType = "pattern_blocked"

	// EventTypePathDepth indicates a path exceeding the maximum allowed depth
	EventTypePathDepth SecurityEventType = "path_depth_exceeded"
)

// SecuEvtCallback is a callback function to process security events.
// It receives security events and can be used for custom handling, logging,
// or integration with external monitoring systems.
// The event parameter is of private type secEvt which contains detailed
// information about the security incident.
type SecuEvtCallback func(event secEvt)

// SecurityConfig configures the integration with WAF (Web Application Firewall),
// IDS (Intrusion Detection System), or EDR (Endpoint Detection and Response) systems.
//
// This configuration allows the static file handler to report security events to
// external systems via webhooks or callbacks. Events can be sent individually or
// batched for efficiency.
//
// Example usage:
//
//	handler.SetSecurityBackend(static.SecurityConfig{
//	    Enabled:        true,
//	    WebhookURL:     "https://waf.example.com/events",
//	    WebhookHeaders: map[string]string{"Authorization": "Bearer token"},
//	    WebhookAsync:   true,
//	    MinSeverity:    "medium",
//	    BatchSize:      100,
//	    BatchTimeout:   30 * time.Second,
//	})
type SecurityConfig struct {
	// Enabled activates the security integration
	Enabled bool

	// WebhookURL is the URL to send security events to (WAF/SIEM/IDS endpoint)
	WebhookURL string

	// WebhookTimeout is the timeout for webhook HTTP requests
	WebhookTimeout time.Duration

	// WebhookHeaders are custom HTTP headers to include in webhook requests (e.g., Authorization)
	WebhookHeaders map[string]string

	// WebhookAsync when true, sends webhooks asynchronously (non-blocking)
	WebhookAsync bool

	// Callbacks is a list of Go callback functions for custom event processing
	Callbacks []SecuEvtCallback

	// MinSeverity is the minimum severity level to notify (low, medium, high, critical)
	MinSeverity string

	// BatchSize is the number of events to accumulate before sending a batch (0 = real-time)
	BatchSize int

	// BatchTimeout is the maximum duration before sending an incomplete batch
	BatchTimeout time.Duration

	// EnableCEFFormat enables CEF (Common Event Format) for SIEM compatibility
	// See: https://www.microfocus.com/documentation/arcsight/arcsight-smartconnectors/
	EnableCEFFormat bool
}

// PathSecurityConfig configures path validation and security rules.
//
// This configuration protects against various path-based attacks including
// path traversal, dot file access, and access to sensitive directories.
type PathSecurityConfig struct {
	// Enabled activates or deactivates strict path validation
	Enabled bool

	// AllowDotFiles permits access to files starting with "." (default: false)
	// When false, blocks access to .env, .git, .htaccess, etc.
	AllowDotFiles bool

	// MaxPathDepth is the maximum allowed path depth (0 = unlimited)
	MaxPathDepth int

	// BlockedPatterns are path patterns to block (e.g., []string{"wp-admin", ".git"})
	BlockedPatterns []string
}

// RateLimitConfig configures IP-based rate limiting to prevent scraping and DoS attacks.
//
// The rate limiting tracks unique file paths requested per IP address within a time window.
// This helps protect against malicious clients that attempt to enumerate or download
// all files from the static file handler.
//
// Example usage:
//
//	handler.SetRateLimit(static.RateLimitConfig{
//	    Enabled:         true,
//	    MaxRequests:     100,
//	    Window:          time.Minute,
//	    CleanupInterval: 5 * time.Minute,
//	    WhitelistIPs:    []string{"127.0.0.1", "::1"},
//	})
//
// The rate limiting is thread-safe and uses atomic operations without mutexes.
type RateLimitConfig struct {
	// Enabled activates or deactivates rate limiting
	Enabled bool

	// MaxRequests is the maximum number of different files allowed per IP
	MaxRequests int

	// Window is the time duration for rate counting (e.g., 1 minute)
	Window time.Duration

	// CleanupInterval is the interval for automatic cache cleanup (e.g., 5 minutes)
	CleanupInterval time.Duration

	// WhitelistIPs is a list of IP addresses exempt from rate limiting (e.g., ["127.0.0.1", "::1"])
	WhitelistIPs []string

	// TrustedProxies is a list of trusted proxy IPs to extract real client IP
	TrustedProxies []string
}

// SuspiciousConfig configures the detection and logging of suspicious file access patterns.
//
// This feature helps identify potential security threats by monitoring access to
// files that are commonly targeted in attacks (e.g., .env files, backup files,
// configuration files).
//
// Example usage:
//
//	handler.SetSuspicious(static.SuspiciousConfig{
//	    Enabled:             true,
//	    LogSuccessfulAccess: true,
//	    SuspiciousPatterns:  []string{".env", ".git", "wp-admin"},
//	    SuspiciousExtensions: []string{".php", ".exe"},
//	})
type SuspiciousConfig struct {
	// Enabled activates or deactivates suspicious access detection
	Enabled bool

	// LogSuccessfulAccess also logs suspicious accesses that succeed (200 OK)
	LogSuccessfulAccess bool

	// SuspiciousPatterns are path patterns considered suspicious
	SuspiciousPatterns []string

	// SuspiciousExtensions are file extensions considered suspicious
	SuspiciousExtensions []string
}

// DefaultHeadersConfig returns a default HTTP headers configuration.
//
// Default values:
//   - EnableCacheControl: true
//   - CacheMaxAge: 3600 seconds (1 hour)
//   - CachePublic: true (allows CDN caching)
//   - EnableETag: true
//   - EnableContentType: true
//   - AllowedMimeTypes: empty (all allowed by default)
//   - DenyMimeTypes: executable types blocked
//   - CustomMimeTypes: includes wasm and webp
func DefaultHeadersConfig() HeadersConfig {
	return HeadersConfig{
		EnableCacheControl: true,
		CacheMaxAge:        3600, // 1 hour
		CachePublic:        true,
		EnableETag:         true,
		EnableContentType:  true,
		AllowedMimeTypes:   []string{}, // All allowed by default
		DenyMimeTypes: []string{
			"application/x-executable",
			"application/x-msdownload",
			"application/x-sh",
		},
		CustomMimeTypes: map[string]string{
			".wasm": "application/wasm",
			".webp": "image/webp",
		},
	}
}

// DefaultSecurityConfig returns a default security backend configuration.
//
// Default values:
//   - Enabled: false (must be explicitly enabled)
//   - WebhookTimeout: 5 seconds
//   - WebhookAsync: true (non-blocking)
//   - MinSeverity: "medium"
//   - BatchSize: 0 (real-time, no batching)
//   - BatchTimeout: 30 seconds
//   - EnableCEFFormat: false (JSON format)
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Enabled:         false, // Disabled by default
		WebhookTimeout:  5 * time.Second,
		WebhookAsync:    true,
		MinSeverity:     "medium",
		BatchSize:       0, // Real-time by default
		BatchTimeout:    30 * time.Second,
		EnableCEFFormat: false,
	}
}

// DefaultRateLimitConfig returns a secure default rate limiting configuration.
//
// Default values:
//   - Enabled: true
//   - MaxRequests: 100 unique files per window
//   - Window: 1 minute
//   - CleanupInterval: 5 minutes
//   - WhitelistIPs: localhost (IPv4 and IPv6)
//   - TrustedProxies: empty
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:         true,
		MaxRequests:     100,
		Window:          1 * time.Minute,
		CleanupInterval: 5 * time.Minute,
		WhitelistIPs:    []string{"127.0.0.1", "::1"},
		TrustedProxies:  []string{},
	}
}

// DefaultPathSecurityConfig returns a secure default path security configuration.
//
// Default values:
//   - Enabled: true
//   - AllowDotFiles: false (blocks .env, .git, etc.)
//   - MaxPathDepth: 10
//   - BlockedPatterns: [".git", ".svn", ".env", "node_modules"]
func DefaultPathSecurityConfig() PathSecurityConfig {
	return PathSecurityConfig{
		Enabled:         true,
		AllowDotFiles:   false,
		MaxPathDepth:    10,
		BlockedPatterns: []string{".git", ".svn", ".env", "node_modules"},
	}
}

// DefaultSuspiciousConfig returns a default suspicious access detection configuration.
//
// Default patterns include:
//   - Configuration files (.env, .git, wp-config, etc.)
//   - Backup files (.bak, .old, .swp, etc.)
//   - Admin panels (wp-admin, phpmyadmin, etc.)
//   - Sensitive paths (etc/passwd, windows/system32)
//   - Database files (.sql, .db)
//   - Executable extensions (.php, .exe, .sh)
func DefaultSuspiciousConfig() SuspiciousConfig {
	return SuspiciousConfig{
		Enabled:             true,
		LogSuccessfulAccess: true,
		SuspiciousPatterns: []string{
			// Configuration files
			".env", ".git", ".svn", ".htaccess", ".htpasswd",
			"web.config", "config.php", "wp-config",
			// Backup files
			".bak", ".backup", ".old", ".orig", ".save", ".swp",
			// Admin panels
			"admin", "wp-admin", "administrator", "phpmyadmin",
			// Sensitive paths
			"etc/passwd", "etc/shadow", "windows/system32",
			// Database
			".sql", ".db", ".sqlite",
			// Archives that might contain source
			".tar.gz", ".zip",
		},
		SuspiciousExtensions: []string{
			".php", ".asp", ".aspx", ".jsp", ".cgi",
			".exe", ".sh", ".bat", ".cmd",
		},
	}
}
