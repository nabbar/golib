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

package njs_static

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	njs_router "github.com/nabbar/golib/njs-router"

	"github.com/gin-gonic/gin/render"

	njs_logger "github.com/nabbar/golib/njs-logger"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
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
	head     func() map[string]string
}

type Static interface {
	Register(register njs_router.RegisterRouter)

	SetDownloadAll()
	SetDownload(file string)
	IsDownload(file string) bool
	Has(file string) bool
	Find(file string) ([]byte, error)

	Health() error
	Get(c *gin.Context)
}

func NewStatic(hasIndex bool, prefix string, box packr.Box, Header func() map[string]string) Static {
	return &staticHandler{
		box:      box,
		debug:    false,
		index:    hasIndex,
		prefix:   "/" + strings.Trim(prefix, "/"),
		head:     Header,
		download: make([]string, 0),
	}
}

func (s staticHandler) Register(register njs_router.RegisterRouter) {
	if s.prefix == "/" {
		for _, f := range s.box.List() {
			register(http.MethodGet, s.prefix+f, s.Get)
		}
	} else {
		register(http.MethodGet, s.prefix, s.Get)
		register(http.MethodGet, s.prefix+"/*file", s.Get)
	}
}

func (s staticHandler) print() {
	if s.debug {
		return
	}

	for _, f := range s.box.List() {
		njs_logger.DebugLevel.Logf("Embedded file : %s", f)
	}

	s.debug = true
}

func (s staticHandler) Health() error {
	s.print()

	if len(s.box.List()) < 1 {
		return fmt.Errorf("empty packed file stored")
	}

	if s.index && !s.box.Has("index.html") && !s.box.Has("index.htm") {
		return fmt.Errorf("cannot find 'index.html' file")
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
			c.Abort()
			return
		}
	}

	if obj, err := s.box.Find(requestPath); !njs_logger.ErrorLevel.LogErrorCtxf(njs_logger.NilLevel, "find file '%s' error for request '%s%s' :", err, calledFile, s.prefix, requestPath) {
		head := s.head()

		if s.allDwnld || s.IsDownload(requestPath) {
			head["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s\"", calledFile)
		}

		c.Render(http.StatusOK, render.Reader{
			ContentLength: int64(len(obj)),
			ContentType:   mime.TypeByExtension(filepath.Ext(calledFile)),
			Headers:       head,
			Reader:        bytes.NewReader(obj),
		})
	} else {
		c.Abort()
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
