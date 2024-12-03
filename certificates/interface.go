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

type FctHttpClient func(def TLSConfig, servername string) *http.Client
type FctTLSDefault func() TLSConfig
type FctRootCA func() []string

type TLSConfig interface {
	RegisterRand(rand io.Reader)

	AddRootCAString(rootCA string) bool
	AddRootCAFile(pemFile string) error
	GetRootCA() []tlscas.Cert
	GetRootCAPool() *x509.CertPool

	AddClientCAString(ca string) bool
	AddClientCAFile(pemFile string) error
	GetClientCA() []tlscas.Cert
	GetClientCAPool() *x509.CertPool
	SetClientAuth(a tlsaut.ClientAuth)

	AddCertificatePairString(key, crt string) error
	AddCertificatePairFile(keyFile, crtFile string) error
	LenCertificatePair() int
	CleanCertificatePair()
	GetCertificatePair() []tls.Certificate

	SetVersionMin(v tlsvrs.Version)
	GetVersionMin() tlsvrs.Version
	SetVersionMax(v tlsvrs.Version)
	GetVersionMax() tlsvrs.Version

	SetCipherList(c []tlscpr.Cipher)
	AddCiphers(c ...tlscpr.Cipher)
	GetCiphers() []tlscpr.Cipher

	SetCurveList(c []tlscrv.Curves)
	AddCurves(c ...tlscrv.Curves)
	GetCurves() []tlscrv.Curves

	SetDynamicSizingDisabled(flag bool)
	SetSessionTicketDisabled(flag bool)

	Clone() TLSConfig
	TLS(serverName string) *tls.Config
	TlsConfig(serverName string) *tls.Config
	Config() *Config
}

var Default = New()

func New() TLSConfig {
	return &config{
		rand:                  nil,
		cert:                  make([]tlscrt.Cert, 0),
		cipherList:            make([]tlscpr.Cipher, 0),
		curveList:             make([]tlscrv.Curves, 0),
		caRoot:                make([]tlscas.Cert, 0),
		clientAuth:            tlsaut.NoClientCert,
		clientCA:              make([]tlscas.Cert, 0),
		tlsMinVersion:         tlsvrs.VersionUnknown,
		tlsMaxVersion:         tlsvrs.VersionUnknown,
		dynSizingDisabled:     false,
		ticketSessionDisabled: false,
	}
}

// Deprecated: use local config and no more globals default config.
func AddRootCAContents(rootContent string) bool {
	return Default.AddRootCAString(rootContent)
}

// Deprecated: use local config and no more globals default config.
func AddRootCAFile(rootFile string) error {
	return Default.AddRootCAFile(rootFile)
}

// Deprecated: use local config and no more globals default config.
func AddCACertificateContents(caContent string) bool {
	return Default.AddClientCAString(caContent)
}

// Deprecated: use local config and no more globals default config.
func AddCACertificateFile(caFile string) error {
	return Default.AddClientCAFile(caFile)
}

// Deprecated: use local config and no more globals default config.
func AddCertificatePairString(key, crt string) error {
	return Default.AddCertificatePairString(key, crt)
}

// Deprecated: use local config and no more globals default config.
func AddCertificatePairFile(keyFile, crtFile string) error {
	return Default.AddCertificatePairFile(keyFile, crtFile)
}

// Deprecated: use local config and no more globals default config.
func CheckCertificates() bool {
	return Default.LenCertificatePair() > 0
}

// Deprecated: use local config and no more globals default config.
func GetCertificates() []tls.Certificate {
	return Default.GetCertificatePair()
}

// Deprecated: use local config and no more globals default config.
func AppendCertificates(cert []tls.Certificate) []tls.Certificate {
	if !CheckCertificates() {
		return cert
	}

	return append(cert, GetCertificates()...)
}

// Deprecated: use local config and no more globals default config.
func GetRootCA() *x509.CertPool {
	return Default.GetRootCAPool()
}

// Deprecated: use local config and no more globals default config.
func GetClientCA() *x509.CertPool {
	return Default.GetClientCAPool()
}

// Deprecated: use local config and no more globals default config.
func SetVersionMin(vers uint16) {
	Default.SetVersionMin(tlsvrs.ParseInt(int(vers)))
}

// Deprecated: use local config and no more globals default config.
func SetVersionMax(vers uint16) {
	Default.SetVersionMax(tlsvrs.ParseInt(int(vers)))
}

// Deprecated: use local config and no more globals default config.
func SetClientAuth(auth string) {
	Default.SetClientAuth(tlsaut.Parse(auth))
}

// Deprecated: use local config and no more globals default config.
func SetCipherList(cipher []uint16) {
	Default.SetCipherList(make([]tlscpr.Cipher, 0))

	for _, i := range cipher {
		c := tlscpr.ParseInt(int(i))
		Default.AddCiphers(c)
	}
}

// Deprecated: use local config and no more globals default config.
func SetCurve(curves []tls.CurveID) {
	Default.SetCurveList(make([]tlscrv.Curves, 0))

	for _, i := range curves {
		c := tlscrv.ParseInt(int(i))
		Default.AddCurves(c)
	}
}

// Deprecated: use local config and no more globals default config.
func SetDynamicSizing(enable bool) {
	Default.SetDynamicSizingDisabled(!enable)
}

// Deprecated: use local config and no more globals default config.
func SetSessionTicket(enable bool) {
	Default.SetSessionTicketDisabled(!enable)
}

// Deprecated: use local config and no more globals default config.
func GetTLSConfig(serverName string) *tls.Config {
	return Default.TlsConfig(serverName)
}

// Deprecated: use local config and no more globals default config.
func GetTlsConfigCertificates() *tls.Config {
	return Default.TlsConfig("")
}

// Deprecated: use local config and no more globals default config.
func AddCertificateContents(keyContents, certContents string) error {
	return Default.AddCertificatePairString(keyContents, certContents)
}

// Deprecated: use local config and no more globals default config.
func AddCertificateFile(keyFile, certFile string) error {
	return Default.AddCertificatePairFile(keyFile, certFile)
}
