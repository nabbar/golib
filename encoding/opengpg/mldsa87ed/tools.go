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
	"os"

	sdkpgp "github.com/ProtonMail/go-crypto/openpgp"
)

// parseKeyRing parses serialized OpenPGP key material from a byte slice.
// It tries two decoding paths in order:
//
//  1. ASCII-armored format — calls ReadArmoredKeyRing, which expects the PGP
//     armor header (e.g., "-----BEGIN PGP PRIVATE KEY BLOCK-----"). Many keys
//     in production are stored in this human-readable format.
//
//  2. Binary sub-packet format — falls back to ReadKeyRing if the armoring
//     fails, which can read raw OpenPGP packet streams produced by
//     Serialize/SerializePrivate without armoring.
//
// If neither path yields at least one entity, os.ErrInvalid is returned. An
// empty input buffer yields os.ErrNotExist.
func parseKeyRing(b []byte) (sdkpgp.EntityList, error) {
	if len(b) == 0 {
		return nil, os.ErrNotExist
	}

	el, err := sdkpgp.ReadArmoredKeyRing(bytes.NewReader(b))
	if err == nil && len(el) > 0 {
		return el, nil
	}

	el, err = sdkpgp.ReadKeyRing(bytes.NewReader(b))
	if err == nil && len(el) > 0 {
		return el, nil
	}

	return nil, os.ErrInvalid
}
