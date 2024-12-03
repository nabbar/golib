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

package curves

import (
	"strings"
)

func (v Cipher) String() string {
	return strings.Join(v.Code(), "_")
}

func (v Cipher) Code() []string {
	switch v {
	case TLS_RSA_WITH_AES_128_GCM_SHA256:
		return []string{"rsa", "aes", "128", "gcm", "sha256"}
	case TLS_RSA_WITH_AES_256_GCM_SHA384:
		return []string{"rsa", "aes", "256", "gcm", "sha384"}
	case TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return []string{"ecdhe", "rsa", "aes", "128", "gcm", "sha256"}
	case TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return []string{"ecdhe", "ecdsa", "aes", "128", "gcm", "sha256"}
	case TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return []string{"ecdhe", "rsa", "aes", "256", "gcm", "sha384"}
	case TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return []string{"ecdhe", "ecdsa", "aes", "256", "gcm", "sha384"}
	case TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
		return []string{"ecdhe", "rsa", "chacha20", "poly1305", "sha256"}
	case TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:
		return []string{"ecdhe", "ecdsa", "chacha20", "poly1305", "sha256"}
	case TLS_AES_128_GCM_SHA256:
		return []string{"aes", "128", "gcm", "sha256"}
	case TLS_AES_256_GCM_SHA384:
		return []string{"aes", "256", "gcm", "sha384"}
	case TLS_CHACHA20_POLY1305_SHA256:
		return []string{"chacha20", "poly1305", "sha256"}
	default:
		return []string{}
	}
}

func (v Cipher) Cipher() uint16 {
	return uint16(v)
}

func (v Cipher) TLS() uint16 {
	return v.Cipher()
}

func (v Cipher) Uint16() uint16 {
	return v.Cipher()
}

func (v Cipher) Uint() uint {
	return uint(v.Cipher())
}

func (v Cipher) Uint32() uint32 {
	return uint32(v.Cipher())
}

func (v Cipher) Uint64() uint64 {
	return uint64(v.Cipher())
}

func (v Cipher) Int() int {
	return int(v.Cipher())
}

func (v Cipher) Int32() int32 {
	return int32(v.Cipher())
}

func (v Cipher) Int64() int64 {
	return int64(v.Cipher())
}
