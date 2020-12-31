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
	return
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
		Err:  goErr.New(r.Message),
		Type: gin.ErrorTypeAny,
	})

	r.GinTonicAbort(ctx, httpCode)
}
