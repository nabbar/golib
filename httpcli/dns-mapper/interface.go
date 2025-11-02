/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// Package dns_mapper provides custom DNS hostname mapping and dialer functionality for HTTP clients.
//
// This package allows overriding DNS resolution by mapping hostnames to specific IP addresses,
// enabling use cases such as testing, development environments, and custom routing without
// modifying system DNS or /etc/hosts files.
//
// Key features:
//   - Hostname-to-IP mapping with wildcard support
//   - Automatic cache management with configurable cleanup
//   - Custom dialer for net.Conn operations
//   - Seamless http.Transport integration
//   - Thread-safe concurrent operations
//   - TLS configuration support
//
// Basic usage:
//
//	import (
//	    "context"
//	    dnsmapper "github.com/nabbar/golib/httpcli/dns-mapper"
//	)
//
//	// Create DNS mapper
//	cfg := &dnsmapper.Config{
//	    DNSMapper: map[string]string{
//	        "api.example.com:443": "192.168.1.100:8443",
//	    },
//	}
//	mapper := dnsmapper.New(context.Background(), cfg, nil, nil)
//	defer mapper.Close()
//
//	// Create HTTP client with custom DNS
//	client := mapper.DefaultClient()
//	resp, _ := client.Get("https://api.example.com")
package dns_mapper

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	tlscas "github.com/nabbar/golib/certificates/ca"
	libdur "github.com/nabbar/golib/duration"
)

// FuncMessage is a callback function type for logging or message handling.
// It receives string messages from the DNS mapper during operations.
type FuncMessage func(msg string)

// DNSMapper defines the interface for DNS mapping and HTTP client creation.
// All methods are thread-safe and can be called concurrently.
type DNSMapper interface {
	// Add registers a new DNS mapping from hostname:port to IP:port.
	// Supports wildcards: "*.example.com:*" or "api.example.com:*"
	Add(from, to string)

	// Get retrieves the mapped address for a given hostname:port.
	// Returns empty string if no mapping exists.
	Get(from string) string

	// Del removes a DNS mapping.
	Del(from string)

	// Len returns the number of active DNS mappings.
	Len() int

	// Walk iterates over all DNS mappings, calling the provided function for each.
	// If the function returns false, iteration stops.
	Walk(func(from, to string) bool)

	// Clean parses an endpoint string into hostname and port components.
	// Returns host, port, and error if parsing fails.
	Clean(endpoint string) (host string, port string, err error)

	// Search resolves an endpoint using DNS mappings without caching.
	// Returns the mapped address or the original endpoint if no mapping exists.
	Search(endpoint string) (string, error)

	// SearchWithCache resolves an endpoint using DNS mappings with caching.
	// Cached results improve performance for repeated lookups.
	SearchWithCache(endpoint string) (string, error)

	// DialContext creates a network connection using custom DNS resolution.
	// This is the custom dialer function used by HTTP transports.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)

	// Transport creates a custom HTTP transport with the given configuration.
	Transport(cfg TransportConfig) *http.Transport

	// TransportWithTLS creates a custom HTTP transport with TLS configuration.
	TransportWithTLS(cfg TransportConfig, ssl *tls.Config) *http.Transport

	// Client creates an HTTP client with custom transport configuration.
	Client(cfg TransportConfig) *http.Client

	// DefaultTransport returns the default HTTP transport for this DNS mapper.
	DefaultTransport() *http.Transport

	// DefaultClient returns the default HTTP client for this DNS mapper.
	DefaultClient() *http.Client

	// GetConfig returns the current DNS mapper configuration.
	GetConfig() Config

	// RegisterTransport registers a custom HTTP transport for the DNS mapper to use.
	RegisterTransport(t *http.Transport)

	// TimeCleaner starts a background goroutine that periodically cleans the cache.
	// The cleanup runs at the specified duration interval until the context is cancelled.
	TimeCleaner(ctx context.Context, dur time.Duration)

	// Close stops the cache cleaner and releases resources.
	// Should be called when the DNS mapper is no longer needed.
	Close() error
}

// GetRootCaCert parses and combines multiple root CA certificates into a single Cert object.
// This utility function is used to aggregate root CA certificates for TLS configuration.
//
// Parameters:
//   - fct: Function that returns a slice of PEM-encoded root CA certificates
//
// Returns a combined Cert object containing all parsed certificates,
// or nil if the function returns an empty slice.
func GetRootCaCert(fct libtls.FctRootCA) tlscas.Cert {
	var res tlscas.Cert

	for _, c := range fct() {
		if res == nil {
			res, _ = tlscas.Parse(c)
		} else {
			_ = res.AppendString(c)
		}
	}

	return res
}

// New creates and initializes a new DNS mapper instance with the given configuration.
//
// This function sets up the DNS mapper with the provided configuration, including:
//   - Initial DNS hostname mappings
//   - HTTP transport configuration
//   - Cache cleanup timer
//   - Optional root CA certificates for TLS
//   - Optional message callback for logging
//
// Parameters:
//   - ctx: Context for lifecycle management and cancellation
//   - cfg: DNS mapper configuration (nil will use default configuration)
//   - fct: Function to retrieve root CA certificates for TLS (nil will use empty function)
//   - msg: Callback function for messages/logging (nil will use no-op function)
//
// Returns a fully initialized DNSMapper instance. The caller should call Close()
// when done to clean up resources.
//
// Example:
//
//	cfg := &Config{
//	    DNSMapper: map[string]string{
//	        "api.example.com:443": "192.168.1.100:8443",
//	    },
//	    TimerClean: libdur.ParseDuration(5 * time.Minute),
//	}
//
//	mapper := New(context.Background(), cfg, nil, func(msg string) {
//	    log.Println("DNS Mapper:", msg)
//	})
//	defer mapper.Close()
func New(ctx context.Context, cfg *Config, fct libtls.FctRootCACert, msg FuncMessage) DNSMapper {
	if cfg == nil {
		cfg = &Config{
			DNSMapper:  make(map[string]string),
			TimerClean: libdur.ParseDuration(3 * time.Minute),
			Transport: TransportConfig{
				Proxy:     nil,
				TLSConfig: nil,
			},
		}
	}

	if fct == nil {
		fct = func() tlscas.Cert {
			return nil
		}
	}

	if msg == nil {
		msg = func(msg string) {}
	}

	d := &dmp{
		d: new(sync.Map),
		z: new(sync.Map),
		c: libatm.NewValue[*Config](),
		t: libatm.NewValue[*http.Transport](),
		f: fct,
		i: msg,
		n: libatm.NewValue[context.CancelFunc](),
		x: libatm.NewValue[context.Context](),
	}

	for edp, adr := range cfg.DNSMapper {
		d.Add(edp, adr)
	}

	d.c.Store(cfg)
	_ = d.DefaultTransport()
	d.TimeCleaner(ctx, cfg.TimerClean.Time())

	return d
}
