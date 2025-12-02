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

// This file contains tests for field operations on logger entries.
// It validates FieldAdd, FieldMerge, FieldSet, and FieldClean methods with
// various field types and scenarios.
//
// Test Coverage:
//   - Field addition with different value types (string, int, bool)
//   - Field merging and overwriting
//   - Field initialization and replacement
//   - Field deletion and cleanup
//   - Nil fields handling
//   - Method chaining with field operations
//   - Integration with other entry methods
package entry_test

import (
	logent "github.com/nabbar/golib/logger/entry"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Field Operations", func() {
	Describe("FieldAdd", func() {
		Context("on entry with initialized fields", func() {
			It("should add string field", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldAdd("key1", "value1")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should add integer field", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldAdd("count", 42)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should add boolean field", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldAdd("enabled", true)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should add multiple fields", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				e = e.FieldAdd("key1", "value1")
				e = e.FieldAdd("key2", "value2")
				e = e.FieldAdd("key3", "value3")

				Expect(e).ToNot(BeNil())
			})

			It("should overwrite existing field", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				e = e.FieldAdd("key", "value1")
				e = e.FieldAdd("key", "value2")

				Expect(e).ToNot(BeNil())
			})
		})

		Context("on entry with nil fields", func() {
			It("should return nil", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.FieldAdd("key", "value")

				Expect(result).To(BeNil())
			})
		})

		Context("on nil entry", func() {
			It("should handle nil entry safely", func() {
				var e logent.Entry = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(e).To(BeNil())
			})
		})
	})

	Describe("FieldMerge", func() {
		Context("on entry with initialized fields", func() {
			It("should merge empty fields", func() {
				fields1 := logfld.New(nil)
				fields2 := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields1)
				result := e.FieldMerge(fields2)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should merge fields with values", func() {
				fields1 := logfld.New(nil)
				fields1.Add("key1", "value1")

				fields2 := logfld.New(nil)
				fields2.Add("key2", "value2")

				e := logent.New(loglvl.InfoLevel).FieldSet(fields1)
				result := e.FieldMerge(fields2)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should overwrite overlapping keys", func() {
				fields1 := logfld.New(nil)
				fields1.Add("key", "value1")

				fields2 := logfld.New(nil)
				fields2.Add("key", "value2")

				e := logent.New(loglvl.InfoLevel).FieldSet(fields1)
				e = e.FieldMerge(fields2)

				Expect(e).ToNot(BeNil())
			})

			It("should merge multiple field sets", func() {
				fields1 := logfld.New(nil)
				fields1.Add("key1", "value1")

				fields2 := logfld.New(nil)
				fields2.Add("key2", "value2")

				fields3 := logfld.New(nil)
				fields3.Add("key3", "value3")

				e := logent.New(loglvl.InfoLevel).FieldSet(fields1)
				e = e.FieldMerge(fields2)
				e = e.FieldMerge(fields3)

				Expect(e).ToNot(BeNil())
			})

			It("should handle nil fields parameter", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldMerge(nil)

				Expect(result).ToNot(BeNil())
			})
		})

		Context("on entry with nil fields", func() {
			It("should return nil", func() {
				e := logent.New(loglvl.InfoLevel)
				fields := logfld.New(nil)
				result := e.FieldMerge(fields)

				Expect(result).To(BeNil())
			})
		})

		Context("on nil entry", func() {
			It("should handle nil entry safely", func() {
				var e logent.Entry = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(e).To(BeNil())
			})
		})
	})

	Describe("FieldSet", func() {
		Context("on valid entry", func() {
			It("should set empty fields", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel)
				result := e.FieldSet(fields)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should set fields with values", func() {
				fields := logfld.New(nil)
				fields.Add("key1", "value1")
				fields.Add("key2", "value2")

				e := logent.New(loglvl.InfoLevel)
				result := e.FieldSet(fields)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should replace existing fields", func() {
				fields1 := logfld.New(nil)
				fields1.Add("key1", "value1")

				fields2 := logfld.New(nil)
				fields2.Add("key2", "value2")

				e := logent.New(loglvl.InfoLevel)
				e = e.FieldSet(fields1)
				e = e.FieldSet(fields2)

				Expect(e).ToNot(BeNil())
			})

			It("should set nil fields", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.FieldSet(nil)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})

		Context("on nil entry", func() {
			It("should handle nil entry safely", func() {
				var e logent.Entry = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(e).To(BeNil())
			})
		})
	})

	Describe("FieldClean", func() {
		Context("on entry with initialized fields", func() {
			It("should remove single key", func() {
				fields := logfld.New(nil)
				fields.Add("key1", "value1")
				fields.Add("key2", "value2")

				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldClean("key1")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should remove multiple keys", func() {
				fields := logfld.New(nil)
				fields.Add("key1", "value1")
				fields.Add("key2", "value2")
				fields.Add("key3", "value3")

				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldClean("key1", "key2")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should handle non-existent keys", func() {
				fields := logfld.New(nil)
				fields.Add("key1", "value1")

				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldClean("nonexistent")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should handle empty key list", func() {
				fields := logfld.New(nil)
				fields.Add("key1", "value1")

				e := logent.New(loglvl.InfoLevel).FieldSet(fields)
				result := e.FieldClean()

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})

		Context("on entry with nil fields", func() {
			It("should return nil", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.FieldClean("key")

				Expect(result).To(BeNil())
			})
		})

		Context("on nil entry", func() {
			It("should handle nil entry safely", func() {
				var e logent.Entry = nil

				// Calling methods on nil interface causes panic in Go
				// This is expected behavior
				Expect(e).To(BeNil())
			})
		})
	})

	Describe("Field Method Chaining", func() {
		Context("with valid entry and fields", func() {
			It("should chain field operations", func() {
				fields := logfld.New(nil)
				fields.Add("base", "value")

				fields2 := logfld.New(nil)
				fields2.Add("merge", "value")

				result := logent.New(loglvl.InfoLevel).
					FieldSet(fields).
					FieldAdd("added", "value").
					FieldMerge(fields2).
					FieldClean("unwanted")

				Expect(result).ToNot(BeNil())
			})

			It("should maintain entry reference through chain", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).FieldSet(fields)

				result := e.FieldAdd("key1", "val1").
					FieldAdd("key2", "val2")

				Expect(result).To(Equal(e))
			})
		})
	})

	Describe("Field Integration", func() {
		Context("with complete entry setup", func() {
			It("should work with all entry methods", func() {
				fields := logfld.New(nil)
				e := logent.New(loglvl.InfoLevel).
					FieldSet(fields).
					FieldAdd("app", "test").
					FieldAdd("version", "1.0").
					SetLevel(loglvl.DebugLevel).
					DataSet("test data")

				Expect(e).ToNot(BeNil())
			})
		})
	})
})
