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
	"os"
	"strings"
	"time"

	libgpg "github.com/nabbar/golib/encoding/opengpg"
	gpgalg "github.com/nabbar/golib/encoding/opengpg/slhdsa256s"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpenGPG ML-DSA87/Ed448 Hybrid Model", func() {
	var (
		err error
		ctx context.Context
		cnl context.CancelFunc
		mod libgpg.OpenGPG
		idt *libgpg.Identity
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, time.Second*30)
		idt = &libgpg.Identity{
			Name:    "Test User",
			Comment: "Ginkgo Test Suite",
			Email:   "test@example.com",
		}
	})

	AfterEach(func() {
		if mod != nil {
			_ = mod.Close()
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("New Instance", func() {
		Context("with empty options", func() {
			It("should successfully instantiate", func() {
				mod = gpgalg.New(ctx, gpgalg.Options{})
				Expect(mod).ToNot(BeNil())
			})
		})
	})

	Describe("Key Lifecycle", func() {
		BeforeEach(func() {
			mod = gpgalg.New(ctx, gpgalg.Options{})
		})

		Context("when creating a new keypair", func() {
			It("should populate the identity with public and private keys", func() {
				err = mod.Create(idt)
				Expect(err).ToNot(HaveOccurred())
				Expect(idt.PublicKey).ToNot(BeEmpty())
				Expect(idt.PrivateKey).ToNot(BeEmpty())
				Expect(string(idt.PublicKey)).To(ContainSubstring("BEGIN PGP PUBLIC KEY BLOCK"))
				Expect(string(idt.PrivateKey)).To(ContainSubstring("BEGIN PGP PRIVATE KEY BLOCK"))
			})
		})

		Context("when loading an existing keypair", func() {
			It("should load keys and pass the self-check", func() {
				var oth libgpg.OpenGPG
				defer func() {
					if oth != nil {
						_ = oth.Close()
					}
				}()

				err = mod.Create(idt)
				Expect(err).ToNot(HaveOccurred())

				oth = gpgalg.New(ctx, gpgalg.Options{})

				err = oth.Load(idt)
				Expect(err).ToNot(HaveOccurred())

				err = oth.Check()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when closed", func() {
			It("should return os.ErrClosed for subsequent operations", func() {
				err = mod.Close()
				Expect(err).ToNot(HaveOccurred())

				err = mod.Create(idt)
				Expect(err).To(MatchError(os.ErrClosed))
			})
		})
	})

	Describe("Streaming Encryption & Decryption", func() {
		var tst = []byte(strings.Repeat("Ginkgo Payload Test ", 100))

		BeforeEach(func() {
			mod = gpgalg.New(ctx, gpgalg.Options{})
			err = mod.Create(idt)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("using Reader interfaces", func() {
			It("should successfully encrypt and decrypt data streams", func() {
				var (
					rdr io.ReadCloser
					ptr *[]byte

					buf = bytes.NewBuffer(make([]byte, 0, len(tst)+4096))
					res = bytes.NewBuffer(make([]byte, 0, len(tst)))
				)

				// Encrypt
				rdr, err = mod.EncryptReader(bytes.NewReader(tst))
				Expect(err).ToNot(HaveOccurred())

				ptr = cpb.Get().(*[]byte)
				_, err = io.CopyBuffer(buf, rdr, *ptr)
				Expect(err).ToNot(HaveOccurred())

				Expect(rdr.Close()).To(Succeed())
				Expect(buf.Bytes()).ToNot(Equal(tst))

				// Decrypt
				rdr, err = mod.DecryptReader(bytes.NewReader(buf.Bytes()))
				Expect(err).ToNot(HaveOccurred())

				ptr = cpb.Get().(*[]byte)
				_, err = io.CopyBuffer(res, rdr, *ptr)
				Expect(err).ToNot(HaveOccurred())

				Expect(rdr.Close()).To(Succeed())
				Expect(res.Bytes()).To(Equal(tst))
			})
		})

		Context("using Writer interfaces", func() {
			It("should successfully encrypt and decrypt data streams", func() {
				var (
					wrt io.WriteCloser
					ptr *[]byte

					buf = bytes.NewBuffer(make([]byte, 0, len(tst)+4096))
					res = bytes.NewBuffer(make([]byte, 0, len(tst)))
				)

				// Encrypt
				wrt, err = mod.EncryptWriter(buf)
				Expect(err).ToNot(HaveOccurred())

				ptr = cpb.Get().(*[]byte)
				_, err = io.CopyBuffer(wrt, bytes.NewReader(tst), *ptr)

				Expect(err).ToNot(HaveOccurred())
				Expect(wrt.Close()).To(Succeed())

				Expect(buf.Len()).To(BeNumerically(">", len(tst)))
				Expect(buf.Bytes()).ToNot(Equal(tst))

				// Decrypt
				wrt, err = mod.DecryptWriter(res)
				Expect(err).ToNot(HaveOccurred())

				ptr = cpb.Get().(*[]byte)
				_, err = io.CopyBuffer(wrt, bytes.NewReader(buf.Bytes()), *ptr)

				Expect(err).ToNot(HaveOccurred())
				Expect(wrt.Close()).To(Succeed())

				Expect(res.Len()).To(Equal(len(tst)))
				Expect(res.Bytes()).To(Equal(tst))
			})
		})
	})
})
