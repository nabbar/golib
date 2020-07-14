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
	"strconv"
)

var msgfct = make([]Message, 0)

type Message func(code CodeError) (message string)
type CodeError uint16

const UNK_ERROR CodeError = 0
const UNK_MESSAGE = "unknown error"

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

	for _, f := range msgfct {
		m := f(c)
		if m != "" {
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

func RegisterFctMessage(fct Message) {
	msgfct = append(msgfct, fct)
}
