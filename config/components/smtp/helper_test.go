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

package smtp_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/smtp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Helper tests verify edge cases, error handling, and special scenarios
// for the SMTP component.
var _ = Describe("Helper Functions and Edge Cases", func() {
	var (
		cpt CptSMTP
		ctx context.Context
		vrs libver.Version
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
		cpt = New(ctx, "")
	})

	Describe("Edge cases", func() {
		Context("nil component operations", func() {
			It("should handle nil context gracefully", func() {
				Expect(func() {
					_ = New(nil, "")
				}).NotTo(Panic())
			})

			It("should handle operations before initialization", func() {
				newCpt := New(ctx, "")

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

		Context("TLS key edge cases", func() {
			It("should handle TLS key updates", func() {
				cpt.SetTLSKey("key1")
				cpt.SetTLSKey("key2")
				cpt.SetTLSKey("key3")

				Expect(cpt).NotTo(BeNil())
			})

			It("should handle empty TLS key", func() {
				Expect(func() {
					cpt.SetTLSKey("")
				}).NotTo(Panic())
			})
		})
	})

	Describe("Dependencies edge cases", func() {
		Context("before initialization", func() {
			It("should return default TLS dependency", func() {
				newCpt := New(ctx, "custom-tls")
				deps := newCpt.Dependencies()
				Expect(deps).To(ContainElement("custom-tls"))
			})

			It("should work when setting dependencies before Init", func() {
				newCpt := New(ctx, "")
				err := newCpt.SetDependencies([]string{"dep1"})
				Expect(err).NotTo(HaveOccurred())
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

				// Should return default TLS dependency
				deps := cpt.Dependencies()
				Expect(deps).NotTo(BeEmpty())
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

	Describe("GetSMTP edge cases", func() {
		Context("before component start", func() {
			It("should return error when not started", func() {
				_, err := cpt.GetSMTP()
				Expect(err).To(HaveOccurred())
			})

			It("should return error even after Init", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }
				cpt.Init("test", ctx, getCpt, vpr, vrs, log)

				_, err := cpt.GetSMTP()
				Expect(err).To(HaveOccurred())
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

		Context("concurrent TLS key updates", func() {
			It("should handle concurrent SetTLSKey calls", func() {
				done := make(chan bool, 10)
				for i := 0; i < 10; i++ {
					go func(id int) {
						defer GinkgoRecover()
						cpt.SetTLSKey("key")
						done <- true
					}(i)
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

			It("should handle concurrent Dependencies calls", func() {
				done := make(chan bool, 10)
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.Dependencies()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Component state consistency", func() {
		Context("state checks", func() {
			It("should maintain consistent state before init", func() {
				newCpt := New(ctx, "")
				Expect(newCpt.IsStarted()).To(BeFalse())
				Expect(newCpt.IsRunning()).To(BeFalse())
			})

			It("should maintain consistent state after init", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }
				cpt.Init("test", ctx, getCpt, vpr, vrs, log)

				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())
			})
		})
	})
})
