/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package multipart

func (m *mpu) RegisterFuncOnPushPart(fct func(eTag string, e error)) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	m.fp = fct
}

func (m *mpu) callFuncOnPushPart(eTag string, e error) {
	if m == nil {
		return
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if m.fp != nil {
		m.fp(eTag, e)
	}
}

func (m *mpu) RegisterFuncOnAbort(fct func(nPart int, obj string, e error)) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	m.fa = fct
}

func (m *mpu) RegisterFuncOnComplete(fct func(nPart int, obj string, e error)) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	m.fc = fct
}

func (m *mpu) callFuncOnComplete(abort bool, nPart int, obj string, e error) {
	if m == nil {
		return
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if !abort && m.fc != nil {
		m.fc(nPart, obj, e)
	} else if abort && m.fa != nil {
		m.fa(nPart, obj, e)
	}
}
