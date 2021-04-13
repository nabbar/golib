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
	"encoding/json"
	goErr "errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReturnError func(code int, msg string, file string, line int)

type Return interface {
	SetError(code int, msg string, file string, line int)
	AddParent(code int, msg string, file string, line int)
}

type DefaultReturn struct {
	Code    string
	Message string
}

func NewDefaultReturn() *DefaultReturn {
	return &DefaultReturn{
		Code:    "",
		Message: "",
	}
}

func (r *DefaultReturn) SetError(code int, msg string, file string, line int) {
	r.Code = fmt.Sprintf("%d", code)
	r.Message = msg
}

func (r *DefaultReturn) AddParent(code int, msg string, file string, line int) {
	// nothing
}

func (r DefaultReturn) JSON() []byte {
	if str, err := json.Marshal(r); err != nil {
		return make([]byte, 0)
	} else {
		return str
	}
}

func (r DefaultReturn) GinTonicAbort(ctx *gin.Context, httpCode int) {
	if ctx == nil || ctx.IsAborted() {
		return
	}

	if httpCode == 0 {
		httpCode = http.StatusInternalServerError
	}

	ctx.AbortWithStatusJSON(httpCode, r)
}

func (r DefaultReturn) GinTonicErrorAbort(ctx *gin.Context, httpCode int) {
	if ctx == nil || ctx.IsAborted() {
		return
	}

	ctx.Errors = append(ctx.Errors, &gin.Error{
		//nolint #goerr113
		Err:  goErr.New(r.Message),
		Type: gin.ErrorTypeAny,
	})

	r.GinTonicAbort(ctx, httpCode)
}
