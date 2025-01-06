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

package udp_test

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsiz "github.com/nabbar/golib/size"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	prt = strconv.Itoa(GetFreePort(libptc.NetworkUDP))
	adr = "127.0.0.1:" + prt
	siz = 10 * libsiz.SizeKilo
	msg []byte
	res []byte
)

func GetFreePort(proto libptc.NetworkProtocol) int {
	var (
		addr *net.UDPAddr
		lstn *net.UDPConn
		err  error
	)

	if addr, err = net.ResolveUDPAddr(proto.Code(), "localhost:0"); err != nil {
		panic(err)
	}

	if lstn, err = net.ListenUDP(proto.Code(), addr); err != nil {
		panic(err)
	}

	defer func() {
		_ = lstn.Close()
	}()

	return addr.Port
}

func Handler(request libsck.Reader, response libsck.Writer) {
	defer func() {
		_ = request.Close()
		_ = response.Close()
	}()
	_, _ = io.Copy(response, request)
}

func ExpectedErr(e error) {
	Expect(e).ToNot(HaveOccurred())
}

func ExpectedInfo(local, remote net.Addr, state libsck.ConnState) {
	Expect(local).ToNot(BeEmpty())
	Expect(local.Network()).To(BeEquivalentTo(libptc.NetworkUDP.String()))
	Expect(local.String()).ToNot(BeEmpty())

	Expect(remote).ToNot(BeEmpty())
	Expect(remote.Network()).To(BeEquivalentTo(libptc.NetworkUDP.String()))
	Expect(remote.String()).ToNot(BeEmpty())

	Expect(state).ToNot(BeEmpty())
	Expect(state.String()).ToNot(BeEmpty())
}

func ExpectedInfoSrv(msg string) {
	Expect(msg).ToNot(BeEmpty())
}

var _ = Describe("socket/server/udp", func() {
	Context("listen a udp socket with client who's closing connection", func() {
		var (
			err error
			sck libsck.Server
			clt libsck.Client
			srv *sckcfg.ServerConfig
			cli *sckcfg.ClientConfig
		)

		It("Create new server config must succeed", func() {
			srv = &sckcfg.ServerConfig{
				Network:   libptc.NetworkUDP,
				Address:   adr,
				PermFile:  0,
				GroupPerm: 0,
			}
			Expect(srv).ToNot(BeNil())
		})

		It("Create new client config must succeed", func() {
			cli = &sckcfg.ClientConfig{
				Network: libptc.NetworkUDP,
				Address: adr,
			}
			Expect(cli).ToNot(BeNil())
		})

		It("Create new server based on config must succeed", func() {
			sck, err = srv.New(nil, Handler)
			Expect(err).ToNot(HaveOccurred())
			Expect(sck).ToNot(BeNil())
		})

		It("Create new client based on config must succeed", func() {
			clt, err = cli.New()
			Expect(err).ToNot(HaveOccurred())
			Expect(clt).ToNot(BeNil())
		})

		It("Listening in a goroutine must succeed", func() {
			go func() {
				defer GinkgoRecover()
				e := sck.Listen(ctx)
				Expect(e).ToNot(HaveOccurred())
			}()

			time.Sleep(5 * time.Second)
			Expect(sck.IsRunning()).To(BeTrue())
		})

		It("Socket must be having one opened connections", func() {
			time.Sleep(100 * time.Millisecond) // Adding delay for better testing synchronization
			Expect(sck.OpenConnections()).To(BeEquivalentTo(int64(1)))
		})

		It("Writing on the client must succeed", func() {
			msg = append([]byte("Hello World"), libsck.EOL)
			err = clt.Once(ctx, bytes.NewBuffer(msg), nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Socket udp must still having an opened connections", func() {
			time.Sleep(10 * time.Second) // Adding delay for better testing synchronization
			Expect(sck.OpenConnections()).To(BeEquivalentTo(int64(1)))
		})

		It("Closing the socket must succeed", func() {
			Expect(sck.Shutdown(ctx)).ToNot(HaveOccurred())
		})

		It("Socket must not having any opened connections", func() {
			time.Sleep(10 * time.Second) // Adding delay for better testing synchronization
			Expect(sck.OpenConnections()).To(BeEquivalentTo(int64(0)))
		})
	})
})
