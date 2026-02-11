/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package httpcli

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
	libptc "github.com/nabbar/golib/network/protocol"
)

// OptionForceIP configures network interface binding for HTTP connections.
// This allows forcing connections through specific network interfaces or IP addresses,
// useful for multi-homed systems or testing specific network paths.
type OptionForceIP struct {
	Enable bool                   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`                                     // Enable ForceIP feature
	Net    libptc.NetworkProtocol `json:"net,omitempty" yaml:"net,omitempty" toml:"net,omitempty" mapstructure:"net,omitempty"`         // Network protocol (IPv4/IPv6)
	IP     string                 `json:"ip,omitempty" yaml:"ip,omitempty" toml:"ip,omitempty" mapstructure:"ip,omitempty"`             // Specific IP address to bind to
	Local  string                 `json:"local,omitempty" yaml:"local,omitempty" toml:"local,omitempty" mapstructure:"local,omitempty"` // Local address for binding
}

// OptionTLS configures TLS/SSL settings for HTTPS connections.
// Provides fine-grained control over certificate validation, cipher suites,
// and other TLS-related parameters.
type OptionTLS struct {
	Enable bool          `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"` // Enable TLS/SSL
	Config libtls.Config `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`             // TLS configuration (certificates, validation, etc.)
}

// OptionProxy configures HTTP/HTTPS proxy settings with authentication support.
// Supports both authenticated and unauthenticated proxies.
type OptionProxy struct {
	Enable   bool     `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`         // Enable proxy
	Endpoint *url.URL `json:"endpoint" yaml:"endpoint" toml:"endpoint" mapstructure:"endpoint"` // Proxy server URL
	Username string   `json:"username" yaml:"username" toml:"username" mapstructure:"username"` // #nosec nolint - Proxy authentication username
	Password string   `json:"password" yaml:"password" toml:"password" mapstructure:"password"` // #nosec nolint - Proxy authentication password
}

// Options defines the complete HTTP client configuration.
// This structure provides comprehensive control over client behavior including
// timeouts, protocol options, security settings, and network configuration.
//
// All fields support JSON, YAML, TOML, and Viper configuration formats through struct tags.
type Options struct {
	Timeout            time.Duration `json:"timeout" yaml:"timeout" toml:"timeout" mapstructure:"timeout"`
	DisableKeepAlive   bool          `json:"disable-keep-alive" yaml:"disable-keep-alive" toml:"disable-keep-alive" mapstructure:"disable-keep-alive"`
	DisableCompression bool          `json:"disable-compression" yaml:"disable-compression" toml:"disable-compression" mapstructure:"disable-compression"`
	Http2              bool          `json:"http2" yaml:"http2" toml:"http2" mapstructure:"http2"`
	TLS                OptionTLS     `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
	ForceIP            OptionForceIP `json:"force_ip" yaml:"force_ip" toml:"force_ip" mapstructure:"force_ip"`
	Proxy              OptionProxy   `json:"proxy" yaml:"proxy" toml:"proxy" mapstructure:"proxy"`
}

// DefaultConfig returns the default DNS mapper configuration in JSON format.
// The indent parameter specifies the indentation string for pretty-printing.
//
// This is a convenience function that delegates to the dns-mapper package's
// DefaultConfig function. The returned JSON can be used as a template for
// creating custom configurations.
//
// Parameters:
//   - indent: String to use for indentation (e.g., "  " for 2 spaces)
//
// Returns a byte slice containing the JSON-formatted default configuration.
//
// Example:
//
//	config := httpcli.DefaultConfig("  ")
//	fmt.Println(string(config))
func DefaultConfig(indent string) []byte {
	return htcdns.DefaultConfig(indent)
}

// Validate checks if the Options configuration is valid.
// It uses struct tags and the validator package to ensure all fields
// meet their specified constraints.
//
// Returns a liberr.Error containing all validation errors, or nil if validation succeeds.
// The error includes detailed information about which fields failed validation and why.
//
// Example:
//
//	opts := httpcli.Options{
//	    Timeout: 30 * time.Second,
//	}
//	if err := opts.Validate(); err != nil {
//	    log.Fatal("Invalid configuration:", err)
//	}
func (o Options) Validate() liberr.Error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(o); err != nil {
		if er, ok := err.(*libval.InvalidValidationError); ok {
			e.Add(er)
		}

		for _, er := range err.(libval.ValidationErrors) {
			//nolint #goerr113
			e.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", er.Namespace(), er.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

// GetClient creates and returns an HTTP client based on the options configuration.
// Currently, this method returns the default client and ignores the parameters.
//
// Parameters:
//   - def: Default TLS configuration (currently unused)
//   - servername: Server name for SNI (currently unused)
//
// Returns the default HTTP client configured with the global DNS mapper,
// along with a nil error.
//
// Note: This method is provided for interface compatibility and may be
// enhanced in future versions to use the provided parameters.
func (o Options) GetClient(def libtls.TLSConfig, servername string) (*http.Client, liberr.Error) {
	return GetClient(), nil
}
