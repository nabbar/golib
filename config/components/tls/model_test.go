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

package tls_test

import (
	"context"

	. "github.com/nabbar/golib/config/components/tls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libtls "github.com/nabbar/golib/certificates"
)

// Model tests verify the internal model behavior for TLS component.
// These tests focus on the Config, GetTLS, and SetTLS methods.
var _ = Describe("Model Methods", func() {
	var (
		ctx context.Context
		cpt CptTlS
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil)
	})

	Describe("Config method", func() {
		Context("getting configuration", func() {
			It("should return nil when not configured", func() {
				config := cpt.Config()
				Expect(config).To(BeNil())
			})

			It("should not panic on uninitialized component", func() {
				Expect(func() {
					_ = cpt.Config()
				}).NotTo(Panic())
			})
		})

		Context("concurrent access", func() {
			It("should handle concurrent Config calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.Config()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("GetTLS method", func() {
		Context("getting TLS configuration", func() {
			It("should return nil initially", func() {
				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).To(BeNil())
			})

			It("should not panic on uninitialized component", func() {
				Expect(func() {
					_ = cpt.GetTLS()
				}).NotTo(Panic())
			})

			It("should return TLS config after SetTLS", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})

			It("should return cloned TLS config", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				tlsCfg1 := cpt.GetTLS()
				tlsCfg2 := cpt.GetTLS()

				// Should be different instances (cloned)
				Expect(tlsCfg1).NotTo(BeIdenticalTo(tlsCfg2))
			})
		})

		Context("concurrent access", func() {
			It("should handle concurrent GetTLS calls", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						tlsCfg := cpt.GetTLS()
						Expect(tlsCfg).NotTo(BeNil())
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("SetTLS method", func() {
		Context("setting TLS configuration", func() {
			It("should accept valid TLS config", func() {
				newTLS := libtls.New()

				Expect(func() {
					cpt.SetTLS(newTLS)
				}).NotTo(Panic())
			})

			It("should handle nil TLS config", func() {
				Expect(func() {
					cpt.SetTLS(nil)
				}).NotTo(Panic())

				// Should create a new TLS config when nil is passed
				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})

			It("should allow changing TLS config multiple times", func() {
				tls1 := libtls.New()
				tls2 := libtls.New()
				tls3 := libtls.New()

				cpt.SetTLS(tls1)
				cpt.SetTLS(tls2)
				cpt.SetTLS(tls3)
				// Should not panic
			})

			It("should update TLS config", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				retrieved := cpt.GetTLS()
				Expect(retrieved).NotTo(BeNil())
			})
		})

		Context("concurrent updates", func() {
			It("should handle concurrent SetTLS calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						newTLS := libtls.New()
						cpt.SetTLS(newTLS)
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Integration between methods", func() {
		Context("combined operations", func() {
			It("should handle SetTLS and GetTLS together", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})

			It("should handle all getters in sequence", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				config := cpt.Config()
				tlsCfg := cpt.GetTLS()

				// Config might be nil if not properly initialized, but GetTLS should work
				Expect(tlsCfg).NotTo(BeNil())
				_ = config // May be nil, which is fine
			})
		})

		Context("concurrent mixed operations", func() {
			It("should handle concurrent reads and writes", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				done := make(chan bool, 30)

				// 10 GetTLS readers
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						tlsCfg := cpt.GetTLS()
						Expect(tlsCfg).NotTo(BeNil())
						done <- true
					}()
				}

				// 10 Config readers
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = cpt.Config()
						done <- true
					}()
				}

				// 10 SetTLS writers
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						newTLS := libtls.New()
						cpt.SetTLS(newTLS)
						done <- true
					}()
				}

				for i := 0; i < 30; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Edge cases", func() {
		Context("with various TLS configurations", func() {
			It("should accept freshly created TLS config", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})

			It("should handle repeated nil sets", func() {
				cpt.SetTLS(nil)
				cpt.SetTLS(nil)
				cpt.SetTLS(nil)

				// Should still have a valid TLS config
				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})

			It("should handle alternating nil and valid sets", func() {
				cpt.SetTLS(nil)
				cpt.SetTLS(libtls.New())
				cpt.SetTLS(nil)
				cpt.SetTLS(libtls.New())

				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})
		})

		Context("state after Stop", func() {
			It("should reset TLS config after Stop", func() {
				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				cpt.Stop()

				// After stop, should still have a TLS config (initialized to new)
				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})

			It("should allow SetTLS after Stop", func() {
				cpt.Stop()

				newTLS := libtls.New()
				cpt.SetTLS(newTLS)

				tlsCfg := cpt.GetTLS()
				Expect(tlsCfg).NotTo(BeNil())
			})
		})
	})
})
