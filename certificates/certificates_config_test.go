/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package certificates_test

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
	"net"
	"os"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	tlscrt "github.com/nabbar/golib/certificates/certs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func genCertififcate() ([]byte, []byte) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).ToNot(HaveOccurred())
	Expect(priv).ToNot(BeNil())

	keyUsage := x509.KeyUsageDigitalSignature
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 365)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	Expect(err).ToNot(HaveOccurred())

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	template.DNSNames = append(template.DNSNames, "example.com")
	template.DNSNames = append(template.DNSNames, "localhost")

	if ip := net.ParseIP("127.0.0.1"); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	Expect(err).ToNot(HaveOccurred())

	bufPub := bytes.NewBuffer(make([]byte, 0))
	err = pem.Encode(bufPub, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	Expect(err).ToNot(HaveOccurred())

	bufKey := bytes.NewBuffer(make([]byte, 0))
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	Expect(err).ToNot(HaveOccurred())

	err = pem.Encode(bufKey, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	Expect(err).ToNot(HaveOccurred())

	return bufPub.Bytes(), bufKey.Bytes()
}

func writeGenCert(pub, key string) {
	p, k := genCertififcate()

	f, e := os.Create(pub)
	Expect(e).ToNot(HaveOccurred())
	_, e = f.Write(p)
	Expect(e).ToNot(HaveOccurred())
	Expect(f.Close()).ToNot(HaveOccurred())

	f, e = os.Create(key)
	Expect(e).ToNot(HaveOccurred())
	_, e = f.Write(k)
	Expect(e).ToNot(HaveOccurred())
	Expect(f.Close()).ToNot(HaveOccurred())
}

func getConfigFile() []byte {
	writeGenCert(pubFile, keyFile)

	str := &tlscrt.ConfigPair{
		Key: keyFile,
		Pub: pubFile,
	}

	p, e := json.Marshal(str)
	Expect(e).ToNot(HaveOccurred())

	var oldConfig string = `{
    "authClient": "none",
    "certs": [` + string(p) + `],
    "cipherList": ["RSA_AES_128_GCM_SHA256", "RSA_AES_256_GCM_SHA384", "ECDHE_RSA_AES_128_GCM_SHA256", "ECDHE_ECDSA_AES_128_GCM_SHA256", "ECDHE_RSA_AES_256_GCM_SHA384", "ECDHE_ECDSA_AES_256_GCM_SHA384", "ECDHE_RSA_CHACHA20_POLY1305_SHA256", "ECDHE_ECDSA_CHACHA20_POLY1305_SHA256", "AES_128_GCM_SHA256", "AES_256_GCM_SHA384", "CHACHA20_POLY1305_SHA256"],
    "clientCA": [],
    "curveList": ["X25519", "P256", "P384", "P521"],
    "dynamicSizingDisable": false,
    "inheritDefault": false,
    "rootCA": [],
    "sessionTicketDisable": false,
    "versionMax": "1.3",
    "versionMin": "1.2"
  }`

	return []byte(oldConfig)
}

func getConfigChain() []byte {
	pub, key := genCertififcate()
	str := "\n" + string(pub) + string(key)
	buf, err := json.Marshal(str)
	Expect(err).ToNot(HaveOccurred())

	var oldConfig string = `{
    "authClient": "none",
    "certs": [` + string(buf) + `],
    "cipherList": ["RSA_AES_128_GCM_SHA256", "RSA_AES_256_GCM_SHA384", "ECDHE_RSA_AES_128_GCM_SHA256", "ECDHE_ECDSA_AES_128_GCM_SHA256", "ECDHE_RSA_AES_256_GCM_SHA384", "ECDHE_ECDSA_AES_256_GCM_SHA384", "ECDHE_RSA_CHACHA20_POLY1305_SHA256", "ECDHE_ECDSA_CHACHA20_POLY1305_SHA256", "AES_128_GCM_SHA256", "AES_256_GCM_SHA384", "CHACHA20_POLY1305_SHA256"],
    "clientCA": [],
    "curveList": ["X25519", "P256", "P384", "P521"],
    "dynamicSizingDisable": false,
    "inheritDefault": false,
    "rootCA": [],
    "sessionTicketDisable": false,
    "versionMax": "1.3",
    "versionMin": "1.2"
  }`

	return []byte(oldConfig)
}

var _ = Describe("certificates test", func() {

	Context("Using a config json", func() {
		It("must success to new Config when certificates are paste as file mode", func() {
			cfg := libtls.Config{}

			p := getConfigFile()
			Expect(len(p)).To(BeNumerically(">", 0))

			e := json.Unmarshal(p, &cfg)
			Expect(e).ToNot(HaveOccurred())

			cnf := cfg.New()
			Expect(cnf).ToNot(BeNil())
			Expect(len(cnf.GetCertificatePair())).To(Equal(1))

			cfgtls := cnf.TLS("localhost")
			Expect(cfgtls).ToNot(BeNil())
			Expect(len(cfgtls.Certificates)).To(Equal(1))
			Expect(len(cfgtls.CipherSuites)).To(BeNumerically(">", 0))
			Expect(len(cfgtls.CurvePreferences)).To(BeNumerically(">", 0))

			p, e = json.Marshal(cnf.Config())
			Expect(e).ToNot(HaveOccurred())
			Expect(len(p)).To(BeNumerically(">", 0))
		})
		It("must success to new Config when certificates are paste as chain mode", func() {
			cfg := libtls.Config{}

			p := getConfigChain()
			Expect(len(p)).To(BeNumerically(">", 0))

			e := json.Unmarshal(p, &cfg)
			Expect(e).ToNot(HaveOccurred())

			cnf := cfg.New()
			Expect(cnf).ToNot(BeNil())
			Expect(len(cnf.GetCertificatePair())).To(Equal(1))

			cfgtls := cnf.TLS("localhost")
			Expect(cfgtls).ToNot(BeNil())
			Expect(len(cfgtls.Certificates)).To(Equal(1))
			Expect(len(cfgtls.CipherSuites)).To(BeNumerically(">", 0))
			Expect(len(cfgtls.CurvePreferences)).To(BeNumerically(">", 0))

			p, e = json.Marshal(cnf)
			Expect(e).ToNot(HaveOccurred())
			Expect(len(p)).To(BeNumerically(">", 0))
		})
	})
})
