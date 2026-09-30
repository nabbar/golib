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

package mldsa65ed

import (
	"context"
	"crypto"
	"io"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	libgpg "github.com/nabbar/golib/encoding/opengpg"
)

// Options configures the OpenGPG instance backed by the Post-Quantum
// ML-DSA65 + Ed25519 key algorithm (ML-DSA65 + Ed25519 for sign, ML-KEM768 + X25519 for crypt, hybrid KEM with SHA-3/512
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

func New(ctx context.Context, opt Options) libgpg.OpenGPG {
	return libgpg.New(ctx, libgpg.Options{
		Rand:      opt.Rand,
		Time:      opt.Time,
		KeyTime:   opt.KeyTime,
		Hash:      crypto.SHA3_512,
		Cipher:    packet.CipherAES256,
		Algorithm: packet.PubKeyAlgoMldsa65Ed25519, // sign: ML-DSA65 + Ed25519, crypt: ML-KEM768 + X25519
	})
}
