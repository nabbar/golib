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

type FuncWalk[K comparable, M any] func(kv KVItem[K, M]) bool

type KVTable[K comparable, M any] interface {
	// Get returns the item associated with the given key.
	// If the key does not exist, the function will return nil and no error.
	// If an error occurs during the retrieval, it will be returned.
	// The iteration order is not defined.
	Get(key K) (KVItem[K, M], error)
	// Del deletes the item associated with the given key.
	// If the key does not exist, the function will return nil.
	// If an error occurs during the deletion, it will be returned.
	// The iteration order is not defined.
	Del(key K) error
	// List returns a list of all items in the table.
	// The function will return an error if it occurs during the iteration.
	// The iteration order is not defined.
	// If no item is found, the function will return an empty list.
	List() ([]KVItem[K, M], error)
	// Search returns a list of all items in the table that match the given pattern.
	// The function will return an error if it occurs during the iteration.
	// The iteration order is not defined.
	// If no item matches the pattern, the function will return an empty list.
	Search(pattern K) ([]KVItem[K, M], error)
	// Walk iterates over all items in the table and calls the provided function.
	// The function is called with the key and the value associated with the key.
	// If the function returns false, the iteration stops.
	// If the function returns an error, the iteration stops and the error is returned.
	// If no error occurs, the function returns nil.
	Walk(fct FuncWalk[K, M]) error
}
