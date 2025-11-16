/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Package config provides SMTP configuration parsing and validation.
//
// This package handles SMTP Data Source Name (DSN) parsing and configuration management.
// It supports various SMTP connection types including plain, STARTTLS, and TLS connections.
//
// DSN Format:
//
//	[user[:password]@][net[(addr)]]/tlsmode[?param1=value1&paramN=valueN]
//
// Examples:
//   - tcp(localhost:25)/
//   - user:pass@tcp(smtp.example.com:587)/starttls
//   - tcp(mail.example.com:465)/tls?ServerName=mail.example.com&SkipVerify=false
//
// The package integrates with:
//   - github.com/nabbar/golib/certificates for TLS configuration
//   - github.com/nabbar/golib/mail/smtp/tlsmode for TLS mode handling
//   - github.com/nabbar/golib/network/protocol for network protocol parsing
//   - github.com/nabbar/golib/monitor/types for health monitoring integration
package config

import (
	"fmt"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	libmon "github.com/nabbar/golib/monitor/types"
)

// ConfigModel represents the SMTP configuration structure that can be used for
// serialization/deserialization from various formats (JSON, YAML, TOML).
//
// It contains:
//   - DSN: The SMTP Data Source Name in the format [user[:password]@][net[(addr)]]/tlsmode[?params]
//   - TLS: Optional TLS certificate configuration (see github.com/nabbar/golib/certificates)
//   - Monitor: Optional monitoring configuration (see github.com/nabbar/golib/monitor/types)
//
// Example JSON:
//
//	{
//	  "dsn": "user:pass@tcp(smtp.example.com:587)/starttls",
//	  "tls": {...},
//	  "monitor": {...}
//	}
type ConfigModel struct {
	DSN     string        `json:"dsn" yaml:"dsn" toml:"dsn" mapstructure:"dsn"`
	TLS     libtls.Config `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty" mapstructure:"tls,omitempty"`
	Monitor libmon.Config `json:"monitor,omitempty" yaml:"monitor,omitempty" toml:"monitor,omitempty" mapstructure:"monitor,omitempty"`
}

// Validate checks if the ConfigModel is valid.
//
// It performs the following validations:
//  1. Struct field validation using validator tags
//  2. DSN presence check (empty DSN returns ErrorConfigInvalidDSN)
//  3. DSN format validation by attempting to parse it with New()
//
// Returns nil if the configuration is valid, or a liberr.Error containing
// all validation errors otherwise. See github.com/nabbar/golib/errors for error handling.
func (c ConfigModel) Validate() liberr.Error {
	err := ErrorConfigValidator.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if c.DSN != "" {
		if _, er := New(c); er != nil {
			err.Add(er)
		}
	} else {
		err.Add(ErrorConfigInvalidDSN.Error(nil))
	}

	if !err.HasParent() {
		err = nil
	}

	return err
}

// Config creates a Config instance from the ConfigModel.
//
// This is a convenience method that calls New() with the current ConfigModel.
// It parses the DSN and returns a Config interface that can be used to access
// and modify SMTP connection parameters.
//
// Returns:
//   - Config: A Config interface for accessing SMTP parameters
//   - liberr.Error: An error if the DSN cannot be parsed (see error.go for error codes)
//
// Example:
//
//	model := ConfigModel{DSN: "tcp(localhost:587)/starttls"}
//	cfg, err := model.Config()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(cfg.GetHost()) // Output: localhost
func (c ConfigModel) Config() (Config, liberr.Error) {
	return New(c)
}
