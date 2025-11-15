/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package gin_test

import (
	ginsdk "github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libgin "github.com/nabbar/golib/context/gin"
	liblog "github.com/nabbar/golib/logger"
)

var _ = Describe("GinTonic Basic Operations", func() {
	var (
		ginCtx  *ginsdk.Context
		gtx     libgin.GinTonic
		logFunc liblog.FuncLog
	)

	BeforeEach(func() {
		// Set Gin to test mode to reduce noise
		ginsdk.SetMode(ginsdk.TestMode)

		// Create a test Gin context
		ginCtx, _ = ginsdk.CreateTestContext(nil)

		// Create a simple logger function (nil is acceptable)
		logFunc = nil

		// Create GinTonic context
		gtx = libgin.New(ginCtx, logFunc)
	})

	Describe("New", func() {
		It("should create a valid GinTonic instance", func() {
			Expect(gtx).ToNot(BeNil())
		})

		It("should create GinTonic with nil Gin context", func() {
			gtxNil := libgin.New(nil, nil)
			Expect(gtxNil).ToNot(BeNil())
			Expect(gtxNil.GinContext()).ToNot(BeNil())
		})

		It("should create GinTonic with custom logger", func() {
			customLog := func() liblog.Logger {
				return liblog.New(nil)
			}
			gtxCustom := libgin.New(ginCtx, customLog)
			Expect(gtxCustom).ToNot(BeNil())
		})
	})

	Describe("GinContext", func() {
		It("should return the underlying Gin context", func() {
			ctx := gtx.GinContext()
			Expect(ctx).ToNot(BeNil())
			Expect(ctx).To(Equal(ginCtx))
		})
	})

	Describe("Set and Get", func() {
		It("should store and retrieve a string value", func() {
			gtx.Set("key", "value")
			val, exists := gtx.Get("key")
			Expect(exists).To(BeTrue())
			Expect(val).To(Equal("value"))
		})

		It("should store and retrieve an integer value", func() {
			gtx.Set("count", 42)
			val, exists := gtx.Get("count")
			Expect(exists).To(BeTrue())
			Expect(val).To(Equal(42))
		})

		It("should store and retrieve a boolean value", func() {
			gtx.Set("active", true)
			val, exists := gtx.Get("active")
			Expect(exists).To(BeTrue())
			Expect(val).To(Equal(true))
		})

		It("should return false for non-existent key", func() {
			val, exists := gtx.Get("nonexistent")
			Expect(exists).To(BeFalse())
			Expect(val).To(BeNil())
		})

		It("should overwrite existing value", func() {
			gtx.Set("key", "value1")
			gtx.Set("key", "value2")
			val, exists := gtx.Get("key")
			Expect(exists).To(BeTrue())
			Expect(val).To(Equal("value2"))
		})

		It("should handle nil values", func() {
			gtx.Set("nilkey", nil)
			val, exists := gtx.Get("nilkey")
			Expect(exists).To(BeTrue())
			Expect(val).To(BeNil())
		})

		It("should store multiple key-value pairs", func() {
			gtx.Set("key1", "value1")
			gtx.Set("key2", 100)
			gtx.Set("key3", true)

			val1, exists1 := gtx.Get("key1")
			val2, exists2 := gtx.Get("key2")
			val3, exists3 := gtx.Get("key3")

			Expect(exists1).To(BeTrue())
			Expect(val1).To(Equal("value1"))
			Expect(exists2).To(BeTrue())
			Expect(val2).To(Equal(100))
			Expect(exists3).To(BeTrue())
			Expect(val3).To(Equal(true))
		})
	})

	Describe("MustGet", func() {
		It("should return value for existing key", func() {
			gtx.Set("key", "value")
			val := gtx.MustGet("key")
			Expect(val).To(Equal("value"))
		})

		It("should panic for non-existent key", func() {
			Expect(func() {
				gtx.MustGet("nonexistent")
			}).To(Panic())
		})
	})

	Describe("Value", func() {
		It("should return value from Gin context", func() {
			gtx.Set("contextKey", "contextValue")
			val := gtx.Value("contextKey")
			Expect(val).To(Equal("contextValue"))
		})

		It("should return nil for non-existent key", func() {
			val := gtx.Value("nonexistent")
			Expect(val).To(BeNil())
		})
	})

	Describe("SetLogger", func() {
		It("should set a custom logger", func() {
			customLog := func() liblog.Logger {
				return liblog.New(nil)
			}
			Expect(func() {
				gtx.SetLogger(customLog)
			}).ToNot(Panic())
		})

		It("should accept nil logger", func() {
			Expect(func() {
				gtx.SetLogger(nil)
			}).ToNot(Panic())
		})
	})
})
