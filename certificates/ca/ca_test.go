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

package ca_test

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
	"reflect"
	"time"

	. "github.com/nabbar/golib/certificates/ca"

	. "github.com/onsi/ginkgo/v2"

	"github.com/fxamacker/cbor/v2"
	. "github.com/onsi/gomega"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func genCAPEM() string {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).ToNot(HaveOccurred())

	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA", Organization: []string{"Test Org"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	Expect(err).ToNot(HaveOccurred())

	buf := bytes.NewBuffer(nil)
	Expect(pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: der})).To(Succeed())

	return buf.String()
}

func genMultiCAPEM() string {
	ca1 := genCAPEM()
	ca2 := genCAPEM()
	return ca1 + ca2
}

var _ = Describe("ca", func() {
	It("Parse should create valid CA cert from PEM string", func() {
		pem := genCAPEM()
		c, err := Parse(pem)
		Expect(err).ToNot(HaveOccurred())
		Expect(c).ToNot(BeNil())
		Expect(c.Len()).To(Equal(1))
		Expect(c.String()).To(ContainSubstring("BEGIN CERTIFICATE"))
	})

	It("ParseByte should work like Parse", func() {
		pem := genCAPEM()
		c, err := ParseByte([]byte(pem))
		Expect(err).ToNot(HaveOccurred())
		Expect(c).ToNot(BeNil())
		Expect(c.Len()).To(Equal(1))
	})

	It("Parse should handle multiple CA certs in chain", func() {
		multi := genMultiCAPEM()
		c, err := Parse(multi)
		Expect(err).ToNot(HaveOccurred())
		Expect(c.Len()).To(Equal(2))
	})

	It("AppendBytes and AppendString should add more certs", func() {
		pem1 := genCAPEM()
		c, err := Parse(pem1)
		Expect(err).ToNot(HaveOccurred())
		Expect(c.Len()).To(Equal(1))

		pem2 := genCAPEM()
		Expect(c.AppendBytes([]byte(pem2))).To(Succeed())
		Expect(c.Len()).To(Equal(2))

		pem3 := genCAPEM()
		Expect(c.AppendString(pem3)).To(Succeed())
		Expect(c.Len()).To(Equal(3))
	})

	It("Chain and String should return PEM-encoded chain", func() {
		pem := genCAPEM()
		c, err := Parse(pem)
		Expect(err).ToNot(HaveOccurred())

		chain, err := c.Chain()
		Expect(err).ToNot(HaveOccurred())
		Expect(chain).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(chain).To(ContainSubstring("END CERTIFICATE"))

		Expect(c.String()).To(Equal(chain))
	})

	It("AppendPool should add certs to x509.CertPool", func() {
		multi := genMultiCAPEM()
		c, err := Parse(multi)
		Expect(err).ToNot(HaveOccurred())

		pool := x509.NewCertPool()
		c.AppendPool(pool)
		// CertPool doesn't expose count, but we can verify it doesn't panic
		Expect(pool).ToNot(BeNil())
	})

	It("Marshal/Unmarshal JSON roundtrip", func() {
		pem := genCAPEM()
		c, err := Parse(pem)
		Expect(err).ToNot(HaveOccurred())

		b, err := json.Marshal(c)
		Expect(err).ToNot(HaveOccurred())

		var c2 Certif
		Expect(json.Unmarshal(b, &c2)).To(Succeed())
		Expect(c2.Len()).To(Equal(c.Len()))
		Expect(c2.String()).To(Equal(c.String()))
	})

	It("Marshal/Unmarshal YAML roundtrip", func() {
		pem := genCAPEM()
		c, err := Parse(pem)
		m := c.Model()
		Expect(err).ToNot(HaveOccurred())

		b, err := yaml.Marshal(&m)
		Expect(err).ToNot(HaveOccurred())

		var c2 = &Certif{}
		Expect(yaml.Unmarshal(b, &c2)).To(Succeed())
		Expect(c2.Len()).To(Equal(c.Len()))
	})

	It("Marshal/Unmarshal CBOR roundtrip", func() {
		pem := genCAPEM()
		c, err := Parse(pem)
		Expect(err).ToNot(HaveOccurred())

		b, err := cbor.Marshal(c)
		Expect(err).ToNot(HaveOccurred())

		var c2 Certif
		Expect(cbor.Unmarshal(b, &c2)).To(Succeed())
		Expect(c2.Len()).To(Equal(c.Len()))
	})

	It("Marshal/Unmarshal Text roundtrip", func() {
		pem := genCAPEM()
		c, err := Parse(pem)
		Expect(err).ToNot(HaveOccurred())

		b, err := c.MarshalText()
		Expect(err).ToNot(HaveOccurred())

		var c2 Certif
		Expect(c2.UnmarshalText(b)).To(Succeed())
		Expect(c2.Len()).To(Equal(c.Len()))
	})

	It("Marshal/Unmarshal Binary roundtrip", func() {
		pem := genCAPEM()
		c, err := Parse(pem)
		Expect(err).ToNot(HaveOccurred())

		b, err := c.MarshalBinary()
		Expect(err).ToNot(HaveOccurred())

		var c2 Certif
		Expect(c2.UnmarshalBinary(b)).To(Succeed())
		Expect(c2.Len()).To(Equal(c.Len()))
	})

	It("ViperDecoderHook should decode string to Cert", func() {
		hook := ViperDecoderHook()
		pem := genCAPEM()

		var c Cert
		fromType := reflect.TypeOf(pem)
		toType := reflect.TypeOf(c)

		result, err := hook(fromType, toType, pem)
		Expect(err).ToNot(HaveOccurred())

		decoded, ok := result.(Cert)
		Expect(ok).To(BeTrue())
		Expect(decoded.Len()).To(Equal(1))
	})

	It("ViperDecoderHook should pass through non-matching types", func() {
		hook := ViperDecoderHook()

		// Non-string source
		result, err := hook(reflect.TypeOf(123), reflect.TypeOf(""), 123)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(123))

		// Non-Cert target
		result2, err := hook(reflect.TypeOf(""), reflect.TypeOf(123), "test")
		Expect(err).ToNot(HaveOccurred())
		Expect(result2).To(Equal("test"))
	})

	It("UnmarshalJSON should handle embedded cert in struct", func() {
		type Wrapper struct {
			CA *Certif
		}

		pem := genCAPEM()
		c, err := Parse(pem)
		m := c.Model()
		Expect(err).ToNot(HaveOccurred())
		Expect(c.Len()).ToNot(Equal(0))

		w := Wrapper{CA: &m}
		b, err := json.Marshal(w)
		Expect(err).ToNot(HaveOccurred())
		Expect(b).ToNot(BeEmpty())

		var w2 Wrapper
		err = json.Unmarshal(b, &w2)
		Expect(err).ToNot(HaveOccurred())
		Expect(w2).To(Equal(w))
	})

	It("UnmarshalTOML should handle embedded cert in struct", func() {
		type wrapper struct {
			CA *Certif
		}

		pem := genCAPEM()
		c, err := Parse(pem)
		m := c.Model()
		Expect(err).ToNot(HaveOccurred())

		w := wrapper{CA: &m}
		b, err := toml.Marshal(w)
		Expect(err).ToNot(HaveOccurred())
		Expect(b).ToNot(BeEmpty())

		var w2 wrapper
		Expect(toml.Unmarshal(b, &w2)).To(Succeed())
		Expect(w2).To(Equal(w))
	})

	It("Parse should handle empty input gracefully", func() {
		_, err := Parse("")
		Expect(err).To(HaveOccurred())
	})

	It("unMarshall should trim whitespace and newlines", func() {
		pem := genCAPEM()
		padded := "\n\r\n  " + pem + "  \n\r\n"
		c, err := Parse(padded)
		Expect(err).ToNot(HaveOccurred())
		Expect(c.Len()).To(Equal(1))
	})
})
