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

// Package httpcli provides advanced HTTP client management with DNS mapping capabilities.
//
// This package offers a simplified API for creating and configuring HTTP clients with
// integrated DNS mapping support through the dns-mapper subpackage. It enables custom
// DNS resolution, TLS configuration, and flexible transport options.
//
// Key features:
//   - Default HTTP client with sensible configuration
//   - Custom DNS mapping for hostname resolution override
//   - TLS/SSL configuration support
//   - Proxy configuration with authentication
//   - Thread-safe singleton DNS mapper management
//   - Automatic resource cleanup and lifecycle management
//
// Basic usage:
//
//	import "github.com/nabbar/golib/httpcli"
//
//	// Get default HTTP client
//	client := httpcli.GetClient()
//
//	// Make HTTP request
//	resp, err := client.Get("https://api.example.com")
//	if err != nil {
//	    panic(err)
//	}
//	defer resp.Body.Close()
//
// With custom DNS mapping:
//
//	import (
//	    "github.com/nabbar/golib/httpcli"
//	    htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
//	)
//
//	// Create DNS mapper
//	cfg := &htcdns.Config{
//	    DNSMapper: map[string]string{
//	        "api.example.com:443": "192.168.1.100:8443",
//	    },
//	}
//	mapper := htcdns.New(context.Background(), cfg, nil, nil)
//	defer mapper.Close()
//
//	// Set as default
//	httpcli.SetDefaultDNSMapper(mapper)
//
//	// Get client with custom DNS mapping
//	client := httpcli.GetClient()
package httpcli

import (
	"context"
	"net/http"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libdur "github.com/nabbar/golib/duration"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

const (
	// ClientTimeout5Sec is a default timeout constant of 5 seconds for HTTP client operations.
	ClientTimeout5Sec = 5 * time.Second // nolint
)

// dns is the global singleton DNS mapper instance stored in an atomic value for thread safety.
var dns = libatm.NewValue[htcdns.DNSMapper]()

// initDNSMapper creates and returns a new DNS mapper with default configuration.
// This function is called automatically by DefaultDNSMapper when no mapper has been set.
//
// Default configuration includes:
//   - Empty DNS mapping (no hostname overrides)
//   - 3-minute cache cleanup interval
//   - Standard HTTP transport settings (50 max idle connections, 30s timeouts)
//   - HTTP/2 enabled
//   - Keep-alive enabled
//
// Returns a configured DNSMapper instance ready for use.
func initDNSMapper() htcdns.DNSMapper {
	return htcdns.New(context.Background(), &htcdns.Config{
		DNSMapper:  make(map[string]string),
		TimerClean: libdur.ParseDuration(3 * time.Minute),
		Transport: htcdns.TransportConfig{
			Proxy:                 nil,
			TLSConfig:             &libtls.Config{},
			DisableHTTP2:          false,
			DisableKeepAlive:      false,
			DisableCompression:    false,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   5,
			MaxConnsPerHost:       25,
			TimeoutGlobal:         libdur.ParseDuration(30 * time.Second),
			TimeoutKeepAlive:      libdur.ParseDuration(15 * time.Second),
			TimeoutTLSHandshake:   libdur.ParseDuration(10 * time.Second),
			TimeoutExpectContinue: libdur.ParseDuration(3 * time.Second),
			TimeoutIdleConn:       libdur.ParseDuration(30 * time.Second),
			TimeoutResponseHeader: 0,
		},
	}, nil, nil)
}

// DefaultDNSMapper returns the default DNS mapper instance.
// If no DNS mapper has been set via SetDefaultDNSMapper, this function
// creates a new one using initDNSMapper with default configuration.
//
// This function is thread-safe and uses atomic operations to ensure
// concurrent access is handled correctly. The DNS mapper is created
// only once on first access (lazy initialization).
//
// Returns the global DNS mapper instance that can be used to:
//   - Add custom hostname-to-IP mappings
//   - Create HTTP clients with custom DNS resolution
//   - Configure custom transport options
//
// Example:
//
//	mapper := httpcli.DefaultDNSMapper()
//	mapper.Add("api.example.com:443", "192.168.1.100:8443")
//	client := httpcli.GetClient()
func DefaultDNSMapper() htcdns.DNSMapper {
	if dns.Load() == nil {
		SetDefaultDNSMapper(initDNSMapper())
	}

	return dns.Load()
}

// SetDefaultDNSMapper replaces the default DNS mapper with a custom instance.
// The previous DNS mapper (if any) is automatically closed to free resources.
//
// This function is thread-safe and uses atomic operations to ensure
// the swap is performed correctly even under concurrent access.
//
// Parameters:
//   - d: The new DNS mapper instance to use as default. If nil, the function returns without action.
//
// The old DNS mapper is automatically closed when replaced, ensuring proper cleanup
// of goroutines, timers, and other resources associated with the old instance.
//
// Example:
//
//	// Create custom DNS mapper
//	cfg := &htcdns.Config{
//	    DNSMapper: map[string]string{
//	        "api.example.com:443": "192.168.1.100:8443",
//	    },
//	}
//	customMapper := htcdns.New(context.Background(), cfg, nil, nil)
//
//	// Set as default (old mapper is automatically closed)
//	httpcli.SetDefaultDNSMapper(customMapper)
func SetDefaultDNSMapper(d htcdns.DNSMapper) {
	if d == nil {
		return
	}

	if o := dns.Swap(d); o != nil {
		_ = o.Close()
	}
}

// FctHttpClient is a function type that returns an HTTP client.
// This type is used for dependency injection and testing purposes.
type FctHttpClient func() *http.Client

// FctHttpClientSrv is a function type that returns an HTTP client configured for a specific server.
// The servername parameter can be used to select different client configurations.
type FctHttpClientSrv func(servername string) *http.Client

// HttpClient defines the minimal interface for HTTP operations.
// This interface is compatible with *http.Client and can be used for testing with mock clients.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GetClient returns an HTTP client configured with the default DNS mapper.
// This is the primary entry point for obtaining HTTP clients in this package.
//
// The returned client includes:
//   - Custom DNS resolution via the default DNS mapper
//   - Connection pooling and keep-alive
//   - Configured transport with timeouts
//   - HTTP/2 support (if enabled in DNS mapper config)
//
// The client is safe for concurrent use and reuses connections for efficiency.
// Always reuse the returned client for multiple requests instead of calling
// GetClient repeatedly.
//
// Example:
//
//	client := httpcli.GetClient()
//	resp, err := client.Get("https://api.example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer resp.Body.Close()
func GetClient() *http.Client {
	return DefaultDNSMapper().DefaultClient()
}
