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
 */

package status

import "sync"

const DefMessageOK = "OK"
const DefMessageKO = "KO"

type Response struct {
	InfoResponse
	StatusResponse

	m          sync.Mutex
	Components []CptResponse `json:"components"`
}

func (r Response) IsOk() bool {
	if len(r.Components) < 1 {
		return true
	}

	for _, c := range r.Components {
		if c.Status != DefMessageOK {
			return false
		}
	}

	return true
}

func (r Response) IsOkMandatory() bool {
	if len(r.Components) < 1 {
		return true
	}

	for _, c := range r.Components {
		if !c.Mandatory {
			continue
		}

		if c.Status != DefMessageOK {
			return false
		}
	}

	return true
}

func (r *Response) appendNewCpt(cpt CptResponse) {
	r.m.Lock()
	defer r.m.Unlock()

	r.Components = append(r.Components, cpt)
}
