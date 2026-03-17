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
	stsmdt "github.com/nabbar/golib/status/mandatory"
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
	// Name is the name of the application.
	Name string `json:"name"`

	// Info contains global descriptive metadata about the service, such as
	// description, links, version, and build information.
	Info map[string]interface{} `json:"info"`

	// Status is the overall health status of the application (OK, Warn, KO).
	Status monsts.Status `json:"status"`

	// Message provides a summary of the overall health status.
	Message string `json:"message"`

	// Component contains the detailed status of individual monitored components.
	// It is omitted in short mode.
	Component encComponent `json:"component,omitempty"`

	// code is the HTTP status code to be returned.
	code int
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
		// In short mode, we don't want to show any component details.
		// Setting Component to nil achieves this for JSON (`omitempty`) and text.
		o.Component = monpol.New(c)
	}

	if isText {
		// Render as plain text.
		c.Render(o.code, ginrdr.Data{
			ContentType: ginsdk.MIMEPlain,
			Data:        o.Bytes(),
		})
	} else {
		// Render as JSON.
		c.JSON(o.code, *o)
	}
}

// stringName formats the application name with its version information for text output.
// Example: "MyApp (release: v1.2.3 hash: abc123)"
func (o *encodeModel) stringName() string {
	var inf []string

	if len(o.Info) > 0 {
		for k, v := range o.Info {
			inf = append(inf, fmt.Sprintf("%s: %v", k, v))
		}
	}

	if len(inf) > 0 {
		return fmt.Sprintf("%s (%s)", o.Name, strings.Join(inf, " "))
	} else {
		return o.Name
	}
}

// stringPart formats the main status line for text output, excluding the status prefix.
// Example: "MyApp (v1.2.3) | All systems operational"
func (o *encodeModel) stringPart() string {
	item := make([]string, 0)
	item = append(item, o.stringName())

	if len(o.Message) > 0 {
		item = append(item, o.Message)
	}

	return strings.Join(item, encTextSepPart)
}

// String returns the complete status as a formatted string for text output.
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
// information. It computes the overall status, gathers all necessary data for
// rendering, and applies any specified filters. This method is thread-safe.
func (o *sts) getEncodeModel(isMap bool, filter []string) Encode {
	o.m.RLock()
	defer o.m.RUnlock()

	var (
		m    string
		s    monsts.Status
		name string
		info = o.cfgGetInfo()
	)

	s, m = o.getStatus()

	if o.fn != nil {
		name = o.fn()
	}

	// Populate Info map with version details if not already present.
	if _, k := info["release"]; !k && o.fr != nil {
		info["release"] = o.fr()
	}

	if _, k := info["hash"]; !k && o.fh != nil {
		info["hash"] = o.fh()
	}

	if _, k := info["date"]; !k && o.fd != nil {
		if ts := o.fd(); !ts.IsZero() {
			info["date"] = ts.Format(time.RFC3339)
		}
	}

	if _, k := info["epoc"]; !k && o.fd != nil {
		if ts := o.fd(); !ts.IsZero() {
			info["epoc"] = fmt.Sprintf("%d", ts.Unix())
		}
	}

	enc := &encodeModel{
		Name:      name,
		Info:      info,
		Status:    s,
		Message:   m,
		Component: nil,
		code:      o.cfgGetReturnCode(s),
	}

	lst := o.cfgGetMandatory()
	var pol = make(map[string]montps.MonitorStatus, 0)

	if p := o.getPool(); p != nil {
		p.MonitorWalk(func(name string, val montps.Monitor) bool {
			pol[name] = val
			return true
		})
	}

	// Apply filters if provided.
	if len(filter) > 0 {
		// First, try to filter by mandatory group name.
		if l := o.cfgFilterMandatory(filter); l != nil && l.Len() > 0 {
			filter = make([]string, 0)
			l.Walk(func(_ string, m stsmdt.Mandatory) bool {
				filter = append(filter, m.KeyList()...)
				return true
			})
			lst = l
		} else {
			// If no groups match, fall back to filtering individual monitor names.
			// In this case, map mode is disabled as there's no group context.
			isMap = false
		}

		// Apply the resulting filter to the monitor pool.
		if p := o.filterPool(filter); len(p) > 0 {
			pol = p
		}
	}

	// Select the appropriate component encoder (map or list).
	if isMap {
		enc.Component = &modControl{
			ctr: lst,
			cpt: pol,
			fct: o.getStatus,
		}
	} else {
		enc.Component = &modPool{
			cpt: pol,
		}
	}

	return enc
}

// getMarshal creates an `Encode` instance after validating that all prerequisite
// information (application name) has been set.
//
// Returns an `Encode` instance or an error if the application info is missing.
func (o *sts) getMarshal(isMap bool, filter []string) (Encode, liberr.Error) {
	if !o.checkFunc() {
		return nil, ErrorParamEmpty.Error(fmt.Errorf("missing status info for API"))
	}
	return o.getEncodeModel(isMap, filter), nil
}

// MarshalText implements the `encoding.TextMarshaler` interface for the status instance,
// allowing it to be marshaled as plain text.
func (o *sts) MarshalText() (text []byte, err error) {
	if enc, e := o.getMarshal(false, nil); e != nil {
		return nil, e
	} else {
		return enc.Bytes(), nil
	}
}

// MarshalJSON implements the `json.Marshaler` interface for the status instance,
// allowing it to be marshaled as JSON.
func (o *sts) MarshalJSON() ([]byte, error) {
	if enc, e := o.getMarshal(false, nil); e != nil {
		return nil, e
	} else {
		return json.Marshal(enc)
	}
}
