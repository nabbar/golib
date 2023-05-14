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

	ginsdk "github.com/gin-gonic/gin"
	ginrdr "github.com/gin-gonic/gin/render"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	librtr "github.com/nabbar/golib/router"
	_ "github.com/ugorji/go/codec"
)

func (s *staticHandler) _makeRoute(group, route string) string {
	if group == "" {
		group = urlPathSeparator
	}
	return path.Join(group, route)
}

func (s *staticHandler) RegisterRouter(route string, register librtr.RegisterRouter, router ...ginsdk.HandlerFunc) {
	s._setRouter(append(s._getRouter(), s._makeRoute(urlPathSeparator, route)))

	router = append(router, s.Get)
	register(http.MethodGet, path.Join(route, urlPathSeparator+"*file"), router...)
}

func (s *staticHandler) RegisterRouterInGroup(route, group string, register librtr.RegisterRouterInGroup, router ...ginsdk.HandlerFunc) {
	s._setRouter(append(s._getRouter(), s._makeRoute(group, route)))

	router = append(router, s.Get)
	register(group, http.MethodGet, path.Join(route, urlPathSeparator+"*file"), router...)
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
		ent := s._getLogger().Entry(liblog.ErrorLevel, "get file info")
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
	head := librtr.NewHeaders()

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
