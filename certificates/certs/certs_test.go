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

package certs_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/fxamacker/cbor/v2"
	. "github.com/nabbar/golib/certificates/certs"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

func genPairPEM() (pub string, key string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).ToNot(HaveOccurred())

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	Expect(err).ToNot(HaveOccurred())

	bufPub := bytes.NewBuffer(nil)
	Expect(pem.Encode(bufPub, &pem.Block{Type: "CERTIFICATE", Bytes: der})).To(Succeed())

	pk, err := x509.MarshalPKCS8PrivateKey(priv)
	Expect(err).ToNot(HaveOccurred())
	bufKey := bytes.NewBuffer(nil)
	Expect(pem.Encode(bufKey, &pem.Block{Type: "PRIVATE KEY", Bytes: pk})).To(Succeed())

	return bufPub.String(), bufKey.String()
}

var _ = Describe("certs", func() {
	It("ParsePair and methods should work", func() {
		pub, key := genPairPEM()
		c, err := ParsePair(key, pub)
		Expect(err).ToNot(HaveOccurred())
		Expect(c).ToNot(BeNil())
		Expect(c.IsPair()).To(BeTrue())
		Expect(c.IsChain()).To(BeFalse())

		tlsC := c.TLS()
		Expect(len(tlsC.Certificate)).To(BeNumerically(">", 0))

		s := c.String()
		Expect(s).To(ContainSubstring("BEGIN CERTIFICATE"))

		cp := c.Model()
		Expect(cp.IsPair()).To(BeTrue())

		// Pair/Chain
		chain, err := c.Chain()
		Expect(err).ToNot(HaveOccurred())
		Expect(chain).To(ContainSubstring("BEGIN CERTIFICATE"))
		p2, k2, err := c.Pair()
		Expect(err).ToNot(HaveOccurred())
		Expect(p2).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(k2).To(ContainSubstring("PRIVATE KEY"))
	})

	It("JSON/YAML/TOML/CBOR/Text/Binary roundtrips", func() {
		var (
			pub, key = genPairPEM()
		)

		c, err := ParsePair(key, pub)
		Expect(err).ToNot(HaveOccurred())

		// JSON
		b, err := json.Marshal(c)
		Expect(err).ToNot(HaveOccurred())
		var m2 Certif
		Expect(json.Unmarshal(b, &m2)).To(Succeed())
		Expect(len(m2.TLS().Certificate)).To(BeNumerically(">", 0))

		// YAML
		y, err := yaml.Marshal(c)
		Expect(err).ToNot(HaveOccurred())
		var m3 Certif
		Expect(yaml.Unmarshal(y, &m3)).To(Succeed())
		Expect(len(m3.TLS().Certificate)).To(BeNumerically(">", 0))

		// TOML
		t, err := toml.Marshal(c)
		Expect(err).ToNot(HaveOccurred())
		var m4 Certif
		Expect(toml.Unmarshal(t, &m4)).To(Succeed())
		Expect(len(m4.TLS().Certificate)).To(BeNumerically(">", 0))

		// CBOR
		cb, err := cbor.Marshal(c)
		Expect(err).ToNot(HaveOccurred())
		var m5 Certif
		Expect(cbor.Unmarshal(cb, &m5)).To(Succeed())
		Expect(len(m5.TLS().Certificate)).To(BeNumerically(">", 0))

		// Text
		txt, err := c.MarshalText()
		Expect(err).ToNot(HaveOccurred())
		var m6 Certif
		Expect(m6.UnmarshalText(txt)).To(Succeed())

		// Binary
		bin, err := c.MarshalBinary()
		Expect(err).ToNot(HaveOccurred())
		var m7 Certif
		Expect(m7.UnmarshalBinary(bin)).To(Succeed())
	})
})
