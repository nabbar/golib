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
 *
 */

package monitor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
	moninf "github.com/nabbar/golib/monitor/types"
	librun "github.com/nabbar/golib/runner"
)

const (
	// Text encoding separators for human-readable output
	encTextSepStatus = ": "  // Separator between status and details
	encTextSepPart   = " | " // Separator between different parts
	encTextSepTime   = " / " // Separator between time durations
)

// Encode provides methods for converting monitor state to different formats.
type Encode interface {
	String() string // Returns a human-readable string representation
	Bytes() []byte  // Returns the byte representation of the string
}

// encMod holds the data needed for encoding monitor state.
type encMod struct {
	Status monsts.Status // Current health status
	Name   string        // Monitor name
	Info   moninf.Info   // Metadata information

	Latency  string // Formatted latency duration
	Uptime   string // Formatted uptime duration
	Downtime string // Formatted downtime duration

	Message string // Error message if any
}

// Bytes returns the byte representation of the encoded monitor state.
func (e *encMod) Bytes() []byte {
	return []byte(e.String())
}

// String returns a human-readable representation of the monitor state.
// Format: "<STATUS>: <name> (<info>) | <latency> / <uptime> / <downtime> | <message>"
func (e *encMod) String() string {
	return e.Status.String() + encTextSepStatus + e.stringPart()
}

// stringDuration formats the duration metrics as a string.
func (e *encMod) stringDuration() string {
	part := append(make([]string, 0), e.Latency, e.Uptime, e.Downtime)
	return strings.Join(part, encTextSepTime)
}

// stringName formats the name and info as a string.
func (e *encMod) stringName() string {
	var inf string

	if e.Info != nil {
		i, _ := e.Info.MarshalText()
		inf = string(i)
	}

	if len(inf) > 0 {
		return fmt.Sprintf("%s (%s)", e.Name, inf)
	} else {
		return e.Name
	}
}

// stringPart combines all parts into a formatted string.
func (e *encMod) stringPart() string {
	item := make([]string, 0)
	item = append(item, e.stringName())
	item = append(item, e.stringDuration())

	if len(e.Message) > 0 {
		item = append(item, e.Message)
	}

	return strings.Join(item, encTextSepPart)
}

// getEncMod creates an Encode instance from the current monitor state.
func (o *mon) getEncMod() Encode {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/encMod/getEncMod", r)
		}
	}()

	return &encMod{
		Status:   o.Status(),
		Name:     o.Name(),
		Info:     o.InfoGet(),
		Latency:  o.Latency().Truncate(time.Millisecond).String(),
		Uptime:   o.Uptime().Truncate(time.Second).String(),
		Downtime: o.Downtime().Truncate(time.Second).String(),
		Message:  o.Message(),
	}
}

// MarshalText implements encoding.TextMarshaler.
// It returns a human-readable text representation of the monitor state.
func (o *mon) MarshalText() (text []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/encMod/MarshalText", r)
		}
	}()

	return o.getEncMod().Bytes(), nil
}

// MarshalJSON implements json.Marshaler.
// It returns a JSON representation of the monitor state.
func (o *mon) MarshalJSON() (text []byte, err error) {
	return json.Marshal(o.getEncMod())
}
