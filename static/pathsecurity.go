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
	"fmt"
	"path"
	"strings"

	ginsdk "github.com/gin-gonic/gin"
	loglvl "github.com/nabbar/golib/logger/level"
)

// SetPathSecurity configures path security validation rules.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) SetPathSecurity(cfg PathSecurityConfig) {
	s.psc.Store(&cfg)
}

// GetPathSecurity returns the current path security configuration.
// This method is thread-safe and uses atomic operations.
func (s *staticHandler) GetPathSecurity() PathSecurityConfig {
	if cfg := s.psc.Load(); cfg != nil {
		return *cfg
	}
	return PathSecurityConfig{}
}

// validatePath validates a path against path traversal and other attacks.
//
// This method performs multiple security checks in sequence:
//  1. Null byte injection (before path cleaning)
//  2. Path traversal detection (.. sequences)
//  3. Path escape attempts (absolute paths)
//  4. Dot file access (hidden files)
//  5. Maximum path depth
//  6. Blocked patterns
//  7. Double slashes (logged but not blocked)
//
// The validation happens BEFORE path.Clean() for critical checks
// to prevent normalization from hiding attacks.
func (s *staticHandler) validatePath(requestPath string) error {
	cfg := s.GetPathSecurity()

	if !cfg.Enabled {
		return nil // Validation disabled
	}

	// 1. Check for null bytes (attack attempt) - BEFORE cleaning
	if strings.Contains(requestPath, "\x00") {
		s.logSecurityEvent("null byte in path", requestPath)
		return ErrorPathInvalid.Error(fmt.Errorf("null byte in path: %s", requestPath))
	}

	// 2. Check for path traversal BEFORE Clean() - look for .. in original path
	if strings.Contains(requestPath, "..") {
		s.logSecurityEvent("path traversal attempt", requestPath)
		return ErrorPathTraversal.Error(fmt.Errorf("path traversal attempt: %s", requestPath))
	}

	// 3. Clean the path for subsequent checks
	cleaned := path.Clean(requestPath)

	// 4. Verify the path doesn't escape above root
	if strings.HasPrefix(cleaned, "..") || strings.HasPrefix(cleaned, "/..") {
		s.logSecurityEvent("path escape attempt", requestPath)
		return ErrorPathTraversal.Error(fmt.Errorf("path escape attempt: %s", requestPath))
	}

	// 5. Check for hidden files (dot files)
	if !cfg.AllowDotFiles {
		parts := strings.Split(cleaned, "/")
		for _, part := range parts {
			if part != "" && part != "." && strings.HasPrefix(part, ".") {
				s.logSecurityEvent("dot file access attempt", requestPath)
				return ErrorPathDotFile.Error(fmt.Errorf("dot file not allowed: %s", requestPath))
			}
		}
	}

	// 6. Check maximum path depth
	if cfg.MaxPathDepth > 0 {
		depth := strings.Count(cleaned, "/")
		if depth > cfg.MaxPathDepth {
			s.logSecurityEvent("max path depth exceeded", requestPath)
			return ErrorPathDepth.Error(fmt.Errorf("path depth %d exceeds maximum %d: %s", depth, cfg.MaxPathDepth, requestPath))
		}
	}

	// 7. Check for blocked patterns
	for _, pattern := range cfg.BlockedPatterns {
		if pattern == "" {
			continue
		}
		if strings.Contains(cleaned, pattern) {
			s.logSecurityEvent("blocked pattern in path", requestPath)
			return ErrorPathBlocked.Error(fmt.Errorf("blocked pattern '%s' in path: %s", pattern, requestPath))
		}
	}

	// 8. Check for double slashes (even though Clean() handles them)
	if strings.Contains(requestPath, "//") {
		// Log but don't reject since Clean() fixes this
		ent := s.getLogger().Entry(loglvl.DebugLevel, "double slashes in path (cleaned)")
		ent.FieldAdd("originalPath", requestPath)
		ent.FieldAdd("cleanedPath", cleaned)
		ent.Log()
	}

	return nil
}

// logSecurityEvent logs a path security violation event.
func (s *staticHandler) logSecurityEvent(event, path string) {
	ent := s.getLogger().Entry(loglvl.WarnLevel, "path security violation")
	ent.FieldAdd("event", event)
	ent.FieldAdd("path", path)
	ent.Log()
}

// notifyPathSecurityEvent notifies external systems of a path security violation.
// This creates a high-severity security event with the violation reason.
func (s *staticHandler) notifyPathSecurityEvent(c *ginsdk.Context, eventType SecurityEventType, reason string) {
	details := map[string]string{
		"reason": reason,
	}

	event := s.newSecuEvt(c, eventType, "high", true, details)
	s.notifySecuEvt(event)
}

// IsPathSafe checks if a path is safe to serve (for external use).
// Returns true if the path passes all security validations.
func (s *staticHandler) IsPathSafe(requestPath string) bool {
	return s.validatePath(requestPath) == nil
}
