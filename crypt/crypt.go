/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package crypt

import (
	"io"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"

	. "github.com/nabbar/golib/errors"
)

var (
	cryptKey   = make([]byte, 32)
	cryptNonce = make([]byte, 12)
)

func SetKeyHex(key, nonce string) Error {
	var err error
	// Load your secret key from a safe place and reuse it across multiple
	// Seal/Open calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	cryptKey, err = hex.DecodeString(key)

	if err != nil {
		return HEXA_KEY.ErrorParent(err)
	}

	cryptNonce, err = hex.DecodeString(nonce)

	if err != nil {
		return HEXA_NONCE.ErrorParent(err)
	}

	return nil
}

func SetKeyByte(key [32]byte, nonce [12]byte) {
	cryptKey = key[:]
	cryptNonce = nonce[:]
}

func GenKeyByte() ([]byte, []byte, Error) {
	// Never use more than 2^32 random key with a given key because of the risk of a repeat.
	if _, err := io.ReadFull(rand.Reader, cryptKey); err != nil {
		return make([]byte, 32), make([]byte, 12), BYTE_KEYGEN.ErrorParent(err)
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	if _, err := io.ReadFull(rand.Reader, cryptNonce); err != nil {
		return make([]byte, 32), make([]byte, 12), BYTE_NONCEGEN.ErrorParent(err)
	}

	return cryptKey, cryptNonce, nil
}

func Encrypt(clearValue []byte) (string, Error) {
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	block, err := aes.NewCipher(cryptKey)
	if err != nil {
		return "", AES_BLOCK.ErrorParent(err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", AES_GCM.ErrorParent(err)
	}

	return hex.EncodeToString(aesgcm.Seal(nil, cryptNonce, clearValue, nil)), nil
}

func Decrypt(hexaVal string) ([]byte, Error) {
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	ciphertext, err := hex.DecodeString(hexaVal)
	if err != nil {
		return nil, HEXA_DECODE.ErrorParent(err)
	}

	block, err := aes.NewCipher(cryptKey)
	if err != nil {
		return nil, AES_BLOCK.ErrorParent(err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, AES_GCM.ErrorParent(err)
	}

	if res, err := aesgcm.Open(nil, cryptNonce, ciphertext, nil); err != nil {
		return res, AES_DECRYPT.ErrorParent(err)
	} else {
		return res, nil
	}
}
