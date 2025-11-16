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

package config_test

import (
	"fmt"
	"log"

	smtpcfg "github.com/nabbar/golib/mail/smtp/config"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
)

// Example demonstrates basic SMTP configuration parsing.
func Example() {
	// Parse a simple DSN
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "tcp(localhost:25)/",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Host: %s\n", cfg.GetHost())
	fmt.Printf("Port: %d\n", cfg.GetPort())
	// Output:
	// Host: localhost
	// Port: 25
}

// ExampleNew_withAuthentication demonstrates parsing a DSN with authentication.
func ExampleNew_withAuthentication() {
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "user:password@tcp(smtp.example.com:587)/starttls",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Host: %s\n", cfg.GetHost())
	fmt.Printf("Port: %d\n", cfg.GetPort())
	fmt.Printf("User: %s\n", cfg.GetUser())
	fmt.Printf("TLS Mode: %s\n", cfg.GetTlsMode())
	// Output:
	// Host: smtp.example.com
	// Port: 587
	// User: user
	// TLS Mode: starttls
}

// ExampleNew_withTLS demonstrates parsing a DSN with direct TLS.
func ExampleNew_withTLS() {
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "tcp(mail.example.com:465)/tls",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Host: %s\n", cfg.GetHost())
	fmt.Printf("Port: %d\n", cfg.GetPort())
	fmt.Printf("TLS Mode: %s\n", cfg.GetTlsMode())
	// Output:
	// Host: mail.example.com
	// Port: 465
	// TLS Mode: tls
}

// ExampleNew_withQueryParameters demonstrates parsing a DSN with query parameters.
func ExampleNew_withQueryParameters() {
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "tcp(mail.example.com:465)/tls?ServerName=smtp.example.com&SkipVerify=false",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Host: %s\n", cfg.GetHost())
	fmt.Printf("Server Name: %s\n", cfg.GetTlSServerName())
	fmt.Printf("Skip Verify: %v\n", cfg.IsTLSSkipVerify())
	// Output:
	// Host: mail.example.com
	// Server Name: smtp.example.com
	// Skip Verify: false
}

// ExampleConfig_SetHost demonstrates modifying configuration programmatically.
func ExampleConfig_SetHost() {
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "tcp(localhost:25)/",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Modify configuration
	cfg.SetHost("smtp.example.com")
	cfg.SetPort(587)
	cfg.SetTlsMode(smtptp.TLSStartTLS)

	fmt.Printf("Host: %s\n", cfg.GetHost())
	fmt.Printf("Port: %d\n", cfg.GetPort())
	fmt.Printf("TLS Mode: %s\n", cfg.GetTlsMode())
	// Output:
	// Host: smtp.example.com
	// Port: 587
	// TLS Mode: starttls
}

// ExampleConfig_GetDsn demonstrates generating a DSN from configuration.
func ExampleConfig_GetDsn() {
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "tcp(localhost:25)/",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Modify configuration
	cfg.SetHost("smtp.example.com")
	cfg.SetPort(587)
	cfg.SetUser("user@example.com")
	cfg.SetPass("secret")
	cfg.SetTlsMode(smtptp.TLSStartTLS)

	// Generate DSN
	fmt.Println(cfg.GetDsn())
	// Output:
	// user@example.com:secret@tcp(smtp.example.com:587)/starttls
}

// ExampleConfigModel_Validate demonstrates configuration validation.
func ExampleConfigModel_Validate() {
	model := smtpcfg.ConfigModel{
		DSN: "tcp(localhost:25)/",
	}

	if err := model.Validate(); err != nil {
		log.Fatal("Invalid configuration:", err)
	}

	fmt.Println("Configuration is valid")
	// Output:
	// Configuration is valid
}

// ExampleConfigModel_Config demonstrates creating a Config from ConfigModel.
func ExampleConfigModel_Config() {
	model := smtpcfg.ConfigModel{
		DSN: "tcp(smtp.example.com:587)/starttls",
	}

	cfg, err := model.Config()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Host: %s\n", cfg.GetHost())
	fmt.Printf("Port: %d\n", cfg.GetPort())
	// Output:
	// Host: smtp.example.com
	// Port: 587
}

// ExampleConfig_SetTLSServerName demonstrates setting a custom TLS server name.
func ExampleConfig_SetTLSServerName() {
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "tcp(192.168.1.100:465)/tls",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set custom server name for SNI
	cfg.SetTLSServerName("smtp.example.com")

	fmt.Printf("Host: %s\n", cfg.GetHost())
	fmt.Printf("Server Name: %s\n", cfg.GetTlSServerName())
	// Output:
	// Host: 192.168.1.100
	// Server Name: smtp.example.com
}

// ExampleConfig_ForceTLSSkipVerify demonstrates disabling TLS verification.
func ExampleConfig_ForceTLSSkipVerify() {
	cfg, err := smtpcfg.New(smtpcfg.ConfigModel{
		DSN: "tcp(localhost:465)/tls",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Enable skip verification (INSECURE - only for testing!)
	cfg.ForceTLSSkipVerify(true)

	fmt.Printf("Skip Verify: %v\n", cfg.IsTLSSkipVerify())
	// Output:
	// Skip Verify: true
}
