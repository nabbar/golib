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
)

type CptResponse struct {
	InfoResponse
	StatusResponse
}

type Component interface {
	Get(x *gin.Context) CptResponse
	Clean()
}

func NewComponent(mandatory bool, info FctInfo, health FctHealth, msg FctMessage, infoCacheDuration, statusCacheDuration time.Duration) Component {
	return &cpt{
		i: NewInfo(info, mandatory, infoCacheDuration),
		s: NewStatus(health, msg, statusCacheDuration),
	}
}

type cpt struct {
	i Info
	s Status
}

func (c *cpt) Get(x *gin.Context) CptResponse {
	return CptResponse{
		InfoResponse:   c.i.Get(x),
		StatusResponse: c.s.Get(x),
	}
}

func (c *cpt) Clean() {
	c.i.Clean()
	c.s.Clean()
}
