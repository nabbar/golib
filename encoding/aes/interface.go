/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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
 *
 *
 */

// Package aes provides AES-256-GCM authenticated encryption with streaming I/O support.
//
// This package implements the encoding.Coder interface for consistent encryption/decryption
// operations. It uses AES-256 with Galois/Counter Mode (GCM) for authenticated encryption,
// providing both confidentiality and integrity.
//
// Features:
//   - AES-256-GCM authenticated encryption
//   - Cryptographically secure key and nonce generation
//   - Streaming encryption/decryption via io.Reader interfaces
//   - Hardware acceleration (AES-NI) on supported CPUs
//   - Memory efficient operations
//   - Thread-safe with separate coder instances
//
// Security specifications:
//   - Algorithm: AES-256 (256-bit key)
//   - Mode: GCM (Galois/Counter Mode)
//   - Nonce: 96 bits (12 bytes) - GCM standard
//   - Authentication tag: 128 bits (16 bytes)
//
// Example usage:
//
//	import encaes "github.com/nabbar/golib/encoding/aes"
//
//	// Generate key and nonce
//	key, _ := encaes.GenKey()
//	nonce, _ := encaes.GenNonce()
//
//	// Create coder
//	coder, _ := encaes.New(key, nonce)
//	defer coder.Reset()
//
//	// Encrypt
//	plaintext := []byte("Secret message")
//	ciphertext := coder.Encode(plaintext)
//
//	// Decrypt
//	decrypted, err := coder.Decode(ciphertext)
//	if err != nil {
//	    log.Fatal("Authentication failed")
//	}
//
// Security warnings:
//   - Never reuse a nonce with the same key (breaks GCM security)
//   - Always check authentication errors during decryption
//   - Store keys securely (never commit to version control)
//   - Rotate keys periodically for high-security applications
package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

	libenc "github.com/nabbar/golib/encoding"
)

// GetHexKey takes a hex encoded string and returns a 32 byte key.
// If the input string is less than 32 bytes, it will be zero-filled.
// If the input string is longer than 32 bytes, it will be truncated.
// If the input string is not a valid hex string, an error will be returned.
func GetHexKey(s string) ([32]byte, error) {
	var (
		err error
		dst []byte
		key [32]byte
	)

	if dst, err = hex.DecodeString(s); err != nil {
		return key, err
	}

	// Copy up to 32 bytes, handling short input gracefully
	n := len(dst)
	if n > 32 {
		n = 32
	}
	copy(key[:], dst[:n])
	return key, nil
}

// GenKey generates a cryptographically secure random 32-byte key.
//
// It reads from the default random number generator and copies the
// resulting bytes into the key.
//
// If there is an error while reading from the random number generator,
// this error will be returned.
//
// The generated key is suitable for use with the AES package.
func GenKey() ([32]byte, error) {
	var (
		slc = make([]byte, 32)
		key [32]byte
	)

	_, err := io.ReadFull(rand.Reader, slc)
	if err != nil {
		return key, err
	}

	copy(key[:], slc[:32])
	return key, nil
}

// GetHexNonce takes a hex encoded string and returns a 12-byte nonce.
// If the input string is less than 12 bytes, it will be zero-filled.
// If the input string is longer than 12 bytes, it will be truncated.
// If the input string is not a valid hex string, an error will be returned.
func GetHexNonce(s string) ([12]byte, error) {
	var (
		err error
		dst []byte
		non [12]byte
	)

	if dst, err = hex.DecodeString(s); err != nil {
		return non, err
	}

	// Copy up to 12 bytes, handling short input gracefully
	n := len(dst)
	if n > 12 {
		n = 12
	}
	copy(non[:], dst[:n])
	return non, nil
}

// GenNonce generates a cryptographically secure random 12-byte nonce.
//
// It reads from the default random number generator and copies the
// resulting bytes into the nonce.
//
// If there is an error while reading from the random number generator,
// this error will be returned.
//
// The generated nonce is suitable for use with the AES package.
func GenNonce() ([12]byte, error) {
	var (
		slc = make([]byte, 32)
		non [12]byte
	)

	_, err := io.ReadFull(rand.Reader, slc)
	if err != nil {
		return non, err
	}

	copy(non[:], slc[:12])
	return non, nil
}

// New creates a new AES coder instance.
//
// The key must be either 16 bytes (AES-128) or 32 (AES-256).
// The nonce must be 12 bytes.
//
// If the key or nonce is invalid, an error is returned.
// If there is an error while creating the AES cipher or
// the Galois/Counter Mode (GCM) cipher, this error is returned.
//
// The returned coder instance is suitable for use with the
// EncodeReader and DecodeReader functions.
func New(key [32]byte, nonce [12]byte) (libenc.Coder, error) {
	var (
		k = make([]byte, 32)
		n = make([]byte, 12)
	)
	copy(k[:], key[:])
	copy(n[:], nonce[:])

	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	blk, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}

	return &crt{
		a: gcm,
		n: n,
	}, nil
}
