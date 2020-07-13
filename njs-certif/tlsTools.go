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

package njs_certif

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	. "github.com/nabbar/golib/njs-errors"
)

var (
	rootCA                = x509.NewCertPool()
	certificates          = make([]tls.Certificate, 0)
	caCertificates        = x509.NewCertPool()
	tlsMinVersion  uint16 = tls.VersionTLS12
	tlsMaxVersion  uint16 = tls.VersionTLS12
	cipherList            = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
	curveList     = []tls.CurveID{tls.X25519}
	dynSizing     = true
	ticketSession = true
	clientAuth    = tls.NoClientCert
)

func init() {
	InitRootCA()
}

func InitRootCA() {
	if runtime.GOOS == "windows" {
		rootCA = x509.NewCertPool()
	} else if c, e := x509.SystemCertPool(); e != nil {
		rootCA = c
	} else {
		rootCA = x509.NewCertPool()
	}
}

func AddRootCAContents(rootContent string) bool {
	if rootContent != "" && rootCA != nil {
		return rootCA.AppendCertsFromPEM([]byte(rootContent))
	}

	return false
}

func AddRootCAFile(rootFile string) Error {
	if rootFile == "" {
		return EMPTY_PARAMS.Error(nil)
	}

	if _, e := os.Stat(rootFile); e != nil {
		return FILE_STAT_ERROR.ErrorParent(e)
	}

	c, e := ioutil.ReadFile(rootFile) // #nosec

	if e == nil {
		if rootCA.AppendCertsFromPEM(c) {
			return nil
		}

		return CERT_APPEND_KO.Error(nil)
	}

	return FILE_READ_ERROR.ErrorParent(e)
}

func AddCACertificateContents(caContent string) bool {
	if caContent != "" {
		return caCertificates.AppendCertsFromPEM([]byte(caContent))
	}

	return false
}

func AddCACertificateFile(caFile string) Error {
	if caFile == "" {
		return EMPTY_PARAMS.Error(nil)
	}

	if _, e := os.Stat(caFile); e != nil {
		return FILE_STAT_ERROR.ErrorParent(e)
	}

	c, e := ioutil.ReadFile(caFile) // #nosec

	if e == nil {
		if caCertificates.AppendCertsFromPEM(c) {
			return nil
		}

		return CERT_APPEND_KO.Error(nil)
	}

	return FILE_READ_ERROR.ErrorParent(e)
}

func CheckCertificates() bool {
	return len(certificates) > 0
}

func AddCertificateContents(keyContents, certContents string) Error {
	keyContents = strings.TrimSpace(keyContents)
	certContents = strings.TrimSpace(certContents)

	if keyContents != "" && keyContents != "\n" && certContents != "" && certContents != "\n" {
		c, err := tls.X509KeyPair([]byte(certContents), []byte(keyContents))
		if err == nil {
			certificates = append(certificates, c)
			return nil
		}

		return CERT_PARSE_KEYPAIR.ErrorParent(err)
	}

	return EMPTY_PARAMS.Error(nil)
}

func AddCertificateFile(keyFile, certFile string) Error {
	if keyFile == "" || certFile == "" {
		return EMPTY_PARAMS.Error(nil)
	}

	if _, e := os.Stat(keyFile); e != nil {
		return FILE_STAT_ERROR.ErrorParent(e)
	}

	if _, e := os.Stat(certFile); e != nil {
		return FILE_STAT_ERROR.ErrorParent(e)
	}

	if c, e := tls.LoadX509KeyPair(certFile, keyFile); e == nil {
		certificates = append(certificates, c)
		return nil
	} else {
		return CERT_LOAD_KEYPAIR.ErrorParent(e)
	}
}

func GetCertificates() []tls.Certificate {
	return certificates
}

func AppendCertificates(cert []tls.Certificate) []tls.Certificate {
	if !CheckCertificates() {
		return cert
	}

	return append(cert, certificates...)
}

func GetRootCA() *x509.CertPool {
	return rootCA
}

func GetClientCA() *x509.CertPool {
	return caCertificates
}

func SetVersionMin(vers uint16) {
	tlsMinVersion = vers
}

func SetVersionMax(vers uint16) {
	tlsMaxVersion = vers
}

func SetStringTlsVersion(tlsVersStr string) uint16 {
	tlsVersStr = strings.ToLower(tlsVersStr)
	tlsVersStr = strings.Replace(tlsVersStr, "TLS", "", -1)
	tlsVersStr = strings.TrimSpace(tlsVersStr)

	switch tlsVersStr {
	case "1", "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

func SetClientAuth(auth string) {
	switch strings.ToLower(auth) {
	case "request":
		clientAuth = tls.RequestClientCert
	case "require":
		clientAuth = tls.RequireAnyClientCert
	case "verify":
		clientAuth = tls.VerifyClientCertIfGiven
	case "strict":
		clientAuth = tls.RequireAndVerifyClientCert
	default:
		clientAuth = tls.NoClientCert
	}
}

func GetCipherKey(cipher string) uint16 {
	cipher = strings.ToLower(cipher)

	RSA := strings.Contains(cipher, "rsa")
	DSA := strings.Contains(cipher, "ecdsa")
	GSM := strings.Contains(cipher, "gsm")
	AES128 := strings.Contains(cipher, "aes128") || strings.Contains(cipher, "aes_128") || strings.Contains(cipher, "aes-128")
	AES256 := strings.Contains(cipher, "aes256") || strings.Contains(cipher, "aes_256") || strings.Contains(cipher, "aes-256")
	CHACHA := strings.Contains(cipher, "chacha20") || strings.Contains(cipher, "chacha_20") || strings.Contains(cipher, "chacha-20")
	POLY := strings.Contains(cipher, "poly1305") || strings.Contains(cipher, "poly_1305") || strings.Contains(cipher, "poly-1305")

	if RSA && AES128 && !GSM {
		return tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
	} else if RSA && AES128 && GSM {
		return tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	} else if RSA && AES256 {
		return tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	} else if RSA && (CHACHA || POLY) {
		return tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
	} else if DSA && AES128 && !GSM {
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
	} else if DSA && AES128 && GSM {
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	} else if DSA && AES256 {
		return tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	} else if DSA && (CHACHA || POLY) {
		return tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
	}

	return 0
}

func GetCurveID(curveRef string) tls.CurveID {
	curveRef = strings.ToLower(curveRef)

	if strings.Contains(curveRef, "p256") {
		return tls.CurveP256
	} else if strings.Contains(curveRef, "p384") {
		return tls.CurveP384
	} else if strings.Contains(curveRef, "p521") {
		return tls.CurveP521
	} else if strings.Contains(curveRef, "x25519") {
		return tls.X25519
	}

	return 0
}

func SetCipherList(cipher []uint16) {
	cipherList = cipher
}

func SetCurve(curves []tls.CurveID) {
	curveList = curves
}

func SetDynamicSizing(enable bool) {
	dynSizing = enable
}

func SetSessionTicket(enable bool) {
	ticketSession = enable
}

func GetTLSConfig(serverName string) *tls.Config {
	cnf := &tls.Config{
		RootCAs:            rootCA,
		ClientCAs:          caCertificates,
		InsecureSkipVerify: false,
	}

	if tlsMinVersion > 0 {
		cnf.MinVersion = tlsMinVersion
	}

	if tlsMaxVersion > 0 {
		cnf.MaxVersion = tlsMaxVersion
	}

	if serverName != "" {
		cnf.ServerName = serverName
	}

	if len(cipherList) > 0 {
		cnf.PreferServerCipherSuites = true
		cnf.CipherSuites = cipherList
	}

	if len(curveList) > 0 {
		cnf.CurvePreferences = curveList
	}

	if dynSizing {
		cnf.DynamicRecordSizingDisabled = false
	} else {
		cnf.DynamicRecordSizingDisabled = true
	}

	if ticketSession {
		cnf.SessionTicketsDisabled = false
	} else {
		cnf.SessionTicketsDisabled = true
	}

	return cnf
}

func GetTlsConfigCertificates() *tls.Config {
	cnf := GetTLSConfig("")

	if clientAuth != tls.NoClientCert {
		cnf.ClientAuth = clientAuth
	}

	if len(cnf.Certificates) > 0 {
		cnf.Certificates = certificates
	}

	return cnf
}
