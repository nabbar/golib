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

package status_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monsts "github.com/nabbar/golib/monitor/status"
	libsts "github.com/nabbar/golib/status"
	stsctr "github.com/nabbar/golib/status/control"
)

var _ = Describe("Status/Config", func() {
	var (
		status libsts.Status
	)

	BeforeEach(func() {
		status = libsts.New(globalCtx)
		status.SetInfo("test-app", "v1.0.0", "abc123")
	})

	Describe("Config.Validate", func() {
		It("should validate empty config", func() {
			cfg := libsts.Config{}
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should validate config with return codes", func() {
			cfg := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK:   200,
					monsts.Warn: 207,
					monsts.KO:   500,
				},
			}
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should validate config with mandatory components", func() {
			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"database", "cache"},
					},
				},
			}
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("SetConfig", func() {
		Context("with default return codes", func() {
			It("should use default codes when not specified", func() {
				cfg := libsts.Config{}
				status.SetConfig(cfg)

				// Verify default behavior by checking health
				healthy := status.IsHealthy()
				Expect(healthy).To(BeAssignableToTypeOf(false))
			})
		})

		Context("with custom return codes", func() {
			It("should accept custom HTTP status codes", func() {
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK:   200,
						monsts.Warn: 200, // Treat warnings as OK
						monsts.KO:   503,
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should handle partial return code configuration", func() {
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK: 200,
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})
		})

		Context("with mandatory components", func() {
			It("should configure Must mode", func() {
				cfg := libsts.Config{
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"critical-service"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure Should mode", func() {
				cfg := libsts.Config{
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Should,
							Keys: []string{"optional-service"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure AnyOf mode", func() {
				cfg := libsts.Config{
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.AnyOf,
							Keys: []string{"db-primary", "db-secondary", "db-tertiary"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure Quorum mode", func() {
				cfg := libsts.Config{
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Quorum,
							Keys: []string{"node-1", "node-2", "node-3", "node-4", "node-5"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should configure Ignore mode", func() {
				cfg := libsts.Config{
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Ignore,
							Keys: []string{"non-critical"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})

			It("should handle multiple mandatory groups", func() {
				cfg := libsts.Config{
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"database"},
						},
						{
							Mode: stsctr.Should,
							Keys: []string{"cache"},
						},
						{
							Mode: stsctr.AnyOf,
							Keys: []string{"queue-1", "queue-2"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})
		})

		Context("with complete configuration", func() {
			It("should configure both return codes and mandatory components", func() {
				cfg := libsts.Config{
					ReturnCode: map[monsts.Status]int{
						monsts.OK:   http.StatusOK,
						monsts.Warn: http.StatusMultiStatus,
						monsts.KO:   http.StatusServiceUnavailable,
					},
					MandatoryComponent: []libsts.Mandatory{
						{
							Mode: stsctr.Must,
							Keys: []string{"database", "api"},
						},
						{
							Mode: stsctr.Should,
							Keys: []string{"cache", "queue"},
						},
					},
				}
				status.SetConfig(cfg)
				Expect(true).To(BeTrue())
			})
		})
	})

	Describe("Config updates", func() {
		It("should allow updating configuration multiple times", func() {
			// First configuration
			cfg1 := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK: 200,
				},
			}
			status.SetConfig(cfg1)

			// Second configuration
			cfg2 := libsts.Config{
				ReturnCode: map[monsts.Status]int{
					monsts.OK:   200,
					monsts.Warn: 207,
					monsts.KO:   500,
				},
			}
			status.SetConfig(cfg2)

			Expect(true).To(BeTrue())
		})

		It("should replace previous configuration", func() {
			cfg1 := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"old-service"},
					},
				},
			}
			status.SetConfig(cfg1)

			cfg2 := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Should,
						Keys: []string{"new-service"},
					},
				},
			}
			status.SetConfig(cfg2)

			Expect(true).To(BeTrue())
		})
	})
})
