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

package cipher_test

import (
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	. "github.com/nabbar/golib/certificates/cipher"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var _ = Describe("cipher", func() {
	It("Parse should recognize known names and retro aliases", func() {
		Expect(Parse("TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")).ToNot(Equal(Unknown))
		Expect(Parse("ECDHE_RSA_AES_128_GCM_SHA256")).ToNot(Equal(Unknown))
		Expect(Parse("ecdhe-rsa-aes-128-gcm-sha256")).ToNot(Equal(Unknown))
		Expect(Parse("tls.chacha20_poly1305_sha256 ecdhe ecdsa")).ToNot(Equal(Unknown))
		Expect(Parse("unknown_cipher")).To(Equal(Unknown))
	})

	It("String/Code should be consistent and non-empty for known ciphers", func() {
		c := TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		Expect(c.String()).To(ContainSubstring("ecdhe"))
		Expect(len(c.Code())).To(BeNumerically(">", 0))
	})

	It("Marshal/Unmarshal JSON/YAML/TOML/CBOR/Text roundtrip", func() {
		var (
			e error
			c = TLS_AES_256_GCM_SHA384
			b []byte
		)

		// JSON
		b, e = json.Marshal(c)
		Expect(e).ToNot(HaveOccurred())
		var c2 Cipher
		Expect(json.Unmarshal(b, &c2)).To(Succeed())
		Expect(c2).To(Equal(c))

		// YAML
		b, e = yaml.Marshal(c)
		Expect(e).ToNot(HaveOccurred())
		var c3 Cipher
		Expect(yaml.Unmarshal(b, &c3)).To(Succeed())
		Expect(c3).To(Equal(c))

		// CBOR
		b, e = cbor.Marshal(c)
		Expect(e).ToNot(HaveOccurred())
		var c5 Cipher
		Expect(cbor.Unmarshal(b, &c5)).To(Succeed())
		Expect(c5).To(Equal(c))

		// Text
		txt, err := c.MarshalText()
		Expect(err).ToNot(HaveOccurred())
		var c6 Cipher
		Expect(c6.UnmarshalText(txt)).To(Succeed())
		Expect(c6).To(Equal(c))
	})

	It("Numeric helpers should map back to same value for known cipher", func() {
		c := TLS_CHACHA20_POLY1305_SHA256
		Expect(ParseInt(int(c.Uint16()))).To(Equal(c))
		Expect(Check(c.Uint16())).To(BeTrue())
	})

	It("JSON embed roundtrip via struct", func() {
		type wrap struct {
			C Cipher `json:"c"`
		}
		w := wrap{C: TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384}
		b, err := json.Marshal(w)
		Expect(err).ToNot(HaveOccurred())
		var w2 wrap
		Expect(json.Unmarshal(b, &w2)).To(Succeed())
		Expect(w2.C).To(Equal(w.C))
	})

	It("TOML embed roundtrip via struct", func() {
		type wrap struct {
			C Cipher `json:"c"`
		}
		w := wrap{C: TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384}
		b, err := toml.Marshal(w)
		Expect(err).ToNot(HaveOccurred())
		var w2 wrap
		Expect(toml.Unmarshal(b, &w2)).To(Succeed())
		Expect(w2.C).To(Equal(w.C))
	})
})
