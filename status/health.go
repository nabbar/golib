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
	"sync"
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
		m:  sync.Mutex{},
		fh: health,
		fm: msg,
		c:  nil,
		t:  time.Time{},
		d:  cacheDuration,
	}
}

type status struct {
	m  sync.Mutex
	fh FctHealth
	fm FctMessage

	c *StatusResponse
	t time.Time
	d time.Duration
}

func (s *status) getInfo() (string, string) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.fm != nil {
		return s.fm()
	}

	return "", ""
}

func (s *status) getHealth() error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.fh != nil {
		return s.fh()
	}

	return nil
}

func (s *status) setCache(obj *StatusResponse) {
	s.m.Lock()
	defer s.m.Unlock()

	s.c = obj
	s.t = time.Now()
}

func (s *status) getCache() StatusResponse {
	s.m.Lock()
	defer s.m.Unlock()

	return s.c.Clone()
}

func (s *status) Get(x *gin.Context) StatusResponse {
	if !s.IsValid() {
		var (
			err   error
			msgOk string
			msgKO string
		)

		msgOk, msgKO = s.getInfo()
		err = s.getHealth()

		c := &StatusResponse{}

		if err != nil {
			c.Status = DefMessageKO
			c.Message = msgKO
			liblog.ErrorLevel.LogErrorCtx(liblog.DebugLevel, "get health status", err)
		} else {
			c.Status = DefMessageOK
			c.Message = msgOk
		}

		s.setCache(c)
	}

	return s.getCache()
}

func (s *status) Clean() {
	s.m.Lock()
	defer s.m.Unlock()

	s.c = nil
	s.t = time.Now()
}

func (s *status) IsValid() bool {
	s.m.Lock()
	defer s.m.Unlock()

	if s.c == nil {
		return false
	} else if s.t.IsZero() {
		return false
	} else if time.Since(s.t) > s.d {
		return false
	}

	return true
}
