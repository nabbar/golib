# Error package
This package allows extending go error interface to add parent tree, code of error, trace, ...
This package still compatible with error interface and can be customized to define what information is return for standard go error implementation.

The main mind of this package is to be simple use, quickly implement and having more capabilities than go error interface

## Example of implement
You will find many uses of this package in the `golib` repos.

We will create an `error.go` file into any package, and we will construct it like this :
```go
import errors "github.com/nabbar/golib/errors"

const (
    // create const for code as type CodeError to use features func
    // we init the iota number to an minimal free code
    // golib/errors expose the minimal available iota with const : errors.MIN_AVAILABLE
    // with this method, all code uint will be predictable
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_AVAILABLE
    MY_ERROR
    // add here all const you will use as code to identify error with code & message
)

func init() {
    // register your function getMessage
	errors.RegisterFctMessage(getMessage)
}

// This function will return the message of one error code
func getMessage(code errors.CodeError) (message string) {
	switch code {
	case EMPTY_PARAMS:
		return "given parameters is empty"
	case My_ERROR:
		return "example of message for code MY_ERROR"
	}

    // CAREFUL : the default return if code is not found must be en empty string !
	return ""
}
```

In go source file, we can implement the call to package `golib/errors` like this :
```go
// will test if err is not nil
// if is nil, will return nil, otherwise will return a Error interface from the MY_ERROR code with a parent err
_ = MY_ERROR.Iferror(err)
```

```go
// return an Error interface from the MY_ERROR code and if err is not nil, add a parent err
_ = MY_ERROR.ErrorParent(err)
```

```go
// error can be cascaded add like this
_ = MY_ERROR.IfError(EMPTY_PARAMS.IfError(err))
return MY_ERROR.Error(EMPTY_PARAMS.ErrorParent(err))
```

