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

package opengpg_test

import (
	"bytes"
	"context"
	"crypto"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	libgpg "github.com/nabbar/golib/encoding/opengpg"
	gpgdsa65 "github.com/nabbar/golib/encoding/opengpg/mldsa65ed"
	gpgdsa87 "github.com/nabbar/golib/encoding/opengpg/mldsa87ed"
	gpgslh1f "github.com/nabbar/golib/encoding/opengpg/slhdsa128f"
	gpgslh1s "github.com/nabbar/golib/encoding/opengpg/slhdsa128s"
	gpgslh2s "github.com/nabbar/golib/encoding/opengpg/slhdsa256s"
)

const (
	suiteTimeout = 2 * time.Minute

	tstDataSize = 32 * 1024
	tstDataBase = "OpenPGP Hybrid verification payload"
)

var cpb = sync.Pool{
	New: func() any {
		b := make([]byte, tstDataSize)
		return &b
	},
}

func getNewOpenGPGP(ctx context.Context) libgpg.OpenGPG {
	return libgpg.New(ctx, libgpg.Options{
		Rand:      nil,
		Time:      nil,
		KeyTime:   0,
		Hash:      crypto.SHA3_512,
		Cipher:    packet.CipherAES256,
		Algorithm: packet.PubKeyAlgoSlhdsaShake256s,
	})
}

func getNewOpenGPGPMLDSA65(ctx context.Context) libgpg.OpenGPG {
	return gpgdsa65.New(ctx, gpgdsa65.Options{})
}

func getNewOpenGPGPMLDSA87(ctx context.Context) libgpg.OpenGPG {
	return gpgdsa87.New(ctx, gpgdsa87.Options{})
}

func getNewOpenGPGPSLHDSA128f(ctx context.Context) libgpg.OpenGPG {
	return gpgslh1f.New(ctx, gpgslh1f.Options{})
}

func getNewOpenGPGPSLHDSA128s(ctx context.Context) libgpg.OpenGPG {
	return gpgslh1s.New(ctx, gpgslh1s.Options{})
}

func getNewOpenGPGPSLHDSA256s(ctx context.Context) libgpg.OpenGPG {
	return gpgslh2s.New(ctx, gpgslh2s.Options{})
}

func readerOpenGPG(b *testing.B, mod libgpg.OpenGPG, buf, res *bytes.Buffer, pld []byte) {
	var (
		err error
		rdr io.ReadCloser
		ptr = cpb.Get().(*[]byte)
	)

	defer cpb.Put(ptr)

	// Encrypt
	if rdr, err = mod.EncryptReader(bytes.NewReader(pld)); err != nil {
		b.Fatalf("EncryptReader failed: %v", err)
	}

	if _, err = io.CopyBuffer(buf, rdr, *ptr); err != nil {
		b.Fatalf("encryption read failed: %v", err)
	}

	_ = rdr.Close()

	// Decrypt
	if rdr, err = mod.DecryptReader(buf); err != nil {
		b.Fatalf("DecryptReader failed: %v", err)
	}

	if _, err = io.CopyBuffer(res, rdr, *ptr); err != nil {
		b.Fatalf("decryption read failed: %v", err)
	}

	_ = rdr.Close()

	if !bytes.Equal(res.Bytes(), pld) {
		b.Fatalf("decryption read failed: got %v, want %v", res.Bytes(), pld)
	}

	buf.Reset()
	res.Reset()
}

func writerOpenGPG(b *testing.B, mod libgpg.OpenGPG, buf, res *bytes.Buffer, pld []byte) {
	var (
		err error
		wrt io.WriteCloser
		ptr = cpb.Get().(*[]byte)
	)

	defer cpb.Put(ptr)

	// Encrypt
	if wrt, err = mod.EncryptWriter(buf); err != nil {
		b.Fatalf("EncryptWriter failed: %v", err)
	}

	if _, err = io.CopyBuffer(wrt, bytes.NewReader(pld), *ptr); err != nil {
		b.Fatalf("encryption write failed: %v", err)
	}

	_ = wrt.Close()

	// Decrypt
	if wrt, err = mod.DecryptWriter(res); err != nil {
		b.Fatalf("DecryptWriter failed: %v", err)
	}

	if _, err = io.CopyBuffer(wrt, buf, *ptr); err != nil {
		b.Fatalf("decryption write failed: %v", err)
	}

	_ = wrt.Close()

	if !bytes.Equal(res.Bytes(), pld) {
		b.Fatalf("decryption read failed: got %v, want %v", res.Bytes(), pld)
	}

	buf.Reset()
	res.Reset()
}
