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

// Name returns the effective name of the monitored component.
// The name is resolved following a specific order of precedence:
// 1. If a dynamic name provider function has been registered (via RegisterName), it is executed and its result returned.
// 2. If no dynamic function is available or it returns an empty string, the name manually set via SetName (if any) is returned.
// 3. If no manual override is present, the default name provided during the initialization of the Info instance is returned.
//
// This method is designed to be thread-safe and can be invoked concurrently by multiple goroutines.
func (o *inf) Name() string {
	if o == nil {
		return ""
	}

	return o.getName()
}

// Data retrieves and returns all metadata associated with the monitored component as a map.
// The resulting map is a combination of two sources:
//   - Static metadata: Key-value pairs manually stored within the Info instance using SetData or AddData.
//   - Dynamic metadata: Data provided by the registered information provider function (if any), which is
//     executed during every call to this method.
//
// If keys between the static and dynamic sources collide, the data from the dynamic provider function
// takes precedence in the final returned map.
//
// This method is thread-safe and ensures consistent data access across concurrent operations.
// It returns an empty map if no metadata is currently available.
func (o *inf) Data() map[string]interface{} {
	if o == nil {
		return nil
	}

	return o.getInfo()
}
