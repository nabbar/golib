## `httpcli` Package

The `httpcli` package provides helpers and abstractions for creating, configuring, and managing HTTP clients in Go. It is designed to simplify HTTP client instantiation, support advanced options (TLS, proxy, timeouts, DNS mapping), and enable easy integration with custom DNS resolvers for testing or advanced routing.

### Features

- Easy creation of HTTP clients with sensible defaults
- Support for custom DNS mapping (mock/fake DNS) via the `dns-mapper` subpackage
- TLS configuration and proxy support
- Configurable timeouts and connection options
- Thread-safe management of default DNS mappers and clients
- Error handling with custom error codes

---

### Main Types & Functions

#### Getting a Default HTTP Client

```go
import "github.com/nabbar/golib/httpcli"

client := httpcli.GetClient()
// Use client for HTTP requests
```

#### Custom DNS Mapper

You can set a custom DNS mapper to control how hostnames are resolved (useful for testing or routing):

```go
import (
    "github.com/nabbar/golib/httpcli"
    htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

dns := htcdns.New(context.Background(), &htcdns.Config{/* ... */}, nil, nil)
httpcli.SetDefaultDNSMapper(dns)
client := httpcli.GetClient()
```

For more details on the `dns-mapper`, see the [dns-mapper subpackage](#subpackage-dns-mapper).

#### Options Structure

The `Options` struct allows you to configure timeouts, keep-alive, compression, HTTP/2, TLS, forced IP, and proxy settings:

```go
import "github.com/nabbar/golib/httpcli"

opt := httpcli.Options{
    Timeout:          10 * time.Second,
    DisableKeepAlive: false,
    TLS:              httpcli.OptionTLS{Enable: true, Config: /* ... */},
    // ... other options
}
```

#### Validation

Validate your options before using them:

```go
err := opt.Validate()
if err != nil {
    // handle validation error
}
```

#### Error Handling

All errors are returned as `liberr.Error` with specific codes (e.g., `ErrorParamEmpty`, `ErrorValidatorError`). Always check errors after each operation.

---

### Example Usage

```go
client := httpcli.GetClient()
resp, err := client.Get("https://example.com")
if err != nil {
    // handle error
}
defer resp.Body.Close()
// process response
```

---

### Subpackage: `dns-mapper`

The `dns-mapper` subpackage provides a flexible and thread-safe DNS mapping and mock resolver for Go HTTP clients. 
<br />It allows you to map specific hostnames (with or without ports, including wildcards) to custom destinations, making it ideal for testing, local development, or advanced routing scenarios.

---

#### Features

- Map hostnames (optionally with port and wildcards) to custom IP:port destinations.
- Transparent integration with HTTP clients and transports.
- Dynamic add, remove, and lookup of DNS mappings at runtime.
- Caching for efficient repeated lookups.
- Customizable cleaning interval for idle connections.
- Full configuration struct for transport and TLS options.
- Thread-safe and context-aware.

---

#### Main Types & Functions

##### `Config` Struct

Defines DNS mapping and transport options:

- `DNSMapper`: `map[string]string` — source to destination mappings (e.g., `"test.example.com:8080": "127.0.0.1:8081"`).
- `TimerClean`: cleaning interval for idle connections.
- `Transport`: HTTP transport configuration (timeouts, proxy, TLS, etc.).
- `TLSConfig`: optional TLS configuration.

##### `DNSMapper` Interface

Main interface for DNS mapping and HTTP client integration:

- `Add(from, to string)`: Add a mapping.
- `Get(from string) string`: Get the mapped destination for a source.
- `Del(from string)`: Remove a mapping.
- `Len() int`: Number of mappings.
- `Walk(func(from, to string) bool)`: Iterate over mappings.
- `Search(endpoint string) (string, error)`: Find the mapped destination for an endpoint.
- `SearchWithCache(endpoint string) (string, error)`: Same as `Search`, with caching.
- `DialContext(ctx, network, address string) (net.Conn, error)`: Custom dialer for HTTP transport.
- `Transport(cfg TransportConfig) *http.Transport`: Create a custom HTTP transport.
- `Client(cfg TransportConfig) *http.Client`: Create an HTTP client using the DNS mapper.
- `DefaultTransport() *http.Transport`: Get the default transport.
- `DefaultClient() *http.Client`: Get the default client.
- `Close() error`: Clean up resources.

---

#### Example Usage

```go
import (
    "context"
    "github.com/nabbar/golib/httpcli/dns-mapper"
    "time"
)

cfg := dns_mapper.Config{
    DNSMapper: map[string]string{
        "test.example.com:8080": "127.0.0.1:8081",
        "*.dev.local:80":        "127.0.0.2:8080",
    },
    TimerClean: dns_mapper.ParseDuration(5 * time.Minute),
    // Transport and TLSConfig can be set as needed
}

dns := dns_mapper.New(context.Background(), &cfg, nil, nil)
defer dns.Close()

// Add or remove mappings dynamically
dns.Add("api.local:443", "10.0.0.1:8443")
dns.Del("test.example.com:8080")

// Use with HTTP client
client := dns.DefaultClient()
resp, err := client.Get("http://api.local:443/health")
if err != nil {
    // handle error
}
defer resp.Body.Close()
// process response
```

---

#### Wildcard and Port Matching

- Hostnames can include wildcards (e.g., `*.example.com` or `*.*.dev.local`).
- Ports can be specified or wildcarded (e.g., `*:8080`).
- The mapping logic matches the most specific rule.

---

#### Error Handling

All errors are wrapped with custom codes for diagnostics. Use `err.Error()` for user-friendly messages.

---

#### Notes

- The DNS mapper is thread-safe and suitable for concurrent use.
- Integrates seamlessly with Go's `http.Transport` and `http.Client`.
- Designed for Go 1.18+.

---

For more details, refer to the GoDoc or the source code in the `dns-mapper` package.