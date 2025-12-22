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

package config

import libtls "github.com/nabbar/golib/certificates"

// TLSClient holds TLS configuration for client connections.
//
// This structure is embedded in the Client configuration to enable TLS/SSL
// encryption for TCP-based client connections.
//
// Fields:
//   - Enabled: Set to true to enable TLS encryption
//   - Config: Certificate configuration from github.com/nabbar/golib/certificates
//   - ServerName: Server hostname for certificate validation (required when Enabled is true)
//
// The ServerName field is used for SNI (Server Name Indication) and certificate
// hostname verification. It must match the server's certificate common name or
// one of its Subject Alternative Names.
type TLSClient struct {
	Enabled    bool          `json:"enabled" yaml:"enabled" toml:"enabled" mapstructure:"enabled"`
	Config     libtls.Config `json:"config" yaml:"config" toml:"config" mapstructure:"config"`
	ServerName string        `json:"server-name" yaml:"server-name" toml:"server-name" mapstructure:"server-name"`
}

// TLSServer holds TLS configuration for server connections.
//
// This structure is embedded in the Server configuration to enable TLS/SSL
// encryption for TCP-based server connections.
//
// Fields:
//   - Enabled: Set to true to enable TLS encryption
//   - Config: Certificate configuration from github.com/nabbar/golib/certificates
//
// When TLS is enabled, the Config must provide at least one valid certificate
// pair (certificate and private key). All client connections will be required
// to use TLS encryption.
type TLSServer struct {
	Enabled bool          `json:"enabled" yaml:"enabled" toml:"enabled" mapstructure:"enabled"`
	Config  libtls.Config `json:"config" yaml:"config" toml:"config" mapstructure:"config"`
}
