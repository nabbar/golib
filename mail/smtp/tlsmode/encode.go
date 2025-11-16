/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package tlsmode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements json.Marshaler for TLSMode.
//
// The TLSMode is marshaled to its string representation:
//   - TLSNone → ""
//   - TLSStartTLS → "starttls"
//   - TLSStrictTLS → "tls"
//
// Example JSON output:
//
//	{"mode": "starttls"}
func (t TLSMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements json.Unmarshaler for TLSMode.
//
// Accepts both string and numeric representations:
//   - String: "starttls", "tls", "" → parsed with Parse
//   - Number: 0, 1, 2 → parsed with ParseInt64
//
// Example JSON inputs:
//
//	{"mode": "starttls"}  // TLSStartTLS
//	{"mode": 1}           // TLSStartTLS
func (t *TLSMode) UnmarshalJSON(p []byte) error {
	var (
		s string
		i int64
	)

	if err := json.Unmarshal(p, &s); err == nil {
		*t = Parse(s)
		return nil
	} else if err = json.Unmarshal(p, &i); err == nil {
		*t = ParseInt64(i)
		return nil
	} else {
		return err
	}
}

// MarshalYAML implements yaml.Marshaler for TLSMode.
//
// The TLSMode is marshaled to its string representation.
// See gopkg.in/yaml.v3 for YAML marshaling details.
func (t TLSMode) MarshalYAML() (interface{}, error) {
	return t.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler for TLSMode.
//
// Accepts both string and numeric representations in YAML format.
// See gopkg.in/yaml.v3 for YAML unmarshaling details.
//
// Example YAML inputs:
//
//	mode: starttls
//	mode: 1
func (t *TLSMode) UnmarshalYAML(value *yaml.Node) error {
	if i, e := strconv.Atoi(value.Value); e == nil {
		*t = ParseInt64(int64(i))
	} else {
		*t = Parse(value.Value)
	}

	return nil
}

// MarshalTOML implements TOML marshaling for TLSMode.
//
// The TLSMode is marshaled to JSON format, which is compatible with TOML.
func (t TLSMode) MarshalTOML() ([]byte, error) {
	return t.MarshalJSON()
}

// UnmarshalTOML implements TOML unmarshaling for TLSMode.
//
// Accepts multiple types:
//   - []byte or string: parsed with Parse/ParseBytes
//   - uint64, int64, float64: parsed with respective Parse functions
//
// Returns an error if the type is not supported.
func (t *TLSMode) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		*t = ParseBytes(p)
		return nil
	}
	if p, k := i.(string); k {
		*t = Parse(p)
		return nil
	}
	if p, k := i.(uint64); k {
		*t = ParseUint64(p)
		return nil
	}
	if p, k := i.(int64); k {
		*t = ParseInt64(p)
		return nil
	}
	if p, k := i.(float64); k {
		*t = ParseFloat64(p)
		return nil
	}
	return fmt.Errorf("size: value not in valid format")
}

// MarshalText implements encoding.TextMarshaler for TLSMode.
//
// The TLSMode is marshaled to its string representation as a byte slice.
// This is used by packages that support text encoding, such as encoding/xml.
func (t TLSMode) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for TLSMode.
//
// The byte slice is parsed as a string using ParseBytes.
// This is used by packages that support text decoding, such as encoding/xml.
func (t *TLSMode) UnmarshalText(p []byte) error {
	*t = ParseBytes(p)
	return nil
}

// MarshalCBOR implements CBOR marshaling for TLSMode.
//
// The TLSMode is marshaled to its string representation in CBOR format.
// See github.com/fxamacker/cbor/v2 for CBOR marshaling details.
func (t TLSMode) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(t.String())
}

// UnmarshalCBOR implements CBOR unmarshaling for TLSMode.
//
// Accepts both string and numeric representations in CBOR format.
// See github.com/fxamacker/cbor/v2 for CBOR unmarshaling details.
//
// Returns an error if the CBOR data cannot be unmarshaled to either
// a string or an int64.
func (t *TLSMode) UnmarshalCBOR(p []byte) error {
	var (
		i int64
		b string
	)

	if e := cbor.Unmarshal(p, &b); e == nil {
		*t = Parse(b)
		return nil
	} else if e = cbor.Unmarshal(p, &i); e == nil {
		*t = ParseInt64(i)
		return nil
	} else {
		return fmt.Errorf("cbor: cannot unmarshall '%s' into TLSMode", string(p))
	}
}

// MarshalBinary implements encoding.BinaryMarshaler for TLSMode.
//
// The TLSMode is marshaled using CBOR format.
// This is used by packages that support binary encoding, such as encoding/gob.
func (t TLSMode) MarshalBinary() ([]byte, error) {
	return t.MarshalCBOR()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for TLSMode.
//
// The binary data is expected to be in CBOR format.
// This is used by packages that support binary decoding, such as encoding/gob.
func (t *TLSMode) UnmarshalBinary(p []byte) error {
	return t.UnmarshalCBOR(p)
}
