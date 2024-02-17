/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package dns_mapper

import (
	"net"
	"strings"
)

type dp struct {
	fqdn []string
	port string
}

func newPartDetail(fqdn, port string) *dp {
	if net.ParseIP(fqdn) != nil {
		return nil
	}

	return &dp{
		fqdn: strings.Split(strings.TrimSpace(fqdn), "."),
		port: strings.TrimPrefix(strings.TrimSpace(port), "0"),
	}
}

func newPart(entry string) *dp {
	var (
		host string
		port = "*"
	)

	if h, p, e := net.SplitHostPort(entry); e != nil {
		const missingPort = "missing port in address"
		if en, k := e.(*net.AddrError); !k {
			host = entry
		} else if en.Err == missingPort {
			host = entry
		} else {
			return nil
		}
	} else {
		host = h
		port = p
	}

	if net.ParseIP(host) != nil {
		return nil
	} else {
		return &dp{
			fqdn: strings.Split(strings.TrimSpace(host), "."),
			port: strings.TrimPrefix(strings.TrimSpace(port), "0"),
		}
	}
}

func (o *dp) String() string {
	if o.port == "*" {
		return o.FQDN()
	} else {
		return o.FQDN() + ":" + o.port
	}
}

func (o *dp) FQDN() string {
	return strings.Join(o.fqdn, ".")
}

func (o *dp) FQDNRaw() []string {
	return o.fqdn
}

func (o *dp) FQDNWildcard() bool {
	return o.fqdn[0] == "*"
}

func (o *dp) FQDNMatch(fqdn []string) bool {
	if len(fqdn) != len(o.fqdn) {
		return false
	} else if o.FQDNEqual(fqdn) {
		return true
	} else if !o.FQDNWildcard() {
		return false
	}

	var fLen = len(o.fqdn) - 1

	for i := 0; i <= fLen; i++ {
		var idx = fLen - i

		if o.fqdn[idx] == "*" {
			continue
		} else if o.fqdn[idx] != fqdn[idx] {
			return false
		}
	}

	return true
}

func (o *dp) FQDNEqual(in []string) bool {
	if len(in) != len(o.fqdn) {
		return false
	}

	for idx := range o.fqdn {
		if !strings.EqualFold(o.fqdn[idx], in[idx]) {
			return false
		}
	}

	return true
}

func (o *dp) Port() string {
	if o.port != "*" {
		return o.port
	}

	return ""
}

func (o *dp) PortWildcard() bool {
	return strings.Contains(o.port, "*")
}

func (o *dp) PortMatch(port string) bool {
	if !o.PortWildcard() {
		return port == o.port
	} else if o.port == "*" {
		return true
	}

	var part = strings.SplitN(o.port, "*", 2)

	if part[0] == "" {
		return port[len(port)-len(part[1]):] == part[1]
	} else {
		return port[:len(part[0])] == part[0]
	}
}
