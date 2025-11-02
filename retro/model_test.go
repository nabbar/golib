/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package retro_test

import (
	"encoding/json"
	"time"

	. "github.com/nabbar/golib/retro"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Test models
type TestModel struct {
	Version      string    `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
	Name         string    `json:"name" yaml:"name" toml:"name" retro:">v1.0.0"`
	Age          int       `json:"age,omitempty" yaml:"age,omitempty" toml:"age,omitempty"`
	Active       bool      `json:"active,omitempty" yaml:"active,omitempty" toml:"active,omitempty" retro:">=v1.0.0"`
	Legacy       string    `json:"legacy" yaml:"legacy" toml:"legacy" retro:"<v1.0.0"`
	DefaultField string    `json:"default_field" yaml:"default_field" toml:"default_field" retro:"default"`
	Salary       float64   `json:"salary,omitempty" yaml:"salary,omitempty" toml:"salary,omitempty" retro:"<=v1.0.0"`
	Phone        string    `json:"phone" yaml:"phone" toml:"phone" retro:"v1.0.0,v2.0.0"`
	Email        string    `json:"email" yaml:"email" toml:"email" retro:">v1.0.0,<v2.0.0"`
	CreatedAt    time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty" toml:"created_at,omitempty" retro:"default,>=v1.0.0"`
}

type SimpleModel struct {
	ID   int    `json:"id" yaml:"id" toml:"id"`
	Name string `json:"name" yaml:"name" toml:"name"`
}

type StandardModel struct {
	A int      `json:"a" yaml:"a" toml:"a"`
	B string   `json:"b" yaml:"b" toml:"b"`
	C []string `json:"c" yaml:"c" toml:"c"`
}

var _ = Describe("Model", func() {
	Describe("Model structure", func() {
		Context("when creating a new model", func() {
			It("should initialize with default values", func() {
				model := Model[TestModel]{}
				Expect(model.Struct).To(Equal(TestModel{}))
				Expect(model.Standard).To(BeFalse())
			})

			It("should allow setting Standard flag", func() {
				model := Model[TestModel]{Standard: true}
				Expect(model.Standard).To(BeTrue())
			})

			It("should hold the struct data", func() {
				data := TestModel{Name: "John", Age: 30}
				model := Model[TestModel]{Struct: data}
				Expect(model.Struct.Name).To(Equal("John"))
				Expect(model.Struct.Age).To(Equal(30))
			})
		})
	})

	Describe("JSON marshaling", func() {
		Context("when marshaling with no version (default)", func() {
			It("should include only fields with default tag and no retro tag", func() {
				model := Model[TestModel]{
					Struct: TestModel{
						Age:          25,
						DefaultField: "test",
						Name:         "John",
					},
				}

				data, err := json.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(HaveKey("age"))
				Expect(result).To(HaveKey("default_field"))
				Expect(result).ToNot(HaveKey("name")) // requires >v1.0.0
			})
		})

		Context("when marshaling with version v1.0.0", func() {
			It("should include fields matching version constraints", func() {
				model := Model[TestModel]{
					Struct: TestModel{
						Version: "v1.0.0",
						Age:     25,
						Active:  true,
						Salary:  50000,
						Phone:   "123-456",
						Name:    "test",
					},
				}

				data, err := json.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(HaveKey("version"))
				Expect(result).To(HaveKey("age"))
				Expect(result).To(HaveKey("active"))   // >=v1.0.0
				Expect(result).To(HaveKey("salary"))   // <=v1.0.0
				Expect(result).To(HaveKey("phone"))    // v1.0.0,v2.0.0
				Expect(result).ToNot(HaveKey("name"))  // >v1.0.0
				Expect(result).ToNot(HaveKey("email")) // >v1.0.0,<v2.0.0
			})
		})

		Context("when marshaling with version v1.5.0", func() {
			It("should include fields in range", func() {
				model := Model[TestModel]{
					Struct: TestModel{
						Version: "v1.5.0",
						Name:    "John",
						Age:     30,
						Active:  true,
						Email:   "john@example.com",
					},
				}

				data, err := json.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(HaveKey("name"))  // >v1.0.0
				Expect(result).To(HaveKey("email")) // >v1.0.0,<v2.0.0
				Expect(result).To(HaveKey("active"))
				Expect(result).ToNot(HaveKey("salary")) // <=v1.0.0
				Expect(result).ToNot(HaveKey("phone"))  // v1.0.0,v2.0.0 only
			})
		})

		Context("when marshaling with omitempty", func() {
			It("should omit zero values with omitempty tag", func() {
				model := Model[TestModel]{
					Struct: TestModel{
						Version: "v1.0.0",
						Active:  false,
						Salary:  0,
					},
				}

				data, err := json.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).ToNot(HaveKey("active")) // omitempty + false
				Expect(result).ToNot(HaveKey("salary")) // omitempty + 0
			})

			It("should include non-zero values with omitempty tag", func() {
				model := Model[TestModel]{
					Struct: TestModel{
						Version: "v1.0.0",
						Age:     25,
						Active:  true,
						Salary:  50000,
					},
				}

				data, err := json.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(HaveKey("age"))
				Expect(result).To(HaveKey("active"))
				Expect(result).To(HaveKey("salary"))
			})
		})

		Context("when using Standard mode", func() {
			It("should use standard JSON marshaling", func() {
				model := Model[StandardModel]{
					Struct: StandardModel{
						A: 42,
						B: "test",
						C: []string{"a", "b"},
					},
					Standard: true,
				}

				data, err := json.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result["a"]).To(BeNumerically("==", 42))
				Expect(result["b"]).To(Equal("test"))
			})
		})
	})

	Describe("JSON unmarshaling", func() {
		Context("when unmarshaling with no version", func() {
			It("should populate fields correctly", func() {
				jsonData := `{"age":25,"default_field":"test"}`
				var model Model[TestModel]

				err := json.Unmarshal([]byte(jsonData), &model)
				Expect(err).ToNot(HaveOccurred())

				Expect(model.Struct.Age).To(Equal(25))
				Expect(model.Struct.DefaultField).To(Equal("test"))
			})
		})

		Context("when unmarshaling with version", func() {
			It("should respect version constraints", func() {
				jsonData := `{"version":"v1.5.0","name":"John","age":30,"email":"john@example.com"}`
				var model Model[TestModel]

				err := json.Unmarshal([]byte(jsonData), &model)
				Expect(err).ToNot(HaveOccurred())

				Expect(model.Struct.Version).To(Equal("v1.5.0"))
				Expect(model.Struct.Name).To(Equal("John"))
				Expect(model.Struct.Age).To(Equal(30))
				Expect(model.Struct.Email).To(Equal("john@example.com"))
			})
		})

		Context("when using Standard mode", func() {
			It("should use standard JSON unmarshaling", func() {
				jsonData := `{"a":42,"b":"test","c":["x","y"]}`
				model := Model[StandardModel]{Standard: true}

				err := json.Unmarshal([]byte(jsonData), &model)
				Expect(err).ToNot(HaveOccurred())

				Expect(model.Struct.A).To(Equal(42))
				Expect(model.Struct.B).To(Equal("test"))
				Expect(model.Struct.C).To(Equal([]string{"x", "y"}))
			})
		})

		Context("when unmarshaling invalid JSON", func() {
			It("should return error for malformed JSON", func() {
				jsonData := `{"name":"John",invalid}`
				var model Model[TestModel]

				err := json.Unmarshal([]byte(jsonData), &model)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("YAML marshaling", func() {
		Context("when marshaling with version", func() {
			It("should include correct fields", func() {
				model := Model[TestModel]{
					Struct: TestModel{
						Version: "v1.0.0",
						Age:     25,
						Active:  true,
						Phone:   "123-456",
					},
				}

				data, err := yaml.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = yaml.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(HaveKey("version"))
				Expect(result).To(HaveKey("age"))
				Expect(result).To(HaveKey("active"))
				Expect(result).To(HaveKey("phone"))
			})
		})

		Context("when using Standard mode", func() {
			It("should use standard YAML marshaling", func() {
				model := Model[StandardModel]{
					Struct: StandardModel{
						A: 42,
						B: "test",
					},
					Standard: true,
				}

				data, err := yaml.Marshal(model)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(ContainSubstring("a: 42"))
				Expect(string(data)).To(ContainSubstring("b: test"))
			})
		})
	})

	Describe("YAML unmarshaling", func() {
		Context("when unmarshaling with version", func() {
			It("should populate fields correctly", func() {
				yamlData := `
version: v1.5.0
name: John
age: 30
email: john@example.com
`
				var model Model[TestModel]

				err := yaml.Unmarshal([]byte(yamlData), &model)
				Expect(err).ToNot(HaveOccurred())

				Expect(model.Struct.Version).To(Equal("v1.5.0"))
				Expect(model.Struct.Name).To(Equal("John"))
				Expect(model.Struct.Age).To(Equal(30))
				Expect(model.Struct.Email).To(Equal("john@example.com"))
			})
		})

		Context("when using Standard mode", func() {
			It("should use standard YAML unmarshaling", func() {
				yamlData := `
a: 42
b: test
c:
  - x
  - y
`
				model := Model[StandardModel]{Standard: true}

				err := yaml.Unmarshal([]byte(yamlData), &model)
				Expect(err).ToNot(HaveOccurred())

				Expect(model.Struct.A).To(Equal(42))
				Expect(model.Struct.B).To(Equal("test"))
				Expect(model.Struct.C).To(Equal([]string{"x", "y"}))
			})
		})
	})

	Describe("TOML marshaling", func() {
		Context("when marshaling with version", func() {
			It("should include correct fields", func() {
				model := Model[TestModel]{
					Struct: TestModel{
						Version: "v1.0.0",
						Age:     25,
						Active:  true,
						Phone:   "123-456",
					},
				}

				data, err := toml.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result map[string]interface{}
				err = toml.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(HaveKey("version"))
				Expect(result).To(HaveKey("age"))
				Expect(result).To(HaveKey("active"))
			})
		})

		Context("when using Standard mode", func() {
			It("should use standard TOML marshaling", func() {
				model := Model[StandardModel]{
					Struct: StandardModel{
						A: 42,
						B: "test",
						C: []string{"x", "y"},
					},
					Standard: true,
				}

				data, err := toml.Marshal(model)
				Expect(err).ToNot(HaveOccurred())

				var result StandardModel
				err = toml.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.A).To(Equal(42))
				Expect(result.B).To(Equal("test"))
			})
		})
	})

	Describe("TOML unmarshaling", func() {
		Context("when unmarshaling with version", func() {
			It("should populate fields correctly", func() {
				tomlData := `
version = "v1.5.0"
name = "John"
age = 30
email = "john@example.com"
`
				var model Model[TestModel]

				err := toml.Unmarshal([]byte(tomlData), &model)
				Expect(err).ToNot(HaveOccurred())

				Expect(model.Struct.Version).To(Equal("v1.5.0"))
				Expect(model.Struct.Name).To(Equal("John"))
				Expect(model.Struct.Age).To(Equal(30))
				Expect(model.Struct.Email).To(Equal("john@example.com"))
			})
		})

		Context("when using Standard mode", func() {
			It("should use standard TOML unmarshaling", func() {
				tomlData := `
a = 42
b = "test"
c = ["x", "y"]
`
				model := Model[StandardModel]{Standard: true}

				err := toml.Unmarshal([]byte(tomlData), &model)
				Expect(err).ToNot(HaveOccurred())

				Expect(model.Struct.A).To(Equal(42))
				Expect(model.Struct.B).To(Equal("test"))
				Expect(model.Struct.C).To(ConsistOf("x", "y"))
			})
		})
	})

	Describe("Round-trip consistency", func() {
		Context("when marshaling and unmarshaling JSON", func() {
			It("should maintain data integrity", func() {
				original := Model[SimpleModel]{
					Struct: SimpleModel{
						ID:   42,
						Name: "test",
					},
				}

				data, err := json.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var result Model[SimpleModel]
				err = json.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.Struct.ID).To(Equal(original.Struct.ID))
				Expect(result.Struct.Name).To(Equal(original.Struct.Name))
			})
		})

		Context("when marshaling and unmarshaling YAML", func() {
			It("should maintain data integrity", func() {
				original := Model[SimpleModel]{
					Struct: SimpleModel{
						ID:   42,
						Name: "test",
					},
				}

				data, err := yaml.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var result Model[SimpleModel]
				err = yaml.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.Struct.ID).To(Equal(original.Struct.ID))
				Expect(result.Struct.Name).To(Equal(original.Struct.Name))
			})
		})

		Context("when marshaling and unmarshaling TOML", func() {
			It("should maintain data integrity", func() {
				original := Model[SimpleModel]{
					Struct: SimpleModel{
						ID:   42,
						Name: "test",
					},
				}

				data, err := toml.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var result Model[SimpleModel]
				err = toml.Unmarshal(data, &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.Struct.ID).To(Equal(original.Struct.ID))
				Expect(result.Struct.Name).To(Equal(original.Struct.Name))
			})
		})
	})
})
