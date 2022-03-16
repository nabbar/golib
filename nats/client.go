/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package nats

import (
	"fmt"
	"time"

	libval "github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	natcli "github.com/nats-io/nats.go"
)

type Client struct {

	// Url represents a single NATS server url to which the client
	// will be connecting. If the Servers option is also set, it
	// then becomes the first server in the Servers array.
	Url string

	// Servers is a configured set of servers which this client
	// will use when attempting to connect.
	Servers []string

	// NoRandomize configures whether we will randomize the
	// server pool.
	NoRandomize bool

	// NoEcho configures whether the server will echo back messages
	// that are sent on this connection if we also have matching subscriptions.
	// Note this is supported on servers >= version 1.2. Proto 1 or greater.
	NoEcho bool

	// Name is an optional name label which will be sent to the server
	// on CONNECT to identify the client.
	Name string

	// Verbose signals the server to send an OK ack for commands
	// successfully processed by the server.
	Verbose bool

	// Pedantic signals the server whether it should be doing further
	// validation of subjects.
	Pedantic bool

	// AllowReconnect enables reconnection logic to be used when we
	// encounter a disconnect from the current server.
	AllowReconnect bool

	// MaxReconnect sets the number of reconnect attempts that will be
	// tried before giving up. If negative, then it will never give up
	// trying to reconnect.
	MaxReconnect int

	// ReconnectWait sets the time to backoff after attempting a reconnect
	// to a server that we were already connected to previously.
	ReconnectWait time.Duration

	// ReconnectJitter sets the upper bound for a random delay added to
	// ReconnectWait during a reconnect when no TLS is used.
	// Note that any jitter is capped with ReconnectJitterMax.
	ReconnectJitter time.Duration

	// ReconnectJitterTLS sets the upper bound for a random delay added to
	// ReconnectWait during a reconnect when TLS is used.
	// Note that any jitter is capped with ReconnectJitterMax.
	ReconnectJitterTLS time.Duration

	// Timeout sets the timeout for a Dial operation on a connection.
	Timeout time.Duration

	// DrainTimeout sets the timeout for a Drain Operation to complete.
	DrainTimeout time.Duration

	// FlusherTimeout is the maximum time to wait for write operations
	// to the underlying connection to complete (including the flusher loop).
	FlusherTimeout time.Duration

	// PingInterval is the period at which the client will be sending ping
	// commands to the server, disabled if 0 or negative.
	PingInterval time.Duration

	// MaxPingsOut is the maximum number of pending ping commands that can
	// be awaiting a response before raising an ErrStaleConnection error.
	MaxPingsOut int

	// ReconnectBufSize is the size of the backing bufio during reconnect.
	// Once this has been exhausted publish operations will return an error.
	ReconnectBufSize int

	// SubChanLen is the size of the buffered channel used between the socket
	// Go routine and the message delivery for SyncSubscriptions.
	// NOTE: This does not affect AsyncSubscriptions which are
	// dictated by PendingLimits()
	SubChanLen int

	// User sets the username to be used when connecting to the server.
	User string

	// Password sets the password to be used when connecting to a server.
	Password string

	// Token sets the token to be used when connecting to a server.
	Token string

	// UseOldRequestStyle forces the old method of Requests that utilize
	// a new Inbox and a new Subscription for each request.
	UseOldRequestStyle bool

	// NoCallbacksAfterClientClose allows preventing the invocation of
	// callbacks after Close() is called. Client won't receive notifications
	// when Close is invoked by user code. Default is to invoke the callbacks.
	NoCallbacksAfterClientClose bool

	// Secure enables TLS secure connections that skip server
	// verification by default. NOT RECOMMENDED.
	Secure bool

	// TLSConfig is a custom TLS configuration to use for secure
	// transports.
	TLSConfig libtls.Config
}

func (c Client) Validate() liberr.Error {
	err := ErrorConfigValidation.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.AddParent(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c Client) NewClient(defTls libtls.TLSConfig) (*natcli.Conn, liberr.Error) {
	opts := natcli.GetDefaultOptions()

	if c.Url != "" {
		opts.Url = c.Url
	}

	if len(c.Servers) > 0 {
		opts.Servers = make([]string, 0)
		for _, s := range c.Servers {
			if s != "" {
				opts.Servers = append(opts.Servers, s)
			}
		}
	}

	if c.NoRandomize {
		opts.NoRandomize = true
	}

	if c.NoEcho {
		opts.NoEcho = true
	}

	if c.Name != "" {
		opts.Name = c.Name
	}

	if c.Verbose {
		opts.Verbose = true
	}

	if c.Pedantic {
		opts.Pedantic = true
	}

	if c.AllowReconnect {
		opts.AllowReconnect = true
	}

	if c.MaxReconnect > 0 {
		opts.MaxReconnect = c.MaxReconnect
	}

	if c.ReconnectWait > 0 {
		opts.ReconnectWait = c.ReconnectWait
	}

	if c.ReconnectJitter > 0 {
		opts.ReconnectJitter = c.ReconnectJitter
	}

	if c.ReconnectJitterTLS > 0 {
		opts.ReconnectJitterTLS = c.ReconnectJitterTLS
	}

	if c.Timeout > 0 {
		opts.Timeout = c.Timeout
	}

	if c.DrainTimeout > 0 {
		opts.DrainTimeout = c.DrainTimeout
	}

	if c.FlusherTimeout > 0 {
		opts.FlusherTimeout = c.FlusherTimeout
	}

	if c.PingInterval > 0 {
		opts.PingInterval = c.PingInterval
	}

	if c.MaxPingsOut > 0 {
		opts.MaxPingsOut = c.MaxPingsOut
	}

	if c.ReconnectBufSize > 0 {
		opts.ReconnectBufSize = c.ReconnectBufSize
	}

	if c.SubChanLen > 0 {
		opts.SubChanLen = c.SubChanLen
	}

	if c.User != "" {
		opts.User = c.User
	}

	if c.Password != "" {
		opts.Password = c.Password
	}

	if c.Token != "" {
		opts.Token = c.Token
	}

	if c.UseOldRequestStyle {
		opts.UseOldRequestStyle = true
	}

	if c.NoCallbacksAfterClientClose {
		opts.NoCallbacksAfterClientClose = true
	}

	if c.Secure {
		if t, e := c.TLSConfig.NewFrom(defTls); e != nil {
			return nil, e
		} else {
			opts.TLSConfig = t.TlsConfig("")
		}
		opts.Secure = true
	}

	if n, e := opts.Connect(); e != nil {
		return nil, ErrorClientConnect.ErrorParent(e)
	} else {
		return n, nil
	}
}
