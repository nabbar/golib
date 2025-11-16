/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

/*
Package tlsmode provides TLS connection mode types and utilities for SMTP connections.

# Overview

This package defines three TLS connection modes for SMTP:
  - TLSNone: Plain SMTP without encryption (port 25)
  - TLSStartTLS: SMTP with opportunistic STARTTLS upgrade (port 587)
  - TLSStrictTLS: Direct TLS connection / SMTPS (port 465)

The package provides comprehensive support for parsing, formatting, and encoding
TLS modes in various formats.

# Basic Usage

Parsing from strings:

	mode := tlsmode.Parse("starttls")
	fmt.Println(mode) // Output: starttls

	mode = tlsmode.Parse("tls")
	fmt.Println(mode) // Output: tls

Parsing from integers:

	mode := tlsmode.ParseInt64(1) // TLSStartTLS
	mode = tlsmode.ParseInt64(2)  // TLSStrictTLS

Converting to various types:

	mode := tlsmode.TLSStartTLS
	str := mode.String()     // "starttls"
	num := mode.Int()        // 1
	f := mode.Float64()      // 1.0

# String Parsing

The Parse function is case-insensitive and handles various formats:

	tlsmode.Parse("starttls")    // TLSStartTLS
	tlsmode.Parse("STARTTLS")    // TLSStartTLS
	tlsmode.Parse("start-tls")   // TLSStartTLS
	tlsmode.Parse("start_tls")   // TLSStartTLS
	tlsmode.Parse("  TLS  ")     // TLSStrictTLS
	tlsmode.Parse("tls")         // TLSStrictTLS
	tlsmode.Parse("")            // TLSNone
	tlsmode.Parse("none")        // TLSNone

The parser normalizes strings by removing whitespace, quotes, underscores,
hyphens, and line breaks before matching.

# Numeric Parsing

All standard numeric types are supported:

	// Integer types
	tlsmode.ParseInt64(int64(1))      // TLSStartTLS
	tlsmode.ParseUint64(uint64(2))    // TLSStrictTLS

	// Float types (floored to integer)
	tlsmode.ParseFloat64(1.5)         // TLSStartTLS (1)
	tlsmode.ParseFloat64(2.9)         // TLSStrictTLS (2)

Values outside the valid range (0-2) or negative numbers return TLSNone.

# JSON Encoding

TLS modes can be marshaled to and unmarshaled from JSON:

	type Config struct {
	    Mode tlsmode.TLSMode `json:"mode"`
	}

	// Marshal
	cfg := Config{Mode: tlsmode.TLSStartTLS}
	data, _ := json.Marshal(cfg)
	// Output: {"mode":"starttls"}

	// Unmarshal from string
	json.Unmarshal([]byte(`{"mode":"tls"}`), &cfg)
	// cfg.Mode == tlsmode.TLSStrictTLS

	// Unmarshal from number
	json.Unmarshal([]byte(`{"mode":1}`), &cfg)
	// cfg.Mode == tlsmode.TLSStartTLS

# YAML Encoding

YAML marshaling and unmarshaling is supported via gopkg.in/yaml.v3:

	type Config struct {
	    Mode tlsmode.TLSMode `yaml:"mode"`
	}

	// YAML: mode: starttls
	// or
	// YAML: mode: 1

# TOML Encoding

TOML encoding supports both string and numeric values:

	[config]
	mode = "starttls"
	# or
	mode = 1

# Other Encoding Formats

The package implements standard Go encoding interfaces:

  - encoding.TextMarshaler / encoding.TextUnmarshaler
  - encoding.BinaryMarshaler / encoding.BinaryUnmarshaler (via CBOR)
  - CBOR marshaling via github.com/fxamacker/cbor/v2

# Viper Integration

For configuration management with Viper, use the ViperDecoderHook:

	import (
	    "github.com/spf13/viper"
	    "github.com/nabbar/golib/mail/smtp/tlsmode"
	)

	type Config struct {
	    TLS tlsmode.TLSMode `mapstructure:"tls"`
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.ReadInConfig()

	var cfg Config
	err := v.Unmarshal(&cfg, viper.DecodeHook(
	    tlsmode.ViperDecoderHook(),
	))

The hook automatically converts strings, integers, and floats to TLSMode values.

# Standard Ports

Each TLS mode has an associated standard port:

  - TLSNone (plain SMTP): port 25
  - TLSStartTLS (submission): port 587
  - TLSStrictTLS (SMTPS): port 465

Alternative port 2525 is also commonly used for submission.

# Security Considerations

TLSNone (Plain SMTP):

Plain SMTP sends all data, including credentials, in clear text. Only use this
mode for:
  - Local testing
  - Communication within a trusted network
  - Relaying to a local mail server

Never use TLSNone for sending mail over the internet or with sensitive data.

TLSStartTLS (Opportunistic TLS):

STARTTLS upgrades a plain connection to TLS. This mode:
  - Is the recommended mode for mail submission (port 587)
  - Allows fallback to plain text if TLS fails (opportunistic)
  - Is vulnerable to downgrade attacks if not properly validated
  - Should verify the server's TLS certificate

TLSStrictTLS (Implicit TLS):

Direct TLS connections (SMTPS):
  - Provide the strongest security
  - Establish encryption before any SMTP commands
  - Are commonly used on port 465
  - Should verify the server's TLS certificate

Always verify TLS certificates in production environments.

# Type Conversions

The package provides comprehensive type conversion methods:

Numeric conversions:

	mode := tlsmode.TLSStartTLS
	mode.Int()      // 1 (int)
	mode.Int32()    // 1 (int32)
	mode.Int64()    // 1 (int64)
	mode.Uint()     // 1 (uint8)
	mode.Uint32()   // 1 (uint32)
	mode.Uint64()   // 1 (uint64)
	mode.Float32()  // 1.0 (float32)
	mode.Float64()  // 1.0 (float64)

String conversion:

	mode.String()   // "starttls"

All conversions are bidirectional - values can be parsed back using the
corresponding Parse functions.

# Error Handling

Parse functions never return errors. Invalid or out-of-range values
return TLSNone:

	tlsmode.Parse("invalid")      // TLSNone
	tlsmode.ParseInt64(-1)        // TLSNone
	tlsmode.ParseInt64(999)       // TLSNone
	tlsmode.ParseFloat64(256.0)   // TLSNone

Unmarshal functions return errors only for malformed input data:

	var mode tlsmode.TLSMode
	err := json.Unmarshal([]byte(`{invalid`), &mode)
	// err != nil (invalid JSON)

	err = json.Unmarshal([]byte(`true`), &mode)
	// err != nil (unsupported type)

# Thread Safety

All functions and methods in this package are safe for concurrent use.
TLSMode values are immutable and can be safely shared across goroutines.

# Performance

The package is designed for high performance:
  - String parsing: ~60ns per operation
  - Numeric parsing: ~120ns per operation
  - String conversion: <10ns per operation
  - JSON roundtrip: ~850ns per operation

Values are cached when appropriate, and all operations avoid allocations
where possible.

# Deprecated Functions

The following functions are deprecated but maintained for backward compatibility:

  - TLSModeFromString: use Parse instead
  - TLSModeFromInt: use ParseInt64 instead

# See Also

Related packages:
  - github.com/nabbar/golib/mail/smtp/config: SMTP configuration with TLS modes
  - gopkg.in/yaml.v3: YAML marshaling support
  - github.com/fxamacker/cbor/v2: CBOR marshaling support
  - github.com/go-viper/mapstructure/v2: Viper configuration decoding
*/
package tlsmode
