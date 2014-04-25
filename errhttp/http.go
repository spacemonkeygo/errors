// Copyright (C) 2014 Space Monkey, Inc.

package errhttp

import (
	"fmt"

	"code.spacemonkey.com/go/errors"
)

var (
	statusCode = errors.GenSym()
	errorBody  = errors.GenSym()
)

func SetStatusCode(code int) errors.ErrorOption {
	return errors.SetData(statusCode, code)
}

func OverrideErrorBody(message string) errors.ErrorOption {
	return errors.SetData(errorBody, message)
}

func RestoreDefaultErrorBody() errors.ErrorOption {
	return errors.SetData(errorBody, nil)
}

func GetStatusCode(err error, default_code int) int {
	rv := errors.GetData(err, statusCode)
	sc, ok := rv.(int)
	if ok {
		return sc
	}
	return default_code
}

func GetErrorBody(err error) string {
	rv := errors.GetData(err, errorBody)
	message, ok := rv.(string)
	if !ok {
		return errors.GetMessage(err)
	}
	class := errors.GetClass(err)
	if class == nil {
		return message
	}
	return fmt.Sprintf("%s: %s", class.String(), message)
}
