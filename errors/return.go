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

type Return interface {
	// SetError set the error with the given code, message, file and line.
	//
	// It will create a new error with the given information and append it to the current error list.
	// If the error list is empty, it will create a new one.
	//
	// Parameters:
	// - code (int): error code
	// - msg (string): error message
	// - file (string): file where the error occurs
	// - line (int): line where the error occurs
	SetError(code int, msg string, file string, line int)

	// AddParent is used to add a parent error to the current error.
	//
	// It will create a new error with the given information and add it to the parent error list of the current error.
	//
	// Parameters:
	// - code (int): error code
	// - msg (string): error message
	// - file (string): file where the error occurs
	// - line (int): line where the error occurs
	AddParent(code int, msg string, file string, line int)

	// JSON return the JSON representation of the current error.
	//
	// It will generate a JSON object with the following structure:
	// {
	// 	"code": <int>,
	// 	"msg": <string>,
	// 	"parents": [
	// 		{
	// 			"code": <int>,
	// 			"msg": <string>,
	// 			"file": <string>,
	// 			"line": <int>
	// 		}
	// 	]
	// }
	//
	// Parameters: None
	//
	// Returns: []byte, a JSON representation of the current error
	JSON() []byte
}

type ReturnGin interface {
	Return

	// GinTonicAbort is used to abort the current request with the given HTTP status code.
	//
	// It will write the JSON representation of the current error to the response writer with the given HTTP status code.
	//
	// Parameters:
	// - ctx (*gin.Context): the gin context
	// - httpCode (int): the HTTP status code to return
	//
	// Returns: None
	GinTonicAbort(ctx *gin.Context, httpCode int)

	// GinTonicErrorAbort is used to abort the current request with the given HTTP status code,
	// and write the JSON representation of the current error to the response writer.
	//
	// Parameters:
	// - ctx (*gin.Context): the gin context
	// - httpCode (int): the HTTP status code to return
	//
	// Returns: None
	GinTonicErrorAbort(ctx *gin.Context, httpCode int)
}

type DefaultReturn struct {
	ReturnGin

	Code    string
	Message string
	err     []error
}

func (r *DefaultReturn) SetError(code int, msg string, file string, line int) {
	r.Code = fmt.Sprintf("%d", code)
	r.Message = msg

	if len(r.err) < 1 {
		r.err = make([]error, 0)
	}

	r.err = append(r.err, NewErrorTrace(code, msg, file, line, nil))
}

func (r *DefaultReturn) AddParent(code int, msg string, file string, line int) {
	if len(r.err) < 1 {
		r.err = make([]error, 0)
	}

	r.err = append(r.err, NewErrorTrace(code, msg, file, line, nil))
}

func (r *DefaultReturn) JSON() []byte {
	if str, err := json.Marshal(r); err != nil {
		return make([]byte, 0)
	} else {
		return str
	}
}

func (r *DefaultReturn) GinTonicAbort(ctx *gin.Context, httpCode int) {
	if ctx == nil || ctx.IsAborted() {
		return
	}

	if httpCode == 0 {
		httpCode = http.StatusInternalServerError
	}

	ctx.AbortWithStatusJSON(httpCode, r)
}

func (r *DefaultReturn) GinTonicErrorAbort(ctx *gin.Context, httpCode int) {
	if ctx == nil || ctx.IsAborted() {
		return
	}

	ctx.Errors = append(ctx.Errors, &gin.Error{
		//nolint #goerr113
		Err:  goErr.New(r.Message),
		Type: gin.ErrorTypeAny,
	})

	if len(r.err) > 0 {
		for _, e := range r.err {
			ctx.Errors = append(ctx.Errors, &gin.Error{
				//nolint #goerr113
				Err:  goErr.New(e.Error()),
				Type: gin.ErrorTypeAny,
			})
		}
	}

	r.GinTonicAbort(ctx, httpCode)
}
