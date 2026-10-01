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

// Options configures the OpenGPG instance. The algorithm, hash, and cipher
// are selected via the corresponding fields. Post-Quantum hybrid algorithms
// use V6 keys; classical algorithms such as RSA use V4 keys. The default
// hash is SHA-3/512, the default cipher is AES-256, and compression is
// always disabled.
//
// Rand provides a cryptographically secure random source. When nil, the Go
// runtime's crypto/rand.Reader is used by the underlying packet layer.
// Since Go 1.26, standard library calls (e.g., key generation) ignore Rand
// unless GODEBUG=cryptocustomrand=1 is set.
//
// Time is an optional clock override. It is useful in tests or when keys must
// be generated with a deterministic timestamp. If Time is nil, time.Now is
// used by the packet layer when creating new entities.
//
// KeyTime specifies the key lifetime in seconds. A value of zero means the
// key never expires (KeyLifetimeSecs is set to math.MaxUint32, approximately
// 136 years). Any positive value is written verbatim into the OpenPGP packet
// configuration as KeyLifetimeSecs, causing the generated key to expire that
// many seconds after creation. KeyTime is a duration, not an absolute Unix
// timestamp: a value of 1700000000 would produce a key that expires roughly
// 54 years after creation. See RFC 9580 section 5.2.3.6. If the value
// exceeds math.MaxUint32, the remaining seconds are clamped to the unsigned
// 32-bit maximum.
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
	// Hash is the default hash function to be used.
	// If zero, SHA-256 is used.
	Hash crypto.Hash
	// Cipher is the cipher to be used.
	// If zero, AES-256 is used.
	Cipher packet.CipherFunction
	// Algorithm is The public key algorithm to use - will always create a signing primary
	// key and encryption subkey.
	Algorithm packet.PublicKeyAlgorithm
	// RSABits is the number of bits in new RSA keys made with NewEntity.
	// If zero, then 4096 bit keys are created.
	RSABits int
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
		RSABits:                o.RSABits,
	}

	if o.Algorithm == packet.PubKeyAlgoRSA {
		cfg.V6Keys = false
	}

	if o.Cipher == 0 {
		cfg.DefaultCipher = packet.CipherAES256
	}

	if o.Hash == 0 {
		cfg.DefaultHash = crypto.SHA3_512
	}

	if o.RSABits < 1 {
		cfg.RSABits = 4096
		cfg.MinRSABits = 4095
	} else if o.RSABits < math.MaxUint16 {
		cfg.RSABits = o.RSABits
		cfg.MinRSABits = uint16(o.RSABits - 1)
	} else {
		cfg.RSABits = o.RSABits
		cfg.MinRSABits = math.MaxUint16
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
