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

func (o *dmp) Len() int {
	var i int
	o.d.Range(func(key, value any) bool {
		i++
		return true
	})
	return i
}

func (o *dmp) Add(from, to string) {
	if d := newPart(from); d == nil {
		return
	} else {
		o.d.Store(d, to)
	}
}

func (o *dmp) Get(endpoint string) string {
	var (
		h, p, _ = o.Clean(endpoint)
		res     string
	)

	if p != "" {
		h = h + ":" + p
	}

	o.Walk(func(from, to string) bool {
		if from == h {
			res = to
			return false
		}

		return true
	})

	return res
}

func (o *dmp) Del(endpoint string) {
	var h, p, _ = o.Clean(endpoint)

	if p != "" {
		h = h + ":" + p
	}

	o.WalkDP(func(from *dp, to string) bool {
		if from.String() == h {
			o.d.Delete(from)
			return false
		}

		return true
	})
}

func (o *dmp) Walk(fct func(from, to string) bool) {
	o.d.Range(func(key, value any) bool {
		if d, l := key.(*dp); !l {
			return true
		} else if t, k := value.(string); !k {
			return true
		} else {
			return fct(d.String(), t)
		}
	})
}

func (o *dmp) WalkDP(fct func(from *dp, to string) bool) {
	o.d.Range(func(key, value any) bool {
		if d, l := key.(*dp); !l {
			return true
		} else if t, k := value.(string); !k {
			return true
		} else {
			return fct(d, t)
		}
	})
}

func (o *dmp) Clean(endpoint string) (host string, port string, err error) {
	host, port, err = net.SplitHostPort(endpoint)

	if err != nil {
		return strings.TrimSpace(endpoint), "", err
	}

	return strings.TrimSpace(host), strings.TrimPrefix(strings.TrimSpace(port), "0"), nil
}

func (o *dmp) Search(endpoint string) (string, error) {
	var (
		h, p, e = o.Clean(endpoint)
		src     *dp
		res     string
	)

	if e != nil {
		return "", e
	} else if src = newPartDetail(h, p); src == nil {
		return endpoint, nil
	}

	o.WalkDP(func(from *dp, to string) bool {
		if from.FQDNMatch(src.FQDNRaw()) {
			if from.PortMatch(src.Port()) {
				res = to

				if _, _, e = net.SplitHostPort(to); e != nil {
					res += ":" + src.Port()
				}

				return false
			}
		}

		return true
	})

	return res, nil
}

func (o *dmp) SearchWithCache(endpoint string) (string, error) {
	var (
		e error
		d string
	)

	if o.CacheHas(endpoint) {
		if d = o.CacheGet(endpoint); len(d) > 0 {
			return d, nil
		}
	}

	if d, e = o.Search(endpoint); e != nil {
		return "", e
	} else if len(d) > 0 {
		o.CacheSet(endpoint, d)
		return d, nil
	} else {
		o.CacheSet(endpoint, endpoint)
		return endpoint, nil
	}
}
