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

package njs_crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

var (
	cryptKey   = make([]byte, 32)
	cryptNonce = make([]byte, 12)
)

func SetKeyHex(key, nonce string) error {
	var err error
	// Load your secret key from a safe place and reuse it across multiple
	// Seal/Open calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	cryptKey, err = hex.DecodeString(key)
	if err != nil {
		return fmt.Errorf("converting hexa key error : %v", err)
	}

	cryptNonce, err = hex.DecodeString(nonce)
	if err != nil {
		return fmt.Errorf("converting hexa nonce error : %v", err)
	}

	return nil
}

func SetKeyByte(key [32]byte, nonce [12]byte) {
	cryptKey = key[:]
	cryptNonce = nonce[:]
}

func GenKeyByte() ([]byte, []byte, error) {
	// Never use more than 2^32 random key with a given key because of the risk of a repeat.
	if _, err := io.ReadFull(rand.Reader, cryptKey); err != nil {
		return make([]byte, 32), make([]byte, 12), fmt.Errorf("key generate error : %v", err)
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	if _, err := io.ReadFull(rand.Reader, cryptNonce); err != nil {
		return make([]byte, 32), make([]byte, 12), fmt.Errorf("nonce generate error : %v", err)
	}

	return cryptKey, cryptNonce, nil
}

func Encrypt(clearValue []byte) (string, error) {
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	block, err := aes.NewCipher(cryptKey)
	if err != nil {
		return "", fmt.Errorf("init AES block error : %v", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("AES GSM cipher init : %v", err)
	}

	return hex.EncodeToString(aesgcm.Seal(nil, cryptNonce, clearValue, nil)), nil
}

func Decrypt(hexaVal string) ([]byte, error) {
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	ciphertext, err := hex.DecodeString(hexaVal)
	if err != nil {
		return nil, fmt.Errorf("hexa decode crypted value error : %v", err)
	}

	block, err := aes.NewCipher(cryptKey)
	if err != nil {
		return nil, fmt.Errorf("AES block init error : %v", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AES GSM cipher init error : %v", err)
	}

	return aesgcm.Open(nil, cryptNonce, ciphertext, nil)
}
