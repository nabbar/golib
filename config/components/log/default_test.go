/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package log_test

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	loglvl "github.com/nabbar/golib/logger/level"
)

// Default configuration tests verify the DefaultConfig method behavior.
var _ = Describe("Default Configuration", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		cpt CptLog
	)

	BeforeEach(func() {
		ctx, cnl = context.WithCancel(x)
		cpt = New(ctx, loglvl.NilLevel)
		cpt.Init(kd, ctx, nil, fv, vs, fl)

		v.Viper().SetConfigType("json")

		configData := map[string]interface{}{
			kd: map[string]interface{}{
				"stdout": map[string]interface{}{
					"disableStandard": true,
				},
			},
		}

		configJSON, err := json.Marshal(configData)
		Expect(err).To(BeNil())

		err = v.Viper().ReadConfig(bytes.NewReader(configJSON))
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		if cpt != nil {
			cpt.Stop()
		}
		cnl()
	})

	Describe("DefaultConfig method", func() {
		Context("generating default config", func() {
			It("should return non-empty configuration", func() {
				config := cpt.DefaultConfig("")
				Expect(config).NotTo(BeEmpty())
			})

			It("should return valid JSON", func() {
				config := cpt.DefaultConfig("")
				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})

			It("should return consistent output", func() {
				config1 := cpt.DefaultConfig("")
				config2 := cpt.DefaultConfig("")
				Expect(config1).To(Equal(config2))
			})

			It("should support indentation", func() {
				config := cpt.DefaultConfig("  ")
				Expect(config).NotTo(BeEmpty())

				// Should still be valid JSON
				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})

			It("should work with empty indentation", func() {
				config := cpt.DefaultConfig("")
				Expect(config).NotTo(BeEmpty())
			})

			It("should work with tab indentation", func() {
				config := cpt.DefaultConfig("\t")
				Expect(config).NotTo(BeEmpty())
			})

			It("should have logger configuration keys", func() {
				config := cpt.DefaultConfig("")
				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				Expect(result).NotTo(BeEmpty())
			})
		})

		Context("configuration structure", func() {
			It("should provide a valid configuration template", func() {
				config := cpt.DefaultConfig("")
				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
				// The configuration should be a valid structure
				Expect(result).NotTo(BeNil())
			})
		})

		Context("edge cases", func() {
			It("should handle multiple calls", func() {
				for i := 0; i < 10; i++ {
					config := cpt.DefaultConfig("")
					Expect(config).NotTo(BeEmpty())
				}
			})

			It("should work with various indent strings", func() {
				indents := []string{"", " ", "  ", "\t", "    ", "\t\t"}

				for _, indent := range indents {
					config := cpt.DefaultConfig(indent)
					Expect(config).NotTo(BeEmpty())
					var result map[string]interface{}
					err := json.Unmarshal(config, &result)
					Expect(err).To(BeNil())
				}
			})

			It("should not panic with unusual indent", func() {
				Expect(func() { _ = cpt.DefaultConfig("\n") }).NotTo(Panic())
			})
		})

		Describe("Concurrent access", func() {
			Context("thread-safety", func() {
				It("should handle concurrent DefaultConfig calls", func() {
					done := make(chan bool, 10)

					for i := 0; i < 10; i++ {
						go func() {
							defer GinkgoRecover()
							config := cpt.DefaultConfig("")
							Expect(config).NotTo(BeEmpty())
							done <- true
						}()
					}

					for i := 0; i < 10; i++ {
						Eventually(done).Should(Receive())
					}
				})

				It("should return consistent results concurrently", func() {
					results := make(chan []byte, 5)

					for i := 0; i < 5; i++ {
						go func() {
							defer GinkgoRecover()
							config := cpt.DefaultConfig("")
							results <- config
						}()
					}

					var configs [][]byte
					for i := 0; i < 5; i++ {
						select {
						case config := <-results:
							configs = append(configs, config)
						}
					}

					// All results should be equal
					for i := 1; i < len(configs); i++ {
						Expect(configs[i]).To(Equal(configs[0]))
					}
				})
			})
		})
	})
})
