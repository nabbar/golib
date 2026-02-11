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

package ftpclient

import (
	"context"
	"fmt"
	"time"

	libval "github.com/go-playground/validator/v10"
	libftp "github.com/jlaffaye/ftp"
	libtls "github.com/nabbar/golib/certificates"
)

// ConfigTimeZone defines the timezone configuration for FTP file timestamps.
// The offset is specified in seconds from UTC.
type ConfigTimeZone struct {
	Name   string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`         // Timezone name (e.g., "America/New_York", "UTC")
	Offset int    `mapstructure:"offset" json:"offset" yaml:"offset" toml:"offset"` // Offset in seconds from UTC
}

// Config holds the FTP client configuration.
// It supports various FTP features including TLS, extended passive mode (EPSV),
// machine-readable listings (MLSD), and file modification time commands (MDTM/MFMT).
//
// The configuration is validated using struct tags and the validator package.
// All public fields can be marshalled to/from JSON, YAML, TOML, and Viper config formats.
type Config struct {
	// Hostname define the host/port to connect to server.
	Hostname string `mapstructure:"hostname" json:"hostname" yaml:"hostname" toml:"hostname" validate:"required,hostname_rfc1123"`

	// Login define the login to use in login command.
	Login string `mapstructure:"login" json:"login" yaml:"login" toml:"login"`

	// Password defined the password to use in the login command.
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"` // #nosec nolint

	// ConnTimeout define a timeout duration for each connection (this is a global connection : connection, store contents, read content, ...).
	ConnTimeout time.Duration `mapstructure:"conn_timeout" json:"conn_timeout" yaml:"conn_timeout" toml:"conn_timeout"`

	// TimeZone force the time zone to use for the connection to the server.
	TimeZone ConfigTimeZone `mapstructure:"timezone" json:"timezone" yaml:"timezone" toml:"timezone"`

	// DisableUTF8 disable the UTF8 translation into the connection to the server.
	DisableUTF8 bool `mapstructure:"disable_utf8" json:"disable_utf8" yaml:"disable_utf8" toml:"disable_utf8"`

	// DisableEPSV disable the EPSV command into the connection to the server. (cf RFC 2428).
	DisableEPSV bool `mapstructure:"disable_epsv" json:"disable_epsv" yaml:"disable_epsv" toml:"disable_epsv"`

	// DisableMLSD disable the MLSD command into the connection to the server. (cf RFC 3659)
	DisableMLSD bool `mapstructure:"disable_mlsd" json:"disable_mlsd" yaml:"disable_mlsd" toml:"disable_mlsd"`

	// EnableMDTM enable the MDTM command into the connection to the server. (cf RFC 3659)
	EnableMDTM bool `mapstructure:"enable_mdtm" json:"enable_mdtm" yaml:"enable_mdtm" toml:"enable_mdtm"`

	// ForceTLS defined if the TLS connection must be forced or not.
	ForceTLS bool `mapstructure:"force_tls" json:"force_tls" yaml:"force_tls" toml:"force_tls"`

	// TLS define the client TLS config used if needed or forced
	TLS libtls.Config `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	fctx func() context.Context
	ftls func() libtls.TLSConfig
}

// Validate allow checking if the config' struct is valid with the awaiting model
func (c *Config) Validate() error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(c); err != nil {
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

// RegisterContext registers a function that provides context for FTP operations.
// This allows for timeout and cancellation support in long-running operations.
// If not registered, operations will not have context support.
func (c *Config) RegisterContext(fct func() context.Context) {
	c.fctx = fct
}

// RegisterDefaultTLS registers a function that provides default TLS configuration.
// This is used as a fallback when the Config.TLS field is not set.
// The function should return a TLSConfig that will be used for secure connections.
func (c *Config) RegisterDefaultTLS(fct func() libtls.TLSConfig) {
	c.ftls = fct
}

// New creates a new FTP server connection based on the current configuration.
// It applies all configured options including TLS, timeouts, timezone, and protocol features.
// The connection is established and authenticated (if credentials are provided) before returning.
//
// Returns the established connection or an error if connection/authentication fails.
func (c *Config) New() (*libftp.ServerConn, error) {
	var opt = make([]libftp.DialOption, 0)

	if tls := c.TLS.NewFrom(c.ftls()); c.ForceTLS {
		opt = append(opt, libftp.DialWithExplicitTLS(tls.TlsConfig("")))
	} else {
		opt = append(opt, libftp.DialWithTLS(tls.TlsConfig("")))
	}

	if c.fctx != nil {
		opt = append(opt, libftp.DialWithContext(c.fctx()))
	}

	if c.ConnTimeout != 0 {
		opt = append(opt, libftp.DialWithTimeout(c.ConnTimeout))
	}

	if c.TimeZone.Name != "" {
		tz := time.FixedZone(c.TimeZone.Name, c.TimeZone.Offset)
		opt = append(opt, libftp.DialWithLocation(tz))
	}

	if c.DisableUTF8 {
		opt = append(opt, libftp.DialWithDisabledUTF8(true))
	}

	if c.DisableEPSV {
		opt = append(opt, libftp.DialWithDisabledEPSV(true))
	}

	if c.DisableMLSD {
		opt = append(opt, libftp.DialWithDisabledMLSD(true))
	}

	if c.EnableMDTM {
		opt = append(opt, libftp.DialWithWritingMDTM(true))
	}

	if cli, err := libftp.Dial(c.Hostname, opt...); err != nil {
		return nil, ErrorFTPConnection.Error(err)
	} else if c.Login == "" && c.Password == "" {
		return cli, nil
	} else if err = cli.Login(c.Login, c.Password); err != nil {
		return cli, ErrorFTPLogin.Error(err)
	} else {
		return cli, nil
	}
}
