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

// Package certs provides certificate pair (private key + certificate) management.
//
// This package handles TLS certificate pairs consisting of a private key and its corresponding
// certificate. It supports multiple input formats including PEM-encoded strings, file paths,
// and certificate chains.
//
// Key Features:
//   - Parse certificate pairs from PEM-encoded strings or files
//   - Support for certificate chains (multiple certificates with one private key)
//   - Multiple configuration formats (pair, chain, single certificate)
//   - Convert to tls.Certificate for use with TLS connections
//   - Multiple encoding format support (JSON, YAML, TOML, CBOR)
//   - Thread-safe operations
//
// Certificate Formats:
//   - ConfigPair: Separate private key and certificate strings
//   - ConfigChain: Combined PEM string with both key and certificate(s)
//   - File paths: Load from files on disk
//
// Example:
//
//	keyPEM := `-----BEGIN RSA PRIVATE KEY-----
//	MIIEpAIBAAKCAQEA...`
//	certPEM := `-----BEGIN CERTIFICATE-----
//	MIIDXTCCAkWgAwIBAgIJ...`
//	cert, err := certs.Parse(keyPEM + "\n" + certPEM)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	tlsCert := cert.GetTLS()
package certs

import (
	"crypto/tls"
	"encoding"
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Cert represents a certificate pair (private key + certificate) for TLS connections.
// It provides methods for managing, parsing, and encoding certificate pairs.
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

	// Chain returns the certificate chain from the internal representation.
	// It returns an empty string and no error if the internal representation
	// is not a chain.
	//
	// The returned string is a PEM encoded certificate chain, where each
	// certificate is separated by a newline.
	//
	// If there is an error during parsing, it returns an empty string and
	// the error.
	//
	// The returned error is of type `x509.ParseCertificate` or
	// `x509.ParsePKCS7` if there is an error during parsing.
	Chain() (string, error)
	// Pair returns a pair of PEM encoded public and private keys.
	// It returns empty strings and no error if the internal representation
	// is not a pair.
	//
	// The returned public key is a PEM encoded public key, and the returned
	// private key is a PEM encoded private key.
	//
	// If there is an error during parsing, it returns empty strings and
	// the error.
	//
	// The returned error is of type `x509.ParseCertificate` or
	// `x509.ParsePKCS7` if there is an error during parsing.
	Pair() (pub string, key string, err error)
	// TLS returns the currently active certificate pair in the TLS configuration.
	//
	// The returned value is a tls.Certificate which contains the currently
	// active certificate pair in the TLS configuration.
	//
	// The returned value is empty if the TLS configuration does not contain
	// any certificate pairs.
	//
	// The TLS configuration is updated when a new certificate pair is added
	// using the `AddCertificatePair` functions.
	//
	// The TLS configuration is not updated when a new certificate pair is added
	// using the `AddCertificatePairString` functions.
	//
	// The TLS configuration is not updated when a new certificate pair is added
	// using the `AddCertificatePairFile` functions.
	TLS() tls.Certificate
	// Model returns the internal representation of the certificate.
	//
	// The returned value is a certificate which contains the internal
	// representation of the certificate.
	//
	// The returned value is empty if the internal representation is not a
	// valid certificate.
	//
	Model() Certif

	// IsChain returns true if the internal representation of the certificate is
	// a chain, and false otherwise.
	//
	// A chain is a PEM encoded certificate chain, where each certificate is
	// separated by a newline.
	//
	// The IsChain function does not check if the certificate chain is valid.
	// It only checks if the internal representation is a valid chain.
	//
	// The IsChain function is thread-safe.
	// Multiple goroutines can call the IsChain function at the same time without
	// affecting the correctness of the TLS configuration.
	IsChain() bool
	// IsPair returns true if the internal representation of the certificate is a pair,
	// and false otherwise.
	//
	// A pair is a PEM encoded private key and a PEM encoded public key.
	//
	// The IsPair function does not check if the pair is valid.
	// It only checks if the internal representation is a valid pair.
	//
	// The IsPair function is thread-safe.
	// Multiple goroutines can call the IsPair function at the same time without
	// affecting the correctness of the TLS configuration.
	IsPair() bool

	// IsFile returns true if the internal representation of the certificate is a file,
	// and false otherwise.
	//
	// A file is a path to a PEM file containing a certificate pair.
	//
	// The IsFile function does not check if the file is valid.
	// It only checks if the internal representation is a valid file.
	//
	// The IsFile function is thread-safe.
	// Multiple goroutines can call the IsFile function at the same time without
	// affecting the correctness of the TLS configuration.
	IsFile() bool
	// GetCerts returns the internal representation of the certificate as a slice of
	// strings.
	//
	// The returned slice of strings contains the internal representation of the
	// certificate. The internal representation can be a chain, a pair or a file.
	//
	// The GetCerts function does not check if the internal representation is
	// valid. It only returns the internal representation as a slice of strings.
	//
	// The GetCerts function is thread-safe.
	// Multiple goroutines can call the GetCerts function at the same time without
	// affecting the correctness of the TLS configuration.
	GetCerts() []string
}

// Parse parses a certificate chain from a PEM encoded string.
//
// The Parse function takes a PEM encoded certificate chain as a string parameter.
// It returns a certificate and an error.
//
// If the PEM encoded string cannot be parsed into a valid certificate chain, the
// Parse function returns an error of type tlscrt.ParseError.
//
// If the certificate chain is empty, the Parse function returns an error of type
// ErrInvalidPairCertificate.
//
// The Parse function is thread-safe.
// Multiple goroutines can call the Parse function at the same time without affecting
// the correctness of the TLS configuration.
func Parse(chain string) (Cert, error) {
	c := ConfigChain(chain)
	return parseCert(&c)
}

// ParsePair parses a certificate pair from a PEM encoded string.
//
// The ParsePair function takes two strings as parameters, the first parameter is a PEM
// encoded private key and the second parameter is a PEM encoded public key.
//
// It returns a certificate and an error.
//
// If the PEM encoded string cannot be parsed into a valid certificate pair, the
// ParsePair function returns an error of type tlscrt.ParseError.
//
// If the certificate pair is empty, the ParsePair function returns an error of type
// ErrInvalidPairCertificate.
//
// The ParsePair function is thread-safe.
// Multiple goroutines can call the ParsePair function at the same time without affecting
// the correctness of the TLS configuration.
func ParsePair(key, pub string) (Cert, error) {
	return parseCert(&ConfigPair{Key: key, Pub: pub})
}

func parseCert(cfg Config) (Cert, error) {
	if c, e := cfg.Cert(); e != nil {
		return nil, e
	} else if c == nil {
		return nil, ErrInvalidPairCertificate
	} else {
		return &Certif{g: cfg, c: *c}, nil
	}
}
