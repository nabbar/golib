/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

// GetName returns the unique identifier name of the server instance.
// Falls back to the bind address if no name is configured.
func (o *srv) GetName() string {
	if i, l := o.c.Load(cfgName); !l {
		return o.GetBindable()
	} else if v, k := i.(string); !k {
		return o.GetBindable()
	} else {
		return v
	}
}

// GetBindable returns the local bind address (host:port) the server listens on.
// Returns empty string if no bind address is configured.
func (o *srv) GetBindable() string {
	if i, l := o.c.Load(cfgListen); !l {
		return ""
	} else if v, k := i.(*url.URL); !k {
		return ""
	} else {
		return v.Host
	}
}

// GetExpose returns the public-facing URL host:port used to access this server externally.
// Falls back to the bind address if no expose URL is configured.
func (o *srv) GetExpose() string {
	if i, l := o.c.Load(cfgExpose); !l {
		return o.GetBindable()
	} else if v, k := i.(*url.URL); !k {
		return o.GetBindable()
	} else {
		return v.Host
	}
}

// IsDisable returns true if the server is configured as disabled and should not start.
// Disabled servers maintain their configuration but do not accept connections.
func (o *srv) IsDisable() bool {
	if i, l := o.c.Load(cfgDisabled); !l {
		return false
	} else if v, k := i.(bool); !k {
		return false
	} else {
		return v
	}
}

// IsTLS returns true if the server is configured to use TLS/HTTPS.
// Checks both TLSMandatory flag and presence of valid certificate pairs.
func (o *srv) IsTLS() bool {
	if o.cfgTLSMandatory() {
		return true
	} else if s := o.cfgGetTLS(); s != nil && s.LenCertificatePair() > 0 {
		return true
	} else {
		return false
	}
}
