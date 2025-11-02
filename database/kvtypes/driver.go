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

type FctWalk[K comparable, M any] func(key K, model M) bool

type KVDriver[K comparable, M any] interface {
	// New returns a new instance of the KVDriver interface.
	// The function takes no parameter and returns a new instance of the interface.
	// The returned instance is not initialized and needs to be initialized before use.
	New() KVDriver[K, M]
	// Get retrieves the value associated with the given key.
	// The function will return nil if the key does not exist.
	// If an error occurs during the retrieval, it will be returned.
	// The iteration order is not defined.
	// If no key is found, the function will return nil.
	Get(key K, model *M) error
	// Set sets the value associated with the given key.
	// If an error occurs during the setting, it will be returned.
	// The iteration order is not defined.
	// If no key is found, the function will return an error.
	Set(key K, model M) error
	// Del deletes the key-value pair associated with the given key.
	// If the key does not exist, the function will return nil.
	// If an error occurs during the deletion, it will be returned.
	// The iteration order is not defined.
	Del(key K) error
	// List returns a list of all keys in the driver.
	// The function will return an error if it occurs during the iteration.
	// The iteration order is not defined.
	// If no key is found, the function will return an empty list.
	List() ([]K, error)
	// Search returns a list of keys that match the given pattern.
	// The function will return an error if it occurs during the iteration.
	// The iteration order is not defined.
	// If no key matches the pattern, the function will return an empty list.
	Search(pattern K) ([]K, error)
	// Walk will call the given function for each key-value pair in the driver.
	// The function will be called with the key and the value associated with the key.
	// If the function returns false, the iteration will stop.
	// If an error occurs during the iteration, it will be returned.
	// The iteration order is not defined.
	Walk(fct FctWalk[K, M]) error
}
