/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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
 *
 */

package retro

import (
	"regexp"
	"strconv"
	"strings"
)

// versionRegex is the regular expression used to validate semantic version strings.
// It matches versions in the format: [operator]vMAJOR.MINOR.PATCH
// where operator is optional and can be: <=, >=, <, >
// Example: ">=v1.2.3", "v2.0.0", "<v3.1.0"
var versionRegex = regexp.MustCompile(`^(<=|>=|<|>)?v(\d+)\.(\d+)\.(\d+)$`)

// isVersionSupported checks if a field should be included based on version constraints.
// This is the core function that evaluates "retro" struct tags to determine field visibility.
//
// The function supports:
//   - Empty retro tag: Field is always included
//   - "default": Matches when version is "default"
//   - Single version: "v1.0.0" - Exact match only
//   - Comparison operators: ">=v1.0.0", "<v2.0.0", ">v1.5.0", "<=v3.0.0"
//   - Range constraints: ">=v1.0.0,<v2.0.0" - Both conditions must be satisfied
//   - Multiple versions: "v1.0.0,v2.0.0" - Any match satisfies
//
// Parameters:
//   - version: The current version from the struct's Version field
//   - retro: The version constraint from the struct tag (e.g., ">=v1.0.0,<v2.0.0")
//
// Returns:
//   - bool: true if the field should be included, false otherwise
//
// Examples:
//
//	// Field included for v1.5.0 and above
//	retro:">=v1.5.0"
//
//	// Field included for versions below v2.0.0
//	retro:"<v2.0.0"
//
//	// Field included for versions between v1.0.0 and v2.0.0
//	retro:">=v1.0.0,<v2.0.0"
//
//	// Field included only for v1.0.0 or v2.0.0
//	retro:"v1.0.0,v2.0.0"
func isVersionSupported(version, retro string) bool {

	var (
		dualBoundaries, standaloneSatisfied       bool
		lowerVersion, upperVersion, lowOp, highOp string
	)

	// No retro tag means the field is always supported.
	if retro == "" {
		return true
	}

	versions := strings.Split(retro, ",")

	if !validRetroTag(versions) {
		return false
	}

	if detectedBoundaries(versions) {
		dualBoundaries = true
	}

	for _, ver := range versions {
		ver = strings.TrimSpace(ver)

		// If the current version is "default", check if retro tag has it.
		if version == "default" {
			if ver == "default" {
				return true
			}
			continue
		}

		// Check for standalone version (exception)
		if !strings.Contains(ver, ">") && !strings.Contains(ver, "<") &&
			!strings.Contains(ver, "<=") && !strings.Contains(ver, ">=") {
			if checkCondition(version, ver, "==") {
				standaloneSatisfied = true
			}

		}

		// Parse the version operator
		operator, compareVersion := parseOperator(ver)

		switch operator {
		case ">=", ">":
			lowerVersion = compareVersion
			lowOp = operator
		case "<=", "<":
			upperVersion = compareVersion
			highOp = operator
		}

	}

	if standaloneSatisfied {
		return true
	}

	if dualBoundaries {
		return checkCondition(version, lowerVersion, lowOp) && checkCondition(version, upperVersion, highOp)
	}

	if upperVersion != "" {
		return checkCondition(version, upperVersion, highOp)
	}

	if lowerVersion != "" {
		return checkCondition(version, lowerVersion, lowOp)
	}

	return false

}

// parseOperator extracts the comparison operator and version from a version constraint string.
// It identifies the operator prefix (>=, <=, >, <) and returns both the operator and the version.
//
// Parameters:
//   - ver: Version constraint string (e.g., ">=v1.0.0", "<v2.0.0", "v1.5.0")
//
// Returns:
//   - operator: The comparison operator (">=", "<=", ">", "<", or empty for exact match)
//   - compareVersion: The version string without the operator
//
// Examples:
//
//	parseOperator(">=v1.0.0") // returns ">=", "v1.0.0"
//	parseOperator("<v2.0.0")  // returns "<", "v2.0.0"
//	parseOperator("v1.5.0")   // returns "", "v1.5.0"
func parseOperator(ver string) (operator, compareVersion string) {
	switch {
	case strings.HasPrefix(ver, ">="):
		operator = ">="
		compareVersion = strings.TrimPrefix(ver, ">=")
	case strings.HasPrefix(ver, "<="):
		operator = "<="
		compareVersion = strings.TrimPrefix(ver, "<=")
	case strings.HasPrefix(ver, ">"):
		operator = ">"
		compareVersion = strings.TrimPrefix(ver, ">")
	case strings.HasPrefix(ver, "<"):
		operator = "<"
		compareVersion = strings.TrimPrefix(ver, "<")
	default:
		compareVersion = ver
	}
	return
}

// isValidVersion validates if a version string follows the expected format.
// A version is valid if it matches the semantic versioning pattern (with optional operator)
// or is the special "default" keyword.
//
// Valid formats:
//   - "default" - Special keyword for default version
//   - "v1.2.3" - Semantic version
//   - ">=v1.2.3" - Version with comparison operator
//   - "<v2.0.0" - Version with less-than operator
//
// Parameters:
//   - version: Version string to validate
//
// Returns:
//   - bool: true if the version format is valid, false otherwise
func isValidVersion(version string) bool {
	return versionRegex.MatchString(version) || version == "default"
}

// checkCondition evaluates a version comparison using the specified operator.
// It compares two semantic versions and returns whether the condition is satisfied.
//
// Supported operators:
//   - ">": Greater than
//   - "<": Less than
//   - ">=": Greater than or equal to
//   - "<=": Less than or equal to
//   - "==": Equal to
//
// Parameters:
//   - version: The current version to check
//   - compareVersion: The version to compare against
//   - operator: The comparison operator
//
// Returns:
//   - bool: true if the condition is satisfied, false otherwise
//
// Examples:
//
//	checkCondition("v1.5.0", "v1.0.0", ">=") // returns true
//	checkCondition("v1.5.0", "v2.0.0", "<")  // returns true
//	checkCondition("v1.5.0", "v1.5.0", "==") // returns true
func checkCondition(version, compareVersion, operator string) bool {
	comparison := compareVersions(strings.TrimPrefix(version, "v"),
		strings.TrimPrefix(compareVersion, "v"))

	switch operator {
	case ">":
		return comparison > 0
	case "<":
		return comparison < 0
	case ">=":
		return comparison >= 0
	case "<=":
		return comparison <= 0
	case "==":
		return comparison == 0
	default:
		return false
	}
}

// compareVersions performs a semantic version comparison between two version strings.
// It breaks down versions into major, minor, and patch components and compares them numerically.
//
// The comparison follows semantic versioning rules:
//  1. Compare major versions first
//  2. If equal, compare minor versions
//  3. If equal, compare patch versions
//
// Parameters:
//   - v1: First version string (without 'v' prefix, e.g., "1.2.3")
//   - v2: Second version string (without 'v' prefix, e.g., "2.0.0")
//
// Returns:
//   - int: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
//
// Examples:
//
//	compareVersions("1.2.3", "1.2.4") // returns -1
//	compareVersions("2.0.0", "1.9.9") // returns 1
//	compareVersions("1.5.0", "1.5.0") // returns 0
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")
	for i := 0; i < 3; i++ {
		num1, num2 := 0, 0
		if i < len(v1Parts) {
			num1, _ = strconv.Atoi(v1Parts[i])
		}
		if i < len(v2Parts) {
			num2, _ = strconv.Atoi(v2Parts[i])
		}
		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}
	return 0
}

// validRetroTag validates that a retro tag constraint is well-formed.
// It checks that:
//  1. All version strings are valid (match versionRegex or are "default")
//  2. Operators are not duplicated (only one '>' operator and one '<' operator allowed)
//
// This prevents invalid constraints like ">=v1.0.0,>=v2.0.0" which would be ambiguous.
//
// Parameters:
//   - values: Slice of version constraint strings from a retro tag
//
// Returns:
//   - bool: true if the tag is valid, false if malformed
//
// Examples:
//
//	validRetroTag([]string{">=v1.0.0", "<v2.0.0"}) // returns true (valid range)
//	validRetroTag([]string{">=v1.0.0", ">=v2.0.0"}) // returns false (duplicate >)
//	validRetroTag([]string{"v1.0.0", "v2.0.0"}) // returns true (multiple exact versions)
func validRetroTag(values []string) bool {
	operatorCount := make(map[string]bool)

	for _, version := range values {
		version = strings.TrimSpace(version)

		if !isValidVersion(version) {
			return false
		}

		if strings.HasPrefix(version, ">") {
			if operatorCount[">"] {
				return false
			}
			operatorCount[">"] = true
		} else if strings.HasPrefix(version, "<") {
			if operatorCount["<"] {
				return false
			}
			operatorCount["<"] = true
		}
	}

	return true
}

// detectedBoundaries checks if a retro tag defines a version range with both lower and upper bounds.
// A dual boundary exists when the tag contains both a '>' operator (lower bound) and a '<' operator (upper bound).
//
// This is used to determine if both conditions must be satisfied (AND logic) versus single boundary
// constraints where only one condition needs to be met.
//
// Parameters:
//   - versions: Slice of version constraint strings from a retro tag
//
// Returns:
//   - bool: true if both lower (>) and upper (<) boundaries are present, false otherwise
//
// Examples:
//
//	detectedBoundaries([]string{">=v1.0.0", "<v2.0.0"}) // returns true (range)
//	detectedBoundaries([]string{">=v1.0.0"}) // returns false (only lower bound)
//	detectedBoundaries([]string{"<v2.0.0"}) // returns false (only upper bound)
func detectedBoundaries(versions []string) bool {
	var (
		hasGreater, hasLess bool
	)

	for _, version := range versions {
		version = strings.TrimSpace(version)

		if strings.HasPrefix(version, ">") {
			hasGreater = true
		} else if strings.HasPrefix(version, "<") {
			hasLess = true
		}

		if hasGreater && hasLess {
			return true
		}
	}

	return false
}
