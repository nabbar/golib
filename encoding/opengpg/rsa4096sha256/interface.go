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

package rsa4096sha256

import (
	"context"
	"crypto"
	"io"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	libgpg "github.com/nabbar/golib/encoding/opengpg"
)

// Options configures the OpenGPG instance backed by the RSA-4096 key algorithm
// with SHA-256 as the default hash and AES-256 as the symmetric cipher.
// This preset uses OpenPGP V4 keys (V6Keys is disabled) and RSA-4096 for
// both the primary signing key and the encryption subkey.
type Options struct {
	// Rand provides a cryptographically secure random source. When nil, the Go
	// runtime's crypto/rand.Reader is used by the underlying packet layer.
	Rand io.Reader
	// Time is an optional clock override. It is useful in tests or when keys must
	// be generated with a deterministic timestamp. Provide a fixed clock to ensure
	// that generated keys are reproducible across test runs.
	Time func() time.Time
	// KeyTime specifies the key lifetime in seconds. A value of zero means the key
	// never expires (KeyLifetimeSecs is set to math.MaxUint32). Any positive value
	// is written verbatim into the OpenPGP packet config, causing the generated key
	// to expire that many seconds after creation. If the value exceeds math.MaxUint32
	// the lifetime is clamped to approximately 136 years. See RFC 4880 section 5.2.3.6.
	KeyTime uint32
}

// New constructs and returns an opengpg.OpenGPG implementation backed by
// RSA-4096 with SHA-256 hashing and AES-256 encryption. The resulting instance
// uses OpenPGP V4 keys (V6Keys is disabled because the ProtonMail/go-crypto
// library restricts post-quantum algorithms to V6). RSA-4096 provides classical
// symmetric-strength security with widely supported tooling.
func New(ctx context.Context, opt Options) libgpg.OpenGPG {
	return libgpg.New(ctx, libgpg.Options{
		Rand:      opt.Rand,
		Time:      opt.Time,
		KeyTime:   opt.KeyTime,
		Hash:      crypto.SHA256,
		Cipher:    packet.CipherAES256,
		Algorithm: packet.PubKeyAlgoRSA,
		RSABits:   4096,
	})
}
