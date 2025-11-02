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

package ldap

import (
	cfgtps "github.com/nabbar/golib/config/types"
	lbldap "github.com/nabbar/golib/ldap"
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
		cli *lbldap.HelperLDAP
		cfg *lbldap.Config
	)

	if cfg, err = o._getConfig(); err != nil {
		return ErrorParamInvalid.Error(err)
	} else if cli, e = lbldap.NewLDAP(o.x.GetContext(), cfg, o.GetAttributes()); e != nil {
		return ErrorConfigInvalid.Error(e)
	} else {
		cli.SetLogger(o.getLogger)
	}

	if i := o.l.Load(); i != nil {
		if v, k := i.(*lbldap.HelperLDAP); k {
			v.Close()
		}
	}

	o.SetConfig(cfg)
	o.SetLDAP(cli)

	return nil
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
