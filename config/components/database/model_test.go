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

package database_test

import (
	"context"
	"time"

	. "github.com/nabbar/golib/config/components/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libdur "github.com/nabbar/golib/duration"
)

// Model tests verify SetLogOptions, SetDatabase, and GetDatabase functions
var _ = Describe("Model Functions", func() {
	var (
		cpt CptDatabase
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
	})

	Describe("SetLogOptions", func() {
		It("should set ignore record not found error to true", func() {
			Expect(func() {
				cpt.SetLogOptions(true, 0)
			}).NotTo(Panic())
		})

		It("should set ignore record not found error to false", func() {
			Expect(func() {
				cpt.SetLogOptions(false, 0)
			}).NotTo(Panic())
		})

		It("should set slow threshold", func() {
			slowThreshold := libdur.Duration(time.Second * 5)
			Expect(func() {
				cpt.SetLogOptions(false, slowThreshold)
			}).NotTo(Panic())
		})

		It("should set zero slow threshold", func() {
			Expect(func() {
				cpt.SetLogOptions(true, 0)
			}).NotTo(Panic())
		})

		It("should set negative slow threshold", func() {
			slowThreshold := libdur.Duration(-time.Second)
			Expect(func() {
				cpt.SetLogOptions(true, slowThreshold)
			}).NotTo(Panic())
		})

		It("should set very large slow threshold", func() {
			slowThreshold := libdur.Duration(time.Hour * 24 * 365)
			Expect(func() {
				cpt.SetLogOptions(false, slowThreshold)
			}).NotTo(Panic())
		})

		It("should be callable multiple times", func() {
			cpt.SetLogOptions(true, 100)
			cpt.SetLogOptions(false, 200)
			cpt.SetLogOptions(true, 300)
			cpt.SetLogOptions(false, 0)
		})

		It("should allow switching ignore flag", func() {
			cpt.SetLogOptions(true, 100)
			cpt.SetLogOptions(false, 100)
			cpt.SetLogOptions(true, 100)
		})

		It("should allow changing slow threshold", func() {
			cpt.SetLogOptions(true, 0)
			cpt.SetLogOptions(true, 100)
			cpt.SetLogOptions(true, 1000)
			cpt.SetLogOptions(true, 10000)
		})
	})

	Describe("SetDatabase and GetDatabase", func() {
		It("should return nil when no database is set", func() {
			db := cpt.GetDatabase()
			Expect(db).To(BeNil())
		})

		It("should handle setting nil database", func() {
			Expect(func() {
				cpt.SetDatabase(nil)
			}).NotTo(Panic())

			db := cpt.GetDatabase()
			Expect(db).To(BeNil())
		})

		It("should be callable multiple times with nil", func() {
			cpt.SetDatabase(nil)
			cpt.SetDatabase(nil)
			cpt.SetDatabase(nil)

			db := cpt.GetDatabase()
			Expect(db).To(BeNil())
		})

		It("should maintain nil state after multiple operations", func() {
			cpt.SetDatabase(nil)
			db1 := cpt.GetDatabase()
			Expect(db1).To(BeNil())

			cpt.SetDatabase(nil)
			db2 := cpt.GetDatabase()
			Expect(db2).To(BeNil())
		})
	})

	Describe("Combined operations", func() {
		It("should allow setting log options and database independently", func() {
			cpt.SetLogOptions(true, 100)
			cpt.SetDatabase(nil)

			db := cpt.GetDatabase()
			Expect(db).To(BeNil())
		})

		It("should maintain log options when database is set", func() {
			cpt.SetLogOptions(true, 500)
			cpt.SetDatabase(nil)
			// Setting database shouldn't affect log options
			// (no way to verify directly, but shouldn't panic)
		})

		It("should maintain database state when log options change", func() {
			cpt.SetDatabase(nil)
			cpt.SetLogOptions(true, 100)
			cpt.SetLogOptions(false, 200)

			db := cpt.GetDatabase()
			Expect(db).To(BeNil())
		})
	})
})

// Concurrent access to model functions
var _ = Describe("Model Thread Safety", func() {
	var (
		cpt CptDatabase
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	AfterEach(func() {
		if cpt != nil && cpt.IsStarted() {
			cpt.Stop()
		}
	})

	It("should handle concurrent SetLogOptions calls", func() {
		done := make(chan bool, 20)

		for i := 0; i < 20; i++ {
			go func(idx int) {
				defer GinkgoRecover()
				ignoreError := idx%2 == 0
				slowThreshold := libdur.Duration(time.Millisecond * time.Duration(idx*10))
				cpt.SetLogOptions(ignoreError, slowThreshold)
				done <- true
			}(i)
		}

		for i := 0; i < 20; i++ {
			Eventually(done).Should(Receive())
		}
	})

	It("should handle concurrent GetDatabase calls", func() {
		done := make(chan bool, 20)

		for i := 0; i < 20; i++ {
			go func() {
				defer GinkgoRecover()
				db := cpt.GetDatabase()
				Expect(db).To(BeNil())
				done <- true
			}()
		}

		for i := 0; i < 20; i++ {
			Eventually(done).Should(Receive())
		}
	})

	It("should handle concurrent SetDatabase calls", func() {
		done := make(chan bool, 20)

		for i := 0; i < 20; i++ {
			go func() {
				defer GinkgoRecover()
				cpt.SetDatabase(nil)
				done <- true
			}()
		}

		for i := 0; i < 20; i++ {
			Eventually(done).Should(Receive())
		}
	})

	It("should handle mixed concurrent operations", func() {
		done := make(chan bool, 30)

		// Concurrent SetLogOptions
		for i := 0; i < 10; i++ {
			go func(idx int) {
				defer GinkgoRecover()
				cpt.SetLogOptions(idx%2 == 0, libdur.Duration(idx*100))
				done <- true
			}(i)
		}

		// Concurrent GetDatabase
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				db := cpt.GetDatabase()
				Expect(db).To(BeNil())
				done <- true
			}()
		}

		// Concurrent SetDatabase
		for i := 0; i < 10; i++ {
			go func() {
				defer GinkgoRecover()
				cpt.SetDatabase(nil)
				done <- true
			}()
		}

		for i := 0; i < 30; i++ {
			Eventually(done).Should(Receive())
		}
	})
})

// Edge cases for model functions
var _ = Describe("Model Edge Cases", func() {
	var (
		cpt CptDatabase
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	Context("with extreme values", func() {
		It("should handle maximum slow threshold", func() {
			maxDuration := libdur.Duration(time.Hour * 24 * 365 * 100)
			Expect(func() {
				cpt.SetLogOptions(true, maxDuration)
			}).NotTo(Panic())
		})

		It("should handle minimum slow threshold", func() {
			minDuration := libdur.Duration(-time.Hour * 24 * 365)
			Expect(func() {
				cpt.SetLogOptions(false, minDuration)
			}).NotTo(Panic())
		})
	})

	Context("with rapid state changes", func() {
		It("should handle rapid log option changes", func() {
			for i := 0; i < 100; i++ {
				cpt.SetLogOptions(i%2 == 0, libdur.Duration(i))
			}
		})

		It("should handle rapid database set/get cycles", func() {
			for i := 0; i < 100; i++ {
				cpt.SetDatabase(nil)
				db := cpt.GetDatabase()
				Expect(db).To(BeNil())
			}
		})

		It("should handle alternating operations", func() {
			for i := 0; i < 50; i++ {
				cpt.SetLogOptions(true, libdur.Duration(i*10))
				cpt.SetDatabase(nil)
				db := cpt.GetDatabase()
				Expect(db).To(BeNil())
			}
		})
	})

	Context("after component stop", func() {
		It("should handle operations after stop", func() {
			cpt.Stop()

			Expect(func() {
				cpt.SetLogOptions(true, 100)
			}).NotTo(Panic())

			Expect(func() {
				cpt.SetDatabase(nil)
			}).NotTo(Panic())

			Expect(func() {
				db := cpt.GetDatabase()
				Expect(db).To(BeNil())
			}).NotTo(Panic())
		})
	})
})
