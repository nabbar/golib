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

package types

import (
	"io"

	liberr "github.com/nabbar/golib/errors"
	spfcbr "github.com/spf13/cobra"
)

// ComponentListWalkFunc is used for ComponentList.Walk.
// For each component the function is call with the component key and the component as params.
// If the function return false, the loop is breaking, other else, the loop will run the function
// until the end of the component list.
type ComponentListWalkFunc func(key string, cpt Component) bool

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

	// ComponentWalk run a function on each component
	ComponentWalk(fct ComponentListWalkFunc)

	// ComponentKeys returns a slice of stored Component keys
	ComponentKeys() []string

	// ComponentStart trigger the Start function of each Component.
	// This function will keep the dependencies of each Component.
	// This function will stop the Start sequence on any error triggered.
	ComponentStart() liberr.Error

	// ComponentIsStarted will trigger the IsStarted function of all registered component.
	// If any component return false, this func return false.
	ComponentIsStarted() bool

	// ComponentReload trigger the Reload function of each Component.
	// This function will keep the dependencies of each Component.
	// This function will stop the Reload sequence on any error triggered.
	ComponentReload() liberr.Error

	// ComponentStop trigger the Stop function of each Component.
	// This function will not keep the dependencies of each Component.
	ComponentStop()

	// ComponentIsRunning will trigger the IsRunning function of all registered component.
	// If asLeast params is false and at least one component return false, this func return false.
	// Otherwise, if the params is true, if at least one component is true, this func return true.
	ComponentIsRunning(atLeast bool) bool

	// DefaultConfig aggregates all registered components' default config
	// Returns a filled buffer with a complete config json model
	DefaultConfig() io.Reader

	// RegisterFlag can be called to register flag to a spf cobra command and link it with viper
	// to retrieve it into the config viper.
	// The key will be use to stay config organisation by compose flag as key.config_key.
	RegisterFlag(Command *spfcbr.Command) error
}
