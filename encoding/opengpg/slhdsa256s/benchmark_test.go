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

package slhdsa256s_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	libgpg "github.com/nabbar/golib/encoding/opengpg"
	gpgalg "github.com/nabbar/golib/encoding/opengpg/slhdsa256s"
)

func BenchmarkStreamReader(b *testing.B) {
	var (
		err error
		rdr io.ReadCloser
		ptr *[]byte

		ctx = context.Background()
		mod = gpgalg.New(ctx, gpgalg.Options{})
		idt = &libgpg.Identity{
			Name:    "Bench User",
			Comment: "Benchmark",
			Email:   "bench@example.com",
		}

		tst = []byte(strings.Repeat(tstDataBase, tstDataSize))
		buf = bytes.NewBuffer(make([]byte, 0, 1000*tstDataSize))
		res = bytes.NewBuffer(make([]byte, 0, tstDataSize))
	)

	defer func() {
		if mod != nil {
			_ = mod.Close()
		}
	}()

	if err = mod.Create(idt); err != nil {
		b.Fatalf("failed to setup key: %v", err)
	}

	b.SetBytes(int64(len(tst)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Encrypt
		if rdr, err = mod.EncryptReader(bytes.NewReader(tst)); err != nil {
			b.Fatalf("EncryptReader failed: %v", err)
		}

		ptr = cpb.Get().(*[]byte)

		if _, err = io.CopyBuffer(buf, rdr, *ptr); err != nil {
			b.Fatalf("encryption read failed: %v", err)
		}

		_ = rdr.Close()

		// Decrypt
		if rdr, err = mod.DecryptReader(buf); err != nil {
			b.Fatalf("DecryptReader failed: %v", err)
		}

		ptr = cpb.Get().(*[]byte)

		if _, err = io.CopyBuffer(res, rdr, *ptr); err != nil {
			b.Fatalf("decryption read failed: %v", err)
		}

		_ = rdr.Close()

		if !bytes.Equal(res.Bytes(), tst) {
			b.Fatalf("decryption read failed: got %v, want %v", res.Bytes(), tst)
		}

		buf.Reset()
		res.Reset()
	}
}

func BenchmarkStreamWriter(b *testing.B) {
	var (
		err error
		wrt io.WriteCloser
		ptr *[]byte

		ctx = context.Background()
		mod = gpgalg.New(ctx, gpgalg.Options{})
		idt = &libgpg.Identity{
			Name:    "Bench User",
			Comment: "Benchmark",
			Email:   "bench@example.com",
		}

		tst = []byte(strings.Repeat(tstDataBase, tstDataSize))
		buf = bytes.NewBuffer(make([]byte, 0, 1000*tstDataSize))
		res = bytes.NewBuffer(make([]byte, 0, tstDataSize))
	)

	defer func() {
		if mod != nil {
			_ = mod.Close()
		}
	}()

	if err = mod.Create(idt); err != nil {
		b.Fatalf("failed to setup key: %v", err)
	}

	b.SetBytes(int64(len(tst)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Encrypt
		if wrt, err = mod.EncryptWriter(buf); err != nil {
			b.Fatalf("EncryptWriter failed: %v", err)
		}

		ptr = cpb.Get().(*[]byte)

		if _, err = io.CopyBuffer(wrt, bytes.NewReader(tst), *ptr); err != nil {
			b.Fatalf("encryption write failed: %v", err)
		}

		_ = wrt.Close()

		// Decrypt
		if wrt, err = mod.DecryptWriter(res); err != nil {
			b.Fatalf("DecryptWriter failed: %v", err)
		}

		ptr = cpb.Get().(*[]byte)

		if _, err = io.CopyBuffer(wrt, buf, *ptr); err != nil {
			b.Fatalf("decryption write failed: %v", err)
		}

		_ = wrt.Close()

		if !bytes.Equal(res.Bytes(), tst) {
			b.Fatalf("decryption read failed: got %v, want %v", res.Bytes(), tst)
		}

		buf.Reset()
		res.Reset()
	}
}
