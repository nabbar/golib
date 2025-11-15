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

	cfgtps "github.com/nabbar/golib/config/types"
	lbldap "github.com/nabbar/golib/ldap"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Helper tests verify edge cases, error handling, and special scenarios
// for the LDAP component.
var _ = Describe("Helper Functions and Edge Cases", func() {
	var (
		cpt CptLDAP
		ctx context.Context
		vrs libver.Version
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
		cpt = New(ctx)
	})

	Describe("Edge cases", func() {
		Context("nil component operations", func() {
			It("should handle nil context gracefully", func() {
				Expect(func() {
					_ = New(nil)
				}).NotTo(Panic())
			})

			It("should handle operations before initialization", func() {
				newCpt := New(ctx)

				Expect(func() {
					_ = newCpt.Start()
					_ = newCpt.Reload()
					newCpt.Stop()
				}).NotTo(Panic())
			})
		})

		Context("multiple lifecycle operations", func() {
			BeforeEach(func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }
				cpt.Init("test", ctx, getCpt, vpr, vrs, log)
			})

			It("should handle multiple Start calls", func() {
				Expect(func() {
					_ = cpt.Start()
					_ = cpt.Start()
					_ = cpt.Start()
				}).NotTo(Panic())
			})

			It("should handle multiple Reload calls", func() {
				Expect(func() {
					_ = cpt.Reload()
					_ = cpt.Reload()
					_ = cpt.Reload()
				}).NotTo(Panic())
			})

			It("should handle multiple Stop calls", func() {
				Expect(func() {
					cpt.Stop()
					cpt.Stop()
					cpt.Stop()
				}).NotTo(Panic())
			})

			It("should handle Start-Stop-Start sequence", func() {
				Expect(func() {
					_ = cpt.Start()
					cpt.Stop()
					_ = cpt.Start()
				}).NotTo(Panic())
			})
		})

		Context("configuration edge cases", func() {
			It("should handle config updates", func() {
				cfg1 := &lbldap.Config{
					Uri:      "ldap://server1",
					PortLdap: 389,
					Basedn:   "dc=example,dc=com",
				}
				cpt.SetConfig(cfg1)

				cfg2 := &lbldap.Config{
					Uri:      "ldap://server2",
					PortLdap: 389,
					Basedn:   "dc=test,dc=org",
				}
				cpt.SetConfig(cfg2)

				result := cpt.GetConfig()
				Expect(result).NotTo(BeNil())
				Expect(result.Uri).To(Equal("ldap://server2"))
			})

			It("should handle attributes updates", func() {
				cpt.SetAttributes([]string{"uid", "cn"})
				cpt.SetAttributes([]string{"mail", "sn"})

				attrs := cpt.GetAttributes()
				Expect(attrs).To(Equal([]string{"mail", "sn"}))
			})
		})

		Context("LDAP helper edge cases", func() {
			It("should handle LDAP helper replacement", func() {
				helper1 := &lbldap.HelperLDAP{}
				cpt.SetLDAP(helper1)

				helper2 := &lbldap.HelperLDAP{}
				cpt.SetLDAP(helper2)

				// GetLDAP may panic if helpers are not properly initialized
				// which is expected for empty HelperLDAP structs
				Expect(func() {
					_ = cpt.GetLDAP()
				}).To(Panic())
			})
		})
	})

	Describe("Dependencies edge cases", func() {
		Context("before initialization", func() {
			It("should work when setting dependencies before Init", func() {
				// After New(), internal structures are initialized
				newCpt := New(ctx)
				err := newCpt.SetDependencies([]string{"dep1"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return empty dependencies before Init", func() {
				newCpt := New(ctx)
				deps := newCpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})
		})

		Context("after initialization", func() {
			BeforeEach(func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }
				cpt.Init("test", ctx, getCpt, vpr, vrs, log)
			})

			It("should handle dependency updates", func() {
				err := cpt.SetDependencies([]string{"dep1", "dep2"})
				Expect(err).To(BeNil())

				err = cpt.SetDependencies([]string{"dep3"})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(Equal([]string{"dep3"}))
			})

			It("should handle clearing dependencies", func() {
				err := cpt.SetDependencies([]string{"dep1", "dep2"})
				Expect(err).To(BeNil())

				err = cpt.SetDependencies([]string{})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})
		})
	})

	Describe("Callback edge cases", func() {
		BeforeEach(func() {
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }
			cpt.Init("test", ctx, getCpt, vpr, vrs, log)
		})

		Context("callback registration", func() {
			It("should handle multiple callback registrations", func() {
				before1 := func(c cfgtps.Component) error { return nil }
				after1 := func(c cfgtps.Component) error { return nil }
				cpt.RegisterFuncStart(before1, after1)

				before2 := func(c cfgtps.Component) error { return nil }
				after2 := func(c cfgtps.Component) error { return nil }
				cpt.RegisterFuncStart(before2, after2)

				Expect(func() {
					_ = cpt.Start()
				}).NotTo(Panic())
			})

			It("should handle nil callback registration", func() {
				cpt.RegisterFuncStart(nil, nil)
				cpt.RegisterFuncReload(nil, nil)

				Expect(func() {
					_ = cpt.Start()
					_ = cpt.Reload()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Monitor pool registration", func() {
		Context("RegisterMonitorPool", func() {
			It("should not panic when registering monitor pool", func() {
				Expect(func() {
					cpt.RegisterMonitorPool(nil)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Concurrent operations", func() {
		BeforeEach(func() {
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }
			cpt.Init("test", ctx, getCpt, vpr, vrs, log)
		})

		Context("concurrent config updates", func() {
			It("should handle concurrent GetConfig calls", func() {
				cfg := &lbldap.Config{
					Uri:      "ldap://localhost",
					PortLdap: 389,
					Basedn:   "dc=example,dc=com",
				}
				cpt.SetConfig(cfg)

				done := make(chan bool, 10)
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.GetConfig()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})

			It("should handle concurrent GetAttributes calls", func() {
				cpt.SetAttributes([]string{"uid", "cn", "mail"})

				done := make(chan bool, 10)
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.GetAttributes()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})

		Context("concurrent lifecycle operations", func() {
			It("should handle concurrent IsStarted calls", func() {
				done := make(chan bool, 10)
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.IsStarted()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})
})
