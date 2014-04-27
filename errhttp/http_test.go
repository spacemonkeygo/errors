// Copyright (C) 2014 Space Monkey, Inc.

package errhttp

import (
	"log"
	"net/http"
	"testing"

	"github.com/SpaceMonkeyGo/errors"
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
