/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package big_test

import (
	"encoding/json"

	durbig "github.com/nabbar/golib/duration/big"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type StructExample struct {
	Value durbig.Duration `json:"value" yaml:"value" toml:"value"`
}

var valueExample = StructExample{Value: (5 * durbig.Day) + (23 * durbig.Hour) + (15 * durbig.Minute) + (13 * durbig.Second)}

func jsonDuration() []byte {
	return []byte(`{"value":"5d23h15m13s"}`)
}

func yamlDuration() []byte {
	return []byte(`value: 5d23h15m13s
`)
}

func tomlDuration() []byte {
	return []byte(`value = "5d23h15m13s"
`)
}

var _ = Describe("duration/big", func() {
	Context("decoding value from json, yaml, toml", func() {
		var (
			err error
			obj = StructExample{}
		)

		It("success when json decoding", func() {
			err = json.Unmarshal(jsonDuration(), &obj)
			Expect(err).ToNot(HaveOccurred())
			Expect(obj.Value).To(Equal(valueExample.Value))
		})

		It("success when yaml decoding", func() {
			err = yaml.Unmarshal(yamlDuration(), &obj)
			Expect(err).ToNot(HaveOccurred())
			Expect(obj.Value).To(Equal(valueExample.Value))
		})

		It("success when toml decoding", func() {
			err = toml.Unmarshal(tomlDuration(), &obj)
			Expect(err).ToNot(HaveOccurred())
			Expect(obj.Value).To(Equal(valueExample.Value))
		})
	})
	Context("encoding value from json, yaml, toml", func() {
		var (
			err error
			res []byte
			str string
			exp string
		)

		It("success when json encoding", func() {
			res, err = json.Marshal(&valueExample)
			str = string(res)
			exp = string(jsonDuration())

			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(Equal(exp))
		})

		It("success when yaml encoding", func() {
			res, err = yaml.Marshal(&valueExample)
			str = string(res)
			exp = string(yamlDuration())

			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(Equal(exp))
		})

		It("success when toml encoding", func() {
			res, err = toml.Marshal(&valueExample)
			str = string(res)
			exp = string(tomlDuration())

			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(Equal(exp))
		})

		It("success when json encoding of maxDuration Value", func() {
			var dur durbig.Duration = 1<<63 - 1
			res, err = json.Marshal(&dur)
			str = string(res)
			exp = `"106751991167300d15h30m7s"`

			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(Equal(exp))
		})
	})
})
