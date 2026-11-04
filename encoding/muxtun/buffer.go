/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package muxtun

func (o *mdl) getBuffer(n int) *[]byte {
	var pCpy *[]byte

	// Attempt to retrieve a recycled pointer from the pool.
	if v := o.b.Get(); v != nil {
		pCpy = v.(*[]byte)
	}

	// Fallback to fresh allocation if the pool is empty or the retrieved buffer is nil.
	if pCpy == nil {
		buf := make([]byte, n)
		return &buf
	}

	// Resizing Logic:
	// If the underlying array's capacity is insufficient, we must allocate a new one.
	if cap(*pCpy) < n {
		buf := make([]byte, n)
		return &buf
	}

	// Re-slice the pooled buffer to the exact requested length; the underlying array is reused.
	*pCpy = (*pCpy)[:n]
	return pCpy
}

func (o *mdl) putBuffer(buf *[]byte) {
	o.b.Put(buf)
}
