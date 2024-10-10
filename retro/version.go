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

var versionRegex = regexp.MustCompile(`^(<=|>=|<|>)?v(\d+)\.(\d+)\.(\d+)$`)

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

// Parse the operator and return it with the version string
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

// Check if the version is valid
func isValidVersion(version string) bool {
	return versionRegex.MatchString(version) || version == "default"
}

// Check the version condition using the operator
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

// Compares two version strings by breaking them into major, minor, and patch parts
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

// Ensure the retro tag has valid boundaries and that operators aren't duplicated
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

// Check if the tag contain a dual boundary
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
