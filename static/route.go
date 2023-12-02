/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package static

import (
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/nabbar/golib/router/header"

	ginsdk "github.com/gin-gonic/gin"
	ginrdr "github.com/gin-gonic/gin/render"
	liberr "github.com/nabbar/golib/errors"
	loglvl "github.com/nabbar/golib/logger/level"
	librtr "github.com/nabbar/golib/router"
	_ "github.com/ugorji/go/codec"
)

func (s *staticHandler) _makeRoute(group, route string) string {
	if group == "" {
		group = urlPathSeparator
	}
	return path.Join(group, route)
}

func (s *staticHandler) genRegisterRouter(route, group string, register any, router ...ginsdk.HandlerFunc) {
	var (
		ok  bool
		rte string
		reg librtr.RegisterRouter
		grp librtr.RegisterRouterInGroup
	)

	if register == nil {
		return
	} else if reg, ok = register.(librtr.RegisterRouter); ok {
		rte = s._makeRoute(urlPathSeparator, route)
		grp = nil
	} else if grp, ok = register.(librtr.RegisterRouterInGroup); ok {
		rte = s._makeRoute(group, route)
		reg = nil
	} else {
		return
	}

	if len(router) > 0 {
		router = append(router, s.Get)
	} else {
		router = append(make([]ginsdk.HandlerFunc, 0), s.Get)
	}

	if rtr := s._getRouter(); len(rtr) > 0 {
		s._setRouter(append(rtr, rte))
	} else {
		s._setRouter(append(make([]string, 0), rte))
	}

	if reg != nil {
		reg(http.MethodGet, path.Join(route, urlPathSeparator+"*file"), router...)
	}

	if grp != nil {
		grp(group, http.MethodGet, path.Join(route, urlPathSeparator+"*file"), router...)
	}
}

func (s *staticHandler) RegisterRouter(route string, register librtr.RegisterRouter, router ...ginsdk.HandlerFunc) {
	s.genRegisterRouter(route, "", register, router...)
}

func (s *staticHandler) RegisterRouterInGroup(route, group string, register librtr.RegisterRouterInGroup, router ...ginsdk.HandlerFunc) {
	s.genRegisterRouter(route, group, register, router...)
}

func (s *staticHandler) Get(c *ginsdk.Context) {
	calledFile := c.Request.URL.Path

	if dest := s.GetRedirect("", calledFile); dest != "" {
		url := c.Request.URL
		url.Path = dest

		c.Redirect(http.StatusPermanentRedirect, url.String())
		return
	}

	if router := s.GetSpecific("", calledFile); router != nil {
		router(c)
		return
	}

	if idx := s.GetIndex("", calledFile); idx != "" {
		calledFile = idx
	} else {
		for _, p := range s._getRouter() {
			if p == urlPathSeparator {
				continue
			}
			calledFile = strings.TrimLeft(calledFile, p)
		}
	}

	calledFile = strings.Trim(calledFile, urlPathSeparator)

	if !s.Has(calledFile) {
		for _, p := range s._getBase() {

			f := path.Join(p, calledFile)

			if s.Has(f) {
				calledFile = f
				break
			}
		}
	}

	if !s.Has(calledFile) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var (
		err liberr.Error
		buf io.ReadCloser
		inf fs.FileInfo
	)

	if inf, buf, err = s._fileGet(calledFile); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		ent := s._getLogger().Entry(loglvl.ErrorLevel, "get file info")
		ent.FieldAdd("filePath", calledFile)
		ent.FieldAdd("requestPath", c.Request.URL.Path)
		ent.ErrorAdd(true, err)
		ent.Log()
		if buf != nil {
			_ = buf.Close()
		}
		return
	}

	defer func() {
		if buf != nil {
			_ = buf.Close()
		}
	}()

	s.SendFile(c, calledFile, inf.Size(), s.IsDownload(calledFile), buf)
}

func (s *staticHandler) SendFile(c *ginsdk.Context, filename string, size int64, isDownload bool, buf io.ReadCloser) {
	head := header.NewHeaders()

	if isDownload {
		head.Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(filename)))
	}

	c.Render(http.StatusOK, ginrdr.Reader{
		ContentLength: size,
		ContentType:   mime.TypeByExtension(path.Ext(filename)),
		Headers:       head.Header(),
		Reader:        buf,
	})
}
