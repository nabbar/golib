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
	"sync"
	"sync/atomic"
	"time"

	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"

	liberr "github.com/nabbar/golib/errors"
)

const JSONIndent = "  "

type ComponentList interface {
	// ComponentHas return true if the key is a registered Component
	ComponentHas(key string) bool

	// ComponentType return the Component Type of the registered key.
	ComponentType(key string) string

	// ComponentGet return the given component associated with the config Key.
	// The component can be transTyped to other interface to be exploited
	ComponentGet(key string) Component

	// ComponentDel remove the given Component key from the config.
	ComponentDel(key string)

	// ComponentSet stores the given Component with a key.
	ComponentSet(key string, cpt Component)

	// ComponentList returns a map of stored couple keyType and Component
	ComponentList() map[string]Component

	// ComponentKeys returns a slice of stored Component keys
	ComponentKeys() []string

	// ComponentStart trigger the Start function of each Component.
	// This function will keep the dependencies of each Component.
	// This function will stop the Start sequence on any error triggered.
	ComponentStart(getCfg FuncComponentConfigGet) liberr.Error

	// ComponentIsStarted will trigger the IsStarted function of all registered component.
	// If any component return false, this func return false.
	ComponentIsStarted() bool

	// ComponentReload trigger the Reload function of each Component.
	// This function will keep the dependencies of each Component.
	// This function will stop the Reload sequence on any error triggered.
	ComponentReload(getCfg FuncComponentConfigGet) liberr.Error

	// ComponentStop trigger the Stop function of each Component.
	// This function will not keep the dependencies of each Component.
	ComponentStop()

	// ComponentIsRunning will trigger the IsRunning function of all registered component.
	// If any component return false, this func return false.
	ComponentIsRunning(atLeast bool) bool

	// DefaultConfig aggregates all registered components' default config
	// Returns a filled buffer with a complete config json model
	DefaultConfig() io.Reader

	// RegisterFlag can be called to register flag to a spf cobra command and link it with viper
	// to retrieve it into the config viper.
	// The key will be use to stay config organisation by compose flag as key.config_key.
	RegisterFlag(Command *spfcbr.Command, Viper *spfvpr.Viper) error
}

func newComponentList() ComponentList {
	return &componentList{
		m: sync.Mutex{},
		l: make(map[string]*atomic.Value, 0),
	}
}

type componentList struct {
	m sync.Mutex
	l map[string]*atomic.Value
}

func (c *componentList) ComponentHas(key string) bool {
	c.m.Lock()
	defer c.m.Unlock()

	_, ok := c.l[key]
	return ok
}

func (c *componentList) ComponentType(key string) string {
	if !c.ComponentHas(key) {
		return ""
	} else if o := c.ComponentGet(key); o == nil {
		return ""
	} else {
		return o.Type()
	}
}

func (c *componentList) ComponentGet(key string) Component {
	if !c.ComponentHas(key) {
		return nil
	}

	c.m.Lock()
	defer c.m.Unlock()

	if len(c.l) < 1 {
		c.l = make(map[string]*atomic.Value, 0)
	}

	if v := c.l[key]; v == nil {
		return nil
	} else if i := v.Load(); i == nil {
		return nil
	} else if o, ok := i.(Component); !ok {
		return nil
	} else {
		return o
	}
}

func (c *componentList) ComponentDel(key string) {
	if !c.ComponentHas(key) {
		return
	}

	c.m.Lock()
	defer c.m.Unlock()

	if len(c.l) < 1 {
		c.l = make(map[string]*atomic.Value, 0)
	}

	if v := c.l[key]; v == nil {
		return
	} else {
		c.l[key] = new(atomic.Value)
	}
}

func (c *componentList) ComponentSet(key string, cpt Component) {
	c.m.Lock()
	defer c.m.Unlock()

	if len(c.l) < 1 {
		c.l = make(map[string]*atomic.Value, 0)
	}

	if v, ok := c.l[key]; !ok || v == nil {
		c.l[key] = new(atomic.Value)
	}

	c.l[key].Store(cpt)
}

func (c *componentList) ComponentList() map[string]Component {
	var res = make(map[string]Component, 0)

	for _, k := range c.ComponentKeys() {
		res[k] = c.ComponentGet(k)
	}

	return res
}

func (c *componentList) ComponentKeys() []string {
	c.m.Lock()
	defer c.m.Unlock()

	var res = make([]string, 0)

	for k := range c.l {
		res = append(res, k)
	}

	return res
}

func (c *componentList) startOne(key string, getCfg FuncComponentConfigGet) liberr.Error {
	var cpt Component

	if !c.ComponentHas(key) {
		return ErrorComponentNotFound.ErrorParent(fmt.Errorf("component: %s", key))
	} else if cpt = c.ComponentGet(key); cpt == nil {
		return ErrorComponentNotFound.ErrorParent(fmt.Errorf("component: %s", key))
	} else if cpt.IsStarted() {
		return nil
	}

	if dep := cpt.Dependencies(); len(dep) > 0 {
		for _, k := range dep {

			var err liberr.Error

			for retry := 0; retry < 3; retry++ {

				if err = c.startOne(k, getCfg); err == nil {
					break
				}

				time.Sleep(100 * time.Millisecond)
			}

			if err != nil {
				return err
			}
		}
	}

	if err := cpt.Start(getCfg); err != nil {
		return err
	} else {
		c.ComponentSet(key, cpt)
	}

	return nil
}

func (c *componentList) ComponentStart(getCfg FuncComponentConfigGet) liberr.Error {
	for _, key := range c.ComponentKeys() {
		if err := c.startOne(key, getCfg); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentList) ComponentIsStarted() bool {
	for _, k := range c.ComponentKeys() {
		if cpt := c.ComponentGet(k); cpt == nil {
			continue
		} else if ok := cpt.IsStarted(); !ok {
			return false
		}
	}

	return true
}

func (c *componentList) reloadOne(isReload []string, key string, getCfg FuncComponentConfigGet) ([]string, liberr.Error) {
	var (
		err liberr.Error
		cpt Component
	)

	if !c.ComponentHas(key) {
		return isReload, ErrorComponentNotFound.ErrorParent(fmt.Errorf("component: %s", key))
	} else if cpt = c.ComponentGet(key); cpt == nil {
		return isReload, ErrorComponentNotFound.ErrorParent(fmt.Errorf("component: %s", key))
	} else if stringIsInSlice(isReload, key) {
		return isReload, nil
	}

	if dep := cpt.Dependencies(); len(dep) > 0 {
		for _, k := range dep {
			if isReload, err = c.reloadOne(isReload, k, getCfg); err != nil {
				return isReload, err
			}
		}
	}

	if err = cpt.Reload(getCfg); err != nil {
		return isReload, err
	} else {
		c.ComponentSet(key, cpt)
		isReload = append(isReload, key)
	}

	return isReload, nil
}

func (c *componentList) ComponentReload(getCfg FuncComponentConfigGet) liberr.Error {
	var (
		err liberr.Error
		key string

		isReload = make([]string, 0)
	)

	for _, key = range c.ComponentKeys() {
		if isReload, err = c.reloadOne(isReload, key, getCfg); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentList) ComponentStop() {
	for _, key := range c.ComponentKeys() {
		if !c.ComponentHas(key) {
			continue
		}

		cpt := c.ComponentGet(key)
		if cpt == nil {
			continue
		}

		cpt.Stop()
	}
}

func (c *componentList) ComponentIsRunning(atLeast bool) bool {
	for _, k := range c.ComponentKeys() {
		if cpt := c.ComponentGet(k); cpt == nil {
			continue
		} else if ok := cpt.IsRunning(atLeast); !ok {
			return false
		}
	}

	return true
}

func (c *componentList) DefaultConfig() io.Reader {
	var buffer = bytes.NewBuffer(make([]byte, 0))

	buffer.WriteString("{")
	buffer.WriteString("\n")

	n := buffer.Len()

	for _, k := range c.ComponentKeys() {
		if cpt := c.ComponentGet(k); cpt == nil {
			continue
		} else if p := cpt.DefaultConfig(JSONIndent); len(p) > 0 {
			if buffer.Len() > n {
				buffer.WriteString(",")
				buffer.WriteString("\n")
			}
			buffer.WriteString(fmt.Sprintf("%s\"%s\": ", JSONIndent, k))
			buffer.Write(p)
		}
	}

	buffer.WriteString("\n")
	buffer.WriteString("}")

	var res = bytes.NewBuffer(make([]byte, 0))
	if err := json.Indent(res, buffer.Bytes(), "", JSONIndent); err != nil {
		return buffer
	}

	return res
}

func (c *componentList) RegisterFlag(Command *spfcbr.Command, Viper *spfvpr.Viper) error {
	var err = ErrorComponentFlagError.Error(nil)

	for _, k := range c.ComponentKeys() {
		if cpt := c.ComponentGet(k); cpt == nil {
			continue
		} else if e := cpt.RegisterFlag(Command, Viper); e != nil {
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
