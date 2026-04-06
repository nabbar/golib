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

// Return is an interface for objects that can collect and format error information for API responses.
// It allows mapping the complex Error hierarchy into a flatter structure suitable for JSON serialization.
type Return interface {
	// SetError sets the primary error details for the response.
	// Parameters:
	// - code: Numeric error code.
	// - msg: Error message.
	// - file: File or function where the error occurred.
	// - line: Line number where the error occurred.
	SetError(code int, msg string, file string, line int)

	// AddParent adds a parent error's details to the response collection.
	AddParent(code int, msg string, file string, line int)

	// JSON returns the JSON-encoded representation of the error information.
	JSON() []byte
}

// ReturnGin extends the Return interface with specific methods for the Gin Gonic framework.
type ReturnGin interface {
	Return

	// GinTonicAbort aborts the current Gin request and sends the error information as a JSON response.
	// Parameters:
	// - ctx: The Gin context.
	// - httpCode: The HTTP status code to use for the response.
	GinTonicAbort(ctx *gin.Context, httpCode int)

	// GinTonicErrorAbort is similar to GinTonicAbort, but it also adds each error in the hierarchy
	// to the Gin context's Error slice before aborting.
	GinTonicErrorAbort(ctx *gin.Context, httpCode int)
}

// DefaultReturn is a basic implementation of the ReturnGin interface.
// It serializes into a simple JSON structure with Code and Message fields.
type DefaultReturn struct {
	ReturnGin `json:"-"` // We don't want to serialize the interface itself.

	Code    string  `json:"code"`    // Code is the string representation of the primary error code.
	Message string  `json:"message"` // Message is the primary error message.
	err     []error `json:"-"`       // err stores the full Error hierarchy for internal processing.
}

// SetError implements the Return interface. It sets the main code and message,
// and initializes the internal error slice with the provided details.
func (r *DefaultReturn) SetError(code int, msg string, file string, line int) {
	r.Code = fmt.Sprintf("%d", code)
	r.Message = msg

	if len(r.err) < 1 {
		r.err = make([]error, 0)
	}

	r.err = append(r.err, NewErrorTrace(code, msg, file, line, nil))
}

// AddParent implements the Return interface. It appends a parent error's details
// to the internal error slice.
func (r *DefaultReturn) AddParent(code int, msg string, file string, line int) {
	if len(r.err) < 1 {
		r.err = make([]error, 0)
	}

	r.err = append(r.err, NewErrorTrace(code, msg, file, line, nil))
}

// JSON returns the JSON-encoded representation of the DefaultReturn struct.
func (r *DefaultReturn) JSON() []byte {
	if str, err := json.Marshal(r); err != nil {
		return make([]byte, 0)
	} else {
		return str
	}
}

// GinTonicAbort aborts the Gin context with the specified HTTP code and the JSON representation of this struct.
// If httpCode is 0, it defaults to 500 Internal Server Error.
func (r *DefaultReturn) GinTonicAbort(ctx *gin.Context, httpCode int) {
	if ctx == nil || ctx.IsAborted() {
		return
	}

	if httpCode == 0 {
		httpCode = http.StatusInternalServerError
	}

	ctx.AbortWithStatusJSON(httpCode, r)
}

// GinTonicErrorAbort adds all errors in the hierarchy to the Gin context's error list before aborting.
func (r *DefaultReturn) GinTonicErrorAbort(ctx *gin.Context, httpCode int) {
	if ctx == nil || ctx.IsAborted() {
		return
	}

	// Add the main message first.
	ctx.Errors = append(ctx.Errors, &gin.Error{
		//nolint #goerr113
		Err:  goErr.New(r.Message),
		Type: gin.ErrorTypeAny,
	})

	// Add each individual error from the hierarchy.
	if len(r.err) > 0 {
		for _, e := range r.err {
			ctx.Errors = append(ctx.Errors, &gin.Error{
				//nolint #goerr113
				Err:  goErr.New(e.Error()),
				Type: gin.ErrorTypeAny,
			})
		}
	}

	// Abort with status.
	r.GinTonicAbort(ctx, httpCode)
}
