/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

// extract of src/crypto/tls/generate_cert.go

func pubKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

func getCert(priv any) ([]byte, error) {
	var (
		err error
		now = time.Now()
		tpl *x509.Certificate
		ser *big.Int
		srl = new(big.Int).Lsh(big.NewInt(1), 128)
	)

	if ser, err = rand.Int(rand.Reader, srl); err != nil {
		return nil, err
	}

	tpl = &x509.Certificate{
		SerialNumber: ser,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           append(make([]net.IP, 0), net.ParseIP("127.0.0.1"), net.ParseIP("::1")),
		DNSNames:              append(make([]string, 0), "localhost"),
	}

	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		tpl.KeyUsage |= x509.KeyUsageKeyEncipherment
	}

	return x509.CreateCertificate(rand.Reader, tpl, tpl, pubKey(priv), priv)
}

func genTLSCertificate(ed bool) (tls.Certificate, error) {
	var (
		err error

		priv any
		cert []byte
		pvbt []byte

		bufKey = bytes.NewBuffer(make([]byte, 0))
		bufCrt = bytes.NewBuffer(make([]byte, 0))
	)

	if ed {
		_, priv, err = ed25519.GenerateKey(rand.Reader)
	} else {
		priv, err = rsa.GenerateKey(rand.Reader, 4096)
	}

	if err != nil {
		panic(err)
	} else if cert, err = getCert(priv); err != nil {
		panic(err)
	} else if err = pem.Encode(bufCrt, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
		panic(err)
	} else if pvbt, err = x509.MarshalPKCS8PrivateKey(priv); err != nil {
		panic(err)
	} else if err = pem.Encode(bufKey, &pem.Block{Type: "PRIVATE KEY", Bytes: pvbt}); err != nil {
		panic(err)
	}

	return tls.X509KeyPair(bufCrt.Bytes(), bufKey.Bytes())
}
