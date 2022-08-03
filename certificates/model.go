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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"strings"

	liberr "github.com/nabbar/golib/errors"
)

type config struct {
	cert                  []tls.Certificate
	cipherList            []uint16
	curveList             []tls.CurveID
	caRoot                *x509.CertPool
	clientAuth            tls.ClientAuthType
	clientCA              *x509.CertPool
	tlsMinVersion         uint16
	tlsMaxVersion         uint16
	dynSizingDisabled     bool
	ticketSessionDisabled bool
}

func (c *config) checkFile(pemFiles ...string) liberr.Error {
	for _, f := range pemFiles {
		if f == "" {
			return ErrorParamsEmpty.Error(nil)
		}

		if _, e := os.Stat(f); e != nil {
			return ErrorFileStat.ErrorParent(e)
		}

		/* #nosec */
		b, e := ioutil.ReadFile(f)
		if e != nil {
			return ErrorFileRead.ErrorParent(e)
		}

		b = bytes.Trim(b, "\n")
		b = bytes.Trim(b, "\r")
		b = bytes.TrimSpace(b)

		if len(b) < 1 {
			return ErrorFileEmpty.Error(nil)
		}
	}

	return nil
}

func (c *config) AddRootCAString(rootCA string) bool {
	if c.caRoot == nil {
		c.caRoot = SystemRootCA()
	}

	if rootCA != "" {
		return c.caRoot.AppendCertsFromPEM([]byte(rootCA))
	}

	return false
}

func (c *config) AddRootCAFile(pemFile string) liberr.Error {
	if e := c.checkFile(pemFile); e != nil {
		return e
	}

	if c.caRoot == nil {
		c.caRoot = SystemRootCA()
	}

	//nolint #nosec
	/* #nosec */
	b, _ := ioutil.ReadFile(pemFile)

	if c.caRoot.AppendCertsFromPEM(b) {
		return nil
	}

	return ErrorCertAppend.Error(nil)
}

func (c *config) AddClientCAString(ca string) bool {
	if c.clientCA == nil {
		c.clientCA = x509.NewCertPool()
	}

	if ca != "" {
		return c.clientCA.AppendCertsFromPEM([]byte(ca))
	}

	return false
}

func (c *config) AddClientCAFile(pemFile string) liberr.Error {
	if e := c.checkFile(pemFile); e != nil {
		return e
	}

	if c.clientCA == nil {
		c.clientCA = x509.NewCertPool()
	}

	//nolint #nosec
	/* #nosec */
	b, _ := ioutil.ReadFile(pemFile)

	if c.clientCA.AppendCertsFromPEM(b) {
		return nil
	}

	return ErrorCertAppend.Error(nil)
}

func (c *config) AddCertificatePairString(key, crt string) liberr.Error {
	if len(c.cert) == 0 {
		c.cert = make([]tls.Certificate, 0)
	}

	key = strings.Trim(key, "\n")
	crt = strings.Trim(crt, "\n")

	key = strings.Trim(key, "\r")
	crt = strings.Trim(crt, "\r")

	key = strings.TrimSpace(key)
	crt = strings.TrimSpace(crt)

	if len(key) < 1 || len(crt) < 1 {
		return ErrorParamsEmpty.Error(nil)
	}

	p, err := tls.X509KeyPair([]byte(crt), []byte(key))
	if err != nil {
		return ErrorCertKeyPairParse.ErrorParent(err)
	}

	c.cert = append(c.cert, p)
	return nil
}

func (c *config) AddCertificatePairFile(keyFile, crtFile string) liberr.Error {
	if e := c.checkFile(keyFile, crtFile); e != nil {
		return e
	}

	if len(c.cert) == 0 {
		c.cert = make([]tls.Certificate, 0)
	}

	if p, e := tls.LoadX509KeyPair(crtFile, keyFile); e != nil {
		return ErrorCertKeyPairLoad.ErrorParent(e)
	} else {
		c.cert = append(c.cert, p)
		return nil
	}
}

func (c *config) TlsConfig(serverName string) *tls.Config {
	/* #nosec */
	cnf := &tls.Config{
		InsecureSkipVerify: false,
	}

	if serverName != "" {
		cnf.ServerName = serverName
	}

	if c.ticketSessionDisabled {
		cnf.SessionTicketsDisabled = true
	}

	if c.dynSizingDisabled {
		cnf.DynamicRecordSizingDisabled = true
	}

	if c.tlsMinVersion != 0 {
		cnf.MinVersion = c.tlsMinVersion
	}

	if c.tlsMaxVersion != 0 {
		cnf.MaxVersion = c.tlsMaxVersion
	}

	if len(c.cipherList) > 0 {
		cnf.PreferServerCipherSuites = true
		cnf.CipherSuites = c.cipherList
	}

	if len(c.curveList) > 0 {
		cnf.CurvePreferences = c.curveList
	}

	if c.caRoot != nil {
		cnf.RootCAs = c.caRoot
	}

	if len(c.cert) > 0 {
		cnf.Certificates = c.cert
	}

	if c.clientAuth != tls.NoClientCert {
		cnf.ClientAuth = c.clientAuth
		if c.clientCA != nil {
			cnf.ClientCAs = c.clientCA
		}
	}

	return cnf
}

func (c *config) cloneCipherList() []uint16 {
	if c.cipherList == nil {
		return nil
	}

	return append(make([]uint16, 0), c.cipherList...)
}

func (c *config) cloneCurveList() []tls.CurveID {
	if c.curveList == nil {
		return nil
	}

	return append(make([]tls.CurveID, 0), c.curveList...)
}

func (c *config) cloneCertificates() []tls.Certificate {
	if c.cert == nil {
		return nil
	}

	return append(make([]tls.Certificate, 0), c.cert...)
}

func (c *config) cloneRootCA() *x509.CertPool {
	if c.caRoot == nil {
		return nil
	}

	list := *c.caRoot

	return &list
}

func (c *config) cloneClientCA() *x509.CertPool {
	if c.clientCA == nil {
		return nil
	}

	list := *c.clientCA

	return &list
}

func (c *config) Clone() TLSConfig {
	return &config{
		caRoot:                c.cloneRootCA(),
		cert:                  c.cloneCertificates(),
		tlsMinVersion:         c.tlsMinVersion,
		tlsMaxVersion:         c.tlsMaxVersion,
		cipherList:            c.cloneCipherList(),
		curveList:             c.cloneCurveList(),
		dynSizingDisabled:     c.dynSizingDisabled,
		ticketSessionDisabled: c.ticketSessionDisabled,
		clientAuth:            c.clientAuth,
		clientCA:              c.cloneClientCA(),
	}
}

func asStruct(cfg TLSConfig) *config {
	if c, ok := cfg.(*config); ok {
		return c
	}

	return nil
}

func (c *config) GetRootCA() *x509.CertPool {
	return c.caRoot
}

func (c *config) GetClientCA() *x509.CertPool {
	return c.clientCA
}

func (c *config) LenCertificatePair() int {
	return len(c.cert)
}

func (c *config) CleanCertificatePair() {
	c.cert = make([]tls.Certificate, 0)
}

func (c *config) GetCertificatePair() []tls.Certificate {
	return c.cert
}

func (c *config) SetVersionMin(vers uint16) {
	c.tlsMinVersion = vers
}

func (c *config) SetVersionMax(vers uint16) {
	c.tlsMaxVersion = vers
}

func (c *config) SetClientAuth(cAuth tls.ClientAuthType) {
	c.clientAuth = cAuth
}

func (c *config) SetCipherList(cipher []uint16) {
	c.cipherList = cipher
}

func (c *config) SetCurveList(curves []tls.CurveID) {
	c.curveList = curves
}

func (c *config) SetDynamicSizingDisabled(flag bool) {
	c.dynSizingDisabled = flag
}

func (c *config) SetSessionTicketDisabled(flag bool) {
	c.ticketSessionDisabled = flag
}
