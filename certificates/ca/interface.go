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

// Package ca provides Certificate Authority (CA) certificate management and parsing.
//
// This package handles CA certificates in various formats (PEM, DER) and provides
// methods for parsing, validating, and managing CA certificates. CA certificates are
// used to verify the authenticity of other certificates in the chain.
//
// Key Features:
//   - Parse CA certificates from PEM-encoded strings or bytes
//   - Support for certificate chains (multiple certificates)
//   - Convert to x509.CertPool for use with TLS
//   - Multiple encoding format support (JSON, YAML, TOML, CBOR)
//   - Thread-safe operations
//
// Example:
//
//	pemData := `-----BEGIN CERTIFICATE-----
//	MIIBkTCB+wIJAKHHCgVZU...`
//	cert, err := ca.Parse(pemData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	pool := cert.GetCertPool()
package ca

import (
	"crypto/x509"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var (
	// ErrInvalidPairCertificate is returned when a certificate pair (key + cert) is invalid or incomplete.
	ErrInvalidPairCertificate = errors.New("invalid pair certificate")

	// ErrInvalidCertificate is returned when a certificate cannot be parsed or is malformed.
	ErrInvalidCertificate = errors.New("invalid certificate")
)

// Cert represents a Certificate Authority certificate or chain.
// It provides methods for managing, parsing, and encoding CA certificates.
// All operations are thread-safe.
type Cert interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler
	fmt.Stringer

	// Len returns the length of the underlying certificate chain.
	//
	// Len is part of the Cert interface.
	//
	// See also Chain and SliceChain
	Len() int
	// AppendPool appends the underlying certificate chain to the given *x509.CertPool.
	//
	// AppendPool is part of the Cert interface.
	//
	// See also Chain and SliceChain
	AppendPool(p *x509.CertPool)
	// AppendBytes appends the given byte slice to the underlying certificate
	// chain. If the given byte slice is empty, AppendBytes returns an
	// error.
	//
	// AppendBytes is part of the Cert interface.
	//
	// See also Chain and SliceChain
	AppendBytes(p []byte) error
	// AppendString appends the given PEM-encoded string to the underlying
	// certificate chain. If the given string is empty, AppendString returns
	// an error.
	//
	// AppendString is part of the Cert interface.
	//
	// See also Chain and SliceChain
	AppendString(str string) error
	// Chain returns the underlying certificate chain as a PEM-encoded
	// string. If the underlying certificate chain is empty, Chain returns
	// an error.
	//
	// Chain is part of the Cert interface.
	//
	// Chain is useful for serializing the underlying certificate chain
	// into a string. This string can be written to disk or sent over the
	// network.
	Chain() (string, error)
	// SliceChain returns a slice of PEM-encoded certificates
	// from the underlying certificate chain. If the underlying
	// certificate chain is empty, SliceChain returns an
	// error.
	//
	// SliceChain is part of the Cert interface.
	SliceChain() ([]string, error)
	// Model returns the certificate model of the Cert instance.
	//
	// Model returns a struct that contains the underlying
	// certificate and certificate chain. It is useful for
	// accessing the raw certificate data.
	//
	// Model is part of the Cert interface.
	//
	// See also certificate and certificate
	Model() Certif
}

// Parse parses the PEM-encoded certificate chain from the given string
// into a Cert instance. If the given string is empty, Parse returns
// an error.
//
// Parse is part of the Cert interface.
//
// # See also Cert, unMarshall, and certificate
//
// Example:
//
// s := "-----BEGIN CERTIFICATE-----\n"+
// "MIICajCCAcICAQEwgZ8wDQYJKoZIhvcNAQELBQAw\n"+
// "c2ExCzAJBgNVHR4ETnOmcxKDQwJjAxMCwxIDAQAB\n"+
// "-----END CERTIFICATE-----\n"
//
// c, e := Parse(s)
//
//	if e != nil {
//		log.Fatal(e)
//	}
//
// fmt.Println(c.Chain())
//
// Output:
//
// -----BEGIN CERTIFICATE-----
// MIICajCCAcICAQEwgZ8wDQYJKoZIhvcNAQELBQAw
// c2ExCzAJBgNVHR4ETnOmcxKDQwJjAxMCwxIDAQAB
// -----END CERTIFICATE-----
func Parse(str string) (Cert, error) {
	return ParseByte([]byte(str))
}

// ParseByte parses the PEM-encoded certificate chain from the given
// byte slice into a Cert instance. If the given byte slice is
// empty, ParseByte returns an error.
//
// ParseByte is part of the Cert interface.
//
// # See also Cert, unMarshall, and certificate
//
// Example:
//
// p := []byte("-----BEGIN CERTIFICATE-----\n"+
// "MIICajCCAcICAQEwgZ8wDQYJKoZIhvcNAQELBQAw\n"+
// "c2ExCzAJBgNVHR4ETnOmcxKDQwJjAxMCwxIDAQAB\n"+
// "-----END CERTIFICATE-----\n")
//
// c, e := ParseByte(p)
//
//	if e != nil {
//		log.Fatal(e)
//	}
//
// fmt.Println(c.Chain())
//
// Output:
//
// -----BEGIN CERTIFICATE-----
// MIICajCCAcICAQEwgZ8wDQYJKoZIhvcNAQELBQAw
// c2ExCzAJBgNVHR4ETnOmcxKDQwJjAxMCwxIDAQAB
// -----END CERTIFICATE-----
func ParseByte(p []byte) (Cert, error) {
	c := &Certif{
		c: make([]*x509.Certificate, 0),
	}

	if e := c.unMarshall(p); e != nil {
		return nil, e
	}

	return c, nil
}
