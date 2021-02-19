/*
 * MIT License
 *
 * Copyright (c) 2021 Nicolas JUHEL
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

import (
	"time"

	"github.com/gin-gonic/gin"
	liblog "github.com/nabbar/golib/logger"
)

type FctHealth func() error
type FctMessage func() (msgOk string, msgKO string)

type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (s *StatusResponse) Clone() StatusResponse {
	return StatusResponse{
		Status:  s.Status,
		Message: s.Message,
	}
}

type Status interface {
	Get(x *gin.Context) StatusResponse
	Clean()
	IsValid() bool
}

func NewStatus(health FctHealth, msg FctMessage, cacheDuration time.Duration) Status {
	return &status{
		fh: health,
		fm: msg,
		c:  nil,
		t:  time.Time{},
		d:  cacheDuration,
	}
}

type status struct {
	fh FctHealth
	fm FctMessage

	c *StatusResponse
	t time.Time
	d time.Duration
}

func (s *status) Get(x *gin.Context) StatusResponse {
	if !s.IsValid() {
		var (
			err   error
			msgOk string
			msgKO string
		)

		if s.fm != nil {
			msgOk, msgKO = s.fm()
		}

		if s.fh != nil {
			err = s.fh()
		}

		c := &StatusResponse{}

		if err != nil {
			c.Status = statusKO
			c.Message = msgKO
			liblog.ErrorLevel.LogGinErrorCtx(liblog.DebugLevel, "get health status", err, x)
		} else {
			c.Status = statusOK
			c.Message = msgOk
		}

		s.c = c
		s.t = time.Now()
	}

	return s.c.Clone()
}

func (s *status) Clean() {
	s.c = nil
	s.t = time.Now()
}

func (s *status) IsValid() bool {
	if s.c == nil {
		return false
	} else if s.t.IsZero() {
		return false
	} else if time.Since(s.t) > s.d {
		return false
	}

	return true
}
