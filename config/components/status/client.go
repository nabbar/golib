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

package status

import (
	cfgtps "github.com/nabbar/golib/config/types"
	libvpr "github.com/nabbar/golib/viper"
)

// getKey retrieves the component key from the internal store.
// It returns an empty string if the key is not found or invalid.
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

// getFctVpr retrieves the Viper factory function from the internal store.
// It returns nil if the function is not found or invalid.
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

// getViper retrieves the Viper instance using the stored factory function.
// It returns nil if the factory function is missing or returns nil.
func (o *mod) getViper() libvpr.Viper {
	if f := o.getFctVpr(); f == nil {
		return nil
	} else if v := f(); v == nil {
		return nil
	} else {
		return v
	}
}

// getFct determines and returns the appropriate Before and After event callbacks.
// If the component is already started, it returns the Reload callbacks.
// Otherwise, it returns the Start callbacks.
func (o *mod) getFct() (cfgtps.FuncCptEvent, cfgtps.FuncCptEvent) {
	if o.IsStarted() {
		return o.getFctEvt(keyFctRelBef), o.getFctEvt(keyFctRelAft)
	} else {
		return o.getFctEvt(keyFctStaBef), o.getFctEvt(keyFctStaAft)
	}
}

// getFctEvt retrieves a specific event callback function by its key.
// It returns nil if the callback is not found or invalid.
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

// runFctEvt executes a given event callback function with the component instance.
// It returns nil if the function is nil, otherwise it returns the function's error result.
func (o *mod) runFctEvt(fct func(cpt cfgtps.Component) error) error {
	if fct != nil {
		return fct(o)
	}

	return nil
}

// run executes the component's main lifecycle logic (Start or Reload).
// It follows the sequence:
// 1. Determine appropriate callbacks (Start vs Reload).
// 2. Execute the 'Before' callback.
// 3. Execute the core logic (runCli).
// 4. Execute the 'After' callback.
// It returns the first error encountered in the sequence.
func (o *mod) run() error {
	fb, fa := o.getFct()

	if err := o.runFctEvt(fb); err != nil {
		return err
	} else if err = o.runCli(); err != nil {
		return err
	} else if err = o.runFctEvt(fa); err != nil {
		return err
	}

	return nil
}

// runCli performs the core configuration loading and application.
// It retrieves the configuration, validates it, marks the component as running,
// and applies the configuration to the underlying status object.
func (o *mod) runCli() error {
	if cfg, err := o._getConfig(); err != nil {
		return ErrorParamInvalid.Error(err)
	} else if cfg == nil {
		return ErrorParamInvalid.Error()
	} else {
		o.r.Store(true)
		o.s.SetConfig(*cfg)
		return nil
	}
}
