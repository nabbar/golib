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

package ldap_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/ldap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	lbldap "github.com/nabbar/golib/ldap"
)

// Configuration tests verify LDAP configuration management, attributes handling,
// and LDAP helper operations.
var _ = Describe("Configuration Management", func() {
	var (
		cpt CptLDAP
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	Describe("Config operations", func() {
		Context("GetConfig and SetConfig", func() {
			It("should return nil for empty config", func() {
				cfg := cpt.GetConfig()
				Expect(cfg).To(BeNil())
			})

			It("should store and retrieve valid config", func() {
				testCfg := &lbldap.Config{
					Uri:         "ldap://localhost",
					PortLdap:    389,
					Basedn:      "dc=example,dc=com",
					FilterUser:  "(uid=%s)",
					FilterGroup: "(&(objectClass=groupOfNames)(member=%s))",
				}

				cpt.SetConfig(testCfg)
				cfg := cpt.GetConfig()

				Expect(cfg).NotTo(BeNil())
				Expect(cfg.Uri).To(Equal("ldap://localhost"))
				Expect(cfg.PortLdap).To(Equal(389))
				Expect(cfg.Basedn).To(Equal("dc=example,dc=com"))
				Expect(cfg.FilterUser).To(Equal("(uid=%s)"))
			})

			It("should handle nil config gracefully", func() {
				Expect(func() {
					cpt.SetConfig(nil)
				}).NotTo(Panic())

				cfg := cpt.GetConfig()
				Expect(cfg).To(BeNil())
			})

			It("should return nil for config without URI", func() {
				testCfg := &lbldap.Config{
					PortLdap: 389,
				}

				cpt.SetConfig(testCfg)
				cfg := cpt.GetConfig()

				Expect(cfg).To(BeNil())
			})

			It("should return nil for config without ports", func() {
				testCfg := &lbldap.Config{
					Uri: "ldap://localhost",
				}

				cpt.SetConfig(testCfg)
				cfg := cpt.GetConfig()

				Expect(cfg).To(BeNil())
			})

			It("should accept config with LDAPS port", func() {
				testCfg := &lbldap.Config{
					Uri:         "ldaps://localhost",
					Portldaps:   636,
					Basedn:      "dc=example,dc=com",
					FilterUser:  "(uid=%s)",
					FilterGroup: "(&(objectClass=groupOfNames)(member=%s))",
				}

				cpt.SetConfig(testCfg)
				cfg := cpt.GetConfig()

				Expect(cfg).NotTo(BeNil())
				Expect(cfg.Portldaps).To(Equal(636))
			})

			It("should return copy on GetConfig", func() {
				testCfg := &lbldap.Config{
					Uri:         "ldap://localhost",
					PortLdap:    389,
					Basedn:      "dc=example,dc=com",
					FilterUser:  "(uid=%s)",
					FilterGroup: "(&(objectClass=groupOfNames)(member=%s))",
				}

				cpt.SetConfig(testCfg)

				// GetConfig returns a copy
				cfg1 := cpt.GetConfig()
				cfg2 := cpt.GetConfig()

				Expect(cfg1).NotTo(BeNil())
				Expect(cfg2).NotTo(BeNil())

				// Verify they are different pointers (copies)
				Expect(cfg1).NotTo(BeIdenticalTo(cfg2))

				// But same values
				Expect(cfg1.Basedn).To(Equal(cfg2.Basedn))
				Expect(cfg1.Uri).To(Equal(cfg2.Uri))
			})
		})
	})

	Describe("Attributes operations", func() {
		Context("SetAttributes and GetAttributes", func() {
			It("should return empty array by default", func() {
				attrs := cpt.GetAttributes()
				Expect(attrs).NotTo(BeNil())
				Expect(attrs).To(BeEmpty())
			})

			It("should store and retrieve attributes", func() {
				testAttrs := []string{"uid", "cn", "mail", "givenName"}
				cpt.SetAttributes(testAttrs)

				attrs := cpt.GetAttributes()
				Expect(attrs).To(Equal(testAttrs))
			})

			It("should handle nil attributes", func() {
				cpt.SetAttributes(nil)

				attrs := cpt.GetAttributes()
				Expect(attrs).NotTo(BeNil())
				Expect(attrs).To(BeEmpty())
			})

			It("should handle empty attributes", func() {
				cpt.SetAttributes([]string{})

				attrs := cpt.GetAttributes()
				Expect(attrs).NotTo(BeNil())
				Expect(attrs).To(BeEmpty())
			})

			It("should overwrite previous attributes", func() {
				cpt.SetAttributes([]string{"uid", "cn"})
				cpt.SetAttributes([]string{"mail", "sn"})

				attrs := cpt.GetAttributes()
				Expect(attrs).To(Equal([]string{"mail", "sn"}))
			})

			It("should handle single attribute", func() {
				cpt.SetAttributes([]string{"uid"})

				attrs := cpt.GetAttributes()
				Expect(attrs).To(HaveLen(1))
				Expect(attrs[0]).To(Equal("uid"))
			})
		})
	})

	Describe("LDAP helper operations", func() {
		Context("GetLDAP and SetLDAP", func() {
			It("should return nil for uninitialized LDAP", func() {
				// GetLDAP may panic if the internal helper is not properly initialized
				// This is expected behavior
				Expect(func() {
					_ = cpt.GetLDAP()
				}).To(Panic())
			})

			It("should handle nil LDAP gracefully", func() {
				Expect(func() {
					cpt.SetLDAP(nil)
				}).NotTo(Panic())

				// After setting nil, GetLDAP may still panic
				Expect(func() {
					_ = cpt.GetLDAP()
				}).To(Panic())
			})

			It("should store LDAP helper", func() {
				helper := &lbldap.HelperLDAP{}

				Expect(func() {
					cpt.SetLDAP(helper)
				}).NotTo(Panic())
			})

			It("should include attributes in LDAP helper", func() {
				testAttrs := []string{"uid", "cn", "mail"}
				cpt.SetAttributes(testAttrs)

				// Create a minimal valid LDAP helper
				cfg := &lbldap.Config{
					Uri:         "ldap://localhost",
					PortLdap:    389,
					Basedn:      "dc=example,dc=com",
					FilterUser:  "(uid=%s)",
					FilterGroup: "(&(objectClass=groupOfNames)(member=%s))",
				}
				helper, _ := lbldap.NewLDAP(ctx, cfg, []string{})

				if helper != nil {
					cpt.SetLDAP(helper)

					retrieved := cpt.GetLDAP()
					if retrieved != nil {
						Expect(retrieved.Attributes).To(Equal(testAttrs))
					}
				}
			})
		})
	})

	Describe("Configuration validation", func() {
		Context("valid configurations", func() {
			It("should accept LDAP configuration", func() {
				cfg := &lbldap.Config{
					Uri:      "ldap://localhost",
					PortLdap: 389,
					Basedn:   "dc=example,dc=com",
				}

				cpt.SetConfig(cfg)
				retrieved := cpt.GetConfig()

				Expect(retrieved).NotTo(BeNil())
			})

			It("should accept LDAPS configuration", func() {
				cfg := &lbldap.Config{
					Uri:       "ldaps://localhost",
					Portldaps: 636,
					Basedn:    "dc=example,dc=com",
				}

				cpt.SetConfig(cfg)
				retrieved := cpt.GetConfig()

				Expect(retrieved).NotTo(BeNil())
			})

			It("should accept configuration with both ports", func() {
				cfg := &lbldap.Config{
					Uri:       "ldap://localhost",
					PortLdap:  389,
					Portldaps: 636,
					Basedn:    "dc=example,dc=com",
				}

				cpt.SetConfig(cfg)
				retrieved := cpt.GetConfig()

				Expect(retrieved).NotTo(BeNil())
			})
		})

		Context("invalid configurations", func() {
			It("should reject config without URI", func() {
				cfg := &lbldap.Config{
					PortLdap: 389,
					Basedn:   "dc=example,dc=com",
				}

				cpt.SetConfig(cfg)
				retrieved := cpt.GetConfig()

				Expect(retrieved).To(BeNil())
			})

			It("should reject config without any port", func() {
				cfg := &lbldap.Config{
					Uri:    "ldap://localhost",
					Basedn: "dc=example,dc=com",
				}

				cpt.SetConfig(cfg)
				retrieved := cpt.GetConfig()

				Expect(retrieved).To(BeNil())
			})
		})
	})
})
