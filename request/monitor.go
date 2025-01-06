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

package request

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	logent "github.com/nabbar/golib/logger/entry"
	loglvl "github.com/nabbar/golib/logger/level"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

const (
	defaultNameMonitor = "HTTP Request Client"
)

func (r *request) HealthCheck(ctx context.Context) error {
	var (
		opts = r.GetOption()
		ednp = r.GetFullUrl()
		head = make(http.Header)
		ent  logent.Entry
	)

	if log := r._getDefaultLogger(); log != nil {
		ent = log.Entry(loglvl.ErrorLevel, "healthcheck")
	}

	if !opts.Health.Enable {
		return nil
	}

	if opts.Health.Endpoint != "" {
		if u, e := url.Parse(opts.Health.Endpoint); e == nil {
			ednp = u
		}
	}

	if opts.Health.Auth.Basic.Enable {
		head.Set(_Authorization, _AuthorizationBasic+" "+base64.StdEncoding.EncodeToString([]byte(opts.Health.Auth.Basic.Username+":"+opts.Health.Auth.Basic.Password)))
	} else if opts.Health.Auth.Bearer.Enable {
		head.Set(_Authorization, _AuthorizationBearer+" "+opts.Health.Auth.Bearer.Token)
	}

	var (
		err error
		buf []byte
		req *http.Request
		rsp *http.Response
	)

	defer func() {
		if len(buf) > 0 {
			buf = buf[:0]
			buf = nil
		}
	}()

	if ent != nil {
		ent.FieldAdd("endpoint", ednp)
		ent.FieldAdd("method", http.MethodGet)
	}

	req, err = r.makeRequest(ctx, ednp, http.MethodGet, nil, head, nil)

	if err != nil {
		if ent != nil {
			ent.ErrorAdd(true, err).Log()
		}
		return err
	}

	rsp, err = r.client().Do(req)
	if err != nil {
		err = ErrorSendRequest.Error(err)
		if ent != nil {
			ent.ErrorAdd(true, err).Log()
		}
		return err
	}

	if buf, err = r.checkResponse(rsp); err != nil {
		if ent != nil {
			ent.ErrorAdd(true, err).Log()
		}
		return err
	}

	if len(opts.Health.Result.ValidHTTPCode) > 0 {
		if !r.isValidCode(opts.Health.Result.ValidHTTPCode, rsp.StatusCode) {
			err = ErrorResponseStatus.Error(fmt.Errorf("status: %s", rsp.Status))
			if ent != nil {
				ent.ErrorAdd(true, err).Log()
			}
			return err
		}
	} else if len(opts.Health.Result.InvalidHTTPCode) > 0 {
		if r.isValidCode(opts.Health.Result.InvalidHTTPCode, rsp.StatusCode) {
			err = ErrorResponseStatus.Error(fmt.Errorf("status: %s", rsp.Status))
			if ent != nil {
				ent.ErrorAdd(true, err).Log()
			}
			return err
		}
	}

	if len(opts.Health.Result.Contain) > 0 {
		if !r.isValidContents(opts.Health.Result.Contain, buf) {
			err = ErrorResponseContainsNotFound.Error(nil)
			if ent != nil {
				ent.ErrorAdd(true, err).Log()
			}
			return err
		}
	} else if len(opts.Health.Result.NotContain) > 0 {
		if r.isValidContents(opts.Health.Result.NotContain, buf) {
			err = ErrorResponseNotContainsFound.Error(nil)
			if ent != nil {
				ent.ErrorAdd(true, err).Log()
			}
			return err
		}
	}

	if ent != nil {
		ent.ErrorClean().Check(loglvl.InfoLevel)
	}

	return nil
}

func (r *request) Monitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error) {
	var (
		e   error
		inf moninf.Info
		mon montps.Monitor
		opt *Options
		res = make(map[string]interface{}, 0)
	)

	if opt = r.GetOption(); opt == nil {
		return nil, fmt.Errorf("cannot load options")
	}

	res["runtime"] = runtime.Version()[2:]
	res["release"] = vrs.GetRelease()
	res["build"] = vrs.GetBuild()
	res["date"] = vrs.GetDate()

	if inf, e = moninf.New(defaultNameMonitor); e != nil {
		return nil, e
	} else {
		inf.RegisterName(func() (string, error) {
			return fmt.Sprintf("%s [%s]", defaultNameMonitor, r.GetEndpoint()), nil
		})
		inf.RegisterInfo(func() (map[string]interface{}, error) {
			return res, nil
		})
	}

	if mon, e = libmon.New(r.getFuncContext(), inf); e != nil {
		return nil, e
	} else if mon == nil {
		return nil, nil
	}

	mon.RegisterLoggerDefault(r._getDefaultLogger)

	if e = mon.SetConfig(r.getFuncContext(), opt.Health.Monitor); e != nil {
		return nil, e
	}

	mon.SetHealthCheck(r.HealthCheck)

	if e = mon.Start(ctx); e != nil {
		return nil, e
	}

	return mon, nil
}
