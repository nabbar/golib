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

package fields

import (
	"encoding/json"

	libctx "github.com/nabbar/golib/context"
	"github.com/sirupsen/logrus"
)

// fldModel is the internal implementation of the Fields interface.
//
// It wraps a github.com/nabbar/golib/context.Config[string] to provide thread-safe
// key-value storage with context integration. This struct should not be used directly;
// use the Fields interface and New() constructor instead.
type fldModel struct {
	c libctx.Config[string]
}

// Add inserts or updates a key-value pair and returns the Fields instance for chaining.
//
// This method delegates to the underlying thread-safe storage, making it safe for
// concurrent use. It's commonly used for building field sets incrementally.
//
// Example:
//
//	flds.Add("key1", "value1").Add("key2", "value2")
func (o *fldModel) Add(key string, val interface{}) Fields {
	o.c.Store(key, val)
	return o
}

// Logrus converts the Fields instance to logrus.Fields format.
//
// This method creates a new map containing all key-value pairs. It's thread-safe
// for concurrent reads. If the receiver is nil, returns an empty map.
//
// Note: A new map is created on each call. For performance-critical code,
// consider caching the result.
//
// Example:
//
//	logrusFields := flds.Logrus()
//	logger.WithFields(logrusFields).Info("message")
func (o *fldModel) Logrus() logrus.Fields {
	var res = make(logrus.Fields, 0)

	if o == nil {
		return res
	} else if o.c == nil {
		return res
	}

	o.c.Walk(func(key string, val interface{}) bool {
		res[key] = val
		return true
	})
	return res
}

// Map applies a transformation function to all field values.
//
// This method iterates over all fields and replaces each value with the result
// of the transformation function. This is a composite operation that requires
// external synchronization if used concurrently with other writes.
//
// Example:
//
//	flds.Map(func(key string, val interface{}) interface{} {
//		if key == "password" {
//			return "[REDACTED]"
//		}
//		return val
//	})
func (o *fldModel) Map(fct func(key string, val interface{}) interface{}) Fields {
	o.c.Walk(func(key string, val interface{}) bool {
		o.c.Store(key, fct(key, val))
		return true
	})

	return o
}

// MarshalJSON implements json.Marshaler interface.
//
// It converts the Fields instance to JSON by first converting to logrus.Fields
// and then marshaling that map. This ensures compatibility with logrus-based systems.
//
// The resulting JSON is a flat object with string keys and arbitrary values.
//
// Example:
//
//	flds.Add("key", "value")
//	data, err := json.Marshal(flds)
//	// data = {"key":"value"}
func (o *fldModel) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Logrus())
}

// UnmarshalJSON implements json.Unmarshaler interface.
//
// It populates the Fields instance from a JSON object. The JSON must be an object
// with string keys. Non-object JSON will result in an error.
//
// Note: This merges the JSON data with existing fields. Use Clean() first if you
// want to completely replace the fields.
//
// Example:
//
//	var flds Fields = fields.New(ctx)
//	err := json.Unmarshal([]byte(`{"key":"value"}`), flds)
func (o *fldModel) UnmarshalJSON(bytes []byte) error {
	var l = make(logrus.Fields)

	if e := json.Unmarshal(bytes, &l); e != nil {
		return e
	} else if len(l) > 0 {
		for k, v := range l {
			o.c.Store(k, v)
		}
	}

	return nil
}

// Clone creates an independent deep copy of the Fields instance.
//
// The returned Fields instance has its own internal storage and can be modified
// without affecting the original. This is essential for creating derived field
// sets or for use in concurrent goroutines.
//
// Note: While the map is deep copied, the values themselves are not. If values
// are pointers, modifications to the pointed data will affect all clones.
//
// Example:
//
//	clone := original.Clone()
//	clone.Add("extra", "field") // Doesn't affect original
func (o *fldModel) Clone() Fields {
	return &fldModel{
		o.c.Clone(o.c), // nolint
	}
}
