# OpenPGP Package

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.27-blue)](https://go.dev/doc/install)

Post-Quantum and classical OpenPGP key management and streaming encryption/decryption library. The root package `encoding/opengpg` provides a single algorithm-agnostic engine (the `OpenGPG` interface and its `packet.Config` driven implementation), and the six sub-packages are thin presets that bind this engine to a specific algorithm: five Post-Quantum hybrid presets (**ML-DSA65 + Ed25519**, **ML-DSA87 + Ed448**, **SLH-DSA-SHAKE128f + Ed25519**, **SLH-DSA-SHAKE128s + Ed25519**, **SLH-DSA-SHAKE256s + Ed448**) and one classical preset (**RSA-4096 + SHA-256**). Built on top of [ProtonMail/go-crypto/openpgp](https://github.com/ProtonMail/go-crypto), it adds deterministic key lifetime management, resource cleanup via `mapCloser`, and memory-efficient streaming through `io.Pipe` with `sync.Pool` buffer reuse.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Subpackages](#subpackages)
    - [mldsa65ed](#mldsa65ed)
    - [mldsa87ed](#mldsa87ed)
    - [slhdsa128f](#slhdsa128f)
    - [slhdsa128s](#slhdsa128s)
    - [slhdsa256s](#slhdsa256s)
    - [rsa4096sha256](#rsa4096sha256)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [Resources](#resources)

---

## Overview

The `encoding/opengpg` root package contains the **complete implementation**. It is not a pure contract package: it defines the `OpenGPG` interface, the `Identity` structure, the typed error registry, and the concrete unexported type `mod` that implements the whole key-management and streaming API. The algorithm is *not* hardcoded — it is selected by the caller through `Options.Algorithm` (plus `Options.Hash` and `Options.Cipher`), which map directly onto `packet.Config`.

The six sub-packages are therefore **preset constructors only** (a single `interface.go` file each). The five Post-Quantum presets expose a reduced `Options` struct (`Rand`, `Time`, `KeyTime`) and a `New(ctx, Options) opengpg.OpenGPG` function that delegates to `opengpg.New` with a fixed algorithm triple (SHA-3/512 hash, AES-256 cipher, one Post-Quantum `packet.PublicKeyAlgorithm`). The `rsa4096sha256` preset uses the same reduced `Options` struct but delegates with a fixed classical algorithm triple (SHA-256 hash, AES-256 cipher, `packet.PubKeyAlgoRSA` with 4096-bit key).

| Sub-package | `packet.PublicKeyAlgorithm` | Sign primary key | Encryption subkey |
|-------------|------------------------------|------------------|-------------------|
| `mldsa65ed` | `PubKeyAlgoMldsa65Ed25519` (30) | ML-DSA-65 + Ed25519 | ML-KEM-768 + X25519 |
| `mldsa87ed` | `PubKeyAlgoMldsa87Ed448` (31) | ML-DSA-87 + Ed448 | ML-KEM-1024 + X448 |
| `slhdsa128f` | `PubKeyAlgoSlhdsaShake128f` (33) | SLH-DSA-SHAKE128f + Ed25519 | ML-KEM-768 + X25519 |
| `slhdsa128s` | `PubKeyAlgoSlhdsaShake128s` (32) | SLH-DSA-SHAKE128s + Ed25519 | ML-KEM-768 + X25519 |
| `slhdsa256s` | `PubKeyAlgoSlhdsaShake256s` (34) | SLH-DSA-SHAKE256s + Ed448 | ML-KEM-1024 + X448 |
| `rsa4096sha256` | `PubKeyAlgoRSA` (1) | RSA-4096 | RSA-4096 |

The five Post-Quantum presets share the same wire parameters: **OpenPGP V6 keys**, **SHA-3/512** as `DefaultHash`, **AES-256** as `DefaultCipher`, and **compression disabled** (`packet.CompressionNone`). The encryption subkey pairing is not configured here: `V6Keys: true` is always set, and go-crypto derives the matching ML-KEM subkey automatically from the primary signature algorithm (`GetMatchingMlkem`).

The `rsa4096sha256` preset uses **OpenPGP V4 keys**, **SHA-256** as `DefaultHash`, **AES-256** as `DefaultCipher`, and **compression disabled** (`packet.CompressionNone`). Key size is fixed at 4096 bits.

### Design Philosophy

1. **One engine, many presets**: the algorithm is a *configuration* (`packet.PublicKeyAlgorithm`) rather than a compile-time branch. The sub-packages contain no duplicated logic — they only fill in three fields — so behaviour, error semantics and resource lifecycle are guaranteed identical across all six presets.
2. **Deterministic resource management**: every pipe end and every streaming writer created during a streaming operation is registered in a `mapCloser.Closer` owned by the instance, and released when `Close()` is called, so no goroutine is left blocked on an orphaned pipe.
3. **Memory-efficient streaming**: `EncryptReader` and `DecryptWriter` run the payload transfer in a background goroutine fed by an `io.Pipe`, copying with a 32 KB buffer taken from a `sync.Pool`. Buffer contents are wiped with `clear()` before being returned to the pool so plaintext or ciphertext does not linger in memory.
4. **Typed error registry**: errors flow through `golib/errors` with a per-package code range (`errors.MinPkgEncodingOpenGPG`), enabling callers to branch on codes rather than on message strings.

### Key Features

- Post-Quantum hybrid key generation (ML-DSA or SLH-DSA for signatures, ML-KEM for key encapsulation)
- OpenPGP V6 key format, SHA-3/512 digest, AES-256 symmetric cipher, no compression
- Streaming encryption and decryption through `io.Pipe` + pooled buffers
- Four I/O shapes: `EncryptReader`, `EncryptWriter`, `DecryptReader`, `DecryptWriter`
- ASCII-armored **and** raw binary key import/export (`parseKeyRing` tries both)
- Configurable key lifetime through `Options.KeyTime`
- Built-in self-test (`Check()`) that performs a full encrypt-then-decrypt round-trip
- Atomic closed flag making the instance single-use and explicitly closable

### Key Benefits

- **Post-Quantum confidentiality and authenticity**: ML-KEM encapsulation and ML-DSA / SLH-DSA signatures provide quantum resilience while remaining hybrid with classical X25519/X448 and Ed25519/Ed448 components.
- **O(1) memory relative to payload size**: the streaming paths never materialise the payload, so a multi-gigabyte file flows through a single 32 KB pooled buffer per goroutine.
- **Uniform API across six presets**: switching from ML-DSA-65 to RSA-4096 is a one-line import and constructor change; the interface, the `Identity` layout and the error codes are unchanged.
- **Explicit lifecycle**: `Close()` is atomic and idempotent, and nils out the entity list and the `packet.Config` so key material becomes garbage-collectable.

---

## Architecture

### Package Structure

```
encoding/opengpg/                           # root package — engine, contract, errors
├── interface.go                            # Identity, OpenGPG interface, Options, New()
├── model.go                                # concrete implementation (mod, waitCloser)
├── tools.go                                # parseKeyRing helper (armored + binary)
├── errors.go                               # Error code constants and registry with golib/errors
├── benchmark_test.go                       # BenchmarkStreamReader / BenchmarkStreamWriter
├── helper_test.go                          # Test helpers (getNewOpenGPGP, buffer pool)
├── model_test.go                           # Unit tests for model methods
├── opengpg_suite_test.go                   # Ginkgo test suite
│
├── mldsa65ed/                              # ML-DSA65 + Ed25519 preset
│   ├── interface.go                        # Options struct and New() constructor
│   ├── benchmark_test.go                   # Benchmark tests
│   ├── helper_test.go                      # Test helpers
│   ├── mldsa65ed_suite_test.go             # Ginkgo test suite
│   └── model_test.go                       # Unit tests
│
├── mldsa87ed/                              # ML-DSA87 + Ed448 preset
│   ├── interface.go                        # Options struct and New() constructor
│   ├── benchmark_test.go                   # Benchmark tests
│   ├── helper_test.go                      # Test helpers
│   ├── mldsa87ed_suite_test.go             # Ginkgo test suite
│   └── model_test.go                       # Unit tests
│
├── slhdsa128f/                             # SLH-DSA-SHAKE128f + Ed25519 preset
│   ├── interface.go                        # Options struct and New() constructor
│   ├── benchmark_test.go                   # Benchmark tests
│   ├── helper_test.go                      # Test helpers
│   ├── slhdsa128f_suite_test.go            # Ginkgo test suite
│   └── model_test.go                       # Unit tests
│
├── slhdsa128s/                             # SLH-DSA-SHAKE128s + Ed25519 preset
│   ├── interface.go                        # Options struct and New() constructor
│   ├── benchmark_test.go                   # Benchmark tests
│   ├── helper_test.go                      # Test helpers
│   ├── slhdsa128s_suite_test.go            # Ginkgo test suite
│   └── model_test.go                       # Unit tests
│
├── slhdsa256s/                             # SLH-DSA-SHAKE256s + Ed448 preset
    │   ├── interface.go                    # Options struct and New() constructor
    │   ├── benchmark_test.go               # Benchmark tests
    │   ├── helper_test.go                  # Test helpers
    │   ├── slhdsa256s_suite_test.go        # Ginkgo test suite
    │   └── model_test.go                   # Unit tests
    │
└── rsa4096sha256/                          # RSA-4096 + SHA-256 classical preset
    ├── interface.go                        # Options struct and New() constructor
    ├── benchmark_test.go                   # Benchmark tests
    ├── helper_test.go                      # Test helpers
    ├── model_test.go                       # Unit tests
    └── rsa4096sha256_suite_test.go         # Ginkgo test suite
```

### Package Architecture

```
                 ┌───────────────────────────────┐
                 │         Consumer Code         │
                 │       (any Go application)    │
                 └───────────────┬───────────────┘
                                 │ uses
                                 ▼
        ┌─────────────────────────────────────────────────┐
        │  sub-package preset (mldsa65ed, slhdsa256s, …)  │
        │  Options{Rand, Time, KeyTime} + New()           │
        │  pins: Hash, Cipher, Algorithm                  │
        └───────────────────────┬─────────────────────────┘
                                │ delegates to
                                ▼
        ┌─────────────────────────────────────────────────┐
        │            opengpg (root package)               │
        │  New(ctx, Options{Rand,Time,KeyTime,            │
        │                     Hash,Cipher,Algorithm})     │
        │  ─────────────────────────────────────────      │
        │  Identity      — user id + armored key material │
        │  OpenGPG       — the contract (interface)       │
        │  mod           — Create/Load/Check              │
        │                  Encrypt/Decrypt Reader/Writer  │
        │  parseKeyRing  — armored / binary key decoding  │
        │  errors        — typed error code registry      │
        └───────┬───────────────────────────┬─────────────┘
                │ depends on                │ depends on
                ▼                           ▼
┌────────────────────────────────┐  ┌────────────────────────────────┐
│ ProtonMail/go-crypto           │  │ golib/errors                   │
│  openpgp + openpgp/packet      │  │  (error code registry)         │
│  (V6 keys, hybrid PQ algos)    │  ├────────────────────────────────┤
│  + golib/ioutils/mapCloser     │  │ golib/ioutils/mapCloser        │
│  (tracked resource cleanup)    │  │  (pipe / writer registration)  │
└────────────────────────────────┘  └────────────────────────────────┘
```

### Dataflow

```
┌──────────────────────────────────────────────────────────────┐
│                     Instance Construction                    │
├──────────────────────────────────────────────────────────────┤
│  New(ctx, Options)                                           │
│    └── packet.Config{                                        │
│          V6Keys: true,                                       │
│          Rand, Time,                                         │
│          DefaultHash, DefaultCipher, Algorithm,              │
│          DefaultCompressionAlgo: CompressionNone,            │
│          KeyLifetimeSecs: KeyTime<1 ? MaxUint32 : KeyTime }  │
│    └── mod{ o: cfg, i: nil, c: false,                        │
│             l: mapCloser.New(ctx), p: sync.Pool(32 KB) }     │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                       Key Management                         │
├──────────────────────────────────────────────────────────────┤
│  Create(Identity):                                           │
│    NewEntity(Name, Comment, Email, cfg)                      │
│      ├── armor(PublicKeyType)  ──► Identity.PublicKey        │
│      └── armor(PrivateKeyType) ──► Identity.PrivateKey       │
│    └── o.i = EntityList{entity}                              │
│                                                              │
│  Load(Identity):                                             │
│    PrivateKey (preferred) or PublicKey                       │
│      └── parseKeyRing: ReadArmoredKeyRing → ReadKeyRing      │
│    └── o.i = parsed entity list                              │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                       Streaming I/O                          │
├──────────────────────────────────────────────────────────────┤
│  EncryptReader(r):  [r] →(goroutine, 32 KB pool)→ [pipe] → R │
│  DecryptReader(r):  [r] → ReadMessage → UnverifiedBody → R   │
│  EncryptWriter(w):  [W] ──────────────→ sdkpgp.Encrypt → w   │
│  DecryptWriter(w):  [W] →(pipe)→ ReadMessage →(32 KB)→ w     │
│  pipe ends & EncryptWriter results are registered in o.l     │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                      Resource Cleanup                        │
├──────────────────────────────────────────────────────────────┤
│  Close():                                                    │
│    c.Store(true)      → reject new Create/Load/Check/        │
│                         EncryptReader with os.ErrClosed      │
│    l.Close()          → close tracked pipes and writers      │
│    o.i = nil          → drop entity list (GC eligible)       │
│    o.o = nil          → drop packet.Config                   │
│    return nil         → always, Close is idempotent          │
└──────────────────────────────────────────────────────────────┘
```

### Data Fields Reference

#### `opengpg.Options` (root constructor)

| Field         | Type                        | Description                                                                                                     |
|---------------|-----------------------------|-----------------------------------------------------------------------------------------------------------------|
| `Rand`        | `io.Reader`                 | Cryptographic random source passed to `packet.Config.Rand`; `nil` lets go-crypto use `crypto/rand.Reader`          |
| `Time`        | `func() time.Time`          | Clock override forwarded to `packet.Config.Time`; `nil` uses the system clock                                     |
| `KeyTime`     | `uint32`                    | Key lifetime **in seconds** (not a Unix timestamp). `0` → `math.MaxUint32` (~136 years); otherwise used verbatim as `KeyLifetimeSecs` |
| `Hash`        | `crypto.Hash`               | `packet.Config.DefaultHash`. Post-Quantum presets set `crypto.SHA3_512`; `rsa4096sha256` sets `crypto.SHA256`   |
| `Cipher`      | `packet.CipherFunction`     | `packet.Config.DefaultCipher`. All presets set `packet.CipherAES256`                                             |
| `Algorithm`   | `packet.PublicKeyAlgorithm` | Primary key algorithm. Post-Quantum presets set a hybrid PQ algorithm; `rsa4096sha256` sets `PubKeyAlgoRSA`       |

#### `opengpg.Identity`

| Field         | Type     | Description                                                                                        |
|---------------|----------|----------------------------------------------------------------------------------------------------|
| `Name`        | `string` | User name portion of the OpenPGP user ID; input to `NewEntity`                                    |
| `Comment`     | `string` | Optional comment portion of the user ID; input to `NewEntity`                                     |
| `Email`       | `string` | Email portion of the user ID; input to `NewEntity`                                                |
| `PublicKey`   | `[]byte` | Input to `Load`, and written by `Create`: ASCII-armored `PUBLIC KEY BLOCK` bytes                  |
| `PrivateKey`  | `[]byte` | Input to `Load` (preferred), and written by `Create`: ASCII-armored `PRIVATE KEY BLOCK` bytes     |

#### Preset `Options` (all six sub-packages)

All sub-packages expose exactly the same reduced `Options` struct — only `Rand`, `Time` and `KeyTime` are configurable; hash, cipher and algorithm are fixed by the preset.

| Field     | Type               | Description                                                       |
|-----------|--------------------|-------------------------------------------------------------------|
| `Rand`    | `io.Reader`        | Cryptographic random source; `nil` → `crypto/rand.Reader`          |
| `Time`    | `func() time.Time` | Clock override; `nil` → system clock                               |
| `KeyTime` | `uint32`           | Key lifetime in seconds; `0` → `math.MaxUint32` (~136 years)       |

#### Error Codes

Codes start at `liberr.MinPkgEncodingOpenGPG` and are registered with the `golib/errors` global registry in `errors.go`.

| Constant                | Message                              | Returned by |
|-------------------------|--------------------------------------|-------------|
| `ErrorParamEmpty`       | "required parameter is nil or empty" | Reserved base code — registered in the registry but not currently returned by any method |
| `ErrorIdentityInvalid`  | "identity is empty or invalid"       | `Create(nil)`, `Load(nil)`, `Load` with both key fields empty, `Check`/`Encrypt*`/`Decrypt*` with no key loaded |
| `ErrorPublicKeyInvalid` | "public key is empty or invalid"    | `Create` on public-key serialization failure, `Load` on public-key parse failure, `EncryptWriter` when `sdkpgp.Encrypt` fails |
| `ErrorPrivateKeyInvalid`| "private key is empty or invalid"   | `Create` on private-key serialization failure, `Load` on private-key parse failure |

When a cause exists it is attached as a parent (`ErrorX.Error(err)`), so the hierarchy is preserved:

```go
var err error = opengpg.ErrorPrivateKeyInvalid.Error(cause)

// Standard sentinels still work through the chain.
errors.Is(err, os.ErrNotExist) // → true

// Code inspection uses the golib/errors Error interface, since a
// CodeError constant is not itself an error value.
var le liberr.Error
if errors.As(err, &le) {
    le.IsCode(opengpg.ErrorPrivateKeyInvalid)  // this error only
    le.HasCode(opengpg.ErrorPrivateKeyInvalid) // this error or any parent
    le.HasParent()                             // is a cause attached?
}
```

Standard library sentinels are also returned directly, unwrapped, in several places: `os.ErrClosed` (see [Best Practices](#best-practices)), and `os.ErrInvalid` when the `Check()` round-trip does not reproduce the payload byte for byte. `parseKeyRing` returns `os.ErrNotExist` for an empty buffer and `os.ErrInvalid` when neither the armored nor the binary parser yields an entity; `Load` always wraps both into `ErrorPrivateKeyInvalid` or `ErrorPublicKeyInvalid`.

---

## Performance

Both streaming paths use `io.Pipe` plus a 32 KB `sync.Pool` buffer, so memory usage is O(1) with respect to payload size.

- **Memory**: `EncryptReader` and `DecryptWriter` never accumulate the payload. A 1 GB input is processed with one 32 KB pooled buffer per streaming goroutine.
- **Buffer reuse**: buffers are drawn from a per-instance `sync.Pool`, wiped with `clear()`, and returned to the pool. `EncryptWriter` and `DecryptReader` do not use the pool — they write straight to the destination through go-crypto.
- **`Check()` cost**: the self-test payload is `strings.Repeat(chkStringBase, 4*1024)`, i.e. 4 096 repetitions of a 41-byte string (~164 KiB) held entirely in memory. It exercises the full symmetric path on every call, so it is the most expensive method on small payloads.
- **Signature cost**: SLH-DSA is hash-based with `h = 63…66` hypertree layers, so it is far more expensive to *sign* than ML-DSA while producing much smaller public keys. `slhdsa128f` signs faster than `slhdsa128s` but yields larger signatures (31 520 B vs 10 768 B); `slhdsa256s` is the slowest and is the only variant paired with the ML-KEM-1024 / X448 subkey. This cost is **absent from the streaming benchmarks below**, which do not sign — see the note there.
- **Hash/cipher**: SHA-3/512 is software-only and slower than SHA-2 but is the digest required by these hybrid algorithms; AES-256 benefits from AES-NI where available.

### Measured Encryption / Decryption Operations Per Second

The root-level benchmarks measure the rate of complete encrypt-then-decrypt round-trips per second, using a small 32 KB test payload. Because the symmetric layer (AES-256 + SHA-3/512) is constant across all presets, the measured difference isolates the **Post-Quantum KEM cost**: ML-KEM-768/X25519 vs ML-KEM-1024/X448.

> **Measurement environment** — results below were produced on a single machine and are indicative only. They are *not* a specification and should not be used for capacity planning.
>
> | | |
> |---|---|
> | CPU | AMD Ryzen 9 9900X3D — 12 cores / 24 threads, Zen 5 (family 26, model 68), 5.58 GHz max boost |
> | Memory | 60 GiB |
> | Vector ISA | **AVX-512 enabled** — `avx512f`, `avx512dq`, `avx512cd`, `avx512bw`, `avx512vl`, `avx512ifma`, `avx512vbmi`, `avx512vbmi2`, `avx512vnni`, `avx512bitalg`, `avx512bf16`, `avx512vpopcntdq`, `avx512vp2intersect` (plus `vaes` and `aes` for the symmetric layer) |
> | Go | go1.27.1 linux/amd64, `GOAMD64=v4` default |
> | Date | 2026-10-01 |

Each benchmark performs a full **encrypt *and* decrypt round trip** over a 32 KB payload. Go's default wall-clock budget determines the number of iterations per benchmark, which makes the disparity in iteration count between KEM families immediately visible. Reproduce with:

```bash
go test -run XXX -bench . -benchmem
```

| Preset | KEM / ECDH | Read ns/op | Read ops/s | Read B/op | Read allocs/op | Write ns/op | Write ops/s | Write B/op | Write allocs/op |
|---------|------------------|-----------------:|-----------------:|------------:|-----------------:|-----------------:|-----------------:|------------:|-----------------:|
| `mldsa65ed / ML-KEM-768 + X25519` | ML-KEM-768/X25519 | 71 243 | 14 036 | 10 420 | 100 | 74 325 | 13 454 | 10 638 | 105 |
| `mldsa87ed / ML-KEM-1024 + X448` | ML-KEM-1024/X448 | 231 076 | **4 328** | 11 924 | 100 | 235 549 | **4 245** | 12 110 | 105 |
| `slhdsa128f / ML-KEM-768 + X25519` | ML-KEM-768/X25519 | 70 514 | 14 182 | 10 431 | 100 | 73 828 | 13 545 | 10 627 | 105 |
| `slhdsa128s / ML-KEM-768 + X25519` | ML-KEM-768/X25519 | 69 619 | 14 364 | 10 423 | 100 | 74 415 | 13 438 | 10 630 | 105 |
| `slhdsa256s / ML-KEM-1024 + X448` | ML-KEM-1024/X448 | 230 084 | **4 346** | 11 927 | 100 | 235 507 | **4 246** | 12 111 | 105 |
| `rsa4096sha256 / RSA-4096 + SHA2-256` | RSA-4096 | 3 669 413 | **272.5** | 79 045 | 196 | 3 665 073 | **272.8** | 78 696 | 200 |

**Two distinct performance tiers emerge.** The data cleanly separates along the KEM/ECDH pair:

- **Tier 1 — ML-KEM-768 + X25519** (`mldsa65ed`, `slhdsa128f`, `slhdsa128s`): ~14 300 ops/s on Read, ~13 500 ops/s on Write. These three presets are nearly indistinguishable from each other, confirming that the signature algorithm (ML-DSA-65 vs SLH-DSA) has negligible impact on the encryption/decryption round-trip.
- **Tier 2 — ML-KEM-1024 + X448** (`mldsa87ed`, `slhdsa256s`): ~4 300 ops/s on Read, ~4 200 ops/s on Write. Roughly **3×</ slower** than the ML-KEM-768/X25519 tier.
- **Non-PQC — Legacy** (`rsa4096sha256`): ~272 ops/s on Read, ~273 ops/s on Write. Roughly **16×</ slower** than the ML-KEM-1024/X448 tier and **50×</ slower** than the ML-KEM-768/X25519 tier.

This difference is structural, not noise. ML-KEM-1024 uses larger matrices (1 568 B public key, 1 568 B ciphertext vs 1 184 B and 1 088 B for ML-KEM-768), and X448 key agreement is inherently more expensive than X25519. Together they make the per-operation overhead of the ML-KEM-1024/X448 pair roughly three times that of ML-KEM-768/X25519.

The iteration count difference is equally telling. Go's default wall-clock budget allows the ML-KEM-768 presets to run ~17 000 iterations each, while the ML-KEM-1024 presets only manage ~5 000 — again confirming a ~3× disparity.

**About the `B/op` column**: the ML-KEM-1024 presets show slightly higher per-operation memory (≈11 900 B vs ≈10 400 B) because the KEM ciphertext and keys are larger. The RSA-4096 preset shows significantly higher memory per operation (≈79 000 B) due to the much larger RSA key material processed per operation. The RSA preset also shows more allocations (196/200 vs 100/105) because RSA key operations involve more internal object creation.

### Key and Signature Sizes

Values below are the raw primitive sizes, before OpenPGP packet framing, and are the main trade-off between the presets.

| Algorithm             | Public key | Private key | Signature | Ciphertext | Armor Public key | Armor Private key |
|-----------------------|------------|-------------|-----------|------------|------------------|-------------------|
| ML-DSA-65             | 1 952 B    | 4 032 B     | 3 309 B   | —          | 18 616 B         | 18 837 B          |
| ML-DSA-87             | 2 592 B    | 4 896 B     | 4 627 B   | —          | 25 628 B         | 25 914 B          |
| SLH-DSA-SHAKE128f     | 32 B       | 64 B        | 31 520 B  | —          | 71 700 B         | 71 922 B          |
| SLH-DSA-SHAKE128s     | 32 B       | 64 B        | 10 768 B  | —          | 34 183 B         | 34 405 B          |
| SLH-DSA-SHAKE256s     | 64 B       | 128 B       | 29 024 B  | —          | 123 908 B        | 124 247 B         |
| ML-KEM-768            | 1 184 B    | 2 400 B     | —         | 1 088 B    | —                | —                 |
| ML-KEM-1024           | 1 568 B    | 3 168 B     | —         | 1 568 B    | —                | —                 |
| RSA-4096              | 512 B      | 3 277 B     | 512 B     | —          | 3 284 B          | 6 780 B           |

*Notes: the SLH-DSA sizes are for the SLH-DSA component only — the OpenPGP primary key additionally carries the classical Ed25519 (32 B) or Ed448 (57 B) key. ML-KEM private key sizes come from the CIRCL implementation and are indicative.*

Combined size of the Post-Quantum component plus its classical counterpart, which is what a preset's primary key actually costs on the wire:

| Preset | Primary public key | Primary signature |
|---------|-------------------:|------------------:|
| `slhdsa128f` | 32 + 32 = **64 B** | 31 520 B |
| `slhdsa128s` | 32 + 32 = **64 B** | 10 768 B |
| `slhdsa256s` | 64 + 57 = **121 B** | 29 024 B |
| `mldsa65ed` | 1 952 + 32 = **1 984 B** | 3 309 B |
| `mldsa87ed` | 2 592 + 57 = **2 649 B** | 4 627 B |
| `rsa4096sha256` | 512 + 32 = **544 B** | 512 B |

---

#### Understanding OpenPGP Key Bundle File Sizes (`.asc`)

In practice, saving an OpenPGP entity to disk or exporting a public key generates a **complete key bundle**, not an isolated key primitive. An OpenPGP v6 key file includes:

1. **Primary Master Key** (Certification/Signing — PQC + Classical hybrid)
2. **User ID Packet** (Identity metadata)
3. **User ID Certification Signature** (Primary key self-signature)
4. **Encryption Subkey** (KEM — PQC + Classical hybrid)
5. **Subkey Binding Signature** (Self-signature by Primary key linking the subkey)
6. **ASCII Armor Framing** (Base64 encoding + OpenPGP packet headers: ~33% overhead)

Because **self-signatures are mandatory** to bind subkeys, algorithms with large signatures (such as `SLH-DSA`) produce significantly larger exported `.asc` key files, even if their raw public key primitive is tiny.

---

## Subpackages

Every sub-package has the same shape: an `Options` struct, a `New(ctx, Options) opengpg.OpenGPG` constructor, and a test suite. None of them contain encryption logic — they delegate to the root package. Import the one that matches your security requirement and your latency budget; nothing else in your code changes. The five Post-Quantum presets use V6 keys, SHA-3/512 and AES-256. The `rsa4096sha256` preset uses V4 keys, SHA-256 and AES-256.

### mldsa65ed

ML-DSA-65 primary signature key paired with Ed25519, and an ML-KEM-768 + X25519 encryption subkey. ML-DSA-65 is the NIST category 3 parameter set (roughly AES-192-level classical security). The ML-KEM-768/X25519 pair places this preset in the faster performance tier: ~14 300 ops/s on encrypt-decrypt round-trips, roughly 3×</ faster than the ML-KEM-1024/X448 tier. Key generation and signing are also among the cheapest of all presets, at the cost of the largest signature among the ML-DSA variants (3 309 B).

**Key Features**:
- `packet.PubKeyAlgoMldsa65Ed25519`, SHA-3/512, AES-256, OpenPGP V6
- Smallest ML-DSA key material of the two ML-DSA presets
- ML-KEM-768/X25519 — fastest performance tier (~3×</ vs ML-KEM-1024/X448)
- Default choice when throughput matters more than signature size

**Signature**: `func New(ctx context.Context, opt Options) opengpg.OpenGPG`

---

### mldsa87ed

ML-DSA-87 primary signature key paired with Ed448, and an ML-KEM-1024 + X448 encryption subkey. ML-DSA-87 is the NIST category 5 parameter set (roughly AES-256-level). The ML-KEM-1024/X448 pair places this preset in the slower performance tier: ~4 300 ops/s on encrypt-decrypt round-trips, roughly 3×</ slower than the ML-KEM-768/X25519 tier. The larger X448 and ML-KEM-1024 subkey requires V6 keys, which `New` always enables.

**Key Features**:
- `packet.PubKeyAlgoMldsa87Ed448`, SHA-3/512, AES-256, OpenPGP V6
- Strongest ML-KEM parameter set (ML-KEM-1024), largest ciphertext (1 568 B)
- Higher security level at the cost of larger key material and slower operations — ML-KEM-1024/X448 is ~3× slower than ML-KEM-768/X25519

**Signature**: `func New(ctx context.Context, opt Options) opengpg.OpenGPG`

---

### slhdsa128f

SLH-DSA-SHAKE128f primary signature key paired with Ed25519, and an ML-KEM-768 + X25519 encryption subkey. SLH-DSA is based on hash-based signatures, so its security assumptions rely only on the hash function. The `f` variant has the shortest tree (`h = 66`, `d = 22`), which makes it the fastest of the three SLH-DSA presets, but produces very large signatures (31 520 B).

**Key Features**:
- `packet.PubKeyAlgoSlhdsaShake128f`, SHA-3/512, AES-256, OpenPGP V6
- Public key of only 32 bytes — the smallest key material of all six presets
- Fastest SLH-DSA variant; use when signature size is not a constraint

**Signature**: `func New(ctx context.Context, opt Options) opengpg.OpenGPG`

---

### slhdsa128s

SLH-DSA-SHAKE128s primary signature key paired with Ed25519, and an ML-KEM-768 + X25519 encryption subkey. The `s` variant uses the short tree (`h = 63`, `d = 7`), trading signing time for a signature roughly three times smaller than `slhdsa128f` (10 768 B) — a good balance when messages are signed frequently.

**Key Features**:
- `packet.PubKeyAlgoSlhdsaShake128s`, SHA-3/512, AES-256, OpenPGP V6
- Smallest SLH-DSA signature (10 768 B) among the three presets
- Recommended default when SLH-DSA is required

**Signature**: `func New(ctx context.Context, opt Options) opengpg.OpenGPG`

---

### slhdsa256s

SLH-DSA-SHAKE256s primary signature key paired with Ed448, and an ML-KEM-1024 + X448 encryption subkey — the only SLH-DSA preset offered with the strongest KEM parameter set. The ML-KEM-1024/X448 pair places this preset in the slower performance tier: ~4 300 ops/s on encrypt-decrypt round-trips. Combined with the largest SLH-DSA public key (64 B) and signature (29 024 B), it is the most resource-intensive preset. Highest security level, at the price of the slowest key generation, signing, and encryption.

**Key Features**:
- `packet.PubKeyAlgoSlhdsaShake256s`, SHA-3/512, AES-256, OpenPGP V6
- SLH-DSA category 5 (256-bit) hash-based security plus ML-KEM-1024 / X448
- Slowest preset in both the ML-KEM-1024/X448 encryption tier and SLH-DSA signing cost; appropriate for low-frequency, long-lived keys

**Signature**: `func New(ctx context.Context, opt Options) opengpg.OpenGPG`

---

### rsa4096sha256

RSA-4096 primary signing key paired with an RSA-4096 encryption subkey, using SHA-256 hashing and AES-256 encryption. The preset uses OpenPGP V4 keys since RSA only supports V4 in the underlying library (V6 is reserved for Post-Quantum algorithms). Both key material and signatures are larger than the Post-Quantum presets, and key generation is the most expensive operation. This preset exists for backward compatibility, tooling interoperability, and environments where Post-Quantum algorithms are not yet available.

**Key Features**:
- `packet.PubKeyAlgoRSA` (1), SHA-256, AES-256, OpenPGP V4
- Universally supported — every OpenPGP implementation understands RSA keys
- Only classical (non–post-quantum) preset; use when Post-Quantum is not yet viable

**Signature**: `func New(ctx context.Context, opt Options) opengpg.OpenGPG`

---

## Use Cases

### 1. Post-Quantum Key Management

Generate and manage Post-Quantum OpenPGP key pairs with a configurable lifetime, suitable for environments requiring FIPS 203/204 compliance.

**References**: [FIPS 203](https://csrc.nist.gov/pubs/fips/203/final), [FIPS 204](https://csrc.nist.gov/pubs/fips/204/final), [RFC 9580 — The OpenPGP Message Format](https://datatracker.ietf.org/doc/html/rfc9580)

```go
package main

import (
	"context"

	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/encoding/opengpg/slhdsa128s"
)

func main() {
	ctx := context.Background()

	// 0 => no practical expiration (KeyLifetimeSecs = math.MaxUint32)
	gpg := slhdsa128s.New(ctx, slhdsa128s.Options{
		KeyTime: 0,
	})
	defer func() {
		_ = gpg.Close()
	}()

	identity := &opengpg.Identity{
		Name:    "Alice",
		Comment: "Post-Quantum Key",
		Email:   "alice@example.com",
	}

	if err := gpg.Create(identity); err != nil {
		panic(err)
	}

	// identity.PublicKey and identity.PrivateKey now hold
	// ASCII-armored key material.
}
```

### 2. Streaming Encryption for Large Files

Encrypt a large data stream (file, network connection) without loading the full payload into memory. `EncryptReader` runs the transfer in a background goroutine over an `io.Pipe`, so memory stays flat regardless of the input size.

```go
package main

import (
	"context"
	"io"
	"os"

	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/encoding/opengpg/mldsa87ed"
)

func main() {
	ctx := context.Background()

	// Load an existing key from stored armored material.
	identity := &opengpg.Identity{
		Name:      "Bob",
		Comment:   "Post-Quantum Key",
		Email:     "bob@example.com",
		PublicKey: []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n..."),
	}

	gpg := mldsa87ed.New(ctx, mldsa87ed.Options{})
	defer func() {
		_ = gpg.Close()
	}()

	if err := gpg.Load(identity); err != nil {
		panic(err)
	}

	// Self-test the key pair before relying on it.
	if err := gpg.Check(); err != nil {
		panic(err)
	}

	in, err := os.Open("large_file.bin")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create("large_file.bin.gpg")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = out.Close()
	}()

	encReader, err := gpg.EncryptReader(in)
	if err != nil {
		panic(err)
	}

	if _, err = io.Copy(out, encReader); err != nil {
		panic(err)
	}

	_ = encReader.Close()
}
```

### 3. Key Expiration for Rotation Policies

Keys can be given a bounded lifetime at construction time. `Options.KeyTime` is a **duration in seconds** written straight into `packet.Config.KeyLifetimeSecs` — it is not an absolute timestamp.

```go
package main

import (
	"context"
	"time"

	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/encoding/opengpg/mldsa65ed"
)

func main() {
	ctx := context.Background()

	// Expire the key 90 days from now: KeyTime is a lifetime in seconds.
	keyTime := uint32(90 * 24 * time.Hour / time.Second)

	gpg := mldsa65ed.New(ctx, mldsa65ed.Options{
		KeyTime: keyTime,
	})
	defer func() {
		_ = gpg.Close()
	}()

	identity := &opengpg.Identity{
		Name:    "TemporaryKey",
		Comment: "90-day rotation",
		Email:   "temp@example.com",
	}

	if err := gpg.Create(identity); err != nil {
		panic(err)
	}
	// The generated primary key and subkey now expire after 90 days.
}
```

### 4. Using the Root Package Directly

The root package is fully functional on its own. Use it when you need a non-preset algorithm, a different hash, a different cipher, or a classic `packet.PublicKeyAlgorithm` such as `packet.PubKeyAlgoEd25519`.

```go
package main

import (
	"context"
	"crypto"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/nabbar/golib/encoding/opengpg"
)

func main() {
	ctx := context.Background()

	gpg := opengpg.New(ctx, opengpg.Options{
		Hash:      crypto.SHA3_512,
		Cipher:    packet.CipherAES256,
		Algorithm: packet.PubKeyAlgoMldsa65Ed25519,
	})
	defer func() {
		_ = gpg.Close()
	}()

	// ... use the full OpenGPG API from here on.
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/encoding/opengpg

# then pick exactly one preset
go get github.com/nabbar/golib/encoding/opengpg/slhdsa128s
```

### Basic Implementation

Create a key pair, self-test it, then encrypt and decrypt a string with the SLH-DSA-SHAKE128s preset:

```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/encoding/opengpg/slhdsa128s"
)

func main() {
	ctx := context.Background()

	gpg := slhdsa128s.New(ctx, slhdsa128s.Options{})
	defer func() {
		_ = gpg.Close()
	}()

	identity := &opengpg.Identity{
		Name:    "Alice",
		Comment: "Test Key",
		Email:   "alice@example.com",
	}

	if err := gpg.Create(identity); err != nil {
		panic(err)
	}

	if err := gpg.Check(); err != nil {
		panic(err)
	}

	// Encrypt with the writer path.
	plaintext := []byte("Hello, Post-Quantum World!")
	buf := bytes.NewBuffer(nil)

	encWriter, err := gpg.EncryptWriter(buf)
	if err != nil {
		panic(err)
	}

	if _, err = encWriter.Write(plaintext); err != nil {
		panic(err)
	}

	if err = encWriter.Close(); err != nil {
		panic(err)
	}

	// Decrypt with the reader path.
	decReader, err := gpg.DecryptReader(buf)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = decReader.Close()
	}()

	decrypted, err := io.ReadAll(decReader)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Decrypted: %s\n", string(decrypted))
	// → Decrypted: Hello, Post-Quantum World!
}
```

### Decrypting into a Writer

`DecryptWriter` is the counterpart path when the destination is a file or a socket rather than memory. Its `Close()` is wrapped in a `waitCloser` that blocks until the background decryption goroutine has finished, so a `Close()` returning `nil` guarantees all plaintext has reached the destination.

```go
// The caller writes ciphertext into wc; decrypted plaintext lands in out.
wc, err := gpg.DecryptWriter(out)
if err != nil {
	return err
}

if _, err = io.Copy(wc, encReader); err != nil {
	return err
}

// Closes the pipe writer, then waits for the decryption goroutine.
if err = wc.Close(); err != nil {
	return err
}
```

---

## Best Practices

### ✅ DO

- **Always call `Close()`** — `defer func() { _ = gpg.Close() }()`. It releases every tracked pipe end and encryption writer, and nils the entity list and `packet.Config` so key material becomes collectable.
  ```go
  defer func() { _ = gpg.Close() }()
  ```
- **Call `Check()` after `Create` or `Load`** — it performs a real encrypt-then-decrypt round-trip and returns `os.ErrInvalid` if the recovered payload does not match byte for byte.
  ```go
  if err := gpg.Check(); err != nil {
      return err // key is loaded but not functional
  }
  ```
- **Set `Options.KeyTime` to a lifetime in seconds** when keys must rotate; leave it at `0` for keys that never expire in practice.
  ```go
  slhdsa128s.New(ctx, slhdsa128s.Options{
      KeyTime: uint32(90 * 24 * time.Hour / time.Second),
  })
  ```
- **Provide `Options.Time` in tests** — a fixed clock makes key creation timestamps deterministic.
  ```go
  slhdsa128s.New(ctx, slhdsa128s.Options{
      Time: func() time.Time { return time.Unix(1700000000, 0) },
  })
  ```
- **Use `EncryptReader` for reading sources and `EncryptWriter` for writing sinks** — the reader path is fully streamed through a pipe, while the writer path is a direct synchronous write to the destination.
- **Close the `io.ReadCloser` returned by `DecryptReader`** — it is an `io.NopCloser` wrapper, so `Close` is a no-op, but closing it keeps call sites uniform.

### ❌ DON'T

- **Don't reuse an instance after `Close()`** — `Close()` nils the entity list and the config. `Create`, `Load`, `Check` and `EncryptReader` return `os.ErrClosed`; `EncryptWriter`, `DecryptReader` and `DecryptWriter` return `ErrorIdentityInvalid` because they only test for a loaded key. Create a new instance with `New()`.
- **Don't pass `nil` to `Create` or `Load`** — both return `ErrorIdentityInvalid`.
- **Don't call `Load` with an `Identity` that has both `PublicKey` and `PrivateKey` empty** — `Load` returns `ErrorIdentityInvalid`.
- **Don't pass an absolute Unix timestamp as `KeyTime`** — it is used verbatim as a lifetime in seconds, so a value like `1700000000` produces a key that claims to live for roughly 54 years.
- **Don't call `Encrypt*` or `Decrypt*` before `Create` or `Load`** — with no key loaded they return `ErrorIdentityInvalid`.
- **Don't call `Check()` in a hot path** — it encrypts and decrypts ~164 KiB on every call.
- **Don't assume signatures are verified** — `DecryptReader` and `DecryptWriter` expose go-crypto's `UnverifiedBody`; the message is decrypted but neither signed nor authenticated.
- **Don't write a private key to disk unprotected** — `Create` serializes the private key with `SerializePrivate` and no passphrase, so `Identity.PrivateKey` is cleartext armor.

---

## API Reference

### `OpenGPG` Interface

| Method | Parameters | Result | Description |
|--------|------------|--------|-------------|
| `Create` | `*Identity` | `error` | Generates a new key pair and writes ASCII-armored material into `Identity.PublicKey` / `Identity.PrivateKey`; stores the entity list. Returns `os.ErrClosed` if closed, `ErrorIdentityInvalid` if `Identity` is nil, `ErrorPublicKeyInvalid` / `ErrorPrivateKeyInvalid` on serialization failure |
| `Load` | `*Identity` | `error` | Imports an existing key pair from `Identity.PrivateKey` (preferred) or `Identity.PublicKey`; accepts armored or binary keyring data. Returns `os.ErrClosed`, `ErrorIdentityInvalid` or the matching key error |
| `Check` | *(none)* | `error` | Round-trip self-test: encrypts ~164 KiB with the loaded keys, decrypts it, compares byte for byte. Returns `os.ErrClosed`, `ErrorIdentityInvalid`, `os.ErrInvalid` on mismatch, or the underlying go-crypto error |
| `EncryptReader` | `io.Reader` | `(io.ReadCloser, error)` | Streaming encryption: background goroutine copies `r` into an `io.Pipe` through go-crypto using a pooled 32 KB buffer; errors surface via `CloseWithError` |
| `EncryptWriter` | `io.Writer` | `(io.WriteCloser, error)` | Direct synchronous encryption into `w`; the returned writer is registered in the `mapCloser` |
| `DecryptReader` | `io.Reader` | `(io.ReadCloser, error)` | Synchronous decryption: `ReadMessage` then `io.NopCloser` over `UnverifiedBody`. Not registered in the `mapCloser` |
| `DecryptWriter` | `io.Writer` | `(io.WriteCloser, error)` | Streaming decryption: caller writes ciphertext into a pipe, a background goroutine decrypts into `w`. Returns a `waitCloser` whose `Close` waits for the goroutine |
| `Close` | *(none)* | `error` | Sets the closed flag, closes all tracked resources, nils the entity list and config. Always returns `nil`; safe to call more than once |

All methods return `ErrorIdentityInvalid` when no key has been created or loaded, except `Create` and `Load` which take the key material as their argument.

### Constructors

| Symbol | Signature | Description |
|--------|-----------|-------------|
| `opengpg.New` | `func New(ctx context.Context, o Options) OpenGPG` | Builds the instance from a full `packet.Config` projection. `V6Keys: true` by default for Post-Quantum presets; V4 keys when `Algorithm` is classical (e.g. `PubKeyAlgoRSA`). Returns an instance with no key loaded |
| `mldsa65ed.New` | `func New(ctx context.Context, opt Options) opengpg.OpenGPG` | ML-DSA-65 + Ed25519 / ML-KEM-768 + X25519, SHA-3/512, AES-256 |
| `mldsa87ed.New` | `func New(ctx context.Context, opt Options) opengpg.OpenGPG` | ML-DSA-87 + Ed448 / ML-KEM-1024 + X448, SHA-3/512, AES-256 |
| `slhdsa128f.New` | `func New(ctx context.Context, opt Options) opengpg.OpenGPG` | SLH-DSA-SHAKE128f + Ed25519 / ML-KEM-768 + X25519, SHA-3/512, AES-256 |
| `slhdsa128s.New` | `func New(ctx context.Context, opt Options) opengpg.OpenGPG` | SLH-DSA-SHAKE128s + Ed25519 / ML-KEM-768 + X25519, SHA-3/512, AES-256 |
| `slhdsa256s.New` | `func New(ctx context.Context, opt Options) opengpg.OpenGPG` | SLH-DSA-SHAKE256s + Ed448 / ML-KEM-1024 + X448, SHA-3/512, AES-256 |
| `rsa4096sha256.New` | `func New(ctx context.Context, opt Options) opengpg.OpenGPG` | RSA-4096 / RSA-4096, SHA-256, AES-256, V4 keys |

### Unexported Implementation Types

| Type | Description |
|------|-------------|
| `mod` | The concrete `OpenGPG` implementation returned by `New`. Fields: `o *packet.Config` (encoding settings), `i EntityList` (loaded keys), `c *atomic.Bool` (closed flag), `l mapCloser.Closer` (tracked resources), `p *sync.Pool` (32 KB streaming buffers) |
| `waitCloser` | Wraps the `DecryptWriter` pipe writer; `Close` closes the writer and blocks on a `done` channel so the background goroutine is finished before returning |

### Package Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `defBufferSize` | `32 * 1024` | Size of each pooled streaming buffer |
| `chkStringBase` | `"OpenPGP ML-DSA65 + ED25519 verification payload"` | Base string repeated 4 096 times to build the `Check()` payload |

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >80%)
   - Pass all tests including race detector
   - Use `gofmt`, `golangci-lint` and `gosec`

2. **AI Usage Policy**
   - ❌ **AI must NEVER be used** to generate package code or core functionality
   - ✅ **AI assistance is limited to**:
       - Testing (writing and improving tests)
       - Debugging (troubleshooting and bug resolution)
       - Documentation (comments, README, TESTING.md)
   - All AI-assisted work must be reviewed and validated by humans

3. **Testing**
   - Add tests for new features
   - Use Ginkgo v2 / Gomega for test framework
   - Ensure zero race conditions
   - Maintain coverage above 80%

4. **Documentation**
   - Update GoDoc comments for public APIs
   - Add examples for new features
   - Update README.md and TESTING.md if needed

5. **Pull Request Process**
   - Fork the repository
   - Create a feature branch
   - Write clear commit messages
   - Ensure all tests pass
   - Update documentation
   - Submit PR with description of changes

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg)** — `OpenGPG` interface, `Options`, `Identity`, error codes and the `New` constructor
- **[GoDoc — mldsa65ed](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg/mldsa65ed)** — ML-DSA-65 + Ed25519 preset
- **[GoDoc — mldsa87ed](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg/mldsa87ed)** — ML-DSA-87 + Ed448 preset
- **[GoDoc — slhdsa128f](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg/slhdsa128f)** — SLH-DSA-SHAKE128f + Ed25519 preset
- **[GoDoc — slhdsa128s](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg/slhdsa128s)** — SLH-DSA-SHAKE128s + Ed25519 preset
- **[GoDoc — slhdsa256s](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg/slhdsa256s)** — SLH-DSA-SHAKE256s + Ed448 preset
- **[GoDoc — rsa4096sha256](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg/rsa4096sha256)** — RSA-4096 + SHA-256 classical preset

### Related golib Packages

- **[github.com/nabbar/golib/errors](https://pkg.go.dev/github.com/nabbar/golib/errors)** — Standardized error code registry; owns `MinPkgEncodingOpenGPG` and the `CodeError` type used by the constants in `errors.go`
- **[github.com/nabbar/golib/ioutils/mapCloser](https://pkg.go.dev/github.com/nabbar/golib/ioutils/mapCloser)** — `Closer` used to register pipe ends and streaming writers for deterministic cleanup on `Close()`

### External References

- **[ProtonMail/go-crypto/openpgp](https://pkg.go.dev/github.com/ProtonMail/go-crypto/openpgp)** — Underlying OpenPGP V6 implementation: entity generation, packet serialization, armor encoding, and the Post-Quantum hybrid public key algorithms used here
- **[ProtonMail/go-crypto/openpgp/packet](https://pkg.go.dev/github.com/ProtonMail/go-crypto/openpgp/packet)** — `packet.Config`, `packet.PublicKeyAlgorithm`, `packet.CipherFunction` and the algorithm identifiers referenced in the tables above
- **[RFC 9580 — The OpenPGP Message Format](https://datatracker.ietf.org/doc/html/rfc9580)** — Message format specification for OpenPGP
- **[FIPS 203 — Module-Lattice-Based Digital Signature Standard](https://csrc.nist.gov/pubs/fips/203/final)** — ML-DSA, used by the `mldsa65ed` and `mldsa87ed` presets
- **[FIPS 204 — Module-Lattice-Based Key-Encapsulation Mechanism Standard](https://csrc.nist.gov/pubs/fips/204/final)** — ML-KEM, used for the encryption subkey of every preset
- **[FIPS 205 — Stateless Hash-Based Digital Signature Standard](https://csrc.nist.gov/pubs/fips/205/final)** — SLH-DSA, used by the `slhdsa128f`, `slhdsa128s` and `slhdsa256s` presets
- **[Cloudflare CIRCL](https://github.com/cloudflare/circl)** — Go implementations of ML-DSA, ML-KEM and SLH-DSA that back the primitives above

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2020-2026 Nicolas JUHEL
