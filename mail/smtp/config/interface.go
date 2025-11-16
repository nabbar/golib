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
	"net"
	"net/url"
	"strconv"
	"strings"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
	libptc "github.com/nabbar/golib/network/protocol"
)

// SMTP is a read-only interface for accessing SMTP configuration parameters.
//
// This interface provides getters for all SMTP connection parameters including
// authentication credentials, network settings, and TLS configuration.
//
// See also:
//   - Config interface for read-write access
//   - github.com/nabbar/golib/mail/smtp/tlsmode for TLS mode constants
//   - github.com/nabbar/golib/network/protocol for network protocol types
type SMTP interface {
	// GetHost returns the SMTP server hostname or IP address.
	GetHost() string

	// GetPort returns the SMTP server port number (typically 25, 465, or 587).
	GetPort() int

	// GetUser returns the username for SMTP authentication.
	GetUser() string

	// GetPass returns the password for SMTP authentication.
	GetPass() string

	// GetNet returns the network protocol (tcp, tcp4, tcp6).
	// See github.com/nabbar/golib/network/protocol for protocol types.
	GetNet() libptc.NetworkProtocol

	// GetTls returns the TLS configuration.
	// See github.com/nabbar/golib/certificates for TLS configuration details.
	GetTls() libtls.Config

	// GetTlSServerName returns the TLS server name for SNI (Server Name Indication).
	GetTlSServerName() string

	// GetTlsMode returns the TLS connection mode (none, starttls, tls).
	// See github.com/nabbar/golib/mail/smtp/tlsmode for TLS mode constants.
	GetTlsMode() smtptp.TLSMode

	// SetTlsMode sets the TLS connection mode.
	SetTlsMode(mode smtptp.TLSMode)

	// IsTLSSkipVerify returns whether TLS certificate verification should be skipped.
	// WARNING: Skipping verification is insecure and should only be used for testing.
	IsTLSSkipVerify() bool

	// GetDsn returns the complete DSN string representation of the configuration.
	// The returned DSN can be used to recreate the configuration.
	GetDsn() string
}

// Config is a read-write interface for managing SMTP configuration.
//
// This interface extends SMTP with setter methods, allowing modification of
// all SMTP connection parameters. Use this interface when you need to
// programmatically build or modify SMTP configurations.
//
// Example:
//
//	cfg, err := New(ConfigModel{DSN: "tcp(localhost:25)/"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	cfg.SetHost("smtp.example.com")
//	cfg.SetPort(587)
//	cfg.SetTlsMode(smtptp.TLSStartTLS)
//	fmt.Println(cfg.GetDsn()) // user:pass@tcp(smtp.example.com:587)/starttls
type Config interface {
	SMTP

	// SetHost sets the SMTP server hostname or IP address.
	SetHost(host string)

	// SetPort sets the SMTP server port number.
	SetPort(port int)

	// SetUser sets the username for SMTP authentication.
	SetUser(user string)

	// SetPass sets the password for SMTP authentication.
	SetPass(pass string)

	// SetNet sets the network protocol (tcp, tcp4, tcp6).
	// See github.com/nabbar/golib/network/protocol for protocol types.
	SetNet(mode libptc.NetworkProtocol)

	// SetTls sets the TLS configuration.
	// See github.com/nabbar/golib/certificates for TLS configuration details.
	SetTls(tls libtls.Config)

	// ForceTLSSkipVerify enables or disables TLS certificate verification.
	// WARNING: Disabling verification is insecure and should only be used for testing.
	ForceTLSSkipVerify(skip bool)

	// SetTLSServerName sets the TLS server name for SNI (Server Name Indication).
	// This is useful when the server's hostname doesn't match the certificate.
	SetTLSServerName(serverName string)
}

// New parses a ConfigModel and returns a Config instance.
//
// This function parses an SMTP Data Source Name (DSN) string and extracts all
// connection parameters including host, port, credentials, network protocol,
// TLS mode, and query parameters.
//
// DSN Format:
//
//	[user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]
//
// Components:
//   - user[:password]@: Optional authentication credentials
//   - net: Network protocol (tcp, tcp4, tcp6)
//   - (addr): Optional address in the format host:port
//   - /tlsmode: TLS mode ("" for none, "starttls" for STARTTLS, "tls" for direct TLS)
//   - ?params: Optional query parameters (ServerName, SkipVerify)
//
// Supported Query Parameters:
//   - ServerName: TLS server name for SNI
//   - SkipVerify: "true" or "false" to skip TLS certificate verification
//
// Examples:
//
//	// Basic connection
//	cfg, _ := New(ConfigModel{DSN: "tcp(localhost:25)/"})
//
//	// With authentication
//	cfg, _ := New(ConfigModel{DSN: "user:pass@tcp(smtp.example.com:587)/starttls"})
//
//	// With TLS and custom server name
//	cfg, _ := New(ConfigModel{DSN: "tcp(mail.example.com:465)/tls?ServerName=smtp.example.com"})
//
// Returns:
//   - Config: A Config interface for accessing and modifying SMTP parameters
//   - liberr.Error: An error if DSN parsing fails (see error.go for error codes)
//
// Possible Errors:
//   - ErrorConfigInvalidDSN: Malformed DSN format
//   - ErrorConfigInvalidNetwork: Missing closing brace in address
//   - ErrorConfigInvalidParams: Invalid query parameters
//   - ErrorConfigInvalidHost: Missing required slash separator
func New(cfg ConfigModel) (Config, liberr.Error) {
	var (
		dsn     = cfg.DSN
		smtpcnf = &smtpConfig{
			DSN:    dsn,
			TLSCfg: cfg.TLS,
		}
	)

	// [user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]
	// Find the last '/' (since the password or the net addr might contain a '/')
	if !strings.ContainsRune(dsn, '?') && !strings.ContainsRune(dsn, '/') {
		dsn += "/"
	} else if strings.ContainsRune(dsn, '?') && !strings.ContainsRune(dsn, '/') {
		v := strings.Split(dsn, "?")
		v[len(v)-2] += "/"
		dsn = strings.Join(v, "?")
	}

	foundSlash := false
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			foundSlash = true
			var j, k int

			// left part is empty if i <= 0
			if i > 0 {
				// [username[:password]@][protocol[(address)]]
				// Find the last '@' in dsn[:i]
				for j = i; j >= 0; j-- {
					if dsn[j] == '@' {
						// username[:password]
						// Find the first ':' in dsn[:j]
						for k = 0; k < j; k++ {
							if dsn[k] == ':' {
								smtpcnf.Pass = dsn[k+1 : j]
								break
							}
						}
						smtpcnf.User = dsn[:k]

						break
					}
				}

				// [protocol[(address)]]
				// Find the first '(' in dsn[j+1:i]
				for k = j + 1; k < i; k++ {
					if dsn[k] == '(' {
						// dsn[i-1] must be == ')' if an address is specified
						if dsn[i-1] != ')' {
							if strings.ContainsRune(dsn[k+1:i], ')') {
								return nil, ErrorConfigInvalidDSN.Error(nil)
							}
							return nil, ErrorConfigInvalidNetwork.Error(nil)
						}

						if strings.ContainsRune(dsn[k+1:i-1], ':') {
							h, p, e := net.SplitHostPort(dsn[k+1 : i-1])
							if e == nil && p != "" {
								pint, er := strconv.ParseInt(p, 10, 64)
								if er == nil {
									if pint <= 65535 {
										smtpcnf.Port = int(pint)
									}
								}
								smtpcnf.Host = h
							}
						}

						if smtpcnf.Host == "" || smtpcnf.Port == 0 {
							smtpcnf.Host = dsn[k+1 : i-1]
						}

						break
					}
				}

				smtpcnf.Net = libptc.Parse(dsn[j+1 : k])

			}

			// [?param1=value1&...&paramN=valueN]
			// Find the first '?' in dsn[i+1:]
			for j = i + 1; j < len(dsn); j++ {
				if dsn[j] == '?' {

					if val, err := url.ParseQuery(dsn[j+1:]); err != nil {
						return nil, ErrorConfigInvalidParams.Error(err)
					} else {

						if val.Get("ServerName") != "" {
							smtpcnf.ServerName = val.Get("ServerName")
						}

						if val.Get("SkipVerify") != "" {
							vi, e := strconv.ParseBool(val.Get("SkipVerify"))
							if e == nil {
								smtpcnf.SkipVerify = vi
							}
						}
					}

					break
				}
			}

			smtpcnf.TLS = smtptp.Parse(dsn[i+1 : j])
			break
		}
	}

	if !foundSlash && len(dsn) > 0 {
		return nil, ErrorConfigInvalidHost.Error(nil)
	}

	return smtpcnf, nil
}
