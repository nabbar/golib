/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package perm_test

import (
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
	. "github.com/nabbar/golib/file/perm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Viper Decoder Hook", func() {
	var hook libmap.DecodeHookFuncType

	BeforeEach(func() {
		hook = ViperDecoderHook()
	})

	Describe("String to Perm Conversion", func() {
		It("should convert string to Perm", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"0644",
			)
			Expect(err).ToNot(HaveOccurred())
			perm, ok := result.(Perm)
			Expect(ok).To(BeTrue())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should convert string 0755 to Perm", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"0755",
			)
			Expect(err).ToNot(HaveOccurred())
			perm, ok := result.(Perm)
			Expect(ok).To(BeTrue())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should convert string without leading zero", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"644",
			)
			Expect(err).ToNot(HaveOccurred())
			perm, ok := result.(Perm)
			Expect(ok).To(BeTrue())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should handle string with quotes", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"\"0777\"",
			)
			Expect(err).ToNot(HaveOccurred())
			perm, ok := result.(Perm)
			Expect(ok).To(BeTrue())
			Expect(perm.Uint64()).To(Equal(uint64(0777)))
		})

		It("should return error for invalid string", func() {
			_, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"invalid",
			)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Non-String Source Types", func() {
		It("should pass through int unchanged", func() {
			result, err := hook(
				reflect.TypeOf(0),
				reflect.TypeOf(Perm(0)),
				123,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(123))
		})

		It("should pass through bool unchanged", func() {
			result, err := hook(
				reflect.TypeOf(true),
				reflect.TypeOf(Perm(0)),
				true,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
		})

		It("should pass through slice unchanged", func() {
			data := []string{"test"}
			result, err := hook(
				reflect.TypeOf(data),
				reflect.TypeOf(Perm(0)),
				data,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})
	})

	Describe("Non-Perm Target Types", func() {
		It("should pass through when target is not Perm", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(""),
				"0644",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("0644"))
		})

		It("should pass through when target is int", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(0),
				"123",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("123"))
		})
	})

	Describe("Mapstructure Integration", func() {
		It("should decode map to struct with Perm field", func() {
			type Config struct {
				Mode Perm
			}

			input := map[string]interface{}{
				"mode": "0644",
			}

			var output Config
			decoder, err := libmap.NewDecoder(&libmap.DecoderConfig{
				DecodeHook: hook,
				Result:     &output,
			})
			Expect(err).ToNot(HaveOccurred())

			err = decoder.Decode(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Mode.Uint64()).To(Equal(uint64(0644)))
		})

		It("should decode map with multiple Perm fields", func() {
			type Config struct {
				FileMode Perm
				DirMode  Perm
			}

			input := map[string]interface{}{
				"FileMode": "0644",
				"DirMode":  "0755",
			}

			var output Config
			decoder, err := libmap.NewDecoder(&libmap.DecoderConfig{
				DecodeHook: hook,
				Result:     &output,
			})
			Expect(err).ToNot(HaveOccurred())

			err = decoder.Decode(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(output.FileMode.Uint64()).To(Equal(uint64(0644)))
			Expect(output.DirMode.Uint64()).To(Equal(uint64(0755)))
		})

		It("should handle nested structs with Perm fields", func() {
			type PermConfig struct {
				Mode Perm
			}
			type Config struct {
				File PermConfig
				Dir  PermConfig
			}

			input := map[string]interface{}{
				"File": map[string]interface{}{
					"Mode": "0644",
				},
				"Dir": map[string]interface{}{
					"Mode": "0755",
				},
			}

			var output Config
			decoder, err := libmap.NewDecoder(&libmap.DecoderConfig{
				DecodeHook: hook,
				Result:     &output,
			})
			Expect(err).ToNot(HaveOccurred())

			err = decoder.Decode(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(output.File.Mode.Uint64()).To(Equal(uint64(0644)))
			Expect(output.Dir.Mode.Uint64()).To(Equal(uint64(0755)))
		})

		It("should return error for invalid permission value", func() {
			type Config struct {
				Mode Perm
			}

			input := map[string]interface{}{
				"mode": "invalid",
			}

			var output Config
			decoder, err := libmap.NewDecoder(&libmap.DecoderConfig{
				DecodeHook: hook,
				Result:     &output,
			})
			Expect(err).ToNot(HaveOccurred())

			err = decoder.Decode(input)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty string", func() {
			_, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"",
			)
			Expect(err).To(HaveOccurred())
		})

		It("should handle permission with special bits", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"04755",
			)
			Expect(err).ToNot(HaveOccurred())
			perm, ok := result.(Perm)
			Expect(ok).To(BeTrue())
			Expect(perm.Uint64()).To(Equal(uint64(04755)))
		})

		It("should handle maximum valid permission", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Perm(0)),
				"07777",
			)
			Expect(err).ToNot(HaveOccurred())
			perm, ok := result.(Perm)
			Expect(ok).To(BeTrue())
			Expect(perm.Uint64()).To(Equal(uint64(07777)))
		})
	})
})
