/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package static

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/gobuffalo/packr"

	. "github.com/nabbar/golib/errors"
	. "github.com/nabbar/golib/logger"

	"github.com/nabbar/golib/router"
)

const (
	FileIndex = "index.html"
)

type staticHandler struct {
	box      packr.Box
	debug    bool
	index    bool
	prefix   string
	download []string
	allDwnld bool
	head     gin.HandlerFunc
}

type Static interface {
	Register(register router.RegisterRouter)
	RegisterInGroup(group string, register router.RegisterRouterInGroup)

	SetDownloadAll()
	SetDownload(file string)
	IsDownload(file string) bool

	Has(file string) bool
	Find(file string) ([]byte, error)

	Health() Error
	Get(c *gin.Context)
}

func cleanPath(p string) string {
	return filepath.Clean(filepath.Join("/", strings.TrimLeft(p, "/")))
}

func cleanJoinPath(p, e string) string {
	return cleanPath(filepath.Join(strings.TrimLeft(p, "/"), e))
}

func NewStatic(hasIndex bool, prefix string, box packr.Box, Header gin.HandlerFunc) Static {
	return &staticHandler{
		box:      box,
		debug:    false,
		index:    hasIndex,
		prefix:   cleanPath(prefix),
		head:     Header,
		download: make([]string, 0),
	}
}

func (s staticHandler) Register(register router.RegisterRouter) {
	if s.prefix == "/" {
		for _, f := range s.box.List() {
			register(http.MethodGet, cleanJoinPath(s.prefix, f), s.Get)
		}
	} else {
		register(http.MethodGet, s.prefix, s.Get)
		register(http.MethodGet, cleanJoinPath(s.prefix, "/*file"), s.Get)
	}
}

func (s staticHandler) RegisterInGroup(group string, register router.RegisterRouterInGroup) {
	if s.prefix == "/" {
		for _, f := range s.box.List() {
			register(group, http.MethodGet, cleanJoinPath(s.prefix, f), s.head, s.Get)
		}
	} else {
		register(group, http.MethodGet, s.prefix, s.head, s.Get)
		register(group, http.MethodGet, cleanJoinPath(s.prefix, "/*file"), s.head, s.Get)
	}
}

func (s staticHandler) print() {
	if s.debug {
		return
	}

	for _, f := range s.box.List() {
		DebugLevel.Logf("Embedded file : %s", f)
	}

	s.debug = true
}

func (s staticHandler) Health() Error {
	s.print()

	if len(s.box.List()) < 1 {
		return EMPTY_PACKED.Error(nil)
	}

	if s.index && !s.box.Has("index.html") && !s.box.Has("index.htm") {
		return INDEX_NOT_FOUND.Error(nil)
	}

	return nil
}

func (s staticHandler) Has(file string) bool {
	return s.box.Has(file)
}

func (s staticHandler) Find(file string) ([]byte, error) {
	return s.box.Find(file)
}

func (s staticHandler) Get(c *gin.Context) {
	partPath := strings.SplitN(c.Request.URL.Path, s.prefix, 2)
	requestPath := partPath[1]

	requestPath = strings.TrimLeft(requestPath, "./")
	requestPath = strings.Trim(requestPath, "/")
	calledFile := filepath.Base(requestPath)

	if requestPath == "" || requestPath == "/" {
		if s.index {
			calledFile = FileIndex
			requestPath = FileIndex
		} else {
			_ = c.AbortWithError(http.StatusNotFound, INDEX_REQUESTED_NOT_SET.Error(nil))
			return
		}
	}

	if obj, err := s.box.Find(requestPath); !ErrorLevel.LogErrorCtxf(DebugLevel, "find file '%s' error for request '%s%s' :", err, calledFile, s.prefix, requestPath) {
		head := router.NewHeaders()

		if s.allDwnld || s.IsDownload(requestPath) {
			head.Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", calledFile))
		}

		c.Render(http.StatusOK, render.Reader{
			ContentLength: int64(len(obj)),
			ContentType:   mime.TypeByExtension(filepath.Ext(calledFile)),
			Headers:       head.Header(),
			Reader:        bytes.NewReader(obj),
		})
	} else {
		_ = c.AbortWithError(http.StatusNotFound, FILE_NOT_FOUND.Error(nil))
	}
}

func (s *staticHandler) SetDownload(file string) {
	s.download = append(s.download, file)
}

func (s *staticHandler) SetDownloadAll() {
	s.allDwnld = true
}

func (s staticHandler) IsDownload(file string) bool {
	for _, f := range s.download {
		if f == file {
			return true
		}
	}

	return false
}
