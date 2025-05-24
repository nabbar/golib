# `errors` Package

The `errors` package provides a comprehensive framework for error handling in Go, supporting error codes, messages, parent/child error relationships, stack traces, and integration with web frameworks like Gin. It is designed to facilitate structured, traceable, and user-friendly error management in complex applications.

## Features

- Custom error type with code, message, and parent errors
- Error code registration and message mapping
- Stack trace capture and filtering
- Error wrapping and unwrapping (compatible with Go's standard `errors` package)
- Flexible error formatting modes (code, message, trace, etc.)
- Integration with Gin for HTTP error responses
- Utilities for error lists, code checks, and string search

---

## Main Types & Interfaces

### `Error` Interface

Represents a single error with code, message, trace, and parent errors. Key methods:

- `IsCode(code CodeError) bool`: Checks if the error has the given code.
- `HasCode(code CodeError) bool`: Checks if the error or any parent has the code.
- `GetCode() CodeError`: Returns the error code.
- `GetParentCode() []CodeError`: Returns all codes in the error chain.
- `Is(e error) bool`: Checks if the error matches another error (compatible with `errors.Is`).
- `HasError(err error) bool`: Checks if the error or any parent matches the given error.
- `HasParent() bool`: Returns true if there are parent errors.
- `GetParent(withMainError bool) []error`: Returns all parent errors.
- `Map(fct FuncMap) bool`: Applies a function to the error and all parents.
- `ContainsString(s string) bool`: Checks if the message contains a substring.
- `Add(parent ...error)`: Adds parent errors.
- `SetParent(parent ...error)`: Replaces parent errors.
- `Code() uint16`: Returns the code as uint16.
- `CodeSlice() []uint16`: Returns all codes in the chain.
- `CodeError(pattern string) string`: Formats the error with code and message.
- `CodeErrorTrace(pattern string) string`: Formats with code, message, and trace.
- `Error() string`: Returns the error string (format depends on mode).
- `StringError() string`: Returns the error message.
- `GetError() error`: Returns a standard error.
- `Unwrap() []error`: Returns parent errors for Go error unwrapping.
- `GetTrace() string`: Returns the stack trace.
- `Return(r Return)`: Fills a `Return` struct for API responses.

### `Errors` Interface

- `ErrorsLast() error`: Returns the last error.
- `ErrorsList() []error`: Returns all errors.

### `Return` and `ReturnGin` Interfaces

For API error responses, including Gin integration.

---

## Error Creation & Usage

### Creating Errors

```go
import "github.com/nabbar/golib/errors"

err := errors.New(1001, "Something went wrong")
err2 := errors.New(1002, "Another error", err)
```

### Wrapping and Checking Errors

```go
if errors.IsCode(err, 1001) {
    // handle specific error code
}

if errors.ContainsString(err, "wrong") {
    // handle error containing substring
}
```

### Error Formatting Modes

Set the global error formatting mode:

```go
errors.SetModeReturnError(errors.ErrorReturnCodeErrorTrace)
```

Modes include: code only, code+message, code+message+trace, message only, etc.

### Gin Integration

```go
import "github.com/gin-gonic/gin"

var r errors.DefaultReturn
err.Return(&r)
r.GinTonicAbort(ctx, 500)
```

---

## Error Codes & Messages

- Register custom error codes and messages using `RegisterIdFctMessage`.
- Retrieve code locations with `GetCodePackages`.

---

## Stack Traces

- Each error captures the file, line, and function where it was created.
- Traces are filtered to remove vendor and runtime paths.

---

## Notes

- Fully compatible with Go's `errors.Is` and `errors.As`.
- Supports error chaining and parent/child relationships.
- Designed for use in both CLI and web applications.

---

Voici une section de documentation en anglais pour aider les développeurs à définir et enregistrer des codes d'erreur personnalisés avec le package `github.com/nabbar/golib/errors` :

---

## Defining and Registering Custom Error Codes

To create your own error codes and messages with the `errors` package, follow these steps:

### 1. Define Error Code Constants

Define your error codes as constants of type `liberr.CodeError`. Use an offset (e.g., your package's minimum code) to avoid collisions with other packages:

```go
import liberr "github.com/nabbar/golib/errors"

const (
    MyErrorInvalidParam liberr.CodeError = iota + liberr.MinPkgMyFeature
    MyErrorFileNotFound
    MyErrorProcessingFailed
)
```

### 2. Register Message Retrieval Function

Implement a function that returns a message for each error code, and register it using `liberr.RegisterIdFctMessage` in an `init()` function. Before registering, check for code collisions with `liberr.ExistInMapMessage`:

```go
func init() {
    if liberr.ExistInMapMessage(MyErrorInvalidParam) {
        panic(fmt.Errorf("error code collision with package myfeature"))
    }
    liberr.RegisterIdFctMessage(MyErrorInvalidParam, getMyFeatureErrorMessage)
}

func getMyFeatureErrorMessage(code liberr.CodeError) string {
    switch code {
    case MyErrorInvalidParam:
        return "invalid parameter provided"
    case MyErrorFileNotFound:
        return "file not found"
    case MyErrorProcessingFailed:
        return "processing failed"
    }
    return liberr.NullMessage
}
```

### 3. Usage Example

Create and use your custom errors as follows:

```go
err := MyErrorInvalidParam.Error(fmt.Errorf("additional context: %s", "details"))

```

And check for specific error codes or messages:

```go
import (
    "fmt"
    liberr "github.com/nabbar/golib/errors"
)

if liberr.IsCode(err, MyErrorInvalidParam) {
    // handle specific error
}

if e := Get(err); e != nil && e.HasParent() {
    for _, parent := range err.GetParent(true) {
        fmt.Println("Parent error:", parent)
    }
}

if ContainsString(err, "missing") {
    // handle error containing substring
}

if e := Get(err); e != nil {
    fmt.Println("Error code:", e.Code())
    fmt.Println("Error message:", e.StringError())
	fmt.Println("Error trace:", e.GetTrace())
	for _, oneErr := range e.GetErrorSlice() {
        fmt.Println("Error code in slice:", oneErr.Code())
		fmt.Println("Error message:", oneErr.StringError())
        fmt.Println("Error trace:", oneErr.GetTrace())
	}
}

```

---

**Notes:**

- Always check for code collisions before registering new error codes.
- Use a unique offset for your package to avoid overlapping with other packages.
- Register a message function for each error code group to provide meaningful error messages.