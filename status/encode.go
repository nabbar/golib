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
)

const (
	// encTextSepStatus defines the separator used between the status and the
	// rest of the message in plain text output.
	// Example: "OK: MyApp..."
	encTextSepStatus = ": "

	// encTextSepPart defines the separator used between different parts of the
	// main status line in plain text output.
	// Example: "MyApp (v1.0.0) | All systems operational"
	encTextSepPart = " | "
)

// Encode defines the interface for status response encoding. It provides methods
// for rendering the status in different formats (JSON, plain text) and for
// integration with the Gin framework.
type Encode interface {
	// String returns the status as a formatted string, suitable for logging or
	// plain text responses.
	String() string

	// Bytes returns the status as a byte slice, equivalent to `[]byte(String())`.
	Bytes() []byte

	// GinRender renders the status response to a Gin context. It handles content
	// negotiation (JSON vs. text) and verbosity (full vs. short).
	GinRender(c *ginsdk.Context, isText bool, isShort bool)

	// GinCode returns the appropriate HTTP status code for the response, based on
	// the overall health status and configuration.
	GinCode() int
}

// encodeModel is the internal implementation of the Encode interface. It serves
// as a data transfer object holding all the information required to render a
// status response.
type encodeModel struct {
	Name      string        `json:"name"`
	Release   string        `json:"release"`
	Hash      string        `json:"hash"`
	DateBuild time.Time     `json:"date_build"`
	Status    monsts.Status `json:"status"`
	Message   string        `json:"message"`
	Component encComponent  `json:"component,omitempty"` // The list of monitored components.
	code      int           // The HTTP status code to be returned.
}

// GinCode returns the HTTP status code for this response.
func (o *encodeModel) GinCode() int {
	return o.code
}

// GinRender renders the status response to the Gin context. It sets the appropriate
// content type and HTTP status code.
//
// If `isShort` is true, the component details are omitted from the response.
// If `isText` is true, the response is rendered as plain text; otherwise, as JSON.
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

// stringName formats the application name with its version information.
// Example: "MyApp (v1.2.3 abc123 2023-10-27T10:00:00Z)"
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

// stringPart formats the main status line, excluding the status prefix.
// Example: "MyApp (v1.2.3) | All systems operational"
func (o *encodeModel) stringPart() string {
	item := make([]string, 0)
	item = append(item, o.stringName())

	if len(o.Message) > 0 {
		item = append(item, o.Message)
	}

	return strings.Join(item, encTextSepPart)
}

// String returns the complete status as a formatted string.
// Example: "OK: MyApp (v1.2.3) | All systems operational\n  database: OK\n"
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

// Bytes returns the status as a byte slice, equivalent to `[]byte(String())`.
func (o *encodeModel) Bytes() []byte {
	return []byte(o.String())
}

// getEncodeModel creates an `Encode` instance populated with the current status
// information. It computes the overall status and gathers all necessary data for
// rendering. This method is thread-safe.
func (o *sts) getEncodeModel(isMap bool) Encode {
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

	enc := &encodeModel{
		Name:      name,
		Release:   release,
		Hash:      hash,
		DateBuild: dateBuild,
		Status:    s,
		Message:   m,
		code:      o.cfgGetReturnCode(s),
	}

	if isMap {
		enc.Component = &modControl{
			ctr: o.cfgGetMandatory(),
			cpt: o._getPool(),
			fct: o.getStatus,
		}
	} else {
		enc.Component = o._getPool()
	}

	return enc
}

// getMarshal creates an `Encode` instance after validating that all prerequisite
// information (application name, release, etc.) has been set.
//
// Returns an `Encode` instance or an error if the application info is missing.
func (o *sts) getMarshal(isMap bool) (Encode, liberr.Error) {
	if !o.checkFunc() {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("missing status info for API"))
	}
	return o.getEncodeModel(isMap), nil
}

// MarshalText implements the `encoding.TextMarshaler` interface for the status instance,
// allowing it to be marshaled as plain text.
func (o *sts) MarshalText() (text []byte, err error) {
	if enc, e := o.getMarshal(false); e != nil {
		return nil, e
	} else {
		return enc.Bytes(), nil
	}
}

// MarshalJSON implements the `json.Marshaler` interface for the status instance,
// allowing it to be marshaled as JSON.
func (o *sts) MarshalJSON() ([]byte, error) {
	if enc, e := o.getMarshal(false); e != nil {
		return nil, e
	} else {
		return json.Marshal(enc)
	}
}
