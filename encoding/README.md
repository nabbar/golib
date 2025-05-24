## `encoding` Package

The `encoding` package provides a unified interface and common utilities for various encoding and cryptographic operations. It serves as the entry point for several specialized subpackages, each implementing a specific encoding or cryptographic algorithm.

### Overview

This package defines the `Coder` interface, which standardizes encoding and decoding operations across different algorithms. The subpackages implement this interface for specific use cases, such as encryption, hashing, hexadecimal encoding, and random data generation.

### Main Features

- Common `Coder` interface for encoding/decoding bytes and streams
- Support for encoding and decoding via both byte slices and `io.Reader`/`io.Writer`
- Memory management with a `Reset` method
- Extensible design for adding new encoding algorithms

### Subpackages

The following subpackages provide concrete implementations of the `Coder` interface and related utilities:

- **aes**: Symmetric encryption and decryption using the AES algorithm. See the [`aes` subpackage](#aes-subpackage) for details.
- **hexa**: Hexadecimal encoding and decoding. See the [`hexa` subpackage](#hexa-subpackage) for details.
- **mux**: Multiplexed encoding, allowing composition of multiple encoders/decoders. See the [`mux` subpackage](#mux-subpackage) for details.
- **randRead**: Secure random byte generation for cryptographic use. See the [`randRead` subpackage](#randread-subpackage) for details.
- **sha256**: SHA-256 hashing and verification.

Refer to each subpackage's documentation for detailed usage, configuration, and examples.

### The `Coder` Interface

All subpackages implement the following interface:

```go
type Coder interface {
    Encode(p []byte) []byte
    Decode(p []byte) ([]byte, error)
    EncodeReader(r io.Reader) io.ReadCloser
    DecodeReader(r io.Reader) io.ReadCloser
    EncodeWriter(w io.Writer) io.WriteCloser
    DecodeWriter(w io.Writer) io.WriteCloser
    Reset()
}
```

This interface allows you to:

- Encode or decode data in memory (`[]byte`)
- Encode or decode data streams (`io.Reader`/`io.Writer`)
- Release resources with `Reset()`

### Usage Example

To use an encoding algorithm, import the relevant subpackage and instantiate its coder:

```go
import (
    "github.com/nabbar/golib/encoding/aes"
)

coder := aes.NewCoder(key)
encoded := coder.Encode([]byte("my data"))
decoded, err := coder.Decode(encoded)
```

### Notes

- Each subpackage provides its own constructor and configuration options.
- Always check for errors when decoding or working with streams.
- Use the `Reset()` method to free resources when done.

---

## `aes` Subpackage

The `aes` subpackage provides symmetric encryption and decryption using the AES-GCM algorithm. It implements the common `Coder` interface from the parent `encoding` package, allowing easy integration for secure data encoding/decoding in memory or via streams.

### Features

- AES-GCM encryption and decryption (256-bit key, 12-byte nonce)
- Secure random key and nonce generation
- Hexadecimal encoding/decoding for keys and nonces
- Implements the `Coder` interface for byte slices and `io.Reader`/`io.Writer`
- Thread-safe and stateless design
- Resource cleanup with `Reset()`

---

### Main Types & Functions

#### Key and Nonce Management

- `GenKey() ([32]byte, error)`: Generate a secure random 256-bit AES key.
- `GenNonce() ([12]byte, error)`: Generate a secure random 12-byte nonce.
- `GetHexKey(s string) ([32]byte, error)`: Decode a hex string to a 256-bit key.
- `GetHexNonce(s string) ([12]byte, error)`: Decode a hex string to a 12-byte nonce.

#### Creating a Coder

- `New(key [32]byte, nonce [12]byte) (encoding.Coder, error)`: Create a new AES-GCM coder instance with the given key and nonce.

#### Example Usage

```go
import (
    "github.com/nabbar/golib/encoding/aes"
)

key, _ := aes.GenKey()
nonce, _ := aes.GenNonce()
coder, err := aes.New(key, nonce)
if err != nil {
    // handle error
}
defer coder.Reset()

// Encrypt data
ciphertext := coder.Encode([]byte("my secret data"))

// Decrypt data
plaintext, err := coder.Decode(ciphertext)
```

#### Stream Encoding/Decoding

- `EncodeReader(r io.Reader) io.ReadCloser`: Returns a reader that encrypts data from `r`.
- `DecodeReader(r io.Reader) io.ReadCloser`: Returns a reader that decrypts data from `r`.
- `EncodeWriter(w io.Writer) io.WriteCloser`: Returns a writer that encrypts data to `w`.
- `DecodeWriter(w io.Writer) io.WriteCloser`: Returns a writer that decrypts data to `w`.

---

### Error Handling

- All decoding and stream operations may return errors (e.g., invalid buffer size, decryption failure).
- Always check errors when decoding or using stream interfaces.

---

### Notes

- The key must be 32 bytes (256 bits) and the nonce 12 bytes, as required by AES-GCM.
- Use `Reset()` to clear sensitive data from memory when done.
- For security, never reuse the same key/nonce pair for different data.

---

## `hexa` Subpackage

The `hexa` subpackage provides hexadecimal encoding and decoding utilities, implementing the common `Coder` interface from the parent `encoding` package. It allows you to encode and decode data as hexadecimal strings, both in memory and via streaming interfaces.

### Features

- Hexadecimal encoding and decoding for byte slices
- Stream encoding/decoding via `io.Reader` and `io.Writer`
- Implements the `Coder` interface for easy integration
- Error handling for invalid buffer sizes and decoding errors
- Stateless and thread-safe design

---

### Main Types & Functions

#### Creating a Coder

Instantiate a new hexadecimal coder:

```go
import (
    "github.com/nabbar/golib/encoding/hexa"
)

coder := hexa.New()
```

#### Encoding and Decoding

- `Encode(p []byte) []byte`: Encodes a byte slice to its hexadecimal representation.
- `Decode(p []byte) ([]byte, error)`: Decodes a hexadecimal byte slice back to its original bytes.

#### Stream Interfaces

- `EncodeReader(r io.Reader) io.ReadCloser`: Returns a reader that encodes data from `r` to hexadecimal.
- `DecodeReader(r io.Reader) io.ReadCloser`: Returns a reader that decodes hexadecimal data from `r`.
- `EncodeWriter(w io.Writer) io.WriteCloser`: Returns a writer that encodes data to hexadecimal and writes to `w`.
- `DecodeWriter(w io.Writer) io.WriteCloser`: Returns a writer that decodes hexadecimal data and writes to `w`.

#### Example Usage

```go
coder := hexa.New()

// Encode bytes
encoded := coder.Encode([]byte("Hello World"))

// Decode bytes
decoded, err := coder.Decode(encoded)

// Stream encoding
r := coder.EncodeReader(myReader)
defer r.Close()

// Stream decoding
w := coder.DecodeWriter(myWriter)
defer w.Close()
```

---

### Error Handling

- Decoding returns an error if the input is not valid hexadecimal.
- Stream operations may return errors for invalid buffer sizes or I/O issues.
- Use the `Reset()` method to release any resources (no-op for this stateless implementation).

---

### Notes

- The package is stateless and safe for concurrent use.
- Buffer sizes must be sufficient for encoding/decoding operations; otherwise, an error is returned.
- Always check errors when decoding or using stream interfaces.

---

## `mux` Subpackage

The `mux` subpackage provides multiplexing and demultiplexing utilities for encoding and decoding data streams over a single `io.Writer` or `io.Reader`. It allows you to send and receive data on multiple logical channels, identified by a key, through a single stream. This is useful for scenarios where you need to transmit different types of data or messages over the same connection.

### Features

- Multiplex multiple logical channels into a single stream
- Demultiplex a stream into multiple channels based on a key
- Channel identification using a `rune` key
- CBOR serialization and hexadecimal encoding for data blocks
- Thread-safe and efficient design
- Error handling for invalid channels and stream issues

---

### Main Types & Functions

#### Multiplexer

The `Multiplexer` interface allows you to create logical channels for writing data:

```go
type Multiplexer interface {
    NewChannel(key rune) io.Writer
}
```

- `NewChannel(key rune) io.Writer`: Returns an `io.Writer` for the given channel key. Data written to this writer is multiplexed into the main stream.

**Example:**

```go
import (
    "github.com/nabbar/golib/encoding/mux"
)

muxer := mux.NewMultiplexer(myWriter, '\n')
chA := muxer.NewChannel('a')
chB := muxer.NewChannel('b')

chA.Write([]byte("data for channel A"))
chB.Write([]byte("data for channel B"))
```

#### Demultiplexer

The `DeMultiplexer` interface allows you to register output channels and read data from the main stream:

```go
type DeMultiplexer interface {
    io.Reader
    Copy() error
    NewChannel(key rune, w io.Writer)
}
```

- `NewChannel(key rune, w io.Writer)`: Registers an `io.Writer` for a given channel key. Data for this key will be written to the provided writer.
- `Copy() error`: Continuously reads from the main stream and dispatches data to the correct channel writers. Intended to be run in a goroutine.

**Example:**

```go
dmx := mux.NewDeMultiplexer(myReader, '\n', 0)
bufA := &bytes.Buffer{}
bufB := &bytes.Buffer{}

dmx.NewChannel('a', bufA)
dmx.NewChannel('b', bufB)

go dmx.Copy()
// bufA and bufB will receive their respective data
```

#### Construction

- `NewMultiplexer(w io.Writer, delim byte) Multiplexer`: Creates a new multiplexer with the given writer and delimiter.
- `NewDeMultiplexer(r io.Reader, delim byte, size int) DeMultiplexer`: Creates a new demultiplexer with the given reader, delimiter, and buffer size.

---

### Data Format

Each data block is serialized using CBOR and includes:
- `K`: The channel key (`rune`)
- `D`: The data payload (hexadecimal encoded)

A delimiter byte is appended to each block to separate messages.

---

### Error Handling

- Returns errors for invalid instances or unknown channel keys.
- `Copy()` returns any error encountered during reading or writing, except for `io.EOF` which is ignored.

---

### Notes

- The package is suitable for use with network sockets, files, or any stream-based transport.
- Always register channels before calling `Copy()` on the demultiplexer.
- The delimiter should not appear in the encoded data.

---

## `randRead` Subpackage

The `randRead` subpackage provides a utility for creating a random byte stream reader from a remote or dynamic source. It is designed to wrap any function that returns an `io.ReadCloser` (such as an HTTP request or a cryptographic random source) and expose it as a buffered, reusable `io.ReadCloser` interface.

### Features

- Wraps any remote or dynamic byte stream as an `io.ReadCloser`
- Buffers data for efficient reading
- Automatically refreshes the underlying source when needed
- Thread-safe using atomic values
- Simple integration with any function returning `io.ReadCloser`

---

### Main Types & Functions

#### `FuncRemote` Type

A function type that returns an `io.ReadCloser` and an error:

```go
type FuncRemote func() (io.ReadCloser, error)
```

#### Creating a Random Reader

Use the `New` function to create a new random reader from a remote source:

```go
import "github.com/nabbar/golib/encoding/randRead"

reader := randRead.New(func() (io.ReadCloser, error) {
    // Return your io.ReadCloser here (e.g., HTTP response body, random source, etc.)
})
```

- The provided function will be called whenever the reader needs to fetch new data.

#### Example Usage

```go
import (
    "github.com/nabbar/golib/encoding/randRead"
    "crypto/rand"
    "io"
)

reader := randRead.New(func() (io.ReadCloser, error) {
    // Wrap crypto/rand.Reader as an io.ReadCloser
    return io.NopCloser(rand.Reader), nil
})

buf := make([]byte, 16)
n, err := reader.Read(buf)
// Use buf[0:n] as random data

_ = reader.Close()
```

---

### How It Works

- The random reader buffers data from the remote source using a `bufio.Reader`.
- If the buffer is empty or the underlying source is exhausted, it automatically calls the provided function to obtain a new `io.ReadCloser`.
- The reader is thread-safe and can be used concurrently.

---

### Error Handling

- If the remote function returns an error, the reader will propagate it on `Read`.
- Always check errors when reading or closing the reader.

---

### Notes

- The package is suitable for scenarios where you need a continuous or on-demand random byte stream from a remote or dynamic source.
- The underlying source is closed and replaced automatically as needed.
- Use `Close()` to release resources when done.

---

## `sha256` Subpackage

The `sha256` subpackage provides a simple and unified interface for computing SHA-256 hashes, implementing the common `Coder` interface from the parent `encoding` package. It allows you to hash data in memory or via streaming interfaces, making it easy to integrate SHA-256 hashing into your applications.

### Features

- SHA-256 hashing for byte slices and data streams
- Implements the `Coder` interface for compatibility with other encoding packages
- Stateless and thread-safe design
- Resource cleanup with `Reset()`
- Stream support for `io.Reader` and `io.Writer` (encoding only)

---

### Main Types & Functions

#### Creating a Coder

Instantiate a new SHA-256 coder:

```go
import (
    "github.com/nabbar/golib/encoding/sha256"
)

coder := sha256.New()
```

#### Hashing Data

- `Encode(p []byte) []byte`: Computes the SHA-256 hash of the input byte slice and returns the hash as a byte slice.
- `Decode(p []byte) ([]byte, error)`: Not supported; always returns an error (SHA-256 is not reversible).

#### Stream Interfaces

- `EncodeReader(r io.Reader) io.ReadCloser`: Returns a reader that passes data through and updates the hash state. Use `Encode(nil)` after reading to get the hash.
- `EncodeWriter(w io.Writer) io.WriteCloser`: Returns a writer that writes data to `w` and updates the hash state. Use `Encode(nil)` after writing to get the hash.
- `DecodeReader(r io.Reader) io.ReadCloser`: Not supported; always returns `nil`.
- `DecodeWriter(w io.Writer) io.WriteCloser`: Not supported; always returns `nil`.

#### Example Usage

```go
coder := sha256.New()

// Hash a byte slice
hash := coder.Encode([]byte("Hello World"))

// Use with io.Writer
buf := &bytes.Buffer{}
w := coder.EncodeWriter(buf)
w.Write([]byte("Hello World"))
w.Close()
hash = coder.Encode(nil) // Get the hash after writing

// Use with io.Reader
r := coder.EncodeReader(bytes.NewReader([]byte("Hello World")))
io.ReadAll(r)
r.Close()
hash = coder.Encode(nil) // Get the hash after reading
```

#### Resetting State

- `Reset()`: Resets the internal hash state, allowing reuse of the coder for new data.

---

### Error Handling

- `Decode` and stream decoding methods are not supported and will return an error or `nil`.
- Always check for errors when using unsupported methods.

---

### Notes

- SHA-256 is a one-way hash function; decoding is not possible.
- The package is stateless and safe for concurrent use.
- Use `Reset()` to clear the internal state before reusing the coder.

