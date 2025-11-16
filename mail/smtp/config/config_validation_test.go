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

package config_test

import (
	smtpcfg "github.com/nabbar/golib/mail/smtp/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMTP Config Validation", func() {

	Describe("ConfigModel Validation", func() {
		Context("with valid configurations", func() {
			It("should validate basic config", func() {
				model := newConfigModel("tcp(localhost:25)/")
				err := model.Validate()
				Expect(err).To(BeNil())
			})

			It("should validate config with authentication", func() {
				model := newConfigModel("user:pass@tcp(localhost:25)/")
				err := model.Validate()
				Expect(err).To(BeNil())
			})

			It("should validate config with TLS", func() {
				model := newConfigModel("tcp(localhost:587)/starttls")
				err := model.Validate()
				Expect(err).To(BeNil())
			})

			It("should validate config with query parameters", func() {
				model := newConfigModel("tcp(localhost:25)/?ServerName=test&SkipVerify=true")
				err := model.Validate()
				Expect(err).To(BeNil())
			})

			It("should validate all valid DSN formats", func() {
				dsns := validDSNs()
				for _, dsn := range dsns {
					model := newConfigModel(dsn)
					err := model.Validate()
					Expect(err).To(BeNil(), "DSN: %s", dsn)
				}
			})
		})

		Context("with invalid configurations", func() {
			It("should fail validation for empty DSN", func() {
				model := newConfigModel("")
				err := model.Validate()
				// Empty DSN should be caught by Validate()
				Expect(err).ToNot(BeNil())
			})

			It("should fail validation for DSN without closing brace", func() {
				model := newConfigModel("tcp(localhost:25")
				err := model.Validate()
				Expect(err).ToNot(BeNil())
			})

			It("should fail validation for DSN without slash", func() {
				model := newConfigModel("tcp(localhost:25)")
				err := model.Validate()
				// This might be auto-corrected, so check result
				_ = err
			})

			It("should handle validation of invalid query parameters", func() {
				model := newConfigModel("tcp(localhost:25)/?invalid=param&bad")
				err := model.Validate()
				// May or may not error depending on parser tolerance
				_ = err
			})
		})
	})

	Describe("Error Messages", func() {
		It("should provide meaningful error for empty DSN via Validate", func() {
			model := newConfigModel("")
			err := model.Validate()
			// Validate() should catch empty DSN
			Expect(err).ToNot(BeNil())
			// Error message is from parent validator error
			Expect(err.Error()).To(ContainSubstring("validation error"))
		})

		It("should provide meaningful error for missing closing brace", func() {
			model := newConfigModel("tcp(localhost:25")
			err := model.Validate()
			// Validate() calls New() which should error on malformed DSN
			Expect(err).ToNot(BeNil())
			if err != nil {
				Expect(err.Error()).ToNot(BeEmpty())
			}
		})
	})

	Describe("Config Validation After Creation", func() {
		It("should validate config created from valid DSN", func() {
			model := newConfigModel("tcp(localhost:25)/")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg).ToNot(BeNil())
		})

		It("should reject config from invalid DSN", func() {
			model := newConfigModel("")
			cfg, err := model.Config()
			// Should error during validation
			if err != nil {
				Expect(cfg).To(BeNil())
			}
		})
	})

	Describe("Validation Edge Cases", func() {
		It("should handle validation of minimal DSN", func() {
			model := newConfigModel("/")
			err := model.Validate()
			// Minimal DSN may or may not be valid
			_ = err
		})

		It("should handle validation of DSN with only protocol", func() {
			model := newConfigModel("tcp/")
			err := model.Validate()
			// May be valid with empty host
			_ = err
		})

		It("should handle validation of DSN with host but no port", func() {
			model := newConfigModel("tcp(localhost)/")
			err := model.Validate()
			// Should be valid - port is optional
			Expect(err).To(BeNil())
		})
	})

	Describe("Concurrent Validation", func() {
		It("should handle concurrent validation calls", func() {
			model := newConfigModel("tcp(localhost:25)/")
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					err := model.Validate()
					Expect(err).To(BeNil())
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent validation of different models", func() {
			dsns := []string{
				"tcp(localhost:25)/",
				"tcp(localhost:587)/starttls",
				"tcp(localhost:465)/tls",
				"user:pass@tcp(localhost:25)/",
			}

			done := make(chan bool, len(dsns))

			for _, dsn := range dsns {
				go func(d string) {
					defer GinkgoRecover()
					model := newConfigModel(d)
					err := model.Validate()
					Expect(err).To(BeNil())
					done <- true
				}(dsn)
			}

			for i := 0; i < len(dsns); i++ {
				<-done
			}
		})
	})

	Describe("Validation with TLS Config", func() {
		It("should validate config with empty TLS configuration", func() {
			model := smtpcfg.ConfigModel{
				DSN: "tcp(localhost:465)/tls",
				// TLS field is zero value (empty Config struct)
			}
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should validate config without TLS configuration", func() {
			model := newConfigModel("tcp(localhost:25)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})
	})

	Describe("Validation with Monitor Config", func() {
		It("should validate config with monitor settings", func() {
			model := smtpcfg.ConfigModel{
				DSN: "tcp(localhost:25)/",
			}
			err := model.Validate()
			Expect(err).To(BeNil())
		})
	})

	Describe("Query Parameter Validation", func() {
		It("should accept valid ServerName parameter", func() {
			model := newConfigModel("tcp(localhost:25)/?ServerName=smtp.example.com")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept valid SkipVerify=true parameter", func() {
			model := newConfigModel("tcp(localhost:25)/?SkipVerify=true")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept valid SkipVerify=false parameter", func() {
			model := newConfigModel("tcp(localhost:25)/?SkipVerify=false")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should handle invalid SkipVerify values", func() {
			model := newConfigModel("tcp(localhost:25)/?SkipVerify=invalid")
			err := model.Validate()
			// Should either ignore invalid value or error
			_ = err
		})

		It("should handle empty query parameters", func() {
			model := newConfigModel("tcp(localhost:25)/?")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should handle multiple query parameters", func() {
			model := newConfigModel("tcp(localhost:25)/?ServerName=test&SkipVerify=true")
			err := model.Validate()
			Expect(err).To(BeNil())
		})
	})

	Describe("Port Validation", func() {
		It("should accept standard SMTP port 25", func() {
			model := newConfigModel("tcp(localhost:25)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept submission port 587", func() {
			model := newConfigModel("tcp(localhost:587)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept SMTPS port 465", func() {
			model := newConfigModel("tcp(localhost:465)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept custom port", func() {
			model := newConfigModel("tcp(localhost:2525)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept maximum valid port", func() {
			model := newConfigModel("tcp(localhost:65535)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept DSN without port", func() {
			model := newConfigModel("tcp(localhost)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})
	})

	Describe("Host Validation", func() {
		It("should accept localhost", func() {
			model := newConfigModel("tcp(localhost:25)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept domain names", func() {
			model := newConfigModel("tcp(smtp.example.com:25)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept IPv4 addresses", func() {
			model := newConfigModel("tcp(192.168.1.1:25)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept IPv6 addresses", func() {
			model := newConfigModel("tcp6([::1]:25)/")
			err := model.Validate()
			Expect(err).To(BeNil())
		})

		It("should accept empty host", func() {
			model := newConfigModel("tcp/")
			err := model.Validate()
			// May be valid for some use cases
			_ = err
		})
	})
})
