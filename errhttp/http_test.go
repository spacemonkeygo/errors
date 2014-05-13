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

package errhttp

import (
	"log"
	"net/http"
	"testing"

	"github.com/spacemonkeygo/errors"
)

func writeError(msg string, code int) {}

func Example(t *testing.T) {
	InvalidRequest := errors.NewClass("Invalid request",
		SetStatusCode(http.StatusBadRequest))

	process := func() error {
		// this method is some sample method somewhere that's doing request
		// processing
		return InvalidRequest.New("missing field or something")
	}

	handler := func(err error) {
		// this method is some sample method somewhere that's figuring out how
		// to display various errors to a user.
		if err != nil {
			code := GetStatusCode(err, 500)
			message := GetErrorBody(err)
			log.Printf("HTTP error %d: %s\n%s", code, message, err)
			writeError(message, code)
		}
	}

	// http event loop
	handler(process())
}
