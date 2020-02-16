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
	"strings"

	logger "github.com/nabbar/golib/njs-logger"
)

var (
	rootCA                = x509.NewCertPool()
	certificates          = make([]tls.Certificate, 0)
	caCertificates        = x509.NewCertPool()
	tlsMinVersion  uint16 = tls.VersionTLS12
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
	if c, e := x509.SystemCertPool(); e != nil {
		rootCA = c
	} else {
		rootCA = x509.NewCertPool()
	}
}

func AddRootCAContents(rootContent string) bool {
	if rootContent != "" {
		return rootCA.AppendCertsFromPEM([]byte(rootContent))
	}

	return false
}

func AddRootCAFile(rootFile string) bool {
	if rootFile == "" {
		return false
	}

	if _, e := os.Stat(rootFile); logger.ErrorLevel.LogErrorCtxf(logger.InfoLevel, "checking certificates file '%s'", e, rootFile) {
		return false
	}

	c, e := ioutil.ReadFile(rootFile) // #nosec
	if !logger.ErrorLevel.LogErrorCtxf(logger.InfoLevel, "loading certificates file '%s'", e, rootFile) {
		return rootCA.AppendCertsFromPEM(c)
	}

	return false
}

func AddCACertificateContents(caContent string) bool {
	if caContent != "" {
		return caCertificates.AppendCertsFromPEM([]byte(caContent))
	}

	return false
}

func AddCACertificateFile(caFile string) bool {
	if caFile == "" {
		return false
	}

	if _, e := os.Stat(caFile); logger.ErrorLevel.LogErrorCtxf(logger.InfoLevel, "checking certificates file '%s'", e, caFile) {
		return false
	}

	c, e := ioutil.ReadFile(caFile) // #nosec
	if !logger.ErrorLevel.LogErrorCtxf(logger.InfoLevel, "loading certificates file '%s'", e, caFile) {
		return caCertificates.AppendCertsFromPEM(c)
	}

	return false
}

func CheckCertificates() bool {
	return len(certificates) > 0
}

func AddCertificateContents(keyContents, certContents string) bool {
	keyContents = strings.TrimSpace(keyContents)
	certContents = strings.TrimSpace(certContents)

	if keyContents != "" && keyContents != "\n" && certContents != "" && certContents != "\n" {
		c, err := tls.X509KeyPair([]byte(certContents), []byte(keyContents))
		if !logger.ErrorLevel.LogErrorCtx(logger.InfoLevel, "loading certificates contents", err) {
			certificates = append(certificates, c)
			return true
		}
	}

	return false
}

func AddCertificateFile(keyFile, certFile string) bool {
	if keyFile == "" || certFile == "" {
		return false
	}

	if _, e := os.Stat(keyFile); logger.ErrorLevel.LogErrorCtxf(logger.InfoLevel, "loading certificates file '%s'", e, keyFile) {
		return false
	}

	if _, e := os.Stat(certFile); logger.ErrorLevel.LogErrorCtxf(logger.InfoLevel, "loading certificates file '%s'", e, certFile) {
		return false
	}

	if c, e := tls.LoadX509KeyPair(certFile, keyFile); !logger.ErrorLevel.LogErrorCtx(logger.InfoLevel, "loading X509 Pair file", e) {
		certificates = append(certificates, c)
		return true
	}

	return false
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

func SetStringTlsVersion(tlsVersStr string) {
	tlsVersStr = strings.ToLower(tlsVersStr)
	tlsVersStr = strings.Replace(tlsVersStr, "TLS", "", -1)
	tlsVersStr = strings.TrimSpace(tlsVersStr)

	switch tlsVersStr {
	case "1", "1.0":
		tlsMinVersion = tls.VersionTLS10
	case "1.1":
		tlsMinVersion = tls.VersionTLS11
	default:
		tlsMinVersion = tls.VersionTLS12
	}
}

func SetTlsVersion(tlsVers uint16) {
	switch tlsVers {
	case tls.VersionTLS10:
		tlsMinVersion = tls.VersionTLS10
	case tls.VersionTLS11:
		tlsMinVersion = tls.VersionTLS11
	default:
		tlsMinVersion = tls.VersionTLS12
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
		MinVersion:         tlsMinVersion,
		InsecureSkipVerify: false,
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

	cnf.Certificates = certificates
	cnf.BuildNameToCertificate()

	return cnf
}
