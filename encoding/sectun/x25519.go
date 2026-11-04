/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package sectun

import (
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"fmt"
	"io"
)

func (o *mdl) oldHandshake(a Algo, _ Curves, h Hash, readFirst bool) ([]byte, error) {
	var (
		err error
		key []byte
		pub []byte
		txt []byte
		shr []byte
		slt []byte
		xch *ecdh.PublicKey

		nfo = "session-encryption-key"
	)

	switch a {
	case AlgoAES256:
		slt = []byte("SecureChannel-X25519-AESGCM")
	case AlgoChaChaPoly:
		slt = []byte("SecureChannel-X25519-ChaChaPoly")
	default:
		return nil, ErrUnsupportedAlgo
	}

	cv := ecdh.X25519()
	pv, e := cv.GenerateKey(o.n())
	if e != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerateKey, e)
	}

	key = pv.Bytes()
	pub = pv.PublicKey().Bytes()

	if len(key) < 1 || len(pub) < 1 {
		return nil, ErrInvalidKey
	}

	if readFirst {
		txt = make([]byte, len(pub))
		if _, err = io.ReadFull(o.r, txt); err != nil {
			return nil, err
		}
		xch, err = cv.NewPublicKey(txt)

		if _, err = o.w.Write(pub); err != nil {
			return nil, err
		}
	} else {
		if _, err = o.w.Write(pub); err != nil {
			return nil, err
		}

		txt = make([]byte, len(pub))
		if _, err = io.ReadFull(o.r, txt); err != nil {
			return nil, err
		}
		xch, err = cv.NewPublicKey(txt)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHandshake, err)
	}

	if shr, err = pv.ECDH(xch); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHandshake, err)
	}

	switch h {
	case HashSha2_256:
		if key, err = hkdf.Key(sha256.New, shr, slt, nfo, 32); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrHKDFDerive, err)
		}
	case HashSha2_512:
		if key, err = hkdf.Key(sha512.New, shr, slt, nfo, 32); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrHKDFDerive, err)
		}
	case HashSha3_256:
		if key, err = hkdf.Key(sha3.New256, shr, slt, nfo, 32); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrHKDFDerive, err)
		}
	case HashSha3_384:
		if key, err = hkdf.Key(sha3.New384, shr, slt, nfo, 32); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrHKDFDerive, err)
		}
	case HashSha3_512:
		if key, err = hkdf.Key(sha3.New512, shr, slt, nfo, 32); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrHKDFDerive, err)
		}
	default:
		return nil, ErrUnsupportedHash
	}

	return key, nil
}
