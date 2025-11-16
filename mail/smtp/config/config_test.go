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
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
	libptc "github.com/nabbar/golib/network/protocol"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMTP Config Basic Operations", func() {

	Describe("Config Creation", func() {
		Context("with valid DSN", func() {
			It("should create a basic config", func() {
				cfg, err := createBasicConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg).ToNot(BeNil())
				Expect(cfg.GetHost()).To(Equal("localhost"))
				Expect(cfg.GetPort()).To(Equal(25))
			})

			It("should create config with authentication", func() {
				cfg, err := createConfigWithAuth("testuser", "testpass")
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg).ToNot(BeNil())
				Expect(cfg.GetUser()).To(Equal("testuser"))
				Expect(cfg.GetPass()).To(Equal("testpass"))
			})

			It("should create config from ConfigModel", func() {
				model := newConfigModel("tcp(localhost:587)/starttls")
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg).ToNot(BeNil())
				Expect(cfg.GetHost()).To(Equal("localhost"))
				Expect(cfg.GetPort()).To(Equal(587))
				Expect(cfg.GetTlsMode()).To(Equal(smtptp.TLSStartTLS))
			})
		})

		Context("with different network protocols", func() {
			It("should handle tcp protocol", func() {
				model := newConfigModel("tcp(localhost:25)/")
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.GetNet()).To(Equal(libptc.NetworkTCP))
			})

			It("should handle tcp4 protocol", func() {
				model := newConfigModel("tcp4(localhost:25)/")
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.GetNet()).To(Equal(libptc.NetworkTCP4))
			})

			It("should handle tcp6 protocol", func() {
				model := newConfigModel("tcp6([::1]:25)/")
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.GetNet()).To(Equal(libptc.NetworkTCP6))
			})
		})

		Context("with different TLS modes", func() {
			It("should handle no TLS", func() {
				model := newConfigModel("tcp(localhost:25)/")
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.GetTlsMode()).To(Equal(smtptp.TLSNone))
			})

			It("should handle STARTTLS", func() {
				model := newConfigModel("tcp(localhost:587)/starttls")
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.GetTlsMode()).To(Equal(smtptp.TLSStartTLS))
			})

			It("should handle strict TLS", func() {
				model := newConfigModel("tcp(localhost:465)/tls")
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.GetTlsMode()).To(Equal(smtptp.TLSStrictTLS))
			})
		})
	})

	Describe("Config Getters", func() {
		var cfg smtpcfg.Config

		BeforeEach(func() {
			model := newConfigModel("user:pass@tcp(smtp.example.com:587)/starttls?ServerName=mail.example.com&SkipVerify=true")
			var err error
			cfg, err = model.Config()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return correct host", func() {
			Expect(cfg.GetHost()).To(Equal("smtp.example.com"))
		})

		It("should return correct port", func() {
			Expect(cfg.GetPort()).To(Equal(587))
		})

		It("should return correct user", func() {
			Expect(cfg.GetUser()).To(Equal("user"))
		})

		It("should return correct password", func() {
			Expect(cfg.GetPass()).To(Equal("pass"))
		})

		It("should return correct network protocol", func() {
			Expect(cfg.GetNet()).To(Equal(libptc.NetworkTCP))
		})

		It("should return correct TLS mode", func() {
			Expect(cfg.GetTlsMode()).To(Equal(smtptp.TLSStartTLS))
		})

		It("should return correct skip verify flag", func() {
			Expect(cfg.IsTLSSkipVerify()).To(BeTrue())
		})

		It("should return correct server name", func() {
			Expect(cfg.GetTlSServerName()).To(Equal("mail.example.com"))
		})

		It("should return valid DSN", func() {
			dsn := cfg.GetDsn()
			Expect(dsn).ToNot(BeEmpty())
			Expect(dsn).To(ContainSubstring("smtp.example.com"))
			Expect(dsn).To(ContainSubstring("587"))
		})
	})

	Describe("Config Setters", func() {
		var cfg smtpcfg.Config

		BeforeEach(func() {
			var err error
			cfg, err = createBasicConfig()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should set host", func() {
			cfg.SetHost("newhost.example.com")
			Expect(cfg.GetHost()).To(Equal("newhost.example.com"))
		})

		It("should set port", func() {
			cfg.SetPort(465)
			Expect(cfg.GetPort()).To(Equal(465))
		})

		It("should set user", func() {
			cfg.SetUser("newuser")
			Expect(cfg.GetUser()).To(Equal("newuser"))
		})

		It("should set password", func() {
			cfg.SetPass("newpass")
			Expect(cfg.GetPass()).To(Equal("newpass"))
		})

		It("should set network protocol", func() {
			cfg.SetNet(libptc.NetworkTCP6)
			Expect(cfg.GetNet()).To(Equal(libptc.NetworkTCP6))
		})

		It("should set TLS mode", func() {
			cfg.SetTlsMode(smtptp.TLSStrictTLS)
			Expect(cfg.GetTlsMode()).To(Equal(smtptp.TLSStrictTLS))
		})

		It("should set TLS skip verify", func() {
			cfg.ForceTLSSkipVerify(true)
			Expect(cfg.IsTLSSkipVerify()).To(BeTrue())

			cfg.ForceTLSSkipVerify(false)
			Expect(cfg.IsTLSSkipVerify()).To(BeFalse())
		})

		It("should set TLS server name", func() {
			cfg.SetTLSServerName("tls.example.com")
			Expect(cfg.GetTlSServerName()).To(Equal("tls.example.com"))
		})
	})

	Describe("DSN Regeneration", func() {
		It("should regenerate DSN after modifications", func() {
			cfg, err := createBasicConfig()
			Expect(err).ToNot(HaveOccurred())

			// Modify config
			cfg.SetHost("modified.example.com")
			cfg.SetPort(587)
			cfg.SetUser("testuser")
			cfg.SetPass("testpass")
			cfg.SetTlsMode(smtptp.TLSStartTLS)

			// Get new DSN
			dsn := cfg.GetDsn()
			Expect(dsn).To(ContainSubstring("modified.example.com"))
			Expect(dsn).To(ContainSubstring("587"))
			Expect(dsn).To(ContainSubstring("testuser"))
			Expect(dsn).To(ContainSubstring("testpass"))
			Expect(dsn).To(ContainSubstring("starttls"))
		})

		It("should handle DSN roundtrip", func() {
			// Create config from DSN
			originalDSN := "user:pass@tcp(smtp.example.com:587)/starttls"
			model := newConfigModel(originalDSN)
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())

			// Get DSN back
			newDSN := cfg.GetDsn()

			// Parse new DSN and compare
			model2 := newConfigModel(newDSN)
			cfg2, err := model2.Config()
			Expect(err).ToNot(HaveOccurred())

			// Compare configs
			assertConfigEquals(cfg, cfg2)
		})
	})

	Describe("Config with Query Parameters", func() {
		It("should parse ServerName parameter", func() {
			model := newConfigModel("tcp(localhost:25)/?ServerName=smtp.example.com")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetTlSServerName()).To(Equal("smtp.example.com"))
		})

		It("should parse SkipVerify=true parameter", func() {
			model := newConfigModel("tcp(localhost:25)/?SkipVerify=true")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.IsTLSSkipVerify()).To(BeTrue())
		})

		It("should parse SkipVerify=false parameter", func() {
			model := newConfigModel("tcp(localhost:25)/?SkipVerify=false")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.IsTLSSkipVerify()).To(BeFalse())
		})

		It("should parse multiple parameters", func() {
			model := newConfigModel("tcp(localhost:25)/?ServerName=smtp.example.com&SkipVerify=true")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetTlSServerName()).To(Equal("smtp.example.com"))
			Expect(cfg.IsTLSSkipVerify()).To(BeTrue())
		})

		It("should handle empty parameter values", func() {
			model := newConfigModel("tcp(localhost:25)/?ServerName=")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetTlSServerName()).To(BeEmpty())
		})
	})

	Describe("Special Characters in DSN", func() {
		It("should handle special characters in password", func() {
			// Using URL-encoded special characters
			model := newConfigModel("user:p@ss!word@tcp(localhost:25)/")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetUser()).To(Equal("user"))
			// Note: The parser uses the last @ to split user/pass from host
		})

		It("should handle IPv6 addresses", func() {
			model := newConfigModel("tcp6([::1]:25)/")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetHost()).To(Equal("::1"))
			Expect(cfg.GetPort()).To(Equal(25))
		})

		It("should handle IPv6 addresses with brackets", func() {
			model := newConfigModel("tcp6([2001:db8::1]:587)/")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetHost()).To(Equal("2001:db8::1"))
			Expect(cfg.GetPort()).To(Equal(587))
		})
	})

	Describe("Empty or Minimal DSN", func() {
		It("should handle DSN with only slash", func() {
			model := newConfigModel("/")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.GetHost()).To(BeEmpty())
			Expect(cfg.GetPort()).To(Equal(0))
		})

		It("should handle DSN with protocol only", func() {
			model := newConfigModel("tcp/")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetNet()).To(Equal(libptc.NetworkTCP))
			Expect(cfg.GetHost()).To(BeEmpty())
		})

		It("should handle DSN with host but no port", func() {
			model := newConfigModel("tcp(localhost)/")
			cfg, err := model.Config()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.GetHost()).To(Equal("localhost"))
			Expect(cfg.GetPort()).To(Equal(0))
		})
	})
})
