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

// Package certificates provides comprehensive TLS/SSL certificate management for secure communications.
//
// This package offers a complete solution for configuring TLS connections including certificate management,
// cipher suite selection, elliptic curve configuration, TLS version control, and client authentication.
//
// Key Features:
//   - Certificate management with support for files and in-memory certificates
//   - Root CA and Client CA management for certificate verification
//   - TLS version control (minimum/maximum version selection)
//   - Cipher suite configuration with support for TLS 1.2 and 1.3
//   - Elliptic curve configuration for ECDHE cipher suites
//   - Client authentication modes (NoClientCert, RequestClientCert, RequireAnyClientCert, VerifyClientCertIfGiven, RequireAndVerifyClientCert)
//   - Dynamic record sizing and session ticket controls
//   - Thread-safe operations for concurrent access
//   - Multiple encoding formats (JSON, YAML, TOML, CBOR)
//
// Subpackages:
//   - auth: Client authentication mode types and parsing
//   - ca: Certificate Authority management and parsing
//   - certs: Certificate pair management (private key + certificate)
//   - cipher: TLS cipher suite selection and management
//   - curves: Elliptic curve configuration for ECDHE
//   - tlsversion: TLS version management and parsing
//
// Example:
//
//	cfg := certificates.New()
//	cfg.SetVersionMin(tlsversion.VersionTLS12)
//	cfg.SetVersionMax(tlsversion.VersionTLS13)
//	cfg.AddRootCAFile("/path/to/ca.pem")
//	cfg.AddCertificatePairFile("/path/to/key.pem", "/path/to/cert.pem")
//	tlsConfig := cfg.TLS("example.com")
package certificates

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"

	tlsaut "github.com/nabbar/golib/certificates/auth"
	tlscas "github.com/nabbar/golib/certificates/ca"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"
)

// FctHttpClient is a function type that creates an HTTP client with TLS configuration.
// It receives a TLS configuration and a server name, and returns a configured *http.Client.
type FctHttpClient func(def TLSConfig, servername string) *http.Client

// FctTLSDefault is a function type that returns a default TLS configuration.
// It is useful for factory patterns or lazy initialization.
type FctTLSDefault func() TLSConfig

// FctRootCA is a function type that returns a list of root CA certificate paths or PEM strings.
type FctRootCA func() []string

// FctRootCACert is a function type that returns a parsed root CA certificate.
type FctRootCACert func() tlscas.Cert

// TLSConfig is the main interface for configuring TLS connections.
// It provides methods for managing certificates, cipher suites, TLS versions, and other TLS parameters.
// All operations are thread-safe and can be called concurrently from multiple goroutines.
type TLSConfig interface {
	// RegisterRand sets the source of randomness for the TLS connection.
	// It can be used to rotate the randomness source for example.
	//
	// The rand parameter should implement the io.Reader interface.
	// The TLS connection will use this reader to generate randomness.
	// If the reader is nil, the TLS connection will use the default source of randomness.
	//
	// The TLS connection will use this reader to generate randomness
	// for the lifetime of the connection. To rotate the randomness source,
	// call RegisterRand with a new reader.
	//
	RegisterRand(rand io.Reader)

	// AddRootCA adds a root CA to the TLS configuration.
	// It returns true if the root CA was added successfully, false otherwise.
	//
	// The root CA is added to the TLS configuration's root CA pool.
	// The root CA pool is used to verify the identity of the server.
	//
	// The root CA parameter should be a parsed certificate.
	// To parse a certificate from a PEM file, use the tlscas.Parse function.
	//
	// The AddRootCA function does not check if the root CA is already in the pool.
	// If you want to avoid adding the same root CA twice, you should check the pool before adding the root CA.
	AddRootCA(rootCA tlscas.Cert) bool
	// AddRootCAString adds a root CA to the TLS configuration from a string.
	// It returns true if the root CA was added successfully, false otherwise.
	//
	// The root CA is added to the TLS configuration's root CA pool.
	// The root CA pool is used to verify the identity of the server.
	//
	// The rootCA parameter should be a PEM encoded certificate.
	// To parse a certificate from a PEM file, use the tlscas.Parse function.
	//
	// The AddRootCAString function does not check if the root CA is already in the pool.
	// If you want to avoid adding the same root CA twice, you should check the pool before adding the root CA.
	AddRootCAString(rootCA string) bool
	// AddRootCAFile adds a root CA to the TLS configuration from a PEM file.
	//
	// The root CA is added to the TLS configuration's root CA pool.
	// The root CA pool is used to verify the identity of the server.
	//
	// The pemFile parameter should be the path to a PEM file containing the root CA.
	//
	// The AddRootCAFile function does not check if the root CA is already in the pool.
	// If you want to avoid adding the same root CA twice, you should check the pool before adding the root CA.
	//
	// The AddRootCAFile function returns an error if the PEM file cannot be read or if the root CA in the PEM file is invalid.
	AddRootCAFile(pemFile string) error
	// GetRootCA returns the root CA pool as a slice of Cert.
	// The root CA pool is used to verify the identity of the server.
	// The returned slice is a copy of the root CA pool and does not reference the original pool.
	// Modifying the returned slice does not affect the original pool.
	// The returned slice is ordered by the order the root CAs were added to the pool.
	GetRootCA() []tlscas.Cert
	// GetRootCAPool returns the root CA pool as a *x509.CertPool.
	// The root CA pool is used to verify the identity of the server.
	// The returned *x509.CertPool is a copy of the root CA pool and does not reference the original pool.
	// Modifying the returned *x509.CertPool does not affect the original pool.
	// The returned *x509.CertPool is ordered by the order the root CAs were added to the pool.
	GetRootCAPool() *x509.CertPool

	// AddClientCAString adds a client CA to the TLS configuration from a PEM encoded string.
	//
	// The client CA is added to the TLS configuration's client CA pool.
	// The client CA pool is used to verify the identity of the client.
	//
	// The ca parameter should be a PEM encoded certificate.
	// To parse a certificate from a PEM file, use the tlscas.Parse function.
	//
	// The AddClientCAString function does not check if the client CA is already in the pool.
	// If you want to avoid adding the same client CA twice, you should check the pool before adding the client CA.
	//
	// The AddClientCAString function returns true if the client CA is successfully added and false otherwise.
	AddClientCAString(ca string) bool
	// AddClientCAFile adds a client CA to the TLS configuration from a PEM file.
	//
	// The client CA is added to the TLS configuration's client CA pool.
	// The client CA pool is used to verify the identity of the client.
	//
	// The pemFile parameter should be the path to a PEM file containing the client CA.
	//
	// The AddClientCAFile function does not check if the client CA is already in the pool.
	// If you want to avoid adding the same client CA twice, you should check the pool before adding the client CA.
	//
	// The AddClientCAFile function returns an error if the PEM file cannot be read or if the client CA in the PEM file is invalid.
	AddClientCAFile(pemFile string) error
	// GetClientCA returns the client CA pool as a slice of tlscas.Cert.
	//
	// The client CA pool is used to verify the identity of the client.
	//
	// The returned slice is ordered by the order the client CAs were added to the pool.
	// Modifying the returned slice does not affect the original pool.
	GetClientCA() []tlscas.Cert
	// GetClientCAPool returns the client CA pool as a *x509.CertPool.
	//
	// The client CA pool is used to verify the identity of the client.
	//
	// The returned *x509.CertPool is ordered by the order the client CAs were added to the pool.
	// Modifying the returned *x509.CertPool does not affect the original pool.
	GetClientCAPool() *x509.CertPool
	// SetClientAuth sets the client authentication requirements for the TLS connection.
	//
	// The a parameter should be a tlsaut.ClientAuth containing the client authentication requirements.
	// The client authentication requirements are used to verify the identity of the client.
	//
	// The SetClientAuth function does not check if the client authentication requirements are already set.
	// If you want to avoid setting the same client authentication requirements twice, you should check the current client authentication requirements before setting the new ones.
	SetClientAuth(a tlsaut.ClientAuth)

	// AddCertificatePairString adds a certificate pair to the TLS configuration from a string.
	//
	// The key parameter should be a PEM encoded private key.
	// The crt parameter should be a PEM encoded certificate.
	//
	// The AddCertificatePairString function does not check if the certificate pair is already in the pool.
	// If you want to avoid adding the same certificate pair twice, you should check the pool before adding the certificate pair.
	//
	// The AddCertificatePairString function returns an error if the PEM encoded string cannot be parsed into a valid certificate pair.
	//
	// The returned error is of type tlscrt.ParseError.
	//
	// The AddCertificatePairString function is used to add a new certificate pair to the TLS configuration.
	// It is used to rotate the certificate pair for example.
	//
	// The AddCertificatePairString function does not affect the currently active certificate pair.
	// The currently active certificate pair is only replaced when the TLS connection is re-established.
	//
	// The AddCertificatePairString function is thread-safe.
	// Multiple goroutines can call the AddCertificatePairString function at the same time without affecting the correctness of the TLS configuration.
	AddCertificatePairString(key, crt string) error
	// AddCertificatePairFile adds a certificate pair to the TLS configuration from a PEM file.
	//
	// The keyFile parameter should be the path to a PEM file containing the private key.
	// The crtFile parameter should be the path to a PEM file containing the certificate.
	//
	// The AddCertificatePairFile function does not check if the certificate pair is already in the pool.
	// If you want to avoid adding the same certificate pair twice, you should check the pool before adding the certificate pair.
	//
	// The AddCertificatePairFile function returns an error if the PEM file cannot be read or if the private key and the certificate in the PEM file are invalid.
	//
	// The returned error is of type tlscrt.ParseError.
	//
	// The AddCertificatePairFile function is used to add a new certificate pair to the TLS configuration.
	// It is used to rotate the certificate pair for example.
	//
	// The AddCertificatePairFile function does not affect the currently active certificate pair.
	// The currently active certificate pair is only replaced when the TLS connection is re-established.
	//
	// The AddCertificatePairFile function is thread-safe.
	// Multiple goroutines can call the AddCertificatePairFile function at the same time without affecting the correctness of the TLS configuration.
	AddCertificatePairFile(keyFile, crtFile string) error
	// LenCertificatePair returns the number of certificate pairs in the TLS configuration.
	//
	// The function is thread-safe.
	// Multiple goroutines can call the LenCertificatePair function at the same time without affecting the correctness of the TLS configuration.
	//
	// The returned value is the number of certificate pairs in the TLS configuration.
	// The returned value does not include the currently active certificate pair.
	// The returned value is zero if the TLS configuration does not contain any certificate pairs.
	LenCertificatePair() int
	// CleanCertificatePair removes all the certificate pairs from the TLS configuration.
	//
	// The CleanCertificatePair function does not affect the currently active certificate pair.
	// The currently active certificate pair is only replaced when the TLS connection is re-established.
	//
	// The CleanCertificatePair function is thread-safe.
	// Multiple goroutines can call the CleanCertificatePair function at the same time without affecting the correctness of the TLS configuration.
	CleanCertificatePair()
	// GetCertificatePair returns all the certificate pairs in the TLS configuration.
	//
	// The returned value is a slice of tls.Certificate.
	// The slice contains all the certificate pairs in the TLS configuration.
	// The slice does not include the currently active certificate pair.
	// The slice is empty if the TLS configuration does not contain any certificate pairs.
	GetCertificatePair() []tls.Certificate

	// SetVersionMin sets the minimum version of TLS supported by the TLS configuration.
	//
	// The minimum version of TLS is the lowest version of TLS that the TLS configuration will support.
	// The TLS configuration will not support any versions of TLS that are lower than the minimum version.
	// The TLS configuration will support all versions of TLS that are equal to or higher than the minimum version.
	//
	// The SetVersionMin function is thread-safe.
	// Multiple goroutines can call the SetVersionMin function at the same time without affecting the correctness of the TLS configuration.
	SetVersionMin(v tlsvrs.Version)
	// GetVersionMin returns the minimum version of TLS supported by the TLS configuration.
	//
	// The returned value is the minimum version of TLS supported by the TLS configuration.
	// The returned value is zero if the TLS configuration does not contain any version of TLS.
	// The returned value is the minimum version of TLS supported by the TLS configuration if the TLS configuration contains multiple versions of TLS.
	// The returned value does not include the version of TLS that is currently active.
	// The returned value is not affected by the version of TLS that is currently active.
	// The returned value is thread-safe.
	// Multiple goroutines can call the GetVersionMin function at the same time without affecting the correctness of the TLS configuration.
	GetVersionMin() tlsvrs.Version
	// SetVersionMax sets the maximum version of TLS supported by the TLS configuration.
	//
	// The function sets the maximum version of TLS supported by the TLS configuration to the specified version.
	//
	// The specified version is the maximum version of TLS supported by the TLS configuration.
	// The specified version must be a valid version of TLS.
	// The specified version must not be less than the minimum version of TLS supported by the TLS configuration.
	//
	// The SetVersionMax function does not affect the currently active version of TLS.
	// The currently active version of TLS is only replaced when the TLS connection is re-established.
	//
	// The SetVersionMax function is thread-safe.
	// Multiple goroutines can call the SetVersionMax function at the same time without affecting the correctness of the TLS configuration.
	SetVersionMax(v tlsvrs.Version)
	// GetVersionMax returns the maximum version of TLS supported by the TLS configuration.
	//
	// The returned value is the maximum version of TLS supported by the TLS configuration.
	// The returned value is zero if the TLS configuration does not contain any version of TLS.
	// The returned value is the maximum version of TLS supported by the TLS configuration if the TLS configuration contains multiple versions of TLS.
	// The returned value does not include the version of TLS that is currently active.
	// The returned value is not affected by the version of TLS that is currently active.
	// The returned value is thread-safe.
	// Multiple goroutines can call the GetVersionMax function at the same time without affecting the correctness of the TLS configuration.
	GetVersionMax() tlsvrs.Version

	// SetCipherList sets the list of ciphers in the TLS configuration.
	//
	// The ciphers to set are specified as a slice of tlscpr.Cipher.
	//
	// The SetCipherList function replaces the current list of ciphers in the TLS configuration.
	// If you want to add ciphers to the current list, you should use the AddCiphers function.
	//
	// The SetCipherList function is thread-safe.
	// Multiple goroutines can call the SetCipherList function at the same time without affecting the correctness of the TLS configuration.
	SetCipherList(c []tlscpr.Cipher)
	// AddCiphers adds one or more ciphers to the TLS configuration.
	//
	// The ciphers to add are specified as a variable number of arguments.
	// Each argument should be of type tlscpr.Cipher.
	//
	// The AddCiphers function does not check if the ciphers are already in the pool.
	// If you want to avoid adding the same ciphers twice, you should check the pool before adding the ciphers.
	//
	// The AddCiphers function is thread-safe.
	// Multiple goroutines can call the AddCiphers function at the same time without affecting the correctness of the TLS configuration.
	AddCiphers(c ...tlscpr.Cipher)
	// GetCiphers returns the list of ciphers in the TLS configuration.
	//
	// The returned value is a slice of tlscpr.Cipher.
	// The slice contains all the ciphers in the TLS configuration.
	// The slice is empty if the TLS configuration does not contain any ciphers.
	// The returned value is ordered by the order the ciphers were added to the configuration.
	// Modifying the returned slice does not affect the original configuration.
	GetCiphers() []tlscpr.Cipher

	// SetCurveList sets the list of curves in the TLS configuration.
	//
	// The list of curves is specified as a slice of tlscrv.Curves.
	//
	// The SetCurveList function replaces the current list of curves in the TLS configuration.
	// If you want to add curves to the current list, you should use the AddCurves function.
	//
	// The SetCurveList function is thread-safe.
	// Multiple goroutines can call the SetCurveList function at the same time without affecting the correctness of the TLS configuration.
	SetCurveList(c []tlscrv.Curves)
	// AddCurves adds one or more curves to the TLS configuration.
	//
	// The curves to add are specified as a variable number of arguments.
	// Each argument should be of type tlscrv.Curves.
	//
	// The AddCurves function does not check if the curves are already in the pool.
	// If you want to avoid adding the same curves twice, you should check the pool before adding the curves.
	//
	// The AddCurves function is thread-safe.
	// Multiple goroutines can call the AddCurves function at the same time without affecting the correctness of the TLS configuration.
	AddCurves(c ...tlscrv.Curves)
	// GetCurves returns the list of curves in the TLS configuration.
	//
	// The returned value is a slice of tlscrv.Curves.
	// The slice contains all the curves in the TLS configuration.
	// The slice is empty if the TLS configuration does not contain any curves.
	// The returned value is ordered by the order the curves were added to the configuration.
	// Modifying the returned slice does not affect the original configuration.
	GetCurves() []tlscrv.Curves

	// SetDynamicSizingDisabled sets the TLS configuration to disable or enable dynamic record sizing.
	//
	// Dynamic record sizing is a feature of TLS that allows the TLS connection to dynamically adjust the size of the records being sent.
	// By default, dynamic record sizing is enabled.
	//
	// The SetDynamicSizingDisabled function takes a boolean as an argument.
	// If the argument is true, dynamic record sizing is disabled.
	// If the argument is false, dynamic record sizing is enabled.
	//
	// The SetDynamicSizingDisabled function is thread-safe.
	// Multiple goroutines can call the SetDynamicSizingDisabled function at the same time without affecting the correctness of the TLS configuration.
	SetDynamicSizingDisabled(flag bool)
	// SetSessionTicketDisabled sets the TLS configuration to disable or enable session tickets.
	//
	// Session tickets are used to resume a TLS connection without needing to re-establish the entire connection.
	// By default, session tickets are enabled.
	//
	// The SetSessionTicketDisabled function takes a boolean as an argument.
	// If the argument is true, session tickets are disabled.
	// If the argument is false, session tickets are enabled.
	//
	// The SetSessionTicketDisabled function is thread-safe.
	// Multiple goroutines can call the SetSessionTicketDisabled function at the same time without affecting the correctness of the TLS configuration.
	SetSessionTicketDisabled(flag bool)

	// Clone returns a copy of the TLSConfig.
	//
	// The returned TLSConfig is safe for concurrent use.
	//
	// The returned TLSConfig is a copy of the TLSConfig.
	// Modifying the returned TLSConfig does not affect the original TLSConfig.
	// The returned TLSConfig is independent of the original TLSConfig.
	// The Clone function is thread-safe.
	// Multiple goroutines can call the Clone function at the same time without affecting the correctness of the TLS configuration.
	Clone() TLSConfig
	// TLS returns a TLS configuration based on the TLSConfig.
	//
	// The returned TLS configuration is safe for concurrent use.
	//
	// The returned TLS configuration is not a copy of the TLSConfig.
	// Instead, it is a reference to the TLSConfig.
	// Modifying the returned TLS configuration affects the TLSConfig.
	// The returned TLS configuration is the same as the TLSConfig.
	//
	// The serverName parameter is the name of the server for which the TLS configuration should be generated.
	// If the serverName parameter is empty, the TLS configuration is generated for an unknown server.
	TLS(serverName string) *tls.Config
	// TlsConfig returns a TLS configuration based on the TLSConfig.
	//
	// The returned TLS configuration is safe for concurrent use.
	//
	// The returned TLS configuration is not a copy of the TLSConfig.
	// Instead, it is a reference to the TLSConfig.
	// Modifying the returned TLS configuration affects the TLSConfig.
	// The returned TLS configuration is the same as the TLSConfig.
	//
	// The serverName parameter is the name of the server for which the TLS configuration is generated.
	// The serverName parameter is used to generate the TLS configuration.
	// The serverName parameter is optional and can be empty.
	// If the serverName parameter is empty, the TLS configuration is generated without a server name.
	TlsConfig(serverName string) *tls.Config
	// Config returns the TLS configuration.
	//
	// The returned TLSConfig is safe for concurrent use.
	//
	// The returned TLSConfig is not a copy of the default TLSConfig.
	// Instead, it is a reference to the default TLSConfig.
	// Modifying the returned TLSConfig affects the default TLSConfig.
	// The returned TLSConfig is the same as the default TLSConfig.
	//
	Config() *Config
}

var Default = New()

// New returns a new TLSConfig with default values.
//
// The returned TLSConfig is safe for concurrent use.
//
// The returned TLSConfig is not a copy of the default TLSConfig.
// Instead, it is a new TLSConfig with default values.
// Modifying the returned TLSConfig does not affect the default TLSConfig.
// The returned TLSConfig is independent of the default TLSConfig.
func New() TLSConfig {
	return &config{
		rand:                  nil,
		cert:                  make([]tlscrt.Cert, 0),
		cipherList:            make([]tlscpr.Cipher, 0),
		curveList:             make([]tlscrv.Curves, 0),
		caRoot:                make([]tlscas.Cert, 0),
		clientAuth:            tlsaut.NoClientCert,
		clientCA:              make([]tlscas.Cert, 0),
		tlsMinVersion:         tlsvrs.VersionTLS12,
		tlsMaxVersion:         tlsvrs.VersionTLS13,
		dynSizingDisabled:     false,
		ticketSessionDisabled: false,
	}
}
