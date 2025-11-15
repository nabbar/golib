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

package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	cpttls "github.com/nabbar/golib/config/components/tls"
	cfgtps "github.com/nabbar/golib/config/types"
	htpool "github.com/nabbar/golib/httpserver/pool"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

func (o *mod) _getKey() string {
	if i, l := o.x.Load(keyCptKey); !l {
		return ""
	} else if i == nil {
		return ""
	} else if v, k := i.(string); !k {
		return ""
	} else {
		return v
	}
}

func (o *mod) _getFctVpr() libvpr.FuncViper {
	if i, l := o.x.Load(keyFctViper); !l {
		return nil
	} else if i == nil {
		return nil
	} else if f, k := i.(libvpr.FuncViper); !k {
		return nil
	} else {
		return f
	}
}

func (o *mod) _getViper() libvpr.Viper {
	if f := o._getFctVpr(); f == nil {
		return nil
	} else if v := f(); v == nil {
		return nil
	} else {
		return v
	}
}

func (o *mod) _getFctCpt() cfgtps.FuncCptGet {
	if i, l := o.x.Load(keyFctGetCpt); !l {
		return nil
	} else if i == nil {
		return nil
	} else if f, k := i.(cfgtps.FuncCptGet); !k {
		return nil
	} else {
		return f
	}
}

func (o *mod) _getVersion() libver.Version {
	if i, l := o.x.Load(keyCptVersion); !l {
		return nil
	} else if i == nil {
		return nil
	} else if v, k := i.(libver.Version); !k {
		return nil
	} else {
		return v
	}
}

func (o *mod) _GetTLS() libtls.TLSConfig {
	if t := o.t.Load(); len(t) < 1 {
		return nil
	} else if i := cpttls.Load(o._getFctCpt(), t); i == nil {
		return nil
	} else if tls := i.GetTLS(); tls == nil {
		return nil
	} else {
		return tls
	}
}

func (o *mod) _GetHandler() map[string]http.Handler {
	if h := o.h.Load(); h == nil {
		return make(map[string]http.Handler)
	} else {
		return h()
	}
}

func (o *mod) _getFct() (cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {
	if o.IsStarted() {
		return o._getFctEvt(keyFctRelBef), o._getFctEvt(keyFctRelAft)
	} else {
		return o._getFctEvt(keyFctStaBef), o._getFctEvt(keyFctStaAft)
	}
}

func (o *mod) _getFctEvt(key uint8) cfgtps.FuncCptEvent {
	if i, l := o.x.Load(key); !l {
		return nil
	} else if i == nil {
		return nil
	} else if f, k := i.(cfgtps.FuncCptEvent); !k {
		return nil
	} else {
		return f
	}
}

func (o *mod) _runFct(fct func(cpt cfgtps.Component) error) error {
	if fct != nil {
		return fct(o)
	}

	return nil
}

func (o *mod) _runCli() error {
	var (
		e   error
		err error
		prt = ErrorComponentReload
		pol htpool.Pool
		cfg *htpool.Config
	)

	if !o.IsStarted() {
		prt = ErrorComponentStart
	}

	if cfg, err = o._getConfig(); err != nil {
		return prt.Error(err)
	}

	if pol, err = cfg.Pool(o.x, o._GetHandler, o.getLogger); err != nil {
		return prt.Error(err)
	} else if s := o.s.Load(); s == nil || s.Len() == 0 {
		o.s.Store(pol)
	} else if e = s.Merge(pol, o.getLogger); e != nil {
		return prt.Error(e)
	} else {
		o.s.Store(s)
	}

	s := o.s.Load()
	if s == nil || s.Len() == 0 {
		return prt.Error(ErrorComponentNotInitialized.Error())
	} else if s.IsRunning() {
		tm, cn := context.WithTimeout(o.x.GetContext(), 5*time.Second)
		defer cn()
		_ = s.Stop(tm)
	}

	tm, cn := context.WithTimeout(o.x.GetContext(), 5*time.Second)
	defer cn()
	e = s.Start(tm)

	if e != nil && errors.Is(e, tm.Err()) {
		return prt.Error(fmt.Errorf("timed out on starting server"))
	} else if e != nil {
		return prt.Error(e)
	}

	o.s.Store(s)

	return o._registerMonitor(prt)
}

func (o *mod) _run() error {
	fb, fa := o._getFct()

	if err := o._runFct(fb); err != nil {
		return err
	} else if err = o._runCli(); err != nil {
		return err
	} else if err = o._runFct(fa); err != nil {
		return err
	}

	return nil
}
