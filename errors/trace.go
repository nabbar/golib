/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package errors

import (
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	// PathSeparator defines the standard path separator used internally.
	PathSeparator = "/"

	// pkgRuntime is the name of the Go runtime package, used for filtering stack frames.
	pkgRuntime = "runtime"

	// fltMod and fltVendor are precomputed strings used for filtering module and vendor paths from stack traces.
	fltMod    = "/pkg/mod/"
	fltVendor = "/vendor/"
)

var (
	// filterPkg stores the base path of the current package, used to filter out internal frames.
	filterPkg = path.Clean(ConvPathFromLocal(reflect.TypeOf(UnknownError).PkgPath()))

	// pcPool is a sync.Pool to reuse uintptr slices for runtime.Callers, reducing memory allocations.
	pcPool = sync.Pool{
		New: func() any {
			return make([]uintptr, 64) // A reasonable default capacity for stack frames.
		},
	}

	// pathCache stores results of filterPath to avoid repeated string manipulations and improve performance.
	pathCache sync.Map

	// trcCache stores formatted trace strings for each program counter (PC) to avoid re-formatting.
	trcCache sync.Map
)

// ClearCachePath clears all cached path and trace strings.
// This might be useful in long-running applications where paths or package structures change dynamically (rare).
func ClearCachePath() {
	pathCache.Range(func(key, value interface{}) bool {
		pathCache.Delete(key)
		return true
	})
	trcCache.Range(func(key, value interface{}) bool {
		trcCache.Delete(key)
		return true
	})
}

// ConvPathFromLocal converts a local file path (using OS-specific separator) to a standard Unix-like path.
func ConvPathFromLocal(str string) string {
	if filepath.Separator == '/' {
		return str
	}
	return strings.ReplaceAll(str, string(filepath.Separator), PathSeparator)
}

// init function runs once on package initialization to set up filterPkg.
func init() {
	// Adjust filterPkg to be the root of the module if it's within a vendor directory.
	if i := strings.LastIndex(filterPkg, fltVendor); i != -1 {
		filterPkg = filterPkg[:i+1]
	}
}

// hasPkgPrefix checks if a function name belongs to a specific package.
func hasPkgPrefix(name, pkg string) bool {
	if pkg == "" {
		return false
	}
	if !strings.HasPrefix(name, pkg) {
		return false
	}
	if len(name) == len(pkg) {
		return true
	}
	next := name[len(pkg)]
	return next == '.' || next == '/'
}

// getFrame captures the first relevant stack frame outside of the current package.
// It skips frames from the 'runtime' package and the 'errors' package itself.
func getFrame() tracer {
	// Acquire a uintptr slice from the pool.
	ptr := pcPool.Get().([]uintptr)
	// Ensure the slice is returned to the pool.
	defer pcPool.Put(ptr) // nolint

	// runtime.Callers(3, ...) skips the current function (getFrame), its caller, and its caller's caller.
	// This helps to get the actual call site of the error creation.
	nbr := runtime.Callers(3, ptr)

	for i := 0; i < nbr; i++ {
		if f := runtime.FuncForPC(ptr[i]); f != nil {
			// Skip frames that are part of the 'errors' package itself.
			if hasPkgPrefix(f.Name(), filterPkg) {
				continue
			}
			return trcPC(ptr[i]) // Return the first external frame.
		}
	}

	return trcNil{} // No relevant frame found.
}

// getFrameVendor captures multiple stack frames, specifically excluding frames from the 'runtime' package
// and the 'errors' package, and also excluding frames from vendor directories.
// It's used for more detailed traces, especially in panic recovery.
func getFrameVendor() []tracer {
	// Acquire a uintptr slice from the pool.
	ptr := pcPool.Get().([]uintptr)
	// Ensure the slice is returned to the pool.
	defer pcPool.Put(ptr) // nolint

	// runtime.Callers(3, ...) skips the current function (getFrameVendor), its caller, and its caller's caller.
	nbr := runtime.Callers(3, ptr)

	if nbr < 1 {
		return make([]tracer, 0)
	}

	var res = make([]tracer, 0, 5) // Pre-allocate for a few frames.

	for i := 0; i < nbr; i++ {
		f := runtime.FuncForPC(ptr[i])
		if f == nil {
			continue
		}

		name := f.Name()
		// Skip frames from the 'errors' package and the 'runtime' package.
		if hasPkgPrefix(name, filterPkg) || hasPkgPrefix(name, pkgRuntime) {
			continue
		}

		// We still need the file for vendor check.
		file, _ := f.FileLine(ptr[i])
		if strings.Contains(file, fltVendor) {
			continue
		}

		// Check for duplicates to avoid adding the same frame multiple times.
		var ok = false
		for j := range res {
			if res[j].PC() == ptr[i] {
				ok = true
				break
			}
		}

		if !ok {
			res = append(res, trcPC(ptr[i]))
			if len(res) >= 5 { // Limit the number of frames to avoid excessive output.
				break
			}
		}
	}

	return res
}

// filterPath cleans up file paths for display in stack traces.
// It removes module paths, vendor paths, and standardizes separators.
func filterPath(pathname string) string {
	if pathname == "" {
		return ""
	}

	// Check cache first to avoid redundant processing.
	if val, ok := pathCache.Load(pathname); ok {
		return val.(string)
	}

	original := pathname // Keep original for caching.
	pathname = ConvPathFromLocal(pathname)

	// Remove the current package's path prefix.
	if i := strings.LastIndex(pathname, filterPkg); i != -1 {
		matchLen := len(filterPkg)
		if i+matchLen == len(pathname) || pathname[i+matchLen] == '/' || pathname[i+matchLen] == '@' {
			pathname = pathname[i+matchLen:]
		}
	}

	// Remove module and vendor path prefixes.
	if i := strings.LastIndex(pathname, fltMod); i != -1 {
		pathname = pathname[i+len(fltMod):]
	} else if i = strings.LastIndex(pathname, fltVendor); i != -1 {
		pathname = pathname[i+len(fltVendor):]
	}

	// Clean up any remaining relative path components or redundant separators.
	if len(pathname) > 0 && (pathname[0] == '/' || strings.Contains(pathname, "//") || strings.Contains(pathname, "/.")) {
		pathname = path.Clean(pathname)
	}
	res := strings.Trim(pathname, PathSeparator)

	// Store the cleaned path in the cache.
	pathCache.Store(original, res)

	return res
}

// tracer is an interface for capturing and formatting stack trace information.
type tracer interface {
	// PC returns the program counter for the trace.
	PC() uintptr

	// String returns a formatted string representation of the trace.
	String() string

	// Compose returns the function name, file path, and line number of the trace.
	Compose() (fct string, file string, line int)

	// IsSame compares two tracers for equality, typically by their PC or composed elements.
	IsSame(other tracer) bool
}

// trcPC implements the tracer interface for a program counter.
type trcPC uintptr

// PC returns the program counter.
func (t trcPC) PC() uintptr {
	return uintptr(t)
}

// frame retrieves the runtime.Frame for the program counter.
func (t trcPC) frame() runtime.Frame {
	if t == 0 {
		return runtime.Frame{}
	}

	f := runtime.FuncForPC(uintptr(t))

	if f == nil {
		return runtime.Frame{}
	}

	file, line := f.FileLine(uintptr(t))

	return runtime.Frame{
		Function: f.Name(),
		File:     file,
		Line:     line,
	}
}

// String returns a cached or newly formatted string representation of the stack frame.
func (t trcPC) String() string {
	if t == 0 {
		return ""
	}

	// Check cache first.
	if val, ok := trcCache.Load(uintptr(t)); ok {
		return val.(string)
	}

	f := t.frame()
	res := stringFrame(f.Function, f.File, f.Line)
	trcCache.Store(uintptr(t), res) // Store in cache.

	return res
}

// Compose returns the function name, file path, and line number.
func (t trcPC) Compose() (string, string, int) {
	f := t.frame()
	return f.Function, f.File, f.Line
}

// IsSame compares two trcPC instances. It prioritizes PC equality, then function and file names.
func (t trcPC) IsSame(other tracer) bool {
	if other == nil {
		return false
	}

	if t.PC() == other.PC() {
		return true
	}

	fc1, fl1, _ := t.Compose()
	fc2, fl2, _ := other.Compose()

	// If PCs differ, compare function and file names for a "semantic" equality.
	if fc1 != "" && fc2 != "" && fc1 == fc2 {
		return true
	}

	return fl1 == fl2
}

// trcNil implements the tracer interface for a nil or manually constructed trace.
type trcNil struct {
	Func string
	File string
	Line int
}

// PC returns 0 for a nil trace.
func (t trcNil) PC() uintptr {
	return 0
}

// String returns a formatted string representation of the manually provided trace info.
func (t trcNil) String() string {
	return stringFrame(t.Func, t.File, t.Line)
}

// Compose returns the manually provided function name, file path, and line number.
func (t trcNil) Compose() (string, string, int) {
	return t.Func, t.File, t.Line
}

// IsSame compares two trcNil instances or a trcNil with a trcPC.
func (t trcNil) IsSame(other tracer) bool {
	if other == nil || other.PC() > 0 { // A trcNil cannot be the same as a trcPC.
		return false
	}

	fc, fl, ln := other.Compose()
	return fc == t.Func && fl == t.File && ln == t.Line
}

// stringFrame formats function, file, and line into a concise string.
func stringFrame(fct, file string, line int) string {
	if file != "" {
		if line != 0 {
			return filterPath(file) + "#" + strconv.Itoa(line)
		}
		return filterPath(file)
	}

	if fct != "" {
		if line != 0 {
			return fct + "#" + strconv.Itoa(line)
		}
		return fct
	}

	if line != 0 {
		return "#" + strconv.Itoa(line)
	}

	return ""
}
