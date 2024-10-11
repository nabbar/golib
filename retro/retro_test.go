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
	"reflect"
	"testing"
	"time"

	"github.com/nabbar/golib/retro"
	"github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v3"
)

type Test struct {
	LastName     string     `json:"lastName" yaml:"lastName" toml:"lastName" retro:"v1.0.0,>v1.0.3"` // greater than v1.0.3 except v1.0.0
	Age          int        `json:"age,omitempty" yaml:"age,omitempty" toml:"age,omitempty"`         // this if not 0 should be included in all the models because no retro tag
	Name         string     `json:"name"  yaml:"name" toml:"name" retro:">v1.0.0"`                   // only if version strict greater than v1.0.0
	Version      string     `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
	Salary       float64    `json:"salary,omitempty" yaml:"salary,omitempty" toml:"salary,omitempty" retro:"<v1.0.2"`                                     // only if version strict lesser than v1.0.2
	Active       bool       `json:"active,omitempty" yaml:"active,omitempty" toml:"active,omitempty" retro:">=v1.0.2"`                                    // only if version greater or equal to v1.0.2
	Address      Address    `json:"address" yaml:"address" toml:"address" retro:"<=v1.0.0"`                                                               // only if version lesser or equal to v1.0.0
	Job          string     `json:"job" yaml:"job" toml:"job" retro:">v1.0.1,<v1.0.4"`                                                                    // only v1.0.2 and v1.0.3
	Status       Status     `json:"status" yaml:"status" toml:"status" retro:"default,>v1.0.1,<v1.0.4"`                                                   // default (meaning no versioning) or only v1.0.2 and v1.0.3
	Married      bool       `json:"married" yaml:"married" toml:"married" retro:">=v1.0.0,<v1.0.2"`                                                       // v1.0.0 v1.0.1
	BirthDate    *time.Time `json:"birthdate,omitempty" yaml:"birthdate,omitempty" toml:"birthdate,omitempty" retro:"default,>=v1.0.0,<v1.0.2"`           // v1.0.0 v1.0.1 and default
	Degree       string     `json:"degree" yaml:"degree" toml:"degree" retro:">v1.0.0,<=v1.0.2"`                                                          // v1.0.1 v1.0.2
	Phone        int32      `json:"phone,omitempty" yaml:"phone,omitempty" toml:"phone,omitempty" retro:"default,>v1.0.0,<=v1.0.2"`                       // v1.0.1 v1.0.2 and default
	Other        []string   `json:"other" yaml:"other" toml:"other" retro:">=v1.0.0,<=v1.0.2"`                                                            // v1.0.0 v1.0.1 v1.0.2
	LuckyNumbers []int      `json:"luckyNumbers,omitempty" yaml:"luckyNumbers,omitempty" toml:"luckyNumbers,omitempty" retro:"default,>=v1.0.0,<=v1.0.2"` // v1.0.0 v1.0.1 v1.0.2 and default
	Weight       int        `json:"weight,omitempty" yaml:"weight,omitempty" toml:"weight,omitempty" retro:"default,v1.0.0"`                              // v1.0.0 and default
	Height       int        `json:"height" yaml:"height" toml:"height" retro:"v1.0.0"`                                                                    // v1.0.0 only
	Id           string     `json:"id" yaml:"id"  toml:"id" retro:"v1.0.0,v1.0.3"`                                                                        // v1.0.0 and v1.0.3 only
	Languages    []string   `json:"languages,omitempty" yaml:"languages,omitempty" toml:"languages,omitempty" retro:"default,v1.0.0,v1.0.3"`              // v1.0.0 and v1.0.3 and default
	Email        string     `json:"email" yaml:"email" toml:"email" retro:"<v1.0.0,v1.0.3"`                                                               // lesser than v1.0.0 expect v1.0.3
	Available    bool       `json:"available" yaml:"available" toml:"available" retro:">v1.0.0,<=v1.0.3"`                                                 // v1.0.1 v1.0.2 v1.0.3
	Sex          string     `json:"sex,omitempty" yaml:"sex,omitempty" toml:"sex,omitempty" retro:">v1.0.0,<=v1.0.3, v0.0.3, default"`                    // between v1.0.0 and v1.0.3 (expect v0.0.3) and default
	Conflict     string     `json:"conflict" yaml:"conflict" toml:"conflict" retro:">v1.0.0,>v1.0.3"`                                                     // this field has non-valid retro definition and should be always ignored
}

type Standard struct {
	A int      `json:"a" yaml:"a" toml:"a"`
	b int      
	C string   `json:"C" yaml:"C" toml:"C"`
	D []string `json:"d" yaml:"d" toml:"d"`
}

type Address struct {
	Street  string `json:"street" `
	City    string `json:"city,omitempty"`
}

type Status int

const (
	Inactive Status = iota
	Active
	Pending
)

func (s Status) MarshalJSON() ([]byte, error) {
	switch s {
	case Inactive:
		return json.Marshal("inactive")
	case Active:
		return json.Marshal("active")
	case Pending:
		return json.Marshal("pending")
	default:
		return json.Marshal("unknown")
	}
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var statusStr string

	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}

	switch statusStr {
	case "inactive":
		*s = Inactive
	case "active":
		*s = Active
	case "pending":
		*s = Pending
	default:
		*s = Inactive
	}
	return nil
}

func TestModel_MarshalJSON(t *testing.T) {
	var birth = time.Date(2024, 10, 8, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		model       retro.Model[Test]
		expected    string
		expectedErr bool
	}{
		{ // Test Default No Versioning + age field with no retro tag
			model: retro.Model[Test]{Struct: Test{
				Age:    25,
				Name:   "Alice",
				Active: true,
				Address: Address{Street: "123 Main St",
					City: "Wonderland"},
				Status: Active},
			},

			expected: `{"age":25,"status":"active"}`,
		},
		{ // v1.0.3
			model: retro.Model[Test]{Struct: Test{
				Age:       25,
				Name:      "Alice",
				Active:    true,
				Job:       "test",
				Id:        "uc123",
				Email:     "test@example.com",
				Available: true,
				Sex:       "M",
				Version:   "v1.0.3",
				Address:   Address{Street: "123 Main St", City: "Wonderland"},
				Status:    Active}},

			expected: `{"version":"v1.0.3","age":25,"status":"active","name":"Alice", 
						"active":true,"job":"test","id":"uc123",
						"email":"test@example.com","available":true,"sex":"M"}`,
		},
		{ // v0.0.3

			model: retro.Model[Test]{Struct: Test{
				Age:     15,
				Version: "v0.0.3",
				Salary:  100,
				Address: Address{Street: "123 Main St", City: "Wonderland"},
				Email:   "test",
				Sex:     "F"}},

			expected: `{"version":"v0.0.3","age":15,"email":"test","sex":"F",
						"salary":100,"address":{"street":"123 Main St","city":"Wonderland"}}`,
		},
		{
			model: retro.Model[Test]{Struct: Test{
				LastName: "test",
				Age:      34,
				Name:     "test",
				Version:  "v1.0.0",
				Salary:   1500,
				Active:   true,
				Address: Address{
					Street: "Joseph Bermond",
					City:   "Valbonne",
				},
				Job:          "test",
				Status:       Active,
				Married:      true,
				BirthDate:    &birth,
				Degree:       "test",
				Phone:        12345850,
				Other:        []string{"tt", "aa", "bb"},
				LuckyNumbers: []int{12, 21},
				Weight:       100,
				Height:       190,
				Id:           "uc123",
				Languages:    []string{"french"},
				Email:        "test@test.com",
				Available:    true,
				Sex:          "M",
				Conflict:     "test",
			}},
			expected: `{"lastName": "test","salary":1500,"age": 34,"version": "v1.0.0",
						"address": {"street": "Joseph Bermond","city": "Valbonne"},"married": true,
						"birthdate": "2024-10-08T00:00:00Z","other": ["tt", "aa", "bb"],"luckyNumbers": 
						[12, 21],"weight": 100,"height": 190,"id": "uc123","languages": ["french"]}`,
		},
	}

	for _, tt := range tests {
		var expectedJSON, resultJSON interface{}

		t.Run(tt.expected, func(t *testing.T) {

			result, err := json.Marshal(tt.model)

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if err = json.Unmarshal([]byte(tt.expected), &expectedJSON); err != nil {
				t.Fatalf("failed to unmarshal expected JSON: %v", err)
			}

			if err = json.Unmarshal(result, &resultJSON); err != nil {
				t.Fatalf("failed to unmarshal result JSON: %v", err)
			}

			if !reflect.DeepEqual(expectedJSON, resultJSON) {
				t.Errorf("expected: %+v, got: %+v", expectedJSON, resultJSON)
			}
		})
	}
}

func TestModel_UnmarshalJSON(t *testing.T) {
	var birth = time.Date(2024, 10, 8, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		input       string
		expected    retro.Model[Test]
		expectedErr bool
	}{
		{
			input:    `{"age":25,"status":"active","weight":70}`,
			expected: retro.Model[Test]{Struct: Test{Age: 25, Status: Active, Weight: 70}},
		},
		{
			input: `{"version":"v1.0.3","age":25,"status":"active","name":"Alice", 
					"active":true,"job":"test","id":"uc123","email":"test@example.com","available":true,"sex":"M"}`,
			expected: retro.Model[Test]{Struct: Test{Age: 25, Name: "Alice", Active: true,
				Job: "test", Id: "uc123", Email: "test@example.com",
				Available: true, Sex: "M", Version: "v1.0.3", Status: Active}},
		},
		{
			input: `{"lastName":"test","age":34,"name":"test","version":"v1.0.0",
					 "salary":1500,"active":true,"address":{"street":"Joseph Bermond","city":"Valbonne"},"job":"test",
					"status":"active","married":true,"birthdate":"2024-10-08T00:00:00Z","degree":"test","phone":12345850,
					"other":["tt","aa","bb"],"luckyNumbers":[12,21],"weight":100,"height":190,
					"id":"uc123","languages":["french"],
					"email":"test@test.com","available":true,"sex":"M","conflict":"test"}`,
			expected: retro.Model[Test]{
				Struct: Test{
					LastName: "test",
					Age:      34,
					Version:  "v1.0.0",
					Salary:   1500,
					Address: Address{
						Street: "Joseph Bermond",
						City:   "Valbonne",
					},
					Married:      true,
					BirthDate:    &birth,
					Other:        []string{"tt", "aa", "bb"},
					LuckyNumbers: []int{12, 21},
					Weight:       100,
					Height:       190,
					Id:           "uc123",
					Languages:    []string{"french"},
				},
			},
		},
	}

	for _, tt := range tests {
		var result retro.Model[Test]

		t.Run(tt.input, func(t *testing.T) {

			err := json.Unmarshal([]byte(tt.input), &result)

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestModel_MarshalYAML(t *testing.T) {
	tests := []struct {
		model       retro.Model[Test]
		expected    string
		expectedErr bool
	}{
		{ // Example Test Case
			model: retro.Model[Test]{Struct: Test{Age: 25, Status: Active}},
			expected: `age: 25
status: 1
`,
			expectedErr: false,
		},
		{ // v0.0.3
			model: retro.Model[Test]{Struct: Test{
				Age:     15,
				Version: "v0.0.3",
				Salary:  100,
				Address: Address{Street: "123 Main St", City: "Wonderland"},
				Email:   "test",
				Sex:     "F",
			}},
			expected: `version: v0.0.3
age: 15
email: test
sex: F
salary: 100
address:
  street: 123 Main St
  city: Wonderland
`,
			expectedErr: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.expected, func(t *testing.T) {
			var (
				m   retro.Model[Test]
				err error
			)

			if _, err = yaml.Marshal(&tt.model); (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if err = yaml.Unmarshal([]byte(tt.expected), &m); err != nil {
				t.Fatalf("failed to unmarshal expected YAML: %v", err)
			}

			if !reflect.DeepEqual(m, tt.model) {

				t.Errorf("expected: %+v, got: %+v", tt.model, m)
			}

		})
	}
}

func TestModel_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		input       string
		expected    retro.Model[Test]
		expectedErr bool
	}{
		{
			input: `age: 25
status: 1
weight: 70`,
			expected: retro.Model[Test]{Struct: Test{Age: 25, Status: Active, Weight: 70}},
		},
		{
			input: `version: v1.0.3
age: 25
status: 1
name: Alice
active: true
job: test
id: uc123
email: test@example.com
available: true
sex: M`,
			expected: retro.Model[Test]{Struct: Test{Age: 25, Name: "Alice", Active: true, Job: "test", Id: "uc123",
				Email:     "test@example.com",
				Available: true, Sex: "M", Version: "v1.0.3", Status: Active}},
		},
	}

	for _, tt := range tests {

		t.Run(tt.input, func(t *testing.T) {

			var result retro.Model[Test]

			err := yaml.Unmarshal([]byte(tt.input), &result)

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if !reflect.DeepEqual(result, tt.expected) {

				t.Errorf("expected: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestModel_MarshalTOML(t *testing.T) {
	var birth = time.Date(2024, 10, 8, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		model       retro.Model[Test]
		expected    string
		expectedErr bool
	}{
		{
			model: retro.Model[Test]{Struct: Test{
				LastName: "test",
				Age:      34,
				Name:     "test",
				Salary:   1500,
				Active:   true,
				Address: Address{
					Street: "Joseph Bermond",
					City:   "Valbonne",
				},
				Job:          "test",
				Status:       Active,
				Married:      true,
				BirthDate:    &birth,
				Degree:       "test",
				Phone:        12345850,
				Other:        []string{"tt", "aa", "bb"},
				LuckyNumbers: []int{12, 21},
				Weight:       100,
				Height:       10,
				Id:           "uc123",
				Languages:    []string{"french"},
				Email:        "test@test.com",
				Available:    true,
				Sex:          "M",
				Conflict:     "test",
			}},
			expected: "age = 34\nbirthdate = \"2024-10-08T00:00:00Z\"\nlanguages = [\"french\"]\nluckyNumbers = [12, 21]" +
				"\nphone = 12345850\nsex = \"M\"\nstatus = 1\nweight = 100\n", // TOML representation
		},
	}
	for _, tt := range tests {
		var expectedTOML, resultTOML interface{}

		t.Run(tt.expected, func(t *testing.T) {

			result, err := toml.Marshal(tt.model)

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if err = toml.Unmarshal([]byte(tt.expected), &expectedTOML); err != nil {
				t.Fatalf("failed to unmarshal expected TOML: %v", err)
			}

			if err = toml.Unmarshal(result, &resultTOML); err != nil {
				t.Fatalf("failed to unmarshal result TOML: %v", err)
			}

			if !reflect.DeepEqual(expectedTOML, resultTOML) {
				t.Errorf("expected: %+v, got: %+v", expectedTOML, resultTOML)
			}
		})
	}
}

func TestModel_UnmarshalTOML(t *testing.T) {
	tests := []struct {
		input       string
		expected    retro.Model[Test]
		expectedErr bool
	}{
		{
			input:    "age = 25\nstatus = 1\nemail = \"test@testcom\"\n",
			expected: retro.Model[Test]{Struct: Test{Age: 25, Status: Active}},
		},
	}

	for _, tt := range tests {
		var result retro.Model[Test]

		t.Run(tt.input, func(t *testing.T) {
			err := toml.Unmarshal([]byte(tt.input), &result)

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

// Standard to true means that the retro Model will Marshalled Unmarshalled directly with the standard encoders
func TestModel_MarshalStandard(t *testing.T) {
	var (
		err error
		// Expected struct after unmarshaling
		expected = retro.Model[Standard]{
			Struct: Standard{
				A: 12,
				C: "test",
				D: []string{"a", "b"},
			},
			Standard: true,
		}
		// Struct to be used for unmarshaling and comparison
		m = retro.Model[Standard]{
			Standard: true,
		}
	)

	tests := []struct {
		input                                    retro.Model[Standard]
		expectedJSON, expectedYAML, expectedTOML string
		expectedErr                              bool
	}{
		{
			input: retro.Model[Standard]{
				Struct: Standard{
					A: 12,
					b: 25,
					C: "test",
					D: []string{"a", "b"},
				},
				Standard: true,
			},

			expectedJSON: `{"a":12,"C":"test","d":["a","b"]}`,
			expectedYAML: `
a: 12
C: test
b: 25
d:
  - a
  - b
`,
			expectedTOML: `a = 12
C = "test"
d = ["a", "b"]
`,
		},
	}

	for _, tt := range tests {

		if err = yaml.Unmarshal([]byte(tt.expectedYAML), &m); (err != nil) != tt.expectedErr {
			t.Fatalf("failed to unmarshal expected YAML: %v", err)
		}

		if !reflect.DeepEqual(m.Struct, expected.Struct) {
			t.Errorf("YAML: expected: %+v, got: %+v", expected, m)
		}

		if err = json.Unmarshal([]byte(tt.expectedJSON), &m); (err != nil) != tt.expectedErr {
			t.Fatalf("failed to unmarshal expected JSON: %v", err)
		}

		if !reflect.DeepEqual(m.Struct, expected.Struct) {
			t.Errorf("JSON: expected: %+v, got: %+v", expected, m)
		}

		if err = toml.Unmarshal([]byte(tt.expectedTOML), &m); (err != nil) != tt.expectedErr {
			t.Fatalf("failed to unmarshal expected TOML: %v", err)
		}

		if !reflect.DeepEqual(m.Struct, expected.Struct) {
			t.Errorf("TOML: expected: %+v, got: %+v", expected, m)
		}
	}
}
