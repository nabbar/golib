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
	"time"

	ginsdk "github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libgin "github.com/nabbar/golib/context/gin"
)

var _ = Describe("GinTonic Type Getters", func() {
	var (
		ginCtx *ginsdk.Context
		gtx    libgin.GinTonic
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		ginCtx, _ = ginsdk.CreateTestContext(nil)
		gtx = libgin.New(ginCtx, nil)
	})

	Describe("GetString", func() {
		It("should return string value for existing key", func() {
			gtx.Set("name", "alice")
			val := gtx.GetString("name")
			Expect(val).To(Equal("alice"))
		})

		It("should return empty string for non-existent key", func() {
			val := gtx.GetString("nonexistent")
			Expect(val).To(Equal(""))
		})

		It("should return empty string for non-string value", func() {
			gtx.Set("number", 123)
			val := gtx.GetString("number")
			Expect(val).To(Equal(""))
		})
	})

	Describe("GetBool", func() {
		It("should return true for boolean true", func() {
			gtx.Set("active", true)
			val := gtx.GetBool("active")
			Expect(val).To(BeTrue())
		})

		It("should return false for boolean false", func() {
			gtx.Set("inactive", false)
			val := gtx.GetBool("inactive")
			Expect(val).To(BeFalse())
		})

		It("should return false for non-existent key", func() {
			val := gtx.GetBool("nonexistent")
			Expect(val).To(BeFalse())
		})

		It("should return false for non-boolean value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetBool("text")
			Expect(val).To(BeFalse())
		})
	})

	Describe("GetInt", func() {
		It("should return integer value for existing key", func() {
			gtx.Set("count", 42)
			val := gtx.GetInt("count")
			Expect(val).To(Equal(42))
		})

		It("should return 0 for non-existent key", func() {
			val := gtx.GetInt("nonexistent")
			Expect(val).To(Equal(0))
		})

		It("should return 0 for non-integer value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetInt("text")
			Expect(val).To(Equal(0))
		})

		It("should handle negative integers", func() {
			gtx.Set("negative", -10)
			val := gtx.GetInt("negative")
			Expect(val).To(Equal(-10))
		})
	})

	Describe("GetInt64", func() {
		It("should return int64 value for existing key", func() {
			var largeNumber int64 = 9876543210
			gtx.Set("large", largeNumber)
			val := gtx.GetInt64("large")
			Expect(val).To(Equal(largeNumber))
		})

		It("should return 0 for non-existent key", func() {
			val := gtx.GetInt64("nonexistent")
			Expect(val).To(Equal(int64(0)))
		})

		It("should return 0 for non-int64 value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetInt64("text")
			Expect(val).To(Equal(int64(0)))
		})
	})

	Describe("GetFloat64", func() {
		It("should return float64 value for existing key", func() {
			gtx.Set("price", 19.99)
			val := gtx.GetFloat64("price")
			Expect(val).To(Equal(19.99))
		})

		It("should return 0.0 for non-existent key", func() {
			val := gtx.GetFloat64("nonexistent")
			Expect(val).To(Equal(0.0))
		})

		It("should return 0.0 for non-float64 value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetFloat64("text")
			Expect(val).To(Equal(0.0))
		})

		It("should handle negative floats", func() {
			gtx.Set("negative", -3.14)
			val := gtx.GetFloat64("negative")
			Expect(val).To(Equal(-3.14))
		})
	})

	Describe("GetTime", func() {
		It("should return time value for existing key", func() {
			now := time.Now()
			gtx.Set("timestamp", now)
			val := gtx.GetTime("timestamp")
			Expect(val).To(BeTemporally("~", now, time.Millisecond))
		})

		It("should return zero time for non-existent key", func() {
			val := gtx.GetTime("nonexistent")
			Expect(val).To(BeZero())
		})

		It("should return zero time for non-time value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetTime("text")
			Expect(val).To(BeZero())
		})
	})

	Describe("GetDuration", func() {
		It("should return duration value for existing key", func() {
			duration := 5 * time.Minute
			gtx.Set("timeout", duration)
			val := gtx.GetDuration("timeout")
			Expect(val).To(Equal(duration))
		})

		It("should return zero duration for non-existent key", func() {
			val := gtx.GetDuration("nonexistent")
			Expect(val).To(Equal(time.Duration(0)))
		})

		It("should return zero duration for non-duration value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetDuration("text")
			Expect(val).To(Equal(time.Duration(0)))
		})

		It("should handle various durations", func() {
			gtx.Set("sec", 30*time.Second)
			gtx.Set("min", 15*time.Minute)
			gtx.Set("hour", 2*time.Hour)

			Expect(gtx.GetDuration("sec")).To(Equal(30 * time.Second))
			Expect(gtx.GetDuration("min")).To(Equal(15 * time.Minute))
			Expect(gtx.GetDuration("hour")).To(Equal(2 * time.Hour))
		})
	})

	Describe("GetStringSlice", func() {
		It("should return string slice for existing key", func() {
			slice := []string{"apple", "banana", "cherry"}
			gtx.Set("fruits", slice)
			val := gtx.GetStringSlice("fruits")
			Expect(val).To(Equal(slice))
		})

		It("should return nil for non-existent key", func() {
			val := gtx.GetStringSlice("nonexistent")
			Expect(val).To(BeNil())
		})

		It("should return nil for non-slice value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetStringSlice("text")
			Expect(val).To(BeNil())
		})

		It("should handle empty slice", func() {
			gtx.Set("empty", []string{})
			val := gtx.GetStringSlice("empty")
			Expect(val).To(Equal([]string{}))
		})
	})

	Describe("GetStringMap", func() {
		It("should return string map for existing key", func() {
			m := map[string]any{
				"name": "alice",
				"age":  30,
			}
			gtx.Set("user", m)
			val := gtx.GetStringMap("user")
			Expect(val).To(Equal(m))
		})

		It("should return nil for non-existent key", func() {
			val := gtx.GetStringMap("nonexistent")
			Expect(val).To(BeNil())
		})

		It("should return nil for non-map value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetStringMap("text")
			Expect(val).To(BeNil())
		})

		It("should handle empty map", func() {
			gtx.Set("empty", map[string]any{})
			val := gtx.GetStringMap("empty")
			Expect(val).To(Equal(map[string]any{}))
		})
	})

	Describe("GetStringMapString", func() {
		It("should return string-to-string map for existing key", func() {
			m := map[string]string{
				"name":  "alice",
				"email": "alice@example.com",
			}
			gtx.Set("contact", m)
			val := gtx.GetStringMapString("contact")
			Expect(val).To(Equal(m))
		})

		It("should return nil for non-existent key", func() {
			val := gtx.GetStringMapString("nonexistent")
			Expect(val).To(BeNil())
		})

		It("should return nil for non-map value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetStringMapString("text")
			Expect(val).To(BeNil())
		})

		It("should handle empty map", func() {
			gtx.Set("empty", map[string]string{})
			val := gtx.GetStringMapString("empty")
			Expect(val).To(Equal(map[string]string{}))
		})
	})

	Describe("GetStringMapStringSlice", func() {
		It("should return string-to-string-slice map for existing key", func() {
			m := map[string][]string{
				"colors": {"red", "green", "blue"},
				"sizes":  {"small", "medium", "large"},
			}
			gtx.Set("options", m)
			val := gtx.GetStringMapStringSlice("options")
			Expect(val).To(Equal(m))
		})

		It("should return nil for non-existent key", func() {
			val := gtx.GetStringMapStringSlice("nonexistent")
			Expect(val).To(BeNil())
		})

		It("should return nil for non-map value", func() {
			gtx.Set("text", "hello")
			val := gtx.GetStringMapStringSlice("text")
			Expect(val).To(BeNil())
		})

		It("should handle empty map", func() {
			gtx.Set("empty", map[string][]string{})
			val := gtx.GetStringMapStringSlice("empty")
			Expect(val).To(Equal(map[string][]string{}))
		})

		It("should handle map with empty slices", func() {
			m := map[string][]string{
				"empty1": {},
				"empty2": {},
			}
			gtx.Set("empties", m)
			val := gtx.GetStringMapStringSlice("empties")
			Expect(val).To(Equal(m))
		})
	})
})
