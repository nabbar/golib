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
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

func (v Cipher) Check() bool {
	switch v {
	case TLS_RSA_WITH_AES_128_GCM_SHA256:
		return true
	case TLS_RSA_WITH_AES_256_GCM_SHA384:
		return true
	case TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return true
	case TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return true
	case TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return true
	case TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return true
	case TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
		return true
	case TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:
		return true
	case TLS_AES_128_GCM_SHA256:
		return true
	case TLS_AES_256_GCM_SHA384:
		return true
	case TLS_CHACHA20_POLY1305_SHA256:
		return true
	default:
		return false
	}
}

func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = Cipher(0)
			t string
			k bool
		)

		// Check if the data type matches the expected one
		if from.Kind() != reflect.String {
			return data, nil
		} else if t, k = data.(string); !k {
			return data, nil
		}

		// Check if the target type matches the expected one
		if reflect.TypeOf(z) != to {
			return data, nil
		}

		// Format/decode/parse the data and return the new value
		if e := z.unmarshall([]byte(t)); e != nil {
			return data, nil
		} else if z == Unknown {
			return data, nil
		} else {
			return z, nil
		}
	}
}
