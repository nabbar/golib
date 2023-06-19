/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package entry

import (
	logfld "github.com/nabbar/golib/logger/fields"
)

// FieldAdd allow to add one couple key/val as type string/interface into the custom field of the entry.
func (e *entry) FieldAdd(key string, val interface{}) Entry {
	e.Fields.Add(key, val)
	return e
}

// FieldMerge allow to merge a Field pointer into the custom field of the entry.
func (e *entry) FieldMerge(fields logfld.Fields) Entry {
	e.Fields.Merge(fields)
	return e
}

// FieldSet allow to change the custom field of the entry with the given Fields in parameter.
func (e *entry) FieldSet(fields logfld.Fields) Entry {
	e.Fields = fields
	return e
}

func (e *entry) FieldClean(keys ...string) Entry {
	for _, k := range keys {
		e.Fields.Delete(k)
	}

	return e
}
