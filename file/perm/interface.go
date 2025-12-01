/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

// Package perm provides type-safe, portable file permission handling for Go applications.
//
// This package wraps os.FileMode with additional parsing, formatting, and encoding capabilities
// to simplify working with file permissions across different platforms and configuration formats.
//
// Key features:
//   - Octal string parsing (e.g., "0644", "0755")
//   - Multiple format encoding (JSON, YAML, TOML, CBOR, Text)
//   - Type conversions (int, uint, FileMode)
//   - Viper integration for configuration files
//   - Special permissions support (setuid, setgid, sticky bit)
//   - Quote handling and validation
//
// Example usage:
//
//	import (
//	    "os"
//	    "github.com/nabbar/golib/file/perm"
//	)
//
//	// Parse permission from string
//	p, err := perm.Parse("0644")
//	if err != nil {
//	    panic(err)
//	}
//
//	// Use with file operations
//	file, err := os.OpenFile("data.txt", os.O_CREATE|os.O_WRONLY, p.FileMode())
//	if err != nil {
//	    panic(err)
//	}
//	defer file.Close()
//
//	// Convert to different formats
//	fmt.Println(p.String())    // "0644"
//	fmt.Println(p.Uint64())    // 420
package perm

import (
	"os"
	"strconv"
)

type Perm os.FileMode

// Parse parses a string representation of a file permission into a Perm.
// It returns an error if the string is not a valid file permission.
//
// The string is expected to be in the format of a octal number, for example "0644".
// The function will return an error if the string is not a valid octal number,
// or if it does not represent a valid file permission.
//
// Example:
// p, err := Parse("0644")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// fmt.Println(p) // Output: 420
func Parse(s string) (Perm, error) {
	return parseString(s)
}

// ParseFileMode converts an os.FileMode to a Perm.
//
// This function is useful when you need to convert file mode information
// obtained from os.Stat() or os.Lstat() into a Perm value for further
// processing or serialization.
//
// Example:
//
//	info, err := os.Stat("file.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	perm := ParseFileMode(info.Mode())
//	fmt.Println(perm.String()) // Output: "0644" (or similar)
func ParseFileMode(p os.FileMode) Perm {
	return Perm(p)
}

// ParseInt parses an integer representation of a file permission into a Perm.
// It returns an error if the integer is not a valid file permission.
//
// The integer is expected to be in the range of a valid file permission, for example 420.
// The function will return an error if the integer is not in the range of a valid file permission.
//
// Example:
// p, err := ParseInt(420)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// fmt.Println(p) // Output: 420
func ParseInt(i int) (Perm, error) {
	return parseString(strconv.FormatInt(int64(i), 8))
}

// ParseInt64 parses an int64 representation of a file permission into a Perm.
// It returns an error if the int64 is not a valid file permission.
//
// The int64 is expected to be in the range of a valid file permission, for example 420.
// The function will return an error if the int64 is not in the range of a valid file permission.
//
// Example:
// p, err := ParseInt64(420)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// fmt.Println(p) // Output: 420
func ParseInt64(i int64) (Perm, error) {
	return parseString(strconv.FormatInt(i, 8))
}

// ParseByte parses a byte slice representation of a file permission into a Perm.
// It returns an error if the byte slice is not a valid file permission.
//
// The byte slice is expected to be in the format of a string representation of an octal number,
// for example "0644". The function will return an error if the byte slice is not a valid string
// representation of an octal number, or if it does not represent a valid file permission.
//
// Example:
// p, err := ParseByte([]byte("0644"))
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// fmt.Println(p) // Output: 420
func ParseByte(p []byte) (Perm, error) {
	return parseString(string(p))
}
