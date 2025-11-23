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
	"strings"

	ginsdk "github.com/gin-gonic/gin"
	loglvl "github.com/nabbar/golib/logger/level"
)

// SetSuspicious configures suspicious access detection.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) SetSuspicious(cfg SuspiciousConfig) {
	s.sus.Store(&cfg)
}

// GetSuspicious returns the current suspicious access detection configuration.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) GetSuspicious() SuspiciousConfig {
	if cfg := s.sus.Load(); cfg != nil {
		return *cfg
	}
	return SuspiciousConfig{}
}

// isSuspiciousPath checks if a path matches suspicious patterns.
// Returns (true, reason) if suspicious, (false, "") otherwise.
//
// This method checks:
//   - Suspicious patterns in the path
//   - Suspicious file extensions
func (s *staticHandler) isSuspiciousPath(path string) (bool, string) {
	cfg := s.GetSuspicious()

	if !cfg.Enabled {
		return false, ""
	}

	pathLower := strings.ToLower(path)

	// Check suspicious patterns
	for _, pattern := range cfg.SuspiciousPatterns {
		if strings.Contains(pathLower, strings.ToLower(pattern)) {
			return true, "suspicious_pattern:" + pattern
		}
	}

	// Check suspicious extensions
	for _, ext := range cfg.SuspiciousExtensions {
		if strings.HasSuffix(pathLower, strings.ToLower(ext)) {
			return true, "suspicious_extension:" + ext
		}
	}

	return false, ""
}

// logSuspiciousAccess logs a suspicious access with full details.
// The log level is determined by the HTTP status code:
//   - 2xx: INFO level (successful suspicious access)
//   - 4xx/5xx: WARN level (failed suspicious access)
func (s *staticHandler) logSuspiciousAccess(c *ginsdk.Context, reason string, statusCode int) {
	cfg := s.GetSuspicious()

	if !cfg.Enabled {
		return
	}

	// If successful and not logging successes, skip
	if statusCode >= 200 && statusCode < 300 && !cfg.LogSuccessfulAccess {
		return
	}

	// Determine log level based on status
	level := loglvl.WarnLevel
	if statusCode >= 200 && statusCode < 300 {
		level = loglvl.InfoLevel // Successful suspicious access = INFO
	} else if statusCode >= 400 {
		level = loglvl.WarnLevel // Failed suspicious access = WARN
	}

	ent := s.getLogger().Entry(level, "suspicious access detected")
	ent.FieldAdd("ip", c.ClientIP())
	ent.FieldAdd("method", c.Request.Method)
	ent.FieldAdd("path", c.Request.URL.Path)
	ent.FieldAdd("reason", reason)
	ent.FieldAdd("status", statusCode)
	ent.FieldAdd("userAgent", c.GetHeader("User-Agent"))
	ent.FieldAdd("referer", c.GetHeader("Referer"))
	ent.FieldAdd("remoteAddr", c.Request.RemoteAddr)

	// Add X-Forwarded-For if present
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ent.FieldAdd("xForwardedFor", xff)
	}

	ent.Log()
}

// logAccessPattern logs suspicious access patterns like scanning and enumeration.
// This helps identify automated attacks or reconnaissance attempts.
func (s *staticHandler) logAccessPattern(c *ginsdk.Context, pattern string) {
	cfg := s.GetSuspicious()

	if !cfg.Enabled {
		return
	}

	ent := s.getLogger().Entry(loglvl.WarnLevel, "suspicious access pattern")
	ent.FieldAdd("ip", c.ClientIP())
	ent.FieldAdd("pattern", pattern)
	ent.FieldAdd("path", c.Request.URL.Path)
	ent.FieldAdd("userAgent", c.GetHeader("User-Agent"))
	ent.Log()
}

// checkAndLogSuspicious checks for and logs suspicious access patterns.
// This method is called for every request and detects:
//   - Suspicious paths (via isSuspiciousPath)
//   - Backup file scanning
//   - Config file scanning
//   - Directory traversal attempts
//   - Path manipulation attempts
//   - Admin panel scanning
func (s *staticHandler) checkAndLogSuspicious(c *ginsdk.Context, statusCode int) {
	if suspicious, reason := s.isSuspiciousPath(c.Request.URL.Path); suspicious {
		s.logSuspiciousAccess(c, reason, statusCode)
	}

	// Detect scanning patterns
	path := c.Request.URL.Path

	// Scanner looking for backup files
	if strings.Contains(path, ".bak") || strings.Contains(path, ".backup") ||
		strings.Contains(path, ".old") || strings.Contains(path, "~") {
		s.logAccessPattern(c, "backup_file_scanning")
	}

	// Scanner looking for config files
	if strings.Contains(strings.ToLower(path), "config") &&
		(strings.HasSuffix(path, ".php") || strings.HasSuffix(path, ".inc")) {
		s.logAccessPattern(c, "config_file_scanning")
	}

	// Directory traversal attempts (already blocked but log the attempt)
	if strings.Contains(path, "..") {
		s.logAccessPattern(c, "directory_traversal_attempt")
	}

	// Multiple slashes (potential path manipulation)
	if strings.Contains(path, "//") || strings.Contains(path, "\\\\") {
		s.logAccessPattern(c, "path_manipulation_attempt")
	}

	// Looking for admin panels
	pathLower := strings.ToLower(path)
	if strings.Contains(pathLower, "admin") || strings.Contains(pathLower, "login") ||
		strings.Contains(pathLower, "console") {
		s.logAccessPattern(c, "admin_panel_scanning")
	}
}
