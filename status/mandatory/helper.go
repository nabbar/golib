/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package mandatory

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
)

var (
	// seq is an atomic counter used to generate unique suffixes for default mandatory group names.
	// This ensures thread-safe generation of unique identifiers across the application.
	seq = new(atomic.Uint32)

	// rgx is the regular expression used for name sanitization. It matches any character
	// that is NOT a lowercase letter (a-z), a digit (0-9), a hyphen (-), or an underscore (_).
	rgx = regexp.MustCompile(`[^a-z0-9\-_]+`)
)

// GetDefaultName generates a unique, sequential default name for a mandatory group.
// This function is typically used when a user does not provide a specific name for a group,
// ensuring that every group has a distinct identifier for logging and management purposes.
//
// The generated name follows the format "mandatory-<sequence_number>", where
// <sequence_number> is an atomically incrementing integer starting from 1.
//
// Returns:
//
//	A unique string identifier (e.g., "mandatory-1", "mandatory-2").
func GetDefaultName() string {
	return fmt.Sprintf("mandatory-%d", seq.Add(1))
}

// GetNameOrDefault processes a candidate name string to ensure it is valid and safe for use.
// It applies sanitization rules and, if the resulting name is empty or invalid, falls back
// to generating a unique default name.
//
// This function is useful for normalizing user input or configuration values before
// assigning them as group identifiers.
//
// Parameters:
//   - s: The candidate name string.
//
// Returns:
//
//	A sanitized version of the input string if valid, otherwise a generated default name.
func GetNameOrDefault(s string) string {
	s = FilterName(strings.ToLower(s))

	if len(s) < 1 {
		return GetDefaultName()
	}

	return s
}

// FilterName sanitizes a string by removing all characters that are not allowed in
// mandatory group names. The allowed character set is restricted to:
//   - Lowercase letters (a-z)
//   - Digits (0-9)
//   - Hyphens (-)
//   - Underscores (_)
//
// This strict filtering ensures that names are compatible with various metric systems
// (like Prometheus) and other downstream consumers that may have restrictions on
// identifier formats.
//
// Parameters:
//   - s: The raw string to be sanitized.
//
// Returns:
//
//	The sanitized string containing only allowed characters.
func FilterName(s string) string {
	return rgx.ReplaceAllString(s, "")
}
