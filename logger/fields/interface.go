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

package fields

import (
	"context"
	"encoding/json"

	libctx "github.com/nabbar/golib/context"
	"github.com/sirupsen/logrus"
)

type Fields interface {
	context.Context
	//libctx.Config[string]
	json.Marshaler
	json.Unmarshaler

	// Clone creates a new Fields instance from the current one, using the provided context.
	// The new Fields instance will have the same values as the current one, but will be a separate entity.
	// This is useful when you want to modify the fields of an entry without affecting the original entry.
	Clone() Fields
	Clean()
	// Add adds a new key/val pair to the Fields instance.
	// It returns the modified Fields instance.
	//
	// If the key already exists in the Fields instance, its value will be overwritten.
	// If the key does not exist in the Fields instance, a new key/val pair will be added.
	//
	// This method is useful when you want to add a new key/val pair to an existing Fields instance.
	Add(key string, val interface{}) Fields
	Delete(key string) Fields
	Merge(f Fields) Fields
	Walk(fct libctx.FuncWalk[string]) Fields
	WalkLimit(fct libctx.FuncWalk[string], validKeys ...string) Fields

	Get(key string) (val interface{}, ok bool)
	LoadOrStore(key string, cfg interface{}) (val interface{}, loaded bool)
	LoadAndDelete(key string) (val interface{}, loaded bool)

	// Logrus returns the logrus.Fields instance associated with the current Fields instance.
	//
	// This method is useful when you want to directly access the logrus.Fields instance
	// associated with the current Fields instance.
	//
	// The returned logrus.Fields instance is a reference to the same instance as the one
	// associated with the current Fields instance. Any modification to the returned logrus.Fields
	// instance will affect the Fields instance.
	//
	// The returned logrus.Fields instance is valid until the Fields instance is modified or
	// until the Fields instance is garbage collected.
	//
	// If the Fields instance is nil, this method will return nil.
	Logrus() logrus.Fields
	// Map applies a transformation function to all key/val pairs in the Fields instance.
	// It returns a new Fields instance with the transformed key/val pairs.
	//
	// The transformation function is called for each key/val pair in the Fields instance.
	// The transformation function takes the key and val as arguments, and returns the new val to be used in the transformed Fields instance.
	//
	// If the transformation function returns a nil value, the key/val pair will not be included in the transformed Fields instance.
	//
	// This method is useful when you want to transform the key/val pairs of an existing Fields instance.
	//
	// The returned Fields instance is a new separate instance, and does not reference the original Fields instance in any way.
	//
	// If the Fields instance is nil, this method will return nil.
	Map(fct func(key string, val interface{}) interface{}) Fields
}

// New creates a new Fields instance from the given context.Context.
//
// It returns a new Fields instance which is associated with the given context.Context.
// The returned Fields instance can be used to add, remove, or modify key/val pairs.
//
// If the given context.Context is nil, this method will return nil.
//
// Example usage:
//
//	 flds := New(context.Background)
//	 flds.Add("key", "value")
//	 flds.Map(func(key string, val interface{}) interface{} {
//		return fmt.Sprintf("%s-%s", key, val)
//	 })
//
// The above example shows how to create a new Fields instance from a context.Context,
// and how to use the returned Fields instance to add, remove, or modify key/val pairs.
func New(ctx context.Context) Fields {
	return &fldModel{
		c: libctx.New[string](ctx),
	}
}
