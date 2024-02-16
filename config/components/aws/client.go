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

package aws

import (
	libaws "github.com/nabbar/golib/aws"
	cfgtps "github.com/nabbar/golib/config/types"
	libreq "github.com/nabbar/golib/request"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvbr "github.com/spf13/viper"
)

func (o *componentAws) _getKey() string {
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

func (o *componentAws) _getFctVpr() libvpr.FuncViper {
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

func (o *componentAws) _getViper() libvpr.Viper {
	if f := o._getFctVpr(); f == nil {
		return nil
	} else if v := f(); v == nil {
		return nil
	} else {
		return v
	}
}

func (o *componentAws) _getSPFViper() *spfvbr.Viper {
	if f := o._getViper(); f == nil {
		return nil
	} else if v := f.Viper(); v == nil {
		return nil
	} else {
		return v
	}
}

func (o *componentAws) _getFctCpt() cfgtps.FuncCptGet {
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

func (o *componentAws) _getVersion() libver.Version {
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

func (o *componentAws) _getFct() (cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {
	if o.IsStarted() {
		return o._getFctEvt(keyFctRelBef), o._getFctEvt(keyFctRelAft)
	} else {
		return o._getFctEvt(keyFctStaBef), o._getFctEvt(keyFctStaAft)
	}
}

func (o *componentAws) _getFctEvt(key uint8) cfgtps.FuncCptEvent {
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

func (o *componentAws) _runFct(fct func(cpt cfgtps.Component) error) error {
	if fct != nil {
		return fct(o)
	}

	return nil
}

func (o *componentAws) _runCli() error {
	var (
		err error
		cli libaws.AWS
		cfg libaws.Config
		mon *libreq.OptionsHealth
		opt *libreq.Options
		req libreq.Request
		prt = ErrorComponentReload
	)

	if !o.IsStarted() {
		prt = ErrorComponentStart
	}

	if cfg, mon, err = o._getConfig(); err != nil {
		return prt.Error(err)
	}

	if cli, err = libaws.New(o.x.GetContext(), cfg, o.getClient()); err != nil {
		return prt.Error(err)
	}

	if mon != nil && mon.Enable {
		opt = &libreq.Options{
			Endpoint: "",
			Auth:     libreq.OptionsAuth{},
			Health:   *mon,
		}

		opt.SetDefaultLog(o.getLogger)

		if req = o.getRequest(); req != nil {
			if req, err = opt.Update(o.x.GetContext, req); err != nil {
				return prt.Error(err)
			}
			req.RegisterHTTPClient(o.getClient())
		} else if req, err = opt.New(o.x.GetContext, o.getClient()); err != nil {
			return prt.Error(err)
		}

		o.setRequest(req)
	}

	if mon != nil {
		if err = o._registerMonitor(mon, cfg); err != nil {
			return prt.Error(err)
		}
	}

	o.SetAws(cli)

	return nil
}

func (o *componentAws) _run() error {
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
