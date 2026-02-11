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

package config

import (
	"bytes"
	"fmt"
	"net/url"

	libtls "github.com/nabbar/golib/certificates"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
	libptc "github.com/nabbar/golib/network/protocol"
)

// smtpConfig is the internal implementation of the Config interface.
//
// This struct stores all SMTP connection parameters parsed from a DSN string.
// It is not exported; users should interact with it through the Config interface.
//
// Fields:
//   - DSN: The original DSN string
//   - Host: SMTP server hostname or IP
//   - Port: SMTP server port (25, 465, 587, etc.)
//   - User: Username for authentication
//   - Pass: Password for authentication
//   - Net: Network protocol (tcp, tcp4, tcp6)
//   - TLS: TLS connection mode (none, starttls, tls)
//   - SkipVerify: Whether to skip TLS certificate verification
//   - ServerName: TLS server name for SNI
//   - TLSCfg: TLS certificate configuration
type smtpConfig struct {
	DSN        string
	Host       string
	Port       int
	User       string
	Pass       string // #nosec nolint
	Net        libptc.NetworkProtocol
	TLS        smtptp.TLSMode
	SkipVerify bool
	ServerName string

	TLSCfg libtls.Config
}

// SetHost sets the SMTP server hostname or IP address.
func (c *smtpConfig) SetHost(host string) {
	c.Host = host
}

// GetHost returns the SMTP server hostname or IP address.
func (c *smtpConfig) GetHost() string {
	return c.Host
}

// SetPort sets the SMTP server port number.
// Common ports: 25 (SMTP), 587 (submission/STARTTLS), 465 (SMTPS).
func (c *smtpConfig) SetPort(port int) {
	c.Port = port
}

// GetPort returns the SMTP server port number.
func (c *smtpConfig) GetPort() int {
	return c.Port
}

// SetUser sets the username for SMTP authentication.
func (c *smtpConfig) SetUser(user string) {
	c.User = user
}

// GetUser returns the username for SMTP authentication.
func (c *smtpConfig) GetUser() string {
	return c.User
}

// SetPass sets the password for SMTP authentication.
// WARNING: The password is stored in plain text in memory.
func (c *smtpConfig) SetPass(pass string) {
	c.Pass = pass
}

// GetPass returns the password for SMTP authentication.
// WARNING: This returns the password in plain text.
func (c *smtpConfig) GetPass() string {
	return c.Pass
}

// SetNet sets the network protocol (tcp, tcp4, tcp6).
// See github.com/nabbar/golib/network/protocol for available protocols.
func (c *smtpConfig) SetNet(mode libptc.NetworkProtocol) {
	c.Net = mode
}

// GetNet returns the network protocol (tcp, tcp4, tcp6).
// See github.com/nabbar/golib/network/protocol for protocol types.
func (c *smtpConfig) GetNet() libptc.NetworkProtocol {
	return c.Net
}

// SetTlsMode sets the TLS connection mode (none, starttls, tls).
// See github.com/nabbar/golib/mail/smtp/tlsmode for available modes:
//   - TLSNone: Plain SMTP without encryption
//   - TLSStartTLS: STARTTLS (upgrade connection)
//   - TLSStrictTLS: Direct TLS connection
func (c *smtpConfig) SetTlsMode(mode smtptp.TLSMode) {
	c.TLS = mode
}

// GetTlsMode returns the TLS connection mode.
// See github.com/nabbar/golib/mail/smtp/tlsmode for mode constants.
func (c *smtpConfig) GetTlsMode() smtptp.TLSMode {
	return c.TLS
}

// SetTls sets the TLS certificate configuration.
// This includes client certificates, CA certificates, and other TLS settings.
// See github.com/nabbar/golib/certificates for configuration details.
func (c *smtpConfig) SetTls(tls libtls.Config) {
	c.TLSCfg = tls
}

// GetTls returns the TLS certificate configuration.
// See github.com/nabbar/golib/certificates for configuration structure.
func (c *smtpConfig) GetTls() libtls.Config {
	return c.TLSCfg
}

// ForceTLSSkipVerify enables or disables TLS certificate verification.
//
// When set to true, the client will accept any certificate presented by the server
// and any host name in that certificate. This makes TLS connections vulnerable to
// man-in-the-middle attacks.
//
// WARNING: Only use this for testing purposes. Never use in production environments.
//
// Parameters:
//   - skip: true to disable verification, false to enable (default)
func (c *smtpConfig) ForceTLSSkipVerify(skip bool) {
	c.SkipVerify = skip
}

// IsTLSSkipVerify returns whether TLS certificate verification is disabled.
//
// Returns:
//   - true: Certificate verification is disabled (insecure)
//   - false: Certificate verification is enabled (secure, default)
func (c *smtpConfig) IsTLSSkipVerify() bool {
	return c.SkipVerify
}

// SetTLSServerName sets the server name for TLS SNI (Server Name Indication).
//
// This is useful when:
//   - The server's hostname doesn't match its certificate
//   - Connecting through a proxy or load balancer
//   - Multiple virtual hosts share the same IP address
//
// Parameters:
//   - serverName: The server name to use for SNI verification
func (c *smtpConfig) SetTLSServerName(serverName string) {
	c.ServerName = serverName
}

// GetTlSServerName returns the TLS server name used for SNI.
//
// Returns an empty string if no custom server name is configured.
func (c *smtpConfig) GetTlSServerName() string {
	return c.ServerName
}

// GetDsn generates and returns the complete SMTP DSN string from the current configuration.
//
// This method reconstructs the DSN string from all configuration parameters,
// ensuring it follows the correct format:
//
//	[user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]
//
// The generated DSN can be used to:
//   - Store the configuration persistently
//   - Recreate an identical Config instance via New()
//   - Display the connection string (with sensitive data)
//
// DSN Components Generated:
//   - Authentication: user:password@ (only if user is set)
//   - Network: protocol(host:port) (port only if > 0)
//   - TLS Mode: /starttls, /tls, or / for none
//   - Parameters: ?ServerName=...&SkipVerify=true (only if set)
//
// Example outputs:
//   - "tcp(localhost:25)/"
//   - "user:pass@tcp(smtp.example.com:587)/starttls"
//   - "tcp(mail.example.com:465)/tls?ServerName=smtp.example.com&SkipVerify=true"
//
// WARNING: The returned DSN contains sensitive information (password) in plain text.
// Handle it carefully and avoid logging or displaying it in production environments.
//
// Returns:
//   - string: The complete DSN string representation
func (c *smtpConfig) GetDsn() string {
	var buf bytes.Buffer

	// [username[:password]@]
	// Only include credentials if username is present
	if len(c.User) > 0 {
		_, _ = buf.WriteString(c.User)
		if len(c.Pass) > 0 {
			_ = buf.WriteByte(':')
			_, _ = buf.WriteString(c.Pass)
		}
		_ = buf.WriteByte('@')
	}

	// [protocol[(address)]]
	// Always include the network protocol
	_, _ = buf.WriteString(c.Net.String())

	// [(host:port)]
	// Only include address if host is present
	if len(c.Host) > 0 {
		_ = buf.WriteByte('(')
		_, _ = buf.WriteString(c.Host)
		if c.Port > 0 {
			_ = buf.WriteByte(':')
			_, _ = buf.WriteString(fmt.Sprintf("%d", c.Port))
		}
		_ = buf.WriteByte(')')
	}

	// /tlsmode
	// Always include the separator, even for TLSNone (which renders as "")
	_ = buf.WriteByte('/')
	_, _ = buf.WriteString(c.TLS.String())

	// [?param1=value1&...&paramN=valueN]
	// Build query parameters only if any are set
	var val = &url.Values{}

	if c.ServerName != "" {
		val.Add("ServerName", c.ServerName)
	}

	if c.SkipVerify {
		val.Add("SkipVerify", "true")
	}

	params := val.Encode()

	// Only append parameters if they exist (length > 2 to account for minimal encoding)
	// nolint: gomnd
	if len(params) > 2 {
		_, _ = buf.WriteString("?" + params)
	}

	return buf.String()
}
