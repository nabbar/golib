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

package smtp

import (
	"crypto/tls"

	libtls "github.com/nabbar/golib/certificates"
	cpttls "github.com/nabbar/golib/config/components/tls"
	cfgtps "github.com/nabbar/golib/config/types"
	lbsmtp "github.com/nabbar/golib/mail/smtp"
	smtpcf "github.com/nabbar/golib/mail/smtp/config"
	moncfg "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvbr "github.com/spf13/viper"
)

func (o *mod) getKey() string {
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

func (o *mod) getFctVpr() libvpr.FuncViper {
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

func (o *mod) getViper() libvpr.Viper {
	if f := o.getFctVpr(); f == nil {
		return nil
	} else if v := f(); v == nil {
		return nil
	} else {
		return v
	}
}

func (o *mod) getSPFViper() *spfvbr.Viper {
	if f := o.getViper(); f == nil {
		return nil
	} else if v := f.Viper(); v == nil {
		return nil
	} else {
		return v
	}
}

func (o *mod) getFctCpt() cfgtps.FuncCptGet {
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

func (o *mod) getVersion() libver.Version {
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

func (o *mod) getTLS() libtls.TLSConfig {
	if t := o.t.Load(); len(t) < 1 {
		return nil
	} else if i := cpttls.Load(o.getFctCpt(), t); i != nil {
		return i.GetTLS()
	}
	return nil
}

func (o *mod) getTLSConfig(cfg libtls.Config) *tls.Config {
	if t := o.getTLS(); t == nil {
		return cfg.NewFrom(nil).TlsConfig("")
	} else {
		return cfg.NewFrom(t).TlsConfig("")
	}
}

func (o *mod) getFct() (cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {
	if o.IsStarted() {
		return o.getFctEvt(keyFctRelBef), o.getFctEvt(keyFctRelAft)
	} else {
		return o.getFctEvt(keyFctStaBef), o.getFctEvt(keyFctStaAft)
	}
}

func (o *mod) getFctEvt(key uint8) cfgtps.FuncCptEvent {
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

func (o *mod) runFct(fct func(cpt cfgtps.Component) error) error {
	if fct != nil {
		return fct(o)
	}

	return nil
}

func (o *mod) runCli() error {
	var (
		err error
		prt = ErrorComponentReload
		obj lbsmtp.SMTP
		cfg smtpcf.Config
		mon *moncfg.Config
	)

	if !o.IsStarted() {
		prt = ErrorComponentStart
	}

	if cfg, mon, err = o._getConfig(); err != nil {
		return prt.Error(err)
	} else if obj, err = lbsmtp.New(cfg, o.getTLSConfig(cfg.GetTls())); err != nil {
		return prt.Error(err)
	} else {
		if s := o.s.Load(); s != nil {
			s.Close()
		}

		o.s.Store(obj)
	}

	return o._registerMonitor(mon)
}

func (o *mod) run() error {
	fb, fa := o.getFct()

	if err := o.runFct(fb); err != nil {
		return err
	} else if err = o.runCli(); err != nil {
		return err
	} else if err = o.runFct(fa); err != nil {
		return err
	}

	return nil
}
