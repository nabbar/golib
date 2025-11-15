/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package ftpclient_test

import (
	"context"
	"time"

	. "github.com/nabbar/golib/ftpclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FTP Config", func() {
	Describe("Config Structure", func() {
		It("should create config with hostname", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
			}

			Expect(cfg.Hostname).To(Equal("ftp.example.com:21"))
		})

		It("should create config with credentials", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
				Login:    "testuser",
				Password: "testpass",
			}

			Expect(cfg.Login).To(Equal("testuser"))
			Expect(cfg.Password).To(Equal("testpass"))
		})

		It("should detect missing hostname in validation", func() {
			cfg := &Config{
				Login:    "testuser",
				Password: "testpass",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should detect invalid hostname in validation", func() {
			cfg := &Config{
				Hostname: "not a valid hostname!@#",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Config Options", func() {
		It("should set DisableUTF8 option", func() {
			cfg := &Config{
				Hostname:    "ftp.example.com:21",
				DisableUTF8: true,
			}

			Expect(cfg.DisableUTF8).To(BeTrue())
		})

		It("should set DisableEPSV option", func() {
			cfg := &Config{
				Hostname:    "ftp.example.com:21",
				DisableEPSV: true,
			}

			Expect(cfg.DisableEPSV).To(BeTrue())
		})

		It("should set DisableMLSD option", func() {
			cfg := &Config{
				Hostname:    "ftp.example.com:21",
				DisableMLSD: true,
			}

			Expect(cfg.DisableMLSD).To(BeTrue())
		})

		It("should set EnableMDTM option", func() {
			cfg := &Config{
				Hostname:   "ftp.example.com:21",
				EnableMDTM: true,
			}

			Expect(cfg.EnableMDTM).To(BeTrue())
		})

		It("should set ForceTLS option", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
				ForceTLS: true,
			}

			Expect(cfg.ForceTLS).To(BeTrue())
		})

		It("should set ConnTimeout option", func() {
			cfg := &Config{
				Hostname:    "ftp.example.com:21",
				ConnTimeout: 30 * time.Second,
			}

			Expect(cfg.ConnTimeout).To(Equal(30 * time.Second))
		})
	})

	Describe("Config TimeZone", func() {
		It("should set timezone configuration", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
				TimeZone: ConfigTimeZone{
					Name:   "America/New_York",
					Offset: -5,
				},
			}

			Expect(cfg.TimeZone.Name).To(Equal("America/New_York"))
			Expect(cfg.TimeZone.Offset).To(Equal(-5))
		})

		It("should set UTC timezone", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
				TimeZone: ConfigTimeZone{
					Name:   "UTC",
					Offset: 0,
				},
			}

			Expect(cfg.TimeZone.Name).To(Equal("UTC"))
			Expect(cfg.TimeZone.Offset).To(Equal(0))
		})

		It("should set positive timezone offset", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
				TimeZone: ConfigTimeZone{
					Name:   "Asia/Tokyo",
					Offset: 9,
				},
			}

			Expect(cfg.TimeZone.Offset).To(Equal(9))
		})
	})

	Describe("Config Context Registration", func() {
		It("should register context function", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
			}

			called := false
			cfg.RegisterContext(func() context.Context {
				called = true
				return context.Background()
			})

			// Context registration should succeed
			Expect(called).To(BeFalse()) // Not called yet
		})

		It("should handle nil context function", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
			}

			// Should not panic with nil
			Expect(func() {
				cfg.RegisterContext(nil)
			}).ToNot(Panic())
		})
	})

	Describe("Config Complete Examples", func() {
		It("should create minimal config", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
			}

			Expect(cfg.Hostname).To(Equal("ftp.example.com:21"))
		})

		It("should create full config", func() {
			cfg := &Config{
				Hostname:    "ftp.example.com:21",
				Login:       "testuser",
				Password:    "testpass",
				ConnTimeout: 10 * time.Second,
				TimeZone: ConfigTimeZone{
					Name:   "UTC",
					Offset: 0,
				},
				DisableUTF8: false,
				DisableEPSV: true,
				DisableMLSD: true,
				EnableMDTM:  true,
				ForceTLS:    false,
			}

			Expect(cfg.Hostname).To(Equal("ftp.example.com:21"))
			Expect(cfg.Login).To(Equal("testuser"))
		})

		It("should create secure config with TLS", func() {
			cfg := &Config{
				Hostname:    "ftps.example.com:990",
				Login:       "secureuser",
				Password:    "securepass",
				ConnTimeout: 15 * time.Second,
				ForceTLS:    true,
			}

			Expect(cfg.ForceTLS).To(BeTrue())
		})
	})

	Describe("Config Edge Cases", func() {
		It("should handle empty password", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
				Login:    "testuser",
				Password: "",
			}

			Expect(cfg.Password).To(BeEmpty())
		})

		It("should handle empty login", func() {
			cfg := &Config{
				Hostname: "ftp.example.com:21",
				Login:    "",
				Password: "testpass",
			}

			Expect(cfg.Login).To(BeEmpty())
		})

		It("should handle zero timeout", func() {
			cfg := &Config{
				Hostname:    "ftp.example.com:21",
				ConnTimeout: 0,
			}

			Expect(cfg.ConnTimeout).To(Equal(time.Duration(0)))
		})

		It("should handle very long timeout", func() {
			cfg := &Config{
				Hostname:    "ftp.example.com:21",
				ConnTimeout: 24 * time.Hour,
			}

			Expect(cfg.ConnTimeout).To(Equal(24 * time.Hour))
		})
	})
})
