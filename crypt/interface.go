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

package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

type Crypt interface {
	Encode(p []byte) []byte
	Reader(r io.Reader) io.Reader

	EncodeHex(p []byte) []byte
	ReaderHex(r io.Reader) io.Reader

	Decode(p []byte) ([]byte, error)
	Writer(w io.Writer) io.Writer

	DecodeHex(p []byte) ([]byte, error)
	WriterHex(w io.Writer) io.Writer
}

func GetHexKey(s string) ([32]byte, error) {
	var (
		err error
		dst = make([]byte, 0)
		key [32]byte
	)

	if dst, err = hex.DecodeString(s); err != nil {
		return key, err
	}

	copy(key[:], dst[:32])
	return key, nil
}

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

func GetHexNonce(s string) ([12]byte, error) {
	var (
		err error
		dst = make([]byte, 0)
		non [12]byte
	)

	if dst, err = hex.DecodeString(s); err != nil {
		return non, err
	}

	copy(non[:], dst[:12])
	return non, nil
}

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

func New(key [32]byte, nonce [12]byte) (Crypt, error) {
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
