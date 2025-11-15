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

package auth_test

import (
	"crypto/tls"
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	. "github.com/nabbar/golib/certificates/auth"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var _ = Describe("auth", func() {
	It("Parse should recognize client auth modes from string", func() {
		Expect(Parse("strict")).To(Equal(RequireAndVerifyClientCert))
		Expect(Parse("require verify")).To(Equal(RequireAndVerifyClientCert))
		Expect(Parse("REQUIRE AND VERIFY")).To(Equal(RequireAndVerifyClientCert))
		Expect(Parse("verify")).To(Equal(VerifyClientCertIfGiven))
		Expect(Parse("VerifyClientCertIfGiven")).To(Equal(VerifyClientCertIfGiven))
		Expect(Parse("require")).To(Equal(RequireAnyClientCert))
		Expect(Parse("request")).To(Equal(RequestClientCert))
		Expect(Parse("none")).To(Equal(NoClientCert))
		Expect(Parse("unknown")).To(Equal(NoClientCert))
		Expect(Parse("")).To(Equal(NoClientCert))
	})

	It("Parse should handle quoted strings and mixed case", func() {
		Expect(Parse("\"strict\"")).To(Equal(RequireAndVerifyClientCert))
		Expect(Parse("'verify'")).To(Equal(VerifyClientCertIfGiven))
		Expect(Parse("  REQUIRE  ")).To(Equal(RequireAnyClientCert))
	})

	It("ParseInt should map tls.ClientAuthType values", func() {
		Expect(ParseInt(int(tls.RequireAndVerifyClientCert))).To(Equal(RequireAndVerifyClientCert))
		Expect(ParseInt(int(tls.VerifyClientCertIfGiven))).To(Equal(VerifyClientCertIfGiven))
		Expect(ParseInt(int(tls.RequireAnyClientCert))).To(Equal(RequireAnyClientCert))
		Expect(ParseInt(int(tls.RequestClientCert))).To(Equal(RequestClientCert))
		Expect(ParseInt(int(tls.NoClientCert))).To(Equal(NoClientCert))
		Expect(ParseInt(999)).To(Equal(NoClientCert))
	})

	It("String and Code should format correctly", func() {
		Expect(RequireAndVerifyClientCert.String()).To(ContainSubstring("strict"))
		Expect(VerifyClientCertIfGiven.String()).To(Equal("verify"))
		Expect(RequireAnyClientCert.String()).To(Equal("require"))
		Expect(RequestClientCert.String()).To(Equal("request"))
		Expect(NoClientCert.String()).To(Equal("none"))

		Expect(RequireAndVerifyClientCert.Code()).To(ContainSubstring("strict"))
	})

	It("TLS should return tls.ClientAuthType", func() {
		Expect(RequireAndVerifyClientCert.TLS()).To(Equal(tls.RequireAndVerifyClientCert))
		Expect(VerifyClientCertIfGiven.TLS()).To(Equal(tls.VerifyClientCertIfGiven))
		Expect(RequireAnyClientCert.TLS()).To(Equal(tls.RequireAnyClientCert))
		Expect(RequestClientCert.TLS()).To(Equal(tls.RequestClientCert))
		Expect(NoClientCert.TLS()).To(Equal(tls.NoClientCert))
	})

	It("List should return all ClientAuth values", func() {
		list := List()
		Expect(list).To(HaveLen(5))
		Expect(list).To(ContainElement(NoClientCert))
		Expect(list).To(ContainElement(RequestClientCert))
		Expect(list).To(ContainElement(RequireAnyClientCert))
		Expect(list).To(ContainElement(VerifyClientCertIfGiven))
		Expect(list).To(ContainElement(RequireAndVerifyClientCert))
	})

	It("Marshal/Unmarshal JSON roundtrip", func() {
		type wrapper struct {
			Auth ClientAuth `json:"auth"`
		}

		w := wrapper{Auth: RequireAndVerifyClientCert}
		b, err := json.Marshal(w)
		Expect(err).ToNot(HaveOccurred())

		var w2 wrapper
		Expect(json.Unmarshal(b, &w2)).To(Succeed())
		Expect(w2.Auth).To(Equal(w.Auth))
	})

	It("Marshal/Unmarshal YAML roundtrip", func() {
		type wrapper struct {
			Auth ClientAuth `yaml:"auth"`
		}

		w := wrapper{Auth: VerifyClientCertIfGiven}
		b, err := yaml.Marshal(w)
		Expect(err).ToNot(HaveOccurred())

		var w2 wrapper
		Expect(yaml.Unmarshal(b, &w2)).To(Succeed())
		Expect(w2.Auth).To(Equal(w.Auth))
	})

	It("Marshal/Unmarshal TOML roundtrip", func() {
		type wrapper struct {
			Auth ClientAuth `toml:"auth"`
		}

		w := wrapper{Auth: RequireAnyClientCert}
		b, err := toml.Marshal(w)
		Expect(err).ToNot(HaveOccurred())

		var w2 wrapper
		Expect(toml.Unmarshal(b, &w2)).To(Succeed())
		Expect(w2.Auth).To(Equal(w.Auth))
	})

	It("Marshal/Unmarshal CBOR roundtrip", func() {
		type wrapper struct {
			Auth ClientAuth `cbor:"1"`
		}

		w := wrapper{Auth: RequestClientCert}
		b, err := cbor.Marshal(w)
		Expect(err).ToNot(HaveOccurred())

		var w2 wrapper
		Expect(cbor.Unmarshal(b, &w2)).To(Succeed())
		Expect(w2.Auth).To(Equal(w.Auth))
	})

	It("Marshal/Unmarshal Text roundtrip", func() {
		a := NoClientCert
		b, err := a.MarshalText()
		Expect(err).ToNot(HaveOccurred())

		var a2 ClientAuth
		Expect(a2.UnmarshalText(b)).To(Succeed())
		Expect(a2).To(Equal(a))
	})

	It("UnmarshalTOML should handle both string and []byte", func() {
		var a ClientAuth
		Expect(a.UnmarshalTOML("verify")).To(Succeed())
		Expect(a).To(Equal(VerifyClientCertIfGiven))

		var a2 ClientAuth
		Expect(a2.UnmarshalTOML([]byte("require"))).To(Succeed())
		Expect(a2).To(Equal(RequireAnyClientCert))

		var a3 ClientAuth
		err := a3.UnmarshalTOML(123) // invalid type
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not in valid format"))
	})

	It("Should test all modes roundtrip correctly", func() {
		modes := []ClientAuth{
			NoClientCert,
			RequestClientCert,
			RequireAnyClientCert,
			VerifyClientCertIfGiven,
			RequireAndVerifyClientCert,
		}

		for _, mode := range modes {
			s := mode.String()
			parsed := Parse(s)
			Expect(parsed).To(Equal(mode), "Failed for mode: %v", mode)
		}
	})
})
