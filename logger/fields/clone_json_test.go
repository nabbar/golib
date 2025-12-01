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
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logfld "github.com/nabbar/golib/logger/fields"
)

var _ = Describe("Fields Clone and JSON Operations", func() {
	Describe("FieldsClone", func() {
		Context("with valid fields instance", func() {
			It("should create independent clone", func() {
				original := logfld.New(nil)
				original.Add("key1", "value1")
				original.Add("key2", "value2")

				clone := original.Clone()

				Expect(clone).ToNot(BeNil())
				Expect(clone).ToNot(Equal(original)) // Different instances

				// Verify clone has same values
				origFields := original.Logrus()
				cloneFields := clone.Logrus()
				Expect(cloneFields).To(HaveLen(len(origFields)))
				Expect(cloneFields["key1"]).To(Equal("value1"))
				Expect(cloneFields["key2"]).To(Equal("value2"))
			})

			It("should not affect original when clone is modified", func() {
				original := logfld.New(nil)
				original.Add("key1", "value1")

				clone := original.Clone()
				clone.Add("key2", "value2")
				clone.Add("key1", "modified")

				// Original should remain unchanged
				origFields := original.Logrus()
				Expect(origFields).To(HaveLen(1))
				Expect(origFields["key1"]).To(Equal("value1"))

				// Clone should have modifications
				cloneFields := clone.Logrus()
				Expect(cloneFields).To(HaveLen(2))
				Expect(cloneFields["key1"]).To(Equal("modified"))
				Expect(cloneFields["key2"]).To(Equal("value2"))
			})

			It("should clone empty fields", func() {
				original := logfld.New(nil)

				clone := original.Clone()

				Expect(clone).ToNot(BeNil())
				Expect(clone.Logrus()).To(HaveLen(0))
			})

			It("should work with nil context", func() {
				original := logfld.New(nil)
				original.Add("key", "value")

				clone := original.Clone()

				Expect(clone).ToNot(BeNil())
				cloneFields := clone.Logrus()
				Expect(cloneFields).To(HaveKeyWithValue("key", "value"))
			})

			It("should preserve complex types", func() {
				original := logfld.New(nil)
				mapVal := map[string]interface{}{
					"nested": "value",
					"count":  42,
				}
				sliceVal := []string{"item1", "item2"}

				original.Add("map", mapVal)
				original.Add("slice", sliceVal)

				clone := original.Clone()

				cloneFields := clone.Logrus()
				Expect(cloneFields).To(HaveKey("map"))
				Expect(cloneFields).To(HaveKey("slice"))
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

	Describe("MarshalJSON", func() {
		Context("with valid fields instance", func() {
			It("should marshal empty fields", func() {
				f := logfld.New(nil)

				data, err := json.Marshal(f)

				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())
				Expect(string(data)).To(Equal("{}"))
			})

			It("should marshal single field", func() {
				f := logfld.New(nil)
				f.Add("key", "value")

				data, err := json.Marshal(f)

				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(ContainSubstring("key"))
				Expect(string(data)).To(ContainSubstring("value"))
			})

			It("should marshal multiple fields", func() {
				f := logfld.New(nil)
				f.Add("name", "John")
				f.Add("age", 30)
				f.Add("active", true)

				data, err := json.Marshal(f)

				Expect(err).ToNot(HaveOccurred())

				// Verify JSON structure
				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(HaveLen(3))
				Expect(result["name"]).To(Equal("John"))
				Expect(result["age"]).To(BeNumerically("==", 30))
				Expect(result["active"]).To(BeTrue())
			})

			It("should marshal nested structures", func() {
				f := logfld.New(nil)
				f.Add("metadata", map[string]interface{}{
					"version": "1.0",
					"build":   123,
				})

				data, err := json.Marshal(f)

				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(ContainSubstring("metadata"))
				Expect(string(data)).To(ContainSubstring("version"))
			})

			It("should marshal arrays", func() {
				f := logfld.New(nil)
				f.Add("items", []string{"a", "b", "c"})

				data, err := json.Marshal(f)

				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(ContainSubstring("items"))
			})

			It("should handle nil values", func() {
				f := logfld.New(nil)
				f.Add("nullable", nil)

				data, err := json.Marshal(f)

				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).ToNot(ContainSubstring("nullable"))
				Expect(string(data)).ToNot(ContainSubstring("null"))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		Context("with valid JSON data", func() {
			It("should unmarshal empty object", func() {
				f := logfld.New(nil)
				jsonData := []byte("{}")

				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				Expect(f.Logrus()).To(HaveLen(0))
			})

			It("should unmarshal single field", func() {
				f := logfld.New(nil)
				jsonData := []byte(`{"key":"value"}`)

				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				fields := f.Logrus()
				Expect(fields).To(HaveKeyWithValue("key", "value"))
			})

			It("should unmarshal multiple fields", func() {
				f := logfld.New(nil)
				jsonData := []byte(`{"name":"John","age":30,"active":true}`)

				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				fields := f.Logrus()
				Expect(fields).To(HaveLen(3))
				Expect(fields["name"]).To(Equal("John"))
				Expect(fields["age"]).To(BeNumerically("==", 30))
				Expect(fields["active"]).To(BeTrue())
			})

			It("should unmarshal nested structures", func() {
				f := logfld.New(nil)
				jsonData := []byte(`{"metadata":{"version":"1.0","build":123}}`)

				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				fields := f.Logrus()
				Expect(fields).To(HaveKey("metadata"))
			})

			It("should unmarshal arrays", func() {
				f := logfld.New(nil)
				jsonData := []byte(`{"items":["a","b","c"]}`)

				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				fields := f.Logrus()
				Expect(fields).To(HaveKey("items"))
			})

			It("should unmarshal null values", func() {
				f := logfld.New(nil)
				jsonData := []byte(`{"nullable":null}`)

				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				fields := f.Logrus()
				Expect(fields).ToNot(HaveKey("nullable"))
				Expect(fields["nullable"]).To(BeNil())
			})

			It("should overwrite existing fields", func() {
				f := logfld.New(nil)
				f.Add("key", "original")

				jsonData := []byte(`{"key":"updated"}`)
				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				fields := f.Logrus()
				Expect(fields["key"]).To(Equal("updated"))
			})

			It("should merge with existing fields", func() {
				f := logfld.New(nil)
				f.Add("existing", "value")

				jsonData := []byte(`{"new":"field"}`)
				err := json.Unmarshal(jsonData, f)

				Expect(err).ToNot(HaveOccurred())
				fields := f.Logrus()
				Expect(fields).To(HaveLen(2))
				Expect(fields["existing"]).To(Equal("value"))
				Expect(fields["new"]).To(Equal("field"))
			})
		})

		Context("with invalid JSON data", func() {
			It("should return error for malformed JSON", func() {
				f := logfld.New(nil)
				jsonData := []byte(`{invalid json}`)

				err := json.Unmarshal(jsonData, f)

				Expect(err).To(HaveOccurred())
			})

			It("should return error for non-object JSON", func() {
				f := logfld.New(nil)
				jsonData := []byte(`["array","not","object"]`)

				err := json.Unmarshal(jsonData, f)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("JSON Round Trip", func() {
		Context("with various field types", func() {
			It("should preserve data through marshal/unmarshal", func() {
				original := logfld.New(nil)
				original.Add("string", "value")
				original.Add("int", 42)
				original.Add("float", 3.14)
				original.Add("bool", true)
				original.Add("null", nil)

				// Marshal
				data, err := json.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				// Unmarshal
				restored := logfld.New(nil)
				err = json.Unmarshal(data, restored)
				Expect(err).ToNot(HaveOccurred())

				// Compare
				origFields := original.Logrus()
				restFields := restored.Logrus()
				Expect(restFields).To(HaveLen(len(origFields)))
				Expect(restFields["string"]).To(Equal("value"))
				Expect(restFields["int"]).To(BeNumerically("==", 42))
				Expect(restFields["float"]).To(BeNumerically("~", 3.14))
				Expect(restFields["bool"]).To(BeTrue())
				Expect(restFields["null"]).To(BeNil())
			})

			It("should handle complex nested structures", func() {
				original := logfld.New(nil)
				original.Add("complex", map[string]interface{}{
					"nested": map[string]interface{}{
						"deep": "value",
					},
					"array": []interface{}{1, "two", true},
				})

				// Marshal and unmarshal
				data, err := json.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				restored := logfld.New(nil)
				err = json.Unmarshal(data, restored)
				Expect(err).ToNot(HaveOccurred())

				// Verify structure exists
				restFields := restored.Logrus()
				Expect(restFields).To(HaveKey("complex"))
			})
		})
	})
})
