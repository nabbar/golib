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
	"crypto/hkdf"
	"crypto/mlkem"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"fmt"
	"io"
)

const (
	mlkem768Size  = 1184
	mlkem1024Size = 1568
)

type mldec interface {
	Decapsulate([]byte) ([]byte, error)
	Bytes() []byte
}

type mlenc interface {
	Encapsulate() ([]byte, []byte)
	Bytes() []byte
}

func (o *mdl) kemHandshake(a Algo, c Curves, h Hash, readFirst bool) ([]byte, error) {
	var (
		err error
		shr []byte
		key []byte
		slt []byte

		nfo = "session-encryption-key"
	)

	switch a {
	case AlgoAES256:
		slt = []byte("SecureChannel-MLKEM-AESGCM")
	case AlgoChaChaPoly:
		slt = []byte("SecureChannel-MLKEM-ChaChaPoly")
	default:
		return nil, ErrUnsupportedAlgo
	}

	if readFirst {
		shr, err = o.kemHandshakeRead(c)
	} else {
		shr, err = o.kemHandshakeWrite(c)
	}

	if err != nil {
		return nil, err
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

func (o *mdl) kemHandshakeRead(c Curves) ([]byte, error) {
	var (
		err error
		cpr mlenc
		key []byte
		txt []byte
	)

	switch c {
	case CurveMLKEM768:
		txt = make([]byte, mlkem768Size)
		if _, err = io.ReadFull(o.r, txt); err != nil {
			return nil, err
		}

		if cpr, err = mlkem.NewEncapsulationKey768(txt); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrHandshake, err)
		}

	case CurveMLKEM1024:
		txt = make([]byte, mlkem1024Size)
		if _, err = io.ReadFull(o.r, txt); err != nil {
			return nil, err
		}

		if cpr, err = mlkem.NewEncapsulationKey1024(txt); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrGenerateKey, err)
		}
	default:
		return nil, ErrUnsupportedCurve
	}

	key, txt = cpr.Encapsulate()
	if len(cpr.Bytes()) < 1 || len(key) < 1 {
		return nil, ErrInvalidKey
	}

	if _, err = o.w.Write(txt); err != nil {
		return nil, err
	}

	return key, nil
}

func (o *mdl) kemHandshakeWrite(c Curves) ([]byte, error) {
	var (
		err error
		cpr mldec
		key []byte
		pub []byte
		txt []byte
	)

	switch c {
	case CurveMLKEM768:
		var km7 *mlkem.DecapsulationKey768
		if km7, err = mlkem.GenerateKey768(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrGenerateKey, err)
		}

		pub = km7.EncapsulationKey().Bytes()
		cpr = km7
		txt = make([]byte, mlkem768Size)

	case CurveMLKEM1024:
		var km1 *mlkem.DecapsulationKey1024
		if km1, err = mlkem.GenerateKey1024(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrGenerateKey, err)
		}

		pub = km1.EncapsulationKey().Bytes()
		cpr = km1
		txt = make([]byte, mlkem1024Size)

	default:
		return nil, ErrUnsupportedCurve
	}

	if len(cpr.Bytes()) < 1 || len(pub) < 1 {
		return nil, ErrInvalidKey
	}

	if _, err = o.w.Write(pub); err != nil {
		return nil, err
	}

	if _, err = io.ReadFull(o.r, txt); err != nil {
		return nil, err
	}

	if key, err = cpr.Decapsulate(txt); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHandshake, err)
	}

	return key, nil
}
