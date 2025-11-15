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

package mail_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/mail"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgtps "github.com/nabbar/golib/config/types"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// Lifecycle tests verify component initialization, Start/Reload/Stop operations,
// and callback execution for the Mail component.
var _ = Describe("Component Lifecycle", func() {
	var (
		cpt CptMail
		ctx context.Context
		vrs libver.Version
		key string
	)

	BeforeEach(func() {
		ctx = context.Background()
		vrs = libver.NewVersion(libver.License_MIT, "test", "", "01/01/1970", "test", "1.0.0", "test", "", struct{}{}, 0)
		key = "test-mail"
		cpt = New(ctx)
	})

	Describe("Component initialization", func() {
		Context("Init method", func() {
			It("should initialize without error", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }

				Expect(func() {
					cpt.Init(key, ctx, getCpt, vpr, vrs, log)
				}).NotTo(Panic())
			})

			It("should store initialization parameters", func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }

				cpt.Init(key, ctx, getCpt, vpr, vrs, log)

				Expect(cpt.Type()).To(Equal("smtp"))
			})
		})
	})

	Describe("Lifecycle operations", func() {
		BeforeEach(func() {
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }
			cpt.Init(key, ctx, getCpt, vpr, vrs, log)
		})

		Context("IsStarted and IsRunning", func() {
			It("should return false before Start", func() {
				Expect(cpt.IsStarted()).To(BeFalse())
				Expect(cpt.IsRunning()).To(BeFalse())
			})
		})

		Context("Start operation", func() {
			It("should not panic on Start", func() {
				Expect(func() {
					_ = cpt.Start()
				}).NotTo(Panic())
			})

			It("should handle start without configuration", func() {
				err := cpt.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Reload operation", func() {
			It("should not panic on Reload", func() {
				Expect(func() {
					_ = cpt.Reload()
				}).NotTo(Panic())
			})
		})

		Context("Stop operation", func() {
			It("should not panic on Stop", func() {
				Expect(func() {
					cpt.Stop()
				}).NotTo(Panic())
			})

			It("should update state after Stop", func() {
				cpt.Stop()
				Expect(cpt.IsStarted()).To(BeFalse())
			})

			It("should handle multiple Stop calls", func() {
				Expect(func() {
					cpt.Stop()
					cpt.Stop()
					cpt.Stop()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Callback registration", func() {
		BeforeEach(func() {
			getCpt := func(k string) cfgtps.Component { return nil }
			vpr := func() libvpr.Viper { return nil }
			log := func() liblog.Logger { return nil }
			cpt.Init(key, ctx, getCpt, vpr, vrs, log)
		})

		Context("RegisterFuncStart", func() {
			It("should register start callbacks", func() {
				before := func(c cfgtps.Component) error { return nil }
				after := func(c cfgtps.Component) error { return nil }

				Expect(func() {
					cpt.RegisterFuncStart(before, after)
				}).NotTo(Panic())
			})

			It("should accept nil callbacks", func() {
				Expect(func() {
					cpt.RegisterFuncStart(nil, nil)
				}).NotTo(Panic())
			})
		})

		Context("RegisterFuncReload", func() {
			It("should register reload callbacks", func() {
				before := func(c cfgtps.Component) error { return nil }
				after := func(c cfgtps.Component) error { return nil }

				Expect(func() {
					cpt.RegisterFuncReload(before, after)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Dependencies", func() {
		Context("Dependencies method", func() {
			It("should return empty dependencies by default", func() {
				deps := cpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})
		})

		Context("SetDependencies and Dependencies", func() {
			BeforeEach(func() {
				getCpt := func(k string) cfgtps.Component { return nil }
				vpr := func() libvpr.Viper { return nil }
				log := func() liblog.Logger { return nil }
				cpt.Init(key, ctx, getCpt, vpr, vrs, log)
			})

			It("should set and retrieve dependencies", func() {
				err := cpt.SetDependencies([]string{"dep1", "dep2"})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(Equal([]string{"dep1", "dep2"}))
			})

			It("should handle empty dependencies", func() {
				err := cpt.SetDependencies([]string{})
				Expect(err).To(BeNil())

				deps := cpt.Dependencies()
				Expect(deps).To(BeEmpty())
			})

			It("should handle nil dependencies", func() {
				err := cpt.SetDependencies(nil)
				Expect(err).To(BeNil())
			})
		})

		Context("SetDependencies errors", func() {
			It("should work even without Init call", func() {
				newCpt := New(ctx)
				err := newCpt.SetDependencies([]string{"dep1"})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Component state", func() {
		Context("Type method", func() {
			It("should return smtp type", func() {
				Expect(cpt.Type()).To(Equal("smtp"))
			})
		})

		Context("GetMail before start", func() {
			It("should return error when not started", func() {
				_, err := cpt.GetMail()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
