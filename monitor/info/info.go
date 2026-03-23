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

// Name returns the current name of the monitor info component.
// If a dynamic name function has been registered via RegisterName, it is invoked
// to retrieve the name. Otherwise, it falls back to any manually set name (via SetName)
// or the default name provided at creation.
//
// This method is thread-safe and safe to call concurrently.
func (o *inf) Name() string {
	if o == nil {
		return ""
	}

	return o.getName()
}

// Info returns the current information data as a map of key-value pairs.
// It combines any manually set data (via SetData/AddData) with the result of
// the registered info function (if any). The registered function is invoked
// on every call to Info.
//
// This method is thread-safe and safe to call concurrently.
// It returns an empty map if no information is available.
func (o *inf) Info() map[string]interface{} {
	if o == nil {
		return nil
	}

	return o.getInfo()
}
