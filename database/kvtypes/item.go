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

package kvtypes

type KVItem[K comparable, M any] interface {
	// Set sets the value of the item to the given model.
	Set(model M)
	// Get returns the current value of the item.
	Get() M
	// Key returns the key associated with the item. This key is used to identify
	// the item in the storage.
	Key() K

	// Load loads the item from the storage. It returns an error if the item
	// cannot be loaded.
	Load() error
	// Store saves the item to the storage. If force is true, the item is saved even
	// if the item has not changed since it was last loaded or stored. If force is
	// false, the item is only saved if it has changed since it was last loaded or
	// stored. It returns an error if the item cannot be stored.
	Store(force bool) error
	// Remove removes the item from the storage. It returns an error if the item
	// has not been loaded or if the item cannot be removed.
	Remove() error
	// Clean resets the item to its initial state. This is useful when the item
	// is re-used and the old state should be forgotten.
	Clean()

	// HasChange returns true if the current value of the item has changed since
	// it was last loaded or stored.
	HasChange() bool
}
