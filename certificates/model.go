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

	tlsaut "github.com/nabbar/golib/certificates/auth"
	tlscas "github.com/nabbar/golib/certificates/ca"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"
)

type config struct {
	rand                  io.Reader
	cert                  []tlscrt.Cert
	cipherList            []tlscpr.Cipher
	curveList             []tlscrv.Curves
	caRoot                []tlscas.Cert
	clientAuth            tlsaut.ClientAuth
	clientCA              []tlscas.Cert
	tlsMinVersion         tlsvrs.Version
	tlsMaxVersion         tlsvrs.Version
	dynSizingDisabled     bool
	ticketSessionDisabled bool
}

func (o *config) RegisterRand(rand io.Reader) {
	o.rand = rand
}

func (o *config) SetVersionMin(v tlsvrs.Version) {
	o.tlsMinVersion = v
}

func (o *config) GetVersionMin() tlsvrs.Version {
	return o.tlsMinVersion
}

func (o *config) SetVersionMax(v tlsvrs.Version) {
	o.tlsMaxVersion = v
}

func (o *config) GetVersionMax() tlsvrs.Version {
	return o.tlsMaxVersion
}

func (o *config) SetDynamicSizingDisabled(flag bool) {
	o.dynSizingDisabled = flag
}

func (o *config) SetSessionTicketDisabled(flag bool) {
	o.ticketSessionDisabled = flag
}

func (o *config) Clone() TLSConfig {
	cfg := &config{
		rand:                  o.rand,
		cert:                  make([]tlscrt.Cert, 0),
		cipherList:            make([]tlscpr.Cipher, 0),
		curveList:             make([]tlscrv.Curves, 0),
		caRoot:                make([]tlscas.Cert, 0),
		clientAuth:            o.clientAuth,
		clientCA:              make([]tlscas.Cert, 0),
		tlsMinVersion:         o.tlsMinVersion,
		tlsMaxVersion:         o.tlsMaxVersion,
		dynSizingDisabled:     o.dynSizingDisabled,
		ticketSessionDisabled: o.ticketSessionDisabled,
	}

	if len(o.cert) > 0 {
		for _, c := range o.cert {
			cfg.cert = append(cfg.cert, c)
		}
	}

	if len(o.cipherList) > 0 {
		for _, c := range o.cipherList {
			if tlscpr.Check(c.Uint16()) {
				cfg.cipherList = append(cfg.cipherList, c)
			}
		}
	}

	if len(o.curveList) > 0 {
		for _, c := range o.curveList {
			if tlscpr.Check(c.Uint16()) {
				cfg.curveList = append(cfg.curveList, c)
			}
		}
	}

	if len(o.caRoot) > 0 {
		for _, c := range o.caRoot {
			cfg.caRoot = append(cfg.caRoot, c)
		}
	}

	if len(o.clientCA) > 0 {
		for _, c := range o.clientCA {
			cfg.clientCA = append(cfg.clientCA, c)
		}
	}

	return cfg
}

func (o *config) TlsConfig(serverName string) *tls.Config {
	return o.TLS(serverName)
}

func (o *config) TLS(serverName string) *tls.Config {
	/* #nosec */
	cnf := &tls.Config{
		Rand:                        nil,
		Certificates:                make([]tls.Certificate, 0),
		RootCAs:                     SystemRootCA(),
		ServerName:                  "",
		ClientAuth:                  tls.NoClientCert,
		ClientCAs:                   x509.NewCertPool(),
		InsecureSkipVerify:          false,
		CipherSuites:                make([]uint16, 0),
		SessionTicketsDisabled:      false,
		MinVersion:                  0,
		MaxVersion:                  0,
		CurvePreferences:            make([]tls.CurveID, 0),
		DynamicRecordSizingDisabled: false,
		Renegotiation:               tls.RenegotiateNever,
	}

	if serverName != "" {
		cnf.ServerName = serverName
	}

	if o.ticketSessionDisabled {
		cnf.SessionTicketsDisabled = true
	}

	if o.dynSizingDisabled {
		cnf.DynamicRecordSizingDisabled = true
	}

	if o.tlsMinVersion != tlsvrs.VersionUnknown {
		cnf.MinVersion = o.tlsMinVersion.TLS()
	}

	if o.tlsMaxVersion != tlsvrs.VersionUnknown {
		cnf.MaxVersion = o.tlsMaxVersion.TLS()
	}

	if len(o.cipherList) > 0 {
		for _, c := range o.cipherList {
			if c != tlscpr.Unknown {
				cnf.CipherSuites = append(cnf.CipherSuites, c.TLS())
			}
		}
	}

	if len(o.curveList) > 0 {
		for _, c := range o.curveList {
			if c != tlscrv.Unknown {
				cnf.CurvePreferences = append(cnf.CurvePreferences, c.TLS())
			}
		}
	}

	if o.caRoot != nil {
		for _, c := range o.caRoot {
			c.AppendPool(cnf.RootCAs)
		}
	}

	if len(o.cert) > 0 {
		for _, c := range o.cert {
			cnf.Certificates = append(cnf.Certificates, c.TLS())
		}
	}

	if o.clientAuth != tlsaut.NoClientCert {
		cnf.ClientAuth = o.clientAuth.TLS()
		if len(o.clientCA) > 0 {
			for _, c := range o.clientCA {
				c.AppendPool(cnf.ClientCAs)
			}
		}
	}

	return cnf
}

func (o *config) Config() *Config {
	var crt = make([]tlscrt.Certif, 0)
	for _, c := range o.cert {
		crt = append(crt, c.Model())
	}

	return &Config{
		CurveList:            o.curveList,
		CipherList:           o.cipherList,
		RootCA:               o.caRoot,
		ClientCA:             o.clientCA,
		Certs:                crt,
		VersionMin:           o.tlsMinVersion,
		VersionMax:           o.tlsMaxVersion,
		AuthClient:           o.clientAuth,
		InheritDefault:       false,
		DynamicSizingDisable: o.dynSizingDisabled,
		SessionTicketDisable: o.ticketSessionDisabled,
	}
}
