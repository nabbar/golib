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

package perm_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nabbar/golib/file/perm"
)

// Example_basic demonstrates basic permission parsing from octal string.
func Example_basic() {
	// Parse a standard file permission
	p, err := perm.Parse("0644")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Octal: %s\n", p.String())
	fmt.Printf("Decimal: %d\n", p.Uint64())

	// Output:
	// Octal: 0644
	// Decimal: 420
}

// Example_symbolicNotation demonstrates parsing Unix symbolic notation.
func Example_symbolicNotation() {
	// Parse symbolic permission format (like ls -l output)
	p, err := perm.Parse("rwxr-xr-x")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Symbolic: rwxr-xr-x\n")
	fmt.Printf("Octal: %s\n", p.String())
	fmt.Printf("Decimal: %d\n", p.Uint64())

	// Output:
	// Symbolic: rwxr-xr-x
	// Octal: 0755
	// Decimal: 493
}

// Example_fileOperations demonstrates using permissions with file operations.
func Example_fileOperations() {
	// Parse permission
	p, err := perm.Parse("0644")
	if err != nil {
		panic(err)
	}

	// Create temporary file with specified permissions
	tmpfile, err := os.CreateTemp("", "example-*.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	// Set file permissions
	if err := os.Chmod(tmpfile.Name(), p.FileMode()); err != nil {
		panic(err)
	}

	// Verify permissions were set
	info, err := os.Stat(tmpfile.Name())
	if err != nil {
		panic(err)
	}

	fmt.Printf("File mode: %s\n", perm.ParseFileMode(info.Mode()).String())

	// Output:
	// File mode: 0644
}

// Example_typeConversions demonstrates various type conversion methods.
func Example_typeConversions() {
	p, _ := perm.Parse("0755")

	// Convert to different types
	fmt.Printf("String: %s\n", p.String())
	fmt.Printf("Uint64: %d\n", p.Uint64())
	fmt.Printf("Uint32: %d\n", p.Uint32())
	fmt.Printf("Uint: %d\n", p.Uint())
	fmt.Printf("Int: %d\n", p.Int())

	// Output:
	// String: 0755
	// Uint64: 493
	// Uint32: 493
	// Uint: 493
	// Int: 493
}

// Example_quotedStrings demonstrates handling of quoted permission strings.
func Example_quotedStrings() {
	// All these formats are equivalent
	p1, _ := perm.Parse("0644")
	p2, _ := perm.Parse("'0644'")
	p3, _ := perm.Parse("\"0644\"")

	fmt.Printf("Unquoted: %s\n", p1.String())
	fmt.Printf("Single quotes: %s\n", p2.String())
	fmt.Printf("Double quotes: %s\n", p3.String())

	// Output:
	// Unquoted: 0644
	// Single quotes: 0644
	// Double quotes: 0644
}

// Example_jsonSerialization demonstrates JSON marshaling and unmarshaling.
func Example_jsonSerialization() {
	type Config struct {
		FilePermission perm.Perm `json:"perm"`
	}

	// Marshal to JSON
	cfg := Config{FilePermission: perm.Perm(0644)}
	data, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("JSON: %s\n", data)

	// Unmarshal from JSON
	var cfg2 Config
	if err := json.Unmarshal(data, &cfg2); err != nil {
		panic(err)
	}

	fmt.Printf("Parsed: %s\n", cfg2.FilePermission.String())

	// Output:
	// JSON: {"perm":"0644"}
	// Parsed: 0644
}

// Example_parseFromInteger demonstrates parsing from integer values.
func Example_parseFromInteger() {
	// Parse from decimal integer (will be converted to octal)
	p1, err := perm.ParseInt(420) // decimal 420 = octal 644
	if err != nil {
		panic(err)
	}

	p2, err := perm.ParseInt64(493) // decimal 493 = octal 755
	if err != nil {
		panic(err)
	}

	fmt.Printf("From int 420: %s\n", p1.String())
	fmt.Printf("From int64 493: %s\n", p2.String())

	// Output:
	// From int 420: 0644
	// From int64 493: 0755
}

// Example_specialPermissions demonstrates parsing permissions with special bits.
func Example_specialPermissions() {
	// Standard permission
	p1, _ := perm.Parse("0755")
	fmt.Printf("Standard: %s (decimal: %d)\n", p1.String(), p1.Uint64())

	// With setuid bit (04755)
	p2, _ := perm.Parse("4755")
	fmt.Printf("With SetUID: %s (decimal: %d)\n", p2.String(), p2.Uint64())

	// With setgid bit (02755)
	p3, _ := perm.Parse("2755")
	fmt.Printf("With SetGID: %s (decimal: %d)\n", p3.String(), p3.Uint64())

	// With sticky bit (01777)
	p4, _ := perm.Parse("1777")
	fmt.Printf("With Sticky: %s (decimal: %d)\n", p4.String(), p4.Uint64())

	// Output:
	// Standard: 0755 (decimal: 493)
	// With SetUID: 04755 (decimal: 2541)
	// With SetGID: 02755 (decimal: 1517)
	// With Sticky: 01777 (decimal: 1023)
}

// Example_commonPermissions demonstrates commonly used permission values.
func Example_commonPermissions() {
	// Use slice to maintain order
	permissions := []struct {
		octal string
		desc  string
	}{
		{"0644", "Regular file (rw-r--r--)"},
		{"0755", "Executable (rwxr-xr-x)"},
		{"0600", "Sensitive file (rw-------)"},
		{"0700", "Private executable (rwx------)"},
		{"0666", "World-writable file (rw-rw-rw-)"},
		{"0777", "World-executable (rwxrwxrwx)"},
	}

	for _, item := range permissions {
		p, _ := perm.Parse(item.octal)
		fmt.Printf("%s: %d - %s\n", p.String(), p.Uint64(), item.desc)
	}

	// Output:
	// 0644: 420 - Regular file (rw-r--r--)
	// 0755: 493 - Executable (rwxr-xr-x)
	// 0600: 384 - Sensitive file (rw-------)
	// 0700: 448 - Private executable (rwx------)
	// 0666: 438 - World-writable file (rw-rw-rw-)
	// 0777: 511 - World-executable (rwxrwxrwx)
}

// Example_errorHandling demonstrates proper error handling.
func Example_errorHandling() {
	// Valid permission
	if p, err := perm.Parse("0644"); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Valid: %s\n", p.String())
	}

	// Invalid octal digit (falls back to symbolic parse, which also fails)
	if _, err := perm.Parse("0888"); err != nil {
		fmt.Printf("Invalid octal: error occurred\n")
	}

	// Invalid format
	if _, err := perm.Parse("invalid"); err != nil {
		fmt.Printf("Invalid format: error occurred\n")
	}

	// Empty string
	if _, err := perm.Parse(""); err != nil {
		fmt.Printf("Empty string: error occurred\n")
	}

	// Output:
	// Valid: 0644
	// Invalid octal: error occurred
	// Invalid format: error occurred
	// Empty string: error occurred
}
