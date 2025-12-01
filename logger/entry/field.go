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

package entry

import (
	logfld "github.com/nabbar/golib/logger/fields"
)

// FieldAdd adds a single key-value pair to the entry's custom fields. The key must be a string,
// and the value can be any type that can be serialized to JSON by logrus.
//
// This method requires that fields have been initialized with FieldSet() before use. If fields
// are nil, this method returns nil.
//
// Parameters:
//   - key: The field key (string)
//   - val: The field value (any JSON-serializable type)
//
// Returns:
//   - The entry itself for method chaining, or nil if entry or fields are nil
//
// Example:
//
//	fields := logfld.New(nil)
//	e := New(loglvl.InfoLevel).FieldSet(fields)
//	e.FieldAdd("user_id", 12345).FieldAdd("action", "login")
func (e *entry) FieldAdd(key string, val interface{}) Entry {
	if e == nil {
		return nil
	} else if e.Fields == nil {
		return nil
	}

	e.Fields.Add(key, val)
	return e
}

// FieldMerge merges another Fields object into the entry's custom fields. Existing keys in the
// entry are overwritten by values from the provided fields object (shallow merge).
//
// This method requires that fields have been initialized with FieldSet() before use. If the
// entry's fields are nil, this method returns nil.
//
// Parameters:
//   - fields: The Fields object to merge into the entry's fields
//
// Returns:
//   - The entry itself for method chaining, or nil if entry or entry's fields are nil
//
// Example:
//
//	baseFields := logfld.New(nil)
//	baseFields.Add("app", "myapp")
//	additionalFields := logfld.New(nil)
//	additionalFields.Add("request_id", "req-123")
//	e := New(loglvl.InfoLevel).FieldSet(baseFields).FieldMerge(additionalFields)
func (e *entry) FieldMerge(fields logfld.Fields) Entry {
	if e == nil {
		return nil
	} else if e.Fields == nil {
		return nil
	}

	e.Fields.Merge(fields)
	return e
}

// FieldSet replaces the entry's entire fields object with the provided Fields object. This
// initializes or resets the custom fields of the entry.
//
// This method must be called before using FieldAdd(), FieldMerge(), or FieldClean() methods.
// It is safe to pass nil to clear the fields.
//
// Parameters:
//   - fields: The Fields object to set, or nil to clear fields
//
// Returns:
//   - The entry itself for method chaining, or nil if entry is nil
//
// Example:
//
//	fields := logfld.New(nil)
//	fields.Add("service", "api")
//	e := New(loglvl.InfoLevel).FieldSet(fields)
func (e *entry) FieldSet(fields logfld.Fields) Entry {
	if e == nil {
		return nil
	}

	e.Fields = fields
	return e
}

// FieldClean removes one or more keys from the entry's custom fields. Keys that do not exist
// are silently ignored. If no keys are provided, the entry is returned unchanged.
//
// This method requires that fields have been initialized with FieldSet() before use. If fields
// are nil, this method returns nil.
//
// Parameters:
//   - keys: Variable number of field keys to remove from the entry
//
// Returns:
//   - The entry itself for method chaining, or nil if entry or fields are nil
//
// Example:
//
//	fields := logfld.New(nil)
//	fields.Add("key1", "value1")
//	fields.Add("key2", "value2")
//	e := New(loglvl.InfoLevel).FieldSet(fields).FieldClean("key1")
func (e *entry) FieldClean(keys ...string) Entry {
	if e == nil {
		return nil
	} else if e.Fields == nil {
		return nil
	}

	for _, k := range keys {
		e.Fields.Delete(k)
	}

	return e
}
