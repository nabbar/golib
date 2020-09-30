package helper

import "github.com/nabbar/golib/errors"

const (
	// minmal are errors.MIN_AVAILABLE + get a hope free range 1000 + 10 for aws-config errors.
	ErrorResponse errors.CodeError = iota + errors.MIN_PKG_Aws + 60
	ErrorConfigEmpty
	ErrorAwsEmpty
	ErrorAws
	ErrorBucketNotFound
)

var isErrInit = false

func init() {
	errors.RegisterIdFctMessage(ErrorResponse, getMessage)
	isErrInit = errors.ExistInMapMessage(ErrorResponse)
}

func IsErrorInit() bool {
	return isErrInit
}

func getMessage(code errors.CodeError) string {
	switch code {
	case errors.UNK_ERROR:
		return ""
	case ErrorResponse:
		return "calling aws api occurred a response error"
	case ErrorConfigEmpty:
		return "the given config is empty or invalid"
	case ErrorAws:
		return "the aws request sent to aws API occurred an error"
	case ErrorAwsEmpty:
		return "the aws request sent to aws API occurred an empty result"
	case ErrorBucketNotFound:
		return "the specified bucket is not found"
	}

	return ""
}
