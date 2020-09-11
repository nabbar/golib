package configAws

import (
	"github.com/nabbar/golib/errors"
)

const (
	ErrorAwsError errors.CodeError = iota + errors.MIN_PKG_Aws + 30
	ErrorConfigLoader
	ErrorConfigValidator
	ErrorConfigJsonUnmarshall
	ErrorEndpointInvalid
	ErrorRegionInvalid
	ErrorRegionEndpointNotFound
	ErrorCredentialsInvalid
)

var isErrInit = false

func init() {
	errors.RegisterIdFctMessage(ErrorAwsError, getMessage)
	isErrInit = errors.ExistInMapMessage(ErrorAwsError)
}

func IsErrorInit() bool {
	return isErrInit
}

func getMessage(code errors.CodeError) string {
	switch code {
	case errors.UNK_ERROR:
		return ""
	case ErrorAwsError:
		return "calling aws api occurred a response error"
	case ErrorConfigLoader:
		return "calling AWS Default config Loader has occurred an error"
	case ErrorConfigValidator:
		return "invalid config, validation error"
	case ErrorConfigJsonUnmarshall:
		return "invalid json config, unmarshall error"
	case ErrorEndpointInvalid:
		return "the specified endpoint seems to be invalid"
	case ErrorRegionInvalid:
		return "the specified region seems to be invalid"
	case ErrorRegionEndpointNotFound:
		return "cannot find the endpoint for the specify region"
	case ErrorCredentialsInvalid:
		return "the specified credentials seems to be incorrect"
	}

	return ""
}
