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

package mldsa87ed

import (
	"context"
	"crypto"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/ioutils/mapCloser"
)

// Options configures the OpenGPG instance backed by the Post-Quantum
// ML-DSA87 + Ed25519 key algorithm (Mldsa87/Ed25519 hybrid KEM with SHA-3/512
// and AES-256 symmetric cipher).
//
// Rand provides a cryptographically secure random source. When nil, the Go
// runtime's crypto/rand.Reader is used by the underlying packet layer.
//
// Time is an optional clock override. It is useful in tests or when keys must
// be generated with a deterministic timestamp.
//
// KeyTime specifies the target key creation time. If KeyTime is in the future
// relative to the current clock (or Time, if provided), the OpenPGP packet
// configuration records a KeyLifetimeSecs so the key expires when that time is
// reached. If the remaining seconds exceed math.MaxUint32, the maximum value
// is clamped (approximately 136 years).
type Options struct {
	// Rand is the cryptographically secure random source for key generation.
	Rand io.Reader
	// Time is an optional clock override for deterministic timestamps.
	Time func() time.Time
	// KeyTime is the target key creation time controlling the key lifetime.
	KeyTime uint32
}

// New constructs and returns an opengpg.OpenGPG implementation backed by the
// ML-DSA87 + Ed25519 (Mldsa87/Ed25519) hybrid key algorithm. The returned
// instance uses SHA-3/512 as the default hash and AES-256 as the default
// symmetric cipher, with compression disabled.
//
// The created instance is in an empty state with no key loaded. Call Create
// to generate a new key pair or Load to import an existing one. The provided
// context is forwarded to a mapCloser.Closer that manages the lifecycle of all
// pipe ends created by streaming encryption/decryption operations.
func New(ctx context.Context, o Options) opengpg.OpenGPG {
	var cfg = &packet.Config{
		V6Keys:                 true,
		Rand:                   o.Rand,
		Time:                   o.Time,
		DefaultHash:            crypto.SHA3_512,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionNone,
		Algorithm:              packet.PubKeyAlgoMldsa87Ed448,
	}

	// If KeyTime is in the future, record a key lifetime so the generated key
	// self-expire when that instant is reached. The remaining seconds are
	// clamped to math.MaxUint32 to avoid overflow in the OpenPGP packet format.
	if o.KeyTime < 1 {
		cfg.KeyLifetimeSecs = math.MaxUint32
	} else {
		cfg.KeyLifetimeSecs = o.KeyTime
	}

	return &mod{
		o: cfg,
		i: nil,
		c: new(atomic.Bool),
		l: mapCloser.New(ctx),
		p: &sync.Pool{
			New: func() any {
				b := make([]byte, defBufferSize)
				return &b
			},
		},
	}
}
