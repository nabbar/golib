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
	"testing"

	libtls "github.com/nabbar/golib/certificates"
	smtpcfg "github.com/nabbar/golib/mail/smtp/config"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
	libptc "github.com/nabbar/golib/network/protocol"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestSMTPConfig is the entry point for the Ginkgo test suite
func TestSMTPConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SMTP Config Package Suite")
}

// Helper functions for test setup

// newConfigModel creates a ConfigModel with the given DSN
func newConfigModel(dsn string) smtpcfg.ConfigModel {
	return smtpcfg.ConfigModel{
		DSN: dsn,
	}
}

// newConfigModelWithTLS creates a ConfigModel with DSN and TLS config
func newConfigModelWithTLS(dsn string, tls libtls.Config) smtpcfg.ConfigModel {
	return smtpcfg.ConfigModel{
		DSN: dsn,
		TLS: tls,
	}
}

// createBasicConfig creates a basic valid SMTP config
func createBasicConfig() (smtpcfg.Config, error) {
	cfg := newConfigModel("tcp(localhost:25)/")
	return cfg.Config()
}

// createConfigWithAuth creates a config with authentication
func createConfigWithAuth(user, pass string) (smtpcfg.Config, error) {
	dsn := user + ":" + pass + "@tcp(localhost:25)/"
	cfg := newConfigModel(dsn)
	return cfg.Config()
}

// createConfigWithTLS creates a config with TLS mode
func createConfigWithTLS(host string, port int, tlsMode smtptp.TLSMode) (smtpcfg.Config, error) {
	dsn := "tcp(" + host + ":"
	// Convert port to string
	portStr := ""
	if port == 0 {
		portStr = "0"
	} else {
		p := port
		for p > 0 {
			portStr = string(rune('0'+p%10)) + portStr
			p /= 10
		}
	}
	dsn += portStr + ")/" + tlsMode.String()
	cfg := newConfigModel(dsn)
	return cfg.Config()
}

// assertConfigEquals checks if two configs have the same values
func assertConfigEquals(cfg1, cfg2 smtpcfg.Config) {
	Expect(cfg1.GetHost()).To(Equal(cfg2.GetHost()))
	Expect(cfg1.GetPort()).To(Equal(cfg2.GetPort()))
	Expect(cfg1.GetUser()).To(Equal(cfg2.GetUser()))
	Expect(cfg1.GetPass()).To(Equal(cfg2.GetPass()))
	Expect(cfg1.GetNet()).To(Equal(cfg2.GetNet()))
	Expect(cfg1.GetTlsMode()).To(Equal(cfg2.GetTlsMode()))
	Expect(cfg1.IsTLSSkipVerify()).To(Equal(cfg2.IsTLSSkipVerify()))
	Expect(cfg1.GetTlSServerName()).To(Equal(cfg2.GetTlSServerName()))
}

// validDSNs returns a list of valid DSN strings for testing
func validDSNs() []string {
	return []string{
		"tcp(localhost:25)/",
		"tcp(localhost:587)/starttls",
		"tcp(localhost:465)/tls",
		"user:pass@tcp(localhost:25)/",
		"user:pass@tcp(smtp.example.com:587)/starttls",
		"tcp(192.168.1.1:25)/",
		"tcp4(localhost:25)/",
		"tcp6([::1]:25)/",
		"user@tcp(localhost:25)/",
		"tcp(localhost:25)/?ServerName=smtp.example.com",
		"tcp(localhost:25)/?SkipVerify=true",
		"user:pass@tcp(localhost:587)/starttls?ServerName=smtp.example.com&SkipVerify=false",
	}
}

// invalidDSNs returns a list of invalid DSN strings for testing
func invalidDSNs() []string {
	return []string{
		"",                                 // Empty DSN
		"tcp(localhost:25",                 // Missing closing brace
		"tcp(localhost:25))",               // Extra closing brace
		"tcp localhost:25)/",               // Missing opening brace
		"user:pass@tcp(localhost:25?param", // Invalid params
		"tcp(localhost:25)/?param=",        // Invalid param value
		"tcp(localhost:65536)/",            // Port out of range
		"tcp(localhost:-1)/",               // Negative port
		"invalid",                          // Invalid format
		"tcp(localhost:25)",                // Missing slash
	}
}

// getDSNComponents extracts components from a config for testing
func getDSNComponents(cfg smtpcfg.Config) map[string]interface{} {
	return map[string]interface{}{
		"host":       cfg.GetHost(),
		"port":       cfg.GetPort(),
		"user":       cfg.GetUser(),
		"pass":       cfg.GetPass(),
		"net":        cfg.GetNet(),
		"tlsMode":    cfg.GetTlsMode(),
		"skipVerify": cfg.IsTLSSkipVerify(),
		"serverName": cfg.GetTlSServerName(),
		"dsn":        cfg.GetDsn(),
	}
}

// createDSNString creates a DSN string from components
func createDSNString(host string, port int, user, pass string, net libptc.NetworkProtocol, tlsMode smtptp.TLSMode) string {
	dsn := ""

	// Add user:pass if provided
	if user != "" {
		dsn += user
		if pass != "" {
			dsn += ":" + pass
		}
		dsn += "@"
	}

	// Add network and address
	dsn += net.String()
	if host != "" {
		dsn += "(" + host
		if port > 0 {
			dsn += ":"
			// Convert port to string manually
			portStr := ""
			p := port
			for p > 0 {
				portStr = string(rune('0'+p%10)) + portStr
				p /= 10
			}
			dsn += portStr
		}
		dsn += ")"
	}

	// Add TLS mode
	dsn += "/"
	if tlsMode != smtptp.TLSNone {
		dsn += tlsMode.String()
	}

	return dsn
}
