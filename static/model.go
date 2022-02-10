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
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
	liblog "github.com/nabbar/golib/logger"
	librtr "github.com/nabbar/golib/router"
	libsts "github.com/nabbar/golib/status"
)

const _textEmbed = "Embed FS"

type staticHandler struct {
	m sync.Mutex
	l func() liblog.Logger

	c embed.FS
	b []string
	z int64

	i *atomic.Value
	d *atomic.Value
	f *atomic.Value
	s *atomic.Value
	r *atomic.Value
}

func (s *staticHandler) _makeRoute(group, route string) string {
	if group == "" {
		group = "/"
	}
	return path.Join(group, route)
}

func (s *staticHandler) _IsInSlice(sl []string, val string) bool {
	for _, v := range sl {
		if v == val {
			return true
		}
	}

	return false
}

func (s *staticHandler) _getSize() int64 {
	s.m.Lock()
	defer s.m.Unlock()

	return s.z
}

func (s *staticHandler) _setSize(size int64) {
	s.m.Lock()
	defer s.m.Unlock()

	s.z = size
}

func (s *staticHandler) _getBase() []string {
	s.m.Lock()
	defer s.m.Unlock()

	return s.b
}

func (s *staticHandler) _setBase(base ...string) {
	s.m.Lock()
	defer s.m.Unlock()

	s.b = base
}

func (s *staticHandler) _setLogger(fct func() liblog.Logger) {
	s.m.Lock()
	defer s.m.Unlock()

	s.l = fct
}

func (s *staticHandler) _getLogger() liblog.Logger {
	s.m.Lock()
	defer s.m.Unlock()

	if s.l == nil {
		return liblog.GetDefault()
	} else if log := s.l(); log == nil {
		return liblog.GetDefault()
	} else {
		return log
	}
}

func (s *staticHandler) _getIndex() map[string][]string {
	s.m.Lock()
	defer s.m.Unlock()

	var def = make(map[string][]string, 0)
	if s.i == nil {
		return def
	}
	if i := s.i.Load(); i == nil {
		return def
	} else if o, ok := i.(map[string][]string); !ok {
		return def
	} else {
		return o
	}
}

func (s *staticHandler) _setIndex(val map[string][]string) {
	s.m.Lock()
	defer s.m.Unlock()

	if val == nil {
		val = make(map[string][]string, 0)
	}

	if s.i == nil {
		s.i = new(atomic.Value)
	}

	s.i.Store(val)
}

func (s *staticHandler) _getDownload() map[string]bool {
	s.m.Lock()
	defer s.m.Unlock()

	var def = make(map[string]bool, 0)
	if s.d == nil {
		return def
	}
	if i := s.d.Load(); i == nil {
		return def
	} else if o, ok := i.(map[string]bool); !ok {
		return def
	} else {
		return o
	}
}

func (s *staticHandler) _setDownload(val map[string]bool) {
	s.m.Lock()
	defer s.m.Unlock()

	if val == nil {
		val = make(map[string]bool, 0)
	}

	if s.d == nil {
		s.d = new(atomic.Value)
	}

	s.d.Store(val)
}

func (s *staticHandler) _getFollow() map[string]string {
	s.m.Lock()
	defer s.m.Unlock()

	var def = make(map[string]string, 0)
	if s.f == nil {
		return def
	}
	if i := s.f.Load(); i == nil {
		return def
	} else if o, ok := i.(map[string]string); !ok {
		return def
	} else {
		return o
	}
}

func (s *staticHandler) _setFollow(val map[string]string) {
	s.m.Lock()
	defer s.m.Unlock()

	if val == nil {
		val = make(map[string]string, 0)
	}

	if s.f == nil {
		s.f = new(atomic.Value)
	}

	s.f.Store(val)
}

func (s *staticHandler) _getSpecific() map[string]gin.HandlerFunc {
	s.m.Lock()
	defer s.m.Unlock()

	var def = make(map[string]gin.HandlerFunc, 0)
	if s.s == nil {
		return def
	}
	if i := s.s.Load(); i == nil {
		return def
	} else if o, ok := i.(map[string]gin.HandlerFunc); !ok {
		return def
	} else {
		return o
	}
}

func (s *staticHandler) _setSpecific(val map[string]gin.HandlerFunc) {
	s.m.Lock()
	defer s.m.Unlock()

	if val == nil {
		val = make(map[string]gin.HandlerFunc, 0)
	}

	if s.s == nil {
		s.s = new(atomic.Value)
	}

	s.s.Store(val)
}

func (s *staticHandler) _getRouter() []string {
	s.m.Lock()
	defer s.m.Unlock()

	var def = make([]string, 0)
	if s.r == nil {
		return def
	}
	if i := s.r.Load(); i == nil {
		return def
	} else if o, ok := i.([]string); !ok {
		return def
	} else {
		return o
	}
}

func (s *staticHandler) _setRouter(val []string) {
	s.m.Lock()
	defer s.m.Unlock()

	if val == nil {
		val = make([]string, 0)
	}

	if s.r == nil {
		s.r = new(atomic.Value)
	}

	s.r.Store(val)
}

func (s *staticHandler) _listEmbed(root string) ([]fs.DirEntry, liberr.Error) {
	if root == "" {
		return nil, ErrorParamsEmpty.ErrorParent(fmt.Errorf("pathfile is empty"))
	}

	s.m.Lock()
	defer s.m.Unlock()

	val, err := s.c.ReadDir(root)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.ErrorParent(err)
	} else if err != nil {
		return nil, ErrorFileOpen.ErrorParent(err)
	} else {
		return val, nil
	}
}

func (s *staticHandler) _fileGet(pathFile string) (fs.FileInfo, io.ReadCloser, liberr.Error) {
	if pathFile == "" {
		return nil, nil, ErrorParamsEmpty.ErrorParent(fmt.Errorf("pathfile is empty"))
	}

	if inf, err := s._fileInfo(pathFile); err != nil {
		return nil, nil, err
	} else if inf.Size() >= s._getSize() {
		r, e := s._fileTemp(pathFile)
		return inf, r, e
	} else {
		r, e := s._fileBuff(pathFile)
		return inf, r, e
	}
}

func (s *staticHandler) _fileInfo(pathFile string) (fs.FileInfo, liberr.Error) {
	if pathFile == "" {
		return nil, ErrorParamsEmpty.ErrorParent(fmt.Errorf("pathfile is empty"))
	}

	s.m.Lock()
	defer s.m.Unlock()

	var inf fs.FileInfo
	obj, err := s.c.Open(pathFile)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.ErrorParent(err)
	} else if err != nil {
		return nil, ErrorFileOpen.ErrorParent(err)
	}

	defer func() {
		_ = obj.Close()
	}()

	inf, err = obj.Stat()

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.ErrorParent(err)
	} else if err != nil {
		return nil, ErrorFileOpen.ErrorParent(err)
	}

	return inf, nil
}

func (s *staticHandler) _fileBuff(pathFile string) (io.ReadCloser, liberr.Error) {
	if pathFile == "" {
		return nil, ErrorParamsEmpty.ErrorParent(fmt.Errorf("pathfile is empty"))
	}

	s.m.Lock()
	defer s.m.Unlock()

	obj, err := s.c.ReadFile(pathFile)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.ErrorParent(err)
	} else if err != nil {
		return nil, ErrorFileOpen.ErrorParent(err)
	} else {
		return libiot.NewBufferReadCloser(bytes.NewBuffer(obj)), nil
	}
}

func (s *staticHandler) _fileTemp(pathFile string) (libiot.FileProgress, liberr.Error) {
	if pathFile == "" {
		return nil, ErrorParamsEmpty.ErrorParent(fmt.Errorf("pathfile is empty"))
	}

	s.m.Lock()
	defer s.m.Unlock()

	var tmp libiot.FileProgress
	obj, err := s.c.Open(pathFile)

	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrorFileNotFound.ErrorParent(err)
	} else if err != nil {
		return nil, ErrorFileOpen.ErrorParent(err)
	}

	defer func() {
		_ = obj.Close()
	}()

	tmp, err = libiot.NewFileProgressTemp()
	if err != nil {
		return nil, ErrorFiletemp.ErrorParent(err)
	}

	_, e := io.Copy(tmp, obj)
	if e != nil {
		return nil, ErrorFiletemp.ErrorParent(e)
	}

	return tmp, nil
}

func (s *staticHandler) RegisterRouter(route string, register librtr.RegisterRouter, router ...gin.HandlerFunc) {
	s._setRouter(append(s._getRouter(), s._makeRoute("/", route)))

	router = append(router, s.Get)
	register(http.MethodGet, path.Join(route, "/*file"), router...)
}

func (s *staticHandler) RegisterRouterInGroup(route, group string, register librtr.RegisterRouterInGroup, router ...gin.HandlerFunc) {
	s._setRouter(append(s._getRouter(), s._makeRoute(group, route)))

	router = append(router, s.Get)
	register(group, http.MethodGet, path.Join(route, "/*file"), router...)
}

func (s *staticHandler) RegisterLogger(log func() liblog.Logger) {
	s._setLogger(log)
}

func (s *staticHandler) SetDownload(pathFile string, flag bool) {
	if pathFile != "" && s.Has(pathFile) {
		obj := s._getDownload()
		obj[pathFile] = flag
		s._setDownload(obj)
	}
}

func (s *staticHandler) SetIndex(group, route, pathFile string) {
	if pathFile != "" && s.Has(pathFile) {
		obj := s._getIndex()

		if obj[pathFile] == nil {
			obj[pathFile] = make([]string, 0)
		}

		obj[pathFile] = append(obj[pathFile], s._makeRoute(group, route))
		s._setIndex(obj)
	}
}

func (s *staticHandler) GetIndex(group, route string) string {
	route = s._makeRoute(group, route)

	for f, r := range s._getIndex() {
		if r == nil {
			continue
		}

		if s._IsInSlice(r, route) {
			return f
		}
	}

	return ""
}

func (s *staticHandler) SetRedirect(srcGroup, srcRoute, dstGroup, dstRoute string) {
	srcRoute = s._makeRoute(srcGroup, srcRoute)
	dstRoute = s._makeRoute(dstGroup, dstRoute)

	obj := s._getFollow()
	obj[srcRoute] = dstRoute
	s._setFollow(obj)
}

func (s *staticHandler) GetRedirect(srcGroup, srcRoute string) string {
	srcRoute = s._makeRoute(srcGroup, srcRoute)

	for src, dst := range s._getFollow() {
		if src == srcRoute {
			return dst
		}
	}

	return ""
}

func (s *staticHandler) SetSpecific(group, route string, router gin.HandlerFunc) {
	route = s._makeRoute(group, route)

	obj := s._getSpecific()
	obj[route] = router
	s._setSpecific(obj)
}

func (s *staticHandler) GetSpecific(group, route string) gin.HandlerFunc {
	route = s._makeRoute(group, route)

	for src, dst := range s._getSpecific() {
		if src == route {
			return dst
		}
	}

	return nil
}

func (s *staticHandler) IsDownload(pathFile string) bool {
	val, ok := s._getDownload()[pathFile]

	if !ok {
		return false
	}

	return val
}

func (s *staticHandler) IsIndex(pathFile string) bool {
	_, ok := s._getIndex()[pathFile]
	return ok
}

func (s *staticHandler) IsIndexForRoute(pathFile, group, route string) bool {
	val, ok := s._getIndex()[pathFile]

	if !ok {
		return false
	}

	return s._IsInSlice(val, s._makeRoute(group, route))
}

func (s *staticHandler) IsRedirect(group, route string) bool {
	_, ok := s._getFollow()[s._makeRoute(group, route)]
	return ok
}

func (s *staticHandler) Has(pathFile string) bool {
	if _, e := s._fileInfo(pathFile); e != nil {
		return false
	} else {
		return true
	}
}

func (s *staticHandler) List(rootPath string) ([]string, liberr.Error) {
	var (
		err error
		res = make([]string, 0)
		lst []string
		ent []fs.DirEntry
		inf fs.FileInfo
	)

	if rootPath == "" {
		for _, p := range s._getBase() {
			inf, err = s._fileInfo(p)
			if err != nil {
				return nil, err.(liberr.Error)
			}

			if !inf.IsDir() {
				res = append(res, p)
				continue
			}

			lst, err = s.List(p)

			if err != nil {
				return nil, err.(liberr.Error)
			}

			res = append(res, lst...)
		}
	} else if ent, err = s._listEmbed(rootPath); err != nil {
		return nil, err.(liberr.Error)
	} else {
		for _, f := range ent {

			if !f.IsDir() {
				res = append(res, path.Join(rootPath, f.Name()))
				continue
			}

			lst, err = s.List(path.Join(rootPath, f.Name()))

			if err != nil {
				return nil, err.(liberr.Error)
			}

			res = append(res, lst...)
		}
	}

	return res, nil
}

func (s *staticHandler) Find(pathFile string) (io.ReadCloser, liberr.Error) {
	_, r, e := s._fileGet(pathFile)
	return r, e
}

func (s *staticHandler) Info(pathFile string) (os.FileInfo, liberr.Error) {
	return s._fileInfo(pathFile)
}

func (s *staticHandler) Temp(pathFile string) (libiot.FileProgress, liberr.Error) {
	return s._fileTemp(pathFile)
}

func (s *staticHandler) Map(fct func(pathFile string, inf os.FileInfo) error) liberr.Error {
	var (
		err error
		lst []string
		inf fs.FileInfo
	)

	if lst, err = s.List(""); err != nil {
		return err.(liberr.Error)
	} else {
		for _, f := range lst {
			if inf, err = s._fileInfo(f); err != nil {
				return err.(liberr.Error)
			} else if err = fct(f, inf); err != nil {
				return err.(liberr.Error)
			}
		}
	}

	return nil
}

func (s *staticHandler) UseTempForFileSize(size int64) {
	s._setSize(size)
}

func (s *staticHandler) StatusInfo() (name string, release string, hash string) {
	return s._statusInfoPath("")
}

func (s *staticHandler) _statusInfoPath(pathFile string) (name string, release string, hash string) {
	vers := strings.TrimLeft(runtime.Version(), "go")
	vers = strings.TrimLeft(vers, "Go")
	vers = strings.TrimLeft(vers, "GO")

	if inf, err := s._fileInfo(pathFile); err != nil {
		return _textEmbed, vers, ""
	} else {
		return fmt.Sprintf("%s [%s]", _textEmbed, inf.Name()), vers, ""
	}
}

func (s *staticHandler) StatusHealth() error {
	for _, p := range s._getBase() {
		if _, err := s._fileInfo(p); err != nil {
			return err
		}
	}

	return nil
}

func (s *staticHandler) _statusHealthPath(pathFile string) error {
	if _, err := s._fileInfo(pathFile); err != nil {
		return err
	}

	return nil
}

func (s *staticHandler) _statusComponentPath(pathFile string, mandatory bool, message libsts.FctMessage, infoCacheTimeout, healthCacheTimeout time.Duration) libsts.Component {
	fctSts := func() (name string, release string, hash string) {
		return s._statusInfoPath(pathFile)
	}

	fctHlt := func() error {
		return s._statusHealthPath(pathFile)
	}

	return libsts.NewComponent(mandatory, fctSts, fctHlt, message, infoCacheTimeout, healthCacheTimeout)
}

func (s *staticHandler) StatusComponent(mandatory bool, message libsts.FctMessage, infoCacheTimeout, healthCacheTimeout time.Duration, sts libsts.RouteStatus) {
	for _, p := range s._getBase() {
		name := fmt.Sprintf("%s-%s", strings.ReplaceAll(_textEmbed, " ", "."), p)
		sts.ComponentNew(name, s._statusComponentPath(p, mandatory, message, infoCacheTimeout, healthCacheTimeout))
	}
}

func (s *staticHandler) Get(c *gin.Context) {
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
			if p == "/" {
				continue
			}
			calledFile = strings.TrimLeft(calledFile, p)
		}
	}

	calledFile = strings.Trim(calledFile, "/")

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

func (s *staticHandler) SendFile(c *gin.Context, filename string, size int64, isDownload bool, buf io.ReadCloser) {
	head := librtr.NewHeaders()

	if isDownload {
		head.Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(filename)))
	}

	c.Render(http.StatusOK, render.Reader{
		ContentLength: size,
		ContentType:   mime.TypeByExtension(path.Ext(filename)),
		Headers:       head.Header(),
		Reader:        buf,
	})
}
