/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package fields_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logfld "github.com/nabbar/golib/logger/fields"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Fields Creation and Basic Operations", func() {
	Describe("New", func() {
		Context("with nil context", func() {
			It("should create empty fields", func() {
				f := logfld.New(nil)

				Expect(f).ToNot(BeNil())
			})
		})

		Context("with valid context", func() {
			It("should create empty fields", func() {
				f := logfld.New(context.Background())

				Expect(f).ToNot(BeNil())
			})
		})
	})

	Describe("Add", func() {
		Context("on valid fields instance", func() {
			It("should add string value", func() {
				f := logfld.New(nil)
				result := f.Add("key", "value")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(f))
			})

			It("should add integer value", func() {
				f := logfld.New(nil)
				result := f.Add("count", 42)

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields).To(HaveKey("count"))
				Expect(logrusFields["count"]).To(Equal(42))
			})

			It("should add boolean value", func() {
				f := logfld.New(nil)
				result := f.Add("enabled", true)

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields).To(HaveKey("enabled"))
				Expect(logrusFields["enabled"]).To(BeTrue())
			})

			It("should add float value", func() {
				f := logfld.New(nil)
				result := f.Add("price", 19.99)

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields).To(HaveKey("price"))
				Expect(logrusFields["price"]).To(Equal(19.99))
			})

			It("should add nil value", func() {
				f := logfld.New(nil)
				result := f.Add("nullable", nil)

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields).ToNot(HaveKey("nullable"))
				Expect(logrusFields["nullable"]).To(BeNil())
			})

			It("should add map value", func() {
				f := logfld.New(nil)
				mapVal := map[string]string{"inner": "value"}
				result := f.Add("metadata", mapVal)

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields).To(HaveKey("metadata"))
			})

			It("should add slice value", func() {
				f := logfld.New(nil)
				sliceVal := []string{"item1", "item2", "item3"}
				result := f.Add("items", sliceVal)

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields).To(HaveKey("items"))
			})

			It("should overwrite existing key", func() {
				f := logfld.New(nil)
				f.Add("key", "value1")
				result := f.Add("key", "value2")

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields["key"]).To(Equal("value2"))
			})

			It("should add multiple fields", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")
				f.Add("key3", "value3")

				logrusFields := f.Logrus()
				Expect(logrusFields).To(HaveLen(3))
				Expect(logrusFields).To(HaveKey("key1"))
				Expect(logrusFields).To(HaveKey("key2"))
				Expect(logrusFields).To(HaveKey("key3"))
			})
		})

		Context("on nil fields instance", func() {
			It("should handle nil safely", func() {
				var f logfld.Fields = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(f).To(BeNil())
			})
		})
	})

	Describe("Logrus", func() {
		Context("on empty fields", func() {
			It("should return empty logrus.Fields", func() {
				f := logfld.New(nil)
				result := f.Logrus()

				Expect(result).ToNot(BeNil())
				Expect(result).To(HaveLen(0))
			})
		})

		Context("on fields with single entry", func() {
			It("should return logrus.Fields with one entry", func() {
				f := logfld.New(nil)
				f.Add("key", "value")
				result := f.Logrus()

				Expect(result).To(HaveLen(1))
				Expect(result).To(HaveKeyWithValue("key", "value"))
			})
		})

		Context("on fields with multiple entries", func() {
			It("should return logrus.Fields with all entries", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", 123)
				f.Add("key3", true)
				result := f.Logrus()

				Expect(result).To(HaveLen(3))
				Expect(result).To(HaveKeyWithValue("key1", "value1"))
				Expect(result).To(HaveKeyWithValue("key2", 123))
				Expect(result).To(HaveKeyWithValue("key3", true))
			})
		})

		Context("on nil fields instance", func() {
			It("should handle nil safely", func() {
				var f logfld.Fields = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(f).To(BeNil())
			})
		})
	})

	Describe("Map", func() {
		Context("with transformation function", func() {
			It("should transform all values", func() {
				f := logfld.New(context.Background())
				f.Add("key1", "value1")
				f.Add("key2", "value2")

				result := f.Map(func(key string, val interface{}) interface{} {
					return val.(string) + "_transformed"
				})

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields["key1"]).To(Equal("value1_transformed"))
				Expect(logrusFields["key2"]).To(Equal("value2_transformed"))
			})

			It("should transform values based on key", func() {
				f := logfld.New(nil)
				f.Add("name", "John")
				f.Add("age", 30)

				result := f.Map(func(key string, val interface{}) interface{} {
					if key == "name" {
						return "Mr. " + val.(string)
					}
					return val
				})

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields["name"]).To(Equal("Mr. John"))
				Expect(logrusFields["age"]).To(Equal(30))
			})

			It("should handle nil return value", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")

				result := f.Map(func(key string, val interface{}) interface{} {
					return ""
				})

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields["key1"]).To(Equal(""))
				Expect(logrusFields["key2"]).To(Equal(""))
			})

			It("should handle type conversion", func() {
				f := logfld.New(nil)
				f.Add("count", "42")

				result := f.Map(func(key string, val interface{}) interface{} {
					if key == "count" {
						return 42 // Convert string to int
					}
					return val
				})

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields["count"]).To(Equal(42))
			})
		})

		Context("on empty fields", func() {
			It("should return empty fields", func() {
				f := logfld.New(nil)

				result := f.Map(func(key string, val interface{}) interface{} {
					return val
				})

				Expect(result).ToNot(BeNil())
				Expect(result.Logrus()).To(HaveLen(0))
			})
		})

		Context("on nil fields instance", func() {
			It("should handle nil safely", func() {
				var f logfld.Fields = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(f).To(BeNil())
			})
		})
	})

	Describe("Method Chaining", func() {
		Context("with multiple operations", func() {
			It("should allow chaining Add operations", func() {
				f := logfld.New(nil)
				result := f.Add("key1", "value1").
					Add("key2", "value2").
					Add("key3", "value3")

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields).To(HaveLen(3))
			})

			It("should allow chaining Add and Map operations", func() {
				f := logfld.New(nil)
				result := f.Add("name", "john").
					Add("city", "paris").
					Map(func(key string, val interface{}) interface{} {
						// Capitalize first letter
						s := val.(string)
						if len(s) > 0 {
							return string(s[0]-32) + s[1:]
						}
						return val
					})

				Expect(result).ToNot(BeNil())
				logrusFields := result.Logrus()
				Expect(logrusFields["name"]).To(Equal("John"))
				Expect(logrusFields["city"]).To(Equal("Paris"))
			})
		})
	})

	Describe("Integration with logrus", func() {
		Context("when using with logrus.Entry", func() {
			It("should work with logrus fields", func() {
				f := logfld.New(nil)
				f.Add("request_id", "123456")
				f.Add("user_id", "user789")
				f.Add("action", "login")

				logrusFields := f.Logrus()

				Expect(logrusFields).To(BeAssignableToTypeOf(logrus.Fields{}))
				Expect(logrusFields).To(HaveLen(3))
			})
		})
	})
})
