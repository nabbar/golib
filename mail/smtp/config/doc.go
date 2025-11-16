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
 *
 */

/*
Package config provides SMTP configuration parsing and validation.

# Overview

This package handles SMTP Data Source Name (DSN) parsing and configuration management.
It supports various SMTP connection types including plain SMTP, STARTTLS, and direct TLS connections.

# DSN Format

The DSN format follows this pattern:

	[user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]

Where:
  - user[:password]@: Optional SMTP authentication credentials
  - net: Network protocol (tcp, tcp4, tcp6)
  - (addr): Optional server address (host:port)
  - /tlsmode: TLS connection mode (empty, "starttls", or "tls")
  - ?params: Optional query parameters

# Supported Query Parameters

  - ServerName: TLS server name for SNI (Server Name Indication)
  - SkipVerify: Set to "true" to skip TLS certificate verification (insecure)

# Usage Examples

Basic SMTP Connection:

	cfg, err := config.New(config.ConfigModel{
		DSN: "tcp(localhost:25)/",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Host: %s, Port: %d\n", cfg.GetHost(), cfg.GetPort())

With Authentication and STARTTLS:

	cfg, err := config.New(config.ConfigModel{
		DSN: "user:password@tcp(smtp.example.com:587)/starttls",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User: %s, TLS Mode: %v\n", cfg.GetUser(), cfg.GetTlsMode())

With Direct TLS and Custom Server Name:

	cfg, err := config.New(config.ConfigModel{
		DSN: "tcp(mail.example.com:465)/tls?ServerName=smtp.example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server Name: %s\n", cfg.GetTlSServerName())

Programmatic Configuration:

	cfg, err := config.New(config.ConfigModel{DSN: "tcp(localhost:25)/"})
	if err != nil {
		log.Fatal(err)
	}

	// Modify configuration
	cfg.SetHost("smtp.example.com")
	cfg.SetPort(587)
	cfg.SetUser("user@example.com")
	cfg.SetPass("secretpassword")
	cfg.SetTlsMode(smtptp.TLSStartTLS)

	// Get updated DSN
	fmt.Println(cfg.GetDsn())
	// Output: user@example.com:secretpassword@tcp(smtp.example.com:587)/starttls

Configuration Validation:

	model := config.ConfigModel{
		DSN: "tcp(localhost:25)/",
	}

	if err := model.Validate(); err != nil {
		log.Fatal("Invalid configuration:", err)
	}

	cfg, err := model.Config()
	if err != nil {
		log.Fatal("Failed to create config:", err)
	}

Loading from JSON:

	jsonData := `{
		"dsn": "user:pass@tcp(smtp.example.com:587)/starttls"
	}`

	var model config.ConfigModel
	if err := json.Unmarshal([]byte(jsonData), &model); err != nil {
		log.Fatal(err)
	}

	cfg, err := model.Config()
	if err != nil {
		log.Fatal(err)
	}

# Integration with Other Packages

This package integrates seamlessly with other golib packages:

TLS Configuration (github.com/nabbar/golib/certificates):

	tlsConfig := certificates.Config{
		// TLS certificate configuration
	}

	model := config.ConfigModel{
		DSN: "tcp(smtp.example.com:465)/tls",
		TLS: tlsConfig,
	}

	cfg, _ := model.Config()
	fmt.Printf("Has TLS config: %v\n", cfg.GetTls())

Health Monitoring (github.com/nabbar/golib/monitor/types):

	monitorConfig := monitor.Config{
		// Monitoring configuration
	}

	model := config.ConfigModel{
		DSN:     "tcp(smtp.example.com:587)/starttls",
		Monitor: monitorConfig,
	}

# Network Protocol Support

The package supports IPv4, IPv6, and generic TCP connections through the
github.com/nabbar/golib/network/protocol package:

IPv4 Specific:

	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp4(smtp.example.com:587)/starttls",
	})

IPv6 Specific:

	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp6([2001:db8::1]:587)/starttls",
	})

Generic TCP (IPv4 or IPv6):

	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp(smtp.example.com:587)/starttls",
	})

# TLS Modes

The package supports three TLS connection modes through the
github.com/nabbar/golib/mail/smtp/tlsmode package:

No Encryption:

	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp(localhost:25)/",
	})
	// TLS mode: TLSNone

STARTTLS (Opportunistic TLS):

	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp(smtp.example.com:587)/starttls",
	})
	// TLS mode: TLSStartTLS

Direct TLS (SMTPS):

	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp(smtp.example.com:465)/tls",
	})
	// TLS mode: TLSStrictTLS

# Error Handling

The package defines several error codes for different parsing failures:

  - ErrorParamEmpty: Required parameter is missing
  - ErrorConfigValidator: Struct validation failed
  - ErrorConfigInvalidDSN: DSN format is invalid
  - ErrorConfigInvalidNetwork: Network address parsing failed
  - ErrorConfigInvalidParams: Query parameters are malformed
  - ErrorConfigInvalidHost: Host portion is invalid

All errors are managed through github.com/nabbar/golib/errors and provide
detailed error messages.

# Security Considerations

Password Storage:

Passwords are stored in plain text in memory. Consider using secure credential
management systems in production environments.

TLS Certificate Verification:

Always verify TLS certificates in production. Only use SkipVerify=true for
testing purposes:

	// INSECURE - Only for testing!
	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp(smtp.example.com:465)/tls?SkipVerify=true",
	})

SNI (Server Name Indication):

When the server's hostname doesn't match its certificate, use ServerName:

	cfg, _ := config.New(config.ConfigModel{
		DSN: "tcp(192.168.1.100:465)/tls?ServerName=smtp.example.com",
	})

# Common Port Numbers

  - 25: Standard SMTP (plain text or with STARTTLS)
  - 587: SMTP submission (typically with STARTTLS)
  - 465: SMTPS (direct TLS connection)
  - 2525: Alternative SMTP port (often used by cloud providers)

# Thread Safety

Config instances are not thread-safe for concurrent writes. If multiple goroutines
need to modify the same configuration, external synchronization is required.

Reading from a Config instance is safe for concurrent access as long as no
goroutine is modifying it.

# See Also

Related packages:
  - github.com/nabbar/golib/mail/smtp/tlsmode - TLS mode constants and parsing
  - github.com/nabbar/golib/certificates - TLS certificate configuration
  - github.com/nabbar/golib/network/protocol - Network protocol types
  - github.com/nabbar/golib/monitor/types - Health monitoring integration
  - github.com/nabbar/golib/errors - Error handling and management
*/
package config
