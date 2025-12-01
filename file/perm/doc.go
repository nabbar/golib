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

// Package perm provides type-safe, portable file permission handling with support for
// multiple formats and serialization protocols.
//
// # Design Philosophy
//
// The perm package wraps os.FileMode to provide a unified, type-safe interface for working
// with file permissions across different platforms and configuration formats. The design
// emphasizes:
//
//  1. Format Flexibility: Support for octal strings ("0644"), symbolic notation ("rwxr-xr-x"),
//     and numeric values.
//  2. Serialization Support: Built-in marshaling/unmarshaling for JSON, YAML, TOML, CBOR,
//     and plain text.
//  3. Type Safety: Strong typing prevents accidental misuse of permission values.
//  4. Configuration Integration: Seamless Viper integration via custom decoder hooks.
//  5. Cross-Platform: Consistent behavior across Linux, macOS, and Windows.
//
// # Package Architecture
//
// The package is organized into specialized files:
//
//	interface.go  - Public API with Parse* constructors and Perm type definition
//	format.go     - Type conversion and formatting (String, Int*, Uint*, FileMode)
//	parse.go      - Parsing logic for octal and symbolic permission strings
//	encode.go     - Marshaling/unmarshaling for various formats (JSON, YAML, TOML, CBOR)
//	model.go      - Viper integration via decoder hooks
//
// Data flow:
//
//	┌──────────────────────────────────────────────────────────┐
//	│                     Input Sources                         │
//	│  "0644"  │  "rwxr-xr-x"  │  420  │  JSON/YAML/TOML/CBOR  │
//	└────┬─────┴────────┬──────┴───┬───┴──────────┬────────────┘
//	     │              │          │              │
//	     ▼              ▼          ▼              ▼
//	┌─────────────────────────────────────────────────────────┐
//	│               Parsing & Unmarshaling                     │
//	│  parseString()  │  parseLetterString()  │  Unmarshal*()│
//	└────────────────────────┬────────────────────────────────┘
//	                         ▼
//	                   ┌──────────┐
//	                   │   Perm   │  (os.FileMode wrapper)
//	                   └─────┬────┘
//	                         │
//	     ┌───────────────────┼───────────────────┐
//	     ▼                   ▼                   ▼
//	┌─────────┐      ┌────────────┐      ┌────────────┐
//	│ String()│      │ FileMode() │      │ Marshal*() │
//	│ Int*()  │      │ Uint*()    │      │ formats    │
//	└─────────┘      └────────────┘      └────────────┘
//
// # Permission Formats
//
// The package supports three input formats:
//
// 1. Octal Strings (Most Common):
//
//	"0644"    - Standard file permission
//	"0755"    - Executable file permission
//	"0777"    - All permissions
//	"644"     - Without leading zero (accepted)
//	"'0644'"  - Quoted strings (quotes stripped)
//
// 2. Symbolic Notation (Unix-style):
//
//	"rwxr-xr-x"    - 0755 equivalent
//	"rw-r--r--"    - 0644 equivalent
//	"-rwxr-xr-x"   - With file type indicator (regular file)
//	"drwxr-xr-x"   - Directory with 0755 permissions
//
// Symbolic format breakdown:
//   - 9 characters: owner(rwx) + group(rwx) + others(rwx)
//   - Optional 10th character prefix for file type (-, d, l, c, b, p, s, D)
//   - Each triplet: r=read(4), w=write(2), x=execute(1), -=none(0)
//
// 3. Numeric Values:
//
//	Parse("644")      - Parsed as octal
//	ParseInt(420)     - Decimal 420 = octal 0644
//	ParseInt64(493)   - Decimal 493 = octal 0755
//
// # Serialization Formats
//
// Automatic marshaling/unmarshaling for:
//
//	JSON:  {"perm": "0644"}
//	YAML:  perm: "0644"
//	TOML:  perm = "0644"
//	CBOR:  Binary encoding of "0644"
//	Text:  0644 (plain text)
//
// All formats use the canonical octal string representation ("0644").
//
// # Type Conversions
//
// The Perm type provides multiple conversion methods:
//
// To os.FileMode:
//
//	p.FileMode() os.FileMode  // For use with os.OpenFile, os.Chmod, etc.
//
// To String:
//
//	p.String() string         // Returns "0644" format
//
// To Integer Types:
//
//	p.Int() int               // With overflow protection
//	p.Int32() int32           // With overflow protection
//	p.Int64() int64           // With overflow protection
//	p.Uint() uint             // With overflow protection
//	p.Uint32() uint32         // With overflow protection
//	p.Uint64() uint64         // Direct conversion
//
// Overflow Handling:
// Integer conversion methods return the maximum value for that type if the permission
// value exceeds the type's capacity (e.g., Int32() returns math.MaxInt32 on overflow).
//
// # Viper Integration
//
// The package provides a decoder hook for Viper configuration library:
//
//	import (
//	    "github.com/nabbar/golib/file/perm"
//	    "github.com/spf13/viper"
//	)
//
//	type Config struct {
//	    FilePermission perm.Perm `mapstructure:"file_perm"`
//	}
//
//	v := viper.New()
//	v.SetConfigFile("config.yaml")
//
//	cfg := Config{}
//	opts := viper.DecoderConfigOption(func(c *mapstructure.DecoderConfig) {
//	    c.DecodeHook = perm.ViperDecoderHook()
//	})
//	v.Unmarshal(&cfg, opts)
//
// Configuration file (config.yaml):
//
//	file_perm: "0644"
//
// # Performance Characteristics
//
// The package is designed for minimal overhead:
//
//	Operation               Time Complexity    Allocations
//	─────────────────────────────────────────────────────────
//	Parse("0644")           O(n)               1-2 allocs
//	ParseInt(420)           O(1)               1-2 allocs
//	p.String()              O(1)               1 alloc
//	p.FileMode()            O(1)               0 allocs
//	p.Uint*(), p.Int*()     O(1)               0 allocs
//	MarshalJSON()           O(1)               2 allocs
//	UnmarshalJSON()         O(n)               2-3 allocs
//
// Parsing symbolic notation ("rwxr-xr-x") is O(n) with constant factor ~9-10.
//
// # Error Handling
//
// The package returns descriptive errors for invalid inputs:
//
//	Parse("0888")           // error: invalid octal digit
//	Parse("invalid")        // error: invalid permission (if not symbolic)
//	Parse("rwx")            // error: invalid permission group length
//	Parse("rwxr-xr-Z")      // error: invalid execute permission character: Z
//	Parse("")               // error: invalid permission
//
// All Parse* functions return (Perm, error). Marshal* functions may return errors
// for encoding failures, while Unmarshal* functions return errors for invalid input.
//
// # Thread Safety
//
// The Perm type is an immutable value type (wrapper around uint64), making it inherently
// thread-safe for concurrent reads. No synchronization is required when accessing the
// same Perm value from multiple goroutines.
//
// However, as with any Go value type, concurrent writes to the same Perm variable
// without synchronization will cause a data race. Protect concurrent writes with
// appropriate synchronization (mutex, channel, etc.).
//
// # Platform Considerations
//
// Windows:
//   - File permissions on Windows are emulated using os.FileMode
//   - Not all Unix permission bits are meaningful on Windows
//   - SetUID, SetGID, and Sticky bits may be ignored
//   - Standard permissions (0644, 0755) work as expected
//
// Unix/Linux/macOS:
//   - Full permission bit support including special bits
//   - SetUID (04000), SetGID (02000), Sticky (01000)
//   - Symbolic notation matches ls -l output format
//
// # Best Practices
//
// 1. Use Standard Permissions:
//
//	perm.Parse("0644")  // Regular files (rw-r--r--)
//	perm.Parse("0755")  // Executables  (rwxr-xr-x)
//	perm.Parse("0600")  // Sensitive files (rw-------)
//	perm.Parse("0700")  // Private executables (rwx------)
//
// 2. Always Check Errors:
//
//	p, err := perm.Parse(userInput)
//	if err != nil {
//	    return fmt.Errorf("invalid permission: %w", err)
//	}
//
// 3. Use FileMode() for os Package:
//
//	p, _ := perm.Parse("0644")
//	os.OpenFile(path, os.O_CREATE|os.O_WRONLY, p.FileMode())
//	os.Chmod(path, p.FileMode())
//
// 4. Leverage Serialization:
//
//	type Config struct {
//	    FileMode perm.Perm `json:"mode" yaml:"mode" toml:"mode"`
//	}
//
// 5. Quote Handling is Automatic:
//
//	perm.Parse("0644")   // Same as
//	perm.Parse("'0644'") // Same as
//	perm.Parse("\"0644\"")
//
// # Security Considerations
//
// Permission Validation:
//   - The package validates that permission values are within uint32 range
//   - Invalid octal digits (8, 9) are rejected
//   - Malformed symbolic notation is rejected
//   - Empty strings and whitespace-only input are rejected
//
// Sensitive Defaults:
//   - No default permissions are applied; caller must specify explicitly
//   - Recommended to use most restrictive permissions that meet requirements
//   - Avoid 0777 (world-writable) unless absolutely necessary
//
// Configuration Files:
//   - When loading from config files, validate against expected values
//   - Consider restricting to a whitelist of acceptable permissions
//   - Log permission changes for audit trails
//
// # Examples
//
// See example_test.go for comprehensive usage examples ranging from basic parsing
// to complex configuration scenarios.
//
// Quick Reference:
//
//	// Basic usage
//	p, _ := perm.Parse("0644")
//	file, _ := os.OpenFile("data.txt", os.O_CREATE, p.FileMode())
//
//	// From symbolic notation
//	p, _ := perm.Parse("rw-r--r--")
//	fmt.Println(p.String())  // "0644"
//
//	// From configuration
//	type Config struct {
//	    Mode perm.Perm `json:"mode"`
//	}
//	json.Unmarshal([]byte(`{"mode":"0755"}`), &cfg)
//
//	// Type conversions
//	p, _ := perm.Parse("0755")
//	fmt.Printf("Octal: %s\n", p.String())         // "0755"
//	fmt.Printf("Decimal: %d\n", p.Uint64())       // 493
//	fmt.Printf("FileMode: %v\n", p.FileMode())    // -rwxr-xr-x
package perm
