/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package info

// defaultName retrieves the default name from the internal storage.
// Returns an empty string if the default name is not set or is not a string.
func (o *inf) defaultName() string {
	if i, l := o.v.Load(keyDefName); !l {
		return ""
	} else if v, k := i.(string); !k {
		return ""
	} else {
		return v
	}
}

// Name returns the name of the monitor info.
// If a name function is registered, it calls the function on first access and caches the result.
// If no function is registered or the cached value exists, it returns the cached or default name.
// This method is thread-safe and can be called concurrently.
func (o *inf) Name() string {
	if i, l := o.v.Load(keyName); !l {
		return o.getName()
	} else if v, k := i.(string); !k {
		return o.getName()
	} else {
		return v
	}
}

// Info returns a map of string to interface{} containing information about the monitor.
// If an info function is registered, it calls the function on first access and caches the results.
// If cached values exist, they are returned. Internal keys (keyDefName, keyName) are filtered out.
// This method is thread-safe and can be called concurrently.
// Returns nil if no info is available.
func (o *inf) Info() map[string]interface{} {
	var res = make(map[string]interface{}, 0)

	o.v.Range(func(key, value any) bool {
		if v, k := key.(string); !k {
			return true
		} else if v == keyDefName {
			return true
		} else if v == keyName {
			return true
		} else {
			res[v] = value
			return true
		}
	})

	if len(res) < 1 {
		return o.getInfo()
	}

	return res
}
