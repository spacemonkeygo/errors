// Copyright (C) 2014 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package errhttp provides some useful helpers on top of the errors package for
HTTP handling.

errhttp is a great example of how to use the errors package SetData and GetData
hierarchical methods.
*/
package errhttp

import (
	"fmt"

	"github.com/SpaceMonkeyGo/errors"
)

var (
	statusCode = errors.GenSym()
	errorBody  = errors.GenSym()
)

// SetStatusCode returns an ErrorOption (for use in ErrorClass creation or
// error instantiation) that controls the error's HTTP status code
func SetStatusCode(code int) errors.ErrorOption {
	return errors.SetData(statusCode, code)
}

// OverrideErrorBody returns an ErrorOption (for use in ErrorClass creation or
// error instantiation) that controls the error body seen by GetErrorBody.
func OverrideErrorBody(message string) errors.ErrorOption {
	return errors.SetData(errorBody, message)
}

// RestoreDefaultErrorBody returns an ErrorOption (for use in ErrorClass
// creation or error instantiation) that restores the default error body shown
// by GetErrorBody for some subhierarchy of errors.
func RestoreDefaultErrorBody() errors.ErrorOption {
	return errors.SetData(errorBody, nil)
}

// GetStatusCode will return the status code associated with an error, and
// default_code if none is found.
func GetStatusCode(err error, default_code int) int {
	rv := errors.GetData(err, statusCode)
	sc, ok := rv.(int)
	if ok {
		return sc
	}
	return default_code
}

// GetErrorBody will return the user-visible error message given an error.
// The message will be determined by errors.GetMessage() unless the error class
// has an error body overridden by OverrideErrorBody.
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
