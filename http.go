// Copyright (C) 2014 Space Monkey, Inc.

package errors

import (
    "fmt"
)

var (
    httpStatusCode = GenSym()
    httpErrorBody  = GenSym()
)

func SetHTTPStatusCode(code int) ErrorOption {
    return SetData(httpStatusCode, code)
}

func OverrideHTTPErrorBody(message string) ErrorOption {
    return SetData(httpErrorBody, message)
}

func RestoreDefaultHTTPErrorBody() ErrorOption {
    return SetData(httpErrorBody, nil)
}

func GetHTTPStatusCode(err error, default_code int) int {
    rv := GetData(err, httpStatusCode)
    sc, ok := rv.(int)
    if ok {
        return sc
    }
    return default_code
}

func GetHTTPErrorBody(err error) string {
    rv := GetData(err, httpErrorBody)
    message, ok := rv.(string)
    if !ok {
        return GetMessage(err)
    }
    class := GetClass(err)
    if class == nil {
        return message
    }
    return fmt.Sprintf("%s: %s", class.String(), message)
}
