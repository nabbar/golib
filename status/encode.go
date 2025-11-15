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

package status

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	ginrdr "github.com/gin-gonic/gin/render"
	liberr "github.com/nabbar/golib/errors"
	monpol "github.com/nabbar/golib/monitor/pool"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
)

const (
	// encTextSepStatus is the separator between status and message in text output.
	// Format: "OK: MyApp (v1.0.0) | message"
	encTextSepStatus = ": "

	// encTextSepPart is the separator between parts in text output.
	// Used to separate application info from messages.
	encTextSepPart = " | "
)

// Encode defines the interface for status response encoding.
// It supports both JSON and plain text output formats.
type Encode interface {
	// String returns the status as a formatted string.
	// Format: "STATUS: Name (Release Hash Date) | Message\nComponent details..."
	String() string

	// Bytes returns the status as a byte slice (same as String).
	Bytes() []byte

	// GinRender renders the status response to a Gin context.
	// It sets the appropriate content type and HTTP status code.
	//
	// Parameters:
	//   - c: the Gin context to render to
	//   - isText: if true, renders as text/plain; if false, renders as JSON
	//   - isShort: if true, omits component details
	GinRender(c *ginsdk.Context, isText bool, isShort bool)

	// GinCode returns the HTTP status code for this response.
	// The code is determined by the health status and configuration.
	GinCode() int
}

// encodeModel is the internal implementation of the Encode interface.
// It contains all information needed to render a status response.
type encodeModel struct {
	Name      string        // Application name
	Release   string        // Release version (e.g., "v1.2.3")
	Hash      string        // Build hash or commit ID
	DateBuild time.Time     // Build date/time
	Status    monsts.Status // Overall health status (OK, Warn, KO)
	Message   string        // Status message from worst component
	Component montps.Pool   // Pool of monitored components (nil if short mode)
	code      int           // HTTP status code to return
}

// GinCode returns the HTTP status code for this response.
// The code is set based on the health status and configuration.
func (o *encodeModel) GinCode() int {
	return o.code
}

// GinRender renders the status response to the Gin context.
// It handles both JSON and plain text output formats.
//
// Parameters:
//   - c: the Gin context to render to
//   - isText: if true, renders as text/plain; if false, renders as JSON
//   - isShort: if true, clears the Component pool (no component details)
func (o *encodeModel) GinRender(c *ginsdk.Context, isText bool, isShort bool) {
	if isShort {
		o.Component = monpol.New(c)
	}

	if isText {
		c.Render(o.code, ginrdr.Data{
			ContentType: ginsdk.MIMEPlain,
			Data:        o.Bytes(),
		})
	} else {
		c.JSON(o.code, *o)
	}
}

// stringName formats the application name with version information.
// Format: "Name (Release Hash Date)" or just "Name" if no version info.
//
// Returns the formatted name string.
func (o *encodeModel) stringName() string {
	var inf []string

	if len(o.Release) > 0 {
		inf = append(inf, o.Release)
	}

	if len(o.Hash) > 0 {
		inf = append(inf, o.Hash)
	}

	if !o.DateBuild.IsZero() {
		inf = append(inf, o.DateBuild.Format(time.RFC3339))
	}

	if len(inf) > 0 {
		return fmt.Sprintf("%s (%s)", o.Name, strings.Join(inf, " "))
	} else {
		return o.Name
	}
}

// stringPart formats the main status line without the status prefix.
// Format: "Name (Release Hash Date) | Message"
//
// Returns the formatted string.
func (o *encodeModel) stringPart() string {
	item := make([]string, 0)
	item = append(item, o.stringName())

	if len(o.Message) > 0 {
		item = append(item, o.Message)
	}

	return strings.Join(item, encTextSepPart)
}

// String returns the complete status as a formatted string.
// Format: "STATUS: Name (Release Hash Date) | Message\nComponent details..."
//
// Returns the formatted status string.
func (o *encodeModel) String() string {
	var buf = bytes.NewBuffer(make([]byte, 0))

	buf.WriteString(o.Status.String() + encTextSepStatus + o.stringPart())
	buf.WriteRune('\n')

	if o.Component != nil {
		if p, err := o.Component.MarshalText(); err == nil {
			buf.Write(p)
		}
	}

	return buf.String()
}

// Bytes returns the status as a byte slice.
// It's equivalent to []byte(String()).
func (o *encodeModel) Bytes() []byte {
	return []byte(o.String())
}

// getEncodeModel creates an Encode instance with current status information.
// It computes the overall status and gathers all necessary data.
// This method is thread-safe.
//
// Returns an Encode instance ready for rendering.
func (o *sts) getEncodeModel() Encode {
	o.m.RLock()
	defer o.m.RUnlock()

	var (
		m         string
		s         monsts.Status
		name      string
		release   string
		hash      string
		dateBuild time.Time
	)

	s, m = o.getStatus()

	// Safely call functions with nil checks to prevent panics
	if o.fn != nil {
		name = o.fn()
	}
	if o.fr != nil {
		release = o.fr()
	}
	if o.fh != nil {
		hash = o.fh()
	}
	if o.fd != nil {
		dateBuild = o.fd()
	}

	return &encodeModel{
		Name:      name,
		Release:   release,
		Hash:      hash,
		DateBuild: dateBuild,
		Status:    s,
		Message:   m,
		Component: o._getPool(),
		code:      o.cfgGetReturnCode(s),
	}
}

// getMarshal creates an Encode instance after validating prerequisites.
// It checks that application information has been set via SetInfo or SetVersion.
//
// Returns an Encode instance or an error if application info is missing.
func (o *sts) getMarshal() (Encode, liberr.Error) {
	if !o.checkFunc() {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("missing status info for API"))
	}
	return o.getEncodeModel(), nil
}

// MarshalText implements encoding.TextMarshaler for the status instance.
// It allows the status to be marshaled as plain text.
//
// Returns the status as text bytes or an error if application info is missing.
func (o *sts) MarshalText() (text []byte, err error) {
	if enc, e := o.getMarshal(); e != nil {
		return nil, e
	} else {
		return enc.Bytes(), nil
	}
}

// MarshalJSON implements json.Marshaler for the status instance.
// It allows the status to be marshaled as JSON.
//
// Returns the status as JSON bytes or an error if application info is missing.
func (o *sts) MarshalJSON() ([]byte, error) {
	if enc, e := o.getMarshal(); e != nil {
		return nil, e
	} else {
		return json.Marshal(enc)
	}
}
