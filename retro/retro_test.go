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
	"github.com/nabbar/golib/retro"
	"gopkg.in/yaml.v3"
	"reflect"
	"testing"
	"time"
)

type Test struct {
	LastName     string     `json:"lastName" yaml:"lastName" retro:"v1.0.0,>v1.0.3"` // greater than v1.0.3 except v1.0.0
	Age          int        `json:"age,omitempty" yaml:"age,omitempty"`              // this if not 0 should be included in all the models because no retro tag
	Name         string     `json:"name"  yaml:"name" retro:">v1.0.0"`               // only if version strict greater than v1.0.0
	Version      string     `json:"version,omitempty" yaml:"version,omitempty"`
	Salary       float64    `json:"salary,omitempty" yaml:"salary,omitempty" retro:"<v1.0.2"`                                // only if version strict lesser than v1.0.2
	Active       bool       `json:"active,omitempty" yaml:"active,omitempty" retro:">=v1.0.2"`                               // only if version greater or equal to v1.0.2
	Address      Address    `json:"address" yaml:"address" retro:"<=v1.0.0"`                                                 // only if version lesser or equal to v1.0.0
	Job          string     `json:"job" yaml:"job" retro:">v1.0.1,<v1.0.4"`                                                  // only v1.0.2 and v1.0.3
	Status       Status     `json:"status" yaml:"status" retro:"default,>v1.0.1,<v1.0.4"`                                    // default (meaning no versioning) or only v1.0.2 and v1.0.3
	Married      bool       `json:"married" yaml:"married" retro:">=v1.0.0,<v1.0.2"`                                         // v1.0.0 v1.0.1
	BirthDate    *time.Time `json:"birthdate,omitempty" yaml:"birthdate,omitempty" retro:"default,>=v1.0.0,<v1.0.2"`         // v1.0.0 v1.0.1 and default
	Degree       string     `json:"degree" yaml:"degree" retro:">v1.0.0,<=v1.0.2"`                                           // v1.0.1 v1.0.2
	Phone        int32      `json:"phone,omitempty" yaml:"phone,omitempty" retro:"default,>v1.0.0,<=v1.0.2"`                 // v1.0.1 v1.0.2 and default
	Other        []string   `json:"other" yaml:"other" retro:">=v1.0.0,<=v1.0.2"`                                            // v1.0.0 v1.0.1 v1.0.2
	LuckyNumbers []int      `json:"luckyNumbers,omitempty" yaml:"luckyNumbers,omitempty" retro:"default, >=v1.0.0,<=v1.0.2"` // v1.0.0 v1.0.1 v1.0.2 and default
	Weight       int        `json:"weight,omitempty" yaml:"weight,omitempty" retro:"default,v1.0.0"`                         // v1.0.0 and default
	Height       int        `json:"height" yaml:"height" retro:"v1.0.0"`                                                     // v1.0.0 only
	Id           string     `json:"id" yaml:"id" retro:"v1.0.0,v1.0.3"`                                                      // v1.0.0 and v1.0.3 only
	Languages    []string   `json:"languages,omitempty" yaml:"languages,omitempty" retro:"default,v1.0.0,v1.0.3"`            // v1.0.0 and v1.0.3 and default
	Email        string     `json:"email" yaml:"email" retro:"<v1.0.0,v1.0.3"`                                               // lesser than v1.0.0 expect v1.0.3
	Available    bool       `json:"available" yaml:"available" retro:">v1.0.0,<=v1.0.3"`                                     // v1.0.1 v1.0.2 v1.0.3
	Sex          string     `json:"sex,omitempty" yaml:"sex,omitempty" retro:">v1.0.0,<=v1.0.3, v0.0.3, default"`            // between v1.0.0 and v1.0.3 (expect v0.0.3) and default
	Conflict     string     `json:"conflict" yaml:"conflict" retro:">v1.0.0,>v1.0.3"`                                        // this field has non-valid retro definition and should be always ignored
}

type Address struct {
	Street string `json:"street" `
	City   string `json:"city,omitempty"`
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
	tests := []struct {
		model       retro.Model[Test]
		expected    string
		expectedErr bool
	}{
		{ // Test Default No Versioning + age field with no retro tag
			model:    retro.Model[Test]{Fields: Test{Age: 25, Name: "Alice", Active: true, Address: Address{Street: "123 Main St", City: "Wonderland"}, Status: Active}},
			expected: `{"age":25,"status":"active"}`,
		},
		{ // v1.0.3
			model: retro.Model[Test]{Fields: Test{Age: 25, Name: "Alice", Active: true, Job: "test", Id: "uc123", Email: "test@example.com",
				Available: true, Sex: "M", Version: "v1.0.3", Address: Address{Street: "123 Main St", City: "Wonderland"}, Status: Active}},
			expected: `{"version":"v1.0.3","age":25,"status":"active","name":"Alice", "active":true,"job":"test","id":"uc123","email":"test@example.com","available":true,"sex":"M"}`,
		},
		{ // v0.0.3

			model:    retro.Model[Test]{Fields: Test{Age: 15, Version: "v0.0.3", Salary: 100, Address: Address{Street: "123 Main St", City: "Wonderland"}, Email: "test", Sex: "F"}},
			expected: `{"version":"v0.0.3","age":15,"email":"test","sex":"F","salary":100,"address":{"street":"123 Main St","city":"Wonderland"}}`,
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
	tests := []struct {
		input       string
		expected    retro.Model[Test]
		expectedErr bool
	}{
		{
			input:    `{"age":25,"status":"active","weight":70}`,
			expected: retro.Model[Test]{Fields: Test{Age: 25, Status: Active, Weight: 70}},
		},
		{
			input: `{"version":"v1.0.3","age":25,"status":"active","name":"Alice", "active":true,"job":"test","id":"uc123","email":"test@example.com","available":true,"sex":"M"}`,
			expected: retro.Model[Test]{Fields: Test{Age: 25, Name: "Alice", Active: true, Job: "test", Id: "uc123", Email: "test@example.com",
				Available: true, Sex: "M", Version: "v1.0.3", Status: Active}},
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
			model: retro.Model[Test]{Fields: Test{Age: 25, Status: Active}},
			expected: `age: 25
status: 1
`,
			expectedErr: false,
		},
		{ // v0.0.3
			model: retro.Model[Test]{Fields: Test{
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
			var expectedMap, resultMap map[string]interface{}

			result, err := tt.model.MarshalYAML()

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if err = yaml.Unmarshal([]byte(tt.expected), &expectedMap); err != nil {
				t.Fatalf("failed to unmarshal expected YAML: %v", err)
			}

			if err = yaml.Unmarshal([]byte(result.(string)), &resultMap); err != nil {
				t.Fatalf("failed to unmarshal actual YAML: %v", err)
			}

			if !reflect.DeepEqual(expectedMap, resultMap) {
				t.Errorf("expected: %+v, got: %+v", expectedMap, resultMap)
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
			expected: retro.Model[Test]{Fields: Test{Age: 25, Status: Active, Weight: 70}},
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
			expected: retro.Model[Test]{Fields: Test{Age: 25, Name: "Alice", Active: true, Job: "test", Id: "uc123", Email: "test@example.com",
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
