/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
 */

package duration

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON returns the JSON encoding of the duration.
//
// The JSON encoding is simply the string representation of the duration
// wrapped in double quotes.
//
// Example:
//
// d := duration.MustParse("1h2m3s")
// b, err := d.MarshalJSON()
//
//	if err != nil {
//	    panic(err)
//	}
//
// fmt.Println(string(b)) // Output: "1h2m3s"
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON parses the JSON-encoded duration and stores the result in
// the receiver.
//
// The JSON encoding is expected to be a string representation of the
// duration wrapped in double quotes.
//
// Example:
//
// b := []byte(`"1h2m3s"`)
// d := &duration.Duration{}
//
//	if err := d.UnmarshalJSON(b); err != nil {
//	    panic(err)
//	}
//
// fmt.Println(d.String()) // Output: "1h2m3s"
func (d *Duration) UnmarshalJSON(bytes []byte) error {
	return d.unmarshall(bytes)
}

// MarshalYAML returns the YAML encoding of the duration.
//
// The YAML encoding is simply the string representation of the duration.
//
// Example:
//
// d := duration.MustParse("1h2m3s")
// y, err := d.MarshalYAML()
//
//	if err != nil {
//	    panic(err)
//	}
//
// fmt.Println(y) // Output: "1h2m3s"
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// UnmarshalYAML parses the YAML-encoded duration and stores the result in
// the receiver.
//
// The YAML encoding is expected to be a string representation of the
// duration.
//
// Example:
//
// y := &yaml.Node{Value: "1h2m3s"}
// d := &duration.Duration{}
//
//	if err := d.UnmarshalYAML(y); err != nil {
//	    panic(err)
//	}
//
// fmt.Println(d.String()) // Output: "1h2m3s"
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	return d.unmarshall([]byte(value.Value))
}

func (d Duration) MarshalTOML() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalTOML(i interface{}) error {
	if b, k := i.([]byte); k {
		return d.unmarshall(b)
	}

	if b, k := i.(string); k {
		return d.parseString(b)
	}

	return fmt.Errorf("duration: value not in valid format")
}

// MarshalText returns the text encoding of the duration.
//
// The text encoding is simply the string representation of the duration.
//
// Example:
//
// d := duration.MustParse("1h2m3s")
// b, err := d.MarshalText()
//
//	if err != nil {
//	    panic(err)
//	}
//
// fmt.Println(string(b)) // Output: "1h2m3s"
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText parses the text-encoded duration and stores the result in the receiver.
//
// The text encoding is expected to be a string representation of the duration.
//
// Example:
//
// d := &duration.Duration{}
// b := []byte("1h2m3s")
//
//	if err := d.UnmarshalText(b); err != nil {
//	    panic(err)
//	}
//
//	fmt.Println(d.String()) // Output: "1h2m3s"
func (d *Duration) UnmarshalText(bytes []byte) error {
	return d.unmarshall(bytes)
}

// MarshalCBOR returns the CBOR encoding of the duration.
//
// The CBOR encoding is simply the CBOR encoding of the string
// representation of the duration.
//
// Example:
//
// d := duration.MustParse("1h2m3s")
// b, err := d.MarshalCBOR()
//
//	if err != nil {
//	    panic(err)
//	}
//
//	fmt.Println(string(b)) // Output: CBOR encoding of "1h2m3s"
func (d Duration) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(d.String())
}

// UnmarshalCBOR parses the CBOR-encoded duration and stores the result in the receiver.
//
// The CBOR encoding is expected to be the CBOR encoding of the string
// representation of the duration.
//
// Example:
//
// d := &duration.Duration{}
// b := []byte{CBOR encoding of "1h2m3s"}
//
//	if err := d.UnmarshalCBOR(b); err != nil {
//	    panic(err)
//	}
//
//	fmt.Println(d.String()) // Output: "1h2m3s"
func (d *Duration) UnmarshalCBOR(bytes []byte) error {
	var s string
	if e := cbor.Unmarshal(bytes, &s); e != nil {
		return e
	} else {
		return d.unmarshall([]byte(s))
	}
}
