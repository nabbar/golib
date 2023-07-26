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

	cfgcst "github.com/nabbar/golib/config/const"
	cfgtps "github.com/nabbar/golib/config/types"
	liberr "github.com/nabbar/golib/errors"
	spfcbr "github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

func (c *configModel) ComponentHas(key string) bool {
	if i, l := c.cpt.Load(key); !l {
		return false
	} else {
		_, k := i.(cfgtps.Component)
		return k
	}
}

func (c *configModel) ComponentType(key string) string {
	if v := c.ComponentGet(key); v == nil {
		return ""
	} else {
		return v.Type()
	}
}

func (c *configModel) ComponentGet(key string) cfgtps.Component {
	if i, l := c.cpt.Load(key); !l {
		return nil
	} else if v, k := i.(cfgtps.Component); !k {
		return nil
	} else {
		return v
	}
}

func (c *configModel) ComponentDel(key string) {
	c.cpt.Delete(key)
}

func (c *configModel) ComponentSet(key string, cpt cfgtps.Component) {
	cpt.Init(key, c.ctx.GetContext, c.ComponentGet, c.getViper, c.getVersion(), c.getDefaultLogger)
	if f := c.getFctMonitorPool(); f != nil {
		cpt.RegisterMonitorPool(f)
	} else {
		cpt.RegisterMonitorPool(c.getMonitorPool)
	}
	c.cpt.Store(key, cpt)
}

func (c *configModel) componentUpdate(key string, cpt cfgtps.Component) {
	c.cpt.Store(key, cpt)
}

func (c *configModel) ComponentList() map[string]cfgtps.Component {
	var res = make(map[string]cfgtps.Component, 0)
	c.cpt.Walk(func(key string, val interface{}) bool {
		if v, k := val.(cfgtps.Component); k {
			res[key] = v
		} else {
			c.cpt.Delete(key)
		}
		return true
	})
	return res
}

func (c *configModel) ComponentKeys() []string {
	var res = make([]string, 0)
	c.cpt.Walk(func(key string, val interface{}) bool {
		if _, k := val.(cfgtps.Component); k {
			res = append(res, key)
		} else {
			c.cpt.Delete(key)
		}
		return true
	})
	return res
}

func (c *configModel) ComponentStart() liberr.Error {
	var err = ErrorComponentStart.Error(nil)

	for _, key := range c.ComponentDependencies() {
		if len(key) < 1 {
			continue
		} else if cpt := c.ComponentGet(key); cpt == nil {
			continue
		} else {
			e := cpt.Start()
			c.componentUpdate(key, cpt)
			if e != nil {
				err.AddParent(e)
			} else if !cpt.IsStarted() {
				err.AddParent(fmt.Errorf("component '%s' has been call to start, but is not started", key))
			}
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c *configModel) ComponentIsStarted() bool {
	isOk := false

	c.cpt.Walk(func(key string, val interface{}) bool {
		if v, k := val.(cfgtps.Component); !k {
			c.cpt.Delete(key)
		} else if v == nil {
			c.cpt.Delete(key)
		} else if v.IsStarted() {
			isOk = true
			return false
		}

		return true
	})

	return isOk
}

func (c *configModel) ComponentReload() liberr.Error {
	var err = ErrorComponentReload.Error(nil)

	for _, key := range c.ComponentDependencies() {
		if len(key) < 1 {
			continue
		} else if cpt := c.ComponentGet(key); cpt == nil {
			continue
		} else {
			e := cpt.Reload()
			c.componentUpdate(key, cpt)
			if e != nil {
				err.AddParent(e)
			} else if !cpt.IsStarted() {
				err.AddParent(fmt.Errorf("component '%s' has been call to reload, but is not started", key))
			}
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c *configModel) ComponentStop() {
	lst := c.ComponentDependencies()

	for i := len(lst) - 1; i >= 0; i-- {
		key := lst[i]

		if len(key) < 1 {
			continue
		} else if cpt := c.ComponentGet(key); cpt == nil {
			continue
		} else {
			cpt.Stop()
		}
	}
}

func (c *configModel) ComponentIsRunning(atLeast bool) bool {
	isOk := false

	c.cpt.Walk(func(key string, val interface{}) bool {
		v, k := val.(cfgtps.Component)

		if !k {
			c.cpt.Delete(key)
			return true
		} else if v == nil {
			c.cpt.Delete(key)
			return true
		}

		if v.IsRunning() {
			isOk = true
		}

		if atLeast && isOk {
			return false
		} else if !atLeast && !isOk {
			return false
		}

		return true
	})

	return isOk
}

func (c *configModel) ComponentDependencies() []string {
	var (
		list = make(map[string][]string, 0)
		keys = make([]string, 0)
	)

	c.cpt.Walk(func(key string, val interface{}) bool {

		if v, k := val.(cfgtps.Component); !k {
			c.cpt.Delete(key)
		} else if v == nil {
			c.cpt.Delete(key)
		} else {
			list[key] = v.Dependencies()
			keys = append(keys, key)
		}

		return true
	})

	return c.orderDependencies(list, keys)
}

func (c *configModel) orderDependencies(list map[string][]string, dep []string) []string {
	var res = make([]string, 0)

	if len(list) < 1 || len(dep) < 1 {
		return res
	}

	for _, d := range dep {
		if _, ok := list[d]; !ok {
			continue
		}

		if len(list[d]) > 0 {
			for _, j := range c.orderDependencies(list, list[d]) {
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

func (c *configModel) DefaultConfig() io.Reader {
	var buffer = bytes.NewBuffer(make([]byte, 0))

	buffer.WriteString("{")
	buffer.WriteString("\n")

	n := buffer.Len()

	c.cpt.Walk(func(key string, val interface{}) bool {
		v, k := val.(cfgtps.Component)

		if !k {
			c.ComponentDel(key)
			return true
		} else if v == nil {
			c.ComponentDel(key)
			return true
		}

		if p := v.DefaultConfig(cfgcst.JSONIndent); len(p) > 0 {
			if buffer.Len() > n {
				buffer.WriteString(",")
				buffer.WriteString("\n")
			}
			buffer.WriteString(fmt.Sprintf("%s\"%s\": ", cfgcst.JSONIndent, key))
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

func (c *configModel) RegisterFlag(Command *spfcbr.Command) error {
	var err = ErrorComponentFlagError.Error(nil)

	for _, k := range c.ComponentKeys() {
		if cpt := c.ComponentGet(k); cpt == nil {
			continue
		} else if e := cpt.RegisterFlag(Command); e != nil {
			err.AddParent(e)
		} else {
			c.ComponentSet(k, cpt)
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c *configModel) ComponentWalk(fct cfgtps.ComponentListWalkFunc) {
	c.cpt.Walk(func(key string, val interface{}) bool {
		if v, k := val.(cfgtps.Component); !k {
			c.cpt.Delete(key)
			return true
		} else if v == nil {
			c.cpt.Delete(key)
			return true
		} else {
			return fct(key, v)
		}
	})
}
