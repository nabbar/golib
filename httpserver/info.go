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

package httpserver

import "net/url"

func (o *srv) GetName() string {
	if i, l := o.c.Load(cfgName); !l {
		return o.GetBindable()
	} else if v, k := i.(string); !k {
		return o.GetBindable()
	} else {
		return v
	}
}

func (o *srv) GetBindable() string {
	if i, l := o.c.Load(cfgListen); !l {
		return ""
	} else if v, k := i.(*url.URL); !k {
		return ""
	} else {
		return v.Host
	}
}

func (o *srv) GetExpose() string {
	if i, l := o.c.Load(cfgExpose); !l {
		return o.GetBindable()
	} else if v, k := i.(*url.URL); !k {
		return o.GetBindable()
	} else {
		return v.Host
	}
}

func (o *srv) IsDisable() bool {
	if i, l := o.c.Load(cfgDisabled); !l {
		return false
	} else if v, k := i.(bool); !k {
		return false
	} else {
		return v
	}
}

func (o *srv) IsTLS() bool {
	if o.cfgTLSMandatory() {
		return true
	} else if s := o.cfgGetTLS(); s != nil && s.LenCertificatePair() > 0 {
		return true
	} else {
		return false
	}
}
