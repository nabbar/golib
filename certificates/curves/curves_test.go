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

package curves_test

import (
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	. "github.com/nabbar/golib/certificates/curves"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var _ = Describe("curves", func() {
	It("Parse should recognize known curves and aliases", func() {
		Expect(Parse("X25519")).To(Equal(X25519))
		Expect(Parse("P256")).To(Equal(P256))
		Expect(Parse("p-384")).To(Equal(P384))
		Expect(Parse("curve p521")).To(Equal(P521))
		Expect(Parse("unknown")).To(Equal(Unknown))
	})

	It("String/Code and TLS conversions work", func() {
		Expect(X25519.String()).To(Equal("X25519"))
		Expect(P256.Code()).To(Equal("p256"))
		Expect(int(P384.TLS())).To(Equal(P384.Int()))
		Expect(P521.Uint16()).ToNot(BeZero())
	})

	It("Marshal/Unmarshal JSON/YAML/TOML/CBOR/Text roundtrip", func() {
		type TestMash struct {
			Crv Curves `json:"curves" yaml:"curves" toml:"curves" cbor:"1"`
		}
		var (
			v = TestMash{
				Crv: X25519,
			}
			b []byte
			e error
		)

		// JSON
		b, e = json.Marshal(v)
		Expect(e).ToNot(HaveOccurred())
		var v2 TestMash
		Expect(json.Unmarshal(b, &v2)).To(Succeed())
		Expect(v2).To(Equal(v))

		// YAML
		b, e = yaml.Marshal(v)
		Expect(e).ToNot(HaveOccurred())
		var v3 TestMash
		Expect(yaml.Unmarshal(b, &v3)).To(Succeed())
		Expect(v3).To(Equal(v))

		// TOML
		b, e = toml.Marshal(v)
		Expect(e).ToNot(HaveOccurred())
		var v4 TestMash
		Expect(toml.Unmarshal(b, &v4)).To(Succeed())
		Expect(v4).To(Equal(v))

		// CBOR
		b, e = cbor.Marshal(v)
		Expect(e).ToNot(HaveOccurred())
		var v5 TestMash
		Expect(cbor.Unmarshal(b, &v5)).To(Succeed())
		Expect(v5).To(Equal(v))

		// Text
		txt, err := v.Crv.MarshalText()
		Expect(err).ToNot(HaveOccurred())
		var v6 Curves
		Expect(v6.UnmarshalText(txt)).To(Succeed())
		Expect(v6).To(Equal(v.Crv))
	})

	It("Numeric helpers should map back to same value for known curve", func() {
		v := X25519
		Expect(ParseInt(v.Int())).To(Equal(v))
		Expect(Check(v.Uint16())).To(BeTrue())
	})
})
