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

package tlsversion_test

import (
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	. "github.com/nabbar/golib/certificates/tlsversion"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var _ = Describe("tlsversion", func() {
	It("Parse should recognize known versions and aliases", func() {
		Expect(Parse("TLS1.2")).To(Equal(VersionTLS12))
		Expect(Parse("tls_1_3")).To(Equal(VersionTLS13))
		Expect(Parse("ssl1.0")).To(Equal(VersionTLS10))
		Expect(Parse("unknown")).To(Equal(VersionUnknown))
	})

	It("String/Code and numeric conversions work", func() {
		Expect(VersionTLS12.String()).To(Equal("TLS 1.2"))
		Expect(VersionTLS13.Code()).To(Equal("tls_1.3"))
		Expect(int(VersionTLS11.Uint16())).To(Equal(VersionTLS11.Int()))
		Expect(VersionTLS10.Uint()).To(BeNumerically(">", 0))
	})

	It("Marshal/Unmarshal JSON/YAML/TOML/CBOR/Text roundtrip", func() {
		type TestMash struct {
			Vrs Version `json:"version" yaml:"version" toml:"version" cbor:"1"`
		}
		var (
			v = TestMash{
				Vrs: VersionTLS12,
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
		b, e = v.Vrs.MarshalText()
		Expect(e).ToNot(HaveOccurred())
		var v6 Version
		Expect(v6.UnmarshalText(b)).To(Succeed())
		Expect(v6).To(Equal(v.Vrs))
	})

	It("ParseInt should map back to same value for known version", func() {
		Expect(ParseInt(VersionTLS13.Int())).To(Equal(VersionTLS13))
	})
})
