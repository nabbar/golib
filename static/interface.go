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
	"embed"
	"io"
	"os"
	"sync/atomic"

	libfpg "github.com/nabbar/golib/file/progress"

	ginsdk "github.com/gin-gonic/gin"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	librtr "github.com/nabbar/golib/router"
	libver "github.com/nabbar/golib/version"
)

type Static interface {
	RegisterRouter(route string, register librtr.RegisterRouter, router ...ginsdk.HandlerFunc)
	RegisterRouterInGroup(route, group string, register librtr.RegisterRouterInGroup, router ...ginsdk.HandlerFunc)

	RegisterLogger(log liblog.FuncLog)

	SetDownload(pathFile string, flag bool)
	SetIndex(group, route, pathFile string)
	GetIndex(group, route string) string
	SetRedirect(srcGroup, srcRoute, dstGroup, dstRoute string)
	GetRedirect(srcGroup, srcRoute string) string
	SetSpecific(group, route string, router ginsdk.HandlerFunc)
	GetSpecific(group, route string) ginsdk.HandlerFunc

	IsDownload(pathFile string) bool
	IsIndex(pathFile string) bool
	IsIndexForRoute(pathFile, group, route string) bool
	IsRedirect(group, route string) bool

	Has(pathFile string) bool
	List(rootPath string) ([]string, liberr.Error)
	Find(pathFile string) (io.ReadCloser, liberr.Error)
	Info(pathFile string) (os.FileInfo, liberr.Error)
	Temp(pathFile string) (libfpg.Progress, liberr.Error)

	Map(func(pathFile string, inf os.FileInfo) error) liberr.Error
	UseTempForFileSize(size int64)

	Monitor(ctx libctx.FuncContext, cfg montps.Config, vrs libver.Version) (montps.Monitor, error)

	Get(c *ginsdk.Context)
	SendFile(c *ginsdk.Context, filename string, size int64, isDownload bool, buf io.ReadCloser)
}

func New(ctx libctx.FuncContext, content embed.FS, embedRootDir ...string) Static {
	s := &staticHandler{
		l: new(atomic.Value),
		c: content,
		b: new(atomic.Value),
		z: new(atomic.Int64),
		i: libctx.NewConfig[string](ctx),
		d: libctx.NewConfig[string](ctx),
		f: libctx.NewConfig[string](ctx),
		s: libctx.NewConfig[string](ctx),
		r: new(atomic.Value),
		h: new(atomic.Value),
	}
	s._setBase(embedRootDir...)
	s._setLogger(nil)
	return s
}
