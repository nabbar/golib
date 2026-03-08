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
	"context"
	"sync/atomic"

	ginsdk "github.com/gin-gonic/gin"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"
	libver "github.com/nabbar/golib/version"
)

// mod implements the CptStatus interface.
// It acts as a component wrapper around the status monitoring functionality provided by libsts.Status.
type mod struct {
	// x is a context-aware configuration store for the component's internal state.
	x libctx.Config[uint8]

	// s is the underlying status instance that provides the core health check and monitoring logic.
	s libsts.Status

	// r is an atomic boolean to track the running state of the component (started/stopped).
	r *atomic.Bool
}

// Expose implements the libsts.Status interface by calling the underlying status object.
// It sets up the health check routes on the provided Gin engine context.
func (o *mod) Expose(ctx context.Context) {
	o.s.Expose(ctx)
}

// MiddleWare implements the libsts.Status interface by calling the underlying status object.
// It returns a Gin middleware that adds status information to the request context.
func (o *mod) MiddleWare(c *ginsdk.Context) {
	o.s.MiddleWare(c)
}

// SetErrorReturn implements the libsts.Status interface by calling the underlying status object.
// It configures a custom function to generate Gin error responses.
func (o *mod) SetErrorReturn(f func() liberr.ReturnGin) {
	o.s.SetErrorReturn(f)
}

// SetInfo implements the libsts.Status interface by calling the underlying status object.
// It sets application information (name, release, hash) for the status endpoint.
func (o *mod) SetInfo(name, release, hash string) {
	o.s.SetInfo(name, release, hash)
}

// SetVersion implements the libsts.Status interface by calling the underlying status object.
// It sets the application version information from a libver.Version instance.
func (o *mod) SetVersion(vers libver.Version) {
	o.s.SetVersion(vers)
}

// MarshalText implements the encoding.TextMarshaler interface by calling the underlying status object.
// It returns a text representation of the overall health status.
func (o *mod) MarshalText() (text []byte, err error) {
	return o.s.MarshalText()
}

// MarshalJSON implements the json.Marshaler interface by calling the underlying status object.
// It returns a JSON representation of the detailed health status.
func (o *mod) MarshalJSON() ([]byte, error) {
	return o.s.MarshalJSON()
}

// MonitorAdd implements the libsts.Status interface by calling the underlying status object.
// It adds a new monitor to the status pool.
func (o *mod) MonitorAdd(mon montps.Monitor) error {
	return o.s.MonitorAdd(mon)
}

// MonitorGet implements the libsts.Status interface by calling the underlying status object.
// It retrieves a monitor from the status pool by its name.
func (o *mod) MonitorGet(name string) montps.Monitor {
	return o.s.MonitorGet(name)
}

// MonitorSet implements the libsts.Status interface by calling the underlying status object.
// It adds or updates a monitor in the status pool.
func (o *mod) MonitorSet(mon montps.Monitor) error {
	return o.s.MonitorSet(mon)
}

// MonitorDel implements the libsts.Status interface by calling the underlying status object.
// It removes a monitor from the status pool by its name.
func (o *mod) MonitorDel(name string) {
	o.s.MonitorDel(name)
}

// MonitorList implements the libsts.Status interface by calling the underlying status object.
// It returns a list of all registered monitor names.
func (o *mod) MonitorList() []string {
	return o.s.MonitorList()
}

// MonitorWalk implements the libsts.Status interface by calling the underlying status object.
// It iterates over registered monitors and executes a callback function for each.
func (o *mod) MonitorWalk(fct func(name string, val montps.Monitor) bool, validName ...string) {
	o.s.MonitorWalk(fct)
}

// RegisterPool implements the libsts.Status interface by calling the underlying status object.
// It registers a function that provides a monitor pool, allowing for dynamic monitor updates.
func (o *mod) RegisterPool(fct montps.FuncPool) {
	o.s.RegisterPool(fct)
}

func (o *mod) RegisterGetConfigCpt(fct libsts.FuncGetCfgCpt) {
	o.s.RegisterGetConfigCpt(fct)
}

// SetConfig implements the libsts.Status interface by calling the underlying status object.
// It applies a new configuration for status checking logic.
func (o *mod) SetConfig(cfg libsts.Config) {
	o.s.SetConfig(cfg)
}

func (o *mod) GetConfig() libsts.Config {
	return o.s.GetConfig()
}

// IsHealthy implements the libsts.Status interface by calling the underlying status object.
// It checks if the specified monitors (or all if none specified) are in a healthy (OK or Warn) state.
func (o *mod) IsHealthy(name ...string) bool {
	return o.s.IsHealthy(name...)
}

// IsStrictlyHealthy implements the libsts.Status interface by calling the underlying status object.
// It checks if the specified monitors (or all if none specified) are in a strictly healthy (OK) state.
func (o *mod) IsStrictlyHealthy(name ...string) bool {
	return o.s.IsStrictlyHealthy(name...)
}

// IsCacheHealthy implements the libsts.Status interface by calling the underlying status object.
// It checks if the cached overall health status is healthy (OK or Warn).
func (o *mod) IsCacheHealthy() bool {
	return o.s.IsCacheHealthy()
}

// IsCacheStrictlyHealthy implements the libsts.Status interface by calling the underlying status object.
// It checks if the cached overall health status is strictly healthy (OK).
func (o *mod) IsCacheStrictlyHealthy() bool {
	return o.s.IsCacheStrictlyHealthy()
}
