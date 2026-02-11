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

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	cfgcst "github.com/nabbar/golib/config/const"
	cfgtps "github.com/nabbar/golib/config/types"
	loglvl "github.com/nabbar/golib/logger/level"
	spfcbr "github.com/spf13/cobra"
)

// ComponentHas checks if a component with the given key is registered.
// Returns true if the component exists, false otherwise.
func (o *model) ComponentHas(key string) bool {
	return o.ComponentGet(key) != nil
}

// ComponentGet retrieves a component by its key.
// Returns the component if found, nil otherwise.
// This method is thread-safe for concurrent access.
func (o *model) ComponentGet(key string) cfgtps.Component {
	if i, l := o.cpt.Load(key); !l {
		return nil
	} else {
		return i
	}
}

// ComponentType returns the type identifier of a component.
// Returns an empty string if the component is not found.
func (o *model) ComponentType(key string) string {
	if v := o.ComponentGet(key); v == nil {
		return ""
	} else {
		return v.Type()
	}
}

// ComponentDel removes a component from the registry.
// This does not call Stop() on the component - stop it first if needed.
func (o *model) ComponentDel(key string) {
	o.cpt.Delete(key)
}

// ComponentSet registers a new component or replaces an existing one.
// The component is initialized with the provided key and necessary dependencies
// (context, viper, version, logger, monitor pool).
// If cpt is nil, this method does nothing.
func (o *model) ComponentSet(key string, cpt cfgtps.Component) {
	if cpt == nil {
		return
	}

	cpt.Init(key, o.ctx, o.ComponentGet, o.getViper, o.getVersion(), o.getDefaultLogger)

	if f := o.getFctMonitorPool(); f != nil {
		cpt.RegisterMonitorPool(f)
	} else {
		cpt.RegisterMonitorPool(o.getMonitorPool)
	}

	o.cpt.Store(key, cpt)
}

func (o *model) componentUpdate(key string, cpt cfgtps.Component) {
	if cpt == nil {
		return
	}

	o.cpt.Store(key, cpt)
}

// ComponentList returns all registered components as a map.
// Keys are component identifiers, values are the component instances.
// This creates a snapshot of the current component registry.
func (o *model) ComponentList() map[string]cfgtps.Component {
	var res = make(map[string]cfgtps.Component)

	o.cpt.Range(func(key string, val cfgtps.Component) bool {
		if len(key) > 0 && val != nil {
			res[key] = val
		} else {
			o.cpt.Delete(key)
		}
		return true
	})

	return res
}

// ComponentKeys returns a list of all registered component keys.
// This creates a snapshot of current component keys in the registry.
func (o *model) ComponentKeys() []string {
	var res = make([]string, 0)

	o.cpt.Range(func(key string, val cfgtps.Component) bool {
		if len(key) > 0 && val != nil {
			res = append(res, key)
		} else {
			o.cpt.Delete(key)
		}

		return true
	})

	return res
}

// ComponentStart starts all registered components in dependency order.
// Components are started sequentially according to their declared dependencies.
// If any component fails to start, an error is returned containing all failures.
// Logging is performed for each component start attempt.
func (o *model) ComponentStart() error {
	var err = ErrorComponentStart.Error(nil)

	for _, key := range o.ComponentDependencies() {
		if len(key) < 1 {
			continue
		} else if cpt := o.ComponentGet(key); cpt == nil {
			continue
		} else {
			ent := o.logEntry(loglvl.InfoLevel, "starting component")
			ent.FieldAdd("component", key)
			ent.Log()

			e := cpt.Start()
			o.componentUpdate(key, cpt)

			if e != nil {
				ent = o.logEntry(loglvl.ErrorLevel, "component return a starting error")
				ent.ErrorAdd(true, e)
				ent.FieldAdd("component", key)
				ent.Log()
				err.Add(e)
			} else if !cpt.IsStarted() {
				e = fmt.Errorf("component '%s' has been call to start, but is not started", key)
				ent = o.logEntry(loglvl.ErrorLevel, "component is not started")
				ent.ErrorAdd(true, e)
				ent.FieldAdd("component", key)
				ent.Log()
				err.Add(e)
			}
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

// ComponentIsStarted checks if all components are in the started state.
// Returns true only if all registered components report IsStarted() as true.
func (o *model) ComponentIsStarted() bool {
	isOk := true

	o.cpt.Range(func(key string, val cfgtps.Component) bool {
		if len(key) > 0 && val != nil {
			isOk = isOk && val.IsStarted()
		} else {
			o.cpt.Delete(key)
		}
		return isOk
	})

	return isOk
}

// ComponentReload reloads all registered components in dependency order.
// This allows components to refresh their configuration without a full restart.
// If any component fails to reload, an error is returned containing all failures.
// Components must remain in the started state after reload.
func (o *model) ComponentReload() error {
	var err = ErrorComponentReload.Error(nil)

	for _, key := range o.ComponentDependencies() {
		if len(key) < 1 {
			continue
		} else if cpt := o.ComponentGet(key); cpt == nil {
			continue
		} else {
			ent := o.logEntry(loglvl.InfoLevel, "reloading component")
			ent.FieldAdd("component", key)
			ent.Log()

			e := cpt.Reload()
			o.componentUpdate(key, cpt)

			if e != nil {
				ent = o.logEntry(loglvl.ErrorLevel, "reloading component return an error")
				ent.FieldAdd("component", key)
				ent.ErrorAdd(true, e)
				ent.Log()
				err.Add(e)
			} else if !cpt.IsStarted() {
				e = fmt.Errorf("component '%s' has been call to reload, but is not started", key)
				ent = o.logEntry(loglvl.ErrorLevel, "reloading component has been call, but component is not started")
				ent.FieldAdd("component", key)
				ent.ErrorAdd(true, e)
				ent.Log()
				err.Add(e)
			}
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

// ComponentStop stops all registered components in reverse dependency order.
// Components are stopped in the opposite order they were started to ensure
// proper cleanup of dependencies (e.g., API stops before database).
// This method does not return errors - it performs best-effort cleanup.
func (o *model) ComponentStop() {
	lst := o.ComponentDependencies()

	for i := len(lst) - 1; i >= 0; i-- {
		key := lst[i]

		if len(key) < 1 {
			continue
		} else if cpt := o.ComponentGet(key); cpt == nil {
			continue
		} else {
			cpt.Stop()
		}
	}
}

// ComponentIsRunning checks if components are in the running state.
// If atLeast is true, returns true if at least one component is running.
// If atLeast is false, returns true only if all components are running.
func (o *model) ComponentIsRunning(atLeast bool) bool {
	isOk := false

	o.cpt.Range(func(key string, val cfgtps.Component) bool {
		if len(key) > 0 && val != nil {
			if !atLeast {
				isOk = isOk && val.IsRunning()
				return isOk
			} else if val.IsRunning() {
				isOk = true
				return false
			}
		} else {
			o.cpt.Delete(key)
		}

		return true
	})

	return isOk
}

// ComponentDependencies returns all component keys in topological dependency order.
// Components with no dependencies come first, followed by components that depend on them.
// This order is used for starting components (forward) and stopping (reverse).
// The ordering algorithm performs a topological sort of the dependency graph.
func (o *model) ComponentDependencies() []string {
	var (
		list = make(map[string][]string)
		keys = make([]string, 0)
	)

	o.cpt.Range(func(key string, val cfgtps.Component) bool {
		if len(key) > 0 && val != nil {
			keys = append(keys, key)
			list[key] = val.Dependencies()
		} else {
			o.cpt.Delete(key)
		}
		return true
	})

	return o.orderDependencies(list, keys)
}

func (o *model) orderDependencies(list map[string][]string, dep []string) []string {
	var res = make([]string, 0)

	if len(list) < 1 || len(dep) < 1 {
		return res
	}

	for _, d := range dep {
		if _, ok := list[d]; !ok {
			continue
		}

		if len(list[d]) > 0 {
			for _, j := range o.orderDependencies(list, list[d]) {
				if len(j) < 1 {
					continue
				} else if slices.Contains(res, j) {
					continue
				}
				res = append(res, j)
			}
		}

		if !slices.Contains(res, d) {
			res = append(res, d)
		}
	}

	return res
}

// DefaultConfig generates a default configuration file for all registered components.
// Returns an io.Reader containing a JSON object with default configuration for each component.
// Each component contributes its default config under its registered key.
// The output is formatted JSON with proper indentation for readability.
func (o *model) DefaultConfig() io.Reader {
	var buffer = bytes.NewBuffer(make([]byte, 0))

	buffer.WriteString("{")
	buffer.WriteString("\n")

	n := buffer.Len()

	o.cpt.Range(func(key string, val cfgtps.Component) bool {
		if len(key) < 1 || val == nil {
			o.cpt.Delete(key)
			return true
		}

		if p := val.DefaultConfig(cfgcst.JSONIndent); len(p) > 0 {
			if buffer.Len() > n {
				buffer.WriteString(",")
				buffer.WriteString("\n")
			}
			buffer.WriteString(fmt.Sprintf("%s\"%s\": ", cfgcst.JSONIndent, key)) // nolint
			buffer.Write(p)
		}

		return true
	})

	buffer.WriteString("\n")
	buffer.WriteString("}")

	var (
		cmp = bytes.NewBuffer(make([]byte, 0))
		ind = bytes.NewBuffer(make([]byte, 0))
	)

	if err := json.Compact(cmp, buffer.Bytes()); err != nil {
		return buffer
	} else if err = json.Indent(ind, cmp.Bytes(), "", cfgcst.JSONIndent); err != nil {
		return buffer
	}

	return ind
}

// RegisterFlag registers CLI flags for all components with the given Cobra command.
// Each component can register its own flags for configuration via command-line.
// Returns an error containing all registration failures, if any.
func (o *model) RegisterFlag(Command *spfcbr.Command) error {
	var err = ErrorComponentFlagError.Error(nil)

	for _, k := range o.ComponentKeys() {
		if cpt := o.ComponentGet(k); cpt == nil {
			continue
		} else if e := cpt.RegisterFlag(Command); e != nil {
			ent := o.logEntry(loglvl.ErrorLevel, "component register flag return an error")
			ent.FieldAdd("component", k)
			ent.ErrorAdd(true, e)
			ent.Log()
			err.Add(e)
		} else {
			o.ComponentSet(k, cpt)
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

// ComponentWalk iterates over all registered components, calling the provided function for each.
// The iteration continues until all components are visited or the function returns false.
// Invalid entries (empty key or nil component) are automatically removed during iteration.
func (o *model) ComponentWalk(fct cfgtps.ComponentListWalkFunc) {
	o.cpt.Range(func(key string, val cfgtps.Component) bool {
		if len(key) > 0 && val != nil {
			return fct(key, val)
		}

		o.cpt.Delete(key)
		return true
	})
}
