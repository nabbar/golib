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

package certificates

import (
	"bytes"
	"crypto/x509"
	"os"
	"runtime"
)

func SystemRootCA() *x509.CertPool {
	if runtime.GOOS == "windows" {
		return x509.NewCertPool()
	} else if c, e := x509.SystemCertPool(); e == nil {
		return c
	} else {
		return x509.NewCertPool()
	}
}

func checkFile(fct func(p []byte) error, pemFiles ...string) error {
	for _, f := range pemFiles {
		if f == "" {
			return ErrorParamEmpty.Error(nil)
		}

		if _, e := os.Stat(f); e != nil {
			return e
		}

		/* #nosec */
		b, e := os.ReadFile(f)
		if e != nil {
			return e
		}

		b = bytes.Trim(b, "\n")
		b = bytes.Trim(b, "\r")
		b = bytes.TrimSpace(b)

		if len(b) < 1 {
			return ErrorFileEmpty.Error(nil)
		} else if fct == nil {
			continue
		}

		if e = fct(b); e != nil {
			return e
		}
	}

	return nil
}
