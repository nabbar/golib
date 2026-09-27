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
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	sdkpgp "github.com/ProtonMail/go-crypto/openpgp"
	sdkamr "github.com/ProtonMail/go-crypto/openpgp/armor"
	sdkpck "github.com/ProtonMail/go-crypto/openpgp/packet"
	libgpg "github.com/nabbar/golib/encoding/opengpg"
	iotclo "github.com/nabbar/golib/ioutils/mapCloser"
)

const (
	defBufferSize = 32 * 1024
	chkStringBase = "OpenPGP ML-DSA87 + ED25519 verification payload"
)

// Ensure writer is closed before closing channel
type waitCloser struct {
	io.WriteCloser
	done <-chan struct{}
}

func (w *waitCloser) Close() error {
	// Closing writer
	err := w.WriteCloser.Close()

	// Lock closing until channel closing is call
	<-w.done

	return err
}

// mod is the concrete implementation of opengpg.OpenGPG backed by the
// DSA44/Ed-KEM512X (Mldsa65/Ed25519) hybrid key algorithm.
//
// Each field serves a specific concern:
//   - o: the packet.Config encoding settings (hash, cipher, algorithm, etc.)
//   - i: the loaded or created entity list holding the public and private keys
//   - c: an atomic flag, once set to true by Close(), disables all further
//     operations and returns os.ErrClosed to the caller
//   - l: a mapCloser.Closer that tracks all ephemeral resources (pipe ends)
//     created during streaming encryption/decryption so they can be cleaned up
//     deterministically when Close() is invoked
type mod struct {
	o *sdkpck.Config    // encoding configuration (hash, cipher, algorithm, key lifetime)
	i sdkpgp.EntityList // loaded or created key pair entities
	c *atomic.Bool      // closed flag — once true, all operations return os.ErrClosed
	l iotclo.Closer     // tracks ephemeral resources for deterministic cleanup
	p *sync.Pool
}

// Create generates a new key pair from the supplied Identity and writes the
// resulting ASCII-armored key material back into Identity.PublicKey and
// Identity.PrivateKey.
//
// The function performs the following steps:
//  1. Validates that the instance is not closed and that Identity is non-nil.
//  2. Calls NewEntity to generate a fresh Mldsa65/Ed25519 key pair using the
//     stored packet.Config, populated by Options passed to New().
//  3. Serializes the entity to an ASCII-armored public key and stores it in
//     id.PublicKey.
//  4. Serializes the entity to an ASCII-armored private key and stores it in
//     id.PrivateKey.
//  5. Assigns the entity to the internal entity list (o.i).
//
// Any failure during entity generation or serialization returns a typed error
// code (ErrorIdentityInvalid, ErrorPublicKeyInvalid, or ErrorPrivateKeyInvalid)
// wrapping the underlying error for chained diagnostics.
func (o *mod) Create(id *libgpg.Identity) error {
	if o.c.Load() {
		return os.ErrClosed
	}

	if id == nil {
		return libgpg.ErrorIdentityInvalid.Error()
	}

	var (
		err error
		ent *sdkpgp.Entity
		buf = bytes.NewBuffer(make([]byte, 0, defBufferSize))
		wrt io.WriteCloser
	)

	// Ensure the armor writer is closed on any early return to avoid resource leaks.
	defer func() {
		if wrt != nil {
			_ = wrt.Close()
		}
	}()

	defer func() {
		if buf != nil {
			buf.Reset()
		}
	}()

	// Step 1: Generate a new Mldsa65/Ed25519 entity with the given identity info.
	ent, err = sdkpgp.NewEntity(id.Name, id.Comment, id.Email, o.o)
	if err != nil {
		return libgpg.ErrorIdentityInvalid.Error(err)
	}

	// Step 2: Serialize and ASCII-armor the public key. The armor header type
	// sdkpgp.PublicKeyType produces the "-----BEGIN PGP PUBLIC KEY BLOCK-----"
	// header recognized by standard OpenPGP tooling.
	wrt, err = sdkamr.Encode(buf, sdkpgp.PublicKeyType, nil)
	if err != nil {
		return libgpg.ErrorPublicKeyInvalid.Error(err)
	}

	if err = ent.Serialize(wrt); err != nil {
		return libgpg.ErrorPublicKeyInvalid.Error(err)
	}

	_ = wrt.Close()
	wrt = nil

	if o.c.Load() {
		return os.ErrClosed
	}

	id.PublicKey = make([]byte, buf.Len())
	copy(id.PublicKey[:], buf.Bytes())

	// clean buffer and recreate it for armored private key
	buf.Reset()
	buf = bytes.NewBuffer(make([]byte, 0, defBufferSize))

	// Step 3: Serialize and ASCII-armor the private key. SerializePrivate includes
	// the unencrypted private key packets (no passphrase protection in this path).
	wrt, err = sdkamr.Encode(buf, sdkpgp.PrivateKeyType, nil)
	if err != nil {
		return libgpg.ErrorPrivateKeyInvalid.Error(err)
	}

	if err = ent.SerializePrivate(wrt, o.o); err != nil {
		return libgpg.ErrorPrivateKeyInvalid.Error(err)
	}

	_ = wrt.Close()
	wrt = nil

	if o.c.Load() {
		return os.ErrClosed
	}

	id.PrivateKey = make([]byte, buf.Len())
	copy(id.PrivateKey[:], buf.Bytes())

	// Step 4: Store the generated entity for subsequent encryption/decryption.
	o.i = sdkpgp.EntityList{ent}

	return nil
}

// Load imports an existing key pair from the supplied Identity by parsing either
// the private key material (preferred) or the public key material.
//
// If both PublicKey and PrivateKey are present, PrivateKey is parsed first.
// The parseKeyRing helper tries ASCII-armored input first, then falls back to
// unarmored binary keyring format. If neither format yields at least one entity,
// the call returns a typed error.
//
// Once successfully loaded, the entity list is stored in o.i and subsequent
// encryption/decryption operations use these keys.
func (o *mod) Load(id *libgpg.Identity) error {
	if o.c.Load() {
		return os.ErrClosed
	}

	if id == nil {
		return libgpg.ErrorIdentityInvalid.Error()
	}

	var (
		err error
		etl sdkpgp.EntityList
	)

	if len(id.PrivateKey) > 0 {
		if etl, err = parseKeyRing(id.PrivateKey); err != nil {
			return libgpg.ErrorPrivateKeyInvalid.Error(err)
		} else if len(etl) < 1 {
			return libgpg.ErrorPrivateKeyInvalid.Error()
		}
	} else if len(id.PublicKey) > 0 {
		if etl, err = parseKeyRing(id.PublicKey); err != nil {
			return libgpg.ErrorPublicKeyInvalid.Error(err)
		} else if len(etl) < 1 {
			return libgpg.ErrorPublicKeyInvalid.Error()
		}
	} else {
		return libgpg.ErrorIdentityInvalid.Error()
	}

	o.i = etl
	return nil
}

// Check validates the key pair by performing a full encrypt-then-decrypt
// round-trip with a synthetic payload. The purpose is to confirm that the
// Mldsa65/Ed25519 key can both encrypt and decrypt without corruption.
//
// The test payload is 4 KiB of repeated "OpenPGP KEM512 verification payload"
// text to exercise the symmetric cipher (AES-256) with enough data for a
// meaningful cryptographic test. The function:
//  1. Encrypts the payload into a buffer using the loaded entity list.
//  2. Decrypts the resulting ciphertext back.
//  3. Compares the decrypted bytes against the original payload with bytes.Equal.
//
// Returns os.ErrInvalid if the round-trip does not produce an identical payload.
func (o *mod) Check() error {
	if o.c.Load() {
		return os.ErrClosed
	}

	if o.i == nil {
		return libgpg.ErrorIdentityInvalid.Error()
	}

	var (
		err error
		wrt io.WriteCloser
		rdr *sdkpgp.MessageDetails

		tst = []byte(strings.Repeat(chkStringBase, 4*1024))
		buf = bytes.NewBuffer(make([]byte, 0, len(tst)+defBufferSize))
		res = bytes.NewBuffer(make([]byte, 0, len(tst)))
	)

	defer func() {
		if wrt != nil {
			_ = wrt.Close()
		}
	}()

	defer func() {
		if buf != nil {
			buf.Reset()
		}
	}()

	defer func() {
		if res != nil {
			res.Reset()
		}
	}()

	// Encrypt the test payload into buf using the Mldsa65/Ed25519 key.
	if wrt, err = sdkpgp.Encrypt(buf, o.i, nil, nil, o.o); err != nil {
		return err
	}

	if _, err = wrt.Write(tst); err != nil {
		return err
	}

	if err = wrt.Close(); err != nil {
		return err
	}

	// Decrypt the ciphertext back and compare with the original payload.
	if rdr, err = sdkpgp.ReadMessage(buf, o.i, nil, o.o); err != nil {
		return err
	}

	if _, err = io.Copy(res, rdr.UnverifiedBody); err != nil {
		return err
	}

	if !bytes.Equal(tst, res.Bytes()) {
		return os.ErrInvalid
	}

	return nil
}

// EncryptReader returns a ReadCloser that transparently encrypts the plaintext
// read from r using the loaded Mldsa65/Ed25519 key. The encryption is performed
// in a background goroutine that writes directly to one end of an io.Pipe; the
// caller reads from the other end.
//
// The pipe ends are tracked by the mapCloser.Closer so they are cleaned up on
// Close(). Error handling propagates through the pipe: if the underlying
// encryption fails, or if io.Copy encounters an error, the error is written to
// the pipe with CloseWithError.
//
// This design is memory-efficient because data is streamed through the pipe
// rather than accumulated in memory.
func (o *mod) EncryptReader(r io.Reader) (io.ReadCloser, error) {
	if o.c.Load() {
		return nil, os.ErrClosed
	}

	if o.i == nil {
		return nil, libgpg.ErrorIdentityInvalid.Error()
	}

	pr, pw := io.Pipe()

	o.l.Add(pr, pw)

	go func() {
		wc, err := sdkpgp.Encrypt(pw, o.i, nil, nil, o.o)

		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		bPtr := o.p.Get().(*[]byte)

		// force cleaning to prevent access resident data
		defer func() {
			clear(*bPtr) // destroy resident data
			o.p.Put(bPtr)
		}()

		if _, err = io.CopyBuffer(wc, r, *bPtr); err != nil {
			_ = wc.Close()
			_ = pw.CloseWithError(err)
			return
		}

		if err = wc.Close(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		_ = pw.Close()
	}()

	return pr, nil
}

// EncryptWriter returns a WriteCloser that accepts plaintext and writes the
// OpenPGP-encrypted output to w using the loaded Mldsa65/Ed25519 key. Unlike
// EncryptReader, this path uses the underlying sdkpgp.Encrypt directly without
// a pipe — the returned WriteCloser is the OpenPGP encryption writer itself.
//
// The writer is tracked by mapCloser.Closer for deterministic cleanup. The
// caller is responsible for writing plaintext and calling Close() to finalize
// the encryption.
func (o *mod) EncryptWriter(w io.Writer) (io.WriteCloser, error) {
	if o.i == nil {
		return nil, libgpg.ErrorIdentityInvalid.Error()
	}

	wrt, err := sdkpgp.Encrypt(w, o.i, nil, nil, o.o)
	if err != nil {
		return nil, libgpg.ErrorPublicKeyInvalid.Error()
	}

	o.l.Add(wrt)

	return wrt, nil
}

// DecryptReader returns a ReadCloser that reads OpenPGP-encrypted data from r
// and yields the decrypted plaintext. The decryption uses the loaded Mldsa65/
// Ed25519 key for key unwrapping and the AES-256 symmetric cipher for the
// symmetric decryption portion.
//
// The function calls ReadMessage synchronously and returns the UnverifiedBody
// (no signature or authenticity verification). The returned reader is wrapped
// with io.NopCloser to satisfy the ReadCloser interface while the underlying
// reader does not have a Close method.
func (o *mod) DecryptReader(r io.Reader) (io.ReadCloser, error) {
	if o.i == nil {
		return nil, libgpg.ErrorIdentityInvalid.Error()
	}

	rdr, err := sdkpgp.ReadMessage(r, o.i, nil, o.o)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(rdr.UnverifiedBody), nil
}

// DecryptWriter returns a WriteCloser that accepts OpenPGP-encrypted data and
// writes the decrypted plaintext to w using the loaded key. The implementation
// uses a bidirectional pipe: the caller writes ciphertext into pw; a background
// goroutine reads from pr, decrypts via ReadMessage, and copies the UnverifiedBody
// to w.
//
// Errors during decryption are propagated through CloseWithError on the read
// end of the pipe, which surfaces to the caller as an io copy error on the
// write side (via io.Copy from pw).
func (o *mod) DecryptWriter(w io.Writer) (io.WriteCloser, error) {
	if o.i == nil {
		return nil, libgpg.ErrorIdentityInvalid.Error()
	}

	pr, pw := io.Pipe()
	o.l.Add(pw, pr)

	// synchronization channel
	done := make(chan struct{})

	go func() {
		defer close(done)

		rdr, err := sdkpgp.ReadMessage(pr, o.i, nil, o.o)
		if err != nil {
			_ = pr.CloseWithError(err)
			return
		}

		bPtr := o.p.Get().(*[]byte)

		// force cleaning to prevent access resident data
		defer func() {
			clear(*bPtr) // destroy resident data
			o.p.Put(bPtr)
		}()

		if _, err = io.CopyBuffer(w, rdr.UnverifiedBody, *bPtr); err != nil {
			_ = pr.CloseWithError(err)
			return
		}

		_ = pr.Close()
	}()

	// return writer wrapper instead of direct PieWriter to prevent race
	return &waitCloser{
		WriteCloser: pw,
		done:        done,
	}, nil
}

// Close terminates the OpenGPG instance by releasing all managed resources.
// The atomic closed flag is set first to prevent new operations from starting.
// Then the mapCloser.Closer is invoked to close all tracked ephemeral resources
// (pipe ends created by streaming encryption/decryption).
//
// After Close, any subsequent call to Create, Load, Check, Encrypt*, or
// Decrypt* returns os.ErrClosed. The internal entity list and config are nilled
// to prevent accidental reuse and to allow the garbage collector to reclaim
// the key material.
func (o *mod) Close() error {
	o.c.Store(true)
	_ = o.l.Close()

	o.i = nil
	o.o = nil

	return nil
}
