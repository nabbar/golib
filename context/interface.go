/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context

import (
	"context"

	libatm "github.com/nabbar/golib/atomic"
)

type FuncContextConfig[T comparable] func() Config[T]
type FuncWalk[T comparable] func(key T, val interface{}) bool

type MapManage[T comparable] interface {
	// Clean removes all the key-value pairs from the map.
	// It is atomic and safe for concurrent use.
	// If the map is empty, it returns immediately.
	// It is safe to call Clean while other goroutines are calling Load or Store.
	Clean()
	// Load loads the value associated with the given key from the map.
	// It returns the loaded value and true if the key was found, false otherwise.
	// It is atomic and safe for concurrent use.
	// If the key doesn't exist, it returns nil and false.
	// If the value is nil, it returns nil and true.
	Load(key T) (val interface{}, ok bool)
	// Store stores the given value in the map associated with the key.
	// It is atomic and safe for concurrent use.
	// If the key already exists, the value is overwritten.
	// If the value is nil, the key is removed from the map.
	// It returns nothing.
	Store(key T, cfg interface{})
	// Delete deletes the value associated with the given key from the map.
	// It returns true if the key was found and deleted, false otherwise.
	// It is atomic and safe for concurrent use.
	Delete(key T)
}

type Context interface {
	// GetContext returns the context associated with the current Config.
	// It is safe for concurrent use.
	// If the context function associated with the current Config is nil,
	// it returns context.Background.
	GetContext() context.Context
}

type Config[T comparable] interface {
	context.Context
	MapManage[T]
	Context

	// Clone creates an independent copy of the current Config.
	// It returns a new Config which references a different underlying map.
	// If the given context is nil, it uses the context from the current Config.
	// If the given context is canceled before the clone operation is complete,
	// it returns nil.
	// It is atomic and safe for concurrent use.
	Clone(ctx context.Context) Config[T]
	// Merge merges the values from another Config into the current one.
	// It returns true if the merge was successful, false otherwise.
	// It is atomic and safe for concurrent use.
	// If the given Config is nil, it returns false.
	// If the given Config shares the same underlying map as the current one,
	// it returns false.
	Merge(cfg Config[T]) bool
	// Walk iterates over all the key-value pairs in the map and calls the given function
	// for each pair. It returns true if all iterations were successful, false otherwise.
	// It is atomic and safe for concurrent use.
	// If the map is empty, it returns true.
	Walk(fct FuncWalk[T])
	// WalkLimit iterates over the given validKeys and calls the given function for each
	// key-value pair. It returns true if all iterations were successful, false otherwise.
	// It is atomic and safe for concurrent use.
	// If the given validKeys are empty, it returns true.
	WalkLimit(fct FuncWalk[T], validKeys ...T)

	// LoadOrStore loads the value for the given key, or stores the given value
	// if the key doesn't exist. It returns the loaded value and whether the key was
	// loaded (true) or not (false). If the key doesn't exist, it returns the
	// stored value and false.
	// It is atomic and safe for concurrent use.
	LoadOrStore(key T, cfg interface{}) (val interface{}, loaded bool)
	// LoadAndDelete loads the value for the given key and deletes it from the map.
	// It returns the loaded value and whether the key was loaded (true) or not (false).
	// If the key doesn't exist, it returns nil and false.
	// It is atomic and safe for concurrent use.
	LoadAndDelete(key T) (val interface{}, loaded bool)
}

// New returns a new Config with the given context function.
// If the context function is nil, it defaults to context.Background.
// The returned Config has the given context and a map to store values.
func New[T comparable](ctx context.Context) Config[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	return &ccx[T]{
		m: libatm.NewMapAny[T](),
		x: ctx,
	}
}

// NewConfig returns a new Config with the given context function.
// It is a shortcut for New, and can be used in the same way.
//
// If the context function is nil, it defaults to context.Background.
// The returned Config has the given context and a map to store values.
// Deprecated: see New
func NewConfig[T comparable](ctx context.Context) Config[T] {
	return New[T](ctx)
}
