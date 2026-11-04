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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"golang.org/x/crypto/chacha20poly1305"
)

const maxPacketSize = 10 << 20 //10 485 760 = 10MB

var (
	ErrUnsupportedAlgo  = errors.New("unsupported algorithm")
	ErrUnsupportedCurve = errors.New("unsupported curve")
	ErrUnsupportedHash  = errors.New("unsupported hash")
	ErrInvalidKey       = errors.New("invalid generated key")
	ErrGenerateKey      = errors.New("generation key error")
	ErrHandshake        = errors.New("handshake error")
	ErrCipher           = errors.New("cipher error")
	ErrAESGCM           = errors.New("aes gcm error")
	ErrChachaPoly       = errors.New("chacha20-Poly1305 error")
	ErrPktOversize      = errors.New("packet oversizing error")
	ErrPktCorrupted     = errors.New("packet corrupted error, too small")
	ErrDeCrypt          = errors.New("decrypt error")
	ErrBufferSmall      = errors.New("buffer too small error")
	ErrHKDFDerive       = errors.New("HKDF derivation error")
)

type mdl struct {
	n func() io.Reader
	r io.Reader
	w io.Writer
	a cipher.AEAD
}

func (o *mdl) Reader() io.ReadCloser {
	return o
}

func (o *mdl) Writer() io.WriteCloser {
	return o
}

func (o *mdl) Read(p []byte) (n int, err error) {
	var (
		lng uint32
		lbf [4]byte
		pkt []byte
		txt []byte
		nsz = o.a.NonceSize()
	)

	if _, err = io.ReadFull(o.r, lbf[:]); err != nil {
		return 0, err
	}

	if lng = binary.BigEndian.Uint32(lbf[:]); lng > maxPacketSize {
		return 0, ErrPktOversize
	}

	pkt = make([]byte, lng)
	if _, err = io.ReadFull(o.r, pkt); err != nil {
		return 0, err
	}

	if len(pkt) < nsz {
		return 0, ErrPktCorrupted
	}

	if txt, err = o.a.Open(nil, pkt[:nsz], pkt[nsz:], nil); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrDeCrypt, err)
	}

	if len(txt) > len(p) {
		return 0, ErrBufferSmall
	}

	copy(p, txt)
	return len(txt), nil
}

func (o *mdl) Write(p []byte) (n int, err error) {
	var (
		nnc = make([]byte, o.a.NonceSize())
		txt []byte
		lng uint32
		pkt []byte
	)

	// TODO: limit send packet to maxPacketSize

	if _, err = o.rand().Read(nnc); err != nil {
		return 0, err
	}

	txt = o.a.Seal(nnc, nnc, p, nil)
	if i := len(txt); i < math.MaxUint32-4 {
		lng = uint32(i)
	} else {
		return 0, ErrPktOversize
	}

	pkt = make([]byte, 4+lng)
	binary.BigEndian.PutUint32(pkt[0:4], lng)
	copy(pkt[4:], txt)

	if _, err = o.w.Write(pkt); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (o *mdl) Close() error {
	if c, k := o.w.(io.Closer); k && c != nil {
		_ = c.Close()
	}
	if c, k := o.r.(io.Closer); k && c != nil {
		_ = c.Close()
	}
	return nil
}

func (o *mdl) rand() io.Reader {
	if o.n == nil {
		return rand.Reader
	}

	if r := o.n(); r != nil {
		return r
	}

	return rand.Reader
}

func (o *mdl) handshake(a Algo, c Curves, h Hash, readFirst bool) error {
	var (
		err error
		blk cipher.Block
		key []byte
	)

	switch c {
	case CurveX25519:
		key, err = o.oldHandshake(a, c, h, readFirst)
	default:
		key, err = o.kemHandshake(a, c, h, readFirst)
	}

	if err != nil {
		return err
	}

	switch a {
	case AlgoAES256:
		if blk, err = aes.NewCipher(key); err != nil {
			return fmt.Errorf("%w: %v", ErrCipher, err)
		}

		if o.a, err = cipher.NewGCM(blk); err != nil {
			return fmt.Errorf("%w: %v", ErrAESGCM, err)
		}

	case AlgoChaChaPoly:
		if o.a, err = chacha20poly1305.New(key); err != nil {
			return fmt.Errorf("%w: %v", ErrChachaPoly, err)
		}

	default:
		return ErrUnsupportedAlgo
	}

	return nil
}
