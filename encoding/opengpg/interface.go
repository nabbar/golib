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

package opengpg

import (
	"context"
	"crypto"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/nabbar/golib/ioutils/mapCloser"
)

// Identity represents an OpenPGP key pair with associated user identity information.
// The Name, Comment, and Email fields form the user ID string as defined by the
// OpenPGP specification. PublicKey and PrivateKey hold the serialized key material
// in ASCII-armored format after a Create or Load operation.
//
// The caller populates Name, Comment, and Email before invoking Create to generate
// a new key pair, or populates PublicKey and/or PrivateKey before invoking Load
// to import an existing key pair from serialized bytes.
type Identity struct {
	// Name is the user name portion of the OpenPGP user ID.
	Name string
	// Comment is the optional comment portion of the OpenPGP user ID.
	Comment string
	// Email is the email address portion of the OpenPGP user ID.
	Email string
	// PublicKey holds the ASCII-armored public key material after Create or Load.
	PublicKey []byte
	// PrivateKey holds the ASCII-armored private key material after Create or Load.
	PrivateKey []byte
}

// OpenGPG defines the contract for an OpenPGP encryption and decryption
// implementation. Any type satisfying this interface provides key management
// (Create/Load/Check) and streaming encryption/decryption via Reader and Writer
// adapters.
//
// Create generates a fresh key pair from the provided Identity and writes the
// resulting public and private keys back into the Identity struct.
//
// Load imports an existing key pair from serialized key material stored in the
// Identity struct. Either PublicKey, PrivateKey, or both may be provided.
//
// Check performs a round-trip self-test by encrypting and then decrypting a
// synthetic payload to verify that the loaded or created key pair is functional.
//
// EncryptReader and EncryptWriter provide OpenPGP hybrid encryption over
// arbitrary data streams. The Encrypt* variants encrypt data for the recipients
// identified by the loaded Identity.
//
// DecryptReader and DecryptWriter decrypt OpenPGP-encrypted data streams
// using the loaded Identity as the key holder. The UnverifiedBody is returned
// without signature or authenticity verification, matching the underlying
// ProtonMail/go-crypto behavior.
//
// Close terminates the OpenGPG instance, releasing any managed resources such
// as pipe ends or temporary buffers. After Close, further operations return
// os.ErrClosed.
type OpenGPG interface {
	io.Closer

	// Create generates a new key pair from the given Identity fields and populates
	// Identity.PublicKey and Identity.PrivateKey with the serialized key material.
	Create(*Identity) error
	// Load imports an existing key pair from Identity.PublicKey or Identity.PrivateKey.
	Load(*Identity) error
	// Check validates the key pair by performing an encrypt-then-decrypt round-trip.
	Check() error

	// EncryptReader returns a ReadCloser that reads plaintext from r and emits
	// OpenPGP-encrypted bytes.
	EncryptReader(r io.Reader) (io.ReadCloser, error)
	// EncryptWriter returns a WriteCloser that accepts plaintext and writes
	// OpenPGP-encrypted bytes to w.
	EncryptWriter(w io.Writer) (io.WriteCloser, error)

	// DecryptReader returns a ReadCloser that reads OpenPGP-encrypted data from r
	// and emits the decrypted plaintext.
	DecryptReader(r io.Reader) (io.ReadCloser, error)
	// DecryptWriter returns a WriteCloser that accepts OpenPGP-encrypted data and
	// writes the decrypted plaintext to w.
	DecryptWriter(w io.Writer) (io.WriteCloser, error)
}

// Options configures the OpenGPG instance backed by the Post-Quantum
// ML-DSA65 + Ed25519 key algorithm (ML-DSA65/Ed25519 hybrid KEM with SHA-3/512
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
	KeyTime   uint32
	Hash      crypto.Hash
	Cipher    packet.CipherFunction
	Algorithm packet.PublicKeyAlgorithm
}

// New constructs and returns an opengpg.OpenGPG implementation backed by the
// ML-DSA65 + Ed25519 (ML-DSA65/Ed25519) hybrid key algorithm. The returned
// instance uses SHA-3/512 as the default hash and AES-256 as the default
// symmetric cipher, with compression disabled.
//
// The created instance is in an empty state with no key loaded. Call Create
// to generate a new key pair or Load to import an existing one. The provided
// context is forwarded to a mapCloser.Closer that manages the lifecycle of all
// pipe ends created by streaming encryption/decryption operations.
func New(ctx context.Context, o Options) OpenGPG {
	var cfg = &packet.Config{
		V6Keys:                 true,
		Rand:                   o.Rand,
		Time:                   o.Time,
		DefaultHash:            o.Hash,
		DefaultCipher:          o.Cipher,
		Algorithm:              o.Algorithm,
		DefaultCompressionAlgo: packet.CompressionNone,
	}

	// If KeyTime is in the future, record a key lifetime so the generated key
	// expires when that instant is reached. The remaining seconds are
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
