/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package errors

import (
	"sort"
	"strconv"
)

var idMsgFct = make(map[CodeError]Message)

type Message func(code CodeError) (message string)
type CodeError uint16

const UNK_ERROR CodeError = 0
const UNK_MESSAGE = "unknown error"
const NUL_MESSAGE = ""

func (c CodeError) GetUint16() uint16 {
	return uint16(c)
}

func (c CodeError) GetInt() int {
	return int(c)
}

func (c CodeError) GetString() string {
	return strconv.Itoa(c.GetInt())
}

func (c CodeError) GetMessage() string {
	if c == UNK_ERROR {
		return UNK_MESSAGE
	}

	if f, ok := idMsgFct[findCodeErrorInMapMessage(c)]; ok {
		if m := f(c); m != "" {
			return m
		}
	}

	return UNK_MESSAGE
}

func (c CodeError) Error(p Error) Error {
	return NewError(c.GetUint16(), c.GetMessage(), p)
}

func (c CodeError) ErrorParent(p ...error) Error {
	e := c.Error(nil)
	e.AddParent(p...)
	return e
}

func (c CodeError) IfError(e Error) Error {
	return NewErrorIfError(c.GetUint16(), c.GetMessage(), e)
}

func (c CodeError) Iferror(e error) Error {
	return NewErrorIferror(c.GetUint16(), c.GetMessage(), e)
}

func RegisterIdFctMessage(minCode CodeError, fct Message) {
	if idMsgFct == nil {
		idMsgFct = make(map[CodeError]Message)
	}

	idMsgFct[minCode] = fct
	orderMapMessage()
}

func ExistInMapMessage(code CodeError) bool {
	if f, ok := idMsgFct[findCodeErrorInMapMessage(code)]; ok {
		if m := f(code); m != NUL_MESSAGE {
			return true
		}
	}

	return false
}

func getMapMessageKey() []CodeError {
	var (
		keys = make([]int, 0)
		res  = make([]CodeError, 0)
	)

	for k := range idMsgFct {
		keys = append(keys, int(k.GetUint16()))
	}

	sort.Ints(keys)

	for _, k := range keys {
		res = append(res, CodeError(k))
	}

	return res
}

func orderMapMessage() {
	var res = make(map[CodeError]Message)

	for _, k := range getMapMessageKey() {
		res[k] = idMsgFct[k]
	}

	idMsgFct = res
}

func findCodeErrorInMapMessage(code CodeError) CodeError {
	var res CodeError = 0

	for _, k := range getMapMessageKey() {
		if k <= code && k > res {
			res = k
		}
	}

	return res
}
