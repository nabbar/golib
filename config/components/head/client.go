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

package head

import (
	cfgtps "github.com/nabbar/golib/config/types"
	liberr "github.com/nabbar/golib/errors"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvbr "github.com/spf13/viper"
)

func (o *componentHead) _getKey() string {
	o.m.RLock()
	defer o.m.RUnlock()

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

func (o *componentHead) _getFctVpr() libvpr.FuncViper {
	o.m.RLock()
	defer o.m.RUnlock()

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

func (o *componentHead) _getViper() libvpr.Viper {
	if f := o._getFctVpr(); f == nil {
		return nil
	} else if v := f(); v == nil {
		return nil
	} else {
		return v
	}
}

func (o *componentHead) _getSPFViper() *spfvbr.Viper {
	if f := o._getViper(); f == nil {
		return nil
	} else if v := f.Viper(); v == nil {
		return nil
	} else {
		return v
	}
}

func (o *componentHead) _getFctCpt() cfgtps.FuncCptGet {
	o.m.RLock()
	defer o.m.RUnlock()

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

func (o *componentHead) _getVersion() libver.Version {
	o.m.RLock()
	defer o.m.RUnlock()

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

func (o *componentHead) _getFct() (cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {
	if o.IsStarted() {
		return o._getFctEvt(keyFctRelBef), o._getFctEvt(keyFctRelAft)
	} else {
		return o._getFctEvt(keyFctStaBef), o._getFctEvt(keyFctStaAft)
	}
}

func (o *componentHead) _getFctEvt(key uint8) cfgtps.FuncCptEvent {
	o.m.RLock()
	defer o.m.RUnlock()

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

func (o *componentHead) _runFct(fct func(cpt cfgtps.Component) liberr.Error) liberr.Error {
	if fct != nil {
		return fct(o)
	}

	return nil
}

func (o *componentHead) _runCli() liberr.Error {
	if cfg, err := o._getConfig(); err != nil {
		return ErrorParamInvalid.Error(err)
	} else {
		o.m.Lock()
		defer o.m.Unlock()

		o.h = cfg.New()
		return nil
	}
}

func (o *componentHead) _run() liberr.Error {
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
