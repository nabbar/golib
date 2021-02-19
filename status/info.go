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

type FctInfo func() (name string, release string, build string)

type InfoResponse struct {
	Name      string `json:"name"`
	Release   string `json:"release"`
	HashBuild string `json:"hash_build"`
	Mandatory bool   `json:"mandatory"`
}

func (i *InfoResponse) Clone() InfoResponse {
	return InfoResponse{
		Name:      i.Name,
		Release:   i.Release,
		HashBuild: i.HashBuild,
		Mandatory: i.Mandatory,
	}
}

type Info interface {
	Get(x *gin.Context) InfoResponse
	Clean()
	IsValid() bool
}

func NewInfo(fct FctInfo, mandatory bool, cacheDuration time.Duration) Info {
	return &info{
		f: fct,
		m: mandatory,
		c: nil,
		t: time.Time{},
		d: cacheDuration,
	}
}

type info struct {
	f FctInfo
	m bool
	c *InfoResponse
	t time.Time
	d time.Duration
}

func (i *info) Get(x *gin.Context) InfoResponse {
	if !i.IsValid() {
		var (
			name string
			vers string
			hash string
		)

		if i.f != nil {
			name, vers, hash = i.f()
		}

		i.c = &InfoResponse{
			Name:      name,
			Release:   vers,
			HashBuild: hash,
			Mandatory: i.m,
		}
		i.t = time.Now()
	}

	return i.c.Clone()
}

func (i *info) Clean() {
	i.c = nil
	i.t = time.Now()
}

func (i *info) IsValid() bool {
	if i.c == nil {
		return false
	} else if i.t.IsZero() {
		return false
	} else if time.Since(i.t) > i.d {
		return false
	}
	return true
}
