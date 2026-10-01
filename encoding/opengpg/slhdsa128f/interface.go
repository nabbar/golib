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

package slhdsa128f

import (
	"context"
	"crypto"
	"io"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	libgpg "github.com/nabbar/golib/encoding/opengpg"
)

// Options configures the OpenGPG instance backed by the Post-Quantum
// SLH-DSA-SHAKE128f + Ed25519 key algorithm (sign: SLH-DSA-SHAKE128f + Ed25519, crypt: ML-KEM768 + X25519, hybrid KEM with SHA-3/512
// and AES-256 symmetric cipher).
type Options struct {
	// Rand provides the source of entropy.
	// If nil, the crypto/rand Reader is used.
	// Since Go 1.26, standard library calls (e.g., key generation) ignore Rand
	// unless GODEBUG=cryptocustomrand=1 is set.
	Rand io.Reader
	// Time returns the current time as the number of seconds since the
	// epoch. If Time is nil, time.Now is used.
	Time func() time.Time
	// KeyTime is The validity period of the key.  This is the number of seconds after
	// the key creation time that the key expires.  If this is not present
	// or has a value of zero, the key never expires.  This is found only on
	// a self-signature.
	// https://tools.ietf.org/html/rfc4880#section-5.2.3.6
	KeyTime uint32
}

func New(ctx context.Context, opt Options) libgpg.OpenGPG {
	return libgpg.New(ctx, libgpg.Options{
		Rand:      opt.Rand,
		Time:      opt.Time,
		KeyTime:   opt.KeyTime,
		Hash:      crypto.SHA3_512,
		Cipher:    packet.CipherAES256,
		Algorithm: packet.PubKeyAlgoSlhdsaShake128f, // sign: SLH-DSA-SHAKE128f + Ed25519, crypt: ML-KEM768 + X25519
	})
}
