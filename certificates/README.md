# Certificates Package

This package provides tools and abstractions to manage TLS/SSL certificates, cipher suites, curves, and related configuration for secure communications in Go applications.

## Features

- **Certificate Management**: Load certificates from files or strings, manage certificate pairs, and support for both file-based and in-memory certificates.
- **Root CA and Client CA**: Add and manage root and client certificate authorities from files or strings.
- **TLS Version Control**: Configure minimum and maximum supported TLS versions.
- **Cipher Suites and Curves**: Select and manage cipher suites and elliptic curves for TLS connections.
- **Client Authentication**: Configure client authentication modes.
- **Dynamic and Session Ticket Options**: Enable or disable dynamic record sizing and session tickets.
- **Configuration Inheritance**: Optionally inherit from a default configuration.

## Installation

```bash
go get github.com/nabbar/golib/certificates
```

## Usage

### Basic Example

```go
import (
    "github.com/nabbar/golib/certificates"
)

cfg := certificates.Config{
    CurveList:            []Curves{...},
    CipherList:           []Cipher{...},
    RootCA:               []Cert{...},
    ClientCA:             []Cert{...},
    Certs:                []Certif{...},
    VersionMin:           tlsversion.VersionTLS12,
    VersionMax:           tlsversion.VersionTLS13,
    AuthClient:           auth.NoClientCert,
    InheritDefault:       false,
    DynamicSizingDisable: false,
    SessionTicketDisable: false,
}

tlsConfig := cfg.New().TLS("your.server.com")
```

### Configuration Fields

- `CurveList`: List of elliptic curves to use (e.g., `["X25519", "P256"]`), see [certificates/curves](#certificatescurves-package) for available options.
- `CipherList`: List of cipher suites (e.g., `["RSA_AES_128_GCM_SHA256"]`), see [certificates/cipher](#certificatescipher-package) for available options.
- `RootCA`: List of root CA certificates, supported formats include PEM strings or file paths. See [certificates/ca](#certificatesca-package) for more details.
- `ClientCA`: List of client CA certificates. Supported formats include PEM strings or file paths. See [certificates/ca](#certificatesca-package) for more details.
- `Certs`: List of certificate pairs (key/cert). Supported formats include PEM strings or file paths. See [certificates/certs](#certificatescerts-package) for more details.
- `VersionMin`: Minimum TLS version (e.g., `"1.2"`). See [certificates/tlsversion](#certificatestlsversion-package) for available options. 
- `VersionMax`: Maximum TLS version (e.g., `"1.3"`). See [certificates/tlsversion](#certificatestlsversion-package) for available options.
- `AuthClient`: Client authentication mode (e.g., `"none"`, `"require"`). See [certificates/auth](#certificatesauth-package) for available modes.
- `InheritDefault`: Inherit from the default configuration if set to `true`. Can be used to apply a base configuration across multiple TLS configurations.
- `DynamicSizingDisable`: Disable dynamic record sizing if set to `true`. Used to control how TLS records are sized dynamically based on the payload.
- `SessionTicketDisable`: Disable session tickets if set to `true`. Used to control whether session tickets are used for session resumption.

### Methods

- `New() TLSConfig`: Create a new TLS configuration from the current config.
- `AddRootCAString(string) bool`: Add a root CA from a string.
- `AddRootCAFile(string) error`: Add a root CA from a file.
- `AddClientCAString(string) bool`: Add a client CA from a string.
- `AddClientCAFile(string) error`: Add a client CA from a file.
- `AddCertificatePairString(key, cert string) error`: Add a certificate pair from strings.
- `AddCertificatePairFile(keyFile, certFile string) error`: Add a certificate pair from files.
- `SetVersionMin(Version)`: Set the minimum TLS version.
- `SetVersionMax(Version)`: Set the maximum TLS version.
- `SetCipherList([]Cipher)`: Set the list of cipher suites.
- `SetCurveList([]Curves)`: Set the list of elliptic curves.
- `SetDynamicSizingDisabled(bool)`: Enable/disable dynamic record sizing.
- `SetSessionTicketDisabled(bool)`: Enable/disable session tickets.
- `TLS(serverName string) *tls.Config`: Get a `*tls.Config` for use in servers/clients.

### Testing

Tests are written using [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/).

```bash
ginkgo -cover .
```

## Liens utiles

- [Documentation Ginkgo](https://onsi.github.io/ginkgo/)
- [crypto/tls (Go)](https://pkg.go.dev/crypto/tls)
- [crypto/x509 (Go)](https://pkg.go.dev/crypto/x509)

## Licence

MIT © Nicolas JUHEL

---

# certificates/auth package

The `certificates/auth` package provides types and helpers to configure TLS client authentication for Go applications. It allows you to select, parse, and serialize the client authentication mode for TLS servers, supporting JSON, YAML, TOML, and text formats.

## Features

- Enumerates all standard TLS client authentication modes
- Parse from string, int, or serialized config (JSON/YAML/TOML)
- Serialize/deserialize for config files and environment variables
- Helper functions for mapping between string, code, and TLS types
- Viper decoder hook for config integration

## Main Types

- `ClientAuth`: Enum type for TLS client authentication (wraps `tls.ClientAuthType`)

### Available Modes

- `NoClientCert` (`"none"`): No client certificate required
- `RequestClientCert` (`"request"`): Request client certificate, but do not require
- `RequireAnyClientCert` (`"require"`): Require any client certificate
- `VerifyClientCertIfGiven` (`"verify"`): Verify client certificate if provided
- `RequireAndVerifyClientCert` (`"strict require verify"`): Require and verify client certificate

## Example: Parse and Use ClientAuth

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/certificates/auth"
)

func main() {
    // Parse from string
    ca := auth.Parse("require")
    fmt.Println("ClientAuth:", ca.String()) // Output: require

    // Use as tls.ClientAuthType
    tlsType := ca.TLS()
    fmt.Println("TLS ClientAuthType:", tlsType)
}
```

## Example: Marshal/Unmarshal JSON

```go
import (
    "encoding/json"
    "github.com/nabbar/golib/certificates/auth"
)

type MyConfig struct {
    Auth auth.ClientAuth `json:"authClient"`
}

func main() {
    // Marshal
    cfg := MyConfig{Auth: auth.RequireAndVerifyClientCert}
    b, _ := json.Marshal(cfg)
    fmt.Println(string(b)) // {"authClient":"strict require verify"}

    // Unmarshal
    var cfg2 MyConfig
    _ = json.Unmarshal([]byte(`{"authClient":"verify"}`), &cfg2)
    fmt.Println(cfg2.Auth.String()) // verify
}
```

## Example: Use with Viper

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/certificates/auth"
    "github.com/go-viper/mapstructure/v2"
)

v := viper.New()
v.Set("authClient", "require")
var cfg struct {
    Auth auth.ClientAuth
}
v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.DecodeHook = auth.ViperDecoderHook()
})
fmt.Println(cfg.Auth.String()) // require
```

## Options

- **String values:** `"none"`, `"request"`, `"require"`, `"verify"`, `"strict require verify"`
- **Int values:** Use `ParseInt(int)` to convert from integer codes
- **Serialization:** Supports JSON, YAML, TOML, text, CBOR

## Advanced

- Use `auth.List()` to get all available modes
- Use `auth.Parse(s string)` to parse from string
- Use `auth.ParseInt(i int)` to parse from int
- Use `auth.ViperDecoderHook()` for config libraries

## Error Handling

Parsing functions return a default value (`NoClientCert`) if the input is invalid. Serialization methods return errors if the format is not supported.

---

# certificates/ca package

The `certificates/ca` package provides tools to parse, manage, and serialize X.509 certificate chains for use as Root or Client Certificate Authorities (CAs) in Go applications. It supports loading certificates from PEM strings or files, and serializing/deserializing in multiple formats (JSON, YAML, TOML, CBOR, text, binary).

## Features

- Parse and manage X.509 certificate chains
- Load certificates from PEM strings or file paths
- Serialize/deserialize in JSON, YAML, TOML, CBOR, text, and binary
- Append certificates to `x509.CertPool`
- Integrate with Viper for configuration

## Main Types

- `Cert`: Interface for certificate chains (implements marshaling/unmarshaling for all supported formats)

## Example: Parse a PEM Certificate Chain

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/certificates/ca"
)

func main() {
    pem := `-----BEGIN CERTIFICATE-----
MIIB... (your PEM data)
-----END CERTIFICATE-----`
    cert, err := ca.Parse(pem)
    if err != nil {
        panic(err)
    }
    fmt.Println("Number of certs:", cert.Len())
}
```

## Example: Load Certificate from File Path

If the PEM block contains a file path, the package will load the certificate from the file.

```go
cert, err := ca.Parse("/etc/ssl/certs/ca-cert.pem")
if err != nil {
    panic(err)
}
```

## Example: Marshal/Unmarshal JSON

```go
import (
    "encoding/json"
    "github.com/nabbar/golib/certificates/ca"
)

type MyConfig struct {
    Root ca.Cert `json:"rootCA"`
}

func main() {
    pem := `-----BEGIN CERTIFICATE-----...`
    cfg := MyConfig{}
    _ = json.Unmarshal([]byte(fmt.Sprintf(`{"rootCA":%q}`, pem)), &cfg)
    b, _ := json.Marshal(cfg)
    fmt.Println(string(b))
}
```

## Example: Append to CertPool

```go
import (
    "crypto/x509"
    "github.com/nabbar/golib/certificates/ca"
)

cert, _ := ca.Parse(pemString)
pool := x509.NewCertPool()
cert.AppendPool(pool)
```

## Example: Use with Viper

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/certificates/ca"
    "github.com/go-viper/mapstructure/v2"
)

v := viper.New()
v.Set("rootCA", "-----BEGIN CERTIFICATE-----...")
var cfg struct {
    Root ca.Cert
}
v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.DecodeHook = ca.ViperDecoderHook()
})
```

## Options & Methods

- **Parse(str string) (Cert, error)**: Parse a PEM string or file path to a `Cert`
- **ParseByte([]byte) (Cert, error)**: Parse from bytes
- **Cert.Len() int**: Number of certificates in the chain
- **Cert.AppendPool(*x509.CertPool)**: Add all certs to a pool
- **Cert.AppendBytes([]byte) error**: Append certs from bytes
- **Cert.AppendString(string) error**: Append certs from string
- **Cert.String() string**: PEM-encoded chain as string
- **Cert.Marshal/Unmarshal**: Supports Text, Binary, JSON, YAML, TOML, CBOR

## Error Handling

- Returns Go `error` or custom errors (`ErrInvalidPairCertificate`, `ErrInvalidCertificate`)
- Always check returned errors

---

# certificates/certs package

The `certificates/certs` package provides types and utilities for handling X.509 certificate pairs and chains in Go. It supports loading certificates from PEM strings or files, serializing/deserializing in multiple formats (JSON, YAML, TOML, CBOR, text, binary), and converting to `tls.Certificate` for use in TLS servers/clients.

## Features

- Parse certificate pairs (key + cert) or chains (cert + private key in one PEM)
- Load from PEM strings or file paths
- Serialize/deserialize in JSON, YAML, TOML, CBOR, text, and binary
- Convert to `tls.Certificate`
- Helper methods for extracting PEM, checking type, and file origin
- Viper integration for config loading

## Main Types

- `Cert`: Interface for certificate objects (pair or chain)
- `Certif`: Implementation of `Cert`
- `ConfigPair`: Struct for key/cert pair
- `ConfigChain`: String type for PEM chain

## Example: Parse a Certificate Pair

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/certificates/certs"
)

func main() {
    cert, err := certs.ParsePair("server.key", "server.crt")
    if err != nil {
        panic(err)
    }
    fmt.Println("Is pair:", cert.IsPair())
    fmt.Println("Is file:", cert.IsFile())
}
```

## Example: Parse a Certificate Chain

```go
cert, err := certs.Parse("chain.pem")
if err != nil {
    panic(err)
}
fmt.Println("Is chain:", cert.IsChain())
```

## Example: Marshal/Unmarshal JSON with Certificates Pair & Chain

```go
import (
    "encoding/json"
    "github.com/nabbar/golib/certificates/certs"
)

type MyConfig struct {
    Cert []certs.Certif `json:"cert"`
}

func main() {
    jsonData := `[{"key":"server1.key","pub":"server1.crt"},{"key":"server2.key","pub":"server2.crt"},"server3.pem","server4.pem"]`
    var cfg MyConfig
    _ = json.Unmarshal([]byte(jsonData), &cfg)
    b, _ := json.Marshal(cfg)
    fmt.Println(string(b))
}
```

## Example: Get PEM Strings

```go
pub, key, err := cert.Pair()
if err != nil {
    panic(err)
}
fmt.Println("Public cert PEM:", pub)
fmt.Println("Private key PEM:", key)
```

## Example: Use with Viper

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/certificates/certs"
    "github.com/go-viper/mapstructure/v2"
)

v := viper.New()
v.Set("cert", "chain.pem")
var cfg struct {
    Cert certs.Certif
}
v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.DecodeHook = certs.ViperDecoderHook()
})
```

## Options & Methods

- **Parse(chain string) (Cert, error)**: Parse a PEM chain or file path
- **ParsePair(key, pub string) (Cert, error)**: Parse a key/cert pair (PEM or file)
- **Cert.TLS() tls.Certificate**: Get as `tls.Certificate`
- **Cert.IsChain() bool**: Is a chain (cert + key in one PEM)
- **Cert.IsPair() bool**: Is a pair (separate key/cert)
- **Cert.IsFile() bool**: Loaded from file(s)
- **Cert.GetCerts() []string**: Get underlying PEMs or file paths
- **Cert.Pair() (pub, key string, error)**: Get PEM strings for cert and key
- **Cert.Chain() (string, error)**: Get full PEM chain
- **Cert.Marshal/Unmarshal**: Supports Text, Binary, JSON, YAML, TOML, CBOR

## Error Handling

- Returns Go `error` or custom errors (`ErrInvalidPairCertificate`, `ErrInvalidCertificate`, `ErrInvalidPrivateKey`)
- Always check returned errors

---

# certificates/cipher package

The `certificates/cipher` package provides types and utilities for handling TLS cipher suites in Go. It allows you to list, parse, serialize, and use cipher suites for configuring TLS servers and clients, supporting multiple serialization formats (JSON, YAML, TOML, CBOR, text).

## Features

- Enumerates all supported TLS cipher suites (TLS 1.2 & 1.3)
- Parse from string or integer
- Serialize/deserialize for config files (JSON, YAML, TOML, CBOR, text)
- Helper methods for code, string, and TLS value conversion
- Viper decoder hook for config integration

## Main Types

- `Cipher`: Enum type for TLS cipher suites (wraps `uint16`)

## Example: List and Parse Cipher Suites

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/certificates/cipher"
)

func main() {
    // List all supported ciphers
    for _, c := range cipher.List() {
        fmt.Println("Cipher:", c.String(), "TLS value:", c.TLS())
    }

    // Parse from string
    c := cipher.Parse("ECDHE_RSA_AES_128_GCM_SHA256")
    fmt.Println("Parsed cipher:", c.String())

    // Parse from int
    c2 := cipher.ParseInt(4865) // Example TLS value
    fmt.Println("Parsed from int:", c2.String())
}
```

## Example: Marshal/Unmarshal JSON

```go
import (
    "encoding/json"
    "github.com/nabbar/golib/certificates/cipher"
)

type MyConfig struct {
    Cipher cipher.Cipher `json:"cipher"`
}

func main() {
    // Marshal
    cfg := MyConfig{Cipher: cipher.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256}
    b, _ := json.Marshal(cfg)
    fmt.Println(string(b)) // {"cipher":"ecdhe_rsa_aes_128_gcm_sha256"}

    // Unmarshal
    var cfg2 MyConfig
    _ = json.Unmarshal([]byte(`{"cipher":"aes_128_gcm_sha256"}`), &cfg2)
    fmt.Println(cfg2.Cipher.String()) // aes_128_gcm_sha256
}
```

## Example: Use with Viper

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/certificates/cipher"
    "github.com/go-viper/mapstructure/v2"
)

v := viper.New()
v.Set("cipher", "chacha20_poly1305_sha256")
var cfg struct {
    Cipher cipher.Cipher
}
v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.DecodeHook = cipher.ViperDecoderHook()
})
fmt.Println(cfg.Cipher.String()) // chacha20_poly1305_sha256
```

## Options & Methods

- **List() []Cipher**: List all supported cipher suites
- **ListString() []string**: List all supported cipher suite names as strings
- **Parse(s string) Cipher**: Parse from string (case-insensitive, flexible format)
- **ParseInt(i int) Cipher**: Parse from integer TLS value
- **Cipher.String() string**: Get string representation
- **Cipher.TLS() uint16**: Get TLS value
- **Cipher.Check() bool**: Check if cipher is supported
- **ViperDecoderHook()**: Viper integration for config loading

## Supported Cipher Suites

- `ecdhe_rsa_aes_128_gcm_sha256`
- `ecdhe_ecdsa_aes_128_gcm_sha256`
- `ecdhe_rsa_aes_256_gcm_sha384`
- `ecdhe_ecdsa_aes_256_gcm_sha384`
- `ecdhe_rsa_chacha20_poly1305_sha256`
- `ecdhe_ecdsa_chacha20_poly1305_sha256`
- `aes_128_gcm_sha256`
- `aes_256_gcm_sha384`
- `chacha20_poly1305_sha256`
- ...and retro-compatible aliases

## Error Handling

- Parsing functions return `Unknown` if the input is invalid.
- Serialization methods return errors if the format is not supported.

---

# certificates/curves package

The `certificates/curves` package provides types and utilities for handling elliptic curves for TLS in Go. It allows you to list, parse, serialize, and use elliptic curves for configuring TLS servers and clients, supporting multiple serialization formats (JSON, YAML, TOML, CBOR, text).

## Features

- Enumerates all supported TLS elliptic curves (X25519, P256, P384, P521)
- Parse from string or integer
- Serialize/deserialize for config files (JSON, YAML, TOML, CBOR, text)
- Helper methods for code, string, and TLS value conversion
- Viper decoder hook for config integration

## Main Types

- `Curves`: Enum type for TLS elliptic curves (wraps `uint16`)

## Example: List and Parse Curves

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/certificates/curves"
)

func main() {
    // List all supported curves
    for _, c := range curves.List() {
        fmt.Println("Curve:", c.String(), "TLS value:", c.TLS())
    }

    // Parse from string
    c := curves.Parse("P256")
    fmt.Println("Parsed curve:", c.String())

    // Parse from int
    c2 := curves.ParseInt(29) // Example TLS value for X25519
    fmt.Println("Parsed from int:", c2.String())
}
```

## Example: Marshal/Unmarshal JSON

```go
import (
    "encoding/json"
    "github.com/nabbar/golib/certificates/curves"
)

type MyConfig struct {
    Curve curves.Curves `json:"curve"`
}

func main() {
    // Marshal
    cfg := MyConfig{Curve: curves.P256}
    b, _ := json.Marshal(cfg)
    fmt.Println(string(b)) // {"curve":"P256"}

    // Unmarshal
    var cfg2 MyConfig
    _ = json.Unmarshal([]byte(`{"curve":"X25519"}`), &cfg2)
    fmt.Println(cfg2.Curve.String()) // X25519
}
```

## Example: Use with Viper

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/certificates/curves"
    "github.com/go-viper/mapstructure/v2"
)

v := viper.New()
v.Set("curve", "P384")
var cfg struct {
    Curve curves.Curves
}
v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.DecodeHook = curves.ViperDecoderHook()
})
fmt.Println(cfg.Curve.String()) // P384
```

## Options & Methods

- **List() []Curves**: List all supported curves
- **ListString() []string**: List all supported curve names as strings
- **Parse(s string) Curves**: Parse from string (case-insensitive, flexible format)
- **ParseInt(i int) Curves**: Parse from integer TLS value
- **Curves.String() string**: Get string representation
- **Curves.TLS() tls.CurveID**: Get TLS value
- **Curves.Check() bool**: Check if curve is supported
- **ViperDecoderHook()**: Viper integration for config loading

## Supported Curves

- `X25519`
- `P256`
- `P384`
- `P521`

## Error Handling

- Parsing functions return `Unknown` if the input is invalid.
- Serialization methods return errors if the format is not supported.

---

# certificates/tlsversion package

The `certificates/tlsversion` package provides types and utilities for handling TLS protocol versions in Go. It allows you to list, parse, serialize, and use TLS versions for configuring secure servers and clients, supporting multiple serialization formats (JSON, YAML, TOML, CBOR, text).

## Features

- Enumerates all supported TLS protocol versions (TLS 1.0, 1.1, 1.2, 1.3)
- Parse from string or integer
- Serialize/deserialize for config files (JSON, YAML, TOML, CBOR, text)
- Helper methods for code, string, and TLS value conversion
- Viper decoder hook for config integration

## Main Types

- `Version`: Enum type for TLS protocol versions (wraps `int`)

## Supported Versions

- `TLS 1.0`
- `TLS 1.1`
- `TLS 1.2`
- `TLS 1.3`

## Example: List and Parse TLS Versions

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/certificates/tlsversion"
)

func main() {
    // List all supported TLS versions
    for _, v := range tlsversion.List() {
        fmt.Println("Version:", v.String(), "TLS value:", v.TLS())
    }

    // Parse from string
    v := tlsversion.Parse("1.2")
    fmt.Println("Parsed version:", v.String())

    // Parse from int
    v2 := tlsversion.ParseInt(0x0304) // Example TLS value for 1.3
    fmt.Println("Parsed from int:", v2.String())
}
```

## Example: Marshal/Unmarshal JSON

```go
import (
    "encoding/json"
    "github.com/nabbar/golib/certificates/tlsversion"
)

type MyConfig struct {
    Version tlsversion.Version `json:"version"`
}

func main() {
    // Marshal
    cfg := MyConfig{Version: tlsversion.VersionTLS12}
    b, _ := json.Marshal(cfg)
    fmt.Println(string(b)) // {"version":"TLS 1.2"}

    // Unmarshal
    var cfg2 MyConfig
    _ = json.Unmarshal([]byte(`{"version":"TLS 1.3"}`), &cfg2)
    fmt.Println(cfg2.Version.String()) // TLS 1.3
}
```

## Example: Use with Viper

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/certificates/tlsversion"
    "github.com/go-viper/mapstructure/v2"
)

v := viper.New()
v.Set("version", "1.2")
var cfg struct {
    Version tlsversion.Version
}
v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.DecodeHook = tlsversion.ViperDecoderHook()
})
fmt.Println(cfg.Version.String()) // TLS 1.2
```

## Options & Methods

- **List() []Version**: List all supported TLS versions
- **ListHigh() []Version**: List only high/secure TLS versions (1.2, 1.3)
- **Parse(s string) Version**: Parse from string (case-insensitive, flexible format)
- **ParseInt(i int) Version**: Parse from integer TLS value
- **Version.String() string**: Get string representation (e.g. "TLS 1.2")
- **Version.Code() string**: Get code representation (e.g. "tls_1_2")
- **Version.TLS() uint16**: Get TLS value
- **Version.Check() bool**: Check if version is supported
- **ViperDecoderHook()**: Viper integration for config loading

## Serialization

- Supports JSON, YAML, TOML, CBOR, and text formats for marshaling/unmarshaling.

## Error Handling

- Parsing functions return `VersionUnknown` if the input is invalid.
- Serialization methods return errors if the format is not supported.

---

This documentation covers all main features, options, and usage examples for the `certificates` package and sub packages.